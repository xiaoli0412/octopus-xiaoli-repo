# 2026-04-28 Phase H AI Learning Host-Friendly External Bootstrap Alignment

## 1. 任务信息

- 任务名称：phase h ai learning host-friendly external bootstrap alignment
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-external-cdp-friendly-defaults.md`
- 本次任务目标：继续停留在同一条 Phase H6 浏览器验证主线，把 `AI 自动化` learning smoke 的 host-friendly `external + cdp` 入口从“默认还会拉本地服务，导致 loopback 宿主提前失败”收口成“默认优先 external CDP，local service bootstrap 仅在显式开关下启用”的更稳语义
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、`scripts/runtime-win.ps1`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 验证入口主线，不扩散到 AI 页面业务逻辑、动态路由 API 或新的浏览器基建分支
- 只改 wrapper、守护和状态文档，保持风险集中且可回退
- 必须围绕当前主线形成真实增量，不能只停留在重复扫描和分析

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 用户可见行为
- 不改动态路由学习数据结构与设置页交互
- 不把宿主 loopback / service-provider 阻塞误写成产品回归

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1` 的 `-UseHostFriendlyExternalDefaults` 不再默认打开 local service bootstrap
- 仍保留显式 `-SelfStartServices` 以及兼容别名 `-SelfStartLocalServices`，便于后续在健康宿主上复用旧路径
- `verify-ai-automation-learning-focus.mjs` 守住新的 wrapper 语义
- 真实 `external + cdp` 复跑不再先死于本地 loopback 端口探测，而是进入更高层的 external 入口失败分类

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改验证入口语义和文档
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；默认 strict external/self-start 语义仍保留，只有 host-friendly external 快捷入口的 local bootstrap 默认值被收紧

## 8. 实施步骤

1. 复核 automation memory、主计划、当前阶段文档与上一轮 H6 worklog，确认本轮仍应围绕 learning browser smoke 同一主线推进。
2. 直接复跑 `AI 自动化` learning 的真实 `external + cdp` host-friendly 命令，捕获当前失败点。
3. 根据复跑结果，只修 `verify-ai-automation-learning-browser-smoke.ps1` 的 host-friendly external 默认语义，并同步更新静态守护与状态文档。
4. 重新执行 `check-only`、真实 `external`、守护脚本、`tsc --noEmit` 与 `git diff --check`。

## 9. 测试与验证

- 通过：`& .\scripts\runtime-win.ps1 -Action status`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 已复跑并按预期失败：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`
  - 本轮真实 external 失败点已从“本地 loopback TCP probing on 127.0.0.1:18081”前移为 external frontend/backend reachability 阶段，不再因为 host-friendly 快捷入口默认拉起 local service bootstrap 而提前死在宿主 loopback 端口探测。
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-28-phase-h-ai-learning-host-friendly-external-bootstrap-alignment.md`

## 10. 风险与兼容性

- 新风险：低；只改变 host-friendly external 快捷入口的默认 local bootstrap 语义
- 兼容性风险：低；需要旧行为时仍可显式加 `-SelfStartServices`，同时新增兼容别名 `-SelfStartLocalServices`
- 是否阻塞下一任务：不阻塞；下一轮可以继续围绕同一条 H6 线补 external/browser 证据，或在有健康宿主时再切回带本地服务自启动的对照路径

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 `runtime-win status`、learning smoke `check-only -UseHostFriendlyExternalDefaults`、`verify-ai-automation-learning-focus.mjs`、`git diff --check`；真实 `external -UseHostFriendlyExternalDefaults` 已复跑并确认失败点上移到 backend/frontend reachability 阶段
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、详细工作流、用户上下文总账、前端主线状态、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` 技能文档
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 repo-local wrapper 与自动 smoke 入口
- 自动 smoke 阻塞原因 / 缺少的环境：当前 external 仍缺真实可达的 backend/frontend/CDP 组合，但已不再因为 host-friendly 快捷入口默认打开 local service bootstrap 而提前撞上 loopback service-provider 阻塞
- worklog 是否更新：是
- 遗留项：
  - 在可用宿主或已启动服务的条件下，继续用 `-Mode external -UseHostFriendlyExternalDefaults` 采集 learning 页真实 browser 证据
  - 需要 local service bootstrap 对照时，再显式加 `-SelfStartServices` 或 `-SelfStartLocalServices`
  - `vitest/esbuild spawn EPERM` 与 Windows loopback service-provider 初始化失败仍维持为环境阻塞，不归类为产品缺陷
- 下一任务前置条件是否满足：满足；下一轮可直接从新的 host-friendly external 入口继续推进同一条 Phase H6 验证主线
