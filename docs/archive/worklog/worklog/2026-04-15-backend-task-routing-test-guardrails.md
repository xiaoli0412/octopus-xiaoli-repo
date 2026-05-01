# Backend Task - Routing Test Guardrails

> 目的：为 `Phase A` 和后续 `Phase B` 提供第一批稳定的后端路由语义测试护栏。

---

## 1. 任务信息

- 任务名称：Routing Test Guardrails
- 日期：2026-04-15
- 当前阶段：Phase A
- 对应 milestone：里程碑 1 可用性核心

## 2. 开工前输入

- 对应 canonical 章节：
  - 第 5 节数据模型设计
  - 第 6 节路由与调度策略
  - 第 13 节里程碑 1
  - 第 14 节验收标准第 1、2、3、13 条
- 对应 workflow 章节：
  - [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 第 3、4 节
- 上一个相关 worklog：
  - `2026-04-14-backend-task-01-channel-key-modes.md`
  - `2026-04-14-backend-task-02-advanced-failover-runtime.md`

## 3. 本次硬规则

- 测试必须优先覆盖确定性、核心、容易退化的路由行为
- 严格顺序与严格权重不能退化
- `classified` 与 `pooled` 语义必须可验证
- 新增测试不能破坏当前可编译状态

## 4. 本次禁止事项

- 不在这轮里顺带改业务语义
- 不把不稳定的网络/DB/HTTP 依赖硬塞进首轮测试
- 不把多个不相干主题混进同一轮测试建设

## 5. 本次实现范围

本轮新增测试覆盖：

- `internal/model/channel_test.go`
  - `NormalizeChannelKeyAllowedModels`
  - `NormalizeChannelKeySourceType`
  - `GetBaseUrl`
  - `GetChannelKey`
  - `EligibleChannelKeysForModel`
  - `GetChannelKeyForModel`
  - `GetChannelKeyForModelExcept`
- `internal/op/channel_test.go`
  - `ChannelUpdate`
  - `ChannelBaseUrlUpdate`
  - `ChannelKeyUpdate` / `ChannelKeySaveDB`
  - `ChannelDel`
- `internal/op/group_test.go`
  - `GroupItemBatchAdd`
  - `GroupUpdate`
- `internal/relay/balancer/balancer_test.go`
  - `RoundRobin`
  - `Failover`
  - `Weighted`
- `internal/relay/balancer/iterator_test.go`
  - sticky 重排
  - reset 恢复基础顺序
  - attempts 顺序记录
  - 并发记录下的 attempt 编号连续性
- `internal/relay/balancer/circuit_test.go`
  - `GetCooldown`
  - `IsTripped`
  - `RecordFailure` / `RecordSuccess`
- `internal/relay/balancer/session_test.go`
  - `GetSticky`
  - `SetSticky`
- `internal/relay/relay_test.go`
  - `allowsRacingByDefault`
  - `isStreamingRequest`
  - `shouldEscalateToRace`
- `internal/relay/relay_more_test.go`
  - `allowsRacingByModel`
  - `finalChannel`
  - `relayAttempt.copyHeaders`
  - `relayAttempt.handleResponse`
  - `relayAttempt.collectResponse`
  - `parseRequest` query passthrough / validation
  - `Handler` key fallback in same channel
  - `Handler` retry across rounds
  - `Handler` stream written-after-no-retry
  - `runRaceFallback` unavailable candidate record
  - `runRaceFallback` earlier-success winner preference
  - `runRaceFallback` 不超过剩余候选的并发上限

## 6. 本次验收条件

- 新增测试通过
- 全量 `go test ./...` 通过
- 全量 `go build ./...` 通过
- 全量 `pnpm build` 通过
- `Phase A` 基础验证脚本通过

## 7. 本次回滚点

- 仅新增 `_test.go` 文件和文档记录
- 如测试设计不合理，可按文件维度回退，不影响业务逻辑

## 8. 测试与验证结果

- `go test ./internal/op`：通过
- `go test ./internal/relay ./internal/relay/balancer`：通过
- `go test ./internal/relay ./internal/relay/balancer ./internal/model`：通过
- 补充 `attempts` 并发编号与 `race_concurrency` 上限边界后，`go test ./internal/relay ./internal/relay/balancer`：通过
- 2026-04-16 继续贴近 canonical `attempts` 语义后，`go test ./internal/relay/balancer ./internal/relay`：通过
- 2026-04-16 收口 `attempts` 真实状态码、429 冷却与 failover-window 边界后，`go test ./...`：通过
- `go test ./...`：通过
- `go build ./...`：通过
- `pnpm build`（`web/`）：通过
- `powershell -ExecutionPolicy Bypass -File .\scripts\phase-a-check.ps1`：通过

## 9. 当前收益

- 第一批核心路由语义已不再完全裸奔
- 后续 `Phase B` 再改 key routing、group fallback、iterator、circuit breaker、sticky session 行为时，至少已有基础护栏
- `Phase A` 已开始从“只保证可编译”进入“开始保证关键语义不回退”

## 10. 遗留项

- `internal/op/channel.go`、`internal/op/group.go` 的业务层测试已补齐
- `internal/relay/balancer/circuit.go`、`internal/relay/balancer/session.go` 的基础单元测试已补齐
- `internal/relay/relay.go` 与 `runRaceFallback` 的关键护栏已补齐
- 目前已补 `attempts` 并发编号连续性、真实 `status_code` 透传、race 结果 `duration/status_code`、429 冷却回归、`failover window exceeded` 不再伪造 attempt、以及 skipped/circuit-break-only 场景的顶层 channel 兜底
- 这轮同时移除了“同渠道换 key”与“升级到 racing”两类过程性 breadcrumb attempt，使 `TotalAttempts` 更贴近真实 outbound 次数
- 若继续推进，可再补更细的并发取消时序、race 多成功结果组合，以及 attempts 导出字段与前端日志展示的一致性验证

## 2026-04-16 增量记录：fill_priority / priority_order 主线现状

本轮继续按 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 收口 key routing 主线，并补齐了 ill_priority / priority_order 在 internal/model/channel.go 中的 phase-1 显式分支。两者当前仍都保持 irst-eligible-first 的最小行为，但代码结构和注释语义已不再共用同一分支；其中 ill_priority 明确为“优先主 key，当前请求中仅在被排除后顺序后退”，priority_order 明确为“每次新请求都从第一个 eligible key 开始，再按 excluded 顺序回退”。

本轮同时补入了 3 条关键回归测试：priority_order 新请求重置首选 key、ill_priority 在主 key 429 冷却恢复后重新优先主 key、以及 NextEligibleChannelKeyAfter 的 strict fallback 顺序直测。验证已通过：go test ./internal/model ./internal/relay ./internal/relay/balancer 与 go test ./...。当前仍未完成的是更进一步的状态化 ill_priority 实现，以及更大范围的 Phase B / 里程碑 2 之后事项。
## 2026-04-16 增量记录：渠道页 key 模式 / 策略可视化

继续按 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 第 9.1 节推进，本轮把此前已在后端与 API 层打通的 key_management_mode / key_routing_policy 直接落到渠道页可见 UI：渠道卡片新增 mode/policy badge，详情页 stats 视图新增“路由策略”区块，显式展示 pooled / classified 与 ound_robin / fill_priority / priority_order，并补充中英繁三套说明文案。

本轮优先复用本地资源建立上下文：以 canonical MD 为主线，直接继承现有 web/src/api/endpoints/channel.ts 中已完成的 normalize 逻辑、现有渠道卡片/详情页组件结构，以及已有 worklog 记录，不重复发明流程。子 agent 调度按 AGENTS.md 规则尝试统一使用 gpt-5.4 做只读分工，但当前工具调用存在参数级阻塞，主线程已记录该阻塞并回退为本地资源优先的串行落地方案，避免中断主线改码。

## 2026-04-16 增量记录：渠道页 key 展示粒度增强

继续沿 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 第 9.1.1 节推进，本轮把详情页中的 key 列表从单行摘要增强为更独立的 key 卡片视图：每个 key 现在会显式展示掩码 key、状态码、累计成本、source_type、备注、llowed_models 模型列表，以及最近使用时间；当 llowed_models 为空时，前端会明确显示“全部模型”，与后端的兼容语义保持一致。

这一轮仍然遵循“本地资源优先”的执行规则：直接复用已存在的 ChannelKey 字段与 channel.detail 文案空间，避免引入额外数据接口或偏离主线的大改。验证已通过 pnpm build（web）与 go test ./...；当前距离 MD 中“每个 key 独立搜索 / 测试 / 折叠 / 编辑”的完整目标，还剩更细的交互能力未做。

## 2026-04-16 增量记录：渠道页 key 搜索 / 折叠

继续按 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 第 9.1.1 节推进，本轮在渠道详情页为 key 列表补上了本地搜索与折叠展开能力。现在用户可以按 key 掩码、备注、source_type、llowed_models 与状态码对 key 做快速筛选；每个 key 也被收口为独立的 Accordion 项，默认只显示关键信息，展开后再查看完整的备注、来源类型、模型列表与最近使用时间。

本轮仍然保持“最小但贴主线”的实现策略：完全复用现有前端组件与本地数据，不新增后端接口、不改动路由语义。验证已通过 pnpm build（web）与 go test ./...。当前距离 MD 中“每个 key 独立测试 / 编辑”的完整终态，还剩更细的 key 级操作入口未补。

## 2026-04-16 增量记录：渠道页 key 级测试入口

继续沿 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 第 9.1 / 9.1.1 节推进，本轮在渠道详情页的每个 key 展开区内补上了“测试此 Key”入口，并把结果原地展示。实现方式复用了现有 useTestChannelModelsByConfig 接口：前端按单个 key 当前配置发起测试，优先测试该 key 的 llowed_models，若为空则回退到渠道模型列表，以最小改动补足“每个 key 都能单独测试”的能力。

本轮仍然没有引入新的后端接口或偏离主线的抽象层，重点是把已有能力更细粒度地暴露到渠道页 UI。验证已通过 pnpm build（web）与 go test ./...。当前距离 MD 中“每个 key 独立搜索、测试、折叠、编辑”的完整终态，主要还差 key 级编辑入口进一步显性化，以及后续分组页的分层展示增强。

## 2026-04-17 增量记录：渠道页 key 级编辑入口与分组页左栏筛选

继续按 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 第 9.1 / 9.2 / 9.3 节推进，本轮做了两段贴主线的前端增强。其一，在渠道详情页的每个 key 展开区内新增“编辑此 Key”入口，点击后会直接进入渠道编辑态，并自动滚动定位到对应 key 行、聚焦输入框，降低“多 key 渠道”下逐个编辑的摩擦。其二，在分组页左栏的模型选择区新增本地搜索筛选框，支持按渠道名和模型名过滤，以贴近 MD 中关于“搜索 / 筛选输入框”和大列表快速筛选的要求。

本轮仍然坚持本地资源优先，不新增后端接口，仅复用现有渠道表单、分组编辑器和前端本地数据完成增强。验证已通过 pnpm build（web）与 go test ./...。当前若继续贴主线推进，下一步更大的工作面会是分组页左栏按 classified / pooled 做真正的分层展示，而不只是纯搜索筛选增强。

## 2026-04-17 增量记录：分组页左栏按 classified / pooled 分层展示

继续按 docs/LLM-Gateway-Refactor-Plan.zh-CN.md 第 9.2.1 节推进，本轮把分组页左栏“添加模型”区域从原来的单层 channel -> models 提升为更贴近主线的层级视图。实现上，主线程复用 useModelChannelList() 与 useChannelList() 的现有前端数据，在左栏按 channel 的 key_management_mode 组织结构：classified 下优先按 key 的 llowed_models 拆成 channel -> key -> models，pooled 下则展示为 channel -> shared model set -> models。同时为 channel 增加了 classified / pooled 可视 badge，并对 classified 场景下未命中的模型补了 unassigned 兜底分组，避免静默丢失。

这一步遵循了“高价值但不破坏现有 group 保存结构”的策略：点击模型后的提交仍保持 channel/model 粒度，不引入新的后端接口与 group item 结构。子 agent（gpt-5.4）本轮已成功给出只读建议，主线程采用其“先做左栏结构可视化、不碰 group 提交语义”的结论并完成落地。验证已通过 pnpm build（web）与 go test ./...。后续若继续贴主线推进，下一步可考虑让分组页左栏进一步显式展示 key 备注 / source_type，或在保存结构允许后再讨论 key/model 级精确绑定。
