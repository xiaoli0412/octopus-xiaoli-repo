import { test, expect } from '@playwright/test';
import { setupAuth, setActiveNav, freezeTime, mockEndpoint } from '../helpers/auth';
import {
    mockUserStatus,
    mockChannelList,
    mockUpstreamSiteList,
    mockUpstreamSiteDetail,
} from '../fixtures/common';

test.describe('upstream visual regression', () => {
    test.beforeEach(async ({ page }) => {
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'upstream');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/channel/list', mockChannelList);
        await mockEndpoint(page, '**/api/v1/upstream/list', mockUpstreamSiteList);
        await mockEndpoint(page, '**/api/v1/upstream/detail/**', mockUpstreamSiteDetail);
    });

    test('upstream list page', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="upstream-page"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('upstream-list.png', {
            fullPage: false,
        });
    });

    test('upstream detail overview tab', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="upstream-page"]', { state: 'visible' });
        await page.waitForSelector('text=NewAPI 演示站', { state: 'visible' });
        await expect(page).toHaveScreenshot('upstream-detail-overview.png', {
            fullPage: false,
        });
    });
});
