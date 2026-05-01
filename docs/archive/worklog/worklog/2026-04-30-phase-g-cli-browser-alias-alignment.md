# 2026-04-30 Phase G CLI Browser Alias Alignment

## 1. 任务信息

- 任务名称：CLI thin forwarder `Browser` / `BrowserPath` alias 收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.4 / 9.6 / 9.7 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-browser-alias-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-cli-wrapper-unification.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-settings-help-cli-wrapper-unification.md`
- 本次任务目标：把 `backup`、`ccswitch-cli` 与 `group-create` 三条 CLI thin forwarder 的公开浏览器参数统一到 `-Browser`，同时保留 `-BrowserPath` 兼容 alias，并把该契约纳入 wrapper guard。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-browser-alias-alignment.md`
  - `scripts/verify-backup-browser-smoke.ps1`
  - `scripts/verify-ccswitch-browser-smoke-cli.ps1`
  - `scripts/verify-group-create-browser-smoke.ps1`
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已统一 page-level CDP forwarder，当前最小相邻任务是补齐 CLI forwarder 的同类参数面。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；用户已明确要求在既有主线内持续落代码，本轮不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 shared-wrapper 家族的 CLI 参数一致性，不需要重开页面业务实现或后端重构链。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小且强依赖现有 wrapper 上下文，串行主线程更稳妥。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 CLI thin forwarder 参数契约与 repo-local guard，不改共享 wrapper 执行逻辑、不改页面级 `mjs` 场景。
- 必须保留旧命令兼容性，不能打断历史 `-BrowserPath` 调用。

## 4. 本次禁止事项

- 不回头修改共享 `verify-channel-create-browser-smoke.ps1` 的参数名。
- 不扩散到页面 UI、CDP scene 逻辑或 AI 自动化 smoke。
- 不在没有验证的前提下顺手调整浏览器运行链。

## 5. 本次验收条件

- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke-cli.ps1 -Mode check-only`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only -BrowserPath $env:COMSPEC`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke-cli.ps1 -Mode check-only -BrowserPath $env:COMSPEC`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli -BrowserPath $env:COMSPEC`
- `git diff --check -- scripts\verify-backup-browser-smoke.ps1 scripts\verify-ccswitch-browser-smoke-cli.ps1 scripts\verify-group-create-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-cli-browser-alias-alignment.md`

完成标准：

- 三条 CLI thin forwarder 都公开 `-Browser`，同时保留 `-BrowserPath` alias；
- alignment guard 能检查三条 CLI 入口的 alias 与透传契约；
- 默认 `check-only`、旧 `-BrowserPath` 兼容 `check-only` 与 `git diff --check` 均通过。

## 6. 本次回滚点

- `scripts/verify-backup-browser-smoke.ps1`
- `scripts/verify-ccswitch-browser-smoke-cli.ps1`
- `scripts/verify-group-create-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-cli-browser-alias-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 CLI thin forwarder 参数契约
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper 与 repo-local guard
- 受影响接口：无运行态 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：只影响 CLI smoke wrapper 的调用一致性；保留 alias 后旧命令仍兼容

## 8. 本轮微型计划

- 当前主线：`Phase G screenshot-first UI closure / browser smoke reliability`
- 当前阶段：browser smoke wrapper 家族参数面收口
- 候选任务：
  - 检查 CLI thin forwarder 是否仍缺 `BrowserPath` alias
  - 检查 wrapper guard 是否遗漏 CLI 入口的 alias/透传断言
  - 检查旧 `-BrowserPath` 命令是否仍能走到共享 wrapper
  - 检查是否还有新的 copied CLI wrapper 回退
- 本轮核心任务：为 `backup / ccswitch-cli / group-create` 三条 CLI thin forwarder 补齐 `BrowserPath` alias，并扩展 guard
- 本轮配套任务：同步前端主线状态与本轮 worklog，给下一轮留下统一入口说明
- 预期验证方式：guard、默认 `check-only`、旧参数兼容 `check-only`、`git diff --check`
- 完成判定：三条入口与 guard 都通过，且状态文档/worklog 同步完成

## 9. 实施步骤

1. 复核 automation memory、当前主线文档、前端主线状态和最近两份 alias/workrapper worklog，确认上一轮已统一 page-level CDP forwarder，当前最小相邻缺口是 CLI thin forwarder 的公开浏览器参数仍未完全对齐。
2. 用 `apply_patch` 为 `verify-backup-browser-smoke.ps1`、`verify-ccswitch-browser-smoke-cli.ps1` 与 `verify-group-create-browser-smoke.ps1` 的 `Browser` 参数补上 `[Alias('BrowserPath')]`。
3. 同步升级 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，固定检查三条 CLI wrapper 也保留 alias 与 `Browser` 透传逻辑，而不只检查 shared-wrapper 引用。
4. 运行默认 `check-only`、旧 `-BrowserPath` 兼容 `check-only`、alignment guard 与 `git diff --check`，确认本轮没有打断当前 screenshot-first 验证链。
5. 更新前端主线状态与本轮 worklog，说明当前统一口径已经覆盖到 CLI 与 CDP 两侧的 thin forwarder。

## 10. 风险与兼容性

- 新风险较低：修改的是 thin forwarder 参数面，不影响共享 wrapper 的执行逻辑。
- 兼容性风险：极低；`[Alias('BrowserPath')]` 保住历史命令。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据或宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。默认 `check-only`、旧 `-BrowserPath` 兼容 `check-only`、alignment guard 与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应留在 screenshot-first / browser smoke reliability 主线，不扩散到页面布局或业务逻辑。
  - 前端主线状态与相邻 worklog：确认 page-level CDP forwarder 已统一，因此最小相邻缺口就是 CLI forwarder 的 alias 一致性。
  - automation memory：明确上一轮已经把同池 CDP 入口统一，本轮自然延续为“CLI 入口补齐”。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 PowerShell `check-only` 与 repo-local guard。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要重开 self-start / external live path。
- 待验证页面清单：下一轮优先在健康宿主上继续验证 `backup`、`ccswitch`、`group-create` 与已统一过的 page-level CDP wrappers 的真实 browser/CDP 证据。
- worklog 是否更新：yes
- 遗留项：
  - 底层共享 wrapper 仍刻意保留现有参数名，这是最小风险选择。
  - 当前 repo-local guard 只覆盖现有同池 wrapper；若后续新增同类入口，需要在同一轮补进断言。
- 下一任务前置条件是否满足：满足。下一轮可继续沿同一 Phase G 主线推进真实 browser/CDP 证据或 host blocker 交接。

## 12. 执行与结果

1. 复核 automation memory、当前主线文档与最近 Phase G wrapper worklog 后确认：最近两轮已经把 page-level CDP thin forwarder 统一到公开 `-Browser` + 兼容 `-BrowserPath`，而 `backup / ccswitch-cli / group-create` 这批 CLI thin forwarder 仍只暴露 `-Browser`，是当前最小且连续的结构不一致点。
2. 因此本轮没有改共享 `verify-channel-create-browser-smoke.ps1`，而是只对三条 CLI thin forwarder 做最小补丁：
   - 继续公开 `Browser`
   - 用 `[Alias('BrowserPath')]` 保持旧命令兼容
   - 保持 forward 到共享 wrapper 的现有参数映射，不改变 CLI 执行链
3. 同时把 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 升级为不仅检查 shared-wrapper 引用和 success-marker/label，也固定检查三条 CLI wrapper 保留 alias 与 `Browser` 透传契约。
4. 验证结果显示：
   - 三条默认 `check-only` 继续通过，说明参数收口没有影响共享 wrapper 的摘要输出；
   - 三条旧 `-BrowserPath` 兼容 `check-only` 也继续通过，而且 `browserName` 摘要正确显示传入的 `C:\Windows\System32\cmd.exe`，证明 alias 真正生效，不只是静态声明；
   - alignment guard 与 `git diff --check` 继续通过，说明新增 alias 断言和实际 thin forwarder 结构一致。

## 13. 验证

- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke-cli.ps1 -Mode check-only`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only -BrowserPath $env:COMSPEC`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke-cli.ps1 -Mode check-only -BrowserPath $env:COMSPEC`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli -BrowserPath $env:COMSPEC`
- passed `git diff --check -- scripts\verify-backup-browser-smoke.ps1 scripts\verify-ccswitch-browser-smoke-cli.ps1 scripts\verify-group-create-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-cli-browser-alias-alignment.md`
