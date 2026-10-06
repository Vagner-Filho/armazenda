/**
 * Rascunhos de NF-e (detached NF-e operation profiles) E2E tests.
 *
 * Covers:
 * - Rascunho lifecycle: create, list, apply (navigation), edit, delete
 * - Data round trip: recipient, ordered items, distinct CFOPs, exact prices
 * - Save-from-form with an explicit inline recipient contact
 * - Preview remains write-free (no person created)
 * - Farm isolation for rascunho read/apply/edit/delete
 * - 'draft' status displays as "Não enviada" on the NF-e list
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

// Status-label fixture rows are marked with a sentinel motive so concurrent
// tests can exclude them from "no invoice was created" side-effect counts.
const STATUS_FIXTURE_MOTIVE = 'e2e-status-fixture';

function detachedInvoiceCount() {
  return db(
    `SELECT COUNT(*) FROM detached_nfe_invoice WHERE farm_id = 1 AND COALESCE(sefaz_motive, '') <> '${STATUS_FIXTURE_MOTIVE}'`
  );
}

/**
 * Navigate to a path on the same origin the session cookie was set on. The
 * login helper lands on 127.0.0.1, while page.goto('/...') would resolve
 * against the Playwright baseURL (localhost), dropping the session cookie.
 */
async function gotoApp(page, path) {
  const origin = new URL(page.url()).origin;
  return page.goto(`${origin}${path}`);
}

async function openNewRascunhoDialog(page) {
  // evaluate-click keeps WebKit happy with htmx/backdrop stacking; the target
  // is a plain toolbar button, not a form control.
  await page.locator('[data-test-id="novo-rascunho"]').evaluate(el => el.click());
  const dialog = page.locator('dialog#addNfeRascunhoDialog');
  await expect(dialog).toBeVisible();
  return dialog;
}

async function fillRascunhoItem(row, { descricao, cfop, preco, cest }) {
  if (descricao !== undefined) {
    await row.locator('[data-test-id="rascunho-item-descricao"]').fill(descricao);
  }
  await row.locator('[data-test-id="rascunho-item-cfop"]').fill(cfop);
  if (preco !== undefined) {
    await row.locator('[data-test-id="rascunho-item-preco"]').fill(preco);
  }
  if (cest !== undefined) {
    await row.locator('input[name$=".cest"]').fill(cest);
  }
}

test.describe('Rascunho de NF-e', () => {
  test('ciclo de vida: criar, listar, aplicar, editar e excluir', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/rascunhos');
    await expect(page.getByRole('heading', { name: 'Rascunhos de NF-e' })).toBeVisible();

    const name = `Rascunho E2E ${uniqueSuffix()}`;
    const productName = `Produto E2E ${uniqueSuffix()}`;
    db(`INSERT INTO farm_product (name, ncm, farm_id) VALUES ('${productName}', '99999999', 1)`);
    const productID = db(`SELECT id FROM farm_product WHERE farm_id = 1 AND name = '${productName}'`);

    // --- Create with two ordered items, distinct CFOPs and exact prices ---
    const dialog = await openNewRascunhoDialog(page);
    await dialog.locator('[data-test-id="rascunho-form-nome"]').fill(name);

    const rows = dialog.locator('.rascunho-item-row');
    await rows.nth(0).locator('select[name$=".farmProductId"]').selectOption({ label: productName });
    await fillRascunhoItem(rows.nth(0), { descricao: 'Soja E2E', cfop: '5101', preco: '150.25', cest: '0100101' });

    await dialog.locator('#add-rascunho-item-btn').evaluate(el => el.click());
    await expect(rows).toHaveCount(2);
    await fillRascunhoItem(rows.nth(1), { descricao: 'Milho E2E', cfop: '5102', preco: '120.75' });

    const numberingBefore = db('SELECT COALESCE(MAX(last_number), 0) FROM nfe_numbering WHERE farm_id = 1');
    const invoicesBefore = detachedInvoiceCount();

    const createResponse = page.waitForResponse(
      r => r.url().includes('/nfe/rascunho') && r.request().method() === 'POST'
    );
    await dialog.locator('button[type="submit"]').evaluate(el => el.click());
    const created = await createResponse;
    expect(created.ok()).toBeTruthy();
    await expect(dialog).toBeHidden();

    const row = page.locator('[data-test-id="rascunho-row"]', { hasText: name });
    await expect(row).toBeVisible();

    // DB round trip: two items, distinct CFOPs, exact prices, no invoice side effects.
    const stored = db(
      `SELECT jsonb_array_length(items_json) || '|' || (items_json->0->>'cfop') || '|' || (items_json->1->>'cfop') || '|' || (items_json->0->>'unit_price') || '|' || (items_json->1->>'unit_price') || '|' || (items_json->0->>'product_name') || '|' || (items_json->1->>'product_name') || '|' || (items_json->0->>'farm_product_id') FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${name}'`
    );
    expect(stored).toBe(`2|5101|5102|150.25|120.75|Soja E2E|Milho E2E|${productID}`);
    expect(detachedInvoiceCount()).toBe(invoicesBefore);
    expect(db('SELECT COALESCE(MAX(last_number), 0) FROM nfe_numbering WHERE farm_id = 1')).toBe(numberingBefore);

    // --- Duplicate name (case-insensitive) is rejected ---
    const dupDialog = await openNewRascunhoDialog(page);
    await dupDialog.locator('[data-test-id="rascunho-form-nome"]').fill(name.toUpperCase());
    await fillRascunhoItem(dupDialog.locator('.rascunho-item-row').nth(0), { cfop: '5101', preco: '10.00' });
    const dupResponse = page.waitForResponse(
      r => r.url().includes('/nfe/rascunho') && r.request().method() === 'POST'
    );
    await dupDialog.locator('button[type="submit"]').evaluate(el => el.click());
    await dupResponse;
    await expect(
      page.locator('#armazenda-toast', { hasText: 'Nome de rascunho já existe' }).first()
    ).toBeVisible();
    await expect(dupDialog).toBeVisible();
    await dupDialog.locator('.cancel-btn').evaluate(el => el.click());
    await expect(dupDialog).toBeHidden();

    // --- Apply navigates to the Emitir page with prefilled, editable values ---
    await row.locator('[data-test-id="gerar-nfe-rascunho"]').evaluate(el => el.click());
    await page.waitForURL(/\/nfe\/emitir\/rascunho\/\d+/);
    await expect(page.locator('[data-test-id="rascunho-aplicado"]')).toContainText(name);

    const itemRows = page.locator('#items-container .item-row');
    await expect(itemRows).toHaveCount(2);
    await expect(itemRows.nth(0).locator('input[name$=".cfop"]')).toHaveValue('5101');
    await expect(itemRows.nth(1).locator('input[name$=".cfop"]')).toHaveValue('5102');
    await expect(itemRows.nth(0).locator('input[name$=".unitPrice"]')).toHaveValue('150.25');
    await expect(itemRows.nth(1).locator('input[name$=".unitPrice"]')).toHaveValue('120.75');
    await expect(itemRows.nth(0).locator('input[name$=".productName"]')).toHaveValue('Soja E2E');
    // Quantity and gross weight are per-emission inputs and stay empty.
    await expect(itemRows.nth(0).locator('input[name$=".quantity"]')).toHaveValue('');
    await expect(itemRows.nth(0).locator('input[name$=".grossWeight"]')).toHaveValue('');
    await expect(itemRows.nth(1).locator('input[name$=".quantity"]')).toHaveValue('');

    // Editing the applied form does not write back to the saved rascunho.
    await itemRows.nth(0).locator('input[name$=".unitPrice"]').fill('999.00');
    expect(
      db(`SELECT items_json->0->>'unit_price' FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${name}'`)
    ).toBe('150.25');

    // --- Edit changes the saved rascunho through an explicit action ---
    const editedName = `${name} Editado`;
    await gotoApp(page, '/nfe/rascunhos');
    const rowToEdit = page.locator('[data-test-id="rascunho-row"]', { hasText: name });
    await rowToEdit.locator('[data-test-id="editar-rascunho"]').evaluate(el => el.click());
    const editDialog = page.locator('dialog#addNfeRascunhoDialog');
    await expect(editDialog).toBeVisible();
    await expect(editDialog.locator('[data-test-id="rascunho-form-nome"]')).toHaveValue(name);
    await expect(editDialog.locator('[data-test-id="rascunho-item-cfop"]').nth(1)).toHaveValue('5102');

    const editRows = editDialog.locator('.rascunho-item-row');
    await editRows.nth(0).locator('[data-test-id="rascunho-item-preco"]').fill('151.00');
    await editDialog.locator('[data-test-id="rascunho-form-nome"]').fill(editedName);

    const updateResponse = page.waitForResponse(
      r => /\/nfe\/rascunho\/\d+/.test(r.url()) && r.request().method() === 'PUT'
    );
    await editDialog.locator('button[type="submit"]').evaluate(el => el.click());
    const updated = await updateResponse;
    expect(updated.ok()).toBeTruthy();
    await expect(editDialog).toBeHidden();
    await expect(page.locator('[data-test-id="rascunho-row"]', { hasText: editedName })).toBeVisible();
    expect(
      db(`SELECT items_json->0->>'unit_price' FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${editedName}'`)
    ).toBe('151'); // decimal String() trims trailing zeros

    // --- Delete asks for confirmation and hard-deletes ---
    const rowToDelete = page.locator('[data-test-id="rascunho-row"]', { hasText: editedName });
    await rowToDelete.locator('[data-test-id="excluir-rascunho"]').evaluate(el => el.click());
    await expect(page.locator('.swal2-confirm')).toBeVisible();
    const deleteResponse = page.waitForResponse(
      r => /\/nfe\/rascunho\/\d+/.test(r.url()) && r.request().method() === 'DELETE'
    );
    await page.locator('.swal2-confirm').evaluate(el => el.click());
    await deleteResponse;
    await expect(rowToDelete).toHaveCount(0);
    expect(db(`SELECT COUNT(*) FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${editedName}'`)).toBe('0');
    expect(detachedInvoiceCount()).toBe(invoicesBefore);
    expect(db('SELECT COALESCE(MAX(last_number), 0) FROM nfe_numbering WHERE farm_id = 1')).toBe(numberingBefore);

    db(`DELETE FROM farm_product WHERE farm_id = 1 AND name = '${productName}'`);
  });

  test('salvar do formulário cria contato inline e o preview não cria contato', async ({ page, browserName }) => {
    await login(page, browserName);

    const name = `Rascunho Formulário ${uniqueSuffix()}`;
    const document = faker.string.numeric(14);

    await gotoApp(page, '/nfe/emitir');
    await expect(page.locator('#detached-nfe-form')).toBeVisible();

    await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
    await page.fill('#recipientName', 'Cliente Inline E2E');
    await page.fill('#recipientDocument', document);
    await page.fill('#recipientStreet', 'Rua Um');
    await page.fill('#recipientNumber', '10');
    await page.fill('#recipientNeighborhood', 'Centro');
    await page.fill('#recipientCity', 'Cuiabá');
    await page.fill('#recipientState', 'MT');
    await page.fill('#recipientCEP', '78000000');

    await page.fill('input[name="items[0].productName"]', 'Soja Formulário');
    await page.fill('input[name="items[0].cfop"]', '5101');
    await page.fill('input[name="items[0].quantity"]', '10');
    await page.fill('input[name="items[0].unitPrice"]', '99.50');
    await page.fill('#naturezaOp', 'Venda de teste');

    const numberingBefore = db('SELECT COALESCE(MAX(last_number), 0) FROM nfe_numbering WHERE farm_id = 1');
    const invoicesBefore = detachedInvoiceCount();

    // Save the populated form as a rascunho: the inline recipient is
    // explicitly saved as a farm contact and linked to the rascunho.
    await page.locator('[data-test-id="open-save-rascunho"]').evaluate(el => el.click());
    const saveDialog = page.locator('dialog#save-rascunho-dialog');
    await expect(saveDialog).toBeVisible();
    await saveDialog.locator('input[name="rascunhoNome"]').fill(name);

    const saveResponse = page.waitForResponse(
      r => r.url().includes('/nfe/rascunho/salvar') && r.request().method() === 'POST'
    );
    await saveDialog.locator('[data-test-id="confirm-save-rascunho"]').evaluate(el => el.click());
    const saved = await saveResponse;
    expect(saved.status()).toBe(204);
    await expect(saveDialog).toBeHidden();

    const profileInfo = db(
      `SELECT COALESCE(recipient_id::text, '') || '|' || (items_json->0->>'cfop') || '|' || (items_json->0->>'unit_price') || '|' || (items_json->0->>'product_name') FROM detached_nfe_profile WHERE farm_id = 1 AND name = '${name}'`
    );
    const [recipientID, cfop, price, productName] = profileInfo.split('|');
    expect(recipientID).not.toBe('');
    expect(cfop).toBe('5101');
    expect(price).toBe('99.5');
    expect(productName).toBe('Soja Formulário');
    expect(db(`SELECT COUNT(*) FROM legal_person WHERE cnpj = '${document}'`)).toBe('1');
    expect(db(`SELECT farm FROM person WHERE id = ${recipientID}`)).toBe('1');

    // Saving a rascunho does not allocate numbers or create invoices.
    expect(detachedInvoiceCount()).toBe(invoicesBefore);
    expect(db('SELECT COALESCE(MAX(last_number), 0) FROM nfe_numbering WHERE farm_id = 1')).toBe(numberingBefore);

    // A plain preview of an inline recipient creates no person and persists nothing.
    const dryDocument = faker.string.numeric(14);
    await gotoApp(page, '/nfe/emitir');
    await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
    await page.fill('#recipientName', 'Cliente Preview E2E');
    await page.fill('#recipientDocument', dryDocument);
    await page.fill('#recipientStreet', 'Rua Dois');
    await page.fill('#recipientNumber', '20');
    await page.fill('#recipientNeighborhood', 'Centro');
    await page.fill('#recipientCity', 'Cuiabá');
    await page.fill('#recipientState', 'MT');
    await page.fill('#recipientCEP', '78000000');
    await page.fill('input[name="items[0].productName"]', 'Soja Preview');
    await page.fill('input[name="items[0].cfop"]', '5101');
    await page.fill('input[name="items[0].quantity"]', '5');
    await page.fill('input[name="items[0].unitPrice"]', '88.00');

    const previewResponse = page.waitForResponse(
      r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
    );
    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate(el => el.click());
    const preview = await previewResponse;
    expect(preview.status()).toBe(200);
    await expect(page.locator('#nfe-detached-preview-panel')).toBeVisible();

    expect(db(`SELECT COUNT(*) FROM legal_person WHERE cnpj = '${dryDocument}'`)).toBe('0');
    expect(detachedInvoiceCount()).toBe(invoicesBefore);
    expect(db('SELECT COALESCE(MAX(last_number), 0) FROM nfe_numbering WHERE farm_id = 1')).toBe(numberingBefore);

    // The preview hidden-field round trip carries the item CFOP so the
    // confirm step submits exactly what was previewed.
    await expect(
      page.locator('#nfe-detached-preview-panel input[name="items[0].cfop"]')
    ).toHaveValue('5101');
  });

  test('rascunho de outra fazenda não pode ser lido, aplicado, editado nem excluído', async ({ page, browserName }) => {
    await login(page, browserName);

    const suffix = uniqueSuffix();
    const inscricao = `9${suffix}`.slice(0, 12).padEnd(12, '0');
    db(`INSERT INTO farm (inscricao_estadual, uf) VALUES ('${inscricao}', 'MT') ON CONFLICT (inscricao_estadual) DO NOTHING`);
    const otherFarm = db(`SELECT id FROM farm WHERE inscricao_estadual = '${inscricao}'`);

    db(
      `INSERT INTO detached_nfe_profile (farm_id, name, items_json) VALUES (${otherFarm}, 'Rascunho de Outra Fazenda ${suffix}', '[{"product_name":"Soja","ncm":"12019000","cfop":"5101","unit":"KG","unit_price":"1.00"}]'::jsonb)`
    );
    const otherProfile = db(
      `SELECT id FROM detached_nfe_profile WHERE farm_id = ${otherFarm} AND name = 'Rascunho de Outra Fazenda ${suffix}'`
    );

    // Read/apply: render the blank emission page with a warning.
    await gotoApp(page, `/nfe/emitir/rascunho/${otherProfile}`);
    await expect(page.locator('[data-test-id="rascunho-warning"]')).toContainText('Rascunho não encontrado');
    await expect(page.locator('#items-container .item-row').first().locator('input[name$=".productName"]')).toHaveValue('');
    // No prefilled rascunho banner and no leaked values.
    await expect(page.locator('[data-test-id="rascunho-aplicado"]')).toHaveCount(0);

    // Edit form: not found. Uses same-origin fetch so the browser session
    // cookie is sent and a body-less 404 cannot surface as a navigation error.
    const editStatus = await page.evaluate(async ({ id }) => {
      const response = await fetch(`/nfe/rascunho/form/${id}`);
      return response.status;
    }, { id: Number(otherProfile) });
    expect(editStatus).toBe(404);

    // Update: rejected.
    const updateStatus = await page.evaluate(async ({ id }) => {
      const body = new URLSearchParams({ name: 'Invadido', itemCount: '1', 'items[0].cfop': '5101', 'items[0].unitPrice': '1.00' });
      const response = await fetch(`/nfe/rascunho/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body,
      });
      return response.status;
    }, { id: Number(otherProfile) });
    expect(updateStatus).toBeGreaterThanOrEqual(400);

    // Delete: rejected, row remains.
    const deleteStatus = await page.evaluate(async ({ id }) => {
      const response = await fetch(`/nfe/rascunho/${id}`, { method: 'DELETE' });
      return response.status;
    }, { id: Number(otherProfile) });
    expect(deleteStatus).toBeGreaterThanOrEqual(400);
    expect(
      db(`SELECT COUNT(*) FROM detached_nfe_profile WHERE id = ${otherProfile}`)
    ).toBe('1');

    db(`DELETE FROM detached_nfe_profile WHERE id = ${otherProfile}`);
    db(`DELETE FROM farm WHERE id = ${otherFarm}`);
  });

  test('navegar com o formulário editado pede confirmação', async ({ page, browserName }) => {
    test.skip(browserName !== 'chromium', 'Somente o Chromium expõe o diálogo de beforeunload nos testes');

    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await expect(page.locator('#detached-nfe-form')).toBeVisible();
    await page.fill('input[name="items[0].productName"]', 'Edição não salva');

    let dialogShown = false;
    page.on('dialog', async (dialog) => {
      dialogShown = true;
      // Accepting proceeds with the navigation (the guard asked for
      // confirmation); dismissing would abort it.
      await dialog.accept();
    });

    await page.click('a[href="/romaneio"]');
    expect(dialogShown).toBeTruthy();
  });

  test('status draft é exibido como "Não enviada" na lista de NF-e', async ({ page, browserName }) => {
    await login(page, browserName);

    const suffix = uniqueSuffix();
    const accessKey = `9${suffix}`.padEnd(44, '0').slice(0, 44);
    const invoiceNumber = Math.floor(Math.random() * 1_000_000) + 1;
    const cnpj = faker.string.numeric(14);

    db(`INSERT INTO person (ie, farm) VALUES ('', 1)`);
    const personId = db('SELECT id FROM person WHERE farm = 1 ORDER BY id DESC LIMIT 1');
    db(
      `INSERT INTO legal_person (cnpj, personid, companyname) VALUES ('${cnpj}', ${personId}, 'Destino E2E Status')`
    );
    db(
      `INSERT INTO detached_nfe_invoice (farm_id, recipient_id, access_key, serie, number, status, natureza_op, mod_frete, total_value, items_json, sefaz_motive) VALUES (1, ${personId}, '${accessKey}', 1, ${invoiceNumber}, 'draft', 'Venda', 9, 100.00, '[{"product_name":"Soja","ncm":"12019000","cfop":"5101","unit":"KG","quantity":1,"gross_weight":1,"unit_price":100,"total_value":100}]'::jsonb, '${STATUS_FIXTURE_MOTIVE}')`
    );

    try {
      await gotoApp(page, '/nfe/list');
      const row = page.locator(`#nfe-${accessKey}`);
      await expect(row).toBeVisible();
      const status = row.locator('td[data-label="Status"]');
      await expect(status).toContainText('Não enviada');
      await expect(status).not.toContainText('Rascunho');
    } finally {
      db(`DELETE FROM detached_nfe_invoice WHERE access_key = '${accessKey}'`);
      db(`DELETE FROM legal_person WHERE personid = ${personId}`);
      db(`DELETE FROM person WHERE id = ${personId}`);
    }
  });
});
