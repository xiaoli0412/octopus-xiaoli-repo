# 2026-04-28 Phase H AI learning stable CDP expectation closure

## 1. 任务信息

- 任务名称：Phase H6 learning smoke stable CDP expectation closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-explicit-external-cdp-preflight-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke wrapper / stable diagnostic 收口主线，让 `check-only` 回放能直接判断“当前命令要求的 external CDP 预检”和“稳定副本当时实际记录的 requireCdp”是否一致，并顺手修复本轮验证链中暴露出的两个前端语法断点，恢复 `tsc --noEmit`
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 副本、`web/src/components/modules/ai-automation/index.tsx`、`web/src/api/endpoints/ai-automation.ts`
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、AI 自动化/动态路由实施计划、最近 H6 worklog、`brainstorming` skill 文档、repo-local stable diagnostic 副本
- 若未使用部分本地资源或上下文，原因：本轮不需要重新展开 relay/runtime 业务逻辑，也不需要继续读取无关 UI 模块或备份主线文档
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / diagnostic 主线，不扩散到 AI 页面产品逻辑、relay 推荐评分或宿主网络层调试
- 动态路由 AI 学习只影响运行时推荐，不覆盖用户配置；本轮只允许改验证入口消费层与为恢复验证链所需的最小语法修复
- 对已存在的前端语法断点只做最小修复，不顺势改写页面结构或接口语义

## 4. 本次禁止事项

- 不把任务扩成真实 external/browser 服务联调
- 不把 `host_networking_blocker`、`vite/esbuild spawn EPERM` 或 Windows 宿主问题误记成产品逻辑回归
- 不借修 `tsc` 之机扩大 AI 自动化页面或 endpoint 的业务改造范围

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 能显示本次命令的 external CDP 预期
- 当本次命令显式要求 `-RequireExternalCdpPreflight`、但 stable copy 仍是历史 `requireCdp=false` 时，回放会直接给出 mismatch 提示和刷新 stable copy 的命令
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- 回退 `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- 回退 `scripts/verify-ai-automation-learning-focus.mjs`
- 回退 `web/src/components/modules/ai-automation/index.tsx`
- 回退 `web/src/api/endpoints/ai-automation.ts`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本消费层；随后仅做恢复 `tsc` 所需的最小前端语法修复
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/ai-automation/index.tsx`、`web/src/api/endpoints/ai-automation.ts`（仅语法修复）
- 受影响接口：无业务 API 语义变化；仅修复前端 endpoint 字符串拼接语法
- 是否影响旧数据：否
- 是否影响旧行为：仅改善 stable diagnostic replay 提示，并恢复前端类型检查链

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、当前状态文档与最近 H6 worklog，确认本轮仍应围绕 learning smoke stable diagnostic 收口推进。
2. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增“本次命令 external CDP 预期”的推导与 mismatch 提示，让 `check-only` 回放能比较 stable copy 的 `requireCdp` 与当前命令预期。
3. 更新 `verify-ai-automation-learning-focus.mjs` 守护新 helper、新输出行与 catch/check-only 传参契约。
4. 跑两次 `check-only` 时发现 `tsc --noEmit` 被已存在的 JSX / 模板字符串语法断点阻塞，定位到 `ai-automation/index.tsx` 与 `api/endpoints/ai-automation.ts`。
5. 对上述两个文件做最小语法修复：删除 profile 预览卡多余的 JSX 关闭片段，并补回 `task artifacts/retry` endpoint 缺失的模板字符串反引号。
6. 重新运行 `verify-ai-automation-learning-focus.mjs`、两次 `check-only`、`tsc --noEmit` 与 `git diff --check`，确认同一条 H6 验证链重新闭环。
7. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，给下一轮留下“健康宿主 external 复跑 `requireCdp=true` stable copy”的明确入口。

## 9. 测试与验证

- 构建命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; . .\\scripts\\use-node-env.ps1; & $env:NODEEXE .\\web\\node_modules\\typescript\\bin\\tsc --noEmit -p .\\web\\tsconfig.json`
- 测试命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; . .\\scripts\\use-node-env.ps1; & $env:NODEEXE .\\scripts\\verify-ai-automation-learning-focus.mjs`
- 专项验证：
  - `$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; & .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
  - `$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; & .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
  - `git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs web/src/components/modules/ai-automation/index.tsx web/src/api/endpoints/ai-automation.ts docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-28-phase-h-ai-learning-stable-cdp-expectation-closure.md`

## 10. 风险与兼容性

- 新风险：本轮 stable replay 新增 mismatch 提示后，下一轮若 stable copy 仍长期不刷新，接手者更容易看到“为什么还没变绿”，但不会误把旧 artifact 当成本次命令结果；这是预期暴露，不是新产品风险
- 兼容性风险：低；wrapper 只增加提示，不改变 external/self-start 主路径，前端改动仅为已存在语法断点修复
- 是否阻塞下一任务：不阻塞；下一轮可直接在健康宿主或可达服务环境执行 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（learning focus 守护、两次 `check-only`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、AI 自动化/动态路由实施计划、最近 H6 worklog、`brainstorming` skill 文档、repo-local stable diagnostic 副本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 H6 worklog 说明上轮已把 external CDP 预检入口显式化，因此本轮最值得推进的是“stable replay 是否能识别和当前命令的 CDP 预期失配”；用户上下文总账和实施计划限制本轮不得越界到业务逻辑；stable diagnostic 副本证明当前 repo-local artifact 仍是 `requireCdp=false`
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未做真实 external/browser；继续使用 repo-local `check-only` 收口验证入口
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 external 真实服务仍是 `host_networking_blocker`；`vite/esbuild spawn EPERM` 仍阻塞真实 jsdom/vitest 入口
- 待验证页面清单：健康宿主上的 `AI 自动化` learning external/browser 证据，尤其是 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：需要在可达 backend/frontend 的宿主上实际生成一份 `requireCdp=true` 的 fresh diagnostic，验证 stable copy 不再把 `cdp` 标成 skipped；真实 jsdom/vitest 入口仍待宿主 `esbuild spawn EPERM` 解阻后补跑
- 下一任务前置条件是否满足：满足；下一轮可直接用同一条 H6 诊断链继续收集 fresh external/browser 证据
