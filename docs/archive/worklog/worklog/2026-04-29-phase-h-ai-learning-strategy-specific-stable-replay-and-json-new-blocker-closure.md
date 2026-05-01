# 2026-04-29 Phase H AI learning strategy-specific stable replay and json-new blocker closure

## 1. 任务信息

- 任务名称：Phase H6 learning strategy-specific stable replay and json-new blocker closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-alternate-execution-path-command-closure.md`
- 本次任务目标：把 learning smoke stable replay 从“只按 requireCdp 选稳定副本”收口到“同时区分 page-bootstrap strategy”，并在 `json-new` 路径拿到双 profile live 证据后，把宿主结论升级为 `page_bootstrap_strategy_blocker_confirmed`
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、上一轮 H6 worklog、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、共享 CDP wrapper、repo-local stable diagnostic 目录与 optional/require-cdp JSON 副本
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON、共享 CDP wrapper
- 若未使用部分本地资源或上下文，原因：本轮不涉及产品 UI、settings 页面、后端 schema 或新的业务接口；工作继续收敛在 learning smoke wrapper、stable replay 诊断和 host-level blocker 分类
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / host-level CDP bootstrap 诊断主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 优先收敛 verifier 契约错误：同一 `requireCdp` 下不同 `page-bootstrap strategy` 的 stable artifact 不允许再互相冒充“matching requirement-specific”证据
- 只调整 no-browser verifier、stable replay 命名/选择逻辑与诊断摘要，不修改用户可见产品行为

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到新的 CDP wrapper 架构改造、后端接口变更、Edge 启动策略重做或动态路由业务逻辑
- 不伪造 stable artifact；新 strategy-specific 副本必须来自真实 external rerun

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 会按 `requireCdp + page-bootstrap strategy` 识别 stable artifact，不再把 `json-new` 副本误判成 `attached-session` invocation 的 matching requirement-specific evidence
- `check-only` 在 strategy-specific 副本缺失时，会明确输出 `same-expectation fallback copy` 与 strategy-aware coverage note，而不是继续宣称“matched this invocation”
- optional `json-new` live rerun 后，repo-local 会新增 `latest-external-preflight-diagnostic-optional-cdp-json-new.json`
- `json-new` 双 profile 证据齐备后，`check-only -CdpPageBootstrapStrategy json-new` 会输出 `page_bootstrap_strategy_blocker_confirmed`
- `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-strategy-specific-stable-replay-and-json-new-blocker-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI；先改 stable replay / verifier 语义，再改 no-browser 守护，再补文档
- 受影响后端模块：无
- 受影响前端模块：无产品前端模块；仅 no-browser verifier `scripts/verify-ai-automation-learning-focus.mjs`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅收紧 learning smoke 的 stable artifact 命名、选择和摘要解释

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态与上一轮 H6 worklog，确认本轮最值得推进的是收敛 stable replay 契约错误，而不是继续盲调 attached-session timeout。
2. 运行 `runtime-win.ps1 status`，确认项目保持默认停驻状态。
3. 先跑 default / require-cdp 的 `check-only` 与两条 `json-new + self-start + 60000ms` live external rerun，确认 optional / require-cdp 两条 profile 在 `json-new` 下都稳定停在 `page_bootstrap_timeout_json_new`。
4. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中引入 strategy-aware stable artifact 命名与选择逻辑：新增 `Normalize-ExternalPreflightPageBootstrapStrategy`、strategy-specific variant path、generic fallback path、`matching-generic / alternate-generic` 状态项，以及 strategy-aware coverage / selection note / fallback action。
5. 同步在 wrapper 中新增 `Get-ExternalPreflightPageBootstrapTimeoutState` 与 `Get-StableExternalPreflightPageBootstrapBlockerConfirmedState`，让 `json-new` 双 profile 证据齐备后能直接升级为 `page_bootstrap_strategy_blocker_confirmed`。
6. 更新 `scripts/verify-ai-automation-learning-focus.mjs`，把静态守护从“只按 requireCdp 选 stable copy”的旧契约改成“variant 模板 + strategy-specific/generic fallback + page-bootstrap blocker confirmed”的新契约。
7. 收工前重新执行 `runtime-win.ps1 stop/status`、PowerShell parser、四次 `check-only`（default/require-cdp + attached-session/json-new 组合）、一条 optional `json-new` live rerun、`verify-ai-automation-learning-focus.mjs`、`tsc --noEmit` 与 `git diff --check`，最后同步更新状态文档、前端主线状态和 automation memory。

## 9. 测试与验证

- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop`
- 通过：PowerShell parser 校验 `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- 通过：`$env:APPDATA=...; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA=...; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA=...; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -CdpPageBootstrapStrategy 'json-new'`
- 通过：`$env:APPDATA=...; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight -CdpPageBootstrapStrategy 'json-new'`
- 观察到预期 `page_bootstrap_timeout_json_new / page_bootstrap_strategy_blocker_confirmed`：`$env:APPDATA=...; powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -NodeSmokeTimeoutSeconds 200 -CdpCommandTimeoutMs 60000 -CdpPageBootstrapStrategy 'json-new' -SelfStartServices`
- 通过：`$env:APPDATA=...; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA=...; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs`

## 10. 风险与兼容性

- 新风险：低；本轮仅收紧 learning smoke verifier 与 stable replay 诊断口径
- 兼容性风险：低；旧 generic stable copy 仍保留，新增的是 strategy-specific 副本与更明确的 fallback 语义
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（`runtime-win stop/status`、PowerShell parser、四次 `check-only`、一条 optional `json-new` live rerun、`verify-ai-automation-learning-focus.mjs`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、上一轮 H6 worklog、repo-local stable diagnostic JSON、共享 CDP wrapper
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：上一轮 H6 worklog 说明 attached-session blocker 已确认，因而本轮应优先检验 `json-new` 是否提供不同失败层；真实 `json-new` live rerun 证明 optional/require-cdp 双 profile 在 `json-new` 下仍共停于 `page_bootstrap_timeout_json_new`；repo-local stable diagnostics 暴露了“只按 requireCdp 选副本”会误把 `json-new` 当成 attached-session matching evidence，因此必须先修 verifier 契约
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：本轮真实 external rerun 已按计划执行，验证后已执行 `runtime-win.ps1 stop`，本机回到默认停驻状态
- 手工 smoke 阻塞原因 / 缺少的环境：当前 remaining blocker 已从 attached-session lane 扩展收口为“本机 `json-new` strategy 也会稳定停在 `page_bootstrap_timeout_json_new`”；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；下一轮若继续 H6，应在不同宿主上复用 `json-new` strategy-aware command，而不是继续在本机调超时/顺序
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - `attached-session` invocation 目前仍只有 generic fallback，尚无 `*-attached-session.json` strategy-specific stable copy；这不是脚本错误，而是当前本机尚未重新跑 attached-session strategy-specific live rerun
  - 下一轮如果继续 H6，应优先换宿主或换真正不同的执行路径，而不是继续在本机围绕 `json-new` 或 `attached-session` 加大超时
- 下一任务前置条件是否满足：满足；strategy-aware stable replay 已闭环，`json-new` page-bootstrap blocker 也已确认，下一轮可直接转到不同宿主或同主线相邻任务
