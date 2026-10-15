const { expect, test } = require('@playwright/test');

test('serves a standalone Portuguese landing page with app destinations', async ({ page }) => {
  const response = await page.goto('/');

  expect(response.ok()).toBeTruthy();
  await expect(page).toHaveTitle('Armazenda — gestão para armazéns de grãos');
  await expect(page.locator('html')).toHaveAttribute('lang', 'pt-BR');
  await expect(page.locator('h1')).toHaveCount(1);
  await expect(page.locator('script')).toHaveCount(0);

  await expect(page.locator('a[href="https://app.armazenda.com/"]').first()).toBeVisible();
  await expect(page.locator('a[href="https://app.armazenda.com/pricing"]').first()).toBeVisible();

  for (const path of ['/styles.css', '/favicon.svg']) {
    const asset = await page.request.get(path);
    expect(asset.ok(), `${path} should be served by the static site`).toBeTruthy();
  }
});

test('keeps the layout, navigation, and accessibility usable across viewport sizes', async ({ page }) => {
  const pageErrors = [];
  page.on('pageerror', error => pageErrors.push(error.message));
  await page.goto('/');

  for (const width of [320, 390, 520, 760, 768, 1024, 1440]) {
    await page.setViewportSize({ width, height: 900 });
    const dimensions = await page.evaluate(() => ({
      viewport: document.documentElement.clientWidth,
      content: document.documentElement.scrollWidth,
    }));

    expect(dimensions.content, `horizontal overflow at ${width}px`).toBeLessThanOrEqual(dimensions.viewport);
    await expect(page.locator('.hero-actions .button')).toBeVisible();
    await expect(page.locator('.header-actions .button')).toBeVisible();
  }

  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.locator('.desktop-nav a[href="#recursos"]').click();
  await expect(page).toHaveURL(/#recursos$/);

  const sectionTop = await page.locator('#recursos').evaluate(node => node.getBoundingClientRect().top);
  const headerBottom = await page.locator('.site-header').evaluate(node => node.getBoundingClientRect().bottom);
  expect(sectionTop).toBeGreaterThanOrEqual(headerBottom - 1);

  await page.goto('/');
  const expectedLinks = await page.locator('a[href]').evaluateAll(links => [...new Set(links.map(link => link.href))]);
  await page.keyboard.press('Tab');
  await expect(page.locator('.skip-link')).toBeFocused();

  const reachedLinks = new Set();
  for (let i = 0; i < expectedLinks.length + 3; i++) {
    const href = await page.evaluate(() => document.activeElement?.href || '');
    if (href) reachedLinks.add(href);
    await page.keyboard.press('Tab');
  }
  expect(expectedLinks.filter(href => !reachedLinks.has(href))).toEqual([]);

  const headingLevels = await page.locator('h1,h2,h3,h4,h5,h6').evaluateAll(nodes => nodes.map(node => Number(node.tagName.slice(1))));
  expect(headingLevels[0]).toBe(1);
  expect(headingLevels.filter(level => level === 1)).toHaveLength(1);
  expect(headingLevels.every((level, index) => index === 0 || level <= headingLevels[index - 1] + 1)).toBeTruthy();
  await expect(page.locator('.dashboard-preview')).toHaveAttribute('aria-label', /fictícios/);
  expect(await page.evaluate(() => getComputedStyle(document.documentElement).scrollBehavior)).toBe('auto');
  expect(pageErrors).toEqual([]);
});
