import { test, expect, type Route } from '@playwright/test';
import { setupAuth, setActiveNav, mockEndpoint } from './helpers/auth';
import {
    mockUserStatus,
    mockChannelList,
    mockStatsTotal,
    mockStatsDaily,
    mockStatsHourly,
    mockStatsTokenBreakdown,
} from './fixtures/common';

/**
 * 实时更新 (SSE) e2e 测试
 *
 * 覆盖场景：
 * 1. SSE 连接建立后，mock 推送统计快照，验证仪表盘收到实时数据
 * 2. SSE 消息接收：mock 推送日志，验证日志列表出现新条目
 * 3. 降级轮询：mock EventSource 持续失败，验证前端降级到 TanStack Query 轮询
 */

const mockLogList = [
    {
        id: 1001,
        time: 1750000000,
        request_model_name: 'gpt-4',
        channel: 1,
        channel_name: 'OpenAI',
        actual_model_name: 'gpt-4',
        input_tokens: 100,
        output_tokens: 50,
        cache_read_tokens: 0,
        cache_write_tokens: 0,
        ftut: 200,
        use_time: 1500,
        cost: 0.002,
        request_content: '',
        response_content: '',
        error: '',
    },
];

test.describe('realtime SSE', () => {
    test.beforeEach(async ({ page }) => {
        page.on('console', (msg) => console.log(`[console ${msg.type()}]`, msg.text()));
        page.on('pageerror', (err) => console.error('[pageerror]', err));
        await setupAuth(page);
        await setActiveNav(page, 'home');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/channel/list', mockChannelList);
        await mockEndpoint(page, '**/api/v1/stats/total', mockStatsTotal);
        await mockEndpoint(page, '**/api/v1/stats/daily', mockStatsDaily);
        await mockEndpoint(page, '**/api/v1/stats/hourly', mockStatsHourly);
        await mockEndpoint(page, '**/api/v1/stats/token-breakdown**', mockStatsTokenBreakdown);
        await mockEndpoint(page, '**/api/v1/log/list**', mockLogList);
    });

    test('SSE stats 连接建立并接收消息', async ({ page }) => {
        // 拦截 SSE stats 端点，返回一条初始快照
        await page.route('**/api/v1/stream/stats**', async (route: Route) => {
            await route.fulfill({
                status: 200,
                contentType: 'text/event-stream',
                headers: {
                    'Cache-Control': 'no-cache',
                    Connection: 'keep-alive',
                },
                body: 'data: {"total":{"id":1,"input_token":500,"output_token":300,"input_cost":0.01,"output_cost":0.02,"wait_time":1200,"request_success":10,"request_failed":1},"today":{"date":"20260711","input_token":500,"output_token":300,"input_cost":0.01,"output_cost":0.02,"wait_time":1200,"request_success":10,"request_failed":1},"hourly":{"hour":10,"date":"20260711","input_token":500,"output_token":300,"input_cost":0.01,"output_cost":0.02,"wait_time":1200,"request_success":10,"request_failed":1}}\n\n',
            });
        });

        await page.goto('/');
        await page.waitForSelector('[data-testid="home-page"]', { state: 'visible' });
        await page.waitForSelector('[data-testid="home-total-summary-card-0"]', { state: 'visible' });

        // SSE 连接成功后应显示实时指示器
        await expect(page.locator('[data-testid="home-total-live-indicator"]')).toBeVisible({ timeout: 5000 });
    });

    test('SSE 日志推送接收', async ({ page }) => {
        await setActiveNav(page, 'log');
        // 拦截 SSE logs 端点，返回一条新日志
        await page.route('**/api/v1/stream/logs**', async (route: Route) => {
            await route.fulfill({
                status: 200,
                contentType: 'text/event-stream',
                headers: {
                    'Cache-Control': 'no-cache',
                    Connection: 'keep-alive',
                },
                body: 'data: {"id":2002,"time":1750000100,"request_model_name":"claude-3","channel":1,"channel_name":"Anthropic","actual_model_name":"claude-3","input_tokens":80,"output_tokens":40,"cache_read_tokens":0,"cache_write_tokens":0,"ftut":150,"use_time":900,"cost":0.001,"request_content":"","response_content":"","error":""}\n\n',
            });
        });

        await page.goto('/');
        await page.waitForSelector('[data-testid="log-pause-resume-btn"]', { state: 'visible' });

        // 暂停/恢复按钮应可见
        await expect(page.locator('[data-testid="log-pause-resume-btn"]')).toBeVisible();

        // 新日志应出现在列表中（SSE 推送的 id=2002）
        await expect(page.locator('text=claude-3')).toBeVisible({ timeout: 5000 });
    });

    test('SSE 失败后降级到轮询', async ({ page }) => {
        await setActiveNav(page, 'log');
        // 拦截 SSE logs 端点，返回 502 使 EventSource 触发 onerror
        await page.route('**/api/v1/stream/logs**', async (route: Route) => {
            await route.fulfill({ status: 502, body: 'Bad Gateway' });
        });
        // 拦截 SSE stats 端点，同样失败
        await page.route('**/api/v1/stream/stats**', async (route: Route) => {
            await route.fulfill({ status: 502, body: 'Bad Gateway' });
        });

        await page.goto('/');
        await page.waitForSelector('[data-testid="home-total-summary-card-0"]', { state: 'visible' });

        // 连续失败后降级轮询仍能通过 TanStack Query 拿到初始数据
        await expect(page.locator('[data-testid="home-total-summary-card-0"]')).toBeVisible();

        // 日志列表仍能通过轮询加载历史数据
        await page.waitForSelector('[data-testid="log-pause-resume-btn"]', { state: 'visible' });
        // 初始日志（mockLogList 中的 gpt-4）应可见
        await expect(page.locator('text=gpt-4')).toBeVisible({ timeout: 10000 });
    });
});
