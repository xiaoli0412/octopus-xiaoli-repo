# 成功案例吸收与反模式规避（调研摘录）

> 本文件是对 canonical plan 第 17 节的调研支撑，不替代 canonical plan。

## 可吸收的方向

### LiteLLM
- 统一 OpenAI-compatible 入口
- 多 key / 多 provider fallback 与 observability
- 层次化 router settings（key/team/global）值得借鉴

### Portkey
- fallback / loadbalance / conditional 策略结构化配置值得借鉴
- observability、预算、可靠性控制值得借鉴

### Bifrost / kgateway / AgentGateway
- priority group / sequential fallback / provider-level backup 链值得借鉴
- "请求失败后按顺序 fallback"、"优先级组"、"可观察性"值得借鉴

## 明确不照搬的反模式

1. 超级模型池 / 超大 catalog 默认全量开放
2. 成本驱动自动排序覆盖用户 priority
3. 完全黑箱自动选模
4. 过度复杂的动态路由，导致系统难理解、难维护、难在低配机器上运行
