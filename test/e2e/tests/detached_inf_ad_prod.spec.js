/**
 * Detached NF-e per-item additional information (det/infAdProd) E2E tests.
 *
 * Covers:
 * - The item row renders the optional "Informações adicionais do item"
 *   input with the 500-character client-side cap
 * - Preview round-trip: the typed value reaches the preview fragment's
 *   hidden items[i].infAdProd input (the confirm step re-submits exactly the
 *   previewed text); a row left blank round-trips blank
 * - Adding a row clones/reindexes the new input (name follows items[1])
 */

import { test, expect } from '@playwright/test';
import { faker } from '@faker-js/faker';
import { login } from '../utils/auth';

// Navigate to a path on the same origin the session cookie was set on. The
// login helper lands on 127.0.0.1, while page.goto('/...') would resolve
// against the Playwright baseURL (localhost), dropping the session cookie.
async function gotoApp(page, path) {
  const origin = new URL(page.url()).origin;
  return page.goto(`${origin}${path}`);
}

// Sections of the emission form are hidden until their tab in #tabs is
// clicked.
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

async function fillRecipient(page) {
  await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
  await page.fill('#recipientName', 'Destinatário Info Item E2E');
  await page.fill('#recipientDocument', faker.string.numeric(14));
  await page.fill('#recipientStreet', 'Rua Info Item');
  await page.fill('#recipientNumber', '10');
  await page.fill('#recipientNeighborhood', 'Centro');
  await page.fill('#recipientCity', 'Cuiabá');
  await page.fill('#recipientState', 'MT');
  await page.fill('#recipientCEP', '78000000');
}

test.describe('Per-item additional information of the detached NF-e form', () => {
  test('round-trips the item note through the preview hidden fields', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    await fillRecipient(page);

    // Transação tab: the item fields are hidden until their tab is active.
    await goToTab(page, 'transac');
    await page.fill('input[name="items[0].productName"]', 'Soja Info Item');
    await selectCfop(page.locator('#items-container .item-row').first(), '5101');
    await page.fill('input[name="items[0].quantity"]', '10');
    await page.fill('input[name="items[0].grossWeight"]', '10.2');
    await page.fill('input[name="items[0].unitPrice"]', '50.00');

    // The optional per-item note renders with the client-side length cap.
    const infoInput = page.locator('[data-test-id="item-inf-ad-prod-0"]');
    await expect(infoInput).toBeVisible();
    await expect(infoInput).toHaveAttribute('maxlength', '500');
    await infoInput.fill('SERIE ABC123');

    // A cloned row keeps the input (reindexed to items[1]) and starts blank.
    // The clone source's data-test-id is not reindexed, so select by name.
    await page.locator('#add-item-btn').click();
    const secondRow = page.locator('#items-container .item-row').nth(1);
    await page.fill('input[name="items[1].productName"]', 'Milho Info Item');
    await selectCfop(secondRow, '5102');
    await page.fill('input[name="items[1].quantity"]', '5');
    await page.fill('input[name="items[1].unitPrice"]', '30.00');
    await expect(secondRow.locator('input[name="items[1].infAdProd"]')).toHaveValue('');

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

    // Hidden round-trip fields carry the exact previewed values, scoped to
    // the preview panel (the main form keeps the same names).
    const previewPanel = page.locator('#nfe-detached-preview-panel');
    await expect(previewPanel.locator('input[name="items[0].infAdProd"]')).toHaveValue('SERIE ABC123');
    await expect(previewPanel.locator('input[name="items[1].infAdProd"]')).toHaveValue('');
  });
});
