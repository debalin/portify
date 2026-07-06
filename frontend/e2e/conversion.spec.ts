import { test, expect } from '@playwright/test';

test.describe('Playlist Conversion Flow', () => {
  test('should successfully complete a mocked conversion', async ({ page }) => {
    // 1. Visit the app
    await page.goto('/');

    // 2. Connect Source (Spotify Mock)
    // Ensure providers are loaded by waiting for the Login buttons
    const loginBtns = page.getByRole('button', { name: /Login to Connect/i });
    // Wait for at least one login button
    await expect(loginBtns.first()).toBeVisible({ timeout: 10000 });
    // Click the first "Login to Connect" button (Source)
    await loginBtns.first().click();

    // The mock adapter instantly redirects us back with ?code=mock-auth-code
    // "Mock Favorites" is the first mock playlist returned and should be auto-selected
    const sourcePlaylistSelect = page.locator('.provider-box.source .playlist-picker select').first();
    await expect(sourcePlaylistSelect).toHaveValue('playlist-with-tracks', { timeout: 10000 });

    // 3. Connect Destination (YouTube Mock)
    // Click the remaining "Login to Connect" button (Destination)
    await expect(loginBtns.first()).toBeVisible();
    await loginBtns.first().click();

    // Destination mock playlist should load, wait for the select element to appear
    const destPlaylistSelect = page.locator('.provider-box.destination .playlist-picker select').first();
    await expect(destPlaylistSelect).toBeVisible({ timeout: 10000 });

    // 4. Start Conversion
    const convertBtn = page.getByRole('button', { name: /Start Conversion/i });
    await expect(convertBtn).toBeVisible();
    await expect(convertBtn).toBeEnabled();
    await convertBtn.click();

    // 5. Verify Progress Bar
    // We should see it transitioning through progress states because of our time.Sleep in mock.go
    await expect(page.getByText(/Converting tracks/i)).toBeVisible({ timeout: 5000 });

    // 6. Verify Final Success State
    await expect(page.getByText(/Successfully converted/i)).toBeVisible({ timeout: 15000 });
    
    // The view playlist button should appear
    const viewPlaylistLink = page.getByRole('link', { name: /Open New Playlist/i });
    await expect(viewPlaylistLink).toBeVisible();
    await expect(viewPlaylistLink).toHaveAttribute('href', /youtube\.com\/playlist\?list=mock-playlist-id/);
  });
});
