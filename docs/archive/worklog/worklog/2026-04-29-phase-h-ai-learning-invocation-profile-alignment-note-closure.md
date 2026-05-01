# 2026-04-29 Phase H AI learning invocation profile alignment note closure

## 1. 任务信息

- 任务名称：Phase H6 learning invocation profile alignment note closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-require-cdp-bridge-artifact-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable replay 主线，把“optional / require-cdp 两条 invocation profile 是否已经落在同一运行层”从手工对照两次 `check-only` 输出，收口成一条直接可消费的 `invocation profile alignment note`
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 副本、`runtime-win.ps1` 状态输出
- 若未使用部分本地资源或上下文，原因：本轮不需要重开 AI 页面 UI、settings 页面、后端 schema 或新的 external/browser 联调；repo-local stable replay 已足够闭环“下一轮先刷新哪个 profile”这件事
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / 状态文档主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 本轮必须形成真实代码增量与直接验证，不能只更新文档或复述阻塞
- 只增强 repo-local stable replay 的对齐判断与静态守护，不改变 external preflight schema、artifact 命名、产品行为或共享 CDP wrapper 主控制流

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到真实 external/browser 复跑或新的宿主修复
- 不伪造新的 stable artifact；本轮只消费已落盘的 optional / require-cdp 两份真实副本

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 会额外打印 `External preflight invocation profile alignment note`
- 当 matched profile 与 opposite profile 落在不同服务可达层时，note 会直接说明哪一边仍停在 `preflight_failed`，哪一边已经进入 `cdp_smoke_failed`
- 默认 `check-only` 与 `-RequireExternalCdpPreflight` 两次回放都会给出对称的对齐判断与下一步刷新建议
- `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-invocation-profile-alignment-note-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI 或外部数据；只改 wrapper 说明层与静态守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 摘要，帮助判断两条 invocation profile 是否已到同一运行层

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 Phase H6 worklog 与 `runtime-win.ps1 -Action status`，确认当前最值得推进的是“给下一轮一个明确的 profile 刷新顺序”，而不是继续改 UI 或重跑 external。
2. 复跑默认 `check-only` 与 `-RequireExternalCdpPreflight` 的 `check-only`，确认当前 repo-local 真实状态已经分叉为：optional profile 仍是 `preflight_failed/backend+frontend unreachable`，require-cdp profile 已是 `cdp_smoke_failed/cdp`。
3. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Get-ExternalPreflightDiagnosticSummaryFragment`、`Get-ExternalPreflightServiceReachabilityState` 与 `Get-StableExternalPreflightInvocationProfileAlignmentNote`，从 matched/alternate 两份 parseable stable diagnostic 中提炼“服务可达层是否对齐”的摘要。
4. 让 `Write-ExternalPreflightDiagnosticSummary` 在 `stable coverage note` 后、`decision summary` 前输出新的 `External preflight invocation profile alignment note`。
5. 同步更新 `scripts/verify-ai-automation-learning-focus.mjs`，把新 helper、输出标签和关键摘要语句纳入静态守护。
6. 运行默认 `check-only`、`-RequireExternalCdpPreflight`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check`，确认新的 note 在当前真实 repo-local 副本上能稳定指出“先刷新 optional profile，再谈 CDP bootstrap”。
7. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，给下一轮留下“先把 optional external profile 刷新到 backend/frontend 已可达层”的直接入口。

## 9. 测试与验证

- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`

## 10. 风险与兼容性

- 新风险：低；本轮只增加 `check-only` 的摘要说明层
- 兼容性风险：低；当 opposite profile 不存在或不可解析时，新的 alignment note 会静默跳过，不影响现有 replay 流程
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（`runtime-win status`、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 目录、`runtime-win.ps1`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：repo-local stable diagnostic 目录与两次 `check-only` 明确显示 optional / require-cdp 两条 profile 的运行层不一致，因此本轮最值得推进的是先把“刷新顺序”输出成机器可读的人类提示，而不是继续扩展新的 external 参数面
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未做新的真实 external/browser 复跑；本轮只消费已落盘的 optional / require-cdp stable replay
- 手工 smoke 阻塞原因 / 缺少的环境：optional profile 仍停在旧的 `host_networking_blocker` 运行层；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；下一轮继续同一条 H6 诊断链
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 仍需执行一次默认 optional external profile 的真实刷新，让 `requireCdp=false` replay 也进入 backend/frontend 已可达的同一运行层
  - 只有在 optional profile 刷新到同层后，继续比较 `attached-session` CDP bootstrap timeout 的调参价值才更可靠
- 下一任务前置条件是否满足：满足；下一轮优先继续同一条 Phase H6 诊断链，先刷新默认 optional external profile，再决定是否继续比较 CDP bootstrap 策略
