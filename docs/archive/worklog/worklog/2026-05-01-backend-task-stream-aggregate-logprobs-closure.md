# 2026-05-01 Backend Task Stream Aggregate Logprobs Closure

## 1. 任务信息

- 任务名称：stream aggregate `logprobs` 聚合闭环
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `OpenAI-compatible` 流式与响应转换兼容硬化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 14` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-openai-responses-refusal-consumer-closure.md`
  - `docs/archive/worklog/worklog/2026-05-01-backend-task-stream-aggregate-refusal-closure.md`
- 本次任务目标：沿同一 Phase B transformer compatibility 主线，补齐共享流聚合器对 `logprobs` 的显式回归覆盖，并确认 `ResponseInbound.GetInternalResponse()` 会保留聚合后的 token logprobs。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/model/stream_aggregate.go`
  - `internal/transformer/model/stream_aggregate_test.go`
  - `internal/transformer/inbound/openai/response_stream_test.go`
- 本次是否启用子 agent 与分工边界：否。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮只涉及单一 transformer lane 下的 helper + inbound 聚合闭环，主线程直接收口风险最低。

## 3. 本次候选任务与选择

- 候选任务 1：补 `logprobs` helper + inbound 聚合回归。
- 候选任务 2：补更碎片化的 tool-call name/arguments 聚合回归。
- 候选任务 3：扩展 Responses stream event 的 `logprobs` 事件承载语义。
- 候选任务 4：切回更大范围 relay/race 主链。

本轮选择：

- 主任务：`1`，只做共享 stream aggregate 的 `logprobs` 合同收口。
- 配套子任务：补 `ResponseInbound.GetInternalResponse()` 的定向回归，证明 helper 聚合结果被真实 consumer 保留。

## 4. 发现的问题

1. `internal/transformer/model/stream_aggregate.go` 之前对 `choice.Logprobs` 只是直接 `append(choice.Logprobs.Content...)`，没有专门测试覆盖其跨 chunk 聚合行为。
2. 现有实现虽然能把 token logprobs 拼接起来，但没有锁定嵌套 `Bytes` / `TopLogprobs` 切片在聚合后不受源 chunk 后续修改影响。
3. `ResponseInbound.GetInternalResponse()` 已切到共享 helper，但此前没有 focused regression test 证明流式 Responses 路径最终会保留聚合后的 `logprobs`。
4. 本轮额外检查发现 `Responses` stream event 当前并没有明确的 `logprobs` 事件语义承载位；这属于同主线下一相邻任务，但会扩大边界，因此本轮先不扩展到 event 协议层。

## 5. 实现结果

- `internal/transformer/model/stream_aggregate.go`
  - 把 `choice.Logprobs` 的合并抽成 `mergeStreamingLogprobs()`。
  - 聚合时对 `Bytes` 与 `TopLogprobs[*].Bytes` 做深拷贝，避免聚合结果与原 chunk 切片共享底层数组。
- `internal/transformer/model/stream_aggregate_test.go`
  - 新增 `TestMergeStreamingResponseAggregate_AppendsLogprobsAcrossStreamChunks`。
  - 覆盖两段 chunk 的 token logprobs 聚合、`finish_reason` 保留，以及源 chunk 后续变更不污染聚合结果。
- `internal/transformer/inbound/openai/response_stream_test.go`
  - 新增 `TestResponseInboundAggregatesLogprobsAcrossStreamChunks`。
  - 证明 Responses inbound 的流式聚合结果在 `GetInternalResponse()` 上仍保留完整 logprobs 列表与深拷贝后的 bytes。

## 6. 验收条件

- `gofmt -w` 目标文件通过
- `go test ./internal/transformer/model ./internal/transformer/inbound/openai`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `git diff --check` 对本轮目标文件无格式问题

## 7. 验证结果

- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w .\internal\transformer\model\stream_aggregate.go .\internal\transformer\model\stream_aggregate_test.go .\internal\transformer\inbound\openai\response_stream_test.go`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/model ./internal/transformer/inbound/openai`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `git diff --check -- internal\transformer\model\stream_aggregate.go internal\transformer\model\stream_aggregate_test.go internal\transformer\inbound\openai\response_stream_test.go`

## 8. 风险与阻塞

- 本轮无直接阻塞。
- `Responses` stream event 当前仍缺 `logprobs` 显式协议语义；如果后续需要真正把 logprobs 透传到 event 层，应单开同主线相邻任务，不要把 helper/test 闭环与协议扩展混在一轮里。
- 同主线剩余缺口仍包括更碎片化的 tool-call name/arguments 分片场景，以及后续更大范围的 relay/race 裁决逻辑。

## 9. 下一轮建议

1. 继续同一 Phase B transformer compatibility 主线，优先补 fragmented tool-call name/arguments merge 的 helper + inbound/outbound focused regression coverage。
2. 如果要继续推进 `logprobs`，下一步应切到 Responses stream event 语义层，明确当前协议是否要承载 token logprobs，而不是重复 helper 级测试。
3. 当 `logprobs` / tool-call 分片契约稳定后，再切回更大范围的 Phase B relay/race 主链，不要回到无关 UI 或文档守卫主题。
