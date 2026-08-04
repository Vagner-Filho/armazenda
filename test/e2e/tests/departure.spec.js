/**
 * Departure E2E Tests
 * 
 * Tests for the departure (romaneio) feature, covering:
 * - Departure creation without analysis values
 */

import { test, expect } from '@playwright/test';
import { login } from '../utils/auth';
import {
  fillDepartureForm,
  getDepartureDependencies,
  getDepartureRows,
  goToDepartures,
  openDepartureForm,
  submitDepartureForm
} from '../utils/departure-helpers';
import { ensureCropExists, ensureVehicleExists } from '../utils/entry-helpers';

test.describe('Departure Creation - Basic Flow', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
    await goToDepartures(page);
  });

  test('should create a basic departure with no analysis values', async ({ page }) => {
    await openDepartureForm(page);

    const { crop, vehicle } = await getDepartureDependencies(page, true);

    // Fill the departure form with basic data
    const grossWeight = 50000;
    const tare = 15000;
    const expectedNetWeight = grossWeight - tare;
    const expectedNetWeightDisplay = "35.000 kg";

    await fillDepartureForm(page, {
      crop: crop.value,
      vehicle: vehicle.value,
      grossWeight: grossWeight,
      tare: tare,
    });

    const netWeightInput = page.locator('input[name="netWeight"]#netWeight');
    expect(netWeightInput).toHaveAttribute('data-raw', expectedNetWeight.toString());

    const rowsBefore = await getDepartureRows(page);
    await submitDepartureForm(page);

    await expect.poll(
      async () => (await getDepartureRows(page)).length,
      { message: 'waiting for new table row after HTMX swap' }
    ).toBe(rowsBefore.length + 1);

    const rows = await getDepartureRows(page);
    expect(rows.length).toBeGreaterThan(0);

    const tableRow = page.locator('#departure-table-body tr').first();
    const displayPlate = await tableRow.locator('td[data-test_id="plate"]').textContent();
    expect(displayPlate.trim()).toBe(vehicle.label);

    const displayNetWeight = await tableRow.locator('td[data-test_id="net_weight"]').textContent();
    expect(displayNetWeight.trim()).toBe(expectedNetWeightDisplay);
  });

  test('should create a departure with analysis values', async ({ page }) => {
    // Open the entry form
    await openDepartureForm(page);

    const { crop, vehicle } = await getDepartureDependencies(page, true)
    await ensureCropExists(page, crop.label);
    await ensureVehicleExists(page, vehicle.label);

    const grossWeight = 47148.55;
    const tare = 9845.13;
    const expectedNetWeight = grossWeight - tare;
    const expectedNetWeightDisplay = "37.303,42 kg";

    await fillDepartureForm(page, {
      crop: crop.value,
      vehicle: vehicle.value,
      grossWeight: grossWeight,
      tare: tare,
      humidity: 5.65,
      impurity: 1.11,
      damage: 1.13
    });

    const netWeightInput = page.locator('input[name="netWeight"]#netWeight');
    expect(netWeightInput).toHaveAttribute('data-raw', expectedNetWeight.toString());

    const rowsBefore = await getDepartureRows(page);
    await submitDepartureForm(page);

    await expect.poll(
      async () => (await getDepartureRows(page)).length,
      { message: 'waiting for new table row after HTMX swap' }
    ).toBe(rowsBefore.length + 1);

    const tableRow = page.locator('#departure-table-body tr').first();
    const displayPlate = await tableRow.locator('td[data-test_id="plate"]').textContent();
    expect(displayPlate.trim()).toBe(vehicle.label);

    const displayNetWeight = await tableRow.locator('td[data-test_id="net_weight"]').textContent();
    expect(displayNetWeight.trim()).toBe(expectedNetWeightDisplay);
  });
});
