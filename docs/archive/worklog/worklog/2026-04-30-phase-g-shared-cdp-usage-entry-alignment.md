# 2026-04-30 Phase G Shared CDP Usage Entry Alignment

## 1. 任务信息

- 任务名称：共享 CDP smoke usage 入口与 `ai-learning` 守门收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-bootstrap-default-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-ai-learning-browser-alias-guard-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-ai-learning-stable-diagnostic-guard-and-user-context-path-alignment.md`
- 本次任务目标：修正共享 `verify-channel-create-browser-smoke-cdp.mjs` 仍残留的错误 usage 入口，并把这条共享入口合同补进 `ai-learning` 静态守门，避免后续接手人或 no-browser 链继续被错误脚本名带偏。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-ai-automation-learning-focus.mjs`
  - `scripts/run-frontend-verification-suite.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已经把 specialized root stable-diagnostic guard 收紧完毕，本轮最小连续任务应转向共享入口文案漂移，而不是重新改 live wrapper。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；本轮属于既有 Phase G 验证链补丁，不进入设计审批流。
- 若未使用部分本地资源或上下文，原因：本轮只处理共享 smoke 脚本 usage 文案与静态守门，不涉及页面业务、后端数据结构或 live browser rerun。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行。
- 若使用分目录 agent，负责目录与禁止越界范围：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、上下文强耦合，主线程串行即可闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改共享 CDP smoke usage 文案、`ai-learning` 静态守门和进展记录，不改页面源码、不改真实 smoke 运行逻辑。
- 必须保持共享 `ai-learning` 场景继续依附 `verify-channel-create-browser-smoke-cdp.mjs`，不能再回退成 page-specific 错误入口提示。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP live rerun。
- 不顺手改动共享 wrapper 的 preflight / bootstrap 逻辑。
- 不把本轮扩大成页面业务、AI 自动化 UI 或宿主 blocker 修复任务。

## 5. 本次验收条件

- `node .\scripts\verify-ai-automation-learning-focus.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs settings`
- `git diff --check -- scripts\verify-channel-create-browser-smoke-cdp.mjs scripts\verify-ai-automation-learning-focus.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-shared-cdp-usage-entry-alignment.md`

完成标准：共享 CDP smoke usage 明确指向真实共享入口，`ai-learning` 守门会拦住回退，且 `settings` no-browser 聚合入口继续通过。

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke-cdp.mjs`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-cdp-usage-entry-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享 smoke usage 合同
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke verifier 与状态记录
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只影响共享 smoke usage 提示与静态守门，不影响真实运行逻辑

## 8. 本轮微型计划

- 当前主线：`Phase G screenshot-first UI closure / browser smoke reliability`
- 当前阶段：shared-wrapper repo-local drift closure
- 候选任务：
  - 修正共享 `verify-channel-create-browser-smoke-cdp.mjs` 残留的错误 usage 入口
  - 把共享 usage 入口合同补进 `verify-ai-automation-learning-focus.mjs`
  - 继续把共享 `ai-learning` 场景的静态守门挂到统一 no-browser 入口
- 本轮核心任务：修共享 CDP smoke usage 入口并补 `ai-learning` 守门
- 本轮配套任务：同步状态文档与 worklog 记录
- 预期验证方式：`verify-ai-automation-learning-focus.mjs`、`settings` 聚合入口、`git diff --check`
- 完成判定：相关守门与聚合入口都通过，且记录层同步完成

## 9. 实施步骤

1. 复核 automation memory、当前主线状态与相邻 Phase G worklog，确认上一轮已把 wrapper 参数面、specialized root 与 page-level 默认 bootstrap 守门收紧，本轮最值得推进的是共享 Node smoke 入口残留文案漂移。
2. 用 `apply_patch` 最小修改 `scripts/verify-channel-create-browser-smoke-cdp.mjs`，把 `printUsage()` 中错误的 `verify-setting-help-browser-smoke-cdp.mjs` 改回真实共享入口 `verify-channel-create-browser-smoke-cdp.mjs`。
3. 同步升级 `scripts/verify-ai-automation-learning-focus.mjs`，新增对共享 usage 文案的正向断言与对旧 `setting-help` 文案的反向断言，确保 `ai-learning` 这条共享场景不会再次被错误入口名带偏。
4. 运行 `verify-ai-automation-learning-focus.mjs`、`run-frontend-verification-suite.mjs settings` 与 `git diff --check`，确认本轮补丁真正在统一 no-browser 验证链内生效。

## 10. 风险与兼容性

- 新风险：低；只改 usage 文案与静态守门。
- 兼容性风险：低；不会影响共享 CDP smoke 的运行路径，只会更早暴露错误入口提示回退。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据与宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke reliability 主线，不能发散回页面业务。
  - 状态文档与相邻 worklog：确认上一轮已把 wrapper 参数与 specialized root 契约收紧，本轮最合理的相邻动作是补共享 usage 入口漂移。
  - automation memory：明确上一轮已决定不碰 live rerun，而是继续 repo-local guard，所以本轮只做共享 smoke usage 与守门收口。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 repo-local no-browser 验证。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮无需重开 self-start/external live path。
- 待验证页面清单：`backup`、`ccswitch`、`group-create`、`settings-help`、`home-layout`、`model-layout`、`channel-page`、`ai-learning` 的真实 browser/CDP 证据与 host blocker 交接。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务范围很小。
- worklog 是否更新：yes
- 遗留项：
  - 当前守门已锁定共享 usage 文案，但没有新增 live rerun 证据；下一轮仍应优先去健康宿主补真实 browser/CDP pass。
  - 如果后续再改共享 `verify-channel-create-browser-smoke-cdp.mjs` 的 usage 或场景入口，需要在同一轮同步更新 `verify-ai-automation-learning-focus.mjs`。
- 下一任务前置条件是否满足：满足。

## 12. 执行与结果

1. 复核当前主线文档、automation memory 与最近几份 Phase G wrapper/worklog 后确认：wrapper 参数面、specialized root 与 page-level 默认 bootstrap 守门已经基本收口，但共享 `verify-channel-create-browser-smoke-cdp.mjs` 仍保留了一处复制遗留的错误 usage 入口，会把接手人错误地引到 `verify-setting-help-browser-smoke-cdp.mjs`。
2. 因此本轮没有再动任何页面或 PowerShell wrapper，只做了一个更小的共享 smoke 合同收口：
   - 把 `verify-channel-create-browser-smoke-cdp.mjs` 的 `printUsage()` 改回真实共享入口；
   - 在 `verify-ai-automation-learning-focus.mjs` 中新增正向/反向断言，锁定共享 `ai-learning` 场景必须继续指向 `verify-channel-create-browser-smoke-cdp.mjs`，不能回退成 `setting-help` 专属入口名。
3. 验证结果显示：`verify-ai-automation-learning-focus.mjs` 与 `settings` 聚合 no-browser 入口均通过，说明这条共享 usage 漂移已被 repo-local 守门收口，而不会继续等到下一轮才靠人工发现。

## 13. 验证

- passed `node .\scripts\verify-ai-automation-learning-focus.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs settings`
- passed `git diff --check -- scripts\verify-channel-create-browser-smoke-cdp.mjs scripts\verify-ai-automation-learning-focus.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-shared-cdp-usage-entry-alignment.md`
