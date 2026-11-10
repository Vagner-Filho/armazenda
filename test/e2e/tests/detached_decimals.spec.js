/**
 * Comma-decimal UX of the NF-e numeric inputs (docs/sdd/decimal_comma_mask).
 *
 * The placeholders teach pt-BR comma decimals ("0,000", "17,00"), while the
 * backend contract is dot-decimal only. The mask (`assets/js/decimalMask.js`)
 * must:
 * - display what the user typed as grouped pt-BR ("1250,90" → "1.250,90"),
 * - keep the canonical dot-decimal in the element's data-raw,
 * - translate the htmx submission to the dot-decimal raw via the
 *   htmx:configRequest parameter override,
 * - and let the comma-driven forms pass the server's decimal parsing exactly
 *   like dot-driven ones (the parse-level toasts — "Quantidade inválida",
 *   "Falha ao identificar as alíquotas" — never fire).
 *
 * The seeded farm carries an EMPTY certificate_data (SEFAZ isolation — never
 * seed certificate rows), so the preview application stops at the
 * certificate gate *after* the form parsers with a "Certificado digital"
 * toast. That toast arriving (instead of the parse toasts) is the accepted
 * evidence that the comma-derived payload parsed clean. The draft-save path
 * has no certificate gate and round-trips comma-typed prices into the
 * detached_nfe_profile rows, which the DB checks assert as dot-decimals.
 */

import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import { faker } from '@faker-js/faker';
import { login } from '../utils/auth';

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

function uniqueSuffix() {
  return `${Date.now().toString(36)}${Math.floor(Math.random() * 1e6).toString(36)}`;
}

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
  // Element click keeps WebKit's dialog/backdrop-stacking click-hang away
  // (the tabs switch is listener-driven; same as the evaluate-click pattern
  // the other emission specs use for buttons).
  await page.locator(`#tabs button#${tabId}`).evaluate(el => el.click());
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

// Clicks the preview submit from the last step and captures its request and
// response promise.
async function submitPreview(page) {
  await goToTab(page, 'trib');
  const previewButton = page.locator('[data-test-id="preview-detached-nfe"]');
  await expect(previewButton).toBeEnabled();
  const previewRequest = page.waitForRequest(
    r => r.url().includes('/nfe/emitir/preview') && r.method() === 'POST'
  );
  const previewResponse = page.waitForResponse(
    r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
  );
  await previewButton.evaluate(el => el.click());
  return { previewRequest, previewResponse };
}

async function gotoEmitPage(page, browserName) {
  await login(page, browserName);
  await gotoApp(page, '/nfe/emitir');
  await page.reload();
}

async function fillInlineRecipient(page, nameSuffix) {
  await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
  await page.fill('#recipientName', `Destinatário Decimais ${nameSuffix}`);
  await page.fill('#recipientDocument', faker.string.numeric(14));
  await page.fill('#recipientStreet', 'Rua Decimais');
  await page.fill('#recipientNumber', '30');
  await page.fill('#recipientNeighborhood', 'Centro');
  await page.fill('#recipientCity', 'Cuiabá');
  await page.fill('#recipientState', 'MT');
  await page.fill('#recipientCEP', '78000000');
}

function bodyParams(request) {
  return new URLSearchParams(request.postData());
}

test.describe('pt-BR decimal mask on the detached NF-e form', () => {
  test('item decimals typed with commas display grouped pt-BR and store the dot raw', async ({ page, browserName }) => {
    await gotoEmitPage(page, browserName);

    await goToTab(page, 'transac');
    const row = page.locator('#items-container .item-row').first();
    await row.locator('input[name="items[0].productName"]').fill('Soja Decimais E2E');
    await selectCfop(row, '5101');

    const quantity = row.locator('input[name="items[0].quantity"]');
    const unitPrice = row.locator('input[name="items[0].unitPrice"]');
    const discount = row.locator('input[name="items[0].discount"]');
    const grossWeight = row.locator('input[name="items[0].grossWeight"]');

    // Typing the placeholder-promised comma format…
    await quantity.fill('1250,90');
    await unitPrice.fill('17,00');
    await discount.fill('1,25');
    await grossWeight.fill('1250,90');

    // …displays the full pt-BR grouping while data-raw keeps the dot contract.
    await expect(quantity).toHaveValue('1.250,90');
    await expect(unitPrice).toHaveValue('17,00');
    await expect(discount).toHaveValue('1,25');
    await expect(grossWeight).toHaveValue('1.250,90');

    await expect(quantity).toHaveAttribute('data-raw', '1250.90');
    await expect(unitPrice).toHaveAttribute('data-raw', '17.00');
    await expect(discount).toHaveAttribute('data-raw', '1.25');
    await expect(grossWeight).toHaveAttribute('data-raw', '1250.90');

    // Focus shows the editable comma raw (no grouping), blur reformats the
    // display. Direct focus() is deterministic in every engine (a synthetic
    // click does not focus on Firefox/WebKit).
    await quantity.evaluate(el => el.focus());
    await expect(quantity).toHaveValue('1250,90');
    await unitPrice.evaluate(el => el.focus());
    await expect(quantity).toHaveValue('1.250,90');

    // Dot-typed input normalizes to the same values (deterministic
    // clear-then-fill: an insert at the caret would append, which is exactly
    // what human caret editing produces).
    await quantity.fill('');
    await quantity.fill('1250.90');
    await expect(quantity).toHaveValue('1.250,90');
    await expect(quantity).toHaveAttribute('data-raw', '1250.90');
  });

  test('comma-typed values reach the preview endpoint as dot-decimal raws', async ({ page, browserName }) => {
    await gotoEmitPage(page, browserName);

    await fillInlineRecipient(page, 'Payload');

    await goToTab(page, 'transac');
    const row = page.locator('#items-container .item-row').first();
    await row.locator('input[name="items[0].productName"]').fill('Soja Decimais E2E');
    await selectCfop(row, '5101');
    await row.locator('input[name="items[0].quantity"]').fill('1250,90');
    await row.locator('input[name="items[0].unitPrice"]').fill('17,00');
    await row.locator('input[name="items[0].discount"]').fill('1,25');
    await row.locator('input[name="items[0].grossWeight"]').fill('7,350');

    await goToTab(page, 'transp');
    await page.locator('input[name="transportType"][value="new"]').check({ force: true });
    await page.fill('#transportPlaca', `${faker.string.alpha({ length: 3 }).toUpperCase()}${faker.string.numeric(4)}`);

    const { previewRequest, previewResponse } = await submitPreview(page);
    const params = bodyParams(await previewRequest);

    // The server contract: dot-decimal; a comma never leaves the browser.
    expect(params.get('items[0].quantity')).toBe('1250.90');
    expect(params.get('items[0].unitPrice')).toBe('17.00');
    expect(params.get('items[0].discount')).toBe('1.25');
    expect(params.get('items[0].grossWeight')).toBe('7.350');

    // The certificate gate answers AFTER the form parsers (SEFAZ isolation:
    // the fixture carries no certificate bytes). Its toast — and not the
    // "Quantidade inválida" / "Falha ao identificar as alíquotas" one —
    // proves the comma-derived dot payload parsed clean.
    const response = await previewResponse;
    const trigger = (await response.headerValue('HX-Trigger')) || '';
    expect(trigger).toContain('Certificado digital');
    expect(trigger).not.toContain('Quantidade');
    expect(trigger).not.toContain('alíquotas');
  });

  test('comma-typed rates submit dot-decimal raws without the rate parse toast', async ({ page, browserName }) => {
    await gotoEmitPage(page, browserName);

    await fillInlineRecipient(page, 'Alíquotas');

    await goToTab(page, 'transac');
    const row = page.locator('#items-container .item-row').first();
    await row.locator('input[name="items[0].productName"]').fill('Soja Alíquotas E2E');
    await selectCfop(row, '5101');
    await row.locator('input[name="items[0].quantity"]').fill('10');
    await row.locator('input[name="items[0].unitPrice"]').fill('50,00');

    await goToTab(page, 'trib');
    // Unchecked "Usar taxa padrão" makes the rate inputs editable.
    await page.locator('#useDefaultTaxRates').uncheck();

    const icmsRate = page.locator('#icmsRate');
    await icmsRate.fill('17,00');
    // Display and raw agree with the mask contract on a rate field too.
    await expect(icmsRate).toHaveValue('17,00');
    await expect(icmsRate).toHaveAttribute('data-raw', '17.00');

    // Re-focus the unit price (WebKit-safe) to keep the editable-raw
    // behavior covered.
    await page.locator('input[name="items[0].unitPrice"]').evaluate(el => el.focus());

    const { previewRequest, previewResponse } = await submitPreview(page);
    const params = bodyParams(await previewRequest);
    expect(params.get('icmsRate')).toBe('17.00');
    expect(params.get('items[0].unitPrice')).toBe('50.00');

    const response = await previewResponse;
    const trigger = (await response.headerValue('HX-Trigger')) || '';
    expect(trigger).toContain('Certificado digital');
    expect(trigger).not.toContain('alíquotas');
  });

  test('cloned rows and the volume prefill follow the mask', async ({ page, browserName }) => {
    await gotoEmitPage(page, browserName);

    await goToTab(page, 'transac');
    const firstRow = page.locator('#items-container .item-row').first();
    await firstRow.locator('input[name="items[0].productName"]').fill('Soja Clones E2E');
    await selectCfop(firstRow, '5101');
    await firstRow.locator('input[name="items[0].quantity"]').fill('1250,90');
    await firstRow.locator('input[name="items[0].grossWeight"]').fill('1250,90');
    await firstRow.locator('input[name="items[0].unitPrice"]').fill('17,00');

    await page.locator('#add-item-btn').evaluate(el => el.click());
    const secondRow = page.locator('#items-container .item-row').nth(1);
    await secondRow.locator('input[name="items[1].productName"]').fill('Milho Clones E2E');
    await selectCfop(secondRow, '5102');
    await secondRow.locator('input[name="items[1].quantity"]').fill('300,5');
    await secondRow.locator('input[name="items[1].unitPrice"]').fill('30,00');

    // Both rows are masked, including the clone.
    await expect(secondRow.locator('input[name="items[1].quantity"]')).toHaveValue('300,5');
    await expect(secondRow.locator('input[name="items[1].quantity"]')).toHaveAttribute('data-raw', '300.5');

    // The volume prefill mirrors the item sums in the pt-BR display.
    await goToTab(page, 'transp');
    await expect(page.locator('#volQtd')).toHaveValue('1');
    await expect(page.locator('#volPesoL')).toHaveValue('1.551,4');
    await expect(page.locator('#volPesoB')).toHaveValue('1.250,9');

    // And the mirrored fields keep the dot raw for submission.
    await expect(page.locator('#volPesoL')).toHaveAttribute('data-raw', '1551.4');
    await expect(page.locator('#volPesoB')).toHaveAttribute('data-raw', '1250.9');
  });

  test('comma-typed prices round-trip a draft save as dot-decimals in the DB', async ({ page, browserName }) => {
    const name = `Decimais Rascunho ${uniqueSuffix()}`;
    try {
      await gotoEmitPage(page, browserName);

      await fillInlineRecipient(page, 'Rascunho');

      await goToTab(page, 'transac');
      const row = page.locator('#items-container .item-row').first();
      await row.locator('input[name="items[0].productName"]').fill('Soja Rascunho E2E');
      await selectCfop(row, '5101');
      await row.locator('input[name="items[0].unitPrice"]').fill('17,25');
      await page.fill('#naturezaOp', 'Venda de decimais');

      await page.locator('[data-test-id="open-save-rascunho"]').evaluate(el => el.click());
      const saveDialog = page.locator('dialog#save-rascunho-dialog');
      await expect(saveDialog).toBeVisible();
      await saveDialog.locator('input[name="rascunhoNome"]').fill(name);

      const saveResponse = page.waitForResponse(
        r => r.url().includes('/nfe/rascunho/salvar') && r.request().method() === 'POST'
      );
      await saveDialog.locator('[data-test-id="confirm-save-rascunho"]').evaluate(el => el.click());
      expect((await saveResponse).status()).toBe(204);

      // The saved rascunho stores the dot-decimal — the comma never reaches
      // the server and the decimal value is exact.
      expect(
        db(`SELECT items_json->0->>'unit_price' FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${name}'`)
      ).toBe('17.25');
    } finally {
      db(`DELETE FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${name}'`);
    }
  });
});
