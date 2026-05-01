# 2026-05-01 Phase G Active Archive Input Doc Entry Guard

## 1. 任务信息

- 任务名称：活跃 archive 输入文档入口守门收口
- 日期：2026-05-01
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性` 的验证/交接支撑层

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收；`1.2 / 1.3` 开工收工固定动作
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-05-01-phase-g-runtime-guard-and-active-doc-entry-alignment.md`
  - `docs/archive/worklog/worklog/2026-05-01-phase-g-runtime-status-netstat-owner-handoff.md`
- 本次任务目标：把仍被当前工作流要求“开工前先读”的 archive requirements/planning/worklog 输入文档入口，从不存在的根目录 `docs/*.md` / `docs/*.zh-CN.md` 收口到真实 `docs/archive/*` 路径，并把这批入口纳入既有 screenshot-first guard。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`
  - `docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`
  - `docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`
  - `docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已经把活跃状态/workflow 文档入口和 runtime 支撑契约收口，当前最连续的小任务就是把同样会被“开工前先读”的 archive 输入文档入口也纳入 guard，而不是重复做 live smoke。
  - `using-superpowers`：按会话要求先核对技能边界。
  - `brainstorming`：仅作流程边界核对；本轮属于既有 Phase G 支撑文档/guard 收口补丁，不进入设计审批流。
- 若未使用部分本地资源或上下文，原因：本轮不涉及页面业务逻辑、后端接口或 live browser/CDP 取证。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、上下文耦合强，主线程直接闭环更稳。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只改活跃 archive 输入文档入口、状态文档同步和相邻 guard，不改页面业务语义、不做 live rerun。
- 必须把“当前工作流要求先读的输入文档”也纳入 screenshot-first 守门，避免下一轮继续从错误入口起步。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP rerun。
- 不顺手清理 archive 历史旧文档全集，只修当前工作流直接点名的输入文档。
- 不改页面、API 或运行时脚本行为语义。

## 5. 本次验收条件

- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.md docs/archive/worklog/worklog/README.zh-CN.md scripts/verify-browser-smoke-wrapper-alignment.mjs`

完成标准：当前工作流点名的 archive 输入文档不再回指不存在的根目录文档；对应入口合同进入 `verify-browser-smoke-wrapper-alignment.mjs`；`screenshot` 聚合入口继续通过。

## 6. 本次回滚点

- `docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`
- `docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`
- `docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`
- `docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.md`
- `docs/archive/worklog/worklog/README.zh-CN.md`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/worklog/worklog/2026-05-01-phase-g-active-archive-input-doc-entry-guard.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改活跃 archive 输入文档入口，再扩 guard，最后同步状态文档和 worklog
- 受影响后端模块：无
- 受影响前端模块：无页面业务语义改动；仅影响 no-browser guard 与文档入口
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只会更早暴露输入文档入口漂移，不影响实际 smoke 执行路径

## 8. 实施步骤

1. 复核 automation memory、主线文档和相邻 worklog，确认上一轮已把活跃状态/workflow 文档入口收口完毕，本轮最连续的小任务应是把“当前工作流仍要求先读”的 archive 输入文档也纳入同一守门。
2. 定向搜索 `USER_CONTEXT_REQUIREMENTS`、动态路由 planning/requirements、AI 自动化 requirements（中英）和 `worklog/README`，确认它们仍回指不存在的根目录 `docs/*.md` / `docs/*.zh-CN.md`。
3. 统一把这些入口改到真实 `docs/archive/planning/`、`docs/archive/requirements/` 与 `docs/archive/status/` 路径，并顺手修掉 `worklog/README` 里一条仍会误导接手人的纯文本旧路径。
4. 扩展 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，新增 `activeArchiveInputDocChecks`，把这批活跃 archive 输入文档的 include/exclude 合同纳入同一条 screenshot-first guard。
5. 运行 guard 与统一 screenshot no-browser 聚合入口，确认没有因为这批入口修正引入新的 verifier 漂移，再同步状态文档与本 worklog。

## 9. 测试与验证

- 构建命令：未涉及构建
- 测试命令：
  - `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
  - `node .\scripts\run-frontend-verification-suite.mjs screenshot`
  - `git diff --check -- docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.md docs/archive/worklog/worklog/README.zh-CN.md scripts/verify-browser-smoke-wrapper-alignment.mjs`

## 10. 风险与兼容性

- 新风险：低；只收口活跃输入文档入口与静态 guard。
- 兼容性风险：低；运行时脚本与页面业务语义未变，只把现有入口约定写成守门。
- 是否阻塞下一任务：不阻塞；下一轮可继续回到真实 browser/CDP 证据与宿主 blocker 交接。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：明确上一轮已经修好活跃状态/workflow 入口，本轮最合理的连续任务是把同级 archive 输入文档也纳入 guard，而不是再回头重复状态层修补或切 live smoke。
  - archive 输入文档与 worklog README：直接暴露“仍被要求开工前先读，但入口仍指向不存在根目录文档”的接手风险。
  - `verify-browser-smoke-wrapper-alignment.mjs`：提供现成 screenshot-first 守门载体，适合把这批入口合同并入，而不需要新造一套 verifier。
- 本次使用了哪些子 agent 及其结论：无。
- 手工 smoke 状态：未执行人工 browser/CDP 操作；本轮只跑 repo-local no-browser guard 与截图聚合。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮没有触碰 live rerun。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮小闭环完全可由主线程串行完成。
- worklog 是否更新：yes
- 遗留项：
  - 活跃 archive 输入文档入口已收口，但 archive-only 历史旧文档仍可能保留旧路径，不属于本轮范围。
  - 当前主线剩余工作仍是健康宿主上的真实 browser/CDP 证据与宿主 blocker 交接。
- 下一任务前置条件是否满足：满足。

## 12. 执行与结果

1. 本轮先确认了一个与上一轮同性质、但仍会直接影响接手效率的 repo-local drift：虽然活跃状态/workflow 文档入口已经修正，但当前工作流仍要求“开工前先读”的 archive requirements/planning/worklog 输入文档中，仍保留了一批指向不存在根目录 `docs/*.md` / `docs/*.zh-CN.md` 的旧入口。
2. 代码与文档层面随后做了两组收口：
   - 把 `USER_CONTEXT_REQUIREMENTS`、动态路由 planning/requirements、AI 自动化 requirements（中英）以及 `worklog/README` 的活跃入口统一改回真实 `docs/archive/planning/`、`docs/archive/requirements/` 与 `docs/archive/status/` 路径；
   - 扩展 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，新增 `activeArchiveInputDocChecks`，把这批文档的 include/exclude 入口合同纳入同一条 screenshot-first 守门。
3. 定向回扫时还发现 `worklog/README` 中保留了一条非链接的纯文本旧路径 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`，本轮一并改成 `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`，并把这类纯文本路径也纳入 `README` 的 guard 断言，避免下一轮继续被误导。
4. 重新验证后，`verify-browser-smoke-wrapper-alignment.mjs` 与 `run-frontend-verification-suite.mjs screenshot` 都继续通过，说明当前 screenshot-first no-browser 接手入口已经从“只守活跃状态文档”扩展到“活跃状态文档 + 活跃 archive 输入文档”同一层。主线剩余工作因此继续稳定收敛为真实 browser/CDP 证据与宿主 blocker，而不是 repo-local 文档入口再漂移。

## 13. 验证

- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.md docs/archive/worklog/worklog/README.zh-CN.md scripts/verify-browser-smoke-wrapper-alignment.mjs`
