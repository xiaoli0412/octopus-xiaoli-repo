import { test, expect } from '@playwright/test';
import { setupAuth, setActiveNav, freezeTime, mockEndpoint } from '../helpers/auth';
import {
    mockUserStatus,
    mockSettingsList,
    mockPublicAccess,
    mockAPIKeyList,
    mockCapabilityInventory,
    mockUpdateLatest,
    mockUpdateNowVersion,
    mockUpdateStatus,
    mockAIGovernanceOverview,
    mockStrategyProfiles,
} from '../fixtures/common';

test.describe('settings visual regression', () => {
    test.beforeEach(async ({ page }) => {
        page.on('pageerror', (err) => console.error('[settings pageerror]', err));
        page.on('console', (msg) => {
            if (msg.type() === 'error') console.error('[settings console error]', msg.text());
        });
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'setting');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/setting/list', mockSettingsList);
        await mockEndpoint(page, '**/api/v1/setting/public-access', mockPublicAccess);
        await mockEndpoint(page, '**/api/v1/apikey/list', mockAPIKeyList);
        await mockEndpoint(page, '**/api/v1/model/capability-inventory', mockCapabilityInventory);
        await mockEndpoint(page, '**/api/v1/update', mockUpdateLatest);
        await mockEndpoint(page, '**/api/v1/update/now-version', mockUpdateNowVersion);
        await mockEndpoint(page, '**/api/v1/update/status', mockUpdateStatus);
        await mockEndpoint(page, '**/api/v1/ai/overview', mockAIGovernanceOverview);
        await mockEndpoint(page, '**/api/v1/ai/strategy-profiles', mockStrategyProfiles);
    });

    test('settings apikey card', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="setting-apikey"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('settings-apikey.png', {
            fullPage: false,
        });
    });

    test('settings info card', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="setting-info"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('settings-info.png', {
            fullPage: false,
        });
    });
});
