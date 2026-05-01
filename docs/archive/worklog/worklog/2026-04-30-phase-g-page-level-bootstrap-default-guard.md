# 2026-04-30 Phase G Page-Level Bootstrap Default Guard

## 1. 任务信息

- 任务名称：page-level browser smoke 默认 bootstrap 契约守门
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.4 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-browser-alias-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-smoke-wrapper-alignment-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-ai-learning-browser-alias-guard-alignment.md`
- 本次任务目标：把已经写入状态文档的 page-level browser smoke 默认 CDP bootstrap 顺序也纳入 repo-local guard，避免 thin forwarder 默认值无声漂移。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - `scripts/verify-ccswitch-browser-smoke.ps1`
  - `scripts/verify-home-layout-browser-smoke.ps1`
  - `scripts/verify-model-layout-browser-smoke.ps1`
  - `scripts/verify-channel-page-browser-smoke.ps1`
  - `scripts/verify-group-create-browser-smoke-cdp.ps1`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已把 wrapper 参数面收口完毕，本轮最小连续任务应转向默认值守门，而不是回到业务页。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；本轮属于既有 Phase G 脚本契约补强，不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 guard 与状态同步，不涉及页面组件、后端逻辑或 live browser rerun。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务范围很小，主线程可直接闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 repo-local guard 与状态/worklog 记录，不改页面业务、不改共享 wrapper 运行逻辑。
- 守门必须锁定已经文档化的默认 bootstrap 契约，而不是新增新的默认值语义。

## 4. 本次禁止事项

- 不回头修改 page-level wrapper 的执行逻辑。
- 不扩散到真实 browser/CDP live rerun。
- 不顺手改动 AI 自动化或 settings/model probe 业务实现。

## 5. 本次验收条件

- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-page-level-bootstrap-default-guard.md`

完成标准：guard 能稳定锁定 `ccswitch` 与其余 page-level wrapper 的默认 bootstrap 组合；直接相关 `check-only`、`screenshot` no-browser 聚合入口和 `git diff --check` 均通过。

## 6. 本次回滚点

- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-bootstrap-default-guard.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改静态 guard
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke guard 与状态记录
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只会更早暴露 page-level thin forwarder 默认值回退

## 8. 本轮微型计划

- 当前主线：`Phase G screenshot-first UI closure / browser smoke reliability`
- 当前阶段：wrapper contract closure / static guard hardening
- 候选任务：
  - 锁定 `ccswitch` 的默认 `runtime-page-lifecycle` 顺序
  - 锁定 `home/model/channel-page/group-create-cdp` 的默认 `page-lifecycle-runtime`
  - 扩展状态文档，说明 guard 已覆盖默认值而不只是 alias/透传
- 本轮核心任务：为 page-level wrapper 家族补默认 bootstrap 契约守门
- 本轮配套任务：同步状态与 worklog 记录
- 预期验证方式：alignment guard、直接相关 `check-only`、screenshot 聚合入口、`git diff --check`
- 完成判定：guard 能防止默认值无声漂移，且不影响当前 no-browser 验证链

## 9. 实施步骤

1. 复核当前状态文档与 page-level wrapper 代码，确认 `ccswitch` 默认 `attached-session + runtime-page-lifecycle`、其余 page-level thin forwarder 默认 `attached-session + page-lifecycle-runtime` 已经是现有文档化口径。
2. 用 `apply_patch` 最小修改 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，把这些默认值加入 `thinForwarderChecks` 的 include 断言。
3. 运行相关 `check-only`、alignment guard 与 `screenshot` no-browser 聚合入口，确认新增守门不会打断现有验证链。
4. 更新前端主线状态、总状态与本 worklog，给下一轮留下明确入口：当前剩余工作是 live browser/CDP 证据与宿主 blocker，而不是默认值漂移。

## 10. 风险与兼容性

- 新风险较低：本轮仅增加静态守门，不触碰运行逻辑。
- 兼容性风险：极低；只要页面 wrapper 默认值不被回退，guard 不会新增噪音。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据或宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。alignment guard、直接相关 `check-only`、`screenshot` no-browser 聚合入口与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `CURRENT_STATUS_AND_PLAN` / `FRONTEND_UI_MAINLINE_STATUS`：确认 `ccswitch` 的默认 `runtime-page-lifecycle` 已是文档承诺，而不是临时试验值。
  - 相邻 wrapper worklog 与 automation memory：确认参数面收口已完成，本轮最合适的是补默认值守门。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 repo-local `check-only` 与 no-browser 聚合入口。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要 live rerun。
- 待验证页面清单：下一轮优先继续 `backup / ccswitch / group-create / home-layout / model-layout / channel-page / settings-help / ai-learning` 的真实 browser/CDP 证据或宿主 blocker 交接。
- worklog 是否更新：yes
- 遗留项：
  - 当前守门只锁定已文档化默认值；如果未来故意调整默认 bootstrap 组合，必须在同一轮同步改 guard 与状态文档。
  - 这条守门只防止 thin forwarder 默认值漂移，不替代 live browser/CDP 证据。
- 下一任务前置条件是否满足：满足。下一轮可以直接沿同一 Phase G 主线继续 live 证据或宿主 blocker 分类。

## 12. 执行与结果

1. 复核当前主线状态与 page-level wrapper 代码后确认：最近几轮已经把 alias、共享引用与高级参数透传收口完毕，但 `verify-browser-smoke-wrapper-alignment.mjs` 仍没有锁定 page-level thin forwarder 的默认 `CdpPageBootstrapStrategy / CdpBootstrapCommandOrder`。这意味着如果后续有人把默认顺序改坏，repo-local guard 不会第一时间报错。
2. 因此本轮没有改任何业务页或共享 wrapper，只在 `scripts/verify-browser-smoke-wrapper-alignment.mjs` 里补上默认值断言：
   - `verify-ccswitch-browser-smoke.ps1` 必须继续保留 `attached-session + runtime-page-lifecycle`
   - `verify-group-create-browser-smoke-cdp.ps1`、`verify-home-layout-browser-smoke.ps1`、`verify-model-layout-browser-smoke.ps1`、`verify-channel-page-browser-smoke.ps1` 必须继续保留 `attached-session + page-lifecycle-runtime`
3. 验证结果显示：alignment guard、`ccswitch` `check-only -RequireExternalCdpPreflight`、以及整条 `screenshot` no-browser 聚合入口都继续通过，说明新增守门没有打断当前验证链，但已经能防止 page-level 默认 bootstrap 契约无声漂移。

## 13. 验证

- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-page-level-bootstrap-default-guard.md`
