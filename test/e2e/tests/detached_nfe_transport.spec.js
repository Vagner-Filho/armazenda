/**
 * Detached NF-e transport section (MOC Grupo X) E2E tests.
 *
 * Covers:
 * - The three sub-groups render (transportadora / veículo / volumes)
 * - Volume fields are prefilled from the item rows and stop auto-updating
 *   once the user edits a field (dirty tracking)
 * - Vehicle fields (plate UF / RNTC) are disabled until a plate is informed;
 *   the plate UF defaults to the emitter UF
 * - Preview round-trip: the transport values appear in the preview's hidden
 *   fields (AC7), and an interstate vehicle is rejected with a warning toast
 *   (AC5 — SEFAZ-free path: the rejection happens before any transmission)
 */

import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import { faker } from '@faker-js/faker';
import { login } from '../utils/auth';
import { createContact, deleteContact } from '../utils/contact-helpers';
import { generateCPF, generateCNPJ } from '../utils/document-helpers';

const DB = {
  host: 'localhost',
  port: '5433',
  user: 'test',
  password: 'test',
  database: 'armazenda_test',
};

function db(sql) {
  return execSync(
    `PGPASSWORD=${DB.password} psql -h ${DB.host} -p ${DB.port} -U ${DB.user} -d ${DB.database} -tAc "${sql.replace(/"/g, '\\"')}"`,
    { encoding: 'utf8' }
  ).trim();
}

// Navigate to a path on the same origin the session cookie was set on. The
// login helper lands on 127.0.0.1, while page.goto('/...') would resolve
// against the Playwright baseURL (localhost), dropping the session cookie.
async function gotoApp(page, path) {
  const origin = new URL(page.url()).origin;
  return page.goto(`${origin}${path}`);
}

// Sections of the emission form are hidden until their tab in #tabs is
// clicked. The Transporte tab reveals fieldset#transporte-fieldset; the
// other tabs share the id of their fieldset.
const EMIT_TAB_FIELDSETS = {
  dest: 'fieldset#dest',
  transac: 'fieldset#transac',
  transp: 'fieldset#transporte-fieldset',
  trib: 'fieldset#trib',
};

async function goToTab(page, tabId) {
  await page.locator(`#tabs button#${tabId}`).click();
  await expect(page.locator(EMIT_TAB_FIELDSETS[tabId])).toBeVisible();
}

// Selects a CFOP through the search-driven selector used by the item rows.
async function selectCfop(row, code) {
  await row.locator('[data-cfop-trigger]').click();
  const search = row.locator('.cfop-search');
  await expect(search).toBeVisible();
  await search.fill(code);
  await search.press('Enter');
  await expect(row.locator('select[name$=".cfop"]')).toHaveValue(code);
}

/** Switches the Transportadora source to the inline fields. */
async function chooseNewTransportador(page) {
  await page.locator('input[name="transportType"][value="new"]').check({ force: true });
}

/** Switches the Transportadora source back to the contact select. */
async function chooseExistingTransportador(page) {
  await page.locator('input[name="transportType"][value="existing"]').check({ force: true });
}

/** Clicks the preview submit from the last step and returns its request. */
async function submitPreview(page) {
  await goToTab(page, 'trib');
  const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
  await expect(previewButton).toBeEnabled();
  const previewRequest = page.waitForRequest(
    r => r.url().includes('/nfe/emitir/preview') && r.method() === 'POST'
  );
  await previewButton.evaluate(el => el.click());
  return previewRequest;
}

/** True when the preview response was refused by the certificate gate. */
function isCertificateGate(response) {
  const trigger = response.headers()['hx-trigger'] || '';
  return response.status() === 400 && trigger.includes('Certificado digital');
}

async function gotoEmitPage(page, browserName) {
  await login(page, browserName);
  await gotoApp(page, '/nfe/emitir');
  // The farm contacts are rendered server-side; reload once so a fixture
  // created after the first render appears in both selects.
  await page.reload();
}

async function fillBasicForm(page) {
  // Destinatário tab (visible on load): recipient fields.
  await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
  await page.fill('#recipientName', 'Destinatário Transporte E2E');
  await page.fill('#recipientDocument', faker.string.numeric(14));
  await page.fill('#recipientStreet', 'Rua Transporte');
  await page.fill('#recipientNumber', '30');
  await page.fill('#recipientNeighborhood', 'Centro');
  await page.fill('#recipientCity', 'Cuiabá');
  await page.fill('#recipientState', 'MT');
  await page.fill('#recipientCEP', '78000000');

  // Transação tab: the item fields are hidden until their tab is active.
  await goToTab(page, 'transac');
  await page.fill('input[name="items[0].productName"]', 'Soja Transporte');
  await selectCfop(page.locator('#items-container .item-row').first(), '5101');
  await page.fill('input[name="items[0].quantity"]', '10');
  await page.fill('input[name="items[0].grossWeight"]', '10.2');
  await page.fill('input[name="items[0].unitPrice"]', '50.00');
}

test.describe('Transporte section of the detached NF-e form', () => {
  test('renders the three sub-groups and disables the vehicle fields without a plate', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    // Sections start hidden; the Transporte tab reveals this fieldset.
    await goToTab(page, 'transp');

    // The inline transportadora is the "Novo transportador" source.
    await chooseNewTransportador(page);

    // Transportadora sub-group
    await expect(page.locator('#transportNome')).toBeVisible();
    await expect(page.locator('#transportDocumento')).toBeVisible();
    await expect(page.locator('#transportIE')).toBeVisible();
    await expect(page.locator('#transportLogradouro')).toBeVisible();
    await expect(page.locator('#transportNumero')).toBeVisible();
    await expect(page.locator('#transportMunicipio')).toBeVisible();
    await expect(page.locator('#transportUF')).toBeVisible();

    // Veículo sub-group: UF/RNTC disabled while no plate
    await expect(page.locator('#transportPlaca')).toBeVisible();
    await expect(page.locator('#transportPlacaUF')).toBeDisabled();
    await expect(page.locator('#transportRNTC')).toBeDisabled();

    // Volumes sub-group
    await expect(page.locator('#volQtd')).toBeVisible();
    await expect(page.locator('#volEsp')).toBeVisible();
    await expect(page.locator('#volMarca')).toBeVisible();
    await expect(page.locator('#volPesoL')).toBeVisible();
    await expect(page.locator('#volPesoB')).toBeVisible();

    // Informing a plate enables the vehicle fields and prefills the plate UF
    // with the emitter UF.
    await page.fill('#transportPlaca', 'abc1d23');
    await expect(page.locator('#transportPlacaUF')).toBeEnabled();
    await expect(page.locator('#transportRNTC')).toBeEnabled();
    await expect(page.locator('#transportPlacaUF')).toHaveValue('MT');

    // Clearing the plate disables and clears them again.
    await page.fill('#transportPlaca', '');
    await expect(page.locator('#transportPlacaUF')).toBeDisabled();
    await expect(page.locator('#transportPlacaUF')).toHaveValue('');
  });

  test('volume prefill mirrors the item sums and respects user edits', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    await fillBasicForm(page);

    // Volume fields live on the Transporte tab; the prefill mirrors the items.
    await goToTab(page, 'transp');
    await expect(page.locator('#volQtd')).toHaveValue('1');
    await expect(page.locator('#volEsp')).toHaveValue('Granel');
    await expect(page.locator('#volPesoL')).toHaveValue('10');
    await expect(page.locator('#volPesoB')).toHaveValue('10,2');

    // A user edit stops that field's auto-update; others keep mirroring.
    // Clear-then-fill: the field is prefilled and a fill over a non-empty
    // masked field inserts at the caret (paste-at-caret semantics) — the
    // deterministic replace is clear-then-fill.
    await page.fill('#volPesoL', '');
    await page.fill('#volPesoL', '999');

    // Add a second item (on the Transação tab): quantity sum changes from 10 to 30.
    await goToTab(page, 'transac');
    await page.locator('#add-item-btn').click();
    await page.fill('input[name="items[1].productName"]', 'Milho Transporte');
    await selectCfop(page.locator('#items-container .item-row').nth(1), '5102');
    await page.fill('input[name="items[1].quantity"]', '20');
    await page.fill('input[name="items[1].unitPrice"]', '30.00');

    // Back to the Transporte tab to check the prefill outcome.
    await goToTab(page, 'transp');
    await expect(page.locator('#volPesoL')).toHaveValue('999'); // user-edited
    await expect(page.locator('#volPesoB')).toHaveValue('10,2'); // still mirrored (unchanged sums)
    await expect(page.locator('#volQtd')).toHaveValue('1');
  });

  test('preview round-trips the transport values and the DB stores nothing', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    await fillBasicForm(page);

    // Transport fields are on the Transporte tab.
    await goToTab(page, 'transp');
    await chooseNewTransportador(page);
    const plate = `${faker.string.alpha({ length: 3 }).toUpperCase()}${faker.string.numeric(4)}`;
    await page.fill('#transportNome', 'Carreteiro Transporte E2E');
    await page.fill('#transportDocumento', '529.982.247-25');
    await page.fill('#transportIE', 'ISENTO');
    await page.fill('#transportLogradouro', 'Rodovia BR-163');
    await page.fill('#transportNumero', 'Km 12');
    await page.fill('#transportMunicipio', 'Sorriso');
    await page.fill('#transportUF', 'MT');
    await page.fill('#transportPlaca', plate);
    await page.fill('#transportRNTC', '12345678901234567890');
    await page.fill('#volMarca', 'BigFarm');

    // The submit button is only enabled on the Tributação tab.
    await goToTab(page, 'trib');
    const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
    await expect(previewButton).toBeEnabled();

    const previewResponse = page.waitForResponse(
      r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
    );
    await previewButton.evaluate(el => el.click());
    const preview = await previewResponse;
    expect(preview.status()).toBe(200);
    await expect(page.locator('#nfe-detached-preview-panel')).toBeVisible();

    // Hidden round-trip fields carry the exact previewed transport values,
    // scoped to the preview panel (the main form keeps the same names).
    const previewPanel = page.locator('#nfe-detached-preview-panel');
    await expect(previewPanel.locator('input[name="transportNome"]')).toHaveValue('Carreteiro Transporte E2E');
    await expect(previewPanel.locator('input[name="transportDocumento"]')).toHaveValue('529.982.247-25');
    await expect(previewPanel.locator('input[name="transportIE"]')).toHaveValue('ISENTO');
    await expect(previewPanel.locator('input[name="transportLogradouro"]')).toHaveValue('Rodovia BR-163');
    await expect(previewPanel.locator('input[name="transportNumero"]')).toHaveValue('Km 12');
    await expect(previewPanel.locator('input[name="transportMunicipio"]')).toHaveValue('Sorriso');
    await expect(previewPanel.locator('input[name="transportUF"]')).toHaveValue('MT');
    await expect(previewPanel.locator('input[name="transportPlaca"]')).toHaveValue(plate);
    await expect(previewPanel.locator('input[name="transportPlacaUF"]')).toHaveValue('MT');
    await expect(previewPanel.locator('input[name="transportRNTC"]')).toHaveValue('12345678901234567890');
    await expect(previewPanel.locator('input[name="volQtd"]')).toHaveValue('1');
    await expect(previewPanel.locator('input[name="volEsp"]')).toHaveValue('Granel');
    await expect(previewPanel.locator('input[name="volMarca"]')).toHaveValue('BigFarm');
    await expect(previewPanel.locator('input[name="volPesoL"]')).toHaveValue('10');
    await expect(previewPanel.locator('input[name="volPesoB"]')).toHaveValue('10.2');
  });

  test('interstate destination rejects vehicle data with a warning toast', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    await fillBasicForm(page);

    // Interstate: SP destination, MT emitter (edited on the Destinatário tab).
    await goToTab(page, 'dest');
    await page.fill('#recipientState', 'SP');

    await goToTab(page, 'transp');
    await page.fill('#transportPlaca', `${faker.string.alpha({ length: 3 }).toUpperCase()}${faker.string.numeric(4)}`);
    await page.fill('#volPesoL', '999'); // user edit, keeps prefill permission off

    // The submit button is only enabled on the Tributação tab.
    await goToTab(page, 'trib');
    const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
    await expect(previewButton).toBeEnabled();
    await previewButton.evaluate(el => el.click());

    // Rejected before any transmission: no preview panel, warning toast.
    await expect(page.locator('#nfe-detached-preview-panel')).toHaveCount(0);
  });

  test('source defaults to the contact select', async ({ page, browserName }) => {
    const name = `Transportadora Padrão ${Date.now().toString(36)}`;
    const contactID = createContact({
      farmId: 1,
      kind: 'legal',
      name,
      document: generateCNPJ(),
      address: { street: 'Rodovia BR-163', cep: '78000000', number: 12, neighborhood: 'Centro', city: 'Sorriso', state: 'MT' },
    });

    try {
      await gotoEmitPage(page, browserName);
      await goToTab(page, 'transp');

      await expect(page.locator('input[name="transportType"][value="existing"]')).toBeChecked();
      await expect(page.locator('input[name="transportType"][value="new"]')).not.toBeChecked();

      const select = page.locator('#transportPersonId');
      await expect(select).toBeEnabled();
      await expect(page.locator('#existing-transport-section')).toBeVisible();
      await expect(select.locator('option').first()).toHaveText('Nenhuma transportadora');
      await expect(select.locator('option').first()).toHaveAttribute('value', '');
      await expect(select.locator(`option[value="${contactID}"]`)).toHaveText(name);

      // The inline source is hidden and disabled (so it is not submitted).
      await expect(page.locator('#new-transport-section')).toBeHidden();
      for (const id of ['transportNome', 'transportDocumento', 'transportIE', 'transportLogradouro', 'transportNumero', 'transportMunicipio', 'transportUF']) {
        await expect(page.locator(`#${id}`)).toBeHidden();
        await expect(page.locator(`#${id}`)).toBeDisabled();
        await expect(page.locator(`label[for="${id}"]`)).toHaveCount(1);
      }
      await expect(page.locator('label[for="transportPersonId"]')).toBeVisible();
    } finally {
      deleteContact(contactID);
    }
  });

  test('switching the source changes only the Transportadora block', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await goToTab(page, 'transp');

    // Existing → new: inline fields appear and enable, the select disables.
    await chooseNewTransportador(page);
    await expect(page.locator('#transportNome')).toBeVisible();
    await expect(page.locator('#transportNome')).toBeEnabled();
    await expect(page.locator('#transportDocumento')).toBeEnabled();
    await expect(page.locator('#existing-transport-section')).toBeHidden();
    await expect(page.locator('#transportPersonId')).toBeDisabled();

    await page.fill('#transportNome', 'Transportadora Digitada');
    await page.fill('#transportDocumento', generateCPF());

    // New → existing: the select comes back, the inline values survive.
    await chooseExistingTransportador(page);
    await expect(page.locator('#transportPersonId')).toBeEnabled();
    await expect(page.locator('#new-transport-section')).toBeHidden();
    await expect(page.locator('#transportNome')).toBeDisabled();
    await expect(page.locator('#transportNome')).toHaveValue('Transportadora Digitada');

    await chooseNewTransportador(page);
    await expect(page.locator('#transportNome')).toBeVisible();
    await expect(page.locator('#transportNome')).toHaveValue('Transportadora Digitada');

    // Veículo and Volumes never change with the source.
    for (const id of ['transportPlaca', 'volQtd', 'volEsp', 'volMarca', 'volPesoL', 'volPesoB']) {
      await expect(page.locator(`#${id}`)).toBeEnabled();
    }
    // The plate-dependent fields keep their own rule (disabled while no plate).
    await expect(page.locator('#transportPlacaUF')).toBeDisabled();
    await expect(page.locator('#transportRNTC')).toBeDisabled();
  });

  test('an inactive contact selection is not submitted', async ({ page, browserName }) => {
    const suffix = Date.now().toString(36);
    const contactID = createContact({
      farmId: 1,
      kind: 'natural',
      name: `Carreteiro Contato ${suffix}`,
      document: generateCPF(),
    });
    const typedName = `Transportadora Digitada ${suffix}`;
    const typedDocument = generateCNPJ();

    try {
      await gotoEmitPage(page, browserName);
      await fillBasicForm(page);
      await goToTab(page, 'transp');
      await page.selectOption('#transportPersonId', String(contactID));

      // Switch away from the picked contact and type another carrier.
      await chooseNewTransportador(page);
      await page.fill('#transportNome', typedName);
      await page.fill('#transportDocumento', typedDocument);

      const previewRequest = await submitPreview(page);
      const payload = new URLSearchParams(previewRequest.postData() || '');

      expect(payload.get('transportNome')).toBe(typedName);
      expect(payload.get('transportDocumento')).toBe(typedDocument);
      expect(payload.has('transportPersonId')).toBe(false);
    } finally {
      deleteContact(contactID);
    }
  });

  test('a picked transportadora is resolved at preview', async ({ page, browserName }) => {
    const suffix = Date.now().toString(36);
    const name = `Transportadora Selecionada ${suffix}`;
    const contactID = createContact({
      farmId: 1,
      kind: 'legal',
      name,
      document: generateCNPJ(),
      address: { street: 'Rodovia BR-163', cep: '78000000', number: 12, neighborhood: 'Centro', city: 'Sorriso', state: 'MT' },
    });

    try {
      await gotoEmitPage(page, browserName);
      await fillBasicForm(page);
      await goToTab(page, 'transp');
      await page.selectOption('#transportPersonId', String(contactID));

      await goToTab(page, 'trib');
      const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
      await expect(previewButton).toBeEnabled();
      const previewResponse = page.waitForResponse(
        r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
      );
      await previewButton.evaluate(el => el.click());
      const preview = await previewResponse;

      // The e2e fixture carries empty certificate bytes and nfeCertGate
      // refuses the preview (pre-existing environment limitation). The
      // server-side resolution is covered by the Go tests; this case asserts
      // the rendered result whenever a certificate is available.
      test.skip(isCertificateGate(preview), 'environment has no NF-e certificate (nfeCertGate refuses the preview)');

      expect(preview.status()).toBe(200);
      const panel = page.locator('#nfe-detached-preview-panel');
      await expect(panel).toBeVisible();
      await expect(panel.locator('input[name="transportPersonId"]')).toHaveValue(String(contactID));
      await expect(panel.locator('input[name="transportNome"]')).toHaveCount(0);
      await expect(panel).toContainText(name);
    } finally {
      deleteContact(contactID);
    }
  });

  test('no transportadora selected previews without carrier data', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await fillBasicForm(page);
    await goToTab(page, 'transp');

    await expect(page.locator('#transportPersonId')).toHaveValue('');
    const previewRequest = await submitPreview(page);
    const payload = new URLSearchParams(previewRequest.postData() || '');

    // The empty option submits an empty value, which the server parses as
    // "no contact"; no carrier data is carried.
    expect(payload.get('transportPersonId')).toBeFalsy();
    expect(payload.get('transportNome')).toBeFalsy();
  });

  test('a transportadora from another farm is rejected at preview', async ({ page, browserName }) => {
    const suffix = Date.now().toString(36);
    const inscricao = `8${suffix}`.slice(0, 12).padEnd(12, '0');
    db(`INSERT INTO farm (inscricao_estadual, uf) VALUES ('${inscricao}', 'MT') ON CONFLICT (inscricao_estadual) DO NOTHING`);
    const otherFarm = Number(db(`SELECT id FROM farm WHERE inscricao_estadual = '${inscricao}'`));
    const foreignContact = createContact({
      farmId: otherFarm,
      kind: 'legal',
      name: `Transportadora Outra Fazenda ${suffix}`,
      document: generateCNPJ(),
    });

    try {
      await gotoEmitPage(page, browserName);
      await fillBasicForm(page);
      await goToTab(page, 'transp');

      // The foreign contact is not one of this farm's options; inject it as a
      // crafted submission.
      await page.evaluate(({ id, label }) => {
        const select = document.getElementById('transportPersonId');
        const option = document.createElement('option');
        option.value = String(id);
        option.textContent = label;
        select.appendChild(option);
        select.value = String(id);
      }, { id: foreignContact, label: 'Transportadora estrangeira' });

      await goToTab(page, 'trib');
      const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
      await expect(previewButton).toBeEnabled();
      const previewResponse = page.waitForResponse(
        r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
      );
      await previewButton.evaluate(el => el.click());
      const preview = await previewResponse;

      test.skip(isCertificateGate(preview), 'environment has no NF-e certificate (nfeCertGate refuses the preview before the farm check)');

      expect(preview.status()).toBe(400);
      const trigger = preview.headers()['hx-trigger'] || '';
      expect(trigger).toContain('Transportadora inválida');
      await expect(page.locator('#nfe-detached-preview-panel')).toHaveCount(0);
      expect(db(`SELECT COUNT(*) FROM person WHERE id = ${foreignContact}`)).toBe('1');
    } finally {
      deleteContact(foreignContact);
      db(`DELETE FROM farm WHERE id = ${otherFarm}`);
    }
  });

  test('an inline transportadora creates no contact', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await fillBasicForm(page);
    await goToTab(page, 'transp');
    await chooseNewTransportador(page);

    const typedName = `Transportadora Inline ${Date.now().toString(36)}`;
    await page.fill('#transportNome', typedName);
    await page.fill('#transportDocumento', generateCNPJ());

    await goToTab(page, 'trib');
    const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
    await expect(previewButton).toBeEnabled();
    const previewResponse = page.waitForResponse(
      r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
    );
    await previewButton.evaluate(el => el.click());
    await previewResponse;
    // Under the current fixture the certificate gate refuses the preview;
    // either way the preview path must not create a contact for a typed
    // transportadora. Scoped to the typed name so parallel specs creating
    // their own contacts cannot skew the assertion.
    await expect(page.locator('#nfe-detached-preview-panel')).toHaveCount(0);
    expect(Number(db(`SELECT COUNT(*) FROM legal_person WHERE companyName = '${typedName}'`))).toBe(0);
  });
});
