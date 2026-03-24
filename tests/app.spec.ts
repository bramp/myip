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
  "Insights": { "Proxy": "none" },
  "Location": {
    "MapURL": "https://maps.googleapis.com/maps/api/staticmap?size=640x400&scale=2&markers=color:red%7C37.7749,-122.4194"
  }
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
    const card = page.locator('.card').first();
    await expect(card).toBeVisible();
    await expect(card).toContainText('1.2.3.4');
    await expect(card).toContainText('IPv4');

    // Expand the card
    await card.locator('button.btn-outline-light').click();
    await expect(card.locator('.card-body')).toHaveClass(/show/);

    // Check for a Font Awesome 6 icon (it might be in the DOM but hidden by CSS if collapse is broken)
    const icon = card.locator('i.fa-brands.fa-chrome');
    await expect(icon).toBeAttached();
  });

  test('should have Bootstrap 5 card classes', async ({ page }) => {
    await page.goto('/');
    const card = page.locator('.card').first();
    await expect(card).toBeVisible();
    await expect(card.locator('.card-header')).toBeVisible();

    // Expand to see body
    await card.locator('button.btn-outline-light').click();
    await expect(card.locator('.card-body')).toHaveClass(/show/);
  });

  test('should expand the section when card header is clicked', async ({ page }) => {
    await page.goto('/');
    const card = page.locator('.card').first();
    const header = card.locator('.card-header');
    const body = card.locator('.card-body');

    // 1. Initially collapsed
    await expect(body).not.toHaveClass(/show/);

    // 2. Click header to expand
    await header.click();

    // 3. Now expanded
    await expect(body).toHaveClass(/show/);
    await expect(body).toBeVisible();

    // 4. Click again to collapse
    await header.click();
    await expect(body).not.toHaveClass(/show/);
  });

  test('should expand the section when "View details" is clicked', async ({ page }) => {
    await page.goto('/');
    const card = page.locator('.card').first();
    const body = card.locator('.card-body');
    const button = card.locator('button.btn-outline-light');

    // 1. Initially it should be collapsed (hidden)
    await expect(body).not.toHaveClass(/show/);
    await expect(body).not.toBeVisible();

    // 2. Click to expand
    await button.click();

    // 3. Now it should have the 'show' class and be visible
    await expect(body).toHaveClass(/show/);
    await expect(body).toBeVisible();

    // 4. Click again to collapse
    await button.click();
    await expect(body).not.toHaveClass(/show/);
    await expect(body).not.toBeVisible();
  });

  test('should not show "View details" for a failed response', async ({ page }) => {
    // Override mock for IPv6 to return an error
    await page.route(url => url.pathname.endsWith('/json') && url.searchParams.get('family') === 'IPv6', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          "RemoteAddrFamily": "IPv6",
          "Error": "unknown error"
        }),
      });
    });

    await page.goto('/');

    // Find the IPv6 card (usually the second one)
    const ipv6Card = page.locator('.card', { hasText: 'IPv6' });
    await expect(ipv6Card).toBeVisible();
    await expect(ipv6Card).toContainText('error: unknown error');

    // The button should NOT be visible
    const button = ipv6Card.locator('button.btn-outline-light');
    await expect(button).not.toBeVisible();

    // The header should NOT have role="button" or data-bs-toggle="collapse"
    const header = ipv6Card.locator('.card-header');
    await expect(header).not.toHaveAttribute('role', 'button');
    await expect(header).not.toHaveAttribute('data-bs-toggle', 'collapse');
  });

  test('should hide "Actual:" text when difference is only the port', async ({ page }) => {
    // Mock data where ActualRemoteAddr has a port but matches RemoteAddr
    const mockWithPort = {
      ...MOCK_DATA,
      "RemoteAddr": "1.2.3.4",
      "ActualRemoteAddr": "1.2.3.4:5678"
    };

    await page.route(url => url.pathname.endsWith('/json'), async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockWithPort),
      });
    });

    await page.goto('/');
    const card = page.locator('.card').first();
    await card.locator('button.btn-outline-light').click();

    // The "Actual Remote Addr" row should NOT be present (since it's just a port difference in our mock)
    await expect(card.locator('tr', { hasText: 'Actual Remote Addr' })).not.toBeVisible();
  });

  test('should hide Whois error alert if Whois body exists', async ({ page }) => {
    const mockWithWhoisErrorAndBody = {
      ...MOCK_DATA,
      "RemoteAddrWhois": {
        "Error": "some error",
        "Body": "valid whois text"
      }
    };

    await page.route(url => url.pathname.endsWith('/json'), async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockWithWhoisErrorAndBody),
      });
    });

    await page.goto('/');
    const card = page.locator('.card').first();
    await card.locator('button.btn-outline-light').click();

    // The error alert should NOT be visible
    await expect(card.locator('.alert-danger', { hasText: 'some error' })).not.toBeVisible();
    // The body should be visible
    await expect(card.locator('pre', { hasText: 'valid whois text' })).toBeVisible();
  });

  test('should show the map when expanded', async ({ page }) => {
    await page.goto('/');
    const card = page.locator('.card').first();
    await card.locator('button.btn-outline-light').click();

    const map = card.locator('img.img-fluid');
    await expect(map).toBeVisible();
    // Expect either a real maps URL or our base64 placeholder
    await expect(map).toHaveAttribute('src', /(maps\.googleapis\.com|^data:image\/svg\+xml;base64)/);
  });

  test('should have a working /json endpoint (integration)', async ({ request }) => {
    // This tests the real backend (not mocked via page.route)
    const response = await request.get('/json');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data).toHaveProperty('RemoteAddr');
  });
});
