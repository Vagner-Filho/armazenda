/**
 * Entry test helpers
 * Provides reusable functions for creating crops, fields, vehicles, and entries
 */

const { expect } = require('@playwright/test');

/**
 * Check if a select element has any valid options (excluding placeholder/error options)
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - The select element selector
 * @returns {Promise<[boolean, string[]]>}
 */
async function hasValidOptions(page, selector) {
  const options = await page.locator(`${selector} option`).allTextContents();
  // Filter out placeholder options like "Nenhuma safra encontrada" or "-1"
  const validOptions = options.filter(text =>
    text &&
    !text.includes('Nenhum') &&
    !text.includes('Nenhuma') &&
    text !== '-1'
  );
  return [validOptions.length > 0, options];
}

/**
 * Get the first valid option value from a select element
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - The select element selector
 * @returns {Promise<string|null>}
 */
async function getFirstValidOptionValue(page, selector) {
  const options = await page.locator(`${selector} option`).all();
  for (const option of options) {
    const value = await option.getAttribute('value');
    const text = await option.textContent();
    if (value && value !== '-1' && text && !text.includes('Nenhum') && !text.includes('Nenhuma')) {
      return value;
    }
  }
  return null;
}

/**
 * Create a new crop via the form dialog
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} name - Crop name
 * @param {string} product - Product name (Milho or Soja)
 * @param {string} startDate - Start date in YYYY-MM-DD format
 */
async function createCrop(page, name, product = 'Milho', startDate = null) {
  // Wait for any existing dialogs to close first
  await page.waitForTimeout(300);

  // Click the add button next to crop selector
  await page.click('button[hx-get="/crop/form"]');

  // Wait for the crop form dialog to appear
  await expect(page.locator('dialog#cropFormDialog')).toBeVisible();

  // Fill in the crop name
  await page.fill('input#crop-name', name);

  // Select the product (grain)
  await page.selectOption('select#grain-selector', { label: product });

  // Fill in the start date (default to today if not provided)
  const date = startDate || new Date().toISOString().split('T')[0];
  await page.fill('input#start-date', date);

  // Submit the form and wait for HTMX request to complete
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/crop') && response.request().method() === 'POST'),
    page.click('dialog#cropFormDialog button[type="submit"]')
  ]);

  // Wait for the dialog to close
  await expect(page.locator('dialog#cropFormDialog')).not.toBeVisible();

  // Wait for HTMX to update the selector and stabilize
  await page.waitForTimeout(800);
}

/**
 * Create a new field via the form dialog
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} name - Field name
 * @param {number} hectares - Field size in hectares
 */
async function createField(page, name, hectares = 10) {
  // Wait for any existing dialogs to close first
  await page.waitForTimeout(300);

  // Click the add button next to field selector
  await page.click('button[hx-get="/entry/field/form"]');

  // Wait for the field form dialog to appear
  await expect(page.locator('dialog#fieldFormDialog')).toBeVisible();

  // Fill in the field name
  await page.fill('input#field-name', name);

  // Fill in the hectares
  await page.fill('input#hectares', hectares.toString());

  // Submit the form and wait for HTMX request to complete
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/field') && response.request().method() === 'POST'),
    page.click('dialog#fieldFormDialog button[type="submit"]')
  ]);

  // Wait for the dialog to close
  await expect(page.locator('dialog#fieldFormDialog')).not.toBeVisible();

  // Wait for HTMX to update the selector and stabilize
  await page.waitForTimeout(800);
}

/**
 * Create a new vehicle via the form dialog
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} plate - Vehicle plate
 * @param {string} [name] - Optional vehicle name
 */
async function createVehicle(page, plate, name = null) {
  // Wait for any existing dialogs to close first
  await page.waitForTimeout(300);

  // Click the add button next to vehicle selector
  await page.click('button[hx-get="/vehicle/form"]');

  // Wait for the vehicle form dialog to appear
  await expect(page.locator('dialog#vehicleFormDialog')).toBeVisible();

  // Fill in the plate
  await page.fill('input#vehicle-plate', plate);

  // Fill in the name if provided
  if (name) {
    await page.fill('input#vehicle-name', name);
  }

  // Submit the form and wait for HTMX request to complete
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/vehicle') && response.request().method() === 'POST'),
    page.click('dialog#vehicleFormDialog button[type="submit"]')
  ]);

  // Wait for the dialog to close
  await expect(page.locator('dialog#vehicleFormDialog')).not.toBeVisible();

  // Wait for HTMX to update the selector and stabilize
  await page.waitForTimeout(800);
}

/**
 * Open the entry form dialog
 * Handles both empty state ("Começar agora" button) and normal state ("Nova Entrada" button)
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function openEntryForm(page) {
  // Try "Nova Entrada" button first (when entries exist)
  const novaEntradaBtn = page.locator('button:has-text("Nova Entrada")');
  const começarBtn = page.locator('button:has-text("Começar agora")');

  if (await novaEntradaBtn.isVisible().catch(() => false)) {
    await novaEntradaBtn.click();
  } else if (await começarBtn.isVisible().catch(() => false)) {
    await começarBtn.click();
  } else {
    throw new Error('Neither "Nova Entrada" nor "Começar agora" button found');
  }

  // Wait for the entry form dialog to appear
  await expect(page.locator('dialog#addEntryDialog')).toBeVisible();
}

/**
 * Ensure a crop exists (create if not)
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} name - Crop name to check/create
 * @param {string} product - Product name
 * @returns {Promise<string>} The selected crop value
 */
async function ensureCropExists(page, name = 'Safra Teste', product = 'Milho') {
  const [hasCrops, cropOptions] = await hasValidOptions(page, 'select#crop-selector');

  if (!hasCrops || !cropOptions.includes(name)) {
    await createCrop(page, name, product);
  }

  // Wait for HTMX to finish updating the selector
  await page.waitForTimeout(500);

  // Return the first valid crop option value
  return await getFirstValidOptionValue(page, 'select#crop-selector');
}

/**
 * Ensure a field exists (create if not)
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} name - Field name to check/create
 * @param {number} hectares - Field size
 * @returns {Promise<string>} The selected field value
 */
async function ensureFieldExists(page, name = 'Talhão Teste', hectares = 10) {
  const [hasFields, fieldOptions] = await hasValidOptions(page, 'select#field-selector');

  if (!hasFields || !fieldOptions.includes(name)) {
    await createField(page, name, hectares);
  }

  // Wait for HTMX to finish updating the selector
  await page.waitForTimeout(500);

  // Return the first valid field option value
  return await getFirstValidOptionValue(page, 'select#field-selector');
}

/**
 * Ensure a vehicle exists (create if not)
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} plate - Vehicle plate to check/create
 * @param {string} [name] - Optional vehicle name
 * @returns {Promise<string>} The selected vehicle value
 */
async function ensureVehicleExists(page, plate = 'TEST1234', name = null) {
  const [hasVehicles, vehicleOptions] = await hasValidOptions(page, 'select#vehicle-selector');

  if (!hasVehicles || !vehicleOptions.includes(name)) {
    await createVehicle(page, plate, name);
  }

  // Wait for HTMX to finish updating the selector
  await page.waitForTimeout(500);

  // Return the first valid vehicle option value
  return await getFirstValidOptionValue(page, 'select#vehicle-selector');
}

/**
 * Fill the entry form with data
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Object} data - Entry data
 * @param {string} [data.crop] - Crop ID (if null, will use first available)
 * @param {string} [data.field] - Field ID (if null, will use first available)
 * @param {string} [data.vehicle] - Vehicle ID (if null, will use first available)
 * @param {number} data.grossWeight - Gross weight in kg
 * @param {number} data.tare - Tare weight in kg
 * @param {string} [data.arrivalDate] - Arrival date in datetime-local format
 */
async function fillEntryForm(page, data) {
  // Select crop if provided
  if (data.crop) {
    await page.locator('select#crop-selector').selectOption(data.crop);
    await page.waitForTimeout(200);
  }

  // Select field if provided
  if (data.field) {
    await page.locator('select#field-selector').selectOption(data.field);
    await page.waitForTimeout(200);
  }

  // Select vehicle if provided
  if (data.vehicle) {
    await page.locator('select#vehicle-selector').selectOption(data.vehicle);
    await page.waitForTimeout(200);
  }

  // Fill gross weight
  await page.fill('input#grossWeight', data.grossWeight.toString());

  // Fill tare
  await page.fill('input#tare', data.tare.toString());

  // Fill arrival date if provided, otherwise use current date/time
  if (data.arrivalDate) {
    await page.fill('input#arrival-date', data.arrivalDate);
  }
}

/**
 * Submit the entry form
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function submitEntryForm(page) {
  // Click the submit button and wait for HTMX request to complete
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/entry') && response.request().method() === 'POST'),
    page.click('dialog#addEntryDialog button[type="submit"]')
  ]);

  // Wait for the dialog to close (successful submission)
  await expect(page.locator('dialog#addEntryDialog')).not.toBeVisible();
}

/**
 * Get all entry rows from the table
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @returns {Promise<Array>}
 */
async function getEntryRows(page) {
  return await page.locator('#entries-table-body tr').all();
}

/**
 * Check if an entry with specific data exists in the table
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Object} expectedData - Expected entry data to look for
 * @returns {Promise<boolean>}
 */
async function entryExistsInTable(page, expectedData) {
  const rows = await getEntryRows(page);

  for (const row of rows) {
    const rowText = await row.textContent();

    // Check if all expected data is present in the row
    let matches = true;
    if (expectedData.field && !rowText.includes(expectedData.field)) matches = false;
    if (expectedData.vehicle && !rowText.includes(expectedData.vehicle)) matches = false;
    if (expectedData.netWeight && !rowText.includes(expectedData.netWeight.toString())) matches = false;

    if (matches) return true;
  }

  return false;
}

module.exports = {
  hasValidOptions,
  getFirstValidOptionValue,
  createCrop,
  createField,
  createVehicle,
  openEntryForm,
  ensureCropExists,
  ensureFieldExists,
  ensureVehicleExists,
  fillEntryForm,
  submitEntryForm,
  getEntryRows,
  entryExistsInTable,
};
