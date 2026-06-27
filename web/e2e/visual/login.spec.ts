import { test, expect } from '@playwright/test';
import { freezeTime } from '../helpers/auth';

test.describe('login page visual regression', () => {
    test.beforeEach(async ({ page }) => {
        await freezeTime(page);
    });

    test('login page default state', async ({ page }) => {
        await page.goto('/login');
        await page.waitForSelector('[data-testid="login-username-input"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('login-default.png', {
            fullPage: false,
        });
    });
});
