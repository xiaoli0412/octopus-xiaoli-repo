// Octopus Load Test - Embeddings
//
// 压测目标: POST /v1/embeddings
// 鉴权: Authorization: Bearer ${API_KEY}
//
// 环境变量:
//   BASE_URL  - Octopus 网关地址(默认 http://localhost:1088)
//   API_KEY   - 客户端 API Key(必填)
//
// 运行示例:
//   k6 run -e API_KEY=sk-xxxx load-test/k6-embeddings.js

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:1088';
const API_KEY = __ENV.API_KEY || '';

const EMBED_URL = `${BASE_URL}/v1/embeddings`;

const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${API_KEY}`,
};

const requestBody = JSON.stringify({
    model: 'text-embedding-3-small',
    input: 'test text',
});

export const options = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '30s', target: 30 },
        { duration: '30s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        // P95 < 2s,错误率 < 5%
        http_req_duration: ['p(95)<2000'],
        http_req_failed: ['rate<0.05'],
    },
    tags: {
        test: 'embeddings',
    },
};

export default function () {
    const res = http.post(EMBED_URL, requestBody, { headers });
    check(res, {
        'status is 200': (r) => r.status === 200,
        'has data array': (r) => {
            try {
                const body = r.json();
                return Array.isArray(body.data) && body.data.length > 0;
            } catch (_) {
                return false;
            }
        },
        'has embedding vector': (r) => {
            try {
                const body = r.json();
                return Array.isArray(body.data) &&
                    body.data.length > 0 &&
                    Array.isArray(body.data[0].embedding) &&
                    body.data[0].embedding.length > 0;
            } catch (_) {
                return false;
            }
        },
    });

    sleep(0.2);
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
        '\n===== k6 embeddings summary =====',
        `http_reqs:           ${fmt(m.http_reqs)}`,
        `http_req_duration:   ${fmt(m.http_req_duration)}`,
        `http_req_failed:     ${fmt(m.http_req_failed)}`,
        `iterations:          ${fmt(m.iterations)}`,
        `vus:                 ${fmt(m.vus)}`,
        '=================================\n',
    ].join('\n');
}
