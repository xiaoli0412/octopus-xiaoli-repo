# 2026-04-29 Phase H AI learning require-cdp bridge artifact closure

## 1. 任务信息

- 任务名称：Phase H6 learning require-cdp bridge artifact closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-action-class-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable replay 主线，解决“真实 `external + self-start + require-cdp` 虽然已经把 backend/frontend/CDP version 拉通，但一旦卡在 CDP page bootstrap timeout，就不会生成新的 repo-local `requireCdp=true` stable artifact”这一最后主阻塞。
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/runtime-win.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、共享 `scripts/verify-channel-create-browser-smoke-cdp.ps1`、repo-local stable diagnostic 目录、两次真实 external temp artifact 目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 目录、共享 CDP wrapper、真实 external 失败 artifact 目录
- 若未使用部分本地资源或上下文，原因：本轮不需要重开 AI 页面 UI、settings 页面、后端 schema 或其它主线；真实 external 失败 artifact 与现有 H6 脚本足以闭环当前问题
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / 状态文档主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 本轮必须形成真实代码增量与直接验证，不能只更新文档或复述阻塞
- 只增强 learning smoke wrapper 在 CDP smoke 失败路径上的诊断桥接与 stable artifact 落盘，不改变产品运行态逻辑或共享 CDP wrapper 的主控制流

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到新的 UI 返工、后端 schema 变更或 unrelated host repair
- 不伪造静态 repo-local artifact；requirement-specific 变体必须来自真实 external 失败或无副作用桥接产物

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1` 在共享 CDP wrapper 抛出 `CDP diagnostic file:` 但没有新的 external preflight JSON 可直接发布时，能桥接生成可解析的 `external-preflight-diagnostic.cdp-bridge.json`
- 真实 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight -SelfStartServices` 失败后，`build/verify-ai-automation-learning/latest-external-preflight-diagnostic-require-cdp.json` 会真实落盘
- sequential `check-only -RequireExternalCdpPreflight` 会直接选中新生成的 `requireCdp=true` stable artifact，并显示 `coverage complete`
- sequential 默认 `check-only` 也能看到 parseable `requireCdp=true / requireCdp=false` 两类 stable variant 共存
- `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit`、`git diff --check` 与 `runtime-win.ps1 -Action status` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-require-cdp-bridge-artifact-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI；只改 wrapper 诊断桥接、stable artifact 落盘与静态守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅让真实 `cdp_smoke_failed` 路径也能沉淀到 repo-local stable replay，且 `check-only` 可直接继承

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog 与 `runtime-win.ps1 -Action status`，确认本轮仍应停留在 learning smoke stable replay 诊断链。
2. 运行真实 `external + self-start + require-cdp`，确认当前宿主的真实阻塞已从 `host_networking_blocker` 进入 `page_bootstrap_timeout_attached_session`，但 repo-local 仍没有新的 `requireCdp=true` stable artifact。
3. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中补 `CDP diagnostic file:` 失败路径的桥接逻辑：新增从异常消息提取 `cdp.diagnostic.json` 路径、从消息文本提取 CDP 摘要、生成 `external-preflight-diagnostic.cdp-bridge.json`、以及按当前 invocation URL 组装 bridge diagnostic 的 helper。
4. 在 `scripts/verify-ai-automation-learning-focus.mjs` 中补对应静态守护，确保 bridge artifact、CDP bridge summary line、CDP-smoke-failed classification 与 bridge helper wiring 不会无声回退。
5. 顺序运行语法检查、静态守护、前端 `tsc --noEmit`、真实 `external + self-start + require-cdp`、两次 sequential `check-only` 与 `runtime-win.ps1 -Action status`，确认新 bridge 路径不仅能落盘 `latest-external-preflight-diagnostic-require-cdp.json`，还能让 repo-local coverage 进入 complete 状态。
6. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，把下一轮入口从“去健康宿主补 `requireCdp=true` artifact”切换成“继续刷新 optional profile 到同一运行层，或直接聚焦当前 attached-session CDP bootstrap timeout”。

## 9. 测试与验证

- 通过：PowerShell 语法解析 `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs`
- 按预期失败但成功落盘新 artifact：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight -SelfStartServices`
- 通过：`Get-ChildItem build\verify-ai-automation-learning`，确认 `latest-external-preflight-diagnostic-require-cdp.json` 与 `latest-external-preflight-diagnostic.json` 于本轮刷新
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`

## 10. 风险与兼容性

- 新风险：低；bridge diagnostic 只在 AI learning wrapper 自身的 catch 路径生成，不会影响共享 CDP wrapper 的主流程
- 兼容性风险：低；旧的 `external-preflight-diagnostic.json` 与 stable replay schema 继续沿用，新增 bridge artifact 仍落在同一 schemaVersion 2 结构上
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（语法解析、静态守护、真实 external self-start require-cdp 复跑按预期失败但成功落盘新 artifact、两次 sequential `check-only`、`runtime-win status`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 目录、共享 CDP wrapper、真实 external temp artifact 目录
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最新 Phase H6 worklog 说明上一轮仍把主阻塞定义为“缺少 `requireCdp=true` artifact”；真实 external temp artifact 目录证明当前宿主已经越过 service reachability，真正卡在 `attached-session` 的 `Runtime.enable / Page.enable`；repo-local stable diagnostic 目录则证明只要 bridge 路径补齐，本机就能真实产出 requirement-specific stable replay，而不需要再等待健康宿主补证
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：真实 external self-start require-cdp 已执行；服务与 CDP version reachability 成功，当前失败稳定收敛为 `page_bootstrap_timeout_attached_session`
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 Edge/CDP attached-session 页面 bootstrap 仍会在 `Runtime.enable / Page.enable` 超时；真实 jsdom/vitest 入口仍受 `vite/esbuild spawn EPERM` 影响
- 待验证页面清单：无新增产品页；同主线下下一步更适合继续验证 optional profile 是否也要刷新到同一运行层，或直接继续收口 attached-session CDP bootstrap 超时的策略证据
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - repo-local stable coverage 现已 complete，但默认 optional profile 仍回放历史 `preflight_failed` artifact；如果后续要让两个 invocation profile 都共享同一“服务可达、只剩 CDP bootstrap timeout”的运行层，需要再决定是否顺手刷新 optional external profile
  - 当前 require-cdp invocation 的真实 remaining blocker 已稳定收敛为 `attached-session + runtime-page-lifecycle` 下的 `Runtime.enable / Page.enable` timeout，下一轮应优先继续同主线下的 CDP bootstrap 策略判断，而不是回头补 artifact
- 下一任务前置条件是否满足：满足；下一轮优先继续同一条 Phase H6 诊断链，要么刷新 optional profile 到同一运行层，要么直接基于现有 require-cdp replay 继续缩小 attached-session bootstrap timeout 的可比较参数面
