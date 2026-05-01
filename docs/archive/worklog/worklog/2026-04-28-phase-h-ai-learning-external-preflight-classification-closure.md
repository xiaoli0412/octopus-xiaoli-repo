# 2026-04-28 Phase H AI Learning External Preflight Classification Closure

## 1. 任务信息

- 任务名称：phase h ai learning external preflight classification closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-host-friendly-external-bootstrap-alignment.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 `external + cdp` 路径里原本粗粒度的 backend/frontend/CDP 等待与失败提示收口成结构化 external preflight 分类，减少下一轮继续靠人工猜测失败层级
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.mjs`、`scripts/verify-ai-automation-learning-focus.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 验证入口主线，不扩散到 AI 页面业务逻辑、设置页 consumer、动态路由 API 或新的浏览器基建主题
- 只改共享 smoke wrapper、守护与状态文档，保持改动集中、可回退、可解释
- 本轮必须形成真实代码增量与直接验证，不能只做阅读和总结

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 用户可见行为
- 不改动态路由学习数据模型、relay 打分逻辑或 settings 交互
- 不把宿主 loopback / service-provider 阻塞误归类为产品回归

## 5. 本次验收条件

- 共享 `verify-channel-create-browser-smoke-cdp.ps1` 在 external 路径下能把 backend / frontend / CDP 依赖预检拆成结构化 reachability 结果
- external 失败时输出具体的 preflight 分类与 diagnostic JSON，而不是只给粗粒度 `Wait-Http` 超时
- `verify-ai-automation-learning-focus.mjs` 守住新的共享 wrapper 语义
- `verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 在当前宿主上失败点前移为 external preflight 分类，且错误信息可直接指出是 backend/frontend/CDP 哪一层不可达

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke-cdp.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改共享 smoke wrapper 的 external preflight 与守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；`external` 路径失败时的报错更具体，`self-start` 与已有 AI learning 页面行为不变

## 8. 实施步骤

1. 复核 automation memory、主计划、当前状态文档、AI 自动化/动态路由实施计划与上一轮 H6 worklog，确认本轮仍应围绕 learning browser smoke 同一主线推进。
2. 复查共享 `verify-channel-create-browser-smoke-cdp.ps1` 与 learning wrapper 当前 external 流程，确认 backend/frontend/CDP 仍分别走通用等待与粗粒度异常。
3. 在共享 wrapper 新增 external 依赖预检 helper，统一输出 backend/frontend/CDP reachability 分类、细化提示语并写出 `external-preflight-diagnostic.json`。
4. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新的共享 wrapper 语义纳入静态守护。
5. 重新执行 learning smoke `check-only`、静态守护、真实 `external` 复跑、类型检查与 `git diff --check`。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-channel-create-browser-smoke-cdp.ps1 scripts/verify-ai-automation-learning-focus.mjs`
- 已复跑并按预期失败：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`
  - 当前失败已不再是粗粒度 `Timed out waiting for ...`，而是先在 external preflight 把 `backend healthcheck` 分类为 `host_networking_blocker`，并输出 `external-preflight-diagnostic.json` 供下一轮或健康宿主直接复用。

## 10. 风险与兼容性

- 新风险：低；只增强 external smoke 的预检、提示和诊断产物
- 兼容性风险：低；`self-start` 既有路径和 AI learning 产品行为完全不受影响
- 是否阻塞下一任务：不阻塞；下一轮可以直接沿着新的 diagnostic JSON 继续推进 external/browser 证据，或在可用宿主上确认 frontend/backend/CDP 外部组合

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 learning smoke `check-only -UseHostFriendlyExternalDefaults`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`；真实 `external -UseHostFriendlyExternalDefaults` 已复跑并确认失败点收口到 external preflight 分类
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、用户上下文总账、环境 next plan、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` 技能文档
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 repo-local wrapper 与 external preflight 诊断
- 自动 smoke 阻塞原因 / 缺少的环境：当前 external 仍缺可达的 backend/frontend/CDP 组合；在本宿主上，失败首先表现为 backend healthcheck 的 `host_networking_blocker`
- worklog 是否更新：是
- 遗留项：
  - 在可用宿主或已暴露服务的环境里，继续用 `-Mode external -UseHostFriendlyExternalDefaults` 采集 learning 页真实 browser 证据
  - 若 external frontend/backend 已可达但 CDP 仍失败，可直接基于新增的 `external-preflight-diagnostic.json` 区分服务不可达与浏览器不可达
  - `vitest/esbuild spawn EPERM` 与 Windows loopback service-provider 初始化失败仍维持为环境阻塞，不归类为产品缺陷
- 下一任务前置条件是否满足：满足；下一轮可直接从新的 external preflight 诊断入口继续推进同一条 Phase H6 验证主线
