import { test, expect } from '@playwright/test';
import { setupAuth, setActiveNav, freezeTime, mockEndpoint } from '../helpers/auth';
import {
    mockUserStatus,
    mockChannelList,
    mockStatsTotal,
    mockStatsDaily,
    mockStatsHourly,
    mockStatsTokenBreakdown,
} from '../fixtures/common';

test.describe('dashboard visual regression', () => {
    test.beforeEach(async ({ page }) => {
        page.on('console', (msg) => console.log(`[console ${msg.type()}]`, msg.text()));
        page.on('pageerror', (err) => console.error('[pageerror]', err));
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'home');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/channel/list', mockChannelList);
        await mockEndpoint(page, '**/api/v1/stats/total', mockStatsTotal);
        await mockEndpoint(page, '**/api/v1/stats/daily', mockStatsDaily);
        await mockEndpoint(page, '**/api/v1/stats/hourly', mockStatsHourly);
        await mockEndpoint(page, '**/api/v1/stats/token-breakdown**', mockStatsTokenBreakdown);
    });

    test('dashboard home page', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="home-page"]', { state: 'visible' });
        await page.waitForSelector('[data-testid="home-total-summary-card-0"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('dashboard-home.png', {
            fullPage: false,
        });
    });
});
