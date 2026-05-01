# 2026-04-30 Phase G AI Learning Stable Diagnostic Guard And User Context Path Alignment

## 1. 任务信息

- 任务名称：`ai-learning` specialized root stable-diagnostic guard 收口与 active 文档 `USER_CONTEXT` 路径对齐
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.4 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-ai-learning-browser-alias-guard-alignment.md`
- 本次任务目标：把 `ai-learning` specialized root 最近已落地的 stable-diagnostic freshness / timeout-comparison 合同补进静态 guard，同时修正 active 状态文档仍指向失效 `USER_CONTEXT` 路径的执行断点。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `scripts/runtime-win.ps1`
  - `scripts/verify-ai-automation-learning-browser-smoke.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已把 `ai-learning` 的浏览器参数 alias 收口，当前最小连续缺口转为 specialized root 的 stable-diagnostic guard 与 active 文档执行入口一致性。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；用户已明确要求沿既有主线直接施工，本轮不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 wrapper guard 与 active 文档入口，不涉及页面源码或真实 browser/CDP live rerun。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否。
- 若使用分目录 agent，负责目录与禁止越界范围：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、范围集中，主线程即可闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 `ai-learning` specialized root 的 stable-diagnostic preview/guard 合同与 active 状态文档中的执行入口路径，不改页面源码与 live smoke 执行流。
- 必须保持 `ai-learning` 现有 host-friendly / self-start / stable-diagnostic 行为不变，只补守门与可执行文档入口。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP rerun。
- 不顺手修改 shared root、page-level wrapper 或页面场景脚本。
- 不批量清理 archive 中所有旧路径，只修当前 active 状态/流程文档里的直接执行入口。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-ai-automation-learning-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\ENV_READY_AND_NEXT_PLAN.zh-CN.md docs\archive\status\DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-ai-learning-stable-diagnostic-guard-and-user-context-path-alignment.md`

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-ai-learning-stable-diagnostic-guard-and-user-context-path-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 smoke guard / workflow 文档入口
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper guard 与 active 文档引用
- 受影响接口：无运行态 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：只影响 `check-only` 预览输出与 repo-local guard；文档入口改为真实存在路径

## 8. 实施步骤

1. 复核 automation memory、主线文档、active 状态文档与 `ai-learning` specialized root 当前实现，确认上一轮已完成 alias 收口，本轮最小连续缺口是 stable-diagnostic guard 与 active 文档执行入口路径。
2. 用 `apply_patch` 最小修改 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 与 `scripts/verify-browser-smoke-wrapper-alignment.mjs`：
   - `check-only` 预览固定输出 freshness threshold；
   - guard 额外锁定 `StableDiagnosticFreshnessThresholdHours`、timeout-comparison helper 与 `-FreshnessThresholdHours` 透传合同。
3. 同步修正 `CURRENT_STATUS`、`FRONTEND_UI_MAINLINE_STATUS`、`ENV_READY_AND_NEXT_PLAN`、`DETAILED_EXECUTION_WORKFLOW` 中仍指向失效旧路径的 `USER_CONTEXT` 链接，并补本轮状态说明。
4. 跑 `ai-learning` `check-only`、alignment guard、screenshot no-browser 聚合入口与 `git diff --check`，确认本轮是实质 guard/入口收口而不是纯文案变更。

## 9. 测试与验证

- 构建命令：未涉及构建。
- 测试命令：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only`
  - `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
  - `node .\scripts\run-frontend-verification-suite.mjs screenshot`
  - `git diff --check -- scripts\verify-ai-automation-learning-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\ENV_READY_AND_NEXT_PLAN.zh-CN.md docs\archive\status\DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-ai-learning-stable-diagnostic-guard-and-user-context-path-alignment.md`
- 专项验证：`powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`

## 10. 风险与兼容性

- 新风险：低；只增加 specialized root `check-only` 输出和静态 guard，不改运行态执行分支。
- 兼容性风险：低；active 文档改为真实存在路径，减少下一轮按 workflow 执行时的假阻塞。
- 是否阻塞下一任务：不阻塞；下一轮仍可继续真实 browser/CDP 证据与宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。`ai-learning check-only`、wrapper alignment guard、screenshot no-browser 聚合入口与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke reliability 主线。
  - `DETAILED_EXECUTION_WORKFLOW`：确认本轮需要补齐 runtime status 与 `USER_CONTEXT` 文档入口，不应只修脚本不修 active 流程文档。
  - `USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW` 实际文件：确认旧路径已失效，active 文档入口需要改到 `docs/archive/requirements/`。
  - automation memory 与相邻 worklog：确认本轮自然承接 alias guard 收口后的下一小步，而不是重新打开新的页面主题。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 runtime status、repo-local `check-only` 与 no-browser 聚合入口。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要 live rerun。
- 待验证页面清单：下一轮优先继续 `backup`、`ccswitch`、`group-create`、`home-layout`、`model-layout`、`channel-page`、`settings-help` 与 `ai-learning` 的真实 browser/CDP 证据或宿主 blocker 分类。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、范围集中，主线程即可闭环。
- worklog 是否更新：yes
- 遗留项：当前 archive 下仍有大量历史 worklog / requirements 文本引用旧 `docs/USER_CONTEXT...` 路径，但它们不是当前 active 执行入口；如后续需要全量一致性清理，应另开同主线小任务。
- 下一任务前置条件是否满足：满足。下一轮可直接沿同一 Phase G 主线继续真实 browser/CDP 证据或宿主 blocker 分类。

## 12. 执行与结果

1. 复核 automation memory、active 主线文档、workflow 与实际存在的 `docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md` 后确认：上一轮已完成 `ai-learning` specialized root 的 `Browser/BrowserPath` alias 收口，但 static guard 还没有锁住 stable-diagnostic freshness / timeout-comparison 合同；同时 active 状态文档仍把执行入口写在已失效的旧 `docs/USER_CONTEXT...` 路径上。
2. 因此本轮没有再改任何页面或 live smoke 执行逻辑，只做两类最小补丁：
   - `verify-ai-automation-learning-browser-smoke.ps1` 的 `check-only` 预览现在会固定输出 `Stable external preflight freshness threshold (hours)`，从而把近期已经落地的 freshness 门槛显式暴露出来；
   - `verify-browser-smoke-wrapper-alignment.mjs` 也同步升级为要求 `ai-learning` specialized root 保留 `StableDiagnosticFreshnessThresholdHours`、timeout-comparison helper 与 `-FreshnessThresholdHours $StableDiagnosticFreshnessThresholdHours` 透传合同，不再只检查 alias 与 host-friendly/self-start 特例；
   - `CURRENT_STATUS`、`FRONTEND_UI_MAINLINE_STATUS`、`ENV_READY_AND_NEXT_PLAN`、`DETAILED_EXECUTION_WORKFLOW` 四份 active 文档中的 `USER_CONTEXT` 链接已统一改到真实存在的 `docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`。
3. 直接验证结果显示：
   - `ai-learning` `check-only` 现在会稳定输出 freshness threshold，并继续展示 stable diagnostic / refresh / alternate-path 预览；
   - wrapper alignment guard 与 screenshot no-browser 聚合入口继续通过，说明本轮没有打断现有 repo-local 验证编排；
   - `git diff --check` 通过，说明补丁保持了整洁文本状态。

## 13. 验证

- passed `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-ai-automation-learning-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\ENV_READY_AND_NEXT_PLAN.zh-CN.md docs\archive\status\DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-ai-learning-stable-diagnostic-guard-and-user-context-path-alignment.md`
