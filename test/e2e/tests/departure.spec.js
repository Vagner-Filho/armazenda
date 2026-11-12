/**
 * Departure E2E Tests
 *
 * Tests for the departure (romaneio) feature, covering:
 * - Departure creation without analysis values
 *
 * The masked weight inputs (assets/js/weight.js) are covered per input mode,
 * as separate test cases: `fill` sets the value in one event (paste-like —
 * the only path dots reach the field through, as decimal or grouping) and
 * `pressSequentially` presses each key (the comma is the only typable
 * separator). Both canonicalize to the same dot-decimal `data-raw`.
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

  // Paste case: fill sets the whole value in one input event, the path dots
  // reach the field through — here as a plain dot decimal (grossWeight) and
  // as a grouped display string (tare). The data-raw assertion is the mask
  // regression guard (the dot-stripping bug produced 4714855/984513 and
  // overflowed NUMERIC(12,6) with SQLSTATE 22003).
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
      tare: '9.845,13',
      humidity: 5.65,
      impurity: 1.11,
      damage: 1.13
    });

    const netWeightInput = page.locator('input[name="netWeight"]#netWeight');
    await expect(netWeightInput).toHaveAttribute('data-raw', expectedNetWeight.toString());

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

  // Typing case: pressSequentially fires one input event per keystroke, like
  // a physical keyboard. The comma is the only typable separator — a dot
  // keystroke is blocked by the mask's keydown filter.
  test('should create a departure with comma-typed weights via keyboard', async ({ page }) => {
    await openDepartureForm(page);

    const { crop, vehicle } = await getDepartureDependencies(page, true);
    await ensureCropExists(page, crop.label);
    await ensureVehicleExists(page, vehicle.label);

    const grossWeight = 47148.55;
    const tare = 9845.13;
    const expectedNetWeight = grossWeight - tare;
    const expectedNetWeightDisplay = "37.303,42 kg";

    await fillDepartureForm(page, {
      crop: crop.value,
      vehicle: vehicle.value,
      grossWeight: '47148,55',
      tare: '9845,13',
      humidity: 5.65,
      impurity: 1.11,
      damage: 1.13
    }, { weightMethod: 'type' });

    const grossWeightInput = page.locator('input#grossWeight');
    const tareInput = page.locator('input#tare');
    await expect(grossWeightInput).toHaveValue('47.148,55 kg');
    await expect(grossWeightInput).toHaveAttribute('data-raw', '47148.55');
    await expect(tareInput).toHaveValue('9.845,13 kg');
    await expect(tareInput).toHaveAttribute('data-raw', '9845.13');

    const netWeightInput = page.locator('input[name="netWeight"]#netWeight');
    await expect(netWeightInput).toHaveAttribute('data-raw', expectedNetWeight.toString());

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
