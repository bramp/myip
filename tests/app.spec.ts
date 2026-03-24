import { test, expect } from '@playwright/test';

const MOCK_DATA = {
  "RemoteAddr": "1.2.3.4",
  "RemoteAddrFamily": "IPv4",
  "RemoteAddrReverse": { "Names": ["test.example.com"] },
  "Header": { "User-Agent": ["Playwright"] },
  "UserAgent": {
    "UserAgent": { "Family": "Chrome", "Major": "120" },
    "Os": { "Family": "macOS" },
    "Device": { "Family": "Desktop" }
  },
  "Insights": { "Proxy": "none" }
};

test.describe('What\'s My IP App', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the /json API call for ANY host (e.g., localhost, ip4.bramp.net, ip6.bramp.net)
    await page.route(url => url.pathname.endsWith('/json'), async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(MOCK_DATA),
      });
    });
  });

  test('should load the home page and show the title', async ({ page }) => {
    await page.goto('/');
    // Escaped the parentheses for literal matching
    await expect(page).toHaveTitle(/\(A better\) What's My IP Address\?/);
    await expect(page.locator('h1')).toContainText('What\'s My IP Address?');
  });

  test('should show the mocked IP address panel', async ({ page }) => {
    await page.goto('/');
    const panel = page.locator('.panel-primary').first();
    await expect(panel).toBeVisible();
    await expect(panel).toContainText('1.2.3.4');
    await expect(panel).toContainText('IPv4');
  });

  test('should have a working /json endpoint (integration)', async ({ request }) => {
    // This tests the real backend (not mocked via page.route)
    const response = await request.get('/json');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data).toHaveProperty('RemoteAddr');
  });
});
