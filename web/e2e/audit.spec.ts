import { test, expect } from '@playwright/test';
import { setupAuth, setActiveNav, freezeTime, mockEndpoint } from './helpers/auth';
import { mockUserStatus } from './fixtures/common';

const mockAuditList = {
    list: [
        {
            id: 1,
            user_id: 1,
            username: 'admin',
            action: 'create',
            resource_type: 'channel',
            resource_id: '1',
            resource_name: 'OpenAI 官方',
            before_json: '',
            after_json: '{"name":"OpenAI 官方","type":0,"enabled":true,"api_key":"***"}',
            ip: '127.0.0.1',
            user_agent: 'Mozilla/5.0 (mock)',
            created_at: '2026-06-27T10:00:00Z',
        },
        {
            id: 2,
            user_id: 1,
            username: 'admin',
            action: 'update',
            resource_type: 'group',
            resource_id: '2',
            resource_name: '默认分组',
            before_json: '{"name":"默认分组","mode":"round_robin"}',
            after_json: '{"name":"默认分组","mode":"failover"}',
            ip: '192.168.1.10',
            user_agent: 'Mozilla/5.0 (mock)',
            created_at: '2026-06-27T11:00:00Z',
        },
        {
            id: 3,
            user_id: 2,
            username: 'operator',
            action: 'delete',
            resource_type: 'apikey',
            resource_id: '5',
            resource_name: '开发测试 Key',
            before_json: '{"id":5,"name":"开发测试 Key","api_key":"***"}',
            after_json: '',
            ip: '10.0.0.5',
            user_agent: 'curl/8.0',
            created_at: '2026-06-27T12:00:00Z',
        },
    ],
    total: 3,
    page: 1,
    page_size: 20,
};

const mockAuditDetail = mockAuditList.list[1];

const mockAuditListFiltered = {
    list: [mockAuditList.list[1]],
    total: 1,
    page: 1,
    page_size: 20,
};

test.describe('audit log page', () => {
    test.beforeEach(async ({ page }) => {
        await freezeTime(page);
        await setupAuth(page);
        await setActiveNav(page, 'audit');
        await mockEndpoint(page, '**/api/v1/user/status', mockUserStatus);
        await mockEndpoint(page, '**/api/v1/audit/list**', mockAuditList);
        await mockEndpoint(page, '**/api/v1/audit/2', mockAuditDetail);
    });

    test('loads audit list and displays rows', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="audit-page"]', { state: 'visible' });

        await page.waitForSelector('[data-testid="audit-row-1"]', { state: 'visible' });
        await expect(page.locator('[data-testid="audit-row-1"]')).toBeVisible();
        await expect(page.locator('[data-testid="audit-row-2"]')).toBeVisible();
        await expect(page.locator('[data-testid="audit-row-3"]')).toBeVisible();

        await expect(page.getByText('admin').first()).toBeVisible();
        await expect(page.getByText('operator')).toBeVisible();
    });

    test('opens detail dialog on row click', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="audit-row-2"]', { state: 'visible' });

        await page.click('[data-testid="audit-row-2"]');

        const dialog = page.locator('[data-slot="dialog-content"]');
        await expect(dialog).toBeVisible();
        await expect(dialog.getByText('差异对比')).toBeVisible();
        await expect(dialog.getByText('默认分组')).toBeVisible();
    });

    test('filters by action via select', async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('[data-testid="audit-row-1"]', { state: 'visible' });

        await mockEndpoint(page, '**/api/v1/audit/list**', mockAuditListFiltered);

        await page.locator('[data-slot="select-trigger"]').first().click();
        await page.getByRole('option', { name: 'update' }).click();
        await page.getByRole('button', { name: '筛选' }).click();

        await page.waitForSelector('[data-testid="audit-row-2"]', { state: 'visible' });
        await expect(page.locator('[data-testid="audit-row-1"]')).toHaveCount(0);
        await expect(page.locator('[data-testid="audit-row-2"]')).toBeVisible();
    });

    test('shows empty state when no audit logs', async ({ page }) => {
        await mockEndpoint(page, '**/api/v1/audit/list**', { list: [], total: 0, page: 1, page_size: 20 });

        await page.goto('/');
        await page.waitForSelector('[data-testid="audit-page"]', { state: 'visible' });

        await expect(page.getByText('暂无审计日志')).toBeVisible();
    });

    test('shows pagination info when total exceeds page size', async ({ page }) => {
        const largeList = {
            list: Array.from({ length: 20 }, (_, i) => ({
                ...mockAuditList.list[0],
                id: i + 10,
                created_at: `2026-06-27T${String(10 + i).padStart(2, '0')}:00:00Z`,
            })),
            total: 55,
            page: 1,
            page_size: 20,
        };
        await mockEndpoint(page, '**/api/v1/audit/list**', largeList);

        await page.goto('/');
        await page.waitForSelector('[data-testid="audit-row-10"]', { state: 'visible' });

        await expect(page.getByText(/共 55 条/)).toBeVisible();
        await expect(page.getByText(/第 1 \/ 3 页/)).toBeVisible();
    });
});
