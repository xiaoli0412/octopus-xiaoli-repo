# 2026-04-30 Phase G Browser Arg Contract Alignment

## 1. 任务信息

- 任务名称：page-level browser smoke thin forwarder `Browser` 参数契约统一
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.2 / 9.6 / 9.7 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-cli-wrapper-unification.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-settings-help-cli-wrapper-unification.md`
- 本次任务目标：把 `group-create-cdp` 与 `settings-help` thin forwarder 的浏览器可执行路径参数统一到同池页面一致的 `-Browser` 口径，同时保留 `-BrowserPath` 兼容 alias，并将该契约纳入 alignment guard。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `scripts/verify-group-create-browser-smoke-cdp.ps1`
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已经把 shared-wrapper 结构做成 repo-local 守门，因此本轮最合适的相邻动作是补“调用参数面”的一致性，而不是再改页面或大 wrapper。
  - `using-superpowers`：按要求核对技能使用边界。
  - `brainstorming`：仅作非门禁式核对；本轮不是新功能设计，而是既有验证主线上的 thin forwarder 契约收口。
- 若未使用部分本地资源或上下文，原因：本轮已明确收敛到 wrapper 参数契约，不需要重新展开业务 UI、后端 schema 或更老的主线规划。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 thin forwarder 的参数契约与守门，不改共享大 wrapper 行为、不改页面业务实现。
- 统一口径必须兼容现有 `-BrowserPath` 调用，避免破坏历史命令或 worklog 记录。

## 4. 本次禁止事项

- 不重开 page-level `mjs` smoke 场景或页面组件修复。
- 不把本轮扩大到 `AI 自动化` 复杂 wrapper 的参数重构。
- 不在没有验证的前提下顺手修改其它 thin wrapper 的默认参数。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke-cdp.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `git diff --check -- scripts\verify-group-create-browser-smoke-cdp.ps1 scripts\verify-setting-help-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs`

完成标准：

- 两条 thin forwarder 都接受统一的 `-Browser` 公开参数，且仍兼容旧的 `-BrowserPath` 调用；
- alignment guard 能直接检查这条契约，不再只守结构不守入口参数面；
- 相关 `check-only` 入口与 `git diff --check` 均通过。

## 6. 本次回滚点

- `scripts/verify-group-create-browser-smoke-cdp.ps1`
- `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-arg-contract-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 thin forwarder 参数契约
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper 与 no-browser 守门
- 受影响接口：无运行态 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：只影响 page-level smoke wrapper 的调用一致性；保留 `-BrowserPath` alias 后，旧命令仍兼容

## 8. 实施步骤

1. 复核当前主线文档、automation memory 与最近两份 Phase G wrapper worklog，确认 shared-wrapper 结构已守住，当前剩余最小不一致是 `group-create-cdp` / `settings-help` 仍直接暴露 `-BrowserPath`，而同池其它页面更常用 `-Browser`。
2. 用 `apply_patch` 将两条 thin forwarder 的浏览器参数统一为：公开参数名 `Browser`，并通过 `[Alias('BrowserPath')]` 保持向后兼容；forward 到共享 wrapper 时仍映射到底层 `BrowserPath` 或 CLI `Browser` 目标字段。
3. 同步更新 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，让 guard 不只检查“有没有共享化”，还检查 `Browser` alias/透传契约是否存在。
4. 运行两条 `check-only`、alignment guard 与 `git diff --check`，确认本轮收口没有打断当前 screenshot-first 验证链。

## 9. 测试与验证

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke-cdp.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `git diff --check -- scripts\verify-group-create-browser-smoke-cdp.ps1 scripts\verify-setting-help-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs`

## 10. 风险与兼容性

- 新风险较低：修改的是 thin forwarder 参数面，不影响共享大 wrapper 的运行逻辑。
- 兼容性风险：极低；保留 `[Alias('BrowserPath')]` 后，现有历史命令与 worklog 中的 `-BrowserPath` 仍可继续使用。
- 是否阻塞下一任务：不阻塞。下一轮可直接继续健康宿主上的真实 browser/CDP 证据，或继续同池 host blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。两条 `check-only`、alignment guard 与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应留在 screenshot-first / browser smoke reliability 主线，不扩散到页面布局或业务逻辑。
  - 前端主线状态与最近 worklog：确认 shared-wrapper 结构守门已完成，因此更值得补的是“thin wrapper 调用参数也统一”。
  - automation memory：明确上一轮已经把 wrapper alignment guard 接进验证链，本轮自然延续为 guard 覆盖参数契约。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 PowerShell `check-only` 与 repo-local no-browser guard。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要重开 self-start/external live path。
- 待验证页面清单：下一轮优先在健康宿主上继续验证 `group-create`、`settings-help`、`channel-page` 等已经完成 shared-wrapper + arg-contract 收口的页面级入口。
- worklog 是否更新：yes
- 遗留项：
  - 当前统一的是 `group-create-cdp` 与 `settings-help` 两条 thin forwarder；如果后续再出现新页面同时公开 `-BrowserPath` / `-Browser` 的分裂，需要继续纳入 alignment guard。
  - 本轮没有触碰 shared wrapper 自身的参数名；底层仍保持 `BrowserPath`，这是刻意保持最小风险的选择。
- 下一任务前置条件是否满足：满足。下一轮可以继续沿同一 Phase G 主线推进真实 browser/CDP 证据或 host blocker 交接。

## 12. 执行与结果

1. 复核 automation memory、当前主线文档与最近 Phase G wrapper worklog 后确认：最近两轮已经把 shared-wrapper 结构本身守住，当前最值得推进的相邻小任务不是继续拆 page-level wrapper，而是把 thin forwarder 的用户参数面也统一掉，避免只有个别页面仍要求单独记忆 `-BrowserPath`。
2. 因此本轮没有去改共享 `verify-channel-create-browser-smoke-cdp.ps1`，而是只对 `verify-group-create-browser-smoke-cdp.ps1` 与 `verify-setting-help-browser-smoke.ps1` 做最小补丁：
   - 公开参数名统一改为 `Browser`
   - 用 `[Alias('BrowserPath')]` 保持旧命令兼容
   - 转发到共享 wrapper 时继续映射到底层 `BrowserPath` / CLI `Browser` 字段
3. 同时把 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 升级为不仅检查 shared-wrapper 引用和 scenario/success-marker/label，还固定检查上述两条 thin wrapper 仍保留 `BrowserPath` alias 与 `Browser -> BrowserPath` 透传契约。
4. 验证结果显示：
   - `group-create-cdp` 与 `settings-help cdp` 的 `check-only -RequireExternalCdpPreflight` 继续通过，说明参数收口没有影响共享 wrapper 的摘要输出；
   - alignment guard 继续通过，说明新增的契约断言与现有 thin forwarder 结构一致；
   - `git diff --check` 通过，工作区未引入格式问题。

## 13. 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke-cdp.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `git diff --check -- scripts\verify-group-create-browser-smoke-cdp.ps1 scripts\verify-setting-help-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs`
