/**
 * Departure-linked NF-e emission inputs.
 * Exercises the modal -> preview path only. The preview route is SEFAZ-free;
 * this spec deliberately never clicks the final confirmation button.
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
    `INSERT INTO vehicle (plate, name, farm) VALUES ('E2E${suffix.slice(-4)}', 'departure nfe test', ${farmId}) RETURNING id`
  ));
  const personID = Number(db(`INSERT INTO person (ie, farm) VALUES ('123456789', ${farmId}) RETURNING id`));
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

test('departure NF-e payment, billing, transport and code choices survive preview', async ({ page, browserName }) => {
  const fixture = createDepartureFixture();
  try {
    await login(page, browserName);
    await goToDepartures(page);
    const emissionButton = page.locator(`#departure-${fixture.departureID} button[title="Emitir NF-e"]`);
    await expect(emissionButton).toBeVisible();
    await emissionButton.click();
    const modal = page.locator('#nfe-emit-dialog');
    await expect(modal).toBeVisible();

    await modal.locator('#unitPrice').fill('2.50');
    await modal.locator('#tpag').selectOption('01');
    await modal.locator('#indPag').selectOption('1');
    await modal.locator('#includeCobranca').uncheck();
    await modal.locator('#modFrete').selectOption('1');
    await modal.locator('#icmsCST').selectOption('51');
    await modal.locator('#pisCST').selectOption('08');
    await modal.locator('#cofinsCST').selectOption('08');

    await modal.locator('#transportNome').fill('Transportador E2E');
    await modal.locator('#transportDocumento').fill('529.982.247-25');
    await modal.locator('#transportLogradouro').fill('Rodovia MT-100');
    await modal.locator('#transportNumero').fill('Km 10');
    await modal.locator('#transportMunicipio').fill('Cuiaba');
    await modal.locator('#transportUF').fill('MT');
    await modal.locator('#transportPlaca').fill('ABC1D23');
    await modal.locator('#transportPlacaUF').fill('MT');
    await modal.locator('#transportRNTC').fill('12345');
    await modal.locator('#volQtd').fill('2');
    await modal.locator('#volEsp').fill('Sacas');
    await modal.locator('#volPesoL').fill('95,5');
    await modal.locator('#volPesoB').fill('100,25');

    await modal.locator('#useDefaultTaxRates').uncheck();
    await modal.locator('#icmsRate').fill('17.00');
    await modal.locator('#pisRate').fill('1.00');
    await modal.locator('#cofinsRate').fill('7.00');
    await modal.locator('#ibsRate').fill('0.10');
    await modal.locator('#cbsRate').fill('0.90');

    const previewResponse = page.waitForResponse(
      response => response.url().includes(`/nfe/preview/${fixture.departureID}`) && response.request().method() === 'POST'
    );
    await modal.locator('button[type="submit"]').evaluate(button => button.click());
    const response = await previewResponse;
    expect(response.status(), await response.headerValue('HX-Trigger')).toBe(200);

    const preview = page.locator('#nfe-emit-step');
    await expect(preview).toBeVisible();
    for (const [name, value] of [
      ['tpag', '01'],
      ['indPag', '1'],
      ['includeCobranca', '0'],
      ['transportNome', 'Transportador E2E'],
      ['transportDocumento', '529.982.247-25'],
      ['transportPlaca', 'ABC1D23'],
      ['transportPlacaUF', 'MT'],
      ['transportRNTC', '12345'],
      ['volQtd', '2'],
      ['volEsp', 'Sacas'],
      ['volPesoL', '95,5'],
      ['volPesoB', '100,25'],
      ['icmsCST', '51'],
      ['pisCST', '08'],
      ['cofinsCST', '08'],
    ]) {
      await expect(preview.locator(`input[name="${name}"]`)).toHaveValue(value);
    }

    // No confirmation/emission request is made in this test.
    await expect(preview.locator('#confirm-nfe')).toBeVisible();
  } finally {
    fixture.cleanup();
  }
});
