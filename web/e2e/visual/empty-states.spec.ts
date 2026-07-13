import { test, expect } from '@playwright/test';
import { setupAuth, setActiveNav, freezeTime, mockEndpoint } from '../helpers/auth';
import { mockUserStatus } from '../fixtures/common';

/**
 * 空状态与骨架屏视觉回归
 *
 * 覆盖任务 8.8：检查 Skeleton/EmptyState 组件在以下场景的视觉表现：
 * - 列表为空时显示 EmptyState 组件
 * - 数据加载中时显示 Skeleton 骨架屏
 *
 * 注意：Playwright config 已设置 reducedMotion: 'reduce'，
 * 骨架屏的 animate-pulse 会被 motion-reduce:animate-none 抑制，
 * 空状态的 motion 淡入动画也会被跳过，保证截图稳定。
 */

test.describe('empty states and skeletons visual regression', () => {
    test.beforeEach(async ({ page }) => {
        await freezeTime(page);
        await setupAuth(page);
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
    });

    test('channel list empty state', async ({ page }) => {
        await setActiveNav(page, 'channel');
        await mockEndpoint(page, '**/api/v1/channel/list', []);
        await page.goto('/');
        await page.waitForSelector('[data-testid="channel-page"]', { state: 'visible' });
        await page.waitForSelector('[role="status"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('channel-empty.png', {
            fullPage: false,
        });
    });

    test('model list empty state', async ({ page }) => {
        await setActiveNav(page, 'model');
        await mockEndpoint(page, '**/api/v1/model/list', []);
        await mockEndpoint(page, '**/api/v1/model/channel', []);
        await mockEndpoint(page, '**/api/v1/model/upstream-prices', []);
        await page.goto('/');
        await page.waitForSelector('[data-testid="model-page"]', { state: 'visible' });
        await page.waitForSelector('[role="status"]', { state: 'visible' });
        await expect(page).toHaveScreenshot('model-empty.png', {
            fullPage: false,
        });
    });

    test('channel list loading skeleton', async ({ page }) => {
        await setActiveNav(page, 'channel');
        // 延迟响应使骨架屏保持可见，便于截图
        await page.route('**/api/v1/channel/list', async (route) => {
            await new Promise((resolve) => setTimeout(resolve, 5000));
            await route.continue();
        });
        await page.goto('/');
        await page.waitForSelector('[data-testid="channel-page"]', { state: 'visible' });
        // 等待骨架屏渲染（animate-pulse 元素出现）
        await page.waitForTimeout(500);
        await expect(page).toHaveScreenshot('channel-loading-skeleton.png', {
            fullPage: false,
        });
    });

    test('model list loading skeleton', async ({ page }) => {
        await setActiveNav(page, 'model');
        await page.route('**/api/v1/model/list', async (route) => {
            await new Promise((resolve) => setTimeout(resolve, 5000));
            await route.continue();
        });
        await page.route('**/api/v1/model/channel', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ success: true, data: [] }),
            });
        });
        await page.route('**/api/v1/model/upstream-prices', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ success: true, data: [] }),
            });
        });
        await page.goto('/');
        await page.waitForSelector('[data-testid="model-page"]', { state: 'visible' });
        await page.waitForTimeout(500);
        await expect(page).toHaveScreenshot('model-loading-skeleton.png', {
            fullPage: false,
        });
    });
});
