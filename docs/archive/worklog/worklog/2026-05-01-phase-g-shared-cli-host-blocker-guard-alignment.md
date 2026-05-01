# 2026-05-01 Phase G Shared CLI Host Blocker Guard Alignment

## 1. 任务信息

- 任务名称：共享 CLI browser smoke 宿主 blocker 守门补齐
- 日期：2026-05-01
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-backup-cli-host-blocker-classification-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-cdp-usage-entry-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-settings-help-bootstrap-default-guard.md`
- 本次任务目标：把共享 CLI wrapper 最近已修复但还未被 guard 锁住的宿主 blocker 契约补成 repo-local 守门，避免 `backup / ccswitch / group-create / settings-help` 后续再回退成误导性的 `channel create` 专属报错。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-backup-cli-host-blocker-classification-closure.md`
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - `scripts/verify-backup-browser-smoke.ps1`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已经收口 shared CDP usage 与 `settings-help` 默认 bootstrap 守门，因此本轮最连续的小任务应继续停留在 wrapper/guard 漂移，而不是切回 live rerun。
  - `using-superpowers`：按会话要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；本轮属于既有 Phase G 验证链守门补丁，不进入设计审批流。
- 若未使用部分本地资源或上下文，原因：本轮只处理共享 CLI wrapper 的静态守门与交接记录，不涉及页面业务、后端数据结构或 live browser 证据。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行。
- 若使用分目录 agent，负责目录与禁止越界范围：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、上下文强耦合，主线程串行即可闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改共享 CLI wrapper、repo-local guard、状态文档和 worklog，不动页面业务源码。
- 必须保持 `backup / ccswitch / group-create / settings-help` 共享 CLI wrapper 在宿主 Node/Playwright 失败时继续使用页面标签，而不是回退成 `channel create` 专属报错。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP live rerun。
- 不顺手改动页面 UI、后端接口或共享 CDP bootstrap 逻辑。
- 不把本轮扩大成宿主环境修复任务。

## 5. 本次验收条件

- `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-channel-create-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-05-01-phase-g-shared-cli-host-blocker-guard-alignment.md`

完成标准：共享 CLI wrapper 的 label-aware Node 失败提示、`spawn EPERM` 宿主 blocker 分类和按页面标签命名临时目录三者都进入静态守门，且 `screenshot` no-browser 聚合入口继续通过。

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-05-01-phase-g-shared-cli-host-blocker-guard-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享 CLI wrapper 报错语义与 guard 契约
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 browser smoke wrapper 守门与状态记录
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只会让共享 CLI wrapper 更早暴露宿主 blocker 语义漂移，不影响真实 smoke 运行路径

## 8. 实施步骤

1. 复核 automation memory、当前主线文档、状态文档与最近几份 Phase G worklog，确认共享 CLI wrapper 最近虽然已补上 `spawn EPERM` host blocker 分类与按标签命名临时目录，但 `verify-browser-smoke-wrapper-alignment.mjs` 还没有锁住这组契约，同时共享根脚本里仍残留一处 `channel create smoke` 硬编码。
2. 用 `apply_patch` 最小修改 `scripts/verify-channel-create-browser-smoke.ps1`，把 Node 解析失败提示改成跟随 `$smokeLabel` 输出当前页面标签。
3. 同步升级 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，新增对 label-aware Node 失败提示、`spawn EPERM` host-blocker 文案，以及 `octopus-<label>-smoke-*` 临时目录命名的断言。
4. 运行共享 guard、`backup` check-only、`screenshot` no-browser 聚合入口与 `git diff --check`，确认契约收口生效，再同步状态文档和本轮 worklog。

## 9. 测试与验证

- 构建命令：未涉及构建
- 测试命令：
  - `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
  - `node .\scripts\run-frontend-verification-suite.mjs screenshot`
  - `git diff --check -- scripts\verify-channel-create-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-05-01-phase-g-shared-cli-host-blocker-guard-alignment.md`
- 专项验证：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`

## 10. 风险与兼容性

- 新风险：低；只改共享 CLI wrapper 的错误提示与静态守门。
- 兼容性风险：低；不会影响 shared CLI smoke 的运行路径，只会更早暴露错误标签或临时目录命名回退。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据与宿主 blocker 交接。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke reliability 主线，不能切回页面业务。
  - 相邻 Phase G worklog：确认上一轮已经补齐 shared CDP usage 与 `settings-help` 默认 bootstrap 守门，因此当前最合理的连续任务就是继续锁住共享 CLI wrapper 的宿主 blocker 契约。
  - automation memory：明确上一轮已决定继续做 repo-local drift closure，不碰 live rerun。
  - `runtime-win.ps1 -Action status`：确认当前主机仍处于低权限 fallback 探测模式，且 `3000 / 8080` 存在监听但无法解析 owning process；这更说明本轮应继续停留在 repo-local 守门，而不是在同一宿主上扩张 live smoke 范围。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 repo-local no-browser / check-only 验证。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；当前主机的运行态探测仍是 low-privilege fallback，且已记录既有 attached-session / Playwright CLI 宿主 blocker。
- 待验证页面清单：`backup`、`ccswitch`、`group-create`、`settings-help`、`home-layout`、`model-layout`、`channel-page`、`ai-learning` 的真实 browser/CDP 证据与 host blocker 交接。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务范围很小。
- worklog 是否更新：yes
- 遗留项：
  - 本轮只补了共享 CLI wrapper 的静态守门，没有新增 live browser pass。
  - `runtime-win.ps1 -Action status` 当前仍显示 `3000 / 8080` 有监听但低权限下无法判定 owning process；若下一轮需要 live 取证，应先确认这些监听是否属于当前项目再决定是否 external/self-start。
  - 若后续继续改共享 CLI wrapper 的报错文案、临时目录或 host-blocker 分类，需要在同一轮同步更新 `verify-browser-smoke-wrapper-alignment.mjs`。
- 下一任务前置条件是否满足：满足。

## 12. 执行与结果

1. 复核当前主线文档、automation memory 与相邻 Phase G worklog 后确认：共享 CLI wrapper 最近已补上 `spawn EPERM` host blocker 分类和按页面标签命名临时目录，但 `verify-browser-smoke-wrapper-alignment.mjs` 还没有把这组契约锁住；同时共享根脚本里 Node 解析失败提示仍残留 `channel create smoke` 硬编码。
2. 因此本轮没有继续碰任何页面或 live smoke，只做了更小的一层共享契约收口：
   - 把 `verify-channel-create-browser-smoke.ps1` 的 Node 解析失败提示改成跟随 `$smokeLabel`；
   - 在 `verify-browser-smoke-wrapper-alignment.mjs` 中补上对 label-aware Node 失败提示、`spawn EPERM` 宿主 blocker 文案与 `octopus-<label>-smoke-*` 临时目录命名的断言。
3. 验证结果显示：guard、`backup` check-only 与 `screenshot` 聚合入口都继续通过，说明这条共享 CLI wrapper 语义已经进入统一 no-browser 守门，而不会继续依赖人工记忆去发现回退。

## 13. 验证

- passed `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-channel-create-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-05-01-phase-g-shared-cli-host-blocker-guard-alignment.md`
