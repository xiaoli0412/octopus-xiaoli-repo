# 2026-05-01 Backend Task Stream Aggregate Refusal Closure

## 1. 任务信息

- 任务名称：流式 `refusal` 聚合兼容收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式语义兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 6 / 14` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-openai-responses-mixed-content-order-closure.md`
- 本次任务目标：沿上一轮 transformer streaming compatibility 主线，补齐 `refusal` 在多段 stream chunk 下的聚合语义，避免 `GetInternalResponse()` 只保留最后一段拒绝文本。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/model/stream_aggregate.go`
  - `internal/transformer/model/stream_aggregate_test.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
  - `internal/transformer/inbound/openai/chat_stream_test.go`
  - `internal/transformer/inbound/anthropic/messages_stream_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务集中在同一 transformer helper 与单一回归入口，主线程直接收口风险更低。

## 3. 本次候选任务与选择

- 候选任务 1：收口流式 `refusal` 聚合语义。
- 候选任务 2：继续扩展更碎片化的 tool-call name/arguments 聚合边界。
- 候选任务 3：补 `logprobs` 聚合回归。
- 候选任务 4：转入更大范围 relay/race 主链任务。

本轮选择：

- 主任务：`1`，只修流式 `refusal` 的 helper 聚合缺口。
- 配套子任务：补 direct helper test 与一个 `ResponseInbound` stream regression，确保不只是 helper 通过而是入口链也锁住。

## 4. 发现的问题

1. `internal/transformer/model/stream_aggregate.go` 对 `Message.Refusal` 仍是覆盖式赋值：后一段 chunk 到来时会直接覆盖前段拒绝文本。
2. 现有 stream aggregate tests 已覆盖 `reasoning_signature`、multipart content、tool-call arguments，但没有覆盖 `refusal` 的分段聚合场景。
3. `ResponseInbound.GetInternalResponse()` 已经依赖统一 helper 聚合，因此 helper 漏洞会直接反映为 Responses stream 结束后的内部聚合结果不完整。

## 5. 实现结果

- `internal/transformer/model/stream_aggregate.go`
  - 将 `src.Refusal` 的处理从覆盖式改为追加式聚合，和 `reasoning_content`、`reasoning_signature`、tool-call partial arguments 的流式累积语义保持一致。
- `internal/transformer/model/stream_aggregate_test.go`
  - 新增 `TestMergeStreamingResponseAggregate_AppendsRefusalAcrossStreamChunks`，验证两段 `refusal` chunk 最终会聚合成完整拒绝文本，并保留 finish reason。
- `internal/transformer/inbound/openai/response_stream_test.go`
  - 新增 `TestResponseInboundAggregatesRefusalAcrossStreamChunks`，验证 `TransformStream -> GetInternalResponse()` 链路不会只剩最后一段 refusal。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/model ./internal/transformer/inbound/openai ./internal/transformer/inbound/anthropic`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无格式问题

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\model\stream_aggregate.go .\internal\transformer\model\stream_aggregate_test.go .\internal\transformer\inbound\openai\response_stream_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/model ./internal/transformer/inbound/openai ./internal/transformer/inbound/anthropic`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal\transformer\model\stream_aggregate.go internal\transformer\model\stream_aggregate_test.go internal\transformer\inbound\openai\response_stream_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- 当前仍未进入 `logprobs`、更极端的 tool-call name/arguments 分片、或更大范围 relay/race 裁决逻辑；这些属于同主线但更靠后的 backend slice。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，补 `logprobs` 与更碎片化 tool-call name/arguments 的 focused helper tests。
2. 复查 `refusal` 在下游 provider-specific 输出转换里是否存在额外 consumer 假设，若有则补 targeted regression。
3. 当 transformer 流式聚合契约足够稳定后，再切入更大范围的 relay/race 主链任务，不要回到无关 UI 或文档漂移主题。
