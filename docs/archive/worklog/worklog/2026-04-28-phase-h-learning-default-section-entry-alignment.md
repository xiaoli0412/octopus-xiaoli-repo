# 2026-04-28 Phase H Learning Default Section Entry 收口

## 1. 任务信息

- 任务名称：phase h learning default section entry alignment
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习 consumer 一致性收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` Phase H / 每轮只选一个最小可验证闭环
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-learning-summary-section-builder-alignment.md`
- 本次任务目标：继续沿 Phase H6 主线，把顶层 `AI 自动化` 学习区的摘要 section 组装也并入共享 helper，并让 settings 学习卡复用同一套默认 section entry 定义，进一步压缩两个 consumer 之间残留的重复分组逻辑
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、DETAILED_EXECUTION_WORKFLOW、USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW、Phase H6 连续 worklog、`learning-summary.ts`、`ai-automation/index.tsx`、`setting/DynamicRouting.tsx`、两条 no-browser 守护脚本
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 no-browser 守护脚本
- 若未使用部分本地资源或上下文，原因：浏览器级 smoke 与真实 jsdom 入口仍受宿主阻塞，本轮继续聚焦 H6 同主线下最小的共享展示收口
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning summary consumer 收口，不扩散到 sample card 共享、后端 schema、AI task 新能力或新的浏览器诊断主题
- 不改 `dynamic_routing_learning_enabled` 的保存语义、toast 语义、test id 与 settings/AI 自动化页既有信息分层
- settings 继续保持“轻摘要 + 轻动作”，`AI 自动化` 页继续承接样本卡与主工作台，不为了抽象强行合并页面职责

## 4. 本次禁止事项

- 不改 learning 样本卡列表结构
- 不改 `useDynamicRouteLearning` / `useResetDynamicRouteLearning` 接口契约
- 不把宿主 `vitest/esbuild spawn EPERM` 与 loopback 问题误记为产品回归

## 5. 本次验收条件

- `learning-summary.ts` 提供默认摘要分组常量，作为两个 consumer 的共享 base section 定义
- 顶层 `AI 自动化` 学习区通过 `buildLearningSummarySections` 组装 `primary / secondary` grid，而不是直接把 `items` 整包塞进单个 grid
- settings 学习卡在插入 `scope / details` 差异字段时复用同一套默认 section entry 定义
- `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs`、`verify-locale-consistency.mjs` 与 `git diff --check` 通过

## 6. 本次回滚点

- `web/src/components/modules/ai-automation/learning-summary.ts`
- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/components/modules/setting/DynamicRouting.tsx`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `scripts/verify-dynamic-routing-help.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先扩展共享 helper 常量，再改顶层 `AI 自动化` consumer，最后把 settings 侧切到复用默认 section entry，并同步守护脚本
- 受影响后端模块：无
- 受影响前端模块：`learning-summary.ts`、顶层 `AI 自动化` 学习摘要、settings 动态路由学习卡、两条 no-browser 守护
- 受影响接口：无新增接口
- 是否影响旧数据：否
- 是否影响旧行为：低风险；摘要字段、文案、按钮与样本列表行为保持不变，只统一了摘要 section 的分组来源与布局装配方式

## 8. 实施步骤

1. 复核 `ai-automation/index.tsx` 当前仍直接把 `learningSummaryView.items` 交给单 grid 渲染，而 settings 已进入 section builder 阶段，确认这是本轮最小且连续的 H6 收口点。
2. 在 `learning-summary.ts` 中新增 `LearningSummarySectionEntries` 与 `DEFAULT_LEARNING_SUMMARY_SECTION_ENTRIES`，把默认摘要分组定义为共享入口。
3. 更新顶层 `AI 自动化` 学习区，改为通过 `buildLearningSummarySections` 输出 `primary / secondary` 两组摘要卡，并保持样本卡列表与 footer action bar 不变。
4. 更新 `DynamicRouting.tsx`，改为在默认 section entry 基础上插入 settings 专属的 `scope / details` 字段，减少两处重复字面量 key。
5. 同步更新 `verify-ai-automation-learning-focus.mjs` 与 `verify-dynamic-routing-help.mjs`，把新的共享常量接线方式纳入静态回归。

## 9. 测试与验证

- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-dynamic-routing-help.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-locale-consistency.mjs`
- 通过：`git diff --check -- web/src/components/modules/ai-automation/learning-summary.ts web/src/components/modules/ai-automation/index.tsx web/src/components/modules/setting/DynamicRouting.tsx scripts/verify-ai-automation-learning-focus.mjs scripts/verify-dynamic-routing-help.mjs docs/worklog/2026-04-28-phase-h-learning-default-section-entry-alignment.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 中途现象：首次并行验证时，`verify-dynamic-routing-help.mjs` 仍残留旧字面量 key 断言，已在本轮同步修正后恢复通过
- 中途现象：首次并行验证时 `tsc` 和 `verify-locale-consistency.mjs` 一度报 locale JSON 解析异常；复核四语 locale 文件后确认当前文件可被 Node 正常解析，重跑后恢复通过，未继续归因为产品回归
- 未执行成功：真实 `vitest` / jsdom 入口，仍受宿主 `vite/esbuild spawn EPERM` 阻塞

## 10. 风险与兼容性

- 新风险：低；若后续默认摘要项增删，两个 consumer 将更一致地跟随共享常量变化，减少一处改了另一处漏改的风险
- 兼容性风险：低；未改变 locale key、摘要文案、动作语义或样本卡展示
- 是否阻塞下一任务：不阻塞；下一轮仍优先继续 Phase H6 浏览器级 `375px / hover / focus` 证据，若宿主继续阻塞，则继续在同页寻找不越界的小收口点

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs`、`verify-locale-consistency.mjs` 与 `git diff --check`
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、DETAILED_EXECUTION_WORKFLOW、USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 no-browser 守护脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与连续 worklog 明确上一轮已完成 settings section builder，因此本轮最值得补的是顶层 `AI 自动化` 学习区仍残留的本地摘要分组装配；canonical plan 与 workflow 限定本轮不得越界；状态文档确认当前主阻塞仍是 browser/jsdom 宿主，而不是 H6 主线断链；现有 no-browser 守护提供本轮可复用的验证路径
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：浏览器级与真实 jsdom 入口仍受宿主 `spawn EPERM` / loopback 初始化问题影响
- 待验证页面清单：`AI 自动化` 学习区浏览器级 `375px / hover / focus`、settings 动态路由学习卡浏览器级 `375px / hover / focus`
- worklog 是否更新：是
- 遗留项：浏览器级 `375px / hover / focus` 证据仍未补齐；`vitest/esbuild spawn EPERM` 仍未解除；sample card 结构继续保持刻意不共享，避免扩大边界
- 下一任务前置条件是否满足：满足；下一轮优先继续 Phase H6 浏览器级证据，若宿主仍阻塞，则继续在同页寻找不越界的小收口点，而不是切回泛化分析
