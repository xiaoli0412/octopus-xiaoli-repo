import type { Page } from '@playwright/test';

export const TEST_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIn0.mock';
export const FUTURE_EXPIRE_AT = '2099-12-31T23:59:59.000Z';

export async function setupAuth(page: Page) {
    await page.addInitScript(({ token, expireAt }: { token: string; expireAt: string }) => {
        const authStorage = JSON.stringify({
            state: {
                token,
                expireAt,
                isAPIKeyAuth: false,
                mustChangePassword: false,
                apiKeyStatus: null,
            },
            version: 0,
        });
        window.localStorage.setItem('auth-storage', authStorage);
    }, { token: TEST_TOKEN, expireAt: FUTURE_EXPIRE_AT });
}

export async function setActiveNav(page: Page, item: string) {
    await page.addInitScript((activeItem: string) => {
        const navStorage = JSON.stringify({
            state: {
                activeItem,
                prevItem: null,
                direction: 0,
            },
            version: 0,
        });
        window.localStorage.setItem('nav-storage', navStorage);
    }, item);
}

export async function freezeTime(page: Page, iso: string = '2026-06-27T12:00:00.000Z') {
    await page.addInitScript((fixedIso: string) => {
        const RealDate = window.Date;
        const fixed = new RealDate(fixedIso);
        // @ts-expect-error override Date for deterministic screenshots
        window.Date = class extends RealDate {
            constructor(...args: unknown[]) {
                if (args.length === 0) {
                    super(fixed);
                } else {
                    // @ts-expect-error dynamic constructor args
                    super(...args);
                }
            }
            static now() {
                return fixed.getTime();
            }
        };
    }, iso);
}

export async function mockEndpoint(page: Page, pattern: string | RegExp, payload: unknown, status = 200) {
    await page.route(pattern, async (route) => {
        await route.fulfill({
            status,
            contentType: 'application/json',
            body: JSON.stringify({ success: true, data: payload }),
        });
    });
}
