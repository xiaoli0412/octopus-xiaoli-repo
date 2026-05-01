# 2026-04-30 Phase G Group Create CLI Wrapper Unification

## 1. 任务信息

- 任务名称：`group-create` 顶层 CLI wrapper 收口到共享 smoke helper
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.2 / 9.6 / 9.7 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-cdp-wrapper-false-positive-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-probe-browser-smoke-closure-and-settings-host-blocker-classification.md`
- 本次任务目标：把 `scripts/verify-group-create-browser-smoke.ps1` 的顶层 CLI 路径从自复制实现收口到共享 `verify-channel-create-browser-smoke.ps1`，让 `group-create` 和 `backup / ccswitch` 一样继承统一的 CLI host-blocker 分类与 success-marker/error-like stderr 护栏。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`
  - `scripts/verify-group-create-browser-smoke.ps1`
  - `scripts/verify-group-create-browser-smoke-cdp.ps1`
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-backup-browser-smoke.ps1`
  - `scripts/verify-ccswitch-browser-smoke-cli.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `using-superpowers` 约束核对
  - `brainstorming` 仅作非门禁式流程核对
- 若未使用部分本地资源或上下文，原因：本轮问题已明确收敛到 `group-create` CLI wrapper 结构不一致，不需要重新展开页面业务组件或更早期实现文档。
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只收口 `group-create` 顶层 CLI wrapper，不改 `mjs` smoke 场景、不动页面组件与后端实现。
- 以共享 wrapper 复用为优先，不重新扩散出另一份 CLI 包装逻辑。

## 4. 本次禁止事项

- 不回头重构 `group-create` 的 Node smoke 脚本内容。
- 不顺手改其它页面的 smoke 入口，除非本轮验证直接暴露同一问题。
- 不把宿主 `spawn EPERM` 重新误判成页面回归或 success-marker 漂移。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`

完成标准：

- `check-only -Driver cli` 能继续输出 `group-create` 自身的 `mjs` 场景摘要。
- `self-start -Driver cli` 若在本机失败，必须直接沿共享 CLI wrapper 的口径报出 Playwright CLI `spawn EPERM` host blocker，而不是保留旧的自复制错误语义。

## 6. 本次回滚点

- `scripts/verify-group-create-browser-smoke.ps1`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-cli-wrapper-unification.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本入口语义
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只影响 `group-create` 顶层 CLI wrapper 的复用路径与错误分类；CDP 路径保持既有 forwarder 结构不变

## 8. 实施步骤

1. 对比 `group-create` 顶层 wrapper 与 `backup / ccswitch` 已共享化的 CLI forwarder，确认 `group-create` 当前仍保留一份自复制 CLI 实现。
2. 用 `apply_patch` 将 `scripts/verify-group-create-browser-smoke.ps1` 收口为双路径 forwarder：`cdp` 继续 forward 到 `verify-group-create-browser-smoke-cdp.ps1`，`cli` 则改为通过环境变量转发到共享 `verify-channel-create-browser-smoke.ps1`。
3. 运行 `check-only -Driver cli` 与真实 `self-start -Driver cli`，确认 `group-create` 场景摘要仍正确，且本机失败已统一落到共享 wrapper 的 `spawn EPERM` host blocker 语义。

## 9. 测试与验证

- 构建命令：无
- 测试命令：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`
- 专项验证：
  - PowerShell parser 校验 `scripts\verify-group-create-browser-smoke.ps1`
  - `git diff --check -- scripts\verify-group-create-browser-smoke.ps1 docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-group-create-cli-wrapper-unification.md`

## 10. 风险与兼容性

- 新风险：较低；本轮只替换顶层 CLI wrapper 复用路径，不改 `group-create` 页面断言。
- 兼容性风险：`group-create` CLI 入口现在直接继承共享 wrapper 的默认 loopback/日志/error-like stderr 规则；若后续 `group-create` 需要独立于共享 wrapper 的 CLI 特殊行为，需要再明确记录，而不是重新复制整份脚本。
- 是否阻塞下一任务：不阻塞。下一轮可继续沿同一 Phase G 主线处理其它 wrapper 漂移或宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过 `check-only`、parser 校验与 `git diff --check`；真实 `self-start -Driver cli` 在本机按预期失败并被统一分类为 `spawn EPERM` host blocker。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke reliability 主线，不回到页面布局返工。
  - 前端主线状态与相邻 worklog：确认 `group-create` 的 CDP 路径已经共享化，但 CLI 顶层仍保留复制版实现，是当前最小且连续的结构不一致点。
  - automation memory：确认上一轮已把复制版 CDP wrapper 收口到共享 helper，因此本轮最自然的相邻动作是补齐 `group-create` CLI 顶层入口的一致性。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 PowerShell / Node browser smoke 入口。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 CLI self-start 仍受本机 Playwright CLI 子进程 `spawn EPERM` 阻塞，属于宿主执行环境问题，不是 `group-create` 页面回归。
- 待验证页面清单：`group-create` CLI 路径仍需在健康宿主上补一条真实 green pass；CDP 路径与其它页面级 browser smoke 仍按既有 host blocker 分类推进。
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。
- worklog 是否更新：yes
- 遗留项：
  - `group-create` 真实 CLI self-start 仍未在本机给出 green pass，当前只确认了共享 wrapper 后的宿主 blocker 语义正确。
  - 本轮没有扩散到 `backup / ccswitch / settings` 之外的其它 CLI 顶层入口；若后续发现更多复制版 CLI wrapper，再按同样方式逐个收口。
- 下一任务前置条件是否满足：满足。下一轮可以直接沿同一 Phase G 主线继续选择“其它 wrapper 漂移收口”或“已有 host blocker 的进一步分类/外部宿主复跑”。

## 12. 执行与结果

1. 先复核 automation memory、当前主线文档、前端主线状态与相邻 Phase G worklog，确认 `group-create` 当前业务 UI 已不缺主链路字段，最值得推进的是 smoke wrapper 结构一致性，而不是继续改 `GroupEditor.tsx`。
2. 对比 `scripts/verify-group-create-browser-smoke.ps1` 与 `scripts/verify-backup-browser-smoke.ps1` / `scripts/verify-ccswitch-browser-smoke-cli.ps1` 后确认：
   - `group-create` 的 `cdp` 路径已经 forward 到共享 `verify-group-create-browser-smoke-cdp.ps1`；
   - 但 `cli` 顶层仍保留一整份自复制实现，没有自动继承共享 `verify-channel-create-browser-smoke.ps1` 最近几轮新增的 loopback 预检、稳定日志读取、`spawn EPERM` host-blocker 分类与 error-like stderr 护栏。
3. 因此本轮没有继续微修旧复制版，而是直接把顶层脚本收口成双路径 forwarder：
   - `Driver=cdp` 维持既有参数透传到 `verify-group-create-browser-smoke-cdp.ps1`；
   - `Driver=cli` 改为通过 `OCTOPUS_UI_SMOKE_SCRIPT / ...SUCCESS_MARKER / ...LABEL` 转发到共享 `verify-channel-create-browser-smoke.ps1`，让它复用同一套 CLI 包装逻辑。
4. 验证结果表明：
   - `check-only -Driver cli` 仍能正确执行 `scripts/verify-group-create-browser-smoke.mjs --check-only`，说明 `group-create` 场景并未因 forwarder 收口而丢失自身参数与摘要；
   - 本机 `self-start -Driver cli` 现在会像 `backup` 一样直接报出 `Node group create is blocked by a host-level child-process 'spawn EPERM' failure while launching Playwright CLI`，不再把这一类宿主问题藏在旧的自复制 wrapper 语义里。
5. 同步更新前端主线状态，把这一轮结论记为“`group-create` CLI 顶层入口也已共享化，剩余 CLI gap 已收敛为宿主环境 blocker”。

## 13. 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only -Driver cli`
- observed expected host blocker `spawn EPERM` from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`
- passed PowerShell parser validation for `scripts\verify-group-create-browser-smoke.ps1`
- passed `git diff --check -- scripts\verify-group-create-browser-smoke.ps1 docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-group-create-cli-wrapper-unification.md`
