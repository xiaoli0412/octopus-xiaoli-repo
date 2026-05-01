# P0-A Relay Live Path Dynamic Circuit Threshold Closure

> 日期：2026-04-23
>
> 目标：在上一轮已完成的 relay live-path 重试 / failover / 同渠道 key fallback 修复基础上，继续把 route-target 驱动的动态熔断阈值真正接入运行时，并把本轮结论同步回主线记忆。

---

## 1. 任务信息

- 任务名称：P0-A relay live path 动态熔断阈值收口
- 日期：2026-04-23
- 当前阶段：P0-A 后端主链继续收口
- 对应 milestone：relay live-path 与 route-target/runtime 语义对齐

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：动态路由主线、双主线规则、当前 P0-A / P0-B 排序
- 上一个相关 worklog：[2026-04-22-all-automation-workflow-memory-plan-normalization.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-all-automation-workflow-memory-plan-normalization.md)
- 本次任务目标：
  - 补写上一轮 relay / smoke 闭环结果到 repo worklog 与 automation memory
  - 继续推进 P0-A，确认 live path 是否还存在未消费的 route-target/runtime 语义
  - 在不破坏当前绿测面的前提下，收口一条可安全落地的 runtime 语义
- 本次已盘点本地资源：
  - [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)
  - [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
  - [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md)
  - [DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md)
  - [FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md)
  - [2026-04-22-215411-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-22-215411-octopus-repo-complete-audit.md)
  - `C:\Users\李昊桐\.codex\automations\octopus\memory.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：当前任务集中在 `internal/relay` 主链和记忆回写，写入点高度耦合，主线程直接收口更稳妥

## 3. 本次实现

### 3.1 已确认并保留的上一轮闭环

- `internal/relay/relay.go` live path 已真实消费：
  - 同渠道多 key fallback
  - `RetryRounds`
  - `RetryDelayMs`
  - `FailoverWindowSec`
  - stale route item skip 语义
- `scripts/smoke-win-backend.ps1` 已能自适应 repo-local `GOCACHE/GOTMPDIR`，Windows smoke 默认执行可通过

### 3.2 本轮新增闭环

- 把 route-target 驱动的动态熔断阈值真正接入 live relay path：
  - 在 [dynamic.go](/D:/GPT-codex/octopus_repo/internal/relay/dynamic.go) 中为 `balancer.RelayEffectiveCircuitThresholdShim` 提供运行时实现
  - 运行时通过 `op.ChannelKeyGet(keyID)` 读取当前 key，再调用 `effectiveCircuitThresholdForRelay(...)`
  - 这样 paid/metered 与 free/public 的 circuit threshold 差异，不再只停留在 helper/test 层，而是实际影响 live `RecordFailure` / `IsTripped`
- 新增 handler 级回归测试，证明 route-target policy 已进入真实请求路径：
  - [relay_more_test.go](/D:/GPT-codex/octopus_repo/internal/relay/relay_more_test.go)
  - 新用例：`TestHandlerUsesDynamicCircuitThresholdFromRouteTargetPolicy`
  - 验证点：全局 threshold 为 `5` 时，`per_request` route-target 仍会在第 `4` 次失败后触发 circuit，后续请求不再打到上游而直接记录 `AttemptCircuitBreak`

## 4. 本轮未做但已明确的边界

- `runRaceFallback` / `finalizeRaceFallbackSuccess` 仍未接入 `Handler` live path。
- 本轮没有强行接 race fallback，原因：
  - 当前 helper 仍需先解决“完整消费剩余候选而不重复探测”的运行时语义
  - 需要明确是否覆盖同渠道剩余 key 的 route-target 粒度
  - `finalizeRaceFallbackSuccess` 当前只适合非流式 JSON 响应，直接接入流式路径风险高
- 因此本轮选择先把“动态熔断阈值进入 live path”这条低风险、高价值语义收口，再把 race fallback 保留为下一闭环

## 5. 验证

- `gofmt -w internal/relay/dynamic.go internal/relay/relay_more_test.go`
- `go test ./internal/relay -count=1`
- `go test ./... -count=1`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`

结果：全部通过

## 6. 风险与兼容性

- 本轮改动不扩展新的配置面，只让已存在 route-target/runtime 语义更真实地影响 live path
- 仍需继续警惕的主线风险：
  - race fallback helper 仍未进入 handler live path
  - `next build` 宿主环境 `spawn EPERM`
  - Linux / Docker / browser 证据仍依赖可执行环境

## 7. 收工记录

- worklog 是否更新：是
- 本次使用的本地资源 / skills / 记忆上下文：主规划、动态路由需求、最新审查报告、automation memory、`internal/relay` 现有 helper/tests
- 本次使用这些资源得到的关键结论：
  - P0-A 不是继续扩新功能，而是要优先把 live path 与已暴露语义对齐
  - race fallback 接线仍有边界条件，不能为了“看起来完整”破坏当前绿测面
  - 动态熔断阈值属于已经设计好但未完全消费的 runtime 语义，适合作为本轮稳妥收口点
- 遗留项：
  - 下一轮继续评估并推进 `runRaceFallback` 的 live-path 受控接线
  - P0-B 仍需继续 browser/screenshot 证据收口
  - P1 备份导入导出闭环保持待办

