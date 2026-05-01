# 2026-05-01 Backend Task Responses Done-Event Identity Recovery Closure

## 1. 任务信息

- 任务名称：Responses done-event 身份恢复边界收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-terminal-event-semantics-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-toolcall-done-consumer-fallback-closure.md`
- 本次任务目标：沿同一 Phase B transformer compatibility 主线，收口 `response.function_call_arguments.done` / `response.output_item.done` 在仅带 `call_id` 或仅靠 `item_id` 时的 tool-call 身份恢复边界，避免多 tool-call 无 `call_id` 场景在 inbound 结束阶段错误复用最后一个 `currentItemID`，或 outbound 侧缺少 focused regression 保护。
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/PLAN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/outbound/openai/response_test.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮范围限定在单一 Responses transformer lane，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：检查 `done` 事件仅带 `call_id` / `item_id` 时的 identity 恢复与回归覆盖。
- 候选任务 2：继续扩展 Responses 终态或 `content_filter` consumer-visible contract。
- 候选任务 3：切回更大范围的 Phase B relay/race 主链。

本轮选择：

- 主任务：`1`，只做 done-event identity recovery。
- 配套子任务：补一条 `only call_id` outbound focused regression，以及一条无 `call_id` 双 tool-call round-trip regression。

## 4. 发现的问题

1. `internal/transformer/inbound/openai/response.go` 只按 `toolCallIndex -> outputIndex` 跟踪 tool-call，没有单独保留 `toolCallIndex -> itemID`；当多个 tool-call 没有 `call_id` 时，结束阶段的 `function_call_arguments.done` / `output_item.done` 会回退到共享的 `currentItemID`，导致多个 tool-call 可能错误复用最后一个 item id。
2. `internal/transformer/outbound/openai/response_test.go` 已覆盖 `added + done` 与 `delta + done dedupe`，但还没有锁住“`done` 仅带 `call_id`”这类更贴近兼容缝隙的恢复边界。
3. 当前 round-trip 回归还没有证明“无 `call_id`、仅靠生成的 `item_id` 区分多个 tool-call”在 inbound SSE 到 outbound internal aggregation 的真实链路里能稳定工作。

## 5. 实现结果

- `internal/transformer/inbound/openai/response.go`
  - 为 `ResponseInbound` 新增 `toolCallItemID map[int]string`，按 tool-call index 独立保留生成出来的 `item_id`。
  - tool-call 首次创建时把生成的 `item_id` 写入该 map，而不再只依赖共享 `currentItemID`。
  - 后续 `response.function_call_arguments.delta` 与结束阶段 `response.function_call_arguments.done` / `response.output_item.done` 都优先使用 `toolCallItemID[index]`，避免无 `call_id` 多 tool-call 结束时串位到同一个 item identity。
- `internal/transformer/outbound/openai/response_test.go`
  - 新增 `TestResponseOutboundTransformStreamRestoresToolCallFromDoneWithOnlyCallID`，锁住“`added` 已暴露 name + `call_id`，后续 `done` 只有 `call_id + arguments` 时仍能补回 arguments，且不会错误重复 id/name”的 consumer 行为。
- `internal/transformer/inbound/openai/response_stream_test.go`
  - 新增 `TestResponsesStreamRoundTripKeepsDistinctItemIDsForCallIDLessToolCalls`，构造两个都没有 `call_id` 的 fragmented tool-call，验证 inbound 发出的两个 `output_item.done` 拥有不同生成 `item_id`，且 outbound round-trip 后仍能正确聚合为两个独立 tool-call，名称、参数和 index 都不串位。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无 patch 格式错误

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\inbound\openai\response.go .\internal\transformer\outbound\openai\response_test.go .\internal\transformer\inbound\openai\response_stream_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/openai ./internal/transformer/outbound/openai`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal/transformer/inbound/openai/response.go internal/transformer/outbound/openai/response_test.go internal/transformer/inbound/openai/response_stream_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `git diff --check` 仍提示工作树现有的 LF/CRLF 触碰警告，但没有新增 patch 格式错误。
- 当前收口的是 tool-call identity 恢复边界，不涉及更大范围的 Responses 事件 schema 扩展；如果后续还要处理 `content_filter` 细分或额外 metadata，可继续作为同主线相邻小任务推进。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，检查 `Responses` 流式事件里是否还有仅靠 `output_index` 才能对齐的残余 consumer 假设，优先看 mixed text/tool output 顺序与空字段 done-event 组合。
2. 若继续停留在 done-event lane，可补 focused regression 覆盖 `output_item.done` 只有 `item_id` 且无 `name/arguments` 新信息时的无重复兜底行为。
3. 当 Responses 的 refusal / logprobs / tool-call / terminal-state 合同都稳定后，再切回更大范围的 Phase B relay/race 主链。
