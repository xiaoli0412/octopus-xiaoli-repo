# 2026-04-29 Phase H AI learning alternate execution path command closure

## 1. 任务信息

- 任务名称：Phase H6 learning alternate execution path command closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-attached-session-blocker-confirmed-closure.md`
- 本次任务目标：在 attached-session blocker confirmed 已成立的前提下，把“切换到不同执行路径”的下一条命令固化到 `check-only` 输出，避免接手人继续手工拼 `json-new` 对照命令或误退回默认 `30000ms` 预算
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/worklog/2026-04-29-phase-h-ai-learning-attached-session-blocker-confirmed-closure.md`、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、共享 CDP wrapper `scripts/verify-channel-create-browser-smoke-cdp.ps1/.mjs`、repo-local stable diagnostic 目录与当前 optional/require-cdp stable JSON
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic 副本、共享 CDP wrapper
- 若未使用部分本地资源或上下文，原因：本轮不涉及产品 UI、settings 页面、后端 schema 或新的 live browser 取证；工作继续收敛在 wrapper 解释层和接手命令生成层
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / host-level CDP bootstrap 诊断主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 只增强 `check-only` 的命令生成与收口说明，不改变 external preflight schema、repo-local stable artifact 命名或产品行为
- 新命令必须继承当前 confirmed blocker 的有效证据层级，不能回退到默认 `30000ms` timeout 而误导下一轮取证

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到新的 CDP wrapper 架构改造、后端接口变更、Edge 启动策略重做或动态路由业务逻辑
- 不伪造 stable artifact，也不重跑无关 external/browser smoke

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 在 `attached_session_bootstrap_blocker_confirmed` 场景下会新增 `External preflight alternate execution path command` 与 `External preflight self-start alternate execution path command`
- 新命令保留当前 invocation profile、保留 `60000ms / 200s` 预算，只切换 `-CdpPageBootstrapStrategy 'json-new'`
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、PowerShell parser 校验、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-alternate-execution-path-command-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI 或外部数据；只改 wrapper 命令推荐层与静态守护
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser verifier `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 在 attached-session blocker confirmed 终态下的接手命令输出

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态与上一轮 H6 worklog，确认当前主线仍是 blocker confirmed 后的交接出口收口，而不是重新回到 timeout 调参。
2. 运行 `runtime-win.ps1 status`，确认本轮不需要 live rerun，项目保持默认停驻状态。
3. 复查 `verify-ai-automation-learning-browser-smoke.ps1` 的 refresh/bootstrap/timeout command 生成链，以及共享 CDP wrapper 对 `auto / json-new / attached-session` 的 page strategy 语义，确认 `json-new` 是同主线下最合适的替代执行路径。
4. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Get-ExternalPreflightAlternatePageBootstrapStrategy`、`Get-ExternalPreflightAlternateExecutionPathCommand` 与 `Get-ExternalPreflightEffectiveAlternateExecutionPathCommand`，并在 `Write-ExternalPreflightDiagnosticSummary` 的 confirmed blocker 分支下新增 `External preflight alternate execution path command` / `self-start` 版本输出。
5. 确保 `check-only` 基于 repo-local stable diagnostic 中已记录的 `commandTimeoutMs` 生成替代路径命令，避免退回 `check-only` 默认的 `30000ms` 预算。
6. 同步更新 `scripts/verify-ai-automation-learning-focus.mjs`，补上新 helper、输出标签和调用签名守护。
7. 复跑 PowerShell parser 校验、两次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check`，最后同步更新状态文档、前端主线状态、automation memory 与本 worklog。

## 9. 测试与验证

- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 通过：`powershell -NoProfile -Command '$tokens=$null; $errors=$null; [void][System.Management.Automation.Language.Parser]::ParseFile("D:\GPT-codex\octopus_repo\scripts\verify-ai-automation-learning-browser-smoke.ps1",[ref]$tokens,[ref]$errors); if ($errors -and $errors.Count -gt 0) { $errors | ForEach-Object { $_.ToString() }; exit 1 }'`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-29-phase-h-ai-learning-alternate-execution-path-command-closure.md`

## 10. 风险与兼容性

- 新风险：低；alternate execution path command 仅作用于 `check-only` 摘要解释层
- 兼容性风险：低；只在 `attached_session_bootstrap_blocker_confirmed` 分支输出，且保留原有 refresh command / replay 结论
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（`runtime-win status`、PowerShell parser、两次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON、共享 CDP wrapper
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：上一轮 H6 worklog 说明当前 timeout lane 已闭环为 blocker confirmed，因此本轮最值得推进的是把“切换执行路径”的下一条命令固化下来；repo-local stable diagnostics 提供了真实 `60000ms` 级别的 `commandTimeoutMs`，用于避免新命令回退到 `check-only` 默认预算；共享 CDP wrapper 证明 `json-new` 是同主线下可直接比较的 page strategy。
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：本轮未新增 live smoke；仅复用 repo-local stable diagnostic 做 `check-only` 收口
- 手工 smoke 阻塞原因 / 缺少的环境：当前 remaining blocker 仍是 attached-session 下 `Runtime.enable / Page.enable / Page.setLifecycleEventsEnabled` 60s timeout；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；下一轮优先比较 `json-new` alternate execution path 或切换宿主
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 下一轮优先复制 `External preflight alternate execution path command`，验证 `json-new` page strategy 是否仍停在同一 host-level 失败层，或是否能把阻塞前移到不同运行层
  - 若 `json-new` 也不能提供新的运行层，再把当前宿主 blocker 作为已确认前提，切到不同宿主继续取证
- 下一任务前置条件是否满足：满足；当前 H6 attached-session timeout lane 与替代路径交接出口都已闭环，下一轮可直接基于现成命令继续同主线推进
