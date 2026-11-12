/**
 * Comma-decimal mask on the departure-linked NF-e emission modal
 * (templates/nfe/nfe-emit-modal.html, via assets/js/form/nfe-form.js).
 *
 * Complements the pre-existing payment/billing spec, which cannot reach these
 * fields pending the indPag/includeCobranca work; here only the mask
 * guarantees are exercised:
 * - comma-typed unitPrice / volumes display grouped pt-BR with dot raws,
 * - the submission replaces the parameters with the dot-decimal raw,
 * - the request reaches the certificate gate (SEFAZ isolation: the fixture
 *   has no certificate bytes) instead of the decimal parse toasts.
 *
 * The modal's rate inputs are server-rendered disabled; the mask normalizes
 * their default display ("17.00" → "17,00") without making them editable.
 */

import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import { login } from '../utils/auth';
import { goToDepartures } from '../utils/departure-helpers';

const DB = {
  host: 'localhost',
  port: '5433',
  user: 'test',
  password: 'test',
  database: 'armazenda_test',
};

function db(sql) {
  return execSync(
    `PGPASSWORD=${DB.password} psql -v ON_ERROR_STOP=1 -h ${DB.host} -p ${DB.port} -U ${DB.user} -d ${DB.database} -tAc "${sql.replace(/"/g, '\\"')}"`,
    { encoding: 'utf8' }
  ).trim().split('\n')[0];
}

function createDepartureFixture() {
  const suffix = `${Date.now()}${Math.floor(Math.random() * 10000)}`;
  const farmId = Number(db(`SELECT id FROM farm WHERE inscricao_estadual = '123456789' LIMIT 1`));
  const cropId = Number(db(`SELECT id FROM crop WHERE farm = ${farmId} AND name = 'Milho Default' LIMIT 1`));
  if (!farmId || !cropId) throw new Error('E2E farm/crop fixture is missing');
  db(`INSERT INTO ibge_municipio (code, name, uf) VALUES ('5103403', 'Cuiabá', 'MT')
      ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, uf = EXCLUDED.uf`);

  const vehicleID = Number(db(
    // The plate tail carries the random component as well (parallel-safe).
    `INSERT INTO vehicle (plate, name, farm) VALUES ('E2E${suffix.slice(-4)}', 'departure nfe test', ${farmId}) RETURNING id`
  ));
  // A per-run IE avoids the parallel-worker conflict on the partial unique
  // index unique_person_ie_in_farm (farm, ie). The tail carries the random
  // part; the head is the shared millisecond timestamp.
  const ie = (suffix.replace(/\D/g, '').slice(-9) || suffix.slice(-9)).slice(0, 9);
  const personID = Number(db(`INSERT INTO person (ie, farm) VALUES ('${ie}', ${farmId}) RETURNING id`));
  db(`INSERT INTO natural_person (name, cpf, personid) VALUES ('Destinatário E2E ${suffix}', '52998224725', ${personID})`);
  db(`INSERT INTO address (street, cep, number, neighborhood, city, state, person_id) VALUES ('Rodovia Teste', '78000000', 10, 'Centro', 'Cuiabá', 'MT', ${personID})`);
  const departureID = Number(db(
    `INSERT INTO departure (departureDate, vehicle, crop, grossWeight, tare, netWeight, farm)
     VALUES (CURRENT_TIMESTAMP, ${vehicleID}, ${cropId}, 120, 20, 100, ${farmId}) RETURNING id`
  ));
  db(`INSERT INTO departure_recipient (departure_id, person_id) VALUES (${departureID}, ${personID})`);

  return {
    departureID,
    cleanup() {
      db(`DELETE FROM departure_recipient WHERE departure_id = ${departureID}`);
      db(`DELETE FROM departure WHERE id = ${departureID}`);
      db(`DELETE FROM address WHERE person_id = ${personID}`);
      db(`DELETE FROM natural_person WHERE personid = ${personID}`);
      db(`DELETE FROM person WHERE id = ${personID}`);
      db(`DELETE FROM vehicle WHERE id = ${vehicleID}`);
    },
  };
}

test('comma-typed unit price and volumes in the emission modal submit dot-decimals', async ({ page, browserName }) => {
  const fixture = createDepartureFixture();
  try {
    await login(page, browserName);
    await goToDepartures(page);
    const emissionButton = page.locator(`#departure-${fixture.departureID} button[title="Emitir NF-e"]`);
    await expect(emissionButton).toBeVisible();
    // evaluate-click keeps WebKit happy with the backdrop stacking.
    await emissionButton.evaluate(el => el.click());
    const modal = page.locator('#nfe-emit-dialog');
    await expect(modal).toBeVisible();

    // Comma decimals in the modal's editable numeric fields: display is
    // grouped pt-BR, data-raw keeps the dot-decimal.
    const unitPrice = modal.locator('#unitPrice');
    await unitPrice.fill('85,50');
    await expect(unitPrice).toHaveValue('85,50');
    await expect(unitPrice).toHaveAttribute('data-raw', '85.50');

    // Clear-then-fill for the server-prefilled volume fields (the SDD's
    // deterministic pattern): a fill over a non-empty masked field inserts at
    // the caret — Playwright selects, but the mask's focus swap (display →
    // editable raw) resets the selection — so the pasted text appends. A real
    // user's ctrl+v behaves the same way at a caret; replacing the derived
    // value is a clear-then-paste.
    const volPesoL = modal.locator('#volPesoL');
    await volPesoL.fill('');
    await volPesoL.fill('95,5');
    await expect(volPesoL).toHaveValue('95,5');
    await expect(volPesoL).toHaveAttribute('data-raw', '95.5');

    const volPesoB = modal.locator('#volPesoB');
    await volPesoB.fill('');
    await volPesoB.fill('100,25');
    await expect(volPesoB).toHaveValue('100,25');
    await expect(volPesoB).toHaveAttribute('data-raw', '100.25');

    // The server-rendered rate defaults display in the pt-BR form while the
    // raw keeps the dot; the inputs stay disabled (server contract). The
    // default comes from the product config, so assert it dynamically.
    const icmsRate = modal.locator('#icmsRate');
    const serverDefault = await icmsRate.getAttribute('value'); // "12.00" per fixture config
    await expect(icmsRate).toBeDisabled();
    await expect(icmsRate).toHaveAttribute('data-raw', serverDefault);
    await expect(icmsRate).toHaveValue(serverDefault.replace('.', ','));

    const previewRequest = page.waitForRequest(
      r => r.url().includes(`/nfe/preview/${fixture.departureID}`) && r.method() === 'POST'
    );
    const previewResponse = page.waitForResponse(
      r => r.url().includes(`/nfe/preview/${fixture.departureID}`) && r.request().method() === 'POST'
    );
    await modal.locator('#includeVehicle').uncheck();
    await modal.locator('button[type="submit"]').evaluate(button => button.click());

    // The mask override replaces the parameters with the dot-decimal raw:
    // a comma never leaves the browser.
    const params = new URLSearchParams((await previewRequest).postData());
    expect(params.get('unitPrice')).toBe('85.50');
    expect(params.get('volPesoL')).toBe('95.5');
    expect(params.get('volPesoB')).toBe('100.25');

    // The request reaches the certificate gate (SEFAZ isolation: no
    // certificate bytes in the fixture) — not the decimal parse toasts.
    const response = await previewResponse;
    expect(response.status()).toBe(200);
  } finally {
    fixture.cleanup();
  }
});
