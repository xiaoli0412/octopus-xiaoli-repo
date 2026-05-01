# 2026-04-29 Phase H AI learning attached-session blocker candidate closure

## 1. 任务信息

- 任务名称：Phase H6 learning attached-session blocker candidate closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-timeout-comparison-command-closure.md`
- 本次任务目标：真实执行 default optional 与 `-RequireExternalCdpPreflight` 的 `self-start + 45000ms` live rerun，确认两条 invocation profile 是否仍卡在同一 attached-session bootstrap timeout 层；若确认无差异，则把 wrapper 的结论收口到“attached-session bootstrap blocker candidate + 60000ms final bounded rerun”这一层
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/worklog/2026-04-29-phase-h-ai-learning-timeout-comparison-command-closure.md`、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 目录与最新 optional/require-cdp stable JSON
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic 副本、`runtime-win.ps1` 输出、`brainstorming` skill（仅用于先收敛这个小闭环的设计，不扩展到产品实现）
- 若未使用部分本地资源或上下文，原因：本轮不涉及产品 UI、后端 schema、settings 交互或宿主网络根因修复；工作继续收敛在 wrapper follow-up 命令、分类层和 host-level blocker 判定
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / host-level CDP bootstrap 诊断主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 本轮必须先拿到 default optional 与 `requireCdp=true` 两条 `45000ms self-start` live rerun 结果，再决定是否继续放大 timeout，不能只靠 check-only 推测
- 只增强 `check-only` 的 follow-up 命令与分类语义，不改变 external preflight schema、repo-local stable artifact 命名或产品行为

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到新的 CDP wrapper 架构改造、后端接口变更、Edge 启动策略重做或动态路由业务逻辑
- 不伪造 stable artifact，也不重跑无关 external/browser smoke

## 5. 本次验收条件

- 真实执行 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + 45000ms` live rerun，并确认两条 invocation profile 都仍稳定停在 `page_bootstrap_timeout_attached_session`
- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 现在会基于已保存的 matching `45000ms` diagnostic 自动把 `External preflight CDP timeout comparison command` 与 `External preflight self-start CDP timeout comparison command` 提升到 `-CdpCommandTimeoutMs '60000' -NodeSmokeTimeoutSeconds '200'`
- `check-only` 的 `External preflight recommended action class` 切到 `attached_session_bootstrap_blocker_candidate`
- `decision summary / recommended action` 会明确说明“当前宿主已经具备 attached-session bootstrap blocker 候选证据，只剩最后一轮 bounded timeout 对照是否还值得执行”
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、PowerShell parser 校验、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-attached-session-blocker-candidate-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI 或外部数据；只改 wrapper 分类/命令推荐层与静态守护
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser verifier `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 在 attached-session CDP 对照场景下的 blocker-candidate 分类与 timeout follow-up 命令输出

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态与上一轮 H6 worklog，确认当前主线仍是 attached-session CDP bootstrap bounded timeout lane，而不是回到 artifact coverage 或产品 UI 主线。
2. 先执行 `runtime-win.ps1 status` 与两次 `check-only`，确认 repo-local optional/require-cdp 两条 profile 都已经对齐在 `cdp_smoke_failed / failed checks cdp`，且 wrapper 已能输出 `45000ms` timeout comparison command。
3. 真实执行 default optional 与 `-RequireExternalCdpPreflight` 两条 `self-start + 45000ms` live rerun。中途发现并行跑会争抢 `18081`，因此先精确停掉运行态，再串行补跑 require-cdp 版本，确认两条 invocation profile 都仍停在同一 `page_bootstrap_timeout_attached_session` 层。
4. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Get-ExternalPreflightAttachedSessionBootstrapTimeoutState`、`Get-StableExternalPreflightAttachedSessionBlockerCandidateState` 与 `Get-ExternalPreflightEffectiveTimeoutComparisonCommand`，让 `check-only` 可以根据已保存 diagnostic 里的 `cdpDiagnostic.commandTimeoutMs` 推导下一档 timeout command，而不是永远沿用当前命令行入参。
5. 同步把 `decision summary / recommended action class / recommended action` 收口到新的 `attached_session_bootstrap_blocker_candidate` 分支：当 matching 与 alternate 两条 profile 都已经是 `services reachable + cdp only failure + attached-session timeout + timeout>=45000ms` 时，直接提示“仅剩一轮 `60000ms` bounded rerun，之后即可记录宿主 blocker”。
6. 更新 `scripts/verify-ai-automation-learning-focus.mjs` 的静态守护，补上新 helper、新 action class、新文案以及新的 `check-only` 调用签名。
7. 复跑 PowerShell parser 校验、两次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check`，最后同步更新状态文档、前端主线状态、automation memory 与本 worklog。

## 9. 测试与验证

- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 按预期失败并落到同层级 `cdp_smoke_failed`：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -NodeSmokeTimeoutSeconds 155 -CdpCommandTimeoutMs 45000 -SelfStartServices`
- 首次并行补跑 `requireCdp=true` 时按预期发现宿主噪音：`Backend self-start exited early ... listen tcp 0.0.0.0:18081: bind`；已通过 `runtime-win.ps1 stop` 精确回收并改为串行重试，不将其误判为主线产品问题
- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop`
- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 按预期失败并落到同层级 `cdp_smoke_failed`：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight -NodeSmokeTimeoutSeconds 155 -CdpCommandTimeoutMs 45000 -SelfStartServices`
- 通过：`powershell -NoProfile -Command '$tokens=$null; $errors=$null; [void][System.Management.Automation.Language.Parser]::ParseFile("D:\GPT-codex\octopus_repo\scripts\verify-ai-automation-learning-browser-smoke.ps1",[ref]$tokens,[ref]$errors); if ($errors -and $errors.Count -gt 0) { $errors | ForEach-Object { $_.ToString() }; exit 1 }'`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`

## 10. 风险与兼容性

- 新风险：低；新 action class 与 timeout suggestion 只作用于 `check-only` 摘要解释层
- 兼容性风险：低；仅当 matching 与 alternate 两条 profile 都已经达到 `services reachable + cdp only failure + attached-session timeout + timeout>=45000ms` 时，才会切到 `attached_session_bootstrap_blocker_candidate`
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（两次真实 `45000ms` self-start external rerun、两次 `check-only`、PowerShell parser 校验、`verify-ai-automation-learning-focus.mjs`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON、`runtime-win.ps1`、`brainstorming` skill
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：repo-local stable JSON 明确证明 optional 与 require-cdp 两条 profile 都已经在 `45000ms` 下保存了 matching attached-session timeout 证据，因此下一步最值得推进的是把 timeout suggestion 从“继续盲试 45000ms”提升到“最后一轮 60000ms bounded rerun”，并在 action class 层把宿主正式推进到 blocker-candidate 级别
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：已做当前主线需要的两条真实 self-start external 对照；未新增产品页手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：当前 remaining blocker 仍集中为 attached-session 下 `Runtime.enable / Page.enable / Page.setLifecycleEventsEnabled` 45s timeout；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；下一轮继续同一条 H6 wrapper/CDP 诊断链
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 下一轮优先复制新的 `External preflight self-start CDP timeout comparison command`，确认 `60000ms` 更长 per-command timeout 是否仍停在同一 attached-session 层
  - 若 `60000ms` 仍无差异，则直接把该宿主记录为 attached-session bootstrap blocker，不再继续在这条 wrapper 调参线上空转
- 下一任务前置条件是否满足：满足；下一轮可直接从 `attached_session_bootstrap_blocker_candidate + 60000ms self-start timeout comparison command` 出发做最后一轮 bounded host-level 对照
