# 2026-04-30 Phase G Copied CDP Wrapper Preflight Closure

## 1. 任务信息

- 任务名称：复制版 CDP wrapper external-preflight 契约收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.2 / 9.6 / 9.7 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-cdp-preflight-forwarder-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-cdp-wrapper-false-positive-closure.md`
- 本次任务目标：把 `group-create-cdp` 与 `settings-help` 这两条仍使用复制版实现的 CDP wrapper 收口到共享 `verify-channel-create-browser-smoke-cdp.ps1` 的 external-preflight 契约，消除同池入口分裂。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-cdp-preflight-forwarder-alignment.md`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - `scripts/verify-group-create-browser-smoke.ps1`
  - `scripts/verify-group-create-browser-smoke-cdp.ps1`
  - `scripts/verify-setting-help-browser-smoke.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `using-superpowers` 约束核对
  - `brainstorming` 仅作非门禁式流程核对
- 若未使用部分本地资源或上下文，原因：本轮问题已收敛到 smoke wrapper 入口契约，不需要重新展开页面业务组件或更早期需求稿。
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只收口 CDP wrapper 入口契约，不改页面组件、不动后端行为。
- 优先复用共享 wrapper，避免继续复制 external-preflight 逻辑。

## 4. 本次禁止事项

- 不扩散到 CLI 路径逻辑重写。
- 不为了统一而回头大改页面级 smoke mjs 脚本。
- 不把宿主级 browser/CDP blocker 误判为本轮脚本未修复。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke-cdp.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`

以上两个入口都必须输出：

- `Explicit external CDP preflight requirement: enabled`
- `External mode initial CDP preflight: required`

并同步更新状态文档与 worklog。

## 6. 本次回滚点

- `scripts/verify-group-create-browser-smoke.ps1`
- `scripts/verify-group-create-browser-smoke-cdp.ps1`
- `scripts/verify-setting-help-browser-smoke.ps1`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本入口语义
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只影响 `group-create` / `settings-help` 的 CDP wrapper 参数透传与 external/check-only 说明口径

## 8. 实施步骤

1. 对比共享 wrapper 与两份复制版 wrapper，确认最小收口方案。
2. 把 `verify-group-create-browser-smoke-cdp.ps1` 改成 shared wrapper forwarder。
3. 把 `verify-group-create-browser-smoke.ps1` 顶层入口补齐 shared wrapper 所需的 CDP 参数透传。
4. 把 `verify-setting-help-browser-smoke.ps1` 的 `Driver=cdp` 路径前移到 shared wrapper，保留 CLI 路径不变。
5. 用 `check-only -RequireExternalCdpPreflight` 证明两个入口都拿到统一 preflight 说明。

## 9. 测试与验证

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke-cdp.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `git diff --check -- scripts\verify-group-create-browser-smoke.ps1 scripts\verify-group-create-browser-smoke-cdp.ps1 scripts\verify-setting-help-browser-smoke.ps1 docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`

## 10. 风险与兼容性

- 风险较低：本轮只做 PowerShell wrapper 入口收口，不改底层 node smoke 断言。
- 兼容性风险：`settings-help` 保留 CLI 路径，因此本轮没有统一所有 driver 的参数集合；这是刻意范围控制。
- 是否阻塞下一任务：不阻塞。下一轮可直接回到真实 host blocker 分类或健康宿主上的 browser 证据复跑。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。两个目标入口的 `check-only -RequireExternalCdpPreflight` 已按预期输出统一 preflight 说明。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke 主线，不回到页面布局返工。
  - 前端主线状态与上一轮 worklog：确认高优先级缺口已收敛成“复制版 wrapper 尚未对齐 shared external-preflight helper”。
  - automation memory：确认上一轮已修好 shared wrapper 假阳性，因此本轮最合理的相邻动作是把剩余复制版入口向 shared helper 收口。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行；本轮只做 wrapper 入口与 `check-only` 验证。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 external/self-start browser smoke 仍受本机 host-level CDP/CLI 条件影响，本轮无需为入口契约问题重跑整条页面流程。
- 待验证页面清单：`group-create`、`settings help`、`channel-create` 等页面级 browser smoke 仍需在健康宿主上补真实 external/self-start 证据。
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。
- worklog 是否更新：yes
- 遗留项：
  - `settings-help` 的 CLI 路径仍保留自有实现，本轮未强行统一该支路。
  - 当前 Phase G 剩余重点已回到 host blocker 分类与健康宿主上的真实 browser 证据，不再是入口契约分裂。
- 下一任务前置条件是否满足：满足。下一轮可直接在同主线下选择“重跑 group/settings 的真实 browser 证据”或“继续收敛宿主级 blocker 记录”。

## 12. 执行与结果

1. 先复核 `UI_MAINLINE_TASK_2026-04-30`、前端主线状态、相邻 worklog 和 automation memory，确认当前主线仍是 `Phase G screenshot-first UI closure / browser smoke reliability`，且上一轮已把 `group-create-cdp` 与 `settings-help` 标成剩余复制版 wrapper 缺口。
2. 对比共享 `verify-channel-create-browser-smoke-cdp.ps1` 与两份目标脚本后，确定最小收口方案不是在复制版里继续补 helper，而是直接把 CDP 路径 forward 到共享 wrapper：
  - `verify-group-create-browser-smoke-cdp.ps1` 改成 shared wrapper forwarder，并配置 `NodeSmokeScript='scripts/verify-group-create-browser-smoke-cdp.mjs'`、`NodeSmokeSuccessMarker='group-create-browser-smoke-cdp passed'`、`SmokeLabel='group create'`。
  - `verify-group-create-browser-smoke.ps1` 继续保留 CLI 路径，但给 `Driver=cdp` 的 forwardParams 补齐 `CdpPort / CdpUrl / CdpCommandTimeoutMs / EdgeLaunchPreset / EdgeProfileStrategy / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder / BootstrapExternalCdpSession / RequireExternalCdpPreflight / SelfStartServices`，避免顶层入口继续截断共享 wrapper 的参数。
  - `verify-setting-help-browser-smoke.ps1` 则把 `Driver=cdp` 路径直接前移为 shared wrapper forwarder，保留原有 CLI 路径不变。
3. 验证结果显示，两条目标入口现在都会输出 shared wrapper 同款的 `Explicit external CDP preflight requirement: enabled` 与 `External mode initial CDP preflight: required`，说明本轮真正收口的是 external-preflight 契约，而不是只在局部脚本上加了表面参数。
4. 同步更新前端主线状态，把“剩余复制版 wrapper 缺口”改写成“已收口，剩余焦点回到 host blocker 和健康宿主证据”。

## 13. 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke-cdp.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`

