# 2026-04-27 Phase H AI Config Source Contract Fallback Closure

## 1. 任务信息

- 任务名称：ai config source contract fallback closure
- 日期：2026-04-27
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H5/H6 设置页来源切换与运行时 consumer 收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 验收标准 19-21
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-h-setting-ai-profile-source-card-closure.md`
  - `docs/worklog/2026-04-27-phase-h-setting-ai-profile-risk-and-fallback-closure.md`
- 本次任务目标：把 AI 配置来源的“请求态 / 生效态 / fallback 原因”收敛成后端显式 contract，并让设置页与 AI 自动化页消费同一语义
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、AI 自动化需求/实施计划、动态路由需求/实施计划、当前状态文档、上一轮 Phase H worklog、`internal/op/ai_automation.go`、`internal/op/ai_automation_test.go`、`internal/server/handlers/ai_automation_test.go`、`web/src/components/modules/setting/AIAutomationSource.tsx`、`web/src/components/modules/ai-automation/index.tsx`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H5/H6 主线，不扩散到无关模块
- AI Profile 只能切换读取来源，不能覆盖手动配置
- 动态路由学习语义保持不变，本轮只收敛 AI 配置来源 contract
- 设置页与 AI 自动化页必须消费同一后端事实，不能继续靠前端双重推断回退状态

## 4. 本次禁止事项

- 不改 AI task 执行语义
- 不改手动配置表或 AI Profile 激活行为的主逻辑
- 不把“运行时 fallback”误做成静默覆盖 persisted setting

## 5. 本次验收条件

- `AIAutomationConfigGet` 显式返回 requested/effective/fallback 语义
- 设置页与 AI 自动化页都可展示“请求 AI Profile 但运行时已回退 manual”的状态
- 相关 Go 测试通过
- 前端 `tsc --noEmit`、locale consistency 与 `git diff --check` 通过

## 6. 本次回滚点

- `internal/model/ai_automation.go`
- `internal/op/ai_automation.go`
- `internal/op/ai_automation_test.go`
- `internal/server/handlers/ai_automation_test.go`
- `web/src/api/endpoints/ai-automation.ts`
- `web/src/components/modules/setting/AIAutomationSource.tsx`
- `web/src/components/modules/setting/AIAutomationSource.test.tsx`
- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/components/modules/ai-automation/index.test.tsx`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/ja.json`

## 7. 实现范围

- 后端：`AIAutomationConfig` 新增 `requested_config_source_mode`、`requested_active_ai_profile_id`、`source_fallback_reason`
- 后端：`applyActiveAIProfileConfig` 改为保留 requested setting，不再在自动 fallback 时静默改写 persisted setting
- 前端：设置页来源卡片改为直接消费后端 fallback 语义，并展示细分 fallback reason
- 前端：AI 自动化页顶部摘要卡片同步展示 requested source 与 runtime fallback 状态
- 测试：补齐 op/handler/setting page/ai automation page 的 contract 断言

## 8. 实施步骤

1. 复核 `AIAutomationConfigGet`、`fallbackManualAIAutomationConfig` 与前端 consumer，确认现有缺口是“前端只能靠 settings/config 不一致自行猜 fallback”。
2. 为 `AIAutomationConfig` 增加 requested/effective/fallback 字段，并在 `AIAutomationConfigGet` / `applyActiveAIProfileConfig` 中产出显式语义。
3. 调整 fallback 行为：保留用户请求值，仅在返回 payload 中回退到 manual，而不再静默改写 setting 表。
4. 更新 Go 测试，覆盖正常 AI Profile 生效、缺失 profile fallback、无效内容 fallback，以及 handler 层返回语义。
5. 更新前端 API 类型、设置页来源卡片、AI 自动化页摘要卡片与相关测试，统一显示 fallback 原因。
6. 补齐四语 locale 的 `fallbackReasons.*` 文案。

## 9. 测试与验证

- 已执行：`. .\scripts\use-go-env.ps1; & $env:GOEXE test ./internal/op ./internal/server/handlers`
- 已执行：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 已执行：`node scripts/verify-locale-consistency.mjs`
- 已执行：`git diff --check -- ...`
- 已执行：`. .\scripts\use-go-env.ps1; & $env:GOFMTEXE -w internal/model/ai_automation.go internal/op/ai_automation.go internal/op/ai_automation_test.go internal/server/handlers/ai_automation_test.go`
- 未通过：`node -r .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run .\web\src\components\modules\setting\AIAutomationSource.test.tsx .\web\src\components\modules\ai-automation\index.test.tsx --config .\web\vitest.config.ts`
  - 失败原因：宿主 `vite/esbuild` 在加载 `vitest.config.ts` 阶段仍报 `spawn EPERM`，属于当前 Windows 宿主对子进程创建的限制，不是新增断言失败

## 10. 风险与兼容性

- 新风险：后端现在会保留 requested source setting，因此管理端在读取 AI 配置时会显式看到“请求 AI Profile，但当前运行时 manual fallback”的双态信息；这属于设计收敛，不是行为回退
- 兼容性风险：低；未改 AI Profile 激活入口、未改手动配置表、未改动态路由学习开关协议
- 仍存风险：前端相关 vitest 仍无法在当前宿主补跑浏览器/配置级单测，只能依赖 `tsc`、locale 和 Go 测试作为本轮可执行验证

## 11. 收工记录

- 构建是否通过：通过
  - Go 相关测试通过
  - 前端 TypeScript 检查通过
  - locale consistency 通过
- 测试是否通过：部分通过
  - 通过：`go test ./internal/op ./internal/server/handlers`
  - 通过：`tsc --noEmit -p web/tsconfig.json`
  - 通过：`verify-locale-consistency.mjs`
  - 通过：`git diff --check`
  - 受宿主阻塞未完成：前端 vitest（`spawn EPERM`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、AI 自动化需求/实施计划、动态路由需求/实施计划、当前状态文档、上一轮 Phase H worklog、automation memory、`using-superpowers` / `brainstorming` skill 文档、相关前后端源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - canonical / AI 自动化文档：确认 H5/H6 必须保证显式切换与无覆盖语义
  - 上一轮 worklog / memory：确认下一步应从“设置页显示回退”推进到“contract 收口，减少前端自行推断”
  - 源码审查：确认真实缺口在 `AIAutomationConfigGet` 缺少 requested/fallback 语义
- worklog 是否更新：已更新
- 遗留项：
  - 当前宿主 `vitest/vite/esbuild` 仍会在 config 启动阶段 `spawn EPERM`
  - 更大的 AI Profile consumer 闭环仍主要停留在 AI 自动化执行配置层，尚未延展到更多配置消费域
- 下一任务前置条件是否满足：满足；下一轮可优先补前端单测环境证据，或继续同主线推进 AI Profile 的更大 consumer 闭环
