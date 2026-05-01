# 2026-05-01 Backend Task Responses Message Done Consumer Fallback Closure

## 1. 任务信息

- 任务名称：Responses `message output_item.done` consumer fallback 收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-toolcall-done-consumer-fallback-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-terminal-event-semantics-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-content-filter-done-marker-terminal-closure.md`
- 本次任务目标：沿同一 Phase B Responses transformer compatibility 主线，补齐 `internal/transformer/outbound/openai/response.go` 对 `response.output_item.done(type=message)` 的 text consumer fallback，避免当流式事件链只保留 message done、缺少 `response.output_text.delta` 时 assistant 文本静默丢失。
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/outbound/openai/response_test.go`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮只涉及单一 transformer consumer fallback，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：补 `response.output_item.done(type=message)` 的 text fallback。
- 候选任务 2：检查 mixed text/tool interleave 在 done-event 组合下是否还有额外 consumer 假设。
- 候选任务 3：回到更大的 Phase B relay/race 主链。

本轮选择：

- 主任务：`1`，只做 message done-event text consumer fallback。
- 配套子任务：补两条 outbound focused regression，分别覆盖“只有 done 恢复文本”和“已有 delta 时 done 不重复追加”。

## 4. 发现的问题

1. `internal/transformer/outbound/openai/response.go` 当前会消费 `response.output_text.delta` 来恢复 assistant 文本，但对 `response.output_item.done` 的 `message` 分支直接忽略。
2. 这意味着当上游或回放链路只保留 `message output_item.done`、没有前置 `output_text.delta` 时，outbound stream consumer 会静默丢掉文本内容。
3. 同时不能简单把所有 `message output_item.done` 都转成 fallback chunk，否则已消费过 `output_text.delta` 的场景会重复追加同一段文本。

## 5. 实现结果

- `internal/transformer/outbound/openai/response.go`
  - 为 `ResponseOutbound` 新增 message-level stream state，按 `output_index / item_id` 记录某个 message output item 的文本是否已通过 `response.output_text.delta` 消费。
  - 新增 `fallbackMessageChunk()`，仅当 `message output_item.done` 携带的 `output_text` 尚未被消费时，才补发一个 assistant text chunk。
  - `response.created` 时会同步重置 message stream state，避免跨响应污染。
  - `response.output_text.delta` 会显式标记对应 message 已看见文本，后续同 item 的 `output_item.done` 不再重复记账。
  - `response.output_item.done` 现在除了既有 `function_call` fallback 之外，也会处理 `message` 分支。
- `internal/transformer/outbound/openai/response_test.go`
  - 新增 `TestResponseOutboundTransformStreamRestoresMessageFromDoneWithoutDelta`，锁定“只有 message done 事件也能恢复 assistant 文本”。
  - 新增 `TestResponseOutboundTransformStreamSkipsMessageDoneAfterTextDelta`，锁定“已有 text delta 时 done 不再重复吐文本”。

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
- `git diff --check` 仍提示工作树现有的 LF/CRLF 触碰警告，但没有新增 patch 格式错误。
- 当前收口的是 message done-event text fallback，不涉及更大范围的 Responses event schema 扩展；若后续还要补 mixed text/image done fallback、annotation 透传或更细粒度 content-part consumer contract，应继续作为同主线相邻小任务推进。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，优先检查 `response.content_part.done` / `response.output_text.done` 在“无 delta 仅 done”场景是否还存在 text-only consumer 缝隙。
2. 若继续留在 Responses consumer lane，可评估 `message output_item.done` 中 mixed `output_text + image/tool` 组合是否还有 fallback 假设，但保持在小闭环测试优先的节奏内。
3. 当 Responses 的 `tool-call / refusal / logprobs / terminal-state / done-event text fallback` 合同都稳定后，再切回更大范围的 Phase B relay/race 主链。
