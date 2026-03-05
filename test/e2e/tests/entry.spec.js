/**
 * Entry E2E Tests
 * 
 * Tests for the entry (romaneio) feature, covering:
 * - Basic entry creation (crop, field, vehicle, gross weight, tare)
 * - Entry creation with dependencies
 * - Form validation
 */

import { faker } from '@faker-js/faker';

const { test, expect } = require('@playwright/test');
const { login } = require('../utils/auth');
const {
  openEntryForm,
  ensureCropExists,
  ensureFieldExists,
  ensureVehicleExists,
  fillEntryForm,
  submitEntryForm,
  getEntryRows,
  entryExistsInTable,
} = require('../utils/entry-helpers');

test.describe('Entry Creation - Basic Flow', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);

    // Navigate to romaneio page
    await page.goto('/romaneio');
    await expect(page).toHaveURL(/.*romaneio/);
  });

  test('should create a basic entry with crop, field, vehicle, gross weight and tare', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Ensure required dependencies exist (create if not)
    const cropValue = await ensureCropExists(page, 'Safra 2024', 'Milho');
    const fieldValue = await ensureFieldExists(page, 'Talhão A', 10);
    const vehicleValue = await ensureVehicleExists(page, 'ABC1234', 'Caminhão Teste');

    // Fill the entry form with basic data
    const grossWeight = 50000;
    const tare = 15000;
    const expectedNetWeight = grossWeight - tare;

    await fillEntryForm(page, {
      crop: cropValue,
      field: fieldValue,
      vehicle: vehicleValue,
      grossWeight: grossWeight,
      tare: tare,
    });

    // Submit the form
    await submitEntryForm(page);

    // Wait for the entry to appear in the table
    await page.waitForTimeout(1000); // Allow HTMX to update

    // Verify the entry was created by checking the table
    const rows = await getEntryRows(page);
    expect(rows.length).toBeGreaterThan(0);

    // Verify the entry contains expected data
    // The table should show the vehicle plate and calculated net weight
    const tableContent = await page.locator('#entries-table-body').textContent();
    expect(tableContent).toContain('ABC1234');

    // Verify gross - tare calculation (35000 kg)
    // Note: The exact format may vary, so we check for the presence of the value
    expect(tableContent).toContain(expectedNetWeight.toString());
  });

  test('should validate required fields', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Try to submit without filling required fields
    // First, ensure we have at least one option in each selector
    await ensureCropExists(page, 'Safra Teste', 'Milho');
    await ensureFieldExists(page, 'Talhão Teste', 5);
    await ensureVehicleExists(page, 'TEST0001');

    // Clear the weight fields to test validation
    await page.fill('input#grossWeight', '');
    await page.fill('input#tare', '');

    // Try to submit
    await page.click('dialog#addEntryDialog button[type="submit"]');

    // The form should still be visible (validation failed)
    await expect(page.locator('dialog#addEntryDialog')).toBeVisible();

    // Check for validation messages or required attribute
    const grossWeightInput = await page.locator('input#grossWeight');
    const tareInput = await page.locator('input#tare');

    expect(grossWeightInput).toHaveAttribute('required');
    expect(tareInput).toHaveAttribute('required');
  });

  test('should calculate net weight correctly', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Ensure dependencies exist
    await ensureCropExists(page, 'Safra Calc', 'Soja');
    await ensureFieldExists(page, 'Talhão Calc', 15);
    await ensureVehicleExists(page, 'CALC0001');

    // Fill in weights
    await page.fill('input#grossWeight', '45000');
    await page.fill('input#tare', '12000');

    // Wait for JavaScript calculation
    await page.waitForTimeout(500);

    // Check if net weight is calculated correctly (33000)
    const netWeightInput = page.locator('input#netWeight');
    expect(netWeightInput).toHaveValue('33.000 kg');
    expect(netWeightInput).toHaveAttribute('data-raw', '33000');
  });
});

test.describe('Entry Creation - Dependencies', () => {
  test.beforeEach(async ({ page, browserName }) => {
    await login(page, browserName);
    await page.goto('/romaneio');
    await expect(page).toHaveURL(/.*romaneio/);
  });

  test('should allow creating crop from entry form', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Click the add button for crop
    await page.click('button[hx-get="/crop/form"]');

    // Verify the crop form dialog opens
    await expect(page.locator('dialog#cropFormDialog')).toBeVisible();

    // Fill and submit the crop form

    const safraName = faker.lorem.word()
    await page.fill('input#crop-name', safraName);
    await page.selectOption('select#grain-selector', { label: 'Soja' });
    await page.fill('input#start-date', '2024-01-01');
    await page.click('dialog#cropFormDialog button[type="submit"]');

    // Wait for dialog to close
    await expect(page.locator('dialog#cropFormDialog')).not.toBeVisible();

    // Verify we're back to entry form
    await expect(page.locator('dialog#addEntryDialog')).toBeVisible();

    // Verify the new crop appears in the selector
    const cropOptions = await page.locator('select#crop-selector option').allTextContents();
    expect(cropOptions.some(text => text.includes(safraName))).toBeTruthy();
  });

  test('should allow creating field from entry form', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Click the add button for field
    await page.click('button[hx-get="/entry/field/form"]');

    // Verify the field form dialog opens
    await expect(page.locator('dialog#fieldFormDialog')).toBeVisible();

    // Fill and submit the field form
    const fieldName = faker.lorem.word();
    await page.fill('input#field-name', fieldName);
    await page.fill('input#hectares', '25.5');
    await page.click('dialog#fieldFormDialog button[type="submit"]');

    // Wait for dialog to close
    await expect(page.locator('dialog#fieldFormDialog')).not.toBeVisible();

    // Verify we're back to entry form
    await expect(page.locator('dialog#addEntryDialog')).toBeVisible();

    // Verify the new field appears in the selector
    const fieldOptions = await page.locator('select#field-selector option').allTextContents();
    expect(fieldOptions.some(text => text.includes(fieldName))).toBeTruthy();
  });

  test('should allow creating vehicle from entry form', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Click the add button for vehicle
    await page.click('button[hx-get="/vehicle/form"]');

    // Verify the vehicle form dialog opens
    await expect(page.locator('dialog#vehicleFormDialog')).toBeVisible();

    // Fill and submit the vehicle form
    const vehiclePlate = faker.vehicle.vrm();
    await page.fill('input#vehicle-plate', vehiclePlate);
    await page.fill('input#vehicle-name', 'Novo Veículo');
    await page.click('dialog#vehicleFormDialog button[type="submit"]');

    // Wait for dialog to close
    await expect(page.locator('dialog#vehicleFormDialog')).not.toBeVisible();

    // Verify we're back to entry form
    await expect(page.locator('dialog#addEntryDialog')).toBeVisible();

    // Verify the new vehicle appears in the selector
    const vehicleOptions = await page.locator('select#vehicle-selector option').allTextContents();
    expect(vehicleOptions.some(text => text.includes(vehiclePlate))).toBeTruthy();
  });
});

test.describe('Entry Creation - Edge Cases', () => {
  test.beforeEach(async ({ page, browserName }) => {
    await login(page, browserName);
    await page.goto('/romaneio');
    await expect(page).toHaveURL(/.*romaneio/);
  });

  test('should handle zero tare weight', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Ensure dependencies exist
    const cropValue = await ensureCropExists(page, 'Safra Zero', 'Milho');
    const fieldValue = await ensureFieldExists(page, 'Talhão Zero', 5);
    const vehicleValue = await ensureVehicleExists(page, 'ZERO0001');

    // Fill with zero tare
    await fillEntryForm(page, {
      crop: cropValue,
      field: fieldValue,
      vehicle: vehicleValue,
      grossWeight: 30000,
      tare: 0,
    });

    // Submit
    await submitEntryForm(page);

    // Verify entry was created
    await page.waitForTimeout(1000);
    const tableContent = await page.locator('#entries-table-body').textContent();
    expect(tableContent).toContain('30000');
  });

  test.only('should handle large weight values', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Ensure dependencies exist
    const cropName = faker.lorem.word();
    const fieldName = faker.lorem.word();
    const vehicleName = faker.vehicle.vrm() + ' LARGE';

    const cropValue = await ensureCropExists(page, cropName, 'Milho');
    const fieldValue = await ensureFieldExists(page, fieldName, 100);
    const vehicleValue = await ensureVehicleExists(page, vehicleName);

    // Fill with large values
    await fillEntryForm(page, {
      crop: cropValue,
      field: fieldValue,
      vehicle: vehicleValue,
      grossWeight: 999999,
      tare: 100000,
    });

    // Submit
    await submitEntryForm(page);

    // Verify entry was created
    await page.waitForTimeout(1000);
    const tableContent = await page.locator('#entries-table-body').textContent();
    expect(tableContent).toContain(vehicleName);
  });
});
