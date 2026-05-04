/**
 * Entry E2E Tests
 * 
 * Tests for the entry (romaneio) feature, covering:
 * - Basic entry creation (crop, field, vehicle, gross weight, tare)
 * - Entry creation with dependencies
 * - Form validation
 */

import { faker } from '@faker-js/faker';
import { test, expect } from '@playwright/test';
import { login } from '../utils/auth';
import {
  openEntryForm,
  ensureCropExists,
  ensureFieldExists,
  ensureVehicleExists,
  fillEntryForm,
  submitEntryForm,
  getEntryRows,
  getEntryDependencies,
} from '../utils/entry-helpers';

test.describe('Entry Creation - Basic Flow', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
  });

  test('should create a basic entry with default dependencies, gross weight and tare', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    const { crop, field, vehicle } = await getEntryDependencies(page, true);
    // Fill the entry form with basic data
    const grossWeight = 50000;
    const tare = 15000;
    const expectedNetWeight = grossWeight - tare;
    const expectedNetWeightDisplay = "35.000 kg";

    await fillEntryForm(page, {
      crop: crop.value,
      field: field.value,
      vehicle: vehicle.value,
      grossWeight: grossWeight,
      tare: tare,
    });

    const netWeightInput = page.locator('#net_weight');
    expect(netWeightInput).toHaveAttribute('data-raw', expectedNetWeight.toString());

    const rowsBefore = await getEntryRows(page);
    await submitEntryForm(page);
    // Wait for HTMX to prepend the new row (hx-swap="afterbegin")
    await expect.poll(
      async () => (await getEntryRows(page)).length,
      { message: 'waiting for new table row after HTMX swap' }
    ).toBe(rowsBefore.length + 1);

    // Verify the entry was created by checking the table
    const rows = await getEntryRows(page);
    expect(rows.length).toBeGreaterThan(0);

    // Verify the entry contains expected data
    // The table should show the vehicle plate and calculated net weight
    const tableRow = page.locator('#entries-table-body tr').first();
    const displayPlate = await tableRow.locator('td[data-test_id="plate"]').textContent();
    expect(displayPlate.trim()).toBe(vehicle.label);

    // Verify gross - tare calculation (35000 kg)
    // Note: The exact format may vary, so we check for the presence of the value
    const displayNetWeight = await tableRow.locator('td[data-test_id="net_weight"]').textContent();
    expect(displayNetWeight.trim()).toBe(expectedNetWeightDisplay);
  });

  test('should validate required fields', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Try to submit without filling required fields
    // First, ensure we have at least one option in each selector
    const { crop, field, vehicle } = await getEntryDependencies(page, true)
    await ensureCropExists(page, crop.label);
    await ensureFieldExists(page, field.label);
    await ensureVehicleExists(page, vehicle.label);

    // Clear the weight fields to test validation
    await page.fill('input#grossWeight', '');
    await page.fill('input#tare', '');

    // Try to submit
    await page.locator('dialog#addEntryDialog button[type="submit"]').evaluate(el => el.click());

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

    const { crop, field, vehicle } = await getEntryDependencies(page, true)
    await ensureCropExists(page, crop.label);
    await ensureFieldExists(page, field.label);
    await ensureVehicleExists(page, vehicle.label);

    // Fill in weights
    await page.fill('input#grossWeight', '45000');
    await page.fill('input#tare', '12000');

    // Wait for JavaScript calculation
    await page.waitForTimeout(500);

    // Check if net weight is calculated correctly (33000)
    const netWeightInput = page.locator('input#net_weight');
    expect(netWeightInput).toHaveValue('33.000 kg');
    expect(netWeightInput).toHaveAttribute('data-raw', '33000');
  });
});

test.describe('Entry Creation - Dependencies', () => {
  test.beforeEach(async ({ page, browserName }) => {
    await login(page, browserName);
  });

  test('should allow creating crop from entry form', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    // Click the add button for crop
    await page.locator('button[hx-get="/crop/form"]').evaluate(el => el.click());

    // Verify the crop form dialog opens
    await expect(page.locator('dialog#cropFormDialog')).toBeVisible();

    // Fill and submit the crop form
    const safraName = faker.lorem.word();
    await page.fill('input#crop-name', safraName);
    await page.selectOption('select#grain-selector', { label: 'Soja' });
    await page.fill('input#start-date', '2024-01-01');

    // Submit and wait for HTMX request to finish
    // Using evaluate to bypass WebKit's stacked dialog backdrop issue
    const cropResponse = page.waitForResponse(response =>
      response.url().includes('/crop') && response.request().method() === 'POST'
    );
    await page.locator('dialog#cropFormDialog button[type="submit"]').evaluate(el => el.click());
    await cropResponse;

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
    await page.locator('button[hx-get="/entry/field/form"]').evaluate(el => el.click());

    // Verify the field form dialog opens
    await expect(page.locator('dialog#fieldFormDialog')).toBeVisible();

    // Fill and submit the field form
    const fieldName = faker.lorem.word();
    await page.fill('input#field-name', fieldName);
    await page.fill('input#hectares', '25.5');

    // Submit and wait for HTMX request to finish
    // Using evaluate to bypass WebKit's stacked dialog backdrop issue
    const fieldResponse = page.waitForResponse(response =>
      response.url().includes('/field') && response.request().method() === 'POST'
    );
    await page.locator('dialog#fieldFormDialog button[type="submit"]').evaluate(el => el.click());
    await fieldResponse;

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
    await page.locator('button[hx-get="/vehicle/form"]').evaluate(el => el.click());

    // Verify the vehicle form dialog opens
    await expect(page.locator('dialog#vehicleFormDialog')).toBeVisible();

    // Fill and submit the vehicle form
    const vehiclePlate = faker.vehicle.vrm();
    await page.fill('input#vehicle-plate', vehiclePlate);
    await page.fill('input#vehicle-name', 'Novo Veículo');

    // Submit and wait for HTMX request to finish
    // Using evaluate to bypass WebKit's stacked dialog backdrop issue
    const vehicleResponse = page.waitForResponse(response =>
      response.url().includes('/vehicle') && response.request().method() === 'POST'
    );
    await page.locator('dialog#vehicleFormDialog button[type="submit"]').evaluate(el => el.click());
    await vehicleResponse;

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
  });

  test('should handle zero tare weight', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    const { crop, field, vehicle } = await getEntryDependencies(page, true)
    await ensureCropExists(page, crop.label);
    await ensureFieldExists(page, field.label);
    await ensureVehicleExists(page, vehicle.label);

    // Fill with zero tare
    await fillEntryForm(page, {
      crop: crop.label,
      field: field.label,
      vehicle: vehicle.label,
      grossWeight: 30000,
      tare: 0,
    });

    const rowsBefore = await getEntryRows(page);
    await submitEntryForm(page);
    // Wait for HTMX to prepend the new row (hx-swap="afterbegin")
    await expect.poll(
      async () => (await getEntryRows(page)).length,
      { message: 'waiting for new table row after HTMX swap' }
    ).toBe(rowsBefore.length + 1);

    const tableRow = page.locator('#entries-table-body tr').first();
    const weight = await tableRow.locator('td[data-test_id="net_weight"]').textContent();
    expect(weight).toContain('30.000 kg');
  });

  test('should handle large weight values', async ({ page }) => {
    // Open the entry form
    await openEntryForm(page);

    const { crop, field, vehicle } = await getEntryDependencies(page, true)
    await ensureCropExists(page, crop.label);
    await ensureFieldExists(page, field.label);
    await ensureVehicleExists(page, vehicle.label);

    // Fill with zero tare
    await fillEntryForm(page, {
      crop: crop.label,
      field: field.label,
      vehicle: vehicle.label,
      grossWeight: 999999,
      tare: 100000,
    });

    const rowsBefore = await getEntryRows(page);
    await submitEntryForm(page);
    // Wait for HTMX to prepend the new row (hx-swap="afterbegin")
    await expect.poll(
      async () => (await getEntryRows(page)).length,
      { message: 'waiting for new table row after HTMX swap' }
    ).toBe(rowsBefore.length + 1);

    const tableRow = page.locator('#entries-table-body tr').first();
    const plate = await tableRow.locator('td[data-test_id="plate"]').textContent();
    const weight = await tableRow.locator('td[data-test_id="net_weight"]').textContent();
    expect(plate.trim()).toBe(vehicle.label);
    expect(weight.trim()).toBe('899.999 kg');
  });
});
