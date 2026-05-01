# 2026-05-01 Phase G Runtime Guard And Active Doc Entry Alignment

## 1. 任务信息

- 任务名称：运行态守门与活跃文档入口对齐收口
- 日期：2026-05-01
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性` 的验证/交接支撑层

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收；`1.2 / 1.3` 开工收工固定动作
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-phase-g-runtime-status-netstat-owner-handoff.md`
  - `docs/archive/worklog/worklog/2026-05-01-phase-g-shared-cli-host-blocker-guard-alignment.md`
- 本次任务目标：把上一轮已经稳定的 `runtime-win.ps1` 低权限输出 contract 与活跃状态文档入口路径正式纳入 repo-local 守门，同时修掉统一 screenshot no-browser 入口里暴露出来的一处 `ai-learning-focus` 旧断言漂移。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `scripts/runtime-win.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - `scripts/verify-ai-automation-learning-focus.mjs`
  - `web/src/components/modules/ai-automation/index.tsx`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已把 `runtime-win.ps1` 的 low-privilege owner hint 收口完，本轮最连续的小任务应是把这层 support contract 变成静态守门，而不是回头改 live smoke。
  - `using-superpowers`：按会话要求先核对技能边界。
  - `brainstorming`：仅作流程边界核对；本轮属于既有 Phase G 支撑脚本/文档入口守门补丁，不进入设计审批流。
- 若未使用部分本地资源或上下文，原因：本轮不涉及页面业务逻辑、后端接口或 live browser/CDP 取证。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、上下文耦合强，主线程直接闭环更稳。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改活跃状态文档、repo-local guard 脚本和相邻 no-browser verifier，不改页面业务语义、不做 live rerun。
- 必须把“上一轮已经口头稳定的 support contract”变成统一入口可自动回归的静态断言。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP rerun。
- 不顺手改 `runtime-win.ps1` 行为语义，只锁现有 contract。
- 不把本轮扩大成全仓历史文档清洗，只修活跃入口和当前验证链阻塞项。

## 5. 本次验收条件

- `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\verify-ai-automation-learning-focus.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs scripts\verify-ai-automation-learning-focus.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\ENV_READY_AND_NEXT_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs\archive\worklog\worklog\2026-05-01-phase-g-runtime-guard-and-active-doc-entry-alignment.md`

完成标准：活跃状态文档不再回指不存在的根目录文档；`runtime-win.ps1` 关键 handoff contract 进入 guard；统一 screenshot no-browser 入口重新跑绿。

## 6. 本次回滚点

- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/archive/worklog/worklog/2026-05-01-phase-g-runtime-guard-and-active-doc-entry-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改活跃文档入口与静态 guard，再修统一验证链里暴露出的旧断言
- 受影响后端模块：无
- 受影响前端模块：无页面业务语义改动；仅影响 no-browser verifier 与状态文档
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只会更早暴露文档入口/guard 漂移，不影响实际 smoke 执行路径

## 8. 实施步骤

1. 复核自动化记忆、主线文档和工作区状态，确认当前最有连续性的任务不是再扩某个 page-level wrapper，而是把“活跃文档入口”与“runtime support contract”这两处 repo-local drift 收口成守门。
2. 检查活跃状态文档里的 canonical/status 链接是否指向真实文件，发现它们仍回指不存在的 `docs/*.zh-CN.md` 根路径，因此先统一改到 `docs/archive/planning/` 与 `docs/archive/status/`。
3. 扩展 `scripts/verify-browser-smoke-wrapper-alignment.mjs`：
   - 新增对 `scripts/runtime-win.ps1` 的关键输出 contract 断言；
   - 新增对活跃状态/workflow 文档路径的 include/exclude 守门，防止再次回指根目录旧路径。
4. 运行 guard 与统一 screenshot no-browser 入口时，发现 `scripts/verify-ai-automation-learning-focus.mjs` 仍要求旧的 `ref={learningSectionRef}` 写法；结合当前 `web/src/components/modules/ai-automation/index.tsx` 的实现，更新为 `bindWorkbenchSection('learning')` 绑定链断言。
5. 重新跑 `runtime-win.ps1 status/check-only`、guard、`ai-learning-focus` 与 `run-frontend-verification-suite.mjs screenshot`，确认整条 repo-local 链恢复全绿，再同步状态文档和本 worklog。

## 9. 测试与验证

- 构建命令：未涉及构建
- 测试命令：
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only`
  - `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
  - `node .\scripts\verify-ai-automation-learning-focus.mjs`
  - `node .\scripts\run-frontend-verification-suite.mjs screenshot`
  - `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs scripts\verify-ai-automation-learning-focus.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\ENV_READY_AND_NEXT_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs\archive\worklog\worklog\2026-05-01-phase-g-runtime-guard-and-active-doc-entry-alignment.md`

## 10. 风险与兼容性

- 新风险：低；只收口状态文档入口和静态 verifier。
- 兼容性风险：低；运行态脚本与页面业务语义没有改变，只把现有约定写成 guard。
- 是否阻塞下一任务：不阻塞；下一轮可以继续回到真实 browser/CDP 证据与宿主 blocker 交接。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：明确上一轮已经修好 `runtime-win.ps1` 低权限 owner hint，本轮应继续锁 support contract 而不是重复做 live host 分类。
  - 活跃状态文档：直接暴露 canonical/status 入口回指根目录旧路径的漂移。
  - `run-frontend-verification-suite.mjs screenshot`：直接暴露 `ai-learning-focus` 里旧 `learningSectionRef` 断言已落后于当前源码实现，是当前统一验证链的同主线阻塞项。
- 本次使用了哪些子 agent 及其结论：无。
- 手工 smoke 状态：未执行人工 browser/CDP 操作；本轮只跑 repo-local no-browser 与 runtime support 验证。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮没有触碰 live rerun。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮小闭环完全可由主线程串行完成。
- worklog 是否更新：yes
- 遗留项：
  - 当前已把 support-layer drift 清空到统一截图入口，但真实 browser/CDP 证据仍待下一轮继续。
  - 活跃状态文档之外的 archive-only 历史旧路径引用仍可能存在，但不属于本轮活跃入口范围。
- 下一任务前置条件是否满足：满足。

## 12. 执行与结果

1. 本轮先确认了一个直接影响接手效率的事实：活跃状态/workflow 文档虽然一直被要求“先读”，但它们里用于跳转 canonical 与状态文档的链接仍指向不存在的根目录 `docs/*.zh-CN.md`。这会让下一轮从错误入口起步，因此被视为当前主线下的有效 repo-local drift。
2. 代码层面随后做了两个收口：
   - 把活跃状态文档里的 canonical/status 链接统一改回真实 `docs/archive/planning/` 与 `docs/archive/status/` 路径；
   - 把 `runtime-win.ps1` 上一轮已收口的 low-privilege handoff contract 正式纳入 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，包括 `Port scan mode`、loopback readiness、runtime entrypoint 提示与 `Automation entrypoints` 输出。
3. 在重新跑统一 screenshot no-browser 入口时，又暴露出一处相邻守门漂移：`scripts/verify-ai-automation-learning-focus.mjs` 仍要求旧的 `ref={learningSectionRef}` JSX 写法，但当前 `AI 自动化` 页已改为 `bindWorkbenchSection('learning')` 绑定并把 ref 保存到 `learningSectionRef.current`。本轮在同一主线下顺手把断言更新到当前真实实现，避免统一入口继续卡在历史 guard 上。
4. 重新验证后，`run-frontend-verification-suite.mjs screenshot` 已恢复全绿。当前 screenshot-first no-browser 主链的剩余工作因此重新稳定收敛为真实 browser/CDP 证据与宿主 blocker，而不是活跃文档入口或 verifier 自身漂移。

## 13. 验证

- passed `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- passed `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action check-only`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\verify-ai-automation-learning-focus.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs scripts\verify-ai-automation-learning-focus.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\ENV_READY_AND_NEXT_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\status\DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs\archive\worklog\worklog\2026-05-01-phase-g-runtime-guard-and-active-doc-entry-alignment.md`
