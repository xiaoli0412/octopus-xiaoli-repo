# 2026-05-01 Backend Task Responses Output Text Done Consumer Fallback Closure

## 1. 任务信息

- 任务名称：Responses `output_text.done` consumer fallback 收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 4 / 14 / 15.1` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-responses-message-done-consumer-fallback-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-toolcall-name-fragment-closure.md`
- 本次任务目标：沿同一 Phase B Responses transformer compatibility 主线，补齐 `internal/transformer/outbound/openai/response.go` 对 `response.output_text.done` 的 text consumer fallback，避免当事件链只保留 `output_text.done`、缺少 `response.output_text.delta` 时 assistant 文本静默丢失。
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/outbound/openai/response_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮仅涉及单一 transformer consumer fallback，主线程直接闭环风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：补 `response.output_text.done` 的 text fallback。
- 候选任务 2：补 `response.content_part.done` 的 text-only fallback。
- 候选任务 3：继续检查多 tool-call 并发 index 的 done-event 边角。

本轮选择：

- 主任务：`1`，只做 `output_text.done` consumer fallback。
- 配套子任务：补两条 outbound focused regression，分别覆盖“只有 done 恢复文本”和“已有 delta 时 done 不重复追加”。

## 4. 发现的问题

1. `internal/transformer/inbound/openai/response.go` 在关闭文本 content part 时会稳定发出 `response.output_text.done`，而 `internal/transformer/outbound/openai/response.go` 当前只消费 `response.output_text.delta` 与 `response.output_item.done(type=message)`。
2. 这意味着当回放链路或上游事件流只保留 `output_text.done`、没有前置 `output_text.delta` 时，assistant 文本会在 outbound stream consumer 里静默丢失。
3. 同时不能无脑把所有 `output_text.done` 都转成 fallback chunk，否则已有 `output_text.delta` 的路径会重复吐同一段文本。

## 5. 实现结果

- `internal/transformer/outbound/openai/response.go`
  - 抽出 `fallbackMessageTextChunk()`，把 message-level 文本 fallback 逻辑统一给 `message output_item.done` 和 `output_text.done` 复用。
  - 新增 `response.output_text.done` 分支：仅当当前 `output_index / item_id` 尚未通过 `response.output_text.delta` 记账时，才补发 assistant text chunk。
  - 继续复用既有 `messageState` 去重语义，因此已消费过 delta 的同一文本 done 事件不会重复输出。
- `internal/transformer/outbound/openai/response_test.go`
  - 新增 `TestResponseOutboundTransformStreamRestoresOutputTextDoneWithoutDelta`，锁定“只有 `output_text.done` 也能恢复 assistant 文本”。
  - 新增 `TestResponseOutboundTransformStreamSkipsOutputTextDoneAfterTextDelta`，锁定“已有 text delta 时 `output_text.done` 不重复输出”。

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
- 当前收口的是 text-only `output_text.done` fallback；如果后续要继续处理 `content_part.done`、mixed text/image done fallback 或更细 content-part 级消费状态，应继续作为同主线相邻小任务推进。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，优先检查 `response.content_part.done` 在 text-only / mixed-content 场景下是否还存在无 delta 仅 done 的 consumer 缝隙。
2. 若继续留在 Responses consumer lane，可评估 `message output_item.done` 与 `output_text.done` 混用时的多 content-part 去重边界，但保持在小闭环测试优先的节奏内。
3. 当 Responses 的 `tool-call / refusal / logprobs / terminal-state / done-event text fallback` 合同都稳定后，再切回更大范围的 Phase B relay/race 主链。
