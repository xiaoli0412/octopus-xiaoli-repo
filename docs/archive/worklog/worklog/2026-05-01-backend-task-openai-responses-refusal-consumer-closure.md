# 2026-05-01 Backend Task OpenAI Responses Refusal Consumer Closure

## 1. 任务信息

- 任务名称：OpenAI Responses `refusal` consumer 闭环收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 6 / 14` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-aggregate-refusal-closure.md`
- 本次任务目标：沿上一轮已完成的流式 `refusal` helper 聚合主线，继续补齐 OpenAI Responses 双向 consumer 转换，避免内部聚合已经保住 `refusal`，但 `TransformResponse` / `convertToLLMResponseFromResponses` 仍静默丢字段。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/model/stream_aggregate.go`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/inbound/openai/response_test.go`
  - `internal/transformer/outbound/openai/response_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮只涉及单一 transformer lane 下的 consumer 收口，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：补 `logprobs` helper / consumer 回归。
- 候选任务 2：补 OpenAI Responses `refusal` 双向 consumer 转换。
- 候选任务 3：继续扩展更碎片化的 tool-call 分片场景。
- 候选任务 4：切入更大范围 relay/race 主链。

本轮选择：

- 主任务：`2`，只做 OpenAI Responses `refusal` 双向 consumer 闭环。
- 配套子任务：补 inbound / outbound 各 1 条 targeted regression test。

## 4. 发现的问题

1. 上一轮 `stream_aggregate.go` 已经能把多段 stream chunk 的 `refusal` 聚合回 `InternalLLMResponse.Message.Refusal`，但 OpenAI Responses 双向转换链并没有消费这个字段。
2. `internal/transformer/inbound/openai/response.go` 的 `convertToResponsesAPIResponse()` 只会输出 reasoning、tool call、text、image，不会把 assistant refusal 转成 Responses output item。
3. `internal/transformer/outbound/openai/response.go` 的 `convertToLLMResponseFromResponses()` 也不会从 Responses output item 恢复 refusal，因此即便上游或回放侧提供 refusal，internal response 仍会丢失。

## 5. 实现结果

- `internal/transformer/inbound/openai/response.go`
  - 为 `ResponsesItem` 补上 `Refusal` 字段。
  - `convertToResponsesAPIResponse()` 现在会把 assistant `message.Refusal` 映射成一个 `type="message"` 的 refusal output item。
  - `convertItemToMessage()` 现在会把 Responses message item 上的 `refusal` 反解回 internal `Message.Refusal`。
- `internal/transformer/outbound/openai/response.go`
  - 为 outbound 侧 `ResponsesItem` 同步补上 `Refusal` 字段。
  - `convertToLLMResponseFromResponses()` 现在会聚合 message item 上的 refusal，并写回 internal `Message.Refusal`。
- 测试：
  - `internal/transformer/inbound/openai/response_test.go`
    - 新增 `TestResponseInboundTransformResponsePreservesRefusalOutput`。
  - `internal/transformer/outbound/openai/response_test.go`
    - 新增 `TestConvertToLLMResponseFromResponsesPreservesRefusal`。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- `git diff --check` 对本轮目标文件无格式问题

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\inbound\openai\response.go .\internal\transformer\outbound\openai\response.go .\internal\transformer\inbound\openai\response_test.go .\internal\transformer\outbound\openai\response_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- passed `git diff --check -- internal\transformer\inbound\openai\response.go internal\transformer\outbound\openai\response.go internal\transformer\inbound\openai\response_test.go internal\transformer\outbound\openai\response_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `git diff --check` 输出仍包含 Windows 工作树的 LF/CRLF 提示，但没有新增 patch 格式错误。
- 同主线剩余缺口仍包括 `logprobs`、更碎片化 tool-call name/arguments 分片场景，以及后续更大范围的 relay/race 裁决逻辑。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，补 `logprobs` 的 helper + inbound/outbound focused regression coverage，避免下一个字段仍停留在“实现有、测试无”。
2. 复查 Responses stream event 侧是否还存在 refusal 专用事件或 `failed/incomplete` 状态映射细节需要更细兼容，但保持在同一 OpenAI-compatible transformer lane 内处理。
3. 当 `refusal` / `logprobs` / tool-call 分片契约都稳定后，再切回更大范围的 Phase B relay/race 主链，不要回到无关 UI 或文档守卫主题。
