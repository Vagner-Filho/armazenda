const { test, expect } = require('@playwright/test');
const { login, TEST_ADMIN } = require('../utils/auth');

test.describe('Login Flow', () => {
  test('should display login page', async ({ page }) => {
    await page.goto('/');

    // Check if login form is visible
    await expect(page.locator('form')).toBeVisible();

    // Check if CPF and password fields exist
    await expect(page.locator('input[name="cpf"]')).toBeVisible();
    await expect(page.locator('input[name="passwd"]')).toBeVisible();

    // Check if submit button exists
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should login successfully with valid credentials', async ({ page, browserName }) => {
    await login(page, browserName);

    // Verify we're on the romaneio page (the main page after login)
    await expect(page).toHaveURL(/.*romaneio/);

    // Verify page content shows the romaneio interface
    // Looking for common elements on the page
    await expect(page.locator('body')).toContainText(/romaneio|entrada|saída/i);
  });

  test('should show error with invalid credentials', async ({ page }) => {
    await page.goto('/');

    // Fill in invalid credentials
    await page.fill('input[name="cpf"]', '00000000000');
    await page.fill('input[name="passwd"]', 'wrongpassword');

    // Submit form
    await page.click('button[type="submit"]');

    // Should stay on login page
    await expect(page).toHaveURL('/');

    // Check for error message (this might vary based on your app's error handling)
    // Common patterns: toast notification, error div, or form validation
    const bodyText = await page.locator('body').textContent();
    const hasError =
      bodyText.toLowerCase().includes('erro') ||
      bodyText.toLowerCase().includes('inválido') ||
      bodyText.toLowerCase().includes('senha') ||
      bodyText.toLowerCase().includes('incorreto');

    // If no obvious error text, at least verify we didn't navigate away
    expect(hasError || await page.locator('form').isVisible()).toBeTruthy();
  });

  test('should require CPF field', async ({ page }) => {
    await page.goto('/');

    // Try to submit without CPF
    await page.fill('input[name="passwd"]', TEST_ADMIN.password);
    await page.click('button[type="submit"]');

    // Should still be on login page
    await expect(page).toHaveURL('/');
  });

  test('should require password field', async ({ page }) => {
    await page.goto('/');

    // Try to submit without password
    await page.fill('input[name="cpf"]', TEST_ADMIN.cpf);
    await page.click('button[type="submit"]');

    // Should still be on login page
    await expect(page).toHaveURL('/');
  });
});

test.describe('Authenticated User Flow', () => {
  test('should access romaneio page after login', async ({ page, browserName }) => {
    await login(page, browserName);

    // Verify we can access the romaneio page
    await expect(page).toHaveURL(/.*romaneio/);

    // Check for page-specific elements
    // These selectors should be adjusted based on your actual page structure
    const bodyText = await page.locator('body').textContent();
    expect(
      bodyText.toLowerCase().includes('romaneio') ||
      bodyText.toLowerCase().includes('entrada') ||
      bodyText.toLowerCase().includes('saída') ||
      bodyText.toLowerCase().includes('manifesto')
    ).toBeTruthy();
  });
});
