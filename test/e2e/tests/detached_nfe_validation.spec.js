/**
 * Detached NF-e emission form: stepper validation E2E tests.
 *
 * Covers (docs/sdd/detached_nfe_tab_validation):
 * - `Avançar` validates only the current step and reports it with the
 *   browser's native validation bubble on a visible field; `Voltar` never
 *   validates.
 * - The live step status badge (⚠ n invalid / ✓ complete).
 * - The preview submit only exists on the last step (hidden and disabled
 *   elsewhere, so Enter stays a no-op on the other steps).
 * - The `invalid` backstop: when a required field of a step that is no longer
 *   active is invalid, submitting reveals that step, focuses the field and
 *   sends no request (the old behaviour aborted the submission silently).
 * - The rascunho save reaches the server with an incomplete emission form
 *   (htmx used to halt it silently on the hx-include'd form).
 */

import { test, expect } from '@playwright/test';
import { faker } from '@faker-js/faker';
import { execSync } from 'child_process';
import { login } from '../utils/auth';

const DB = {
  host: 'localhost',
  port: '5433',
  user: 'test',
  password: 'test',
  database: 'armazenda_test',
};

/** Runs a query against the e2e database and returns its first output line. */
function db(sql) {
  return execSync(
    `PGPASSWORD=${DB.password} psql -v ON_ERROR_STOP=1 -h ${DB.host} -p ${DB.port} -U ${DB.user} -d ${DB.database} -tAc "${sql.replace(/"/g, '\\"')}"`,
    { encoding: 'utf8' }
  ).trim().split('\n')[0];
}

// Navigate to a path on the same origin the session cookie was set on. The
// login helper lands on 127.0.0.1, while page.goto('/...') would resolve
// against the Playwright baseURL (localhost), dropping the session cookie.
async function gotoApp(page, path) {
  const origin = new URL(page.url()).origin;
  return page.goto(`${origin}${path}`);
}

// Step fieldset per stepper chip id.
const STEP_FIELDSETS = {
  dest: 'fieldset#dest',
  transac: 'fieldset#transac',
  transp: 'fieldset#transporte-fieldset',
  trib: 'fieldset#trib',
};

/**
 * Clicks a stepper chip (free navigation, no validation). evaluate-click is the
 * pattern the other detached NF-e specs use to keep WebKit happy with
 * htmx/backdrop stacking.
 */
async function goToStep(page, tabId) {
  await page.locator(`#tabs button#${tabId}`).evaluate((el) => el.click());
  await expect(page.locator(STEP_FIELDSETS[tabId])).toBeVisible();
}

/**
 * Clicks Avançar, which validates the current step first. The footer buttons
 * carry iconify icons, whose late rendering keeps the button box moving;
 * evaluate-click is the pattern the other detached NF-e specs use for WebKit.
 */
async function advance(page) {
  await page.locator('#wizard-next').evaluate((el) => el.click());
}

/** Clicks Voltar, which never validates. */
async function goBack(page) {
  await page.locator('#wizard-back').evaluate((el) => el.click());
}

/** The status badge of a step: "⚠ n" while it has invalid controls, "✓" when complete. */
function statusBadge(page, tabId) {
  return page.locator(`[data-step-status="${tabId}"]`);
}

/** Records every request to a URL fragment so a test can assert none was sent. */
function trackRequests(page, fragment) {
  const sent = [];
  page.on('request', (request) => {
    if (request.url().includes(fragment)) sent.push(request);
  });
  return sent;
}

async function fillRecipient(page) {
  await page.locator('input[name="recipientType"][value="new"]').check({ force: true });
  await page.fill('#recipientName', 'Destinatário Wizard E2E');
  await page.fill('#recipientDocument', faker.string.numeric(14));
  // The DANFE preview needs a complete recipient address (emitter validation).
  await page.fill('#recipientStreet', 'Rua Wizard');
  await page.fill('#recipientNumber', '10');
  await page.fill('#recipientNeighborhood', 'Centro');
  await page.fill('#recipientCity', 'Cuiabá');
  await page.fill('#recipientState', 'MT');
  await page.fill('#recipientCEP', '78000000');
}

// Selects a CFOP through the search-driven selector used by the item rows.
// The trigger click is an evaluate-click: a real click on it hangs under
// WebKit parallel load (the same fragility that makes the pre-existing CFOP
// specs flaky there).
async function selectCfop(row, code) {
  await row.locator('[data-cfop-trigger]').evaluate((el) => el.click());
  const search = row.locator('.cfop-search');
  await expect(search).toBeVisible();
  await search.fill(code);
  await search.press('Enter');
  await expect(row.locator('select[name$=".cfop"]')).toHaveValue(code);
}

async function fillFirstItem(page) {
  await page.fill('input[name="items[0].productName"]', 'Soja Wizard E2E');
  await selectCfop(page.locator('#items-container .item-row').first(), '5101');
  await page.fill('input[name="items[0].quantity"]', '10');
  await page.fill('input[name="items[0].unitPrice"]', '50.00');
}

test.describe('Detached NF-e emission form stepper', () => {
  test('Avançar blocks on an invalid step and reports it natively', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    const previews = trackRequests(page, '/nfe/emitir/preview');

    // "Novo destinatário" makes the name and document required.
    await page.locator('input[name="recipientType"][value="new"]').check({ force: true });

    // The step badge reports the missing fields before anything is submitted.
    await expect(statusBadge(page, 'dest')).toContainText('⚠ 2');

    await advance(page);

    // Still on Destinatário: the native bubble is shown on a visible field.
    await expect(page.locator(STEP_FIELDSETS.dest)).toBeVisible();
    await expect(page.locator(STEP_FIELDSETS.transac)).toBeHidden();
    const name = page.locator('#recipientName');
    await expect(name).toBeFocused();
    expect(await name.evaluate((el) => el.validationMessage)).not.toBe('');
    await expect(name).toHaveClass(/nfe-field-invalid/);
    expect(previews).toHaveLength(0);

    // Completing the step lets Avançar move on.
    await page.fill('#recipientName', 'Destinatário Wizard E2E');
    await page.fill('#recipientDocument', faker.string.numeric(14));
    await expect(statusBadge(page, 'dest')).toContainText('✓');
    await advance(page);
    await expect(page.locator(STEP_FIELDSETS.transac)).toBeVisible();
  });

  test('the Destinatário step requires a picked recipient in "Selecionar existente" mode', async ({ page, browserName }) => {
    await login(page, browserName);

    // A registered contact, so the recipient select has a real option. The IE
    // is unique per farm (partial unique index), so it must not collide with
    // the contacts other specs create in parallel.
    const suffix = `${Date.now().toString(36)}`;
    const ie = String(Math.floor(100000000000 + Math.random() * 900000000000));
    const farmId = Number(db(`SELECT id FROM farm WHERE inscricao_estadual = '123456789' LIMIT 1`));
    const personID = Number(db(`INSERT INTO person (ie, farm) VALUES ('${ie}', ${farmId}) RETURNING id`));
    if (!personID) throw new Error('could not create the recipient fixture');
    db(`INSERT INTO natural_person (name, cpf, personid) VALUES ('Destinatário Existente ${suffix}', '52998224725', ${personID})`);

    try {
      await gotoApp(page, '/nfe/emitir');
      const previews = trackRequests(page, '/nfe/emitir/preview');

      // The default mode is "Selecionar existente": the person select is the
      // recipient (the server accepts a person id *or* the inline name and
      // document), and nothing is picked yet.
      await expect(statusBadge(page, 'dest')).toContainText('⚠ 1');

      await advance(page);

      await expect(page.locator(STEP_FIELDSETS.dest)).toBeVisible();
      const select = page.locator('#recipientPersonId');
      await expect(select).toBeFocused();
      expect(await select.evaluate((el) => el.validationMessage)).not.toBe('');
      expect(previews).toHaveLength(0);

      // Picking the contact completes the step.
      await select.selectOption(String(personID));
      await expect(statusBadge(page, 'dest')).toContainText('✓');
      await advance(page);
      await expect(page.locator(STEP_FIELDSETS.transac)).toBeVisible();
    } finally {
      db(`DELETE FROM natural_person WHERE personid = ${personID}`);
      db(`DELETE FROM person WHERE id = ${personID}`);
    }
  });

  test('Avançar reports the missing required item fields', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    const previews = trackRequests(page, '/nfe/emitir/preview');

    await fillRecipient(page);
    await advance(page);
    await expect(page.locator(STEP_FIELDSETS.transac)).toBeVisible();

    // Description, quantity and unit price are required on this step.
    await expect(statusBadge(page, 'transac')).toContainText('⚠ 3');

    await advance(page);

    await expect(page.locator(STEP_FIELDSETS.transac)).toBeVisible();
    const description = page.locator('input[name="items[0].productName"]');
    await expect(description).toBeFocused();
    expect(await description.evaluate((el) => el.validationMessage)).not.toBe('');
    await expect(description).toHaveClass(/nfe-field-invalid/);
    expect(previews).toHaveLength(0);

    // The marker disappears as soon as the field becomes valid.
    await description.fill('Soja Wizard E2E');
    await expect(description).not.toHaveClass(/nfe-field-invalid/);
    await expect(statusBadge(page, 'transac')).toContainText('⚠ 2');
  });

  test('the footer and the preview submit follow the active step', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');

    const back = page.locator('#wizard-back');
    const next = page.locator('#wizard-next');
    const preview = page.locator('[data-test-id="preview-detached-nfe"]');

    // First step: nothing to go back to, and no submit action.
    await expect(back).toBeHidden();
    await expect(next).toBeVisible();
    await expect(preview).toBeHidden();
    await expect(preview).toBeDisabled();

    await fillRecipient(page);
    await advance(page);
    await expect(back).toBeVisible();
    await expect(preview).toBeHidden();
    await expect(preview).toBeDisabled();

    // Voltar never validates: an invalid step can still be left.
    await goBack(page);
    await expect(page.locator(STEP_FIELDSETS.dest)).toBeVisible();

    // Walk the remaining steps.
    await fillRecipient(page);
    await advance(page);
    await fillFirstItem(page);
    await advance(page);
    await expect(page.locator(STEP_FIELDSETS.transp)).toBeVisible();
    await advance(page);

    // Last step: Avançar is gone and the single network action is available.
    await expect(page.locator(STEP_FIELDSETS.trib)).toBeVisible();
    await expect(next).toBeHidden();
    await expect(back).toBeVisible();
    await expect(preview).toBeVisible();
    await expect(preview).toBeEnabled();
  });

  test('the submit reveals the step of a hidden invalid field instead of failing silently', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    const previews = trackRequests(page, '/nfe/emitir/preview');

    await fillRecipient(page);
    await advance(page);
    await fillFirstItem(page);
    await advance(page);
    await advance(page);
    await expect(page.locator(STEP_FIELDSETS.trib)).toBeVisible();

    // Break a required field of an already-completed step, then jump forward
    // through the chip (free navigation, no validation).
    await goToStep(page, 'transac');
    await page.fill('input[name="items[0].productName"]', '');
    await goToStep(page, 'trib');
    await expect(page.locator('[data-test-id="preview-detached-nfe"]')).toBeEnabled();

    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate((el) => el.click());

    // The owning step is revealed, the field focused, and nothing is sent.
    await expect(page.locator(STEP_FIELDSETS.transac)).toBeVisible();
    await expect(page.locator(STEP_FIELDSETS.trib)).toBeHidden();
    const description = page.locator('input[name="items[0].productName"]');
    await expect(description).toBeFocused();
    expect(await description.evaluate((el) => el.validationMessage)).not.toBe('');
    await expect(description).toHaveClass(/nfe-field-invalid/);
    expect(previews).toHaveLength(0);

    // Fixing it and submitting again reaches the server: the backstop no
    // longer blocks the flow client-side. (The response itself depends on the
    // farm certificate — nfeCertGate — which the e2e fixture does not seed;
    // see the SDD's Verification Results.)
    await description.fill('Soja Wizard E2E');
    await goToStep(page, 'trib');
    const previewRequest = page.waitForRequest(
      (r) => r.url().includes('/nfe/emitir/preview') && r.method() === 'POST'
    );
    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate((el) => el.click());
    await previewRequest;
    expect(previews).toHaveLength(1);
  });

  test('saving a rascunho reaches the server with an incomplete form', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    const saves = trackRequests(page, '/nfe/rascunho/salvar');

    // Incomplete form: no recipient, no items.
    await page.locator('[data-test-id="open-save-rascunho"]').evaluate((el) => el.click());
    const dialog = page.locator('dialog#save-rascunho-dialog');
    await expect(dialog).toBeVisible();
    await page.fill('#rascunhoNome', `Rascunho Wizard ${Date.now().toString(36)}`);

    const saveResponse = page.waitForResponse(
      (r) => r.url().includes('/nfe/rascunho/salvar') && r.request().method() === 'POST'
    );
    await dialog.locator('[data-test-id="confirm-save-rascunho"]').evaluate((el) => el.click());

    // The request is no longer halted client-side: the server answers and
    // reports the incomplete data with its own pt-br toast.
    const response = await saveResponse;
    expect(response.status()).toBe(400);
    await expect(
      page.locator('#armazenda-toast', { hasText: 'Preço unitário inválido' }).first()
    ).toBeVisible();
    await expect(dialog).toBeVisible();
    expect(saves.length).toBeGreaterThan(0);

    await dialog.locator('#cancel-save-rascunho').evaluate((el) => el.click());
    await expect(dialog).toBeHidden();
  });

  test('the happy path still submits the preview from the last step', async ({ page, browserName }) => {
    await login(page, browserName);
    await gotoApp(page, '/nfe/emitir');
    const previews = trackRequests(page, '/nfe/emitir/preview');

    await fillRecipient(page);
    await advance(page);
    await fillFirstItem(page);
    await advance(page);
    await advance(page);
    await expect(page.locator(STEP_FIELDSETS.trib)).toBeVisible();

    // The last step's submit is the single network action of the form: exactly
    // one POST is issued. Whether the server renders the DANFE depends on the
    // farm certificate (nfeCertGate), which the e2e fixture does not seed —
    // that is a pre-existing environment limitation, not a regression of this
    // change (see the SDD's Verification Results).
    const previewRequest = page.waitForRequest(
      (r) => r.url().includes('/nfe/emitir/preview') && r.method() === 'POST'
    );
    await page.locator('[data-test-id="preview-detached-nfe"]').evaluate((el) => el.click());
    await previewRequest;
    expect(previews).toHaveLength(1);
  });
});
