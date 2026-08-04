/**
 * Departure test helpers
 */

import { expect } from '@playwright/test';
import { faker } from '@faker-js/faker';
import { ensureCropExists, ensureVehicleExists } from './entry-helpers';

/**
 * Open the departure form dialog
 * Handles both empty state ("Começar agora" button) and normal state ("Nova Saída" button)
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function openDepartureForm(page) {
	// Try "Nova Entrada" button first (when entries exist)
	const novaSaidaBtn = page.locator('button:has-text("Nova Saída")');
	const comecarBtn = page.locator('button:has-text("Começar agora")');

	await expect(novaSaidaBtn.or(comecarBtn).first()).toBeVisible({ timeout: 10000 });

	if (await novaSaidaBtn.isVisible()) {
		await novaSaidaBtn.click({ force: true });
	} else if (await comecarBtn.isVisible()) {
		await comecarBtn.click({ force: true });
	} else {
		throw new Error('Neither "Nova Saída" nor "Começar agora" button found');
	}

	// Wait for the departure form dialog to appear
	await expect(page.locator('dialog#departure-form-dialog')).toBeVisible();
}

/**
 * Go to the departure section of the romaneio page.
 * @param {import('@playwright/test').Page} page - Playwright page object
*/
async function goToDepartures(page) {
	await page.click('[data-test_id="romaneio-menu-option"]');
	await expect(page).toHaveURL(/romaneio/);
	await page.click('[data-test-id="departure-toggler"]');
}

/**
 * Ensures departure form dependencies exist and returns their selected values.
 * Creates crops and vehicles as needed using faker for default names.
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Object} args - Dependency options
 * @param {string} [args.cropName] - Crop name (uses faker if not provided)
 * @param {string} [args.cropProduct] - Product name (default: 'Milho')
 * @param {string} [args.vehiclePlate] - Vehicle plate (uses faker if not provided)
 * @param {string} [args.vehicleName] - Optional vehicle name
 * @returns {Promise<{crop: {value: string, label: string}, vehicle: {value: string, label: string}}>}
 */
async function getDepartureDependencies(page, args = { cropName: faker.lorem.word(), cropProduct: 'Milho', vehiclePlate: faker.vehicle.vrm(), vehicleName: null }) {
	const dependencies = {}
	if (typeof args === 'boolean' && args === true) {
		dependencies.cName = 'Milho Default'
		dependencies.cProd = 1
		dependencies.vPlate = 'ABC123'
	} else {
		dependencies.cName = args.cropName
		dependencies.cProd = args.cropProduct
		dependencies.vPlate = args.vehiclePlate
	}

	const cropValue = await ensureCropExists(page, dependencies.cName, dependencies.cProd);
	const vehicleValue = await ensureVehicleExists(page, dependencies.vPlate, args.vehicleName);

	return {
		crop: { value: cropValue, label: dependencies.cName },
		vehicle: { value: vehicleValue, label: dependencies.vPlate }
	};
}

/**
 * Fill the departure form with data
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Object} data - Entry data
 * @param {string} [data.crop] - Crop ID (if null, will use first available)
 * @param {string} [data.vehicle] - Vehicle ID (if null, will use first available)
 * @param {number} data.grossWeight - Gross weight in kg
 * @param {number} data.tare - Tare weight in kg
 * @param {string} [data.arrivalDate] - Arrival date in datetime-local format
 * @param {number} data.humidity - Humidity percentage
 * @param {number} data.impurity - Impurity percentage
 * @param {number} data.damage - Damage percentage
 */
async function fillDepartureForm(page, data) {
	if (data.crop) {
		await page.locator('select#crop-selector').selectOption(data.crop);
		await page.waitForTimeout(200);
	}

	if (data.person) {
		await page.locator('select#origin-selector').selectOption(data.person);
		await page.waitForTimeout(200);
	}

	if (data.vehicle) {
		await page.locator('select#vehicle-selector').selectOption(data.vehicle);
		await page.waitForTimeout(200);
	}

	await page.fill('input#grossWeight', data.grossWeight.toString());

	await page.fill('input#tare', data.tare.toString());

	if (data.arrivalDate) {
		await page.fill('input#arrival-date', data.arrivalDate);
	}

	if (data.humidity) {
		await page.fill('input#humidity', data.humidity.toString());
	}

	if (data.impurity) {
		await page.fill('input#impurity', data.impurity.toString());
	}

	if (data.damage) {
		await page.fill('input#damage', data.damage.toString());
	}
}

/**
 * Get all departure rows from the table
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @returns {Promise<Array>}
 */
async function getDepartureRows(page) {
	return await page.locator('#departure-table-body tr').all();
}

/**
 * Submit the departure form
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function submitDepartureForm(page) {
	const responsePromise = page.waitForResponse(
		response => response.url().includes('/departure') && response.request().method() === 'POST'
	);

	//INFO: bypass webkit backdrop issue
	await page.locator('dialog#departure-form-dialog button[type="submit"]').evaluate(el => el.click());
	await responsePromise;

	await expect(page.locator('dialog#departure-form-dialog')).not.toBeVisible();
}

export {
	openDepartureForm,
	goToDepartures,
	getDepartureDependencies,
	fillDepartureForm,
	getDepartureRows,
	submitDepartureForm
}
