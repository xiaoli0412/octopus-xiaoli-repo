# P0-A Relay Live Race Fallback Handler Closure

时间：`2026-04-23`

## 本轮目标

继续推进已合并主线中的 `P0-A`，把此前只存在于 helper/test 层的 live race fallback 以受控方式接入真实 `Handler` 请求路径，同时保持现有 handler 级绿色链路不回退。

## 本轮收口内容

### 1. live race fallback 已接入真实请求主链路

本轮把 `internal/relay/relay.go` 中的真实 failover 主路径与 `runRaceFallbackWindow(...)` 接通，当前生效范围是：

- 仅 `GroupModeFailover`
- 仅非流式请求
- 仅在连续失败次数达到 `shouldEscalateToRace(...)` 阈值后
- 仅在当前 channel 已经没有剩余同模型 key 可继续顺序回退时
- 仍然受 `FailoverWindowSec` 总窗口约束

这样做的目的，是在不破坏已落地的“同渠道多 key 顺序回退”语义前提下，把跨 channel 的并发竞速能力接到真实主链路上。

### 2. race helper 从“单批次”收口为“窗口内批次消费”

此前的 `runRaceFallback(...)` 只会从 `iter.Index()+1` 之后抽取第一批 `maxConcurrency` 候选；如果这一批无 winner，真实 `Handler` 后续顺序循环还可能再次打到这些候选，存在重复探测风险。

本轮新增并接入：

- `runRaceFallbackWindow(...)`
- `buildRaceCandidateBatch(...)`
- `executeRaceCandidateBatch(...)`

现在 live path 会在 race 阶段按批次消费整个剩余候选窗口，并把 `consumedToIndex` 回传给 `Handler`，使顺序主循环跳过已被 race 消费过的候选，避免重复 probe。

### 3. winner 重复 attempt 记录风险已关闭

此前已识别的风险之一，是：

- `recordRaceAttemptOutcomes(...)` 会记录 `race fallback winner`
- `finalizeRaceFallbackSuccess(...)` 也会再次记录一次 winner

本轮新增 `attemptResult.AttemptRecorded`，并在 live race 成功路径中只保留一次 winner attempt 记录，handler 级回归已证明 winner 不会被重复 probe / 重复 attempt 记录。

### 4. race 失败/成功结果已补齐运行时副作用

为了避免“attempt 记了，但状态没动”的假象，本轮还补齐了 live race 批次中的运行时副作用：

- race winner 在 `finalizeRaceFallbackSuccess(...)` 中会更新 key 状态、channel 统计、sticky、circuit success
- race failure 会更新 key 状态、channel failure 统计、circuit failure、probe failure event
- race success / selected 仍会继续写 probe event

### 5. route-target / stale-route / 类型兼容校验已进入 race batch 构建

`buildRaceCandidateBatch(...)` 现在和主顺序路径保持一致地检查：

- channel enabled
- `SupportsModel(...)`
- `HasConfiguredKeyForModel(...)`
- circuit breaker
- request type compatibility
- route-target 是否允许 racing
- race budget 是否允许并发 probe

这保证 live race 不会绕过已经在主链路里收口好的安全检查。

## 新增回归验证

在 `internal/relay/relay_more_test.go` 中新增 handler 级用例：

- `TestHandlerEscalatesToLiveRaceAfterSequentialFailuresAndDoesNotDuplicateWinnerProbe`
  - 证明非流式 failover 请求在两次连续失败后进入 live race
  - 证明 winner 返回成功响应
  - 证明 winner 不会被顺序路径再次重复请求
  - 证明 relay log attempts 收口为“两次顺序失败 + 一次 race winner 成功”

- `TestHandlerStreamingRequestDoesNotEscalateToLiveRace`
  - 证明流式请求即便失败，也不会走 live race 分支

## 本轮验证

执行并通过：

```powershell
$env:GOCACHE='D:\GPT-codex\octopus_repo\.tools\gocache'
$env:GOTMPDIR='D:\GPT-codex\octopus_repo\.tools\gotmp'
$env:TEMP='D:\GPT-codex\octopus_repo\.tools\tmp'
$env:TMP='D:\GPT-codex\octopus_repo\.tools\tmp'
gofmt -w internal/relay/type.go internal/relay/probe.go internal/relay/dynamic_runtime.go internal/relay/relay.go internal/relay/relay_more_test.go
go test ./internal/relay -count=1
go test ./... -count=1
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1
```

结果：全部通过。

## 本轮后剩余边界

这轮把 live race fallback 以“安全最小范围”接进了主链路，但仍保留一个明确边界：

- 当前 live race 仍是“跨剩余 group candidates 的并发竞速”
- 尚未把“同 channel 的剩余 key”纳入 race 粒度

这是有意保留的，因为规划文档明确建议最终粒度应当收敛到 `(channel, key, model)`；本轮为了不破坏已经稳定的同渠道顺序 key fallback 语义，先把 live race 约束在“当前 channel 已无剩余 key 后，再对后续 candidates 竞速”的范围。

## 建议的下一入口

`P0-A` 的主 blocker 已从“live race helper 完全不可达”收窄为：

- 是否要把 same-channel remaining keys 也纳入 race 粒度
- 若纳入，应如何与当前已稳定的 key 顺序 fallback / route-target / budget 语义合并而不造成语义漂移

在这之前，可以把 backend 主线优先级下调一档，继续回到已约定的 `P0-B` 浏览器级前端主线；若下一轮仍继续后端，则建议专注于“same-channel key 级 race 粒度设计与验证”。
