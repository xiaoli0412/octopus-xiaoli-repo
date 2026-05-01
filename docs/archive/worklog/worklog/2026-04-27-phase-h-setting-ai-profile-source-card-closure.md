# 2026-04-27 Phase H Setting AI Profile Source Card Closure

## 1. 任务信息

- 任务名称：setting ai profile source card closure
- 日期：2026-04-27
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H5 设置页 `manual / ai_profile` 切换收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 设置来源双轨切换要求
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：
  - `docs/worklog/2026-04-24-phase-g-ai-automation-compile-unblock.md`
  - `docs/worklog/2026-04-23-ai-automation-center-mainline-plan.md`
- 本次任务目标：把设置页 `AIAutomationSource` 从“只显示当前 active profile ID”收口到“显式选择 Profile + 展示元数据 + 明确不覆盖手动配置”的最小闭环
- 本次已盘点本地资源：AGENTS.md、canonical plan、AI 自动化需求/实施计划、当前状态文档、workflow、审计报告、`web/src/components/modules/setting/AIAutomationSource.tsx`、`web/src/api/endpoints/ai-automation.ts`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H 设置页来源切换主线
- 不扩散到 AI 自动化整页重构，不改后端 contract
- 设置页必须向需求文档收敛：展示当前激活来源、方案名称、更新时间、置信度、回退说明
- AI Profile 切换不得覆盖手动配置，只改变读取来源

## 4. 本次禁止事项

- 不改 AI task / profile schema
- 不把“方案切换”误改成静默覆盖手动配置
- 不引入新的 raw i18n key 或中英混杂主文案

## 5. 本次验收条件

- 设置页来源卡片支持从 profile 列表中显式选择方案
- 卡片展示当前选中方案的名称、ID、版本、更新时间、置信度与说明
- 卡片明确说明“只切换读取来源，不覆盖手动配置”
- focused 验证至少包含 `tsc --noEmit`、locale consistency 与 `git diff --check`

## 6. 本次回滚点

- 仅回退 `web/src/components/modules/setting/AIAutomationSource.tsx`
- 仅回退 `web/src/components/modules/setting/AIAutomationSource.test.tsx`
- 仅回退四语 locale 中 `setting.aiAutomationSource` 新增键

## 7. 实现范围

- 受影响前端模块：`web/src/components/modules/setting/AIAutomationSource.tsx`
- 受影响测试：`web/src/components/modules/setting/AIAutomationSource.test.tsx`
- 受影响 locale：`web/public/locale/en.json`、`zh-Hans.json`、`zh-Hant.json`、`ja.json`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强设置页来源卡片的展示与激活入口

## 8. 实施步骤

1. 复核 AI 自动化需求文档、审计报告和当前设置页实现，确认真实缺口是“只显示 active profile ID，缺少显式选择与元数据展示”。
2. 复核后端与前端现有 hooks，确认无需扩 API，直接复用 `useAIProfiles` 与 `useActivateAIProfile`。
3. 重写 `AIAutomationSource.tsx`，补齐 profile 选择器、显式启用按钮、元数据摘要卡片和手动配置保护说明。
4. 新增 `AIAutomationSource.test.tsx`，覆盖元数据展示、切换到 `ai_profile` 时激活所选 Profile、切回 manual 时仅更新 setting。
5. 补齐四语 locale 新键，并修正激活失败 toast 文案，避免引用不存在的根级 `setting.saveFailed`。

## 9. 测试与验证

- 已执行：`. .\scripts\use-node-env.ps1; node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 已执行：`. .\scripts\use-node-env.ps1; node .\scripts\verify-locale-consistency.mjs`
- 已执行：`git diff --check -- web/src/components/modules/setting/AIAutomationSource.tsx web/src/components/modules/setting/AIAutomationSource.test.tsx web/public/locale/zh-Hans.json web/public/locale/en.json web/public/locale/zh-Hant.json web/public/locale/ja.json`
- 未通过：`. .\scripts\use-node-env.ps1; node .\web\node_modules\vitest\vitest.mjs run .\web\src\components\modules\setting\AIAutomationSource.test.tsx --config .\web\vitest.config.ts`
  - 失败原因：宿主 `vite/esbuild` 启动阶段 `spawn EPERM`，属于当前 Windows 宿主进程创建限制，不是本轮业务逻辑断言失败

## 10. 风险与兼容性

- 兼容性风险：低；只增强设置页展示与激活入口
- 已收敛风险：设置页现在不再只暴露 Profile ID，减少“像方案切换、实际看不到方案”的认知偏差
- 仍存风险：本轮只补了设置页可见层闭环，AI Profile 的真实 consumer 仍主要停留在 AI 自动化执行配置层，尚未把分组/渠道/价格方案消费闭环补齐
- 工作区状态说明：`web/src/components/modules/setting/AIAutomationSource.tsx` 当前在 git 视角为未跟踪文件（`git ls-files` 无该路径），并非本轮误删 tracked 文件；下一轮不应把这一状态误判为“删除用户文件”

## 11. 收工记录

- 构建是否通过：focused 类型检查通过
- 测试是否通过：部分通过
  - 通过：TypeScript `tsc --noEmit`
  - 通过：`verify-locale-consistency.mjs`
  - 通过：`git diff --check`
  - 受宿主阻塞未完成：单测 `vitest`（`spawn EPERM`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、AI 自动化需求与实施计划、当前状态文档、workflow、审计报告、`using-superpowers` / `brainstorming` skill 文档、相关前端源码与 locale
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - AI 自动化需求文档：确认设置页必须展示方案名称、更新时间、置信度与风险提示，并支持显式切换
  - 审计报告：确认当前最大前端缺口是 `AIAutomationSource.tsx` 只支持模式切换和显示 ID
  - 前端 hooks / handlers：确认现有 `useAIProfiles` + `useActivateAIProfile` 已足够支撑本轮小闭环
- worklog 是否更新：是
- 遗留项：
  - `vitest` 仍受宿主 `spawn EPERM` 阻塞，需在环境正常的宿主或 CI 补跑
  - AI Profile 主线下一步更适合继续补“真实方案 consumer / 设置页更完整元数据 / 风险提示”而不是回到只看 endpoint 配置
- 下一任务前置条件是否满足：满足；下一轮可继续同一主线，优先把设置页与 AI Profile consumer 之间的契约再向需求文档收敛，或者在环境允许时补跑这次新增测试
