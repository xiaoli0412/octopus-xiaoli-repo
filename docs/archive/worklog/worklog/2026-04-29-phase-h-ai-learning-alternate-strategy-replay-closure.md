# 2026-04-29 Phase H AI learning alternate-strategy replay closure

## 1. 任务信息

- 任务名称：Phase H6 learning alternate-strategy replay closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-alternate-execution-path-command-closure.md`
- 本次任务目标：在不再继续 live 调参的前提下，把 `attached-session` invocation 对已存在 `json-new` repo-local 诊断的消费语义收口到 strategy-aware 状态，避免 `check-only` 继续把下一轮错误引向“先 refresh matching copy”
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/worklog/2026-04-29-phase-h-ai-learning-attached-session-blocker-confirmed-closure.md`、`docs/worklog/2026-04-29-phase-h-ai-learning-alternate-execution-path-command-closure.md`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、`build/verify-ai-automation-learning/` 下最新 stable JSON
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON、`brainstorming` skill（仅用于收敛本轮 replay 语义边界）
- 若未使用部分本地资源或上下文，原因：本轮不涉及产品 UI、后端 schema、真实浏览器页交互或宿主网络修复；工作继续限定在 wrapper replay 解释层
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / repo-local stable replay / host-level CDP 诊断主线，不扩散到产品 UI、动态路由业务逻辑或宿主修复
- 只修正 replay 解释层与交接命令预算，不改 stable artifact 命名、不伪造新的 repo-local JSON
- 必须先复核当前 stable JSON 的真实 page strategy，再决定 summary/action 是否要从 refresh 导向切回 alternate-path 导向

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不新增新的 live external rerun 作为主任务，不再继续在当前宿主上放大 timeout/order/page-mode 调参
- 不改 shared CDP smoke wrapper schema，不改后端接口、数据库或模型逻辑

## 5. 本次验收条件

- `check-only` 在 `attached-session` invocation 下能够明确标出 `matching-generic / alternate-generic` 是来自 alternate page-bootstrap strategy `json-new`
- `decision summary / recommended action / recommended action class` 不再误导去优先 refresh `attached-session` matching copy，而是先建议复用现有 `json-new` 证据或直接执行 alternate execution path command
- `alternate execution path command` 继承现有 repo-local `json-new` 证据中的 `60000ms / 200s` 预算，不回落到默认 `30000ms / 110s`
- PowerShell parser 校验、两次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-alternate-strategy-replay-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：只改 wrapper replay 语义与交接命令，不改产品 UI
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser verifier `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只影响 `check-only` 文本分类、动作建议与 alternate-path 命令预算

## 8. 实施步骤

1. 复核 repo-local stable JSON 与当前 `check-only` 输出，确认 `build/verify-ai-automation-learning/` 已有 `json-new` optional / require-cdp 变体，而 `attached-session` invocation 仍被误导为优先 refresh matching copy。
2. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中新增 strategy-aware helper，显式识别 `matching-generic / alternate-generic` 所记录的 alternate page-bootstrap strategy，并把 coverage note、decision summary、recommended action class、recommended action 与 status line 调整为“先消费现有 alternate-strategy 证据，再决定是否 refresh strategy-specific copy”。
3. 修正 `Get-ExternalPreflightEffectiveAlternateExecutionPathCommand`，让 alternate-path 命令在预览已保存 `json-new` 诊断时继承其 `60000ms` 命令超时与换算后的 `200s` Node 预算，而不是回落到默认 host-friendly 基线。
4. 更新 `scripts/verify-ai-automation-learning-focus.mjs` 静态守护，补上 alternate strategy 文案和新的 fallback-ready 语义约束。
5. 同步更新当前状态文档、前端主线状态、automation memory 与本 worklog。

## 9. 测试与验证

- 通过：`powershell -NoProfile -Command '$tokens=$null; $errors=$null; [void][System.Management.Automation.Language.Parser]::ParseFile("D:\GPT-codex\octopus_repo\scripts\verify-ai-automation-learning-browser-smoke.ps1",[ref]$tokens,[ref]$errors); if ($errors -and $errors.Count -gt 0) { $errors | ForEach-Object { $_.ToString() }; exit 1 }'`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-29-phase-h-ai-learning-alternate-strategy-replay-closure.md`

## 10. 风险与兼容性

- 新风险：低；仅调整 `check-only` 摘要、动作分类和命令预算
- 兼容性风险：低；不影响 live external、本地服务、自定义 artifact 命名或前后端接口
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（PowerShell parser、两次 `check-only`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON、`brainstorming` skill
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：最新 stable JSON 说明 `json-new` 双 profile 证据已经存在，但 `attached-session` invocation 的 replay 摘要仍按“缺 matching copy”导向 refresh；因此最值得推进的是修 replay 解释层和 alternate-path 预算，而不是继续 live 调参
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未新增 live rerun；本轮以 repo-local `check-only` 解释层收口为主
- 手工 smoke 阻塞原因 / 缺少的环境：真实宿主级 `attached-session / json-new` page bootstrap timeout 结论已足够明确，本轮没有继续在同宿主上扩大 live 对照
- 待验证页面清单：无新增产品页；下一轮可直接复制 `json-new` alternate execution 命令到健康宿主复跑，或在当前宿主明确止损并切换相邻主线
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 若下一轮仍留在 Phase H6，优先在不同宿主使用现成 `json-new` alternate execution 命令取证，而不是在当前宿主继续调参
  - 若当前宿主不再有价值，可把这条主线正式挂账为“attached-session + json-new 双 strategy 均已在本机确认 page bootstrap blocker”，再切换同主线相邻任务
- 下一任务前置条件是否满足：满足；当前 replay 解释层和交接命令都已收口，下一轮可直接执行 alternate-path live 命令或切宿主
