# 2026-04-28 Phase H AI Learning External Preflight Full Diagnostic Closure

## 1. 任务信息

- 任务名称：phase h ai learning external preflight full diagnostic closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-external-preflight-classification-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 external preflight 从“首个失败即停止”收口成“尽量收集 backend/frontend/CDP 全量诊断后统一报错”的闭环，减少下一轮继续靠人工重跑猜测失败层级
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 验证入口主线，不扩散到 AI 页面业务逻辑、settings consumer、动态路由 API 或新的浏览器基建主题
- 只改共享 smoke wrapper、learning 守护和状态记录，保持改动集中、可回退、可解释
- 本轮必须形成真实代码增量与直接验证，不能只做阅读和总结

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 用户可见行为
- 不改动态路由学习数据模型、relay 打分逻辑或 settings 交互
- 不把宿主 loopback / service-provider 阻塞误归类为产品回归

## 5. 本次验收条件

- external preflight 失败时能同时输出 backend/frontend/CDP 的已知 reachability 结果，而不是在第一处失败就停止
- 诊断 JSON 能明确包含 `schemaVersion`、`failedChecks` 与 `overallClassification`
- 非必需的 CDP 预检项会被标记为 `skipped`，而不是混入 `unreachable`
- `verify-ai-automation-learning-focus.mjs` 守住新的共享 wrapper 语义

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke-cdp.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改共享 smoke wrapper 的 external preflight 聚合诊断与守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 external 路径失败时的诊断深度与文案表达，`self-start` 和既有 AI learning 页面行为不变

## 8. 实施步骤

1. 复核 automation memory、主计划、当前状态文档、AI 自动化/动态路由实施计划与上一轮 H6 worklog，确认本轮仍应围绕 learning browser smoke 同一主线推进。
2. 复查共享 `verify-channel-create-browser-smoke-cdp.ps1` 当前 external preflight 逻辑，确认它仍按 backend -> frontend -> CDP 的顺序在第一处失败时直接抛错。
3. 在共享 wrapper 中新增聚合 helper，统一生成 skipped entry、diagnostic path、summary lines、failed checks、hint 集合和统一 failure message。
4. 把 external preflight 的输出升级为 `schemaVersion = 2`，写入 `failedChecks / overallClassification`，并让非必需 CDP 项显式标记为 `skipped`。
5. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新的共享 wrapper 语义纳入静态守护。
6. 重新执行 learning smoke `check-only`、静态守护、类型检查、`git diff --check` 与真实 `external` 复跑，确认失败摘要已从“首个失败点”收口成聚合诊断。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-channel-create-browser-smoke-cdp.ps1 scripts/verify-ai-automation-learning-focus.mjs`
- 已复跑并按预期失败：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`
  - 当前失败已不再只报首个 backend 阻塞点，而是一次性报告 `backend + frontend` 的 `host_networking_blocker`，并把未要求的 CDP 项标记成 `skipped`；诊断产物仍统一写入 `external-preflight-diagnostic.json`。

## 10. 风险与兼容性

- 新风险：低；仅增强 external smoke 的预检、提示和诊断产物
- 兼容性风险：低；`self-start` 路径和 AI learning 产品行为完全不受影响
- 是否阻塞下一任务：不阻塞；下一轮可以直接基于新的聚合诊断继续推进 external/browser 证据，或在可用宿主上验证 backend/frontend 通后是否只剩 CDP 问题

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 learning smoke `check-only -UseHostFriendlyExternalDefaults`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`；真实 `external -UseHostFriendlyExternalDefaults` 已复跑并确认失败摘要收口成聚合诊断
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` 技能文档
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 repo-local wrapper 与 external preflight 聚合诊断
- 自动 smoke 阻塞原因 / 缺少的环境：当前 external 仍缺可达的 backend/frontend/CDP 组合；在本宿主上，backend 与 frontend 同时表现为 `host_networking_blocker`
- worklog 是否更新：是
- 遗留项：
  - 在可用宿主或已暴露服务的环境里，继续用 `-Mode external -UseHostFriendlyExternalDefaults` 采集 learning 页真实 browser 证据
  - 若 backend/frontend 已可达但 CDP 仍失败，继续沿同一份 `external-preflight-diagnostic.json` 聚焦浏览器/CDP bootstrap，而不是回到服务可达性排查
  - `vitest/esbuild spawn EPERM` 与 Windows loopback/service-provider 初始化失败仍维持为环境阻塞，不归类为产品缺陷
- 下一任务前置条件是否满足：满足；下一轮可直接从新的聚合 preflight 诊断入口继续推进同一条 Phase H6 验证主线
