# 2026-05-01 Backend Task Stream Aggregate Mixed Message Delta And Multipart Closure

## 1. 任务信息

- 任务名称：stream aggregate mixed message/delta 与 multipart 收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `可用性核心` 中的 `OpenAI-compatible` 流式兼容稳态化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 6 / 14` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-inbound-stream-aggregate-shared-helper.md`
- 本次任务目标：继续停留在同一 backend compatibility 主线，把共享 stream aggregate helper 里剩余的 `choice.Message + choice.Delta` 混合聚合缺口，以及文本流转 multipart/image 时的内容保留缺口收口，并补成直接 helper 级回归测试。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/model/stream_aggregate.go`
  - `internal/transformer/inbound/openai/chat_stream_test.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
  - `internal/transformer/inbound/anthropic/messages_stream_test.go`
  - `internal/transformer/outbound/openai/response.go`
  - `internal/transformer/inbound/anthropic/messages.go`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上轮已把共享 helper 引入主线，本轮应直接沿 helper 补缺口，不回到 Phase G guard/doc 漂移。
  - canonical plan：继续把 `OpenAI-compatible` 流式输出语义视为不可妥协约束。
  - `using-superpowers` / `brainstorming`：只作流程边界核对；本轮仍属于既有实现收口，不进入新的设计审批流。
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮改动集中在单一 helper 与测试文件，主线程直接收口风险最低。

## 3. 本次硬规则

- 只做与 `OpenAI-compatible` 流式聚合兼容直接相关的 helper 与测试闭环。
- 不扩散到 relay、balancer、UI、browser smoke、backup/import 文档链。
- 优先补真实缺口，不做无关清理式重构。

## 4. 本次候选任务与选择

- 候选任务 1：为共享 helper 增加 mixed `choice.Message` / `choice.Delta` 聚合覆盖。
- 候选任务 2：为 `MultipleContent` / image 混合流补直接 helper 测试。
- 候选任务 3：复查 downstream consumer 是否会因 helper 聚合形态变化而吞文本。
- 候选任务 4：回到更大的 stream race / retry 裁决 TODO。

本轮选择：

- 主任务：`1 + 2` 合并收口，集中在 `stream_aggregate.go` 与 helper 级测试。
- 配套子任务：`3` 做最小 consumer 复查，只在确认存在真实丢信息风险时补 helper，不额外改其他包逻辑。

原因：

- 与上轮连续，改动面最小。
- 有明确完成标准和验证手段。
- 能形成真实后端增量，不会再次滑回外围 guard/doc 循环。

## 5. 本次验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/model ./internal/transformer/inbound/openai ./internal/transformer/inbound/anthropic`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无格式问题

## 6. 实施步骤

1. 复查 `internal/transformer/model/stream_aggregate.go`，确认当前 helper 仍把 `choice.Message` 与 `choice.Delta` 当成二选一来源。
2. 复查 `internal/transformer/outbound/openai/response.go` 与 `internal/transformer/inbound/anthropic/messages.go` 的 downstream 消费逻辑，确认一旦聚合结果进入 `MultipleContent` 分支，旧的纯文本字段若不提升会有被吞掉的风险。
3. 修改 helper：
   - 同一 chunk 同时存在 `choice.Message` 与 `choice.Delta` 时，两者都合并。
   - 当先收到纯文本、后收到 image/multipart part 时，把已有文本提升为 `MultipleContent` 的首个 text part。
   - 继续保持纯 text-only 流不必无故切成 multipart。
4. 新增 `internal/transformer/model/stream_aggregate_test.go`，直接锁住：
   - mixed `choice.Message` + `choice.Delta`
   - text -> image -> text 的 multipart promotion
5. 跑 formatter、model/inbound/outbound 定向测试与 diff check，确认本轮闭环稳定。

## 7. 发现的问题

1. 共享 helper 引入后，`choice.Message` 与 `choice.Delta` 的聚合仍是“二选一”，这会在单个 chunk 同时携带 completion-style `message` 和 streaming-style `delta` 时漏掉一部分信息。
2. 旧 helper 在先累积 `Content.Content`、后追加 image/multipart parts 时，不会把已累积文本提升到 `MultipleContent`。而 downstream 的部分 consumer 进入 multipart 分支后不会再读取 `Content.Content`，因此早期文本存在被吞掉的真实风险。
3. 这两个问题都属于 helper 层 contract 漏洞，不需要扩散到三个 inbound 包各自写不同修复。

## 8. 实现结果

- `internal/transformer/model/stream_aggregate.go`
  - 现在会先合并 `choice.Message`，再合并 `choice.Delta`，不再互斥。
  - 新增文本与 content-part 的聚合辅助逻辑：仅文本流继续保留 `Content.Content`；一旦收到非 text part，会把已有文本提升成 `MultipleContent`，确保后续 image/multipart 不丢前文。
- `internal/transformer/model/stream_aggregate_test.go`
  - 新增 helper 级直接回归测试，覆盖 mixed message/delta 和 text/image/text multipart promotion。

本轮未修改：

- relay / balancer / routing 行为
- UI / browser smoke / Phase G guard 文档链
- backup/import / AI automation 主线

## 9. 测试与验证

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\model\stream_aggregate.go .\internal\transformer\model\stream_aggregate_test.go .\internal\transformer\inbound\openai\chat_stream_test.go .\internal\transformer\inbound\openai\response_stream_test.go .\internal\transformer\inbound\anthropic\messages_stream_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/model ./internal/transformer/inbound/openai ./internal/transformer/inbound/anthropic`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal\transformer\model\stream_aggregate.go internal\transformer\model\stream_aggregate_test.go internal\transformer\inbound\openai\chat_stream_test.go internal\transformer\inbound\openai\response_stream_test.go internal\transformer\inbound\anthropic\messages_stream_test.go docs\archive\worklog\worklog\2026-05-01-backend-task-stream-aggregate-mixed-message-delta-and-multipart-closure.md`

## 10. 风险与阻塞

- 本轮无直接阻塞；helper 级测试与 transformer 定向回归均通过。
- 尚未处理的同主线问题：
  - 更广义的 stream race / retry arbitration 仍未进入本轮。
  - 目前新增的是 helper 级 contract 测试，还没有把 text+image 的完整 round-trip 用例补到更下游的 response/anthropic 变换链。

## 11. 下一轮建议

按优先级排序：

1. 继续同一 backend compatibility 主线，为 `outbound/openai/response.go` 与 `inbound/anthropic/messages.go` 增加 end-to-end 回归，证明聚合后的 `MultipleContent` text+image 不会在更下游转换链被吞文本。
2. 复查 stream aggregate 其他字段是否仍有非对称 merge 行为，例如 `logprobs`、`refusal`、tool-call name/arguments 的极端分片场景，并补 focused helper tests。
3. 在 helper 兼容带稳定后，再回到 Phase B 里程碑 1 其余更大的 backend slice，而不是重新掉回 repo-local guard/doc 漂移轮次。
