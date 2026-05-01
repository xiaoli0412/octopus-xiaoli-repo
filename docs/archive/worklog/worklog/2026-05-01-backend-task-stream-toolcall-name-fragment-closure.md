# 2026-05-01 Backend Task Stream ToolCall Name Fragment Closure

## 1. 任务信息

- 任务名称：stream tool-call `name / arguments` 分片闭环
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-aggregate-logprobs-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-openai-responses-refusal-consumer-closure.md`
- 本次任务目标：沿同一 Phase B transformer compatibility 主线，收口 streaming tool-call 在 `function name` 与 `arguments` 分片到达时的聚合与 Responses SSE 桥接兼容，避免后续 name 片段在 inbound SSE 事件链里静默丢失或被重复累积。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/model/stream_aggregate.go`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
  - `internal/transformer/outbound/openai/response.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮仅涉及单一 transformer lane 下的 helper / inbound SSE / outbound round-trip 收口，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：补 fragmented tool-call `name / arguments` helper + inbound/outbound 回归。
- 候选任务 2：扩展 Responses stream event 的 `logprobs` 协议承载语义。
- 候选任务 3：切入更大范围 Phase B relay/race 主链。

本轮选择：

- 主任务：`1`，只做 tool-call 分片兼容闭环。
- 配套子任务：补 helper + inbound 聚合 + inbound SSE -> outbound round-trip 三层回归测试，并落盘 worklog。

## 4. 发现的问题

1. `internal/transformer/model/stream_aggregate.go` 的 `mergeStreamingToolCall()` 能做基础字符串追加，但测试只覆盖了 arguments 分片，没有锁住 function name 分片和同一 index 的多 chunk 合同。
2. `internal/transformer/inbound/openai/response.go` 的 `handleToolCalls()` 之前只把后续 chunk 的 `Function.Arguments` 追加到已跟踪 tool call，后续到达的 `ID / Type / Function.Name` 片段会静默丢失。
3. 当首个 tool-call chunk 同时通过 `response.output_item.added` 与 `response.function_call_arguments.delta` 发出时，如果 delta 再重复携带首段 name，会在 outbound 消费并重新聚合后把函数名加倍，属于真实 SSE round-trip 兼容问题。

## 5. 实现结果

- `internal/transformer/inbound/openai/response.go`
  - `handleToolCalls()` 现在会持续累计同一 `toolCallIndex` 的 `ID / Type / Function.Name / Function.Arguments`，而不是只累积 arguments。
  - `response.function_call_arguments.delta` 事件现在会携带当前追踪后的 `CallID`，并且只在“非首次创建”场景传播 name 分片，避免与 `response.output_item.added` 的首段 name 重复记账。
  - `function_call_arguments.done` / `output_item.done` 继续使用累积后的完整 tool call，确保最终 done 事件和 `GetInternalResponse()` 一致。
- `internal/transformer/model/stream_aggregate_test.go`
  - 新增 `TestMergeStreamingResponseAggregate_AppendsToolCallNameAndArgumentsAcrossStreamChunks`。
  - 直接锁住 helper 层的 `name + arguments + finish_reason` 聚合合同。
- `internal/transformer/inbound/openai/response_stream_test.go`
  - 新增 `TestResponseInboundAggregatesToolCallNameAndArgumentsAcrossStreamChunks`，验证 `GetInternalResponse()` 会保留完整函数名与参数。
  - 新增 `TestResponsesStreamRoundTripPreservesToolCallNameFragments`，验证 inbound 发出的 SSE 事件被 outbound 消费并重新聚合后，函数名不会丢失，也不会因首段事件重复而被加倍。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/model ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无 patch 格式问题

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\inbound\openai\response.go .\internal\transformer\model\stream_aggregate_test.go .\internal\transformer\inbound\openai\response_stream_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/model ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal\transformer\inbound\openai\response.go internal\transformer\model\stream_aggregate_test.go internal\transformer\inbound\openai\response_stream_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `git diff --check` 仍会提示 Windows 工作树现有的 LF/CRLF 触碰警告，但没有新增 patch 格式错误。
- 当前只收口了现有 Responses SSE 链里已存在的 tool-call name/arguments 兼容合同；如果后续要扩展更多事件字段（如更细粒度的 tool-call metadata 或 logprobs event lane），应单开同主线相邻任务，不要混到本轮 closure 里。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，优先检查剩余 tool-call 分片边角，如多 tool-call 并发 index、跨 chunk 的 `ID / Type` 变更与 done 事件消费者假设。
2. 若要继续推进 stream 可见性，下一步更适合单独处理 Responses stream event 的 `logprobs` 协议语义，而不是再回到 helper-only 测试补洞。
3. 当 `tool-call / refusal / logprobs / mixed-content` 契约稳定后，再切回更大范围的 Phase B relay/race 主链，不要回到无关 UI 或文档守卫主题。
