# 2026-04-28 Phase H AI learning toggle runtime fallback alignment

## 1. 任务信息

- 任务名称：phase h ai learning toggle runtime fallback alignment
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习双 consumer 一致性收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-settings-learning-runtime-source-alignment.md`
- 本次任务目标：把顶层 `AI 自动化` 页的 learning 开关状态来源与 settings 卡收口到同一条“草稿态 -> 运行时 learning 查询 -> persisted config”的回退链，减少 learning query 短空窗里的闪烁或错误关闭态
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN、DYNAMIC_ROUTING_IMPLEMENTATION_PLAN、Phase H6 连续 worklog、`web/src/components/modules/ai-automation/index.tsx`、`web/src/components/modules/ai-automation/index.test.tsx`、`scripts/verify-ai-automation-learning-focus.mjs`、`web/public/locale/*.json`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 动态路由学习 consumer 收口，不扩散到新任务类型、后端 schema 或新的 profile 能力
- `AI 自动化` 页和 settings 卡必须共享同一运行时学习状态解释，不能一个用 runtime learning、一个只用 persisted config
- 必须同步补 no-browser 守护、locale 和 worklog，避免只改实现不补回归护栏

## 4. 本次禁止事项

- 不改 `dynamic_routing_learning_enabled` 的后端含义
- 不新增接口或改动 `dynamic-routing/learning` 返回结构
- 不因宿主 `vitest/esbuild spawn EPERM` 删除现有 jsdom 测试入口

## 5. 本次验收条件

- `AI 自动化` 页 learning 开关状态优先采用草稿态，其次 runtime learning 查询，最后回退 `ai/config.dynamic_routing_learning_enabled`
- learning 开关保存失败时，页面要回退草稿态并显示专用失败 toast
- `verify-ai-automation-learning-focus.mjs` 与 `index.test.tsx` 同步覆盖新的回退链
- `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs`、`verify-locale-consistency.mjs` 通过

## 6. 本次回滚点

- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/components/modules/ai-automation/index.test.tsx`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/en.json`
- `web/public/locale/ja.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先收口 `AI 自动化` learning toggle 的状态来源和失败回退，再补测试和 locale
- 受影响后端模块：无
- 受影响前端模块：`AI 自动化` learning 开关、学习状态摘要、对应 no-browser 脚本与组件测试桩
- 受影响接口：无新增接口；继续复用 `GET /api/v1/ai/config` 与 `GET /api/v1/dynamic-routing/learning`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只减少查询短空窗里的错误状态闪烁，并补上失败反馈

## 8. 实施步骤

1. 复核 `AI 自动化` 页与 settings 卡的 learning 状态来源，确认 settings 已有草稿态和 runtime 回退链，而 `AI 自动化` 页仍只依赖 `learning.data?.enabled`。
2. 在 `web/src/components/modules/ai-automation/index.tsx` 中新增 `learningEnabledDraft`，把 learning 开关状态改成“草稿态 -> runtime learning -> persisted config”的顺序。
3. 在同一处为 learning toggle 保存失败补上草稿回退和专用 toast，避免失败时 UI 停留在错误状态。
4. 更新 `index.test.tsx` 和 `verify-ai-automation-learning-focus.mjs`，再补四语 locale 的 `learningSaveFailed` 文案。

## 9. 测试与验证

- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-dynamic-routing-help.mjs`
- 通过：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-locale-consistency.mjs`
- 通过：`git diff --check -- web/src/components/modules/ai-automation/index.tsx web/src/components/modules/ai-automation/index.test.tsx scripts/verify-ai-automation-learning-focus.mjs web/public/locale/zh-Hans.json web/public/locale/zh-Hant.json web/public/locale/en.json web/public/locale/ja.json`
- 未执行成功：`. .\scripts\use-node-env.ps1; & $env:NODEEXE -r .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run .\web\src\components\modules\ai-automation\index.test.tsx --config .\web\vitest.config.ts` 仍受宿主 `vite/esbuild spawn EPERM` 阻塞

## 10. 风险与兼容性

- 新风险：低；learning toggle 现在多了一层 persisted config 回退，但其语义与 settings 卡保持一致
- 兼容性风险：低；不改接口结构，只收口已有状态源优先级并补失败提示
- 是否阻塞下一任务：不阻塞；下一轮可继续同一 Phase H6 主线，补浏览器级证据或继续压缩 settings / `AI 自动化` 学习区的剩余 consumer 差异

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、`verify-dynamic-routing-help.mjs`、`verify-locale-consistency.mjs` 与 `git diff --check`；真实 `vitest` 仍受宿主 `spawn EPERM` 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 no-browser 守护脚本
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：浏览器级与 jsdom 入口仍受宿主环境阻塞
- worklog 是否更新：是
- 遗留项：浏览器级 `375px / hover / focus` 证据仍未补齐；`vitest/esbuild spawn EPERM` 仍未解除
- 下一任务前置条件是否满足：满足；下一轮优先继续同一 Phase H6 主线，补浏览器级证据，或继续收口学习入口在失败态/空窗态下的细节一致性
