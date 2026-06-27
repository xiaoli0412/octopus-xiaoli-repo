import { defineConfig, devices } from '@playwright/test';

const PORT = Number(process.env.PLAYWRIGHT_WEB_PORT ?? 3456);
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL ?? `http://127.0.0.1:${PORT}`;

export default defineConfig({
    testDir: './e2e',
    fullyParallel: false,
    forbidOnly: Boolean(process.env.CI),
    retries: process.env.CI ? 2 : 1,
    workers: 1,
    reporter: [['list'], ['html', { outputFolder: 'playwright-report' }]],
    use: {
        baseURL: BASE_URL,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        viewport: { width: 1280, height: 900 },
        contextOptions: {
            reducedMotion: 'reduce',
        },
    },
    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
    expect: {
        toHaveScreenshot: {
            maxDiffPixels: 50,
            threshold: 0.2,
        },
    },
    webServer: process.env.PLAYWRIGHT_BASE_URL
        ? undefined
        : {
            command: `cmd /c node_modules\\.bin\\serve.CMD out -l tcp://127.0.0.1:${PORT} --single`,
            url: `${BASE_URL}/login`,
            reuseExistingServer: !process.env.CI,
            timeout: 120 * 1000,
        },
});
