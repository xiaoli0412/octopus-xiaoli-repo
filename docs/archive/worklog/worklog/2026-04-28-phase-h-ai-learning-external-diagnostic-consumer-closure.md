# 2026-04-28 Phase H AI Learning External Diagnostic Consumer Closure

## 1. 任务信息

- 任务名称：phase h ai learning external diagnostic consumer closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-external-preflight-full-diagnostic-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 aggregated external preflight 从“只有 JSON 可复用”继续收口成“wrapper 失败时即可直接消费并打印摘要”的闭环，减少下一轮还要手动进入临时目录翻诊断产物
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 验证入口主线，不扩散到 AI 页面业务逻辑、settings consumer、动态路由 API 或新的浏览器基建主题
- 只改 shared smoke wrapper、learning wrapper 守护和状态记录，保持改动集中、可回退、可解释
- 本轮必须形成真实代码增量与直接验证，不能只做阅读和总结

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 与 `web/src/components/modules/setting/*` 的用户可见行为
- 不改动态路由学习数据模型、relay 打分逻辑或 settings 交互
- 不把宿主 loopback / service-provider 阻塞误归类为产品回归

## 5. 本次验收条件

- external 失败时，learning wrapper 会先打印稳定的 `Latest external preflight diagnostic` 摘要，而不是只抛 shared wrapper 原始报错
- 诊断摘要至少包含 `overallClassification`、`failedChecks`、`skippedChecks`、`primaryBlockingCheck`、`summaryLines` 和 artifact 路径
- shared diagnostic JSON 保留结构化字段，供下一轮或健康宿主继续直接复用
- `verify-ai-automation-learning-focus.mjs` 守住新的 wrapper 消费语义

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke-cdp.ps1`
- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改 external preflight 诊断字段与 learning wrapper 失败摘要消费链
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 external 路径失败时的摘要输出与诊断可消费性，`self-start`、页面逻辑和动态路由学习行为保持不变

## 8. 实施步骤

1. 复核 automation memory、canonical plan、用户上下文总账、当前状态文档、前端主线状态、环境 next plan 与上一轮 H6 worklog，确认本轮仍应围绕 learning browser smoke 同一主线推进。
2. 复查 shared `verify-channel-create-browser-smoke-cdp.ps1` 当前已经有 aggregated external preflight JSON，但 learning wrapper 失败时仍主要依赖原始异常消息，确认“诊断产物未被 wrapper 二次消费”是本轮最小且连续的闭环点。
3. 在 shared wrapper summary 中补充 `skippedChecks / primaryBlockingCheck / summaryLines / hints / checkDetails`，让 aggregated preflight 既能用于 JSON，也能直接复用到错误消息和后续 wrapper 摘要输出。
4. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增诊断路径解析与 JSON 读取逻辑：当 external 失败消息里带 `Diagnostic:` 时，先读取 `external-preflight-diagnostic.json`，再打印稳定的 `Latest external preflight diagnostic` 摘要后重新抛错。
5. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新的 wrapper 消费链和 shared summary 字段纳入静态守护。
6. 重新执行 learning smoke `check-only`、静态守护、类型检查、`runtime-win status`、`git diff --check` 与真实 `external` 复跑，确认失败时已先打印聚合摘要再抛错。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`& .\scripts\runtime-win.ps1 -Action status`
- 通过：`git diff --check -- scripts/verify-channel-create-browser-smoke-cdp.ps1 scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-external-diagnostic-consumer-closure.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 已复跑并按预期失败：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`
  - 当前失败会先打印 `Latest external preflight diagnostic`，明确给出 `overallClassification=preflight_failed`、`failedChecks=backend, frontend`、`skippedChecks=cdp`、`primaryBlockingCheck=backend`、summary lines、hints 与 artifact 路径，之后再抛 shared wrapper 的原始失败消息。

## 10. 风险与兼容性

- 新风险：低；仅增强 external smoke 的摘要字段与 wrapper 失败输出
- 兼容性风险：低；`self-start` 路径、AI learning 产品行为和动态路由评分逻辑不受影响
- 是否阻塞下一任务：不阻塞；下一轮可直接基于 wrapper 打印的摘要决定是先补服务可达性，还是在健康宿主上继续追 CDP/browser 证据

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 learning smoke `check-only -UseHostFriendlyExternalDefaults`、`verify-ai-automation-learning-focus.mjs`、`runtime-win status`、`git diff --check`；真实 `external -UseHostFriendlyExternalDefaults` 已复跑并确认失败时先打印聚合摘要再抛错
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、当前状态文档、前端主线状态、详细工作流、环境 next plan、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` 技能文档
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 repo-local wrapper 的 external 失败摘要消费链
- 自动 smoke 阻塞原因 / 缺少的环境：当前 external 仍缺可达的 backend/frontend/CDP 组合；在本宿主上，backend 与 frontend 同时表现为 `host_networking_blocker`
- worklog 是否更新：是
- 遗留项：
  - 在可用宿主或已暴露服务的环境里，继续用 `-Mode external -UseHostFriendlyExternalDefaults` 采集 learning 页真实 browser 证据
  - 若 backend/frontend 已可达但 CDP 仍失败，继续沿 wrapper 打印的摘要与同一份 `external-preflight-diagnostic.json` 聚焦浏览器/CDP bootstrap，而不是回到服务可达性排查
  - `vitest/esbuild spawn EPERM` 与 Windows loopback/service-provider 初始化失败仍维持为环境阻塞，不归类为产品缺陷
- 下一任务前置条件是否满足：满足；下一轮可直接从新的 wrapper 摘要入口继续推进同一条 Phase H6 验证主线
