# 2026-04-30 Phase G Browser Wrapper Inventory Guard

## 1. 任务信息

- 任务名称：browser smoke wrapper inventory 与 shared-root 契约守门
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.4 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-browser-alias-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-cli-browser-alias-alignment.md`
- 本次任务目标：把 `verify-browser-smoke-wrapper-alignment.mjs` 从“只检查若干 thin forwarder”升级为“固定 browser smoke PowerShell 家族清单 + shared root wrapper 契约守门”，避免后续新增或回退 wrapper 时无声漂移。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-page-level-browser-alias-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-cli-browser-alias-alignment.md`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - `scripts/verify-ai-automation-learning-browser-smoke.ps1`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已把 thin forwarder 参数面统一到 `Browser`/`BrowserPath`，当前最小相邻任务是给 wrapper 家族本身补更完整的 repo-local guard。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；用户已明确要求在既有主线内直接推进代码，本轮不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 wrapper 守门，不涉及页面源码、后端重构链或运行态业务契约。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、上下文强耦合，主线程串行即可闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改 repo-local 守护、状态记录与 worklog，不改页面源码、不改共享运行逻辑。
- 必须把 `verify-ai-automation-learning-browser-smoke.ps1` 明确标记为“受控特例”，而不是默认允许任何大脚本脱离共享 wrapper 漂移。

## 4. 本次禁止事项

- 不重构 `verify-channel-create-browser-smoke.ps1` / `verify-channel-create-browser-smoke-cdp.ps1` 的执行逻辑。
- 不扩散到真实 browser/CDP live path、host blocker 文案或页面 `mjs` 场景脚本。
- 不因为补 guard 而顺手调整无关 UI/前端业务文件。

## 5. 本次验收条件

- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-browser-wrapper-inventory-guard.md`

完成标准：

- guard 会固定枚举全部 `verify-*-browser-smoke*.ps1`，若家族成员新增/删除而未同步 guard 会直接失败；
- guard 除 thin forwarder 外，还会检查 shared root wrapper 与 `ai-learning` specialized root 的关键契约；
- alignment guard、screenshot no-browser 聚合入口与 `git diff --check` 均通过。

## 6. 本次回滚点

- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-wrapper-inventory-guard.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 repo-local guard
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper guard 与文档记录
- 受影响接口：无运行态 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：不影响运行态；只提升 wrapper 漂移的静态门禁强度

## 8. 本轮微型计划

- 当前主线：`Phase G screenshot-first UI closure / browser smoke reliability`
- 当前阶段：browser smoke wrapper guard 完整性收口
- 候选任务：
  - 固定 browser smoke PowerShell wrapper 清单，防止新增脚本漏进 guard
  - 把 shared root wrapper 关键契约纳入静态断言
  - 明确 `ai-learning` 是否属于受控特例而非新的独立家族
  - 复跑 screenshot no-browser 聚合入口，确认 guard 升级未打断现有验证链
- 本轮核心任务：升级 `verify-browser-smoke-wrapper-alignment.mjs`，补 inventory + shared-root contract guard
- 本轮配套任务：同步前端主线状态与 worklog，给下一轮留下“ai-learning 是受控特例”的入口说明
- 预期验证方式：alignment guard、screenshot 聚合入口、`git diff --check`
- 完成判定：上述验证均通过，且状态文档/worklog 同步完成

## 9. 实施步骤

1. 复核 automation memory、当前主线文档与最近两份 wrapper alias worklog，确认 thin forwarder 参数面已统一，当前剩余风险不在单个入口，而在“guard 只覆盖局部家族”。
2. 用 `apply_patch` 升级 `scripts/verify-browser-smoke-wrapper-alignment.mjs`：
   - 固定枚举全部 `verify-*-browser-smoke*.ps1` 清单；
   - 把 wrapper 家族分成 `thin forwarders / shared roots / specialized root(ai-learning)`；
   - 除了继续检查 thin forwarder，也额外检查 `verify-channel-create-browser-smoke.ps1`、`verify-channel-create-browser-smoke-cdp.ps1` 与 `verify-ai-automation-learning-browser-smoke.ps1` 的关键契约。
3. 运行 alignment guard 与 screenshot no-browser 聚合入口，确认 guard 升级没有打断当前 repo-local 验证链。
4. 更新前端主线状态与本轮 worklog，记录 `ai-learning` 作为受控特例的边界，给下一轮继续真实 browser/CDP 证据留入口。

## 10. 风险与兼容性

- 新风险较低：修改的是静态 guard，不影响共享 wrapper 的执行路径。
- 兼容性风险：极低；仅当后续新增/删除 browser smoke PowerShell 脚本却不更新 guard 时才会主动失败，这正是本轮希望前置暴露的漂移。
- 是否阻塞下一任务：不阻塞；下一轮可继续真实 browser/CDP 证据或宿主 blocker 交接。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。alignment guard、screenshot no-browser 聚合入口与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应留在 screenshot-first / browser smoke reliability 主线，不扩散到页面布局或业务逻辑。
  - 相邻 wrapper alias worklog：确认参数面收口已完成，下一轮最合理的小闭环是 guard 完整性，而不是继续改单个 wrapper。
  - automation memory：明确上一轮已把 thin forwarder 公开参数面收齐，本轮自然承接为“静态家族门禁补完”。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 repo-local guard 与 no-browser 聚合入口。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要重新打开 self-start / external live path。
- 待验证页面清单：下一轮优先在健康宿主上继续验证 `backup`、`ccswitch`、`group-create`、`home-layout`、`model-layout`、`channel-page` 与 `settings-help` 的真实 browser/CDP 证据。
- worklog 是否更新：yes
- 遗留项：
  - 当前 guard 明确允许 `verify-ai-automation-learning-browser-smoke.ps1` 作为 specialized root 保留 host-handoff/stable-diagnostic 逻辑，但后续若新增第二个特例，必须在同一轮写清原因并同步 guard。
  - shared root wrapper 的运行逻辑本轮未改；若后续需要调整公共参数面或 Node 解析逻辑，应另开同主线任务并重跑 alias/host-blocker 验证。
- 下一任务前置条件是否满足：满足。下一轮可直接沿同一 Phase G 主线继续真实 browser/CDP 证据与宿主 blocker 分类。

## 12. 执行与结果

1. 复核当前主线文档、automation memory 与最近两份 wrapper alias worklog 后确认：薄包装器参数面已经基本统一，当前更真实的风险是 `verify-browser-smoke-wrapper-alignment.mjs` 只守住了若干入口，但并没有把整套 browser smoke PowerShell 家族的成员关系与 shared-root 公共契约固定下来。
2. 因此本轮没有再改单个 wrapper，而是只升级 guard：
   - 先用文件清单固定当前全部 `verify-*-browser-smoke*.ps1` 成员，要求它们只能属于 `thin forwarders / shared roots / specialized root(ai-learning)` 三类；
   - 再对 `verify-channel-create-browser-smoke.ps1` 与 `verify-channel-create-browser-smoke-cdp.ps1` 补 shared-root 关键断言，确保 Codex bundled Node 过滤、环境变量驱动的 scenario/success-marker 入口，以及 `Browser -> BrowserPath` 的公共契约不会无声回退；
   - 同时把 `verify-ai-automation-learning-browser-smoke.ps1` 明确标成受控特例，要求它继续 forward 到共享 CDP wrapper，并保留当前 host-handoff / stable diagnostic 逻辑，而不是默认允许新的独立 browser smoke 栈长出来。
3. 验证结果显示：
   - 单独运行 alignment guard 通过，说明新的 inventory + root-contract 断言和当前仓库状态一致；
   - `screenshot` no-browser 聚合入口也继续通过，说明 guard 升级没有打断现有前端验证编排；
   - `git diff --check` 通过，工作区未引入格式问题。

## 13. 验证

- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- passed `git diff --check -- scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-browser-wrapper-inventory-guard.md`
