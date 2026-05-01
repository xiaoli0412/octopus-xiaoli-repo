# 2026-04-30 Phase G Page-Level Browser Alias Alignment

## 1. 任务信息

- 任务名称：剩余 page-level CDP thin forwarder `Browser` / `BrowserPath` alias 收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.4 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-arg-contract-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-settings-help-cli-wrapper-unification.md`
- 本次任务目标：把仍然只公开 `-Browser` 但未声明 `-BrowserPath` 兼容 alias 的 page-level CDP thin forwarder 收口到同池统一口径，并把 guard 覆盖扩到这批入口。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-arg-contract-alignment.md`
  - `scripts/verify-home-layout-browser-smoke.ps1`
  - `scripts/verify-model-layout-browser-smoke.ps1`
  - `scripts/verify-channel-page-browser-smoke.ps1`
  - `scripts/verify-ccswitch-browser-smoke.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮只把 `group-create-cdp` 与 `settings-help` 做了参数统一，当前最小相邻任务是把同池剩余 page-level CDP forwarder 补齐。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；用户已明确要求在既有主线内持续落代码，本轮不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 shared-wrapper 家族的参数一致性，不需要重新展开页面业务实现或后端链路。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务过小、强依赖现有 wrapper 上下文，串行主线程更稳妥。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 thin forwarder 参数契约与 repo-local guard，不改共享大 wrapper 行为、不改 page-level `mjs` 场景。
- 必须保留旧命令兼容性，不能把历史 `-BrowserPath` 调用打断。

## 4. 本次禁止事项

- 不回头重构 shared CDP wrapper 本身的参数名。
- 不扩散到 CLI wrapper、新页面或 AI 自动化 smoke。
- 不在没有验证的前提下顺手修改页面业务代码或运行态脚本。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `git diff --check -- scripts\verify-home-layout-browser-smoke.ps1 scripts\verify-model-layout-browser-smoke.ps1 scripts\verify-channel-page-browser-smoke.ps1 scripts\verify-ccswitch-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-page-level-browser-alias-alignment.md`

完成标准：

- `home-layout / model-layout / channel-page / ccswitch` 四条 page-level CDP thin forwarder 都公开 `-Browser`，同时保留 `-BrowserPath` alias；
- alignment guard 能检查这四条入口的 alias 与 `Browser -> BrowserPath` 透传契约；
- 相关 `check-only`、guard 与 `git diff --check` 均通过。

## 6. 本次回滚点

- `scripts/verify-home-layout-browser-smoke.ps1`
- `scripts/verify-model-layout-browser-smoke.ps1`
- `scripts/verify-channel-page-browser-smoke.ps1`
- `scripts/verify-ccswitch-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-browser-alias-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 thin forwarder 参数契约
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper 与 repo-local guard
- 受影响接口：无运行态 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：只影响 page-level smoke wrapper 的调用一致性；保留 alias 后旧命令仍兼容

## 8. 本轮微型计划

- 当前主线：`Phase G screenshot-first UI closure / browser smoke reliability`
- 当前阶段：page-level smoke wrapper 家族一致性收口
- 候选任务：
  - 检查 `home / model / channel-page / ccswitch` 是否仍缺 `BrowserPath` alias
  - 检查 shared-wrapper guard 是否遗漏这批入口
  - 检查是否还有 page-level forwarder 未透传 `RequireExternalCdpPreflight`
  - 检查是否还有复制版 CLI wrapper 残留
- 本轮核心任务：为剩余 page-level CDP thin forwarder 补齐 `BrowserPath` alias，并把 guard 扩到这批入口
- 本轮配套任务：同步前端主线状态与 worklog，给下一轮留下统一入口说明
- 预期验证方式：四条 `check-only`、alignment guard、`git diff --check`
- 完成判定：四条入口与 guard 都通过，且状态文档/worklog 同步完成

## 9. 实施步骤

1. 复核 automation memory、当前主线文档与最近 Phase G wrapper worklog，确认上一轮只统一了 `group-create-cdp` 与 `settings-help`，剩余同池 page-level CDP forwarder 仍是最值得推进的小缺口。
2. 用 `apply_patch` 为 `verify-home-layout-browser-smoke.ps1`、`verify-model-layout-browser-smoke.ps1`、`verify-channel-page-browser-smoke.ps1` 与 `verify-ccswitch-browser-smoke.ps1` 的 `Browser` 参数补上 `[Alias('BrowserPath')]`。
3. 同步升级 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，固定检查这四条入口仍保留 alias 与 `Browser -> BrowserPath` 透传，而不只是检查 shared-wrapper 引用。
4. 运行四条 `check-only`、alignment guard 与 `git diff --check`，确认本轮没有打断当前 screenshot-first 验证链。
5. 更新前端主线状态与本轮 worklog，说明当前统一口径已经覆盖到 page-level CDP wrapper 家族。

## 10. 风险与兼容性

- 新风险较低：修改的是 thin forwarder 参数面，不影响共享 wrapper 运行逻辑。
- 兼容性风险：极低；`[Alias('BrowserPath')]` 保住历史命令。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据或宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。四条 `check-only`、alignment guard 与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应留在 screenshot-first / browser smoke reliability 主线，不扩散到页面布局或业务逻辑。
  - 前端主线状态与相邻 worklog：确认 shared-wrapper 结构与前两条参数统一已完成，因此最小相邻缺口就是剩余 page-level CDP forwarder 的 alias 一致性。
  - automation memory：明确上一轮已经把 `group-create` / `settings-help` 做成统一公开参数面，本轮自然延续为“同池剩余页面补齐”。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 PowerShell `check-only` 与 repo-local guard。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要重开 self-start / external live path。
- 待验证页面清单：下一轮优先在健康宿主上继续验证 `home-layout`、`model-layout`、`channel-page`、`ccswitch` 与已完成统一的 `group-create` / `settings-help`。
- worklog 是否更新：yes
- 遗留项：
  - 目前统一的是 page-level CDP thin forwarder 的公开浏览器参数面；底层共享 wrapper 仍保留 `BrowserPath`，这是刻意保持最小风险的选择。
  - 当前 repo-local guard 只覆盖现有同池页面；如果后续新增同类 wrapper，需要在同一轮把 alias 断言补进去。
- 下一任务前置条件是否满足：满足。下一轮可继续沿同一 Phase G 主线推进真实 browser/CDP 证据或 host blocker 交接。

## 12. 执行与结果

1. 复核 automation memory、当前主线文档与最近 Phase G wrapper worklog 后确认：最近一轮只统一了 `group-create-cdp` 与 `settings-help` 的用户参数面，而 `home / model / channel-page / ccswitch` 这批 page-level CDP forwarder 仍没有显式 `BrowserPath` alias，是当前最小且连续的结构不一致点。
2. 因此本轮没有去改共享 `verify-channel-create-browser-smoke-cdp.ps1`，而是只对四条 thin forwarder 做最小补丁：
   - 继续公开 `Browser`
   - 用 `[Alias('BrowserPath')]` 保持旧命令兼容
   - 转发到共享 wrapper 时仍映射到底层 `BrowserPath`
3. 同时把 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 升级为不仅检查 shared-wrapper 引用和 scenario/success-marker/label，还固定检查上述四条 thin wrapper 也保留 `BrowserPath` alias 与 `Browser -> BrowserPath` 透传契约。
4. 验证结果显示：
   - 四条 `check-only -RequireExternalCdpPreflight` 都继续通过，说明参数收口没有影响共享 wrapper 的摘要输出；
   - alignment guard 继续通过，说明新增的 alias 断言与现有 thin forwarder 结构一致；
   - `git diff --check` 通过，工作区未引入格式问题。

## 13. 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `git diff --check -- scripts\verify-home-layout-browser-smoke.ps1 scripts\verify-model-layout-browser-smoke.ps1 scripts\verify-channel-page-browser-smoke.ps1 scripts\verify-ccswitch-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-page-level-browser-alias-alignment.md`
