# 2026-04-29 Phase H AI learning timeout comparison command closure

## 1. 任务信息

- 任务名称：Phase H6 learning timeout comparison command closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-self-start-comparison-command-closure.md`
- 本次任务目标：在确认 self-start command-order 对照不会改变 attached-session 失败层后，把最后一轮 bounded timeout 对照命令固化到 `check-only` 输出
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/worklog/2026-04-29-phase-h-ai-learning-self-start-comparison-command-closure.md`、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic 副本、`runtime-win.ps1` 状态输出
- 若未使用部分本地资源或上下文，原因：本轮不需要回到产品 UI、settings 页面、后端 schema 或宿主网络根因修复；工作继续收敛在 wrapper follow-up 命令层
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / 状态文档主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 本轮必须先拿到真实 self-start command-order 对照结果，再决定是否补 timeout 命令出口，不能只凭 check-only 推测
- 只增强 `check-only` 的 timeout follow-up 命令输出，不改变 external preflight schema、repo-local stable artifact 命名或产品行为

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到新的 CDP wrapper 架构改造、后端接口变更、Edge 启动策略重做或动态路由业务逻辑
- 不伪造 stable artifact，也不重跑无关 external/browser smoke

## 5. 本次验收条件

- 真实执行 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + page-lifecycle-runtime` 外部对照，确认 command order 切换后仍停在同一 attached-session timeout 层
- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会额外打印 `External preflight CDP timeout comparison command` 与 `External preflight self-start CDP timeout comparison command`
- timeout 对照命令会复用当前 invocation profile，并自动把 `CdpCommandTimeoutMs` 放大到 45000、`NodeSmokeTimeoutSeconds` 提升到与该超时相容的 155
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-timeout-comparison-command-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI 或外部数据；只改 wrapper follow-up 命令层与静态守护
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser verifier `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 在 attached-session CDP 对照场景下的 timeout follow-up 命令输出

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态与上一轮 H6 worklog，确认当前主线仍是 attached-session CDP bootstrap 对照，而不是回到 artifact coverage 或 UI 主线。
2. 先执行 `runtime-win.ps1 status` 与两次 `check-only`，确认 repo-local optional/require-cdp 两条 profile 都已经对齐在 `cdp_smoke_failed / failed checks cdp`，且 wrapper 已能输出 self-start comparison 命令。
3. 真实执行 optional 与 require-cdp 两条 `self-start + page-lifecycle-runtime` live comparison，确认它们都稳定停在 `page_bootstrap_timeout_attached_session`，没有因为 bootstrap order 反转而出现新的运行层。
4. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中新增 timeout helper：`Get-ExternalPreflightSuggestedComparisonCommandTimeoutMs`、`Get-ExternalPreflightSuggestedNodeSmokeTimeoutSeconds` 与 `Get-ExternalPreflightTimeoutComparisonCommand`，并把结果接入 `Write-StableExternalPreflightDiagnosticPreview` / `Write-ExternalPreflightDiagnosticSummary` 的新输出标签。
5. 同步更新 `scripts/verify-ai-automation-learning-focus.mjs`，把新 helper、timeout 输出标签、推荐的 `45000 / 155` 参数与新的调用签名纳入静态守护。
6. 复跑默认与 `-RequireExternalCdpPreflight` 两次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit`、`git diff --check` 与 `runtime-win.ps1 -Action status`，确认本轮没有打断 Phase H6 repo-local replay 验证链；最后同步更新状态文档、前端主线状态、automation memory 与本 worklog。

## 9. 测试与验证

- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 按预期失败并落到同层级 `cdp_smoke_failed`：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -CdpBootstrapCommandOrder 'page-lifecycle-runtime' -SelfStartServices`
- 按预期失败并落到同层级 `cdp_smoke_failed`：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight -CdpBootstrapCommandOrder 'page-lifecycle-runtime' -SelfStartServices`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-29-phase-h-ai-learning-timeout-comparison-command-closure.md`

## 10. 风险与兼容性

- 新风险：低；timeout comparison command 只作用于 `check-only` 摘要解释层
- 兼容性风险：低；仅当 `aligned_cdp_bootstrap_focus` 成立时才会输出 timeout comparison command，其余路径保持原有 comparison/refresh/next-step 逻辑
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（runtime status、两次 real self-start external comparison、两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic 目录、`runtime-win.ps1`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：上一轮 H6 worklog 与当前 `check-only` 输出说明 self-start comparison 命令已经具备可直接执行的 host-ready 形式；真实对照结果进一步证明 command order 反转后仍停在同一 attached-session 层，因此本轮最值得推进的是把更长 timeout 的 bounded rerun 命令固化下来，而不是再回头扩散到 UI/runtime 其它主题
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：已做当前主线需要的两条真实 self-start external 对照；未新增产品页手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：当前 remaining blocker 仍集中为 attached-session 下 `Page.enable / Page.setLifecycleEventsEnabled / Runtime.enable` 30s timeout；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；下一轮继续同一条 H6 wrapper/CDP 诊断链
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 下一轮优先复制新的 `External preflight self-start CDP timeout comparison command`，确认 `45000ms` 更长 per-command timeout 是否仍停在同一 attached-session 层
  - 若更长 timeout 仍无差异，再决定是把该宿主正式定性为 attached-session bootstrap blocker，还是在同主线下再做最后一轮更高超时的 bounded 对照
- 下一任务前置条件是否满足：满足；下一轮可直接从 self-start timeout comparison command 出发做最后一轮 bounded host-level 对照
