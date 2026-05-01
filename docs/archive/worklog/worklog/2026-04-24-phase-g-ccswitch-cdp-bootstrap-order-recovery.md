# 2026-04-24 Phase G CC Switch CDP Bootstrap Order Recovery

## 1. 任务信息

- 任务名称：`CC Switch` 浏览器级主证据恢复与 CDP bootstrap 顺序收口
- 日期：`2026-04-24`
- 当前阶段：`Phase G` screenshot-first UI closure
- 对应 milestone：`P1` 图片问题池 / `CC Switch` browser evidence closure

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`11.4.2`、`11.4.3` 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-24-phase-g-ccswitch-browser-evidence-selector-and-host-blocker.md`
  - `docs/worklog/2026-04-23-phase-g-ccswitch-375px-and-verification-sync.md`
- 本次任务目标：
  - 重新验证 `CC Switch` 的共享 CDP browser smoke，确认上一轮记录的 blocker 是否仍然成立。
  - 若当前宿主存在可通过的 bootstrap 顺序，则把默认 wrapper 固定到该顺序，避免下一轮继续误判为 host blocker。
  - 在同一条 CDP smoke 中补上 `CC Switch` 帮助提示的真实 `hover` 断言，让本轮不仅是“能打开页面”，而是收紧到 `focus + hover + 375px`。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `scripts/verify-ccswitch-browser-smoke.ps1`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-ccswitch-flow.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、当前状态、详细工作流、环境计划、前端主线状态、用户总账、上一轮 `CC Switch` worklog、automation memory、当前 smoke 脚本源码
- 若未使用部分本地资源或上下文，原因：本轮只收口 `CC Switch` 同池 browser smoke，不扩散到备份、动态路由或后端业务逻辑
- 本次是否启用子 agent 与分工边界：`否`
- 本次子 agent 使用模型：`N/A`
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：`否`
- 若未使用子 agent，原因：用户已要求主线程串行推进，且本轮任务集中在同一套 smoke 脚本与状态文档

## 3. 本次硬规则

- 只处理 `CC Switch` browser smoke、共享 CDP smoke 辅助逻辑和同池状态文档，不改深链协议与后端契约。
- 浏览器级证据只有在命令真实通过后才能记为闭环；宿主波动必须明确记录，不得用“有脚本”代替“已验证”。
- 本轮必须保持默认停驻策略，验证结束后不留下 `octopus_repo` 常驻进程。

## 4. 本次禁止事项

- 不回滚仓库中无关脏改动。
- 不扩散到 `channel create`、`group create`、homepage、settings 或其他 screenshot-first 主题。
- 不把 CLI `spawn EPERM` 直接写成 `CC Switch` 功能缺陷。

## 5. 本次验收条件

- `scripts/verify-ccswitch-browser-smoke.ps1 -Mode self-start` 在当前宿主至少有一条默认路径可通过。
- `scripts/verify-channel-create-browser-smoke-cdp.mjs` 中 `ccswitch` 场景补上真实 `hover` 断言。
- `node scripts/verify-ccswitch-flow.mjs` 与 `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 保持通过。

## 6. 本次回滚点

- `scripts/verify-ccswitch-browser-smoke.ps1`
- `scripts/verify-channel-create-browser-smoke-cdp.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先跑现有 smoke 复现实情，再收口 wrapper 默认参数与 smoke 断言
- 受影响后端模块：`无`
- 受影响前端模块：`无业务组件变更；仅 browser smoke 对 CC Switch 的交互验收更完整`
- 受影响脚本：`verify-ccswitch-browser-smoke.ps1`、`verify-channel-create-browser-smoke-cdp.mjs`
- 是否影响旧数据：`否`
- 是否影响旧行为：`否`；仅影响浏览器 smoke 默认 bootstrap 顺序和断言覆盖范围

## 8. 实施步骤

1. 复核主规划、用户总账、当前状态、前端主线状态、详细工作流、automation memory 与最近 `CC Switch` worklog，确认本轮继续留在 `Phase G` screenshot-first / `CC Switch` browser evidence 主线。
2. 先执行 `runtime-win.ps1 -Action status`、`verify-ccswitch-flow.mjs`、`tsc --noEmit`，确认当前工作区停驻且 no-browser 基线保持绿色。
3. 运行 `verify-ccswitch-browser-smoke.ps1 -Mode self-start`，复核上一轮记录的宿主 blocker；再用 `runtime-page-lifecycle`、更长 timeout 和不同 preset 做最小对照。
4. 发现 `runtime-page-lifecycle` 在当前宿主可通过后，把 `verify-ccswitch-browser-smoke.ps1` 的默认 `CdpBootstrapCommandOrder` 固定到该顺序，避免下一轮继续把默认入口跑到失败顺序上。
5. 在共享 CDP smoke 脚本中补 `hoverSelectorViaCdp`、`resetHelpHintInteractionViaCdp` 与 `hoverScopedHelpHintAndReadTooltip`，把 `CC Switch` 场景从只验 `focus` 收紧到 `focus + hover`。
6. 重新运行 `verify-ccswitch-flow.mjs`、`tsc --noEmit` 和 `verify-ccswitch-browser-smoke.ps1 -Mode self-start`，确认修改后的默认入口可通过。
7. 同步更新前端主线状态、当前状态、环境 next plan 和 automation memory，写清“主证据已恢复”和“仍保留的宿主波动项”。

## 9. 测试与验证

- 已通过：
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `node scripts/verify-ccswitch-flow.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120 -CdpBootstrapCommandOrder runtime-page-lifecycle`
- 已执行但失败：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 60`
    - 失败点：默认旧顺序 `page-lifecycle-runtime` 下，`attached-session` bootstrap 卡在 `Page.enable / Page.setLifecycleEventsEnabled / Runtime.enable`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120 -CdpCommandTimeoutMs 30000`
    - 失败点：个别复跑中 backend 端口仍可能被前一轮残留占用，说明同进程快速连跑存在宿主级端口回收波动
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
    - 失败点：个别复跑中 Edge 远程调试端口未在限定时间内完成启动，属于宿主级启动波动，不影响已取得的一次默认路径通过结论
- 专项验证结论：
  - `CC Switch` 的 CDP browser smoke 在当前宿主并非完全阻塞，关键差异是 bootstrap 命令顺序；把默认顺序切到 `runtime-page-lifecycle` 后，默认 `self-start + cdp` 路径可通过。
  - 本轮新增的 `hover` 断言基于 `Input.dispatchMouseEvent(mouseMoved)`，因此 `CC Switch` 帮助提示现在不再只验键盘 `focus`，而是和 settings CLI smoke 一样收紧到真实交互层。

## 10. 风险与兼容性

- 新风险：低；本轮只调整 smoke 默认参数与交互断言，不改业务实现。
- 兼容性风险：中低；`runtime-page-lifecycle` 是当前宿主已验证通过的默认顺序，但 Edge 自启动和端口准备仍有偶发波动，因此保留 `CdpBootstrapCommandOrder` 参数用于宿主诊断。
- 是否阻塞下一任务：`否`

## 11. 收工记录

- 构建是否通过：`通过（TypeScript noEmit）`
- 测试是否通过：`通过（no-browser + CC Switch self-start browser smoke 默认路径）`
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、当前状态、详细工作流、环境计划、前端主线状态、用户总账、上一轮 `CC Switch` worklog、automation memory、当前 smoke 脚本源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 当前状态 / 前端主线状态 / automation memory 一致指出 `CC Switch` 仍是同池最高优先 browser evidence 缺口之一。
  - 详细工作流要求本轮先保持停驻，再只选一个 screenshot-first 小闭环任务推进。
  - 上一轮 `CC Switch` worklog 提供了失败命令和 `attached-session` blocker 基线，便于本轮直接做 bootstrap 顺序对照而不是重新摸索。
- 本次使用了哪些子 agent 及其结论：`未使用`
- 手工 smoke 状态：`本轮通过 browser smoke 脚本完成浏览器级主证据，不再额外做人工浏览器点击`
- 手工 smoke 阻塞原因 / 缺少的环境：CLI 路径仍受宿主 `Node spawn EPERM` 影响；Edge remote debugging 端口在个别复跑里仍有启动波动
- 待验证页面清单：
  - `CC Switch` CLI browser smoke 路径（宿主 `spawn EPERM` 解除后再复跑）
  - 同池剩余页面的浏览器级中文残留与 hover/focus 证据
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务范围小、脚本上下文强耦合
- worklog 是否更新：`是`
- 遗留项：
  - `CC Switch` CLI browser smoke 仍未恢复
  - Edge remote debugging 端口偶发启动波动仍需保留为宿主诊断项
- 下一任务前置条件是否满足：`满足`

## 12. 下一轮建议

- 下一轮最适合继续推进：回到同一 `Phase G` screenshot-first 池，优先处理剩余浏览器级中文主显示泄漏，或补 `channel / group / model` 相邻页面里还未收紧的 `hover / focus` 小缺口。
- 同主线候选顺序：
  1. 继续清理同池页面的中文主显示残留与 browser evidence 漂移
  2. 视宿主情况复查 `CC Switch` CLI 路径的 `spawn EPERM`
  3. 仅在出现新的 `CC Switch` CDP 回归时，再回到 bootstrap 诊断层
