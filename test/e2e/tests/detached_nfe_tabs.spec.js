/**
 * Detached NF-e form tabs (assets/js/detachedNFeTabs.js) E2E tests.
 *
 * Covers:
 * - Only the Destinatário section is visible on load; every tab reveals its
 *   own fieldset and hides the previous one.
 * - The preview submit button is disabled on load, stays disabled while the
 *   Destinatário/Transação/Transporte tabs are active, and is only enabled
 *   by the Tributação tab (switching away disables it again).
 */

import { test, expect } from '@playwright/test';
import { login } from '../utils/auth';

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
const EMIT_TABS = [
  { tab: 'dest', fieldset: 'fieldset#dest' },
  { tab: 'transac', fieldset: 'fieldset#transac' },
  { tab: 'transp', fieldset: 'fieldset#transporte-fieldset' },
  { tab: 'trib', fieldset: 'fieldset#trib' },
];

test.describe('Tabs of the detached NF-e form', () => {
  test('each tab reveals exactly its own section', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    await expect(page.locator('#detached-nfe-form')).toBeVisible();

    // Only Destinatário is visible on load.
    await expect(page.locator('fieldset#dest')).toBeVisible();
    for (const { fieldset } of EMIT_TABS.slice(1)) {
      await expect(page.locator(fieldset)).toBeHidden();
    }

    for (const { tab, fieldset } of EMIT_TABS) {
      await page.locator(`#tabs button#${tab}`).click();
      await expect(page.locator(fieldset)).toBeVisible();
      for (const { fieldset: other } of EMIT_TABS) {
        if (other !== fieldset) {
          await expect(page.locator(other)).toBeHidden();
        }
      }
    }
  });

  test('the preview submit is only enabled on the Tributação tab', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    const preview = page.locator('[data-test-id="preview-detached-nfe"]');
    await expect(preview).toBeDisabled();

    for (const { tab } of EMIT_TABS.slice(0, 3)) {
      await page.locator(`#tabs button#${tab}`).click();
      await expect(preview).toBeDisabled();
    }

    await page.locator('#tabs button#trib').click();
    await expect(preview).toBeEnabled();

    // Leaving Tributação disables the submit again.
    await page.locator('#tabs button#dest').click();
    await expect(preview).toBeDisabled();
  });
});
