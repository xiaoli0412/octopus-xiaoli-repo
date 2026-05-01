# 2026-04-29 Phase H AI learning bootstrap comparison command closure

## 1. 任务信息

- 任务名称：Phase H6 learning bootstrap comparison command closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-aligned-cdp-bootstrap-focus-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable replay 主线，把“两个 invocation profile 已对齐后，下一条 attached-session CDP bootstrap live rerun 该怎么跑”收口成 repo-local `check-only` 的固定输出命令
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/worklog/2026-04-29-phase-h-ai-learning-aligned-cdp-bootstrap-focus-closure.md`、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 副本、`runtime-win.ps1` 状态输出
- 若未使用部分本地资源或上下文，原因：本轮不需要重开 AI 页面 UI、settings 页面、后端 schema 或宿主网络修复；工作继续收敛在 wrapper 诊断输出层
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / 状态文档主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 本轮必须形成真实代码增量与直接验证，不能只更新文档或复述阻塞
- 只增强 `check-only` 的命令级 follow-up 出口，不改变 external preflight schema、stable artifact 命名或产品行为

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到新的 CDP wrapper 架构改造、后端接口变更或动态路由业务逻辑
- 不伪造 stable artifact，也不重跑无关 external/browser smoke

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 在 `aligned_cdp_bootstrap_focus` 场景下会额外打印 `External preflight CDP bootstrap comparison command`
- 默认与 `-RequireExternalCdpPreflight` 两条 invocation profile 都会生成保持相同 profile、仅切换 `-CdpBootstrapCommandOrder` 的 external comparison command
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-bootstrap-comparison-command-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI 或外部数据；只改 wrapper 说明层与静态守护
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser verifier `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 在对齐后 CDP bootstrap 场景下的 follow-up 命令输出

## 8. 实施步骤

1. 复核 automation memory、主规划、当前状态文档、前端主线状态、上一轮 `aligned_cdp_bootstrap_focus` worklog 与当前 repo-local stable artifact，确认当前主阻塞已经稳定收敛为 attached-session `Runtime.enable / Page.enable` timeout，而不是 profile 未对齐。
2. 复跑默认与 `-RequireExternalCdpPreflight` 两次 `check-only`，确认当前输出虽然已给出 `aligned_cdp_bootstrap_focus`、decision summary 与 recommended action，但下一条 live rerun 命令仍需人工从当前 invocation 参数中手工拼接。
3. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Get-ExternalPreflightAlternateBootstrapCommandOrder` 与 `Get-ExternalPreflightBootstrapComparisonCommand`，基于当前 `PSBoundParameters` 复用同一 invocation profile，仅翻转 `CdpBootstrapCommandOrder` 生成 comparison command。
4. 扩展 `Get-ExternalPreflightRecommendedRefreshCommand` 支持 `Overrides`，避免 comparison command 需要重复拼接所有已绑定的 external 参数。
5. 在 `Write-ExternalPreflightDiagnosticSummary` 中接线 `External preflight CDP bootstrap comparison command`，只在 `aligned_cdp_bootstrap_focus` 分支下输出，确保这条命令与当前 invocation 的 `requireCdp`、host-friendly defaults 和其它参数保持一致。
6. 同步更新 `scripts/verify-ai-automation-learning-focus.mjs`，把新 helper、comparison command 输出标签、`Overrides` 调用与 `check-only` / summary 调用签名纳入静态守护。
7. 运行两次 learning smoke `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check`，确认本轮没有打断 Phase H6 repo-local replay 验证链；最后同步更新状态文档、前端主线状态、automation memory 与本 worklog。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs`

## 10. 风险与兼容性

- 新风险：低；comparison command 只作用于 `check-only` 摘要解释层
- 兼容性风险：低；仅当 `aligned_cdp_bootstrap_focus` 成立时才会输出 comparison command，其余路径保持原有 refresh command / next steps 逻辑
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（两次 sequential `check-only`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic 目录、`runtime-win.ps1`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：上一轮 H6 worklog 与当前 `check-only` 输出说明 profile 对齐工作已完成，因此本轮最值得推进的是把 attached-session bootstrap 比较命令固定下来；repo-local stable diagnostics 证明 optional/require-cdp 两条 profile 都已停在 `cdp_smoke_failed / failed checks cdp`，适合补 command-level follow-up，而不是回头刷新 artifact
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未做新的真实 external/browser 复跑；继续使用 repo-local `check-only` 收口对齐后 CDP bootstrap 比较入口
- 手工 smoke 阻塞原因 / 缺少的环境：当前 remaining blocker 仍集中为 attached-session 下 `Runtime.enable / Page.enable` 30s timeout；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；下一轮继续同一条 H6 wrapper/CDP 诊断链
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 下一轮优先执行新输出的 `External preflight CDP bootstrap comparison command`，确认 `page-lifecycle-runtime` 在 attached-session 下是否仍与当前 `runtime-page-lifecycle` 一样卡在 `Page.enable / Runtime.enable`
  - 若 command-order 对照仍同样失败，再决定是否值得在同一 profile 下继续放大 `-CdpCommandTimeoutMs`，还是直接把该宿主定性为 attached-session bootstrap blocker
- 下一任务前置条件是否满足：满足；下一轮可直接从 comparison command 出发做 host-level attached-session bootstrap 对照
