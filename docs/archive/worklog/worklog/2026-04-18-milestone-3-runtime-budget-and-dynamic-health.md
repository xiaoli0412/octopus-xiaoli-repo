# 2026-04-18 Milestone 3 Runtime Budget And Dynamic Health

## 任务信息

- 主线来源：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 对齐目标：里程碑 3 `策略与性能`
- 本轮范围：`internal/relay/*`、`internal/relay/balancer/*`、`internal/model/setting.go`、`internal/db/migrate/009.go`

## 开工前输入

- 复用本地资源：`AGENTS.md`、主 MD、既有 relay/balancer 测试、已有 race fallback / circuit breaker 实现。
- 本轮子 agent 使用模型：`gpt-5.4`
- 子 agent 职责边界：只读分析里程碑 3 测试缺口，不修改实现文件。
- 主线程职责：直接落地代码、跑测试、汇总结论并写回 worklog。

## 本轮完成

- 为里程碑 3 新增运行时设置项：
  - `dynamic_routing_health_enabled`
  - `race_global_budget`
  - `race_group_budget`
  - `race_channel_budget`
  - `race_key_budget`
  - `race_probe_budget`
- 在迁移阶段补默认值写入，保证旧库升级后有安全兜底。
- 新增 `internal/relay/budget.go`：实现并发竞速预算计数器，覆盖全局、group、channel、key、probe 五层预算。
- 新增 `internal/relay/dynamic.go`：实现动态健康调节骨架。
- 动态调节坚持硬规则：
  - 不改用户显式优先级顺序。
  - 只调 `race_after_fails`、`race_concurrency`、`circuit threshold` 这类阈值参数。
  - `paid / metered / per_request / per_quota / flat` 默认更保守。
  - `free / public` 允许更积极，但不越过用户显式把 `race_concurrency` 设为 `1` 的限制。
- `relay.go` 接入动态调节与并发预算：
  - race 触发阈值改为动态计算。
  - race 并发数改为动态计算。
  - 每个 race candidate 启动前先申请预算，不满足则跳过并记录 attempt。
- `balancer/circuit.go` 接入动态 circuit threshold 桥接，并保留原有熔断行为。

## 测试与验证

- `go test ./internal/relay/...`
  - 通过
- `go test ./internal/op/...`
  - 通过

新增或调整的关键测试点：

- `TestEffectiveDynamicRoutingTuningPreservesPriorityWhileAdjustingThresholds`
- `TestEffectiveRaceConcurrencyPaidAndFreeProfiles`
- `TestRunRaceFallbackSkipsCandidateWhenRaceBudgetExhausted`
- 更新 `TestShouldEscalateToRace` 以匹配动态调节签名

## 与 MD 的对齐结果

- 已实做：
  - 成本感知探测的一部分
  - 并发预算的一部分
  - 动态调整只调阈值、不改用户顺序
- 已有明确证据：
  - paid/default 路径不会默认激进并发
  - race winner 确定后会取消其余请求
  - race 预算耗尽时会跳过候选而不是无上限放大并发

## 未完成 / 下一步

- 还没有做配置面的完整前后端暴露：`channel/group` 页面尚未提供里程碑 3 新设置的显式编辑能力。
- 动态调整目前主要体现在 relay 请求路径里的即时策略；后台每日任务现阶段只是 summary scan，不做 runtime mutation。
- 预算目前已覆盖 race 路径，尚未扩展为更通用的全局请求级预算器。
- 还缺更完整的 `paid / free / public` 执行级差异测试，尤其是 handler 级整链路验证。

## 子 agent 与本地资源记录

- 子 agent：`019d9cca-2e61-7161-8d1e-d90f6cb61064`
- 模型：`gpt-5.4`
- 职责：只读分析里程碑 3 测试缺口，给出最该补的测试建议。
- 采用结论：优先补“动态阈值不改排序”和“并发预算硬上限”的测试证据。
