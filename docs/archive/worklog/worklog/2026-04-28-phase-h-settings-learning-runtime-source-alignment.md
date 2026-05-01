# 2026-04-28 Phase H settings learning runtime source alignment

## 1. 任务信息

- 任务名称：phase h settings learning runtime source alignment
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习 settings / AI Automation consumer 一致性收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-setting-learning-runtime-summary-closure.md`
- 本次任务目标：把 settings 动态路由学习卡对“当前是否参与 runtime scoring”的判断从本地 setting 视角收口到与顶层 `AI 自动化` 页一致的运行时 learning 查询视角，避免两个 consumer 在刷新窗口里出现状态不一致
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN、DYNAMIC_ROUTING_IMPLEMENTATION_PLAN、Phase H6 连续 worklog、`web/src/components/modules/setting/DynamicRouting.tsx`、`web/src/components/modules/ai-automation/index.tsx`、`web/src/api/endpoints/ai-automation.ts`、`scripts/verify-dynamic-routing-help.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 动态路由学习 consumer 收口，不扩散到后端 schema 或新的 AI task 能力
- settings 与顶层 `AI 自动化` 页必须对齐同一运行时语义，不能一个看 persisted setting、一个看 learning API
- 必须同步更新 no-browser 守护、测试桩与主线记录，避免只修实现不补护栏

## 4. 本次禁止事项

- 不改 `dynamic_routing_learning_enabled` 的持久化语义
- 不改动态路由学习 API 返回结构
- 不因宿主 `vitest/esbuild spawn EPERM` 删除现有 jsdom 测试入口

## 5. 本次验收条件

- settings 学习卡的 runtime scoring 摘要优先读取 `useDynamicRouteLearning().data.enabled`
- 切换学习开关后，settings 学习查询会主动刷新，减少与顶层 `AI 自动化` 页的显示窗口差
- `DynamicRouting.test.tsx`、`verify-dynamic-routing-help.mjs` 与相关 query invalidation 同步更新
- `tsc --noEmit`、`verify-dynamic-routing-help.mjs`、`verify-ai-automation-learning-focus.mjs`、`verify-locale-consistency.mjs` 通过

## 6. 本次回滚点

- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/components/modules/setting/DynamicRouting.test.tsx`
- `web/src/api/endpoints/ai-automation.ts`
- `scripts/verify-dynamic-routing-help.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先收口 settings 学习卡的数据来源与刷新逻辑，再补测试桩和 no-browser 守护
- 受影响后端模块：无
- 受影响前端模块：`DynamicRouting` 设置卡、动态路由学习 query mutation invalidation、对应 no-browser 守护和测试桩
- 受影响接口：无新增接口；继续复用 `GET /api/v1/dynamic-routing/learning` 与 `POST /api/v1/dynamic-routing/learning/reset`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只把 settings 卡的运行时状态判断与刷新链对齐到现有运行时事实

## 8. 实施步骤

1. 复核 Phase H6 连续 worklog 与 `DynamicRouting.tsx` / `ai-automation/index.tsx`，确认两处 consumer 对 learning enable 状态分别读取 setting 与 learning API，存在短暂不一致风险。
2. 在 `DynamicRouting.tsx` 中把 runtime scoring 摘要调整为“草稿态优先，其次运行时 learning 查询，最后回退 persisted setting”，同时在布尔 setting 保存成功后主动 `learning.refetch()`。
3. 在 `useResetDynamicRouteLearning` 中补齐对 `ai-automation/config` 与 `settings/list` 的 invalidation，确保 reset 后轻入口与顶层页更快收敛到同一状态。
4. 更新 `DynamicRouting.test.tsx` 与 `scripts/verify-dynamic-routing-help.mjs`，把新的运行时数据来源和刷新动作纳入静态回归。

## 9. 测试与验证

- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-dynamic-routing-help.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-locale-consistency.mjs`
- 通过：`git diff --check -- web/src/components/modules/setting/DynamicRouting.tsx web/src/components/modules/setting/DynamicRouting.test.tsx web/src/api/endpoints/ai-automation.ts scripts/verify-dynamic-routing-help.mjs docs/worklog/2026-04-28-phase-h-settings-learning-runtime-source-alignment.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 未执行成功：`vitest` 真实 jsdom 入口，仍受宿主 `vite/esbuild spawn EPERM` 阻塞

## 10. 风险与兼容性

- 新风险：低；settings 学习卡现在更依赖运行时 learning 查询返回，但仍保留 persisted setting 作为回退
- 兼容性风险：低；不改接口结构，只收口已有数据源顺序和刷新链
- 是否阻塞下一任务：不阻塞；下一轮可继续同一 Phase H6 主线，补浏览器级证据，或继续压缩 settings / AI Automation 之间剩余 consumer 差异

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：通过 `verify-dynamic-routing-help.mjs`、`verify-ai-automation-learning-focus.mjs`、`verify-locale-consistency.mjs` 与 `git diff --check`；真实 `vitest` 仍受宿主 `spawn EPERM` 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 no-browser 守护脚本
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：浏览器级与 jsdom 入口仍受宿主环境阻塞
- worklog 是否更新：是
- 遗留项：浏览器级 `375px / hover / focus` 证据仍未补齐；`vitest/esbuild spawn EPERM` 仍未解除
- 下一任务前置条件是否满足：满足；下一轮优先继续同一 Phase H6 主线，补浏览器级证据或继续收口 settings / AI Automation 剩余 consumer 差异
