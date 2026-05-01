# 2026-05-01 Backend Task Inbound Stream Aggregate Shared Helper

## 1. 任务信息

- 任务名称：入站流式聚合共享 helper 收口
- 日期：2026-05-01
- 当前阶段：`LLM Gateway canonical refactor / Phase B 可用性核心收口`
- 对应 milestone：里程碑 1 `可用性核心` 中的 `OpenAI-compatible` 流式兼容路径稳态化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `1.2 / 6 / 14` 节
- 对应 workflow 章节：`docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.2`、`4 Phase B`
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-14-backend-task-02-advanced-failover-runtime.md`
  - `docs/archive/worklog/worklog/2026-05-01-phase-g-active-archive-input-doc-entry-guard.md`
- 本次任务目标：在不扩散协议事件逻辑的前提下，收口三条入站 transformer 中重复的 stream chunk 聚合实现，并补齐 `reasoning_signature` 聚合保留，形成一处真实后端增量。
- 本次已盘点本地资源：
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `internal/transformer/inbound/anthropic/messages.go`
  - `internal/transformer/inbound/openai/chat.go`
  - `internal/transformer/inbound/openai/response.go`
  - `internal/transformer/model/model.go`
  - 现有 stream tests
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认最近多轮主要停留在 Phase G 文档/guard 收口，本轮应切回真实后端代码增量。
  - `using-superpowers`：按会话要求先核对技能边界。
  - `brainstorming`：仅作流程边界核对；本轮属于既有实现收口，不进入新的设计审批流。
- 若未使用部分本地资源或上下文，原因：本轮不涉及 UI 页面、浏览器 smoke、导入回滚页面交互与运行态脚本。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程串行执行。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务改动面小、文件耦合强，主线程直接收口更稳。

## 3. 本次硬规则

- 只做与 `OpenAI-compatible` 入站流式兼容直接相关的后端小闭环。
- 不改变各协议现有事件输出语义，只收口重复聚合逻辑。
- 必须保留并验证 `reasoning` / `tool_calls` 等流式增量字段的聚合结果。

## 4. 本次禁止事项

- 不扩散到 outbound transformer 语义改写。
- 不顺手调整 UI、文档总状态或浏览器 smoke 支撑脚本。
- 不对无关的 relay / routing / backup 模块做清理式重构。

## 5. 本次验收条件

- `go test ./internal/transformer/inbound/anthropic ./internal/transformer/inbound/openai ./internal/transformer/model`
- `go test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- `gofmt` 后目标文件保持无格式问题

## 6. 本次回滚点

- `internal/transformer/model/stream_aggregate.go`
- `internal/transformer/inbound/openai/chat.go`
- `internal/transformer/inbound/openai/response.go`
- `internal/transformer/inbound/anthropic/messages.go`
- `internal/transformer/inbound/anthropic/messages_stream_test.go`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改后端共享聚合逻辑，再补测试。
- 受影响后端模块：`internal/transformer/model`、`internal/transformer/inbound/openai`、`internal/transformer/inbound/anthropic`
- 受影响前端模块：无
- 受影响接口：无 HTTP 接口形状变化；仅影响流式 chunk 聚合后的内部响应结构
- 是否影响旧数据：否
- 是否影响旧行为：只会让聚合后的内部流式结果更一致，并补保留 `reasoning_signature`

## 8. 实施步骤

1. 先跑 `backup/import` 与 `transformer` 定向测试，确认 Phase F 当前并无真实失败点，再切到更适合本轮形成闭环的协议兼容层。
2. 复查 `anthropic/messages.go`、`openai/chat.go`、`openai/response.go` 的 stream 聚合实现，确认三处存在重复逻辑且当前不会聚合保留 `ReasoningSignature`。
3. 在 `internal/transformer/model/stream_aggregate.go` 新增共享 helper，把 choice/message/tool-call 聚合统一到一处，并让三条 inbound 路径复用它。
4. 为 Anthropic 流补一条 `reasoning_signature` 聚合测试，锁住这次修复点。
5. 跑 formatter 与 transformer 定向回归，确认收口没有破坏 inbound/outbound 现有测试链。

## 9. 测试与验证

- 构建命令：`gofmt -w` 目标文件
- 测试命令：
  - `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/op -run 'TestDBExportAll|TestDBImportIncremental|TestExportDumpLegacyView'`
  - `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/server/handlers -run 'TestExportDB|TestImport|TestSettingImport'`
  - `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/anthropic ./internal/transformer/inbound/openai ./internal/transformer/model`
  - `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- 专项验证：`reasoning_signature` 现在会通过共享 helper 进入聚合后的 `InternalLLMResponse`

## 10. 风险与兼容性

- 新风险：低；改动集中在内部聚合 helper 和三个调用点。
- 兼容性风险：低到中；如果共享 helper 聚合字段漏项，会影响依赖 `GetInternalResponse()` 的日志、统计或二次协议转换，因此本轮补了更贴近真实字段的测试。
- 是否阻塞下一任务：不阻塞。

## 11. 收工记录

- 构建是否通过：通过 formatter；未单独跑全量 `go build ./...`。
- 测试是否通过：通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：提醒最近多轮偏向 guard/doc drift，本轮应优先回到真实代码增量。
  - canonical plan：明确 `OpenAI-compatible` 行为和流式输出语义属于不可妥协约束。
  - workflow：限定本轮只选一个小闭环任务并做直接验证。
  - 现有 stream tests：说明“即时聚合替代 chunk retention”的重构已在推进，但还缺共享化与签名字段覆盖。
- 本次使用了哪些子 agent 及其结论：无。
- 手工 smoke 状态：未执行；本轮仅做 Go 后端与 transformer 回归。
- 手工 smoke 阻塞原因 / 缺少的环境：无；本轮不需要运行态页面验证。
- 待验证页面清单：无新增页面项。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且任务范围小。
- worklog 是否更新：yes
- 遗留项：
  - 共享 helper 目前只锁住了 `reasoning_signature` 新增覆盖；后续仍可继续补 multipart content / image / message-source 分支的直接 helper 级测试。
  - 最近多轮 Phase G 仍有浏览器/CDP 宿主 blocker 未解，但已不再影响本轮后端小闭环。
- 下一任务前置条件是否满足：满足。

## 12. 执行与结果

1. 先对 `backup/import` 相关后端测试做了定向复查，确认 `contains_secrets` / 凭据导入契约当前是绿的，不适合继续在该池空转。
2. 随后切到 `OpenAI-compatible` 流式兼容主链，确认 `anthropic/messages.go`、`openai/chat.go`、`openai/response.go` 各自都维护了一套近似相同的 stream chunk 聚合逻辑。
3. 新增 [`internal/transformer/model/stream_aggregate.go`](/D:/GPT-codex/octopus_repo/internal/transformer/model/stream_aggregate.go)，把 choice/message/tool-call 的聚合统一到共享 helper；三条 inbound 流路径改为直接复用该 helper，不再各自维护同一套实现。
4. 这次共享化顺带修掉了一个真实缺口：原聚合逻辑不会保留 `ReasoningSignature`，现在会随着 reasoning 流块一起累积保留，避免 Anthropic thinking 流在 `GetInternalResponse()` 路径丢签名。
5. 新增了 `MessagesInbound` 的 `reasoning_signature` 聚合用例，并跑通 inbound/outbound transformer 定向回归，确认这次收口没有带出新的协议层回归。

## 13. 验证

- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/op -run 'TestDBExportAll|TestDBImportIncremental|TestExportDumpLegacyView'`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/server/handlers -run 'TestExportDB|TestImport|TestSettingImport'`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/anthropic ./internal/transformer/inbound/openai ./internal/transformer/model`
- passed `. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/transformer/inbound/... ./internal/transformer/outbound/...`
- passed `. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w ...`
