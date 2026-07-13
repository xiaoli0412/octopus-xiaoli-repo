# Octopus 负载测试基线报告

> 状态:**待实际环境执行后填写**
>
> 本文档为模板,所有"实际结果"列在生产/预发环境完成 `load-test/run.sh` 全量压测后填入。
> 基线指标基于 Octopus 的架构特点(Rust FFI 加速 token 计数、出站 HTTP 连接池调优、Gin + GORM)给出预期范围,实际值以压测为准。

---

## 1. 测试环境说明

| 项目 | 配置 |
|------|------|
| Octopus 版本 | _(待填写,见 `VERSION` 文件)_ |
| 部署方式 | _(Docker / 二进制 / k8s)_ |
| 实例规格 | _(CPU / 内存 / 核数)_ |
| 实例数 | _(待填写)_ |
| 数据库类型 | _(sqlite / mysql / postgresql)_ |
| 数据库规格 | _(待填写)_ |
| 操作系统 | _(待填写)_ |
| 内核版本 | _(待填写)_ |
| Docker 版本 | _(待填写,如适用)_ |
| k6 版本 | _(待填写,`k6 version`)_ |
| 压测机规格 | _(与 Octopus 实例分离时填写)_ |
| 压测机与被测机网络 | _(同主机 / 同机房 / 跨地域)_ |

### 1.1 Octopus 配置(关键项)

| 配置项 | 值 | 说明 |
|--------|----|----|
| `OCTOPUS_HTTP_MAX_IDLE_CONNS` | 200(默认) | 出站 HTTP 连接池上限 |
| `OCTOPUS_HTTP_MAX_CONNS_PER_HOST` | 100(默认) | 单上游 host 连接上限 |
| `OCTOPUS_HTTP_MAX_IDLE_CONNS_PER_HOST` | 100(默认) | 单上游 host 空闲连接上限 |
| `OCTOPUS_HTTP_IDLE_CONN_TIMEOUT` | 90s(默认) | 空闲连接超时 |
| `OCTOPUS_SHUTDOWN_TIMEOUT` | 30s(默认) | 优雅停机超时 |
| `OCTOPUS_LOG_LEVEL` | info | 日志级别(压测时建议 warn 以减少 IO 干扰) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | _(未设置)_ | OTLP trace 导出端点(压测时建议关闭以隔离变量) |

### 1.2 上游渠道配置

| 渠道 | provider_type | base_url | 是否启用熔断 | 备注 |
|------|---------------|----------|--------------|------|
| _(示例)OpenAI_ | openai | https://api.openai.com | 是 | 真实上游,延迟受外网影响 |
| _(示例)Mock_ | openai | http://mock-upstream:8080 | 否 | 本地 mock,用于隔离网关自身性能 |

> **建议**:基线测试应优先使用 **mock 上游**(固定 50ms 延迟、固定响应体),以隔离 Octopus 网关自身的转发/转换开销。真实上游测试用于端到端 SLO 验证。

---

## 2. 测试场景

### 2.1 chat-completions(中继 API)

| 项 | 值 |
|----|----|
| 端点 | `POST /v1/chat/completions` |
| 鉴权 | `Authorization: Bearer ${API_KEY}` |
| 请求体(非流式) | `{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}],"max_tokens":10}` |
| 请求体(流式) | 同上,追加 `"stream":true` |
| 压测模型 | `gpt-3.5-turbo` |
| VU 阶梯 | 10 → 30 → 50 → 50 → 0,每段 30s,总时长 ~2.5min |
| thresholds | `http_req_duration p(95)<2000ms`,`http_req_failed rate<0.05` |
| 脚本 | `load-test/k6-chat-completions.js` |

### 2.2 embeddings(中继 API)

| 项 | 值 |
|----|----|
| 端点 | `POST /v1/embeddings` |
| 鉴权 | `Authorization: Bearer ${API_KEY}` |
| 请求体 | `{"model":"text-embedding-3-small","input":"test text"}` |
| VU 阶梯 | 10 → 30 → 50 → 50 → 0,每段 30s |
| thresholds | `http_req_duration p(95)<2000ms`,`http_req_failed rate<0.05` |
| 脚本 | `load-test/k6-embeddings.js` |

### 2.3 stats(管理 API)

| 项 | 值 |
|----|----|
| 端点 | `GET /api/v1/stats/total`、`/api/v1/stats/hourly`、`/api/v1/stats/daily` |
| 鉴权 | `Authorization: Bearer ${JWT_TOKEN}` |
| VU | 10 持续 1 分钟 |
| thresholds | `http_req_duration p(95)<500ms`,`http_req_failed rate<0.01` |
| 脚本 | `load-test/k6-stats.js` |

---

## 3. 预期基线指标

> 预期值基于以下架构假设:
> - **Rust FFI token 计数**:相对纯 Go 实现,token 计数开销可忽略,不会成为瓶颈
> - **出站 HTTP 连接池调优**:`max_conns_per_host=100` 默认值在 50 VU 并发下应能复用连接,无新建连接抖动
> - **SQLite 单机**:写并发受 WAL 模式限制,stats 写入走内存聚合 + 后台批量落盘,前端读路径应稳定
> - **mock 上游 50ms 固定延迟**:网关端到端延迟 ≈ 上游延迟 + 转换开销(预期 < 50ms 增量)

### 3.1 chat-completions 预期基线

| 指标 | 预期范围(mock 上游) | 预期范围(真实上游) | 实际结果 |
|------|----------------------|----------------------|----------|
| RPS(峰值) | 200 - 600 | 50 - 200 | _待填写_ |
| P50 延迟 | 60 - 120 ms | 300 - 800 ms | _待填写_ |
| P95 延迟 | 100 - 300 ms | 800 - 1800 ms | _待填写_ |
| P99 延迟 | 200 - 500 ms | 1200 - 2500 ms | _待填写_ |
| 错误率 | < 0.1% | < 2% | _待填写_ |
| 峰值 VU | 50 | 50 | _待填写_ |
| thresholds 通过 | 是 | 是 | _待填写_ |

### 3.2 embeddings 预期基线

| 指标 | 预期范围(mock 上游) | 预期范围(真实上游) | 实际结果 |
|------|----------------------|----------------------|----------|
| RPS(峰值) | 300 - 800 | 80 - 300 | _待填写_ |
| P50 延迟 | 50 - 100 ms | 200 - 600 ms | _待填写_ |
| P95 延迟 | 80 - 200 ms | 600 - 1500 ms | _待填写_ |
| P99 延迟 | 150 - 400 ms | 1000 - 2000 ms | _待填写_ |
| 错误率 | < 0.1% | < 2% | _待填写_ |
| thresholds 通过 | 是 | 是 | _待填写_ |

### 3.3 stats 预期基线

| 指标 | 预期范围 | 实际结果 |
|------|----------|----------|
| RPS | 200 - 500(3 端点轮询,10 VU) | _待填写_ |
| P50 延迟 | 5 - 30 ms | _待填写_ |
| P95 延迟 | 30 - 200 ms | _待填写_ |
| P99 延迟 | 80 - 400 ms | _待填写_ |
| 错误率 | 0% | _待填写_ |
| thresholds 通过 | 是 | _待填写_ |

---

## 4. 实际结果(待填写)

> 执行命令:
> ```bash
> cd load-test
> BASE_URL=http://<octopus-host>:1088 API_KEY=sk-xxxx JWT_TOKEN=eyJ... ./run.sh all
> ```
>
> 结果文件:
> - JSON 明细: `results/k6-<scenario>-<timestamp>.json`
> - 汇总摘要: `results/k6-<scenario>-summary-<timestamp>.json`
> - 文本汇总: `results/summary-<timestamp>.txt`

### 4.1 chat-completions 实测

| 指标 | 实测值 | 是否达标 | 备注 |
|------|--------|----------|------|
| RPS(峰值) | _待填写_ | _待填写_ | |
| P50 延迟 | _待填写_ | _待填写_ | |
| P95 延迟 | _待填写_ | _待填写_ | 阈值 < 2000ms |
| P99 延迟 | _待填写_ | — | |
| 错误率 | _待填写_ | _待填写_ | 阈值 < 5% |

### 4.2 embeddings 实测

| 指标 | 实测值 | 是否达标 | 备注 |
|------|--------|----------|------|
| RPS(峰值) | _待填写_ | _待填写_ | |
| P50 延迟 | _待填写_ | _待填写_ | |
| P95 延迟 | _待填写_ | _待填写_ | 阈值 < 2000ms |
| P99 延迟 | _待填写_ | — | |
| 错误率 | _待填写_ | _待填写_ | 阈值 < 5% |

### 4.3 stats 实测

| 指标 | 实测值 | 是否达标 | 备注 |
|------|--------|----------|------|
| RPS | _待填写_ | _待填写_ | |
| P50 延迟 | _待填写_ | _待填写_ | |
| P95 延迟 | _待填写_ | _待填写_ | 阈值 < 500ms |
| P99 延迟 | _待填写_ | — | |
| 错误率 | _待填写_ | _待填写_ | 阈值 < 1% |

---

## 5. 性能优化记录

> 每轮压测后,将调优动作与对应指标变化记录于此,形成性能演进轨迹。

| 轮次 | 日期 | 调优动作 | 指标变化 | 备注 |
|------|------|----------|----------|------|
| 1 | _待填写_ | 基线建立 | — | |
| 2 | _待填写_ | _(如:调大 `max_conns_per_host`)_ | _(如:P95 下降 X%)_ | |
| 3 | _待填写_ | _(如:`log_level=warn`)_ | | |

---

## 6. 附录:关键性能影响因素

### 6.1 Rust FFI 加速 token 计数

Octopus 通过 `internal/rustbridge/`(Rust 静态库 + Go FFI)加速 token 计数。预期:
- 在 50 VU 并发下,token 计数开销 < 总延迟 1%
- 若 Rust FFI 未编译(纯 Go 回退路径),token 计数开销可能上升至 5-10%
- 验证方式:`curl http://<host>:1088/metrics | grep octopus_token_throughput_total` 观察 token 吞吐

### 6.2 出站 HTTP 连接池调优

| 配置项 | 默认 | 调优建议 |
|--------|------|----------|
| `max_idle_conns` | 200 | 高并发(>100 VU)可调至 500 |
| `max_conns_per_host` | 100 | 单上游瓶颈时调至 200-500,注意上游限流 |
| `max_idle_conns_per_host` | 100 | 与 `max_conns_per_host` 保持一致 |
| `idle_conn_timeout` | 90s | 长连接上游可调至 5min 减少重建 |
| `response_header_timeout` | 30s | 慢上游可调至 60s,但会延长故障检测 |

### 6.3 数据库选型影响

| 数据库 | 适用场景 | 压测预期 |
|--------|----------|----------|
| SQLite | 单实例、低并发 | 写并发 > 100 QPS 时可能成为瓶颈;读路径(stats)走内存缓存,影响小 |
| MySQL | 多实例、中高并发 | 推荐生产使用;连接池建议 20-50 |
| PostgreSQL | 多实例、复杂查询 | 推荐;JSONB 对 relay_log 查询友好 |

---

## 7. 变更记录

| 日期 | 作者 | 变更 |
|------|------|------|
| _(待填写)_ | _(待填写)_ | 初始基线报告创建 |
