/**
 * Detached NF-e CFOP selector E2E tests.
 *
 * Covers:
 * - Per-row search filtering (code + accent-insensitive description) with no
 *   cross-row influence and selected-option survival.
 * - Exact-code typing selecting the matching option.
 * - Farm CFOP registration from the emission page and from the rascunho
 *   editor: modal round trip, auto-selection in the originating row, option
 *   refresh with farm ordering, catalog/farm duplicate rejection.
 * - "Mais utilizados" ordering from the farm use counters.
 * - items[i].cfop submission through the preview hidden-field round trip.
 *
 * SEFAZ isolation: only SEFAZ-free paths (page render, selector, preview) are
 * exercised; emission is never confirmed and no nfe_farm_config row is
 * created or modified here.
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

/**
 * Navigate to a path on the same origin the session cookie was set on (see
 * detached_nfe_rascunho.spec.js for the localhost/127.0.0.1 rationale).
 */
async function gotoApp(page, path) {
  const origin = new URL(page.url()).origin;
  return page.goto(`${origin}${path}`);
}

/** Random valid farm CFOP in the exterior family (7xxx), unused elsewhere. */
function randomCfop() {
  const tail = String(Math.floor(Math.random() * 1000)).padStart(3, '0');
  return `7${tail}`;
}

function cleanupCfop(code) {
  db(`DELETE FROM farm_cfop_use WHERE farm_id = 1 AND code = '${code}'`);
  db(`DELETE FROM farm_cfop WHERE farm_id = 1 AND code = '${code}'`);
}

function optionHidden(row, code) {
  return row.locator(`option[value="${code}"]`).evaluate((el) => el.hidden);
}

async function fillInlineRecipient(page) {
  await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
  await page.fill('#recipientName', 'Cliente CFOP E2E');
  await page.fill('#recipientDocument', faker.string.numeric(14));
  await page.fill('#recipientStreet', 'Rua CFOP');
  await page.fill('#recipientNumber', '10');
  await page.fill('#recipientNeighborhood', 'Centro');
  await page.fill('#recipientCity', 'Cuiabá');
  await page.fill('#recipientState', 'MT');
  await page.fill('#recipientCEP', '78000000');
}

test.describe('CFOP avulso', () => {
  test('seletor de CFOP: busca por linha e contrato do formulário', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await expect(page.locator('#detached-nfe-form')).toBeVisible();

    const rows = page.locator('#items-container .item-row');
    await expect(rows).toHaveCount(1);

    // Default CFOP (fixture config 5101) is rendered and selected.
    const row0Select = rows.nth(0).locator('select[name$=".cfop"]');
    await expect(row0Select).toHaveValue('5101');
    await expect(row0Select.locator('option[value="5101"]').first()).toHaveAttribute('selected', '');

    // Add a second item: it starts from the farm/default CFOP as well.
    await page.locator('#add-item-btn').evaluate((el) => el.click());
    await expect(rows).toHaveCount(2);
    const row1Select = rows.nth(1).locator('select[name$=".cfop"]');
    await expect(row1Select).toHaveValue('5101');

    // Row 0 search: accent-insensitive description match. 1202 ("Devolução…")
    // stays; 1102 ("Compra…") is filtered out; the selected option survives.
    const row0Search = rows.nth(0).locator('.cfop-search');
    await row0Search.fill('devolucao');
    expect(await optionHidden(rows.nth(0), '1202')).toBe(false);
    expect(await optionHidden(rows.nth(0), '1102')).toBe(true);
    expect(await optionHidden(rows.nth(0), '5101')).toBe(false);

    // Rows do not influence each other.
    expect(await optionHidden(rows.nth(1), '1102')).toBe(false);

    // Clearing the search restores the full list.
    await row0Search.fill('');
    expect(await optionHidden(rows.nth(0), '1102')).toBe(false);

    // Typing the exact code selects the matching option (keyboard selection).
    await rows.nth(1).locator('.cfop-search').fill('6102');
    await expect(row1Select).toHaveValue('6102');

    // Fill the emission form and preview: per-item CFOPs must round-trip.
    await rows.nth(0).locator('input[name$=".productName"]').fill('Soja CFOP E2E');
    await rows.nth(0).locator('input[name$=".quantity"]').fill('10');
    await rows.nth(0).locator('input[name$=".unitPrice"]').fill('100.00');
    await expect(row0Select).toHaveValue('5101');

    await rows.nth(1).locator('input[name$=".productName"]').fill('Milho CFOP E2E');
    await rows.nth(1).locator('input[name$=".quantity"]').fill('5');
    await rows.nth(1).locator('input[name$=".unitPrice"]').fill('50.00');
    await fillInlineRecipient(page);

    const previewResponse = page.waitForResponse(
      (r) => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
    );
    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate((el) => el.click());
    expect((await previewResponse).status()).toBe(200);
    await expect(page.locator('#nfe-detached-preview-panel')).toBeVisible();
    await expect(
      page.locator('#nfe-detached-preview-panel input[name="items[0].cfop"]')
    ).toHaveValue('5101');
    await expect(
      page.locator('#nfe-detached-preview-panel input[name="items[1].cfop"]')
    ).toHaveValue('6102');
  });

  test('cadastro de CFOP pela emissão: autoseleção, re-render e duplicados', async ({ page, browserName }) => {
    await login(page, browserName);
    const code = randomCfop();

    try {
      await gotoApp(page, '/nfe/emitir');
      await page.locator('#add-item-btn').evaluate((el) => el.click());

      const rows = page.locator('#items-container .item-row');
      await expect(rows).toHaveCount(2);
      await rows.nth(0).locator('select[name$=".cfop"]').selectOption('5102');

      // "+" on the second row opens the register modal.
      await rows.nth(1).locator('[data-test-id="cfop-add"]').evaluate((el) => el.click());
      const dialog = page.locator('dialog#nfeCfopFormDialog');
      await expect(dialog).toBeVisible();

      await dialog.locator('[data-test-id="cfop-form-code"]').fill(code);
      // Client-side convenience: Origem/Destino is auto-derived and editable.
      await expect(dialog.locator('[data-test-id="cfop-form-origin"]')).toHaveValue('Exterior');
      await dialog.locator('[data-test-id="cfop-form-description"]').fill('Operação E2E');

      const createResponse = page.waitForResponse(
        (r) => r.url().includes('/nfe/cfop') && r.request().method() === 'POST'
      );
      await dialog.locator('[data-test-id="cfop-form-submit"]').evaluate((el) => el.click());
      expect((await createResponse).status()).toBe(201);
      await expect(dialog).toBeHidden();

      // Originating row auto-selects the new code; the other row keeps its own.
      await expect(rows.nth(1).locator('select[name$=".cfop"]')).toHaveValue(code);
      await expect(rows.nth(0).locator('select[name$=".cfop"]')).toHaveValue('5102');
      expect(
        db(`SELECT origin_destination FROM farm_cfop WHERE farm_id = 1 AND code = '${code}'`)
      ).toBe('Exterior');
      expect(db(`SELECT COUNT(*) FROM farm_cfop WHERE farm_id = 1 AND code = '${code}'`)).toBe('1');

      // Fresh render: the farm CFOP appears in "Todos os CFOPs".
      await gotoApp(page, '/nfe/emitir');
      const freshSelect = page.locator('#items-container .item-row').first().locator('select[name$=".cfop"]');
      const groupLabel = await freshSelect
        .locator(`option[value="${code}"]`)
        .evaluate((el) => (el.closest('optgroup') ? el.closest('optgroup').label : ''));
      expect(groupLabel).toBe('Todos os CFOPs');
      expect(await freshSelect.locator(`option[value="${code}"]`).getAttribute('data-cfop-farm')).toBe('true');

      // Catalog duplicate is rejected and the modal stays open.
      await page.locator('#items-container .item-row').first()
        .locator('[data-test-id="cfop-add"]').evaluate((el) => el.click());
      const catalogDialog = page.locator('dialog#nfeCfopFormDialog');
      await expect(catalogDialog).toBeVisible();
      await catalogDialog.locator('[data-test-id="cfop-form-code"]').fill('5101');
      await expect(catalogDialog.locator('[data-test-id="cfop-form-origin"]')).toHaveValue('Mesmo estado');
      await catalogDialog.locator('[data-test-id="cfop-form-description"]').fill('Duplicado do catálogo');
      const catalogResponse = page.waitForResponse(
        (r) => r.url().includes('/nfe/cfop') && r.request().method() === 'POST'
      );
      await catalogDialog.locator('[data-test-id="cfop-form-submit"]').evaluate((el) => el.click());
      expect((await catalogResponse).status()).toBe(400);
      await expect(page.locator('#armazenda-toast', { hasText: 'CFOP já cadastrado' }).first()).toBeVisible();
      await expect(catalogDialog).toBeVisible();
      await catalogDialog.locator('.cancel-btn').evaluate((el) => el.click());
      await expect(catalogDialog).toBeHidden();

      // Same-farm duplicate is rejected too.
      await page.locator('#items-container .item-row').first()
        .locator('[data-test-id="cfop-add"]').evaluate((el) => el.click());
      const farmDialog = page.locator('dialog#nfeCfopFormDialog');
      await expect(farmDialog).toBeVisible();
      await farmDialog.locator('[data-test-id="cfop-form-code"]').fill(code);
      await farmDialog.locator('[data-test-id="cfop-form-description"]').fill('Duplicado da fazenda');
      const farmResponse = page.waitForResponse(
        (r) => r.url().includes('/nfe/cfop') && r.request().method() === 'POST'
      );
      await farmDialog.locator('[data-test-id="cfop-form-submit"]').evaluate((el) => el.click());
      expect((await farmResponse).status()).toBe(400);
      await expect(page.locator('#armazenda-toast', { hasText: 'CFOP já cadastrado' }).first()).toBeVisible();
      await expect(farmDialog).toBeVisible();
      await farmDialog.locator('.cancel-btn').evaluate((el) => el.click());
      await expect(farmDialog).toBeHidden();
    } finally {
      cleanupCfop(code);
    }
  });

  test('cadastro de CFOP pelo editor de rascunho', async ({ page, browserName }) => {
    await login(page, browserName);
    const code = randomCfop();

    try {
      await gotoApp(page, '/nfe/rascunhos');
      await page.locator('[data-test-id="novo-rascunho"]').evaluate((el) => el.click());
      const editor = page.locator('dialog#addNfeRascunhoDialog');
      await expect(editor).toBeVisible();

      const row = editor.locator('.rascunho-item-row').first();
      await expect(row.locator('select[name$=".cfop"]')).toHaveValue('5101');

      await row.locator('[data-test-id="cfop-add"]').evaluate((el) => el.click());
      const dialog = page.locator('dialog#nfeCfopFormDialog');
      await expect(dialog).toBeVisible();
      await dialog.locator('[data-test-id="cfop-form-code"]').fill(code);
      await dialog.locator('[data-test-id="cfop-form-description"]').fill('Operação rascunho E2E');

      const createResponse = page.waitForResponse(
        (r) => r.url().includes('/nfe/cfop') && r.request().method() === 'POST'
      );
      await dialog.locator('[data-test-id="cfop-form-submit"]').evaluate((el) => el.click());
      expect((await createResponse).status()).toBe(201);
      await expect(dialog).toBeHidden();

      // The editor row auto-selects the freshly created farm CFOP.
      await expect(row.locator('select[name$=".cfop"]')).toHaveValue(code);
      expect(
        db(`SELECT origin_destination FROM farm_cfop WHERE farm_id = 1 AND code = '${code}'`)
      ).toBe('Exterior');

      await editor.locator('.cancel-btn').evaluate((el) => el.click());
      await expect(editor).toBeHidden();
    } finally {
      cleanupCfop(code);
    }
  });

  test('CFOP: grupo Mais utilizados ordena pelos contadores da fazenda', async ({ page, browserName }) => {
    // Counter mutations are shared by the whole farm; keep them in one
    // browser project to avoid cross-project races.
    test.skip(browserName !== 'chromium', 'Contadores são globais por fazenda');

    await login(page, browserName);
    db(
      `INSERT INTO farm_cfop_use (farm_id, code, use_count) VALUES (1, '6102', 5)
       ON CONFLICT (farm_id, code) DO UPDATE SET use_count = 5`
    );
    db(
      `INSERT INTO farm_cfop_use (farm_id, code, use_count) VALUES (1, '5101', 2)
       ON CONFLICT (farm_id, code) DO UPDATE SET use_count = 2`
    );

    try {
      await gotoApp(page, '/nfe/emitir');
      const select = page.locator('#items-container .item-row').first().locator('select[name$=".cfop"]');
      const group = select.locator('optgroup[label="Mais utilizados"]');
      await expect(group).toHaveCount(1);

      const codes = await group.locator('option').evaluateAll((options) => options.map((o) => o.value));
      expect(codes[0]).toBe('6102'); // count 5
      expect(codes[1]).toBe('5101'); // count 2

      // The configured default keeps being selected even when ranked second.
      await expect(select).toHaveValue('5101');
    } finally {
      db(`DELETE FROM farm_cfop_use WHERE farm_id = 1 AND code IN ('6102', '5101')`);
    }
  });
});
