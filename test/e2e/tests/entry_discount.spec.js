/**
 * Entry E2E Tests
 * 
 * Tests for the entry (romaneio) feature, covering:
 * - Discounts with humidity, damage and impurity
 * - Humidity Progression for own entry;
 * - Humidity Progression as farm default;
 * - Humidity Progression for person;
 */

import { test, expect } from '@playwright/test';
import { login } from '../utils/auth';
import {
  createEntryWithDiscount,
} from '../utils/entry-helpers';
import { fillPersonForm, openForm, generatePersonFormData } from '../utils/helpers';

test.describe('Entry With Humidity Discount - System Wide Progression', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
  });

  test('should create entry with default dependencies, gross weight, tare and humidity 14.01', async ({ page }) => {
    await createEntryWithDiscount(page, { humidity: 14.01, expectedNetWeightDisplay: "24.995,75 kg" });
  });

  test('should create entry with default dependencies, gross weight, tare and humidity 16', async ({ page }) => {
    await createEntryWithDiscount(page, { humidity: 16, expectedNetWeightDisplay: "24.100 kg" });
  });

  test('should create entry with default dependencies, gross weight, tare and humidity 18', async ({ page }) => {
    await createEntryWithDiscount(page, { humidity: 18, expectedNetWeightDisplay: "23.000 kg" });
  });

  test('should create entry with default dependencies, gross weight, tare and humidity 20', async ({ page }) => {
    await createEntryWithDiscount(page, { humidity: 20, expectedNetWeightDisplay: "21.700 kg" });
  });
});

test.describe('Entry With Damage Discount', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
  });

  test('should create entry with default dependencies, gross weight, tare and damage 10', async ({ page }) => {
    await createEntryWithDiscount(page, { damage: 10, expectedNetWeightDisplay: "24.500 kg" });
  });

  test('should create entry with default dependencies, gross weight, tare and damage 8.01', async ({ page }) => {
    await createEntryWithDiscount(page, { damage: 8.01, expectedNetWeightDisplay: "24.997,5 kg" });
  });
});

test.describe('Entry With Impurity Discount', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
  });

  test('should create entry with default dependencies, gross weight, tare and damage 3', async ({ page }) => {
    await createEntryWithDiscount(page, { impurity: 3, expectedNetWeightDisplay: "24.500 kg" });
  });

  test('should create entry with default dependencies, gross weight, tare and damage 1.01', async ({ page }) => {
    await createEntryWithDiscount(page, { impurity: 1.01, expectedNetWeightDisplay: "24.997,5 kg" });
  });
});

test.describe('Entry With All Discounts', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
  });

  test('should create entry with default dependencies, gross weight, tare, damage 8.01, impurity 1.01, humidity 14.01', async ({ page }) => {
    await createEntryWithDiscount(page, { impurity: 1.01, damage: 8.01, humidity: 14.01, expectedNetWeightDisplay: "24.990,75 kg" });
  });

  test('should create entry with default dependencies, gross weight, tare, damage 10, impurity 3, humidity 16', async ({ page }) => {
    await createEntryWithDiscount(page, { impurity: 3, damage: 10, humidity: 16, expectedNetWeightDisplay: "23.100 kg" });
  });
});

test.describe('Entry With Default Storage Tax', () => {
  test.beforeEach(async ({ page, browserName }) => {
    // Login before each test
    await login(page, browserName);
    await page.locator('[data-test_id="pessoa-menu-option"]').click()
    await expect(page).toHaveURL(/.*pessoa/);
    await openForm(page, { novaTxt: "Nova Pessoa", comecarTxt: "Começar Agora", formId: "personFormDialog" });
  });

  test('should create entry with default dependencies and external origin for corn', async ({ page }) => {
    const nPerson = generatePersonFormData('natural', {
      entryCornDiscount: 5.5,
    });
    await fillPersonForm(page, 'natural', nPerson);
    const responsePromise = page.waitForResponse(
      response => response.url().includes('/person/natural') && response.request().method() === 'POST'
    );
    await page.locator('[data-test_id="submit-natural-btn"]').click({ force: true });
    const response = await responsePromise;
    expect(response.ok()).toBeTruthy();
    await expect(page.locator('dialog#personFormDialog')).not.toBeVisible();

    await page.locator('[data-test_id="romaneio-menu-option"]').click()
    await expect(page).toHaveURL(/.*romaneio/);
    await createEntryWithDiscount(page, { person: nPerson.name, expectedNetWeightDisplay: "23.625 kg" });
  });
});
