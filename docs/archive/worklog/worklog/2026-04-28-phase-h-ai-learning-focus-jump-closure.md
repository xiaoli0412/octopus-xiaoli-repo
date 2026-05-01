# 2026-04-28 Phase H AI 学习区跳转闭环收口

## 1. 任务信息

- 任务名称：phase h ai learning focus jump closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习 consumer 行为收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-setting-dynamic-routing-learning-entry-closure.md`
- 本次任务目标：把设置页 `动态路由学习` 轻入口补成真实的“打开 AI 自动化后聚焦到学习区”的最小行为闭环，并接入现有 no-browser 验证链
- 本次已盘点本地资源：AGENTS.md、canonical plan、USER_CONTEXT_REQUIREMENTS、DETAILED_EXECUTION_WORKFLOW、CURRENT_STATUS_AND_PLAN、上一轮 Phase H worklog、`web/src/components/modules/setting/DynamicRouting.tsx`、`web/src/components/modules/ai-automation/index.tsx`、`web/package.json`
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、Phase H 连续 worklog、现有 no-browser 验证脚本、`using-superpowers` / `brainstorming` skill 文档
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求本自动化主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 相邻 consumer 收口，不扩散到新的后端能力或新的 AI 页面模块
- 不改变动态路由学习的运行时语义，只补页面间跳转与聚焦契约
- 必须同步补 no-browser 验证和脚本入口，避免形成“只有 UI 代码、没有回归护栏”的假闭环

## 4. 本次禁止事项

- 不改 `dynamic_routing_learning_enabled` 的保存语义
- 不把重型学习样本列表重新搬回 settings
- 不因为宿主 `vitest/esbuild spawn EPERM` 去删除现有测试文件或绕过验证记录

## 5. 本次验收条件

- 设置页动态路由学习入口不再只是 `setActiveItem('ai')`，而是会显式排队一个 `learning` 聚焦目标
- `AI 自动化` 页面在载入时能消费该聚焦目标，并把学习区滚动到可见位置
- 新增 repo-local no-browser 验证脚本，并接入 `web/package.json` 的 `test:settings-no-browser`
- `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs`、`verify-ai-config-profile-summary.mjs` 通过

## 6. 本次回滚点

- `web/src/components/modules/ai-automation/focus-target.ts`
- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/components/modules/setting/DynamicRouting.test.tsx`
- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/components/modules/ai-automation/index.test.tsx`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `web/package.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先补前端 consumer 之间的跳转契约，再补 no-browser 验证和脚本入口
- 受影响后端模块：无
- 受影响前端模块：`DynamicRouting` 设置卡片、`AI Automation` 学习区、相关测试
- 受影响接口：无新增接口
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只把“跳到 AI 自动化页面”细化成“跳到 AI 自动化页面的学习区”

## 8. 实施步骤

1. 复核上一轮 H6 worklog，确认当前缺口是“入口存在，但还没有明确带用户落到学习区”。
2. 新增轻量 `focus-target` helper，用 `sessionStorage` 排队一次性 `learning` 聚焦目标，避免把临时跳转状态塞进持久导航 store。
3. 设置页点击 `动态路由学习` 入口时先写入聚焦目标，再切到 `AI 自动化` 页面。
4. `AI 自动化` 页面挂接学习区锚点，在首次载入时消费聚焦目标并执行 `scrollIntoView + focus`。
5. 新增 `verify-ai-automation-learning-focus.mjs`，并把它并入 `web/package.json` 的 `test:settings-no-browser` 聚合入口。

## 9. 测试与验证

- 已执行：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 已执行：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 已执行：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-dynamic-routing-help.mjs`
- 已执行：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-config-profile-summary.mjs`
- 已执行：`git diff --check -- web/package.json web/src/components/modules/ai-automation/focus-target.ts web/src/components/modules/setting/DynamicRouting.tsx web/src/components/modules/setting/DynamicRouting.test.tsx web/src/components/modules/ai-automation/index.tsx web/src/components/modules/ai-automation/index.test.tsx scripts/verify-ai-automation-learning-focus.mjs`
- 已确认阻塞：`. .\scripts\use-node-env.ps1; & $env:NODEEXE -r .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run .\web\src\components\modules\setting\DynamicRouting.test.tsx .\web\src\components\modules\ai-automation\index.test.tsx --config .\web\vitest.config.ts` 仍在加载 `web/vitest.config.ts` 时因宿主 `vite/esbuild spawn EPERM` 失败

## 10. 风险与兼容性

- 新风险：低；本轮只增加一次性前端聚焦状态，不改设置接口和 AI 学习数据语义
- 兼容性风险：低；若浏览器或宿主不支持该一次性聚焦状态，行为会退化为仅跳转到 `AI 自动化` 页面，不会破坏现有主路径
- 是否阻塞下一任务：不阻塞；下一轮可以继续补浏览器级证据，或在同主线下继续收口相邻 consumer 行为

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：部分通过
  - 通过：`verify-ai-automation-learning-focus.mjs`
  - 通过：`verify-dynamic-routing-help.mjs`
  - 通过：`verify-ai-config-profile-summary.mjs`
  - 通过：`git diff --check`
  - 未通过：`vitest` 仍受宿主 `vite/esbuild spawn EPERM` 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、Phase H 连续 worklog、现有 no-browser 脚本、`using-superpowers` / `brainstorming` skill 文档
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：确认上一轮已把 settings 学习入口轻量化，本轮应继续补“入口是否真正把用户带到学习区”的相邻闭环
  - Phase H 连续 worklog：确认 `vitest/esbuild spawn EPERM` 是环境阻塞，不是本轮回归
  - 现有 no-browser 脚本：证明当前仓库已有稳定的 repo-local 合同验证模式，适合继续扩 Phase H consumer 护栏
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮只做 no-browser consumer 闭环；浏览器级 `375px / hover / focus` 仍需后续在宿主允许时补证据
- worklog 是否更新：是
- 遗留项：`DynamicRouting.test.tsx` 与 `ai-automation/index.test.tsx` 的真实 jsdom 运行仍未补上；浏览器级跳转与聚焦证据仍待收口
- 下一任务前置条件是否满足：满足；下一轮可继续同一 Phase H6 主线，优先补 settings -> AI learning 的浏览器/jsdom 证据，或继续同一主线下另一处 consumer 行为收口
