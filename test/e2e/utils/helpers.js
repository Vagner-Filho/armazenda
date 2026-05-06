import { faker } from '@faker-js/faker';
import { expect } from '@playwright/test';

/**
 * Generate fake person form data using faker.
 * @param {'legal'|'natural'} type - Person type to generate
 * @param {Object} [overrides] - Optional overrides for any field
 * @param {string} [overrides.name] - Company name (legal) or full name (natural)
 * @param {string} [overrides.fantasyName] - Fantasy name (legal only)
 * @param {string} [overrides.document] - CNPJ (legal) or CPF (natural)
 * @param {string} [overrides.ie] - Inscrição Estadual
 * @param {string} [overrides.humidityProgressionId] - Humidity progression select value
 * @param {number} [overrides.entrySoyDiscount] - Soy entry discount
 * @param {number} [overrides.entryCornDiscount] - Corn entry discount
 * @param {Object} [overrides.address] - Address data overrides
 * @returns {Object} Data object compatible with fillPersonForm
 */
function generatePersonFormData(type, overrides = {}) {
	const isLegal = type === 'legal';
	const rawDoc = isLegal
		? faker.string.numeric(14)
		: faker.string.numeric(11);

	const base = {
		name: isLegal ? faker.company.name() : faker.person.fullName(),
		document: rawDoc,
		ie: faker.string.numeric(12),
		entrySoyDiscount: parseFloat(faker.number.float({ min: 0, max: 10, fractionDigits: 1 }).toFixed(1)),
		entryCornDiscount: parseFloat(faker.number.float({ min: 0, max: 10, fractionDigits: 1 }).toFixed(1)),
		address: {
			cep: faker.string.numeric(8),
			state: faker.string.alpha({ length: 2, casing: 'upper' }),
			city: faker.location.city(),
			neighborhood: faker.location.county(),
			street: faker.location.street(),
			number: faker.number.int({ min: 1, max: 9999 }).toString(),
			complement: faker.location.secondaryAddress(),
			email: faker.internet.email(),
			phoneNumber: faker.string.numeric(11)
		}
	};

	if (isLegal) {
		base.fantasyName = faker.company.name();
	}

	return {
		...base,
		...overrides,
		address: {
			...base.address,
			...(overrides.address || {})
		}
	};
}

/**
 * Opens a form dialog by clicking either a "new" or "start" button.
 * Handles both empty state ("Começar agora") and normal state ("Nova ...") buttons.
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Object} locators - Locator configuration
 * @param {string} locators.novaTxt - Text of the "new" button
 * @param {string} locators.comecarTxt - Text of the "start" button
 * @param {string} locators.formId - ID of the form dialog to wait for
 * @throws {Error} If required locators are missing or neither button is found
 */
async function openForm(page, locators) {
	if (!locators.novaTxt || !locators.comecarTxt || !locators.formId) {
		throw new Error("Must provide full locators config")
	}
	// Try "Nova Entrada" button first (when entries exist)
	const novaBtn = page.locator(`button:has-text("${locators.novaTxt}")`);
	const comecarBtn = page.locator(`button:has-text("${locators.comecarTxt}")`);

	await expect(novaBtn.or(comecarBtn).first()).toBeVisible({ timeout: 10000 });

	if (await novaBtn.isVisible()) {
		await novaBtn.click({ force: true });
	} else if (await comecarBtn.isVisible()) {
		await comecarBtn.click({ force: true });
	} else {
		throw new Error('Neither "Nova Btn" nor "Começar agora" button found');
	}

	// Wait for the entry form dialog to appear
	await expect(page.locator(`dialog#${locators.formId}`)).toBeVisible();
}

/**
 * Fill the person form dialog.
 * Handles toggling between legal and natural person types when the toggler is present.
 * Address fields are scoped to the active form to avoid duplicate ID conflicts.
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {'legal'|'natural'} type - Person type to fill
 * @param {Object} data - Form data
 * @param {string} data.name - Company name (legal) or full name (natural)
 * @param {string} [data.fantasyName] - Fantasy name (legal only)
 * @param {string} data.document - CNPJ (legal) or CPF (natural)
 * @param {string} data.ie - Inscrição Estadual
 * @param {string} [data.humidityProgressionId] - Humidity progression select value
 * @param {number} [data.entrySoyDiscount] - Soy entry discount
 * @param {number} [data.entryCornDiscount] - Corn entry discount
 * @param {Object} [data.address] - Address data
 * @param {string} [data.address.cep] - CEP
 * @param {string} [data.address.state] - State (UF)
 * @param {string} [data.address.city] - City
 * @param {string} [data.address.neighborhood] - Neighborhood
 * @param {string} [data.address.street] - Street
 * @param {string} [data.address.number] - Number
 * @param {string} [data.address.complement] - Complement
 * @param {string} [data.address.email] - Email
 * @param {string} [data.address.phoneNumber] - Phone number
 */
async function fillPersonForm(page, type, data) {
	const isLegal = type === 'legal';
	const formTestId = isLegal ? 'person-legal-form' : 'person-natural-form';
	const formLocator = page.locator(`[data-test_id="${formTestId}"]`);

	// Toggle type only if the target form is not already visible
	const toggler = page.locator('#type-toggler');
	if (await toggler.isVisible().catch(() => false)) {
		const isHidden = await formLocator.isHidden().catch(() => true);
		if (isHidden) {
			const toggleBtn = page.locator(`[data-test_id="person-type-${type}-btn"]`);
			await toggleBtn.click();
		}
	}

	// Wait for the correct form to be visible before filling it
	await expect(formLocator).toBeVisible();

	// Fill identification fields
	if (isLegal) {
		await page.fill('[data-test_id="person-legal-name"]', data.name);
		if (data.fantasyName !== undefined) {
			await page.fill('[data-test_id="person-legal-fantasy-name"]', data.fantasyName);
		}
		await page.fill('[data-test_id="person-legal-document"]', data.document);
		await page.fill('[data-test_id="person-legal-ie"]', data.ie);
		if (data.humidityProgressionId !== undefined) {
			await page.selectOption('[data-test_id="person-legal-humidity"]', data.humidityProgressionId);
		}
		if (data.entrySoyDiscount !== undefined) {
			await page.fill('[data-test_id="person-legal-soy-discount"]', data.entrySoyDiscount.toString());
		}
		if (data.entryCornDiscount !== undefined) {
			await page.fill('[data-test_id="person-legal-corn-discount"]', data.entryCornDiscount.toString());
		}
	} else {
		await page.fill('[data-test_id="person-natural-name"]', data.name);
		await page.fill('[data-test_id="person-natural-document"]', data.document);
		await page.fill('[data-test_id="person-natural-ie"]', data.ie);
		if (data.humidityProgressionId !== undefined) {
			await page.selectOption('[data-test_id="person-natural-humidity"]', data.humidityProgressionId);
		}
		if (data.entrySoyDiscount !== undefined) {
			await page.fill('[data-test_id="person-natural-soy-discount"]', data.entrySoyDiscount.toString());
		}
		if (data.entryCornDiscount !== undefined) {
			await page.fill('[data-test_id="person-natural-corn-discount"]', data.entryCornDiscount.toString());
		}
	}

	// Fill address fields (scoped to the active form to avoid duplicate IDs)
	if (data.address) {
		const addr = data.address;
		if (addr.cep !== undefined) {
			await formLocator.locator('input#cep-input').fill(addr.cep);
		}
		if (addr.state !== undefined) {
			await formLocator.locator('input#state-input').fill(addr.state);
		}
		if (addr.city !== undefined) {
			await formLocator.locator('input#city-input').fill(addr.city);
		}
		if (addr.neighborhood !== undefined) {
			await formLocator.locator('input#neighborhood-input').fill(addr.neighborhood);
		}
		if (addr.street !== undefined) {
			await formLocator.locator('input#street-input').fill(addr.street);
		}
		if (addr.number !== undefined) {
			await formLocator.locator('input#number-input').fill(addr.number);
		}
		if (addr.complement !== undefined) {
			await formLocator.locator('input#complement-input').fill(addr.complement);
		}
		if (addr.email !== undefined) {
			await formLocator.locator('input#email-input').fill(addr.email);
		}
		if (addr.phoneNumber !== undefined) {
			await formLocator.locator('input#phoneNumber-input').fill(addr.phoneNumber);
		}
	}
}

export {
	openForm,
	fillPersonForm,
	generatePersonFormData
}
