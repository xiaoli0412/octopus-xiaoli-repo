# 2026-04-30 Phase G Browser Smoke Wrapper Alignment Guard

## 1. 任务信息

- 任务名称：page-level browser smoke shared-wrapper 对齐守门
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.2 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-cli-wrapper-unification.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-settings-help-cli-wrapper-unification.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`
- 本次任务目标：把最近几轮已经收口到共享 helper 的 page-level smoke wrapper 固化成 repo-local 守门，避免后续再次回退成复制版 PowerShell 入口。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `scripts/run-frontend-verification-suite.mjs`
  - `scripts/verify-backup-browser-smoke.ps1`
  - `scripts/verify-ccswitch-browser-smoke-cli.ps1`
  - `scripts/verify-ccswitch-browser-smoke.ps1`
  - `scripts/verify-group-create-browser-smoke.ps1`
  - `scripts/verify-group-create-browser-smoke-cdp.ps1`
  - `scripts/verify-home-layout-browser-smoke.ps1`
  - `scripts/verify-model-layout-browser-smoke.ps1`
  - `scripts/verify-channel-page-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已完成 `settings-help` CLI wrapper 收口，因此本轮最合理的相邻动作不是继续改页面，而是把 shared-wrapper 结构加上 repo-local 护栏。
  - `using-superpowers`：按要求核对技能使用边界。
  - `brainstorming`：仅作非门禁式约束核对；本轮属于既有 Phase G 验证守门补强，不进入设计审批流。
- 若未使用部分本地资源或上下文，原因：本轮问题已缩小到验证脚本结构守门，不需要重新展开业务组件、后端契约或更早的需求稿。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只补 repo-local 验证守门，不回头修改业务页面或后端逻辑。
- 守门脚本必须对已收口的 wrapper 结构给出可重复、可失败的断言，而不是只做人工说明。

## 4. 本次禁止事项

- 不把本轮扩大成对 `AI 自动化` browser smoke wrapper 的再次重构。
- 不顺手修改 page-level `mjs` smoke 场景或页面组件。
- 不新增与当前主线无关的 lint/test 编排。

## 5. 本次验收条件

- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs scripts\run-frontend-verification-suite.mjs web\package.json docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`

完成标准：

- 新守门脚本能稳定通过并检查共享 wrapper 关键入口未回退；
- `screenshot` no-browser 套件会自动覆盖该守门；
- 状态文档与 worklog 明确记录本轮收口和下一轮入口。

## 6. 本次回滚点

- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `scripts/run-frontend-verification-suite.mjs`
- `web/package.json`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证守门
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整前端 no-browser 验证编排
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只影响 repo-local 验证链，会更早暴露 wrapper 回退问题

## 8. 实施步骤

1. 先复核当前主线文档、前端状态和相邻 Phase G worklog，确认 `backup / ccswitch / group-create / settings-help` 的 CLI wrapper 与多条 page-level CDP wrapper 都已收口到共享 helper。
2. 新增 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，固定检查：
   - 相关 wrapper 是否仍引用共享 `verify-channel-create-browser-smoke.ps1` 或 `verify-channel-create-browser-smoke-cdp.ps1`
   - 各自的 scenario / success marker / label 是否仍指向本页专属配置
   - 脚本长度是否仍维持“薄 forwarder”级别
   - 是否重新长回 `Wait-HttpOk / Start-LocalOctopusBackend / Assert-NodeSmokeSucceeded` 等复制版 helper
3. 将该脚本接入 `scripts/run-frontend-verification-suite.mjs screenshot`，并在 `web/package.json` 增加 `test:browser-smoke-wrapper-alignment` 入口。
4. 跑新脚本与整条 `screenshot` no-browser 套件，确认守门已经真正接到日常验证链，而不是孤立脚本。

## 9. 测试与验证

- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs scripts\run-frontend-verification-suite.mjs web\package.json docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`

## 10. 风险与兼容性

- 新风险较低：新增的是 repo-local 结构守门，不改变页面运行时逻辑。
- 兼容性风险：若后续某个页面确实需要脱离共享 helper，必须同步更新守门脚本，否则会先在 no-browser 套件失败。
- 是否阻塞下一任务：不阻塞。下一轮可以直接切回“健康宿主上的真实 browser/CDP 证据”或继续收宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。新守门脚本与整条 `screenshot` no-browser 套件均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍属于 screenshot-first / browser smoke reliability 主线，不应发散回业务页改造。
  - 前端主线状态与相邻 worklog：确认共享 wrapper 收口本身已基本完成，最值得补的是 repo-local 守门，而不是继续人工盘点脚本结构。
  - automation memory：明确上一轮已经把 `settings-help` CLI 收口完，本轮应补“防回退护栏”。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行；本轮只做 repo-local no-browser 验证守门。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser-grade 证据仍需健康宿主或外部会话；本轮不需要重新拉起页面级 self-start/external 流程。
- 待验证页面清单：下一轮优先回到已分类 host blocker 的真实 browser 证据补跑，尤其 `settings-help`、`group-create`、`channel-page` 等已完成 wrapper 收口的入口。
- worklog 是否更新：yes
- 遗留项：
  - 当前守门只覆盖已进入共享 wrapper 家族的 page-level 入口；若后续新增新页面 smoke，需要把对应 forwarder 合同纳入这条脚本。
  - 守门确认的是“结构未回退”，不是“真实宿主已通过 browser smoke”；宿主 blocker 仍需另外取证。
- 下一任务前置条件是否满足：满足。下一轮可以直接沿同一 Phase G 主线继续跑健康宿主上的真实 browser/CDP 证据，或补最后剩余的 host blocker 交接说明。

## 12. 执行与结果

1. 复核 automation memory、当前主线文档和最近几份 Phase G worklog 后确认：`settings-help` / `group-create` 最近两轮已经把 CLI wrapper 收口到共享 helper，`home-layout / model-layout / channel-page / ccswitch / group-create-cdp` 的 CDP 入口也已经共享化，因此本轮最有价值的相邻动作不是再改脚本，而是补“防回退护栏”。
2. 直接新增 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，把 shared-wrapper 家族当前应满足的结构固化成 repo-local 断言：
   - CLI forwarder 必须继续指向共享 `verify-channel-create-browser-smoke.ps1`
   - CDP forwarder 必须继续指向共享 `verify-channel-create-browser-smoke-cdp.ps1`
   - 每个入口仍保留本页专属 `scenario / success marker / label`
   - 已收口的脚本不能重新长回复制版 helper
3. 把该守门接入 `scripts/run-frontend-verification-suite.mjs screenshot` 与 `web/package.json`，让这条结构检查不再依赖人工自觉，而是自动成为同池 no-browser 验证的一部分。
4. 验证结果显示：新脚本单独运行通过，整条 `screenshot` no-browser 套件也继续保持绿色，说明本轮增加的是有效护栏，而不是引入新的验证噪音。

## 13. 验证

- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs scripts\run-frontend-verification-suite.mjs web\package.json docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`
