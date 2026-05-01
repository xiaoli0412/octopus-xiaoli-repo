# 2026-04-27 Phase H Setting AI Profile Risk And Fallback Closure

## 1. 任务信息

- 任务名称：setting ai profile risk and fallback closure
- 日期：2026-04-27
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H5 设置页 `manual / ai_profile` 切换收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 验收标准 19-21
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-27-phase-h-setting-ai-profile-source-card-closure.md`
- 本次任务目标：把设置页 `AIAutomationSource` 从“只显示元数据”继续收口到“显式展示 Profile 状态、风险提示、以及后端自动回退 manual 的运行时语义”
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、AI 自动化需求/实施计划、当前状态文档、上一轮 Phase H worklog、`internal/op/ai_automation.go`、`internal/op/ai_automation_test.go`、`web/src/components/modules/setting/AIAutomationSource.tsx`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档与源码，以及 session 开头读取的 `using-superpowers` / `brainstorming` 本地 skill 文件
- 若未使用部分本地资源或上下文，原因：本轮是设置页来源卡片的小闭环，不需要扩展到更大的 AI task / backup / browser smoke 资源池
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：不适用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：不适用
- 若未使用子 agent，原因：用户明确要求不要创建子 agent；本轮任务边界也足够小，主线程可直接闭环

## 3. 本次硬规则

- 设置页必须向 Phase H5 契约收拢：展示当前来源、Profile 状态、风险提示和回退语义
- 继续保持“只切换读取来源，不覆盖手动配置”
- 风险提示必须对齐后端真实语义，不能凭空扩写运行时行为

## 4. 本次禁止事项

- 不扩散到 AI Automation 顶层页面的大范围 UI 改造
- 不改后端 `AIAutomationConfigGet` 语义，只消费并显性化当前已存在的 fallback 事实
- 不触碰与本轮无关的设置页卡片或其他主线模块

## 5. 本次验收条件

- 设置页来源卡片可区分 `active / ready / draft / archived / invalid` 等 Profile 状态
- 低置信度、草稿、归档、无效 Profile 会显示可见风险提示
- 当后端已经把 `ai_profile` 回退成 `manual` 时，前端会显式提示“运行时已回退 manual”
- `manualSafety / fallbackHint` 继续保留，且四语 locale 一致

## 6. 本次回滚点

- 仅回退 `web/src/components/modules/setting/AIAutomationSource.tsx`
- 仅回退 `web/src/components/modules/setting/AIAutomationSource.test.tsx`
- 仅回退 `web/public/locale/en.json`
- 仅回退 `web/public/locale/zh-Hans.json`
- 仅回退 `web/public/locale/zh-Hant.json`
- 仅回退 `web/public/locale/ja.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先核对后端 fallback 语义，再改前端 UI
- 受影响后端模块：无代码改动；仅读取 `internal/op/ai_automation.go` / test 作为语义依据
- 受影响前端模块：`web/src/components/modules/setting/AIAutomationSource.tsx`
- 受影响接口：`web/src/api/endpoints/ai-automation.ts` 现有 `useAIAutomationConfig` 消费链
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强设置页来源卡片的状态与风险可见性

## 8. 实施步骤

1. 核对需求文档与 `AIAutomationConfigGet` / `fallbackManualAIAutomationConfig` 的真实行为，确认“无效 Profile 自动回退 manual”是代码事实。
2. 在 `AIAutomationSource.tsx` 中接入 `useAIAutomationConfig`，同时保留原 `Setting` 视角，补齐状态 badge、风险提示、runtime fallback notice 和无效 Profile 激活保护。
3. 扩展 `AIAutomationSource.test.tsx`，覆盖 ready/invalid/runtime fallback 三类关键场景，并补齐四语 locale 文案。

## 9. 测试与验证

- 构建命令：`$env:APPDATA='C:\Users\李昊桐\AppData\Roaming'; $env:LOCALAPPDATA='C:\Users\李昊桐\AppData\Local'; $env:ProgramData='C:\ProgramData'; $env:SystemRoot='C:\Windows'; $env:WINDIR='C:\Windows'; $env:TEMP='C:\Users\李昊桐\AppData\Local\Temp'; $env:TMP='C:\Users\李昊桐\AppData\Local\Temp'; $env:USERPROFILE='C:\Users\李昊桐'; $env:COMSPEC='C:\Windows\System32\cmd.exe'; node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：`$env:APPDATA='C:\Users\李昊桐\AppData\Roaming'; $env:LOCALAPPDATA='C:\Users\李昊桐\AppData\Local'; $env:ProgramData='C:\ProgramData'; $env:SystemRoot='C:\Windows'; $env:WINDIR='C:\Windows'; $env:TEMP='C:\Users\李昊桐\AppData\Local\Temp'; $env:TMP='C:\Users\李昊桐\AppData\Local\Temp'; $env:USERPROFILE='C:\Users\李昊桐'; $env:COMSPEC='C:\Windows\System32\cmd.exe'; node -r .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run .\web\src\components\modules\setting\AIAutomationSource.test.tsx --config .\web\vitest.config.ts`
- 专项验证：`node scripts/verify-locale-consistency.mjs`、`git diff --check -- web/src/components/modules/setting/AIAutomationSource.tsx web/src/components/modules/setting/AIAutomationSource.test.tsx web/public/locale/en.json web/public/locale/zh-Hans.json web/public/locale/zh-Hant.json web/public/locale/ja.json`

## 10. 风险与兼容性

- 新风险：来源卡片现在同时参考 settings 与 `AIAutomationConfig`，若后续两者刷新节奏变化，需留意显示瞬态
- 兼容性风险：低；仅新增文案和风险层，不改变实际设置提交协议
- 是否阻塞下一任务：不阻塞；下一轮可继续同主线，优先补 vitest 环境证据或继续向真实 consumer 收口

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 在补齐 Windows 环境变量后通过
- 测试是否通过：部分通过；`verify-locale-consistency.mjs` 通过，目标 vitest 仍在 `vite/esbuild` 启动阶段报 `spawn EPERM`
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、AI 自动化需求/实施计划、当前状态文档、上一轮 Phase H worklog、automation memory、`internal/op/ai_automation.go`、`internal/op/ai_automation_test.go`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：需求文档给出 H5 契约；后端 op/test 明确了 invalid/archived/缺内容时自动回退 manual；上一轮 worklog 给出本轮应继续补“状态/风险提示”而非再改选择器；automation memory 提示上一轮 vitest 宿主限制仍在
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：不适用
- 手工 smoke 状态：未执行浏览器 smoke；本轮聚焦 no-browser 小闭环
- 手工 smoke 阻塞原因 / 缺少的环境：不需要浏览器；前端单测仍受宿主 `spawn EPERM` 影响
- 待验证页面清单：设置页 `AIAutomationSource` 浏览器级状态展示与交互；环境允许时补跑 `AIAutomationSource.test.tsx`
- 若未使用子 agent，原因：用户明确禁止创建子 agent，且本轮范围足够小
- worklog 是否更新：已更新
- 遗留项：`vitest` / `vite` / `esbuild` 子进程创建仍受宿主限制；AI Profile 的真实 consumer 仍主要停留在执行配置层
- 下一任务前置条件是否满足：满足；可继续同主线补测试环境证据，或转向更大的 AI Profile consumer 闭环
