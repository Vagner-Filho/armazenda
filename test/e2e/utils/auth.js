/**
 * Authentication helper for e2e tests
 * Provides reusable functions for logging in as the test admin user
 */

const { expect } = require('@playwright/test');

const TEST_ADMIN = {
  cpf: '52998224725',
  password: 'TestAdmin123!',
};

/**
 * Parse Set-Cookie header and extract cookie data
 * @param {string} setCookieHeader - The Set-Cookie header value
 * @param {string} url - The URL to associate with the cookie
 * @returns {Object} Parsed cookie object
 */
function parseSetCookieHeader(setCookieHeader, url) {
  const parts = setCookieHeader.split(';').map(p => p.trim());
  const [nameValue, ...attributes] = parts;
  const [name, value] = nameValue.split('=');

  const cookie = { name, value, url };

  for (const attr of attributes) {
    const lowerAttr = attr.toLowerCase();
    if (lowerAttr === 'httponly') {
      cookie.httpOnly = true;
    } else if (lowerAttr === 'secure') {
      // Override: don't set secure for WebKit compatibility
      cookie.secure = false;
    } else if (lowerAttr.startsWith('max-age=')) {
      cookie.expires = Date.now() / 1000 + parseInt(attr.substring(8), 10);
    }
    // Note: path and domain are ignored when url is set
  }

  return cookie;
}

/**
 * Login via API and manually set cookies (WebKit workaround)
 * WebKit doesn't accept secure cookies over HTTP, so we:
 * 1. Login via API request
 * 2. Parse Set-Cookie headers
 * 3. Set cookies manually with secure: false
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function loginViaApi(page) {
  const baseUrl = process.env.BASE_URL || 'http://localhost:8100';
  // Use 127.0.0.1 for API request to avoid IPv6 resolution issues
  const apiUrl = baseUrl.replace('localhost', '127.0.0.1');

  const response = await page.request.post(`${apiUrl}/login`, {
    form: {
      cpf: TEST_ADMIN.cpf,
      passwd: TEST_ADMIN.password,
    },
  });

  if (!response.ok()) {
    throw new Error(`Login failed with status ${response.status()}`);
  }

  const setCookieHeaders = response.headersArray().filter(
    h => h.name.toLowerCase() === 'set-cookie'
  );

  if (setCookieHeaders.length === 0) {
    throw new Error('No cookies received from login');
  }

  // Use the original baseUrl (localhost) for cookies, not apiUrl (127.0.0.1)
  const cookies = setCookieHeaders.map(h => parseSetCookieHeader(h.value, baseUrl));

  // Set cookies on the browser context with secure: false for localhost
  await page.context().addCookies(cookies);

  // Navigate to the main page
  await page.goto('/romaneio', { waitUntil: 'domcontentloaded' });
  await expect(page).toHaveURL(/.*romaneio/);
}

async function visitHome(page) {
  const baseUrl = (process.env.BASE_URL || 'http://localhost:8100')
    .replace('localhost', '127.0.0.1');

  const doGoto = () =>
    page.goto(baseUrl + '/', { waitUntil: 'domcontentloaded', timeout: 10000 });

  try {
    await doGoto();
  } catch (err) {
    if (err.message?.includes('Timeout') || err.message?.includes('timeout')) {
      // Retry once on navigation stall
      await doGoto();
    } else {
      throw err;
    }
  }
}

/**
 * Login to the application using the test admin credentials
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} [browserName] - Optional browser name for WebKit workaround
 */
async function login(page, browserName) {
  // WebKit workaround: secure cookies don't work over HTTP
  if (browserName === 'webkit') {
    return loginViaApi(page);
  }

  await visitHome(page);

  await expect(page.locator('form')).toBeVisible();
  await page.fill('input[name="cpf"]', TEST_ADMIN.cpf);
  await page.fill('input[name="passwd"]', TEST_ADMIN.password);
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/.*romaneio/);
}

/**
 * Logout from the application
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function logout(page) {
  await page.context().clearCookies();
  await page.goto('/', { waitUntil: 'domcontentloaded' });
  await expect(page.locator('form')).toBeVisible();
}

/**
 * Check if user is currently logged in
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @returns {Promise<boolean>}
 */
async function isLoggedIn(page) {
  try {
    await page.goto('/romaneio', { waitUntil: 'domcontentloaded' });
    const url = page.url();
    return url.includes('romaneio');
  } catch {
    return false;
  }
}

module.exports = {
  TEST_ADMIN,
  login,
  loginViaApi,
  logout,
  isLoggedIn,
  visitHome
};
