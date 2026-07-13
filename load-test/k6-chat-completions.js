// Octopus Load Test - Chat Completions (非流式 + 流式)
//
// 压测目标: POST /v1/chat/completions
// 鉴权: Authorization: Bearer ${API_KEY}
//
// 环境变量:
//   BASE_URL  - Octopus 网关地址(默认 http://localhost:1088)
//   API_KEY   - 客户端 API Key(必填,见 /api/v1/apikey 管理)
//
// 运行示例:
//   k6 run -e API_KEY=sk-xxxx load-test/k6-chat-completions.js
//   k6 run -e BASE_URL=http://10.0.0.5:1088 -e API_KEY=sk-xxxx load-test/k6-chat-completions.js

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:1088';
const API_KEY = __ENV.API_KEY || '';

const CHAT_URL = `${BASE_URL}/v1/chat/completions`;

const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${API_KEY}`,
};

// 非流式请求体
const nonStreamBody = JSON.stringify({
    model: 'gpt-3.5-turbo',
    messages: [{ role: 'user', content: 'Hello' }],
    max_tokens: 10,
});

// 流式请求体(仅 stream 字段不同)
const streamBody = JSON.stringify({
    model: 'gpt-3.5-turbo',
    messages: [{ role: 'user', content: 'Hello' }],
    max_tokens: 10,
    stream: true,
});

export const options = {
    stages: [
        { duration: '30s', target: 10 },   // 30s 内 ramp 到 10 VU
        { duration: '30s', target: 30 },   // 30s 内 ramp 到 30 VU
        { duration: '30s', target: 50 },   // 30s 内 ramp 到 50 VU
        { duration: '30s', target: 50 },   // 保持 50 VU 30s
        { duration: '30s', target: 0 },    // 30s 内 ramp 到 0
    ],
    thresholds: {
        // P95 < 2s,错误率 < 5%
        http_req_duration: ['p(95)<2000'],
        http_req_failed: ['rate<0.05'],
    },
    tags: {
        test: 'chat-completions',
    },
};

export default function () {
    // 非流式请求
    const nonStreamRes = http.post(CHAT_URL, nonStreamBody, { headers });
    check(nonStreamRes, {
        'non-stream status is 200': (r) => r.status === 200,
        'non-stream has choices': (r) => {
            try {
                const body = r.json();
                return Array.isArray(body.choices) && body.choices.length > 0;
            } catch (_) {
                return false;
            }
        },
    });

    sleep(0.2);

    // 流式请求(SSE)
    const streamRes = http.post(CHAT_URL, streamBody, { headers });
    check(streamRes, {
        'stream status is 200': (r) => r.status === 200,
        'stream content-type is event-stream or json': (r) => {
            const ct = (r.headers['Content-Type'] || r.headers['content-type'] || '').toLowerCase();
            return ct.includes('event-stream') || ct.includes('application/json');
        },
        'stream body contains data prefix': (r) => {
            // SSE 响应体应以 "data: " 前缀返回 chunk
            return typeof r.body === 'string' && (r.body.includes('data:') || r.body.includes('"choices"'));
        },
    });

    sleep(0.2);
}

export function handleSummary(data) {
    return {
        stdout: textSummary(data),
    };
}

// 极简文本摘要(k6 内置 handleSummary,这里输出到 stdout)
function textSummary(data) {
    const m = data.metrics || {};
    const fmt = (v) => (v && typeof v === 'object' && 'values' in v) ? JSON.stringify(v.values) : String(v);
    return [
        '\n===== k6 chat-completions summary =====',
        `http_reqs:           ${fmt(m.http_reqs)}`,
        `http_req_duration:   ${fmt(m.http_req_duration)}`,
        `http_req_failed:     ${fmt(m.http_req_failed)}`,
        `iterations:          ${fmt(m.iterations)}`,
        `vus:                 ${fmt(m.vus)}`,
        '=======================================\n',
    ].join('\n');
}
