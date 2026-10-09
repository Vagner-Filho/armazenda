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

// Selects a CFOP through the search-driven selector used by the item rows.
async function selectCfop(row, code) {
  await row.locator('[data-cfop-trigger]').click();
  const search = row.locator('.cfop-search');
  await expect(search).toBeVisible();
  await search.fill(code);
  await search.press('Enter');
  await expect(row.locator('select[name$=".cfop"]')).toHaveValue(code);
}

async function fillBasicForm(page) {
  await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
  await page.fill('#recipientName', 'Destinatário Transporte E2E');
  await page.fill('#recipientDocument', faker.string.numeric(14));
  await page.fill('#recipientStreet', 'Rua Transporte');
  await page.fill('#recipientNumber', '30');
  await page.fill('#recipientNeighborhood', 'Centro');
  await page.fill('#recipientCity', 'Cuiabá');
  await page.fill('#recipientState', 'MT');
  await page.fill('#recipientCEP', '78000000');
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
    await expect(page.locator('#transporte-fieldset')).toBeVisible();

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
    await expect(page.locator('#transporte-fieldset')).toBeVisible();

    await fillBasicForm(page);

    // Prefill stays live while fields are untouched.
    await expect(page.locator('#volQtd')).toHaveValue('1');
    await expect(page.locator('#volEsp')).toHaveValue('Granel');
    await expect(page.locator('#volPesoL')).toHaveValue('10');
    await expect(page.locator('#volPesoB')).toHaveValue('10.2');

    // A user edit stops that field's auto-update; others keep mirroring.
    await page.fill('#volPesoL', '999');

    // Add a second item: quantity sum changes from 10 to 30.
    await page.locator('#add-item-btn').click();
    await page.fill('input[name="items[1].productName"]', 'Milho Transporte');
    await selectCfop(page.locator('#items-container .item-row').nth(1), '5102');
    await page.fill('input[name="items[1].quantity"]', '20');
    await page.fill('input[name="items[1].unitPrice"]', '30.00');

    await expect(page.locator('#volPesoL')).toHaveValue('999'); // user-edited
    await expect(page.locator('#volPesoB')).toHaveValue('10.2'); // still mirrored (unchanged sums)
    await expect(page.locator('#volQtd')).toHaveValue('1');
  });

  test('preview round-trips the transport values and the DB stores nothing', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await expect(page.locator('#transporte-fieldset')).toBeVisible();

    await fillBasicForm(page);

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

    const previewResponse = page.waitForResponse(
      r => r.url().includes('/nfe/emitir/preview') && r.request().method() === 'POST'
    );
    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate(el => el.click());
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
    await expect(page.locator('#transporte-fieldset')).toBeVisible();

    await fillBasicForm(page);
    await page.fill('#recipientState', 'SP'); // interstate: SP destination, MT emitter

    await page.fill('#transportPlaca', `${faker.string.alpha({ length: 3 }).toUpperCase()}${faker.string.numeric(4)}`);
    await page.fill('#volPesoL', '999'); // user edit, keeps prefill permission off
    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate(el => el.click());

    // Rejected before any transmission: no preview panel, warning toast.
    await expect(page.locator('#nfe-detached-preview-panel')).toHaveCount(0);
  });
});
