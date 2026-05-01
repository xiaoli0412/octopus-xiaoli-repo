# 2026-04-30 Phase G AI Learning Browser Alias Guard Alignment

## 1. 任务信息

- 任务名称：`ai-learning` specialized root `Browser` / `BrowserPath` alias 与 guard 契约收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.4 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-wrapper-inventory-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-root-browser-alias-alignment.md`
- 本次任务目标：把 browser smoke PowerShell 家族里最后一个仍未显式声明 `BrowserPath` 兼容 alias 的受控特例 `verify-ai-automation-learning-browser-smoke.ps1` 收口到同池统一口径，并把它额外持有的 host-friendly / self-start 特例契约补进 shared guard。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-wrapper-inventory-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-root-browser-alias-alignment.md`
  - `scripts/verify-ai-automation-learning-browser-smoke.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已把 thin forwarder 与 shared root 的 `Browser/BrowserPath` 参数面收口，当前最小连续任务是 specialized root 的最后一处参数/guard 缺口。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；用户已明确要求在既有主线内直接推进代码，本轮不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 wrapper 参数与 guard 契约，不涉及页面源码、后端业务或真实 live rerun。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、范围集中，主线程即可闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 `ai-learning` specialized root 的公开参数面、repo-local guard 与文档记录，不改页面 smoke `mjs`、不改共享运行逻辑。
- 必须保留 `SelfStartServices / SelfStartLocalServices` 双入口与 host-friendly 外部默认参数语义，不能打断既有 worklog 与 replay 命令。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP live rerun。
- 不顺手修改 shared root 或其它 page-level wrapper 的执行逻辑。
- 不改变 `ai-learning` 的 stable diagnostic / host handoff / next-step 语义。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -BrowserPath $env:COMSPEC`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -Browser $env:COMSPEC`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-ai-automation-learning-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-ai-learning-browser-alias-guard-alignment.md`

完成标准：`ai-learning` shared-specialized root 新旧两种浏览器参数都能通过 `check-only`，guard 明确要求 alias 与 self-start / host-friendly 特例契约，screenshot no-browser 聚合入口继续通过。

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-ai-learning-browser-alias-guard-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 specialized root 参数语义
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper guard 与文档记录
- 受影响接口：无运行态 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：只影响 `ai-learning` wrapper 的公开参数面和静态 guard，不影响既有旧命令兼容性

## 8. 本轮微型计划

- 当前主线：`Phase G screenshot-first UI closure / browser smoke reliability`
- 当前阶段：browser smoke wrapper 参数契约收口
- 候选任务：
  - 统一 `ai-learning` specialized root 的 `Browser` / `BrowserPath` 参数面
  - 升级 shared guard，补 specialized root 的特例契约
  - 继续新增 browser wrapper 家族静态分类校验
- 本轮核心任务：把 `verify-ai-automation-learning-browser-smoke.ps1` 改为显式接受 `-BrowserPath` alias
- 本轮配套任务：同步升级 guard 与状态/worklog 记录
- 预期验证方式：`ai-learning` `check-only`、alignment guard、screenshot 聚合入口、`git diff --check`
- 完成判定：`ai-learning` wrapper 新旧两种命令都可通过，guard 与 no-browser 聚合入口通过

## 9. 实施步骤

1. 复核 automation memory、当前主线文档和最近 Phase G wrapper worklog，确认 thin forwarder 与 shared root 已收口，当前剩余结构不一致只在 `ai-learning` specialized root。
2. 用 `apply_patch` 最小修改 `scripts/verify-ai-automation-learning-browser-smoke.ps1`：
   - 为 `Browser` 参数补上 `[Alias('BrowserPath')]`；
   - 保留内部 `Browser -> BrowserPath` 映射，不改执行逻辑。
3. 同步升级 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，把 `ai-learning` specialized root 的 alias、`SelfStartServices / SelfStartLocalServices` 合流与 host-friendly 默认 `BootstrapExternalCdpSession` 契约纳入 guard。
4. 运行 `ai-learning` `check-only`、alignment guard、screenshot no-browser 聚合入口和 `git diff --check`，确认本轮是实质兼容，而不是静态文本变更。
5. 更新状态文档、automation memory 与本 worklog，为下一轮继续真实 browser/CDP 证据留下统一入口。

## 10. 风险与兼容性

- 新风险较低：`ai-learning` wrapper 的内部执行逻辑未改，只是统一公开参数面并加严静态 guard。
- 兼容性风险：极低；`[Alias('BrowserPath')]` 保证旧命令继续可用。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据与宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。`ai-learning` `check-only`、alignment guard、screenshot no-browser 聚合入口与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke reliability 主线，不扩散到页面布局。
  - `DETAILED_EXECUTION_WORKFLOW`：确认当前窗口优先级仍是图片问题池之后的 wrapper/reliability 收口，而不是回到其它 canonical 阶段。
  - 相邻 wrapper worklog 与 automation memory：确认上一轮已完成 thin forwarder 与 shared root 参数统一，本轮自然承接为 specialized root 参数面收口。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 repo-local `check-only` 与 no-browser 聚合入口。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要重开 self-start / external live path。
- 待验证页面清单：下一轮优先在健康宿主上继续验证 `backup`、`ccswitch`、`group-create`、`home-layout`、`model-layout`、`channel-page` 与 `settings-help` 的真实 browser/CDP 证据，以及 `ai-learning` 的外部 replay/handoff 命令。
- worklog 是否更新：yes
- 遗留项：
  - 当前 specialized root 已补 alias，本轮只收口公开参数面与 guard；如后续还要调整 `ai-learning` replay/host-handoff 输出，应另开同主线任务。
  - 当前主线剩余工作继续收敛为真实 browser/CDP 证据与宿主 blocker，不再是 wrapper 参数面分裂。
- 下一任务前置条件是否满足：满足。下一轮可直接沿同一 Phase G 主线继续真实 browser/CDP 证据或宿主 blocker 分类。

## 12. 执行与结果

1. 复核 automation memory、当前主线文档与最近三份 wrapper worklog 后确认：`backup / ccswitch / group-create / settings-help / home / model / channel-page / shared roots` 这些入口的 `Browser/BrowserPath` 参数面都已收口，但 `verify-ai-automation-learning-browser-smoke.ps1` 仍只公开 `-Browser`，是当前最小且连续的结构不一致点。
2. 因此本轮没有再改任何页面场景脚本，只对 specialized root 做最小补丁：
   - `verify-ai-automation-learning-browser-smoke.ps1` 现在显式公开 `-Browser`，同时保留 `[Alias('BrowserPath')]`；
   - 内部 helper 与 bootstrap 仍继续映射到底层 `BrowserPath`，避免动到实际执行逻辑；
   - `verify-browser-smoke-wrapper-alignment.mjs` 也同步升级为要求 specialized root 保留 alias，并显式检查 `SelfStartServices / SelfStartLocalServices` 合流与 host-friendly 默认 `BootstrapExternalCdpSession` 契约。
3. 验证结果显示：
   - `ai-learning` wrapper 的旧 `-BrowserPath` 与新 `-Browser` 两种 `check-only` 调用都通过，说明兼容与统一同时成立；
   - alignment guard 与 screenshot no-browser 聚合入口继续通过，说明本轮没有打断当前 repo-local 验证编排。

## 13. 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -BrowserPath $env:COMSPEC`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -Browser $env:COMSPEC`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-ai-automation-learning-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-ai-learning-browser-alias-guard-alignment.md`
