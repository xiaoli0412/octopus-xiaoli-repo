// Octopus Load Test - Stats(管理 API)
//
// 压测目标:
//   GET /api/v1/stats/total
//   GET /api/v1/stats/hourly
//   GET /api/v1/stats/daily
// 鉴权: Authorization: Bearer ${JWT_TOKEN}(管理 API 使用 JWT,非 API Key)
//
// 环境变量:
//   BASE_URL   - Octopus 网关地址(默认 http://localhost:1088)
//   JWT_TOKEN  - 管理 API JWT Token(必填,通过 POST /api/v1/user/login 获取)
//
// 运行示例:
//   k6 run -e JWT_TOKEN=eyJhbGciOi... load-test/k6-stats.js

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:1088';
const JWT_TOKEN = __ENV.JWT_TOKEN || '';

const STATS_TOTAL = `${BASE_URL}/api/v1/stats/total`;
const STATS_HOURLY = `${BASE_URL}/api/v1/stats/hourly`;
const STATS_DAILY = `${BASE_URL}/api/v1/stats/daily`;

const headers = {
    'Authorization': `Bearer ${JWT_TOKEN}`,
};

export const options = {
    // 10 VU 持续 1 分钟
    vus: 10,
    duration: '1m',
    thresholds: {
        // 管理 API 读路径,P95 < 500ms
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
    tags: {
        test: 'stats',
    },
};

export default function () {
    // 轮询三个 stats 端点,模拟监控面板的并发查询
    const totalRes = http.get(STATS_TOTAL, { headers });
    check(totalRes, {
        'total status is 200': (r) => r.status === 200,
        'total has success field': (r) => {
            try {
                const body = r.json();
                return body && (typeof body.data === 'object' || typeof body.success === 'boolean');
            } catch (_) {
                return false;
            }
        },
    });

    sleep(0.1);

    const hourlyRes = http.get(STATS_HOURLY, { headers });
    check(hourlyRes, {
        'hourly status is 200': (r) => r.status === 200,
    });

    sleep(0.1);

    const dailyRes = http.get(STATS_DAILY, { headers });
    check(dailyRes, {
        'daily status is 200': (r) => r.status === 200,
    });

    sleep(0.1);
}

export function handleSummary(data) {
    return {
        stdout: textSummary(data),
    };
}

function textSummary(data) {
    const m = data.metrics || {};
    const fmt = (v) => (v && typeof v === 'object' && 'values' in v) ? JSON.stringify(v.values) : String(v);
    return [
        '\n===== k6 stats summary =====',
        `http_reqs:           ${fmt(m.http_reqs)}`,
        `http_req_duration:   ${fmt(m.http_req_duration)}`,
        `http_req_failed:     ${fmt(m.http_req_failed)}`,
        `iterations:          ${fmt(m.iterations)}`,
        `vus:                 ${fmt(m.vus)}`,
        '============================\n',
    ].join('\n');
}
