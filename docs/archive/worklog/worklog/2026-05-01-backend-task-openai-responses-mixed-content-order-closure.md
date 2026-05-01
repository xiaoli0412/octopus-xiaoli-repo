# 2026-05-01 Backend Task OpenAI Responses Mixed Content Order Closure

## 1. 任务信息

- 任务名称：OpenAI Responses mixed text/image 下游顺序与保留收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `可用性核心` 中的 `OpenAI-compatible` 响应转换稳态化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 6 / 14` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-aggregate-mixed-message-delta-and-multipart-closure.md`
- 本次任务目标：把上轮 helper 已经能正确聚合的 `text -> image -> text` mixed content，继续推进到下游 `OpenAI Responses` 与 `Anthropic Messages` 转换链，避免输出顺序漂移或前文文本被图片分支挤压。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/inbound/anthropic/messages.go`
  - 现有 stream aggregate tests
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮集中在单一 transformer 链路，主线程收口更稳。

## 3. 本次候选任务与选择

- 候选任务 1：为 `OpenAI Responses` 下游响应转换补 mixed content 回归。
- 候选任务 2：修正 `Responses` 输出中 text/image 顺序漂移。
- 候选任务 3：为 `Anthropic Messages` 下游响应转换补同类顺序回归。
- 候选任务 4：继续扩展到 `refusal / logprobs` merge 细节。

本轮选择：

- 主任务：`1 + 2`，直接修 `Responses` 双向转换链的 mixed content 顺序/保留问题。
- 配套子任务：`3`，补 `Anthropic Messages` 回归，锁住当前已正确的顺序行为。

## 4. 发现的问题

1. `internal/transformer/inbound/openai/response.go` 把 internal assistant `MultipleContent` 转回 Responses 输出时，会先把 image 直接 append 到 `Output`，最后再把全部 text 拼成一个 `message` item，导致 `text -> image -> text` 变成 `image -> text+text`。
2. `internal/transformer/outbound/openai/response.go` 把 Responses 输出转回 internal response 时，会把所有 text 先并到一个 builder，再把 image 放进 `contentParts`，导致 mixed content 的相对顺序被压扁成“全部文本在前、图片在后”。
3. `internal/transformer/inbound/anthropic/messages.go` 当前实现其实能保序，但缺少直接回归测试，下一轮容易被无意回退。

## 5. 实现结果

- `internal/transformer/inbound/openai/response.go`
  - mixed `MultipleContent` 现在按顺序 flush 成多个 Responses output items：连续 text 形成 `message` item，image 则在遇到时立即落成 `image_generation_call`，不再把图片提前、文本挤成单块。
  - data URL image 会补出 `output_format`，让回放更接近原始图像类型。
- `internal/transformer/outbound/openai/response.go`
  - Responses 输出转 internal response 时改为按事件顺序累计 text/image；遇到 image 前会先提升已有纯文本，image 后再来的 text 也会落到尾部，不再统一被折叠到图片前面。
- 测试：
  - 新增 `ResponseInbound.TransformResponse` mixed text/image 顺序回归。
  - 新增 `convertToLLMResponseFromResponses` mixed text/image 顺序回归。
  - 新增 `MessagesInbound.TransformResponse` mixed text/image 顺序回归。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai ./internal/transformer/inbound/anthropic`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无格式问题

## 7. 风险与阻塞

- 本轮无直接阻塞。
- 本次仍未进入更大的 stream race / retry 裁决任务；该任务保留给下一轮更大的 backend slice。

## 8. 下一轮建议

1. 继续同一 backend compatibility 主线，补 `refusal`、`logprobs`、tool-call 极端分片的 focused helper/consumer tests。
2. 复查 `Responses` 请求侧 assistant mixed content 表达是否还存在 image 兼容缺口，若要处理，保持在同一 transformer compatibility lane 内收口。
3. 在 transformer mixed content 契约稳定后，再回到 Phase B 更大的 relay/race 主链任务。
