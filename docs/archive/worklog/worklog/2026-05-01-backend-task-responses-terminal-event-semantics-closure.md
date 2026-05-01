# 2026-05-01 Backend Task Responses Terminal Event Semantics Closure

## 1. 任务信息

- 任务名称：Responses 终态事件语义收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-toolcall-done-consumer-fallback-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-toolcall-index-stability-closure.md`
- 本次任务目标：沿同一 Phase B transformer compatibility 主线，收口 `Responses` 流式终态事件的 `completed / incomplete / failed` 语义，保证 inbound 侧发出的终态事件与 outbound 侧消费后的 `finish_reason / usage` 一致，不再把 `length` 或 `error` 静默折叠成 `completed`。
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/outbound/openai/response_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent；本轮范围是单一 transformer lane 内的流式终态语义闭环，主线程直接完成风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：修正 `Responses` 流式终态事件 `response.completed / response.incomplete / response.failed` 的语义映射。
- 候选任务 2：继续检查 tool-call done 事件只带 `call_id / item_id` 的回补边界。
- 候选任务 3：切回更大范围的 Phase B relay/race 主链。

本轮选择：

- 主任务：`1`，只做 Responses 终态事件语义与 round-trip 收口。
- 配套子任务：补 inbound/outbound focused regression，覆盖 `length`、`error`、usage 尾块和无 usage + `[DONE]` 兜底场景。

## 4. 发现的问题

1. `internal/transformer/inbound/openai/response.go` 在流式路径里只要收到了 `finish_reason`，最终几乎都会发 `response.completed + status=completed`，导致 `length` / `error` 终态语义丢失。
2. inbound 侧原逻辑把终态事件与 usage 尾块强耦合；如果没有 usage 尾块，就可能只收到 `[DONE]` 而缺失真正的 terminal response event。
3. `internal/transformer/outbound/openai/response.go` 虽然能识别 `response.incomplete / response.failed`，但原实现把 `response.incomplete` 直接落到 `error` 分支之外且不透传 usage，round-trip 后会出现 finish reason 或 usage 不完整。

## 5. 实现结果

- `internal/transformer/inbound/openai/response.go`
  - 为 `ResponseInbound` 增加 `terminalEventType / terminalStatus` 状态，显式保存 `finish_reason -> response.*` 的终态映射结果。
  - 新增 `responsesTerminalState()`，把 `length / content_filter` 映射到 `response.incomplete`，`error` 映射到 `response.failed`，其余保持 `response.completed`。
  - 新增 `emitTerminalResponseEvent()`，统一构造终态 `ResponsesResponse`，避免后续分支继续硬编码 `completed`。
  - 终态事件改为两段式：
    - 若 finish 之后收到 usage 尾块，则在该尾块发出 terminal response event 并携带 usage。
    - 若没有 usage 尾块，则在 `[DONE]` 前兜底发出 terminal response event，避免终态事件缺失。
- `internal/transformer/outbound/openai/response.go`
  - 将 `response.incomplete` 单独映射为 internal `finish_reason=length`。
  - 为 `response.incomplete` 和 `response.failed` 都补齐 `Response.Usage` 透传，保证 round-trip 后 usage 不丢。
- `internal/transformer/inbound/openai/response_stream_test.go`
  - 新增 `TestResponseInboundStreamEmitsIncompleteTerminalEventOnLengthFinish`，覆盖“无 usage 尾块时在 `[DONE]` 前仍发 `response.incomplete`”场景。
  - 新增 `TestResponsesStreamRoundTripMapsIncompleteAndFailedTerminalEvents`，验证 `length / error` 在 inbound SSE 到 outbound internal chunk 的 round-trip 中能得到正确的 `finish_reason` 和 usage。
- `internal/transformer/outbound/openai/response_test.go`
  - 新增 `TestResponseOutboundTransformStreamMapsIncompleteAndFailedTerminalEvents`，锁住 outbound 对 `response.incomplete / response.failed` 的终态消费语义和 usage 透传。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无 patch 格式错误

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\inbound\openai\response.go .\internal\transformer\inbound\openai\response_stream_test.go .\internal\transformer\outbound\openai\response.go .\internal\transformer\outbound\openai\response_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal/transformer/inbound/openai/response.go internal/transformer/inbound/openai/response_stream_test.go internal/transformer/outbound/openai/response.go internal/transformer/outbound/openai/response_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `git diff --check` 仍提示工作树现有的 LF/CRLF 触碰警告，但没有新增 patch 格式错误。
- 目前已收口的是终态事件语义，不包含更大范围的 Responses event schema 扩展；如需新增外显 `logprobs` event lane 或更丰富的 incomplete details，应继续作为同主线相邻小任务处理。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，检查 `response.output_item.done` / `response.function_call_arguments.done` 在只带 `call_id` 或只带 `item_id` 时的恢复边界。
2. 若继续留在 Responses 终态 lane，可评估是否需要把 `content_filter` 从当前 `incomplete` 语义进一步细分到单独的 consumer-visible contract，但不要与 relay/race 主链混做。
3. 当终态、tool-call、refusal、logprobs 契约都稳定后，再切回更大范围的 Phase B relay/race 主链。
