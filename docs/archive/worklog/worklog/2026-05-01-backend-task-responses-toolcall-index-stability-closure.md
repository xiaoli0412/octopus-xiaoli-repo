# 2026-05-01 Backend Task Responses ToolCall Index Stability Closure

## 1. 任务信息

- 任务名称：Responses tool-call index 稳定性闭环
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-toolcall-name-fragment-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-aggregate-logprobs-closure.md`
- 本次任务目标：沿同一 Phase B transformer compatibility 主线，收口 Responses 非流式与流式 round-trip 中的 tool-call index 语义，避免多个 function call 在 internal `ToolCall.Index` 上串位、全部落成 `0`，或错误复用 `output_index`。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/outbound/openai/response_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮只涉及单一 transformer 响应转换链，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：多 tool-call 并发场景下的 index 稳定性与 done 事件一致性。
- 候选任务 2：Responses stream event 的 `logprobs` 协议语义扩展。
- 候选任务 3：切回更大范围的 Phase B relay/race 主链。

本轮选择：

- 主任务：`1`，只做 tool-call index stability。
- 配套子任务：补非流式 `Responses -> Internal` 与流式 inbound SSE -> outbound aggregate 两层回归测试，并落盘 worklog。

## 4. 发现的问题

1. `internal/transformer/outbound/openai/response.go` 的 `convertToLLMResponseFromResponses()` 在处理多个 `function_call` 时没有恢复递增的 `ToolCall.Index`，导致多个工具调用都会落成默认 `0`。
2. 同文件 `TransformStream()` 处理 `response.output_item.added` / `response.function_call_arguments.delta` 时直接把 `output_index` 塞进 internal `ToolCall.Index`；一旦流里先有文本 message、再有多个 tool call，`output_index` 与 tool-call list index 就会错位。
3. 上轮已收口 tool-call `name / arguments` 分片，但尚未锁住“多个 tool call 且前面已有其他 output item”这一更真实的 round-trip 场景，因此同主线仍有回归风险。

## 5. 实现结果

- `internal/transformer/outbound/openai/response.go`
  - 为 `ResponseOutbound` 新增仅用于流式消费内部恢复的 tool-call index 状态：按 `output_index / item_id / call_id` 建立映射，不再把 `output_index` 直接当作 internal `ToolCall.Index`。
  - 在 `response.created` 时重置该映射，避免跨 response 复用旧状态。
  - `response.output_item.added` 与 `response.function_call_arguments.delta` 现在都会恢复稳定的 tool-call list index，再交给现有聚合器。
  - 非流式 `convertToLLMResponseFromResponses()` 处理多个 `function_call` 时改为显式写入递增 `ToolCall.Index`。
- `internal/transformer/outbound/openai/response_test.go`
  - 新增 `TestConvertToLLMResponseFromResponsesAssignsSequentialToolCallIndexes`，锁住非流式多 tool-call 的 index 语义。
- `internal/transformer/inbound/openai/response_stream_test.go`
  - 新增 `TestResponsesStreamRoundTripKeepsToolCallIndexesStableAcrossOutputItems`。
  - 构造“先输出文本 message，再输出两个 fragmented function call”的真实 round-trip，证明 inbound 发出的 Responses SSE 被 outbound 消费后，聚合结果仍保持 `ToolCall.Index = [0,1]`，不会被中间的 text output item 挤成 `[1,2]` 或全部 `0`。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无 patch 格式问题

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\outbound\openai\response.go .\internal\transformer\outbound\openai\response_test.go .\internal\transformer\inbound\openai\response_stream_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal/transformer/outbound/openai/response.go internal/transformer/outbound/openai/response_test.go internal/transformer/inbound/openai/response_stream_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `git diff --check` 仍输出工作树现有的 LF/CRLF 提示，但没有新增 patch 格式错误。
- 当前只收口了 tool-call index 语义；如果后续需要把更多 tool-call metadata 暴露到 Responses stream event 层，应单开同主线相邻任务，不要把事件协议扩展混进当前闭环。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，优先检查 `response.function_call_arguments.done` / `response.output_item.done` 的消费者是否仍隐含依赖 `output_index`，并补 focused regression。
2. 若 stream 可见性还需增强，再单独评估是否要扩展 Responses 事件层的 `logprobs` 或更细粒度 tool-call metadata，而不是回到 helper-only 修补。
3. 当 tool-call / refusal / logprobs / mixed-content 这些兼容合同稳定后，再切回更大范围的 Phase B relay/race 主链。
