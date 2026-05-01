# 2026-04-27 深度审查：AI 自动化任务取消与生命周期收口

## 1. 任务信息

- 任务名称：AI 自动化任务取消与停机生命周期审查和小修
- 日期：2026-04-27
- 当前阶段：Phase A 工程稳定性收口 / Phase H AI 自动化中心风险复审
- 对应 milestone：里程碑 7 AI 自动化中心与动态路由 AI 学习

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 6、11、12、14 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、9.5、11.5 节
- 上一个相关 worklog：`docs/worklog/2026-04-27-deep-audit-update-restart-order-and-shutdown-nilsafe.md`
- 本次任务目标：确认并修复 AI 自动化任务取消只改 DB 状态、无法真正打断执行链的生命周期缺陷
- 本次已盘点本地资源：automation memory、当前状态/主规划/工作流文档、`internal/op/ai_automation*.go`、`internal/server/handlers/ai_automation*.go`、现有测试与 shutdown 收口代码
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、主规划、用户上下文总账、工作流文档、已有 AI 自动化测试
- 若未使用部分本地资源或上下文，原因：本轮聚焦高风险后端取消语义，未展开前端/UI 主线文档
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否，主线程串行执行
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求本自动化不要创建子 agent，且本轮缺陷位于单条后端链路，主线程可直接完成

## 3. 本次硬规则

- 优先处理安全/稳定性/生命周期高风险问题，不做大重构
- 小修必须可验证，并补最小必要测试
- 不覆盖工作区其他已有修改

## 4. 本次禁止事项

- 不改 AI Profile/动态路由主语义
- 不扩展到无关 UI 或文档重写
- 不做数据库破坏性迁移

## 5. 本次验收条件

- 取消中的 AI 任务必须真正触发执行上下文取消，而不是只改数据库状态
- 取消后的任务不得继续保存 `result_json`、`result_profile_id` 或把状态覆盖成 `failed/succeeded`
- 进程停机时应具备统一取消在途 AI 任务的收口入口
- 至少补充后端单元测试覆盖上述链路

## 6. 本次回滚点

- `internal/op/ai_automation_executor.go`
- `internal/op/ai_automation.go`
- `internal/op/ai_automation_test.go`
- `cmd/start.go`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改后端生命周期语义
- 受影响后端模块：`internal/op/ai_automation.go`、`internal/op/ai_automation_executor.go`、`cmd/start.go`
- 受影响前端模块：无
- 受影响接口：`POST /api/v1/ai/tasks/:id/cancel` 的真实取消语义
- 是否影响旧数据：否
- 是否影响旧行为：是，取消行为从“仅状态标记”收紧为“实际中断执行并同步收口步骤状态”

## 8. 实施步骤

1. 复核 `AITaskStartAsync()`、`executeAITask()`、`AITaskCancel()` 与 handler/test，确认取消链路只写 DB、不取消真实上下文。
2. 为运行中的 AI 任务增加进程内 cancel registry，取消时触发 `context.CancelFunc`，并把停机流程接到统一的 `CancelAllAITasks()`。
3. 增加回归测试，覆盖“在途请求取消后不保存结果”和“全局取消钩子生效”。

## 9. 测试与验证

- 构建命令：未单独执行 `go build ./...`
- 测试命令：`go test ./internal/op -run 'Test(AITaskCancelStopsInFlightExecution|CancelAllAITasksCancelsRunningContext|AITaskCreateExecutesAndSavesProfile|AITaskCreateExecutesProtectedActionsWhenAuthorized|AITaskCreateToolKeysLimitContextAndSkipProfileWrite|AIAutomationConfigGetUsesActiveProfileEffectiveConfig|AIAutomationConfigGetFallsBackToManualWhenActiveProfileInvalid|AIAutomationConfigGetFallsBackToManualWhenProfileContentInvalid)' -count=1`
- 专项验证：`gofmt -w internal/op/ai_automation_executor.go internal/op/ai_automation.go internal/op/ai_automation_test.go cmd/start.go`

## 10. 风险与兼容性

- 新风险：任务取消后，当前实现会把正在运行的 step 标记为 `failed(task canceled)`，其余 pending step 标记为 `skipped`；该语义是刻意保留可观测性的行为变化，需要后续前端若展示 step 状态时一并接受
- 兼容性风险：`cmd/start.go` 当前工作区本就有未提交改动，本轮只在其上追加 `CancelAllAITasks` 注册，没有回退其他差异
- 是否阻塞下一任务：不阻塞；但 handler 侧复验仍需在可下载依赖的环境补跑

## 11. 收工记录

- 构建是否通过：未执行完整构建
- 测试是否通过：与本次修复直接相关的 `internal/op` 定向测试通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、现有 AI 自动化测试
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：上一轮 memory 指明 AI 任务取消/停机生命周期仍是高优先级风险；主规划要求 AI Profile 不覆盖手动配置且需可回退；现有测试提供了最小回归基线
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮属纯后端生命周期修复，无需手工 smoke；handler 包测试受离线依赖下载限制未补跑
- 待验证页面清单：AI 自动化任务页取消按钮的实时状态展示、step 状态文案、任务取消后的列表刷新
- 若未使用子 agent，原因：用户要求本自动化主线程串行执行，不创建子 agent
- worklog 是否更新：是
- 遗留项：`internal/server/handlers` 包测试仍因 `github.com/tmaxmax/go-sse` 离线下载失败无法复验；Docker 最小权限面、自更新完整性校验仍是更高层剩余风险
- 下一任务前置条件是否满足：满足，下一轮可继续深审 handler/relay/Docker 风险

## Findings 与处理

1. `internal/op/ai_automation.go:388`、`internal/op/ai_automation_executor.go:142`
   - 问题：`AITaskCancel()` 原先只把 `AITask.status` 更新为 `canceled`，异步执行 goroutine 仍跑在 `context.Background()` 派生上下文上；正在进行的外部 AI 请求不会立即中断。
   - 影响：取消后的任务仍可能继续请求上游 AI、继续解析结果甚至继续落盘，造成“用户以为取消成功，但后台仍执行”的权限/资源边界偏差。
   - 证据：取消路径没有持有任何 `CancelFunc`；异步启动时直接 `context.WithTimeout(context.Background(), 2*time.Minute)`，DB 状态与 goroutine 生命周期脱钩。
   - 处理：已修复。新增进程内 `aiTaskCancelFuncs` 注册表，`AITaskStartAsync()` 存储每个任务的 `CancelFunc`，`AITaskCancel()` 先触发真实取消，再更新任务与步骤状态；`cmd/start.go` 同时把 `op.CancelAllAITasks` 注册到 shutdown 链路。

2. `internal/op/ai_automation_executor.go:190-231`
   - 问题：取消错误路径原先用 `err == context.Canceled` 精确比较，并在已取消上下文上继续回写步骤状态/任务失败，容易遗漏包装后的取消错误，或因上下文已失效导致步骤状态收口不完整。
   - 影响：取消时可能把任务误记为 `failed`、步骤长期停在 `running`，增加可观测性噪音，且可能误触发失败处理分支。
   - 证据：原逻辑仅比较 `== context.Canceled`；`finishAITaskFailure()`、步骤回写等都沿用原 ctx。
   - 处理：已修复。统一改为 `errors.Is(err, context.Canceled)`；取消后的步骤清理和失败收口改用新的 background timeout context 执行，避免被已取消的请求上下文吞掉。

3. `internal/op/ai_automation_test.go:921`、`:972`
   - 问题：原测试未覆盖“在途请求取消”与“全局停机取消”两条高风险生命周期分支。
   - 影响：之前即使取消只是改数据库状态，也难以及时在回归里暴露。
   - 证据：现有测试覆盖创建/成功/受保护动作，但没有阻塞中的取消用例。
   - 处理：已补测试 `TestAITaskCancelStopsInFlightExecution` 与 `TestCancelAllAITasksCancelsRunningContext`，并复验现有成功路径测试未回退。
