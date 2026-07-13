import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { setupAuth, setActiveNav, freezeTime, mockEndpoint } from './helpers/auth';
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
    mockChannelList,
    mockStatsTotal,
    mockStatsDaily,
    mockStatsHourly,
    mockStatsTokenBreakdown,
} from './fixtures/common';

/**
 * 可访问性（a11y）自动化扫描。
 *
 * 使用 @axe-core/playwright 对关键页面跑 axe 扫描，
 * 断言无 critical 与 serious 级别违规。
 *
 * 覆盖页面：登录、仪表盘、渠道列表、设置-APIKey。
 */

/** 断言 axe 扫描结果中无 critical / serious 违规。 */
function expectNoCriticalViolations(
    results: Awaited<ReturnType<AxeBuilder['analyze']>>
) {
    const serious = results.violations.filter(
        (v) => v.impact === 'critical' || v.impact === 'serious'
    );
    if (serious.length > 0) {
        const summary = serious
            .map(
                (v) =>
                    `  - [${v.impact}] ${v.id}: ${v.description} (${v.nodes.length} nodes)`
            )
            .join('\n');
        throw new Error(
            `发现 ${serious.length} 条 critical/serious 级别 a11y 违规:\n${summary}`
        );
    }
    expect(serious).toHaveLength(0);
}

test.describe('a11y 扫描：关键页面', () => {
    test('登录页无 critical/serious 违规', async ({ page }) => {
        await freezeTime(page);
        await page.goto('/login');
        await page.waitForSelector('[data-testid="login-username-input"]', {
            state: 'visible',
        });

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
            .analyze();

        expectNoCriticalViolations(results);
    });

    test('仪表盘无 critical/serious 违规', async ({ page }) => {
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'home');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/channel/list', mockChannelList);
        await mockEndpoint(page, '**/api/v1/stats/total', mockStatsTotal);
        await mockEndpoint(page, '**/api/v1/stats/daily', mockStatsDaily);
        await mockEndpoint(page, '**/api/v1/stats/hourly', mockStatsHourly);
        await mockEndpoint(
            page,
            '**/api/v1/stats/token-breakdown**',
            mockStatsTokenBreakdown
        );

        await page.goto('/');
        await page.waitForSelector('[data-testid="home-page"]', {
            state: 'visible',
        });
        await page.waitForSelector('[data-testid="home-total-summary-card-0"]', {
            state: 'visible',
        });

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
            .analyze();

        expectNoCriticalViolations(results);
    });

    test('渠道列表无 critical/serious 违规', async ({ page }) => {
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'channel');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/channel/list', mockChannelList);
        await mockEndpoint(page, '**/api/v1/group/list', []);
        await mockEndpoint(page, '**/api/v1/stats/total', mockStatsTotal);

        await page.goto('/');
        await page.waitForSelector('[data-testid="channel-page"]', {
            state: 'visible',
        });

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
            .analyze();

        expectNoCriticalViolations(results);
    });

    test('设置-APIKey 无 critical/serious 违规', async ({ page }) => {
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'setting');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/setting/list', mockSettingsList);
        await mockEndpoint(page, '**/api/v1/setting/public-access', mockPublicAccess);
        await mockEndpoint(page, '**/api/v1/apikey/list', mockAPIKeyList);
        await mockEndpoint(
            page,
            '**/api/v1/model/capability-inventory',
            mockCapabilityInventory
        );
        await mockEndpoint(page, '**/api/v1/update', mockUpdateLatest);
        await mockEndpoint(page, '**/api/v1/update/now-version', mockUpdateNowVersion);
        await mockEndpoint(page, '**/api/v1/update/status', mockUpdateStatus);
        await mockEndpoint(page, '**/api/v1/ai/overview', mockAIGovernanceOverview);
        await mockEndpoint(
            page,
            '**/api/v1/ai/strategy-profiles',
            mockStrategyProfiles
        );

        await page.goto('/');
        await page.waitForSelector('[data-testid="setting-apikey"]', {
            state: 'visible',
        });

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
            .analyze();

        expectNoCriticalViolations(results);
    });
});
