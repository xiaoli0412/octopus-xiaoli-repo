# 2026-04-28 Phase H 设置页动态路由学习轻入口收口

## 1. 任务信息

- 任务名称：phase h setting dynamic routing learning entry closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习设置页 consumer 收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 动态路由 AI 学习不改写用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-27-phase-h-ai-config-no-browser-contract-closure.md`
- 本次任务目标：把设置页动态路由区域收口成“轻量学习开关 + 摘要说明 + 跳转 AI 自动化中心”的最小闭环，消除文档与前端事实不一致
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、上一轮 Phase H worklog、`web/src/components/modules/setting/DynamicRouting.tsx`、`web/src/components/modules/setting/DynamicRouting.test.tsx`、`web/src/components/modules/ai-automation/index.tsx`、`scripts/verify-dynamic-routing-help.mjs`、四语 locale
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、主规划、前端主线状态、当前状态文档、最近 Phase H worklog、`using-superpowers` / `brainstorming` skill 文档、现有 no-browser 验证脚本
- 若未使用部分本地资源或上下文，原因：本轮只聚焦设置页动态路由学习入口，不扩散到 AI task、backup、browser smoke 或后端 schema
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求本自动化主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 动态路由 AI 学习主线，不扩散到无关 UI 或新的后端能力
- 设置页只补轻入口，不把学习样本重列表和重置流量重新堆回 settings
- 必须同步补类型检查、locale 与 no-browser 护栏，避免“只有 UI、没有验证”的假闭环

## 4. 本次禁止事项

- 不改动态路由学习的运行时语义
- 不扩散到 AI 自动化整页重构或 settings 其它卡片
- 不因为宿主 `vitest/esbuild spawn EPERM` 去删除现有测试或绕过 repo-local 验证链

## 5. 本次验收条件

- 设置页动态路由卡片出现 `dynamic_routing_learning_enabled` 轻量开关和边界说明
- 设置页明确说明学习只影响运行时推荐，并提供前往 `AI 自动化` 页面查看明细的入口
- `DynamicRouting` 测试、locale 与 no-browser 校验同步更新
- `tsc --noEmit`、`verify-dynamic-routing-help.mjs` 与 `verify-locale-consistency.mjs` 通过

## 6. 本次回滚点

- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/components/modules/setting/DynamicRouting.test.tsx`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/ja.json`
- `scripts/verify-dynamic-routing-help.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改设置页 UI consumer 和验证链
- 受影响后端模块：无
- 受影响前端模块：`DynamicRouting` 设置卡片、对应测试、四语 locale
- 受影响接口：无新增接口；继续复用已有 `setting/list` 与导航切换
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强设置页动态路由学习入口与说明，不改变后端学习评分逻辑

## 8. 实施步骤

1. 复核主线文档、前端主线状态与当前实现，确认真实缺口是“文档要求 settings 有学习开关与入口，但实际仍主要停留在 AI 自动化页”。
2. 在 `SettingDynamicRouting` 中补上学习开关、摘要说明和跳转 AI 自动化中心的轻入口，保持 settings 轻量边界。
3. 同步更新 `DynamicRouting.test.tsx`、四语 locale 和 `verify-dynamic-routing-help.mjs`，并执行定向验证。

## 9. 测试与验证

- 构建命令：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-dynamic-routing-help.mjs`
- 专项验证：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-locale-consistency.mjs`；尝试 `vitest` 运行 `DynamicRouting.test.tsx`，仍在加载 `web/vitest.config.ts` 时因宿主 `vite/esbuild spawn EPERM` 失败

## 10. 风险与兼容性

- 新风险：低；本轮只补 settings 轻入口与说明，不改变 API contract 或运行时评分
- 兼容性风险：低；仍复用已有 `dynamic_routing_learning_enabled` 设置键和导航 store
- 是否阻塞下一任务：不阻塞；但若下一轮要补 `DynamicRouting.test.tsx` 的真实 jsdom 运行，需要继续处理宿主 `vitest/esbuild spawn EPERM`

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：部分通过
  - 通过：`verify-dynamic-routing-help.mjs`
  - 通过：`verify-locale-consistency.mjs`
  - 通过：`git diff --check`
  - 未通过：`vitest` 仍在加载 `web/vitest.config.ts` 时受宿主 `vite/esbuild spawn EPERM` 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、前端主线状态、当前状态文档、最近 Phase H worklog、现有 no-browser 脚本、`using-superpowers` / `brainstorming` skill 文档
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：确认上一轮已把 AI config/profile contract 与 no-browser 验证链收口，本轮适合补 settings 相邻 consumer 缺口
  - 前端主线状态：明确 settings 需要新增 `dynamic_routing_learning_enabled` 开关和学习状态入口
  - 现有 no-browser 脚本：证明当前仓库对 settings 卡片已有稳定静态验证模式，可直接复用
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮属设置页轻入口收口，未启动浏览器；更高层 `vitest` / 浏览器级验证仍受宿主 `spawn EPERM` 与 CDP 问题影响
- 待验证页面清单：设置页动态路由卡片 `375px`、帮助提示 hover/focus、跳转 `AI 自动化` 页后的学习区域可达性
- 若未使用子 agent，原因：用户要求本自动化主线程串行执行，不创建子 agent
- worklog 是否更新：是
- 遗留项：`DynamicRouting.test.tsx` 真实运行仍未补上；浏览器级 `375px / hover / focus` 证据仍待后续同池任务处理
- 下一任务前置条件是否满足：满足；下一轮可继续同一主线，优先在宿主允许时补 settings 动态路由学习轻入口的 browser/jsdom 证据，或继续收口 AI Profile / 动态路由相邻 consumer
