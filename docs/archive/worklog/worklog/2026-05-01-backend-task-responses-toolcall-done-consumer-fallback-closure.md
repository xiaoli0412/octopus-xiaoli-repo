# 2026-05-01 Backend Task Responses ToolCall Done Consumer Fallback Closure

## 1. 任务信息

- 任务名称：Responses tool-call `done` consumer fallback 收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-toolcall-index-stability-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-toolcall-name-fragment-closure.md`
- 本次任务目标：沿同一 Phase B transformer compatibility 主线，补齐 `internal/transformer/outbound/openai/response.go` 对 `response.function_call_arguments.done` / `response.output_item.done` 的 tool-call 消费兜底，避免当 delta 链不完整或只剩 done 事件时，internal tool-call 聚合静默缺失。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/outbound/openai/response_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮只涉及单一 transformer lane 的 consumer 兜底，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：补 `Responses` tool-call done-event consumer 兜底。
- 候选任务 2：继续扩展 `Responses` stream event 的 `logprobs` 协议语义。
- 候选任务 3：切回更大范围的 Phase B relay/race 主链。

本轮选择：

- 主任务：`1`，只做 tool-call done-event consumer fallback。
- 配套子任务：补 outbound focused regression，覆盖“只有 done 事件也能恢复”和“已有 delta 时 done 不重复记账”。

## 4. 发现的问题

1. `internal/transformer/outbound/openai/response.go` 当前只消费 `response.output_item.added` 和 `response.function_call_arguments.delta`，对 `response.function_call_arguments.done` / `response.output_item.done` 基本忽略。
2. 如果上游或回放链路只留下 done 事件，或者 delta 片段不完整，outbound 侧就无法把完整 tool-call 恢复成 internal streaming chunk。
3. 同时不能简单把 done 事件也直接重复吐回 internal chunk，否则会把已有 delta 聚合过的 `name / arguments` 再追加一遍，造成 tool-call 内容翻倍。

## 5. 实现结果

- `internal/transformer/outbound/openai/response.go`
  - 为 `ResponseOutbound` 新增仅用于 stream 消费的 tool-call state，记录每个 internal tool-call index 是否已经见过 `item_id / call_id / name / arguments / output_item.added`。
  - 新增 `fallbackToolCallChunk()`，只在 done 事件携带的是“此前未见过的 tool-call 信息”时才补发 internal tool-call chunk。
  - `response.function_call_arguments.done` 现在会在 arguments 尚未被 delta 链消费时补发 arguments-only fallback chunk。
  - `response.output_item.done` 现在会在 name / arguments / id 中仍有缺口时补发 fallback chunk；如果前面的 added/delta 已经完整消费，则直接跳过，避免重复追加。
- `internal/transformer/outbound/openai/response_test.go`
  - 新增 `TestResponseOutboundTransformStreamRestoresToolCallFromAddedPlusDoneWithoutDelta`。
  - 新增 `TestResponseOutboundTransformStreamSkipsRedundantToolCallDoneAfterDelta`。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/outbound/openai ./internal/transformer/inbound/openai`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无 patch 格式错误

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\outbound\openai\response.go .\internal\transformer\outbound\openai\response_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/outbound/openai ./internal/transformer/inbound/openai`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal/transformer/outbound/openai/response.go internal/transformer/outbound/openai/response_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `git diff --check` 仍会提示工作树现有的 LF/CRLF 触碰警告，但没有新增 patch 格式错误。
- 当前收口的是 done-event 兜底消费，不是更大范围的 `Responses` 事件协议扩展；如果后续需要更细粒度 metadata / logprobs event lane，应单开同主线相邻任务。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，复查 `response.completed` / `response.failed` / `response.incomplete` 的 finish-reason 映射与 `Responses` 终态消费是否仍存在隐式兼容缝隙。
2. 若继续留在 tool-call lane，更适合检查 `output_item.done` 对仅有 `call_id` 或仅有 `item_id` 的恢复边角，而不是重复 helper 层修补。
3. 当 tool-call / refusal / logprobs / mixed-content / done-event 契约都稳定后，再切回更大范围的 Phase B relay/race 主链。
