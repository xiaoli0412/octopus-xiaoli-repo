# 2026-04-28 Phase H AI learning explicit external CDP preflight closure

## 1. 任务信息

- 任务名称：Phase H6 learning smoke explicit external CDP preflight closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-freshness-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke wrapper / diagnostic 收口主线，把“external 第一步是否强制检查 CDP”从隐式分支改成显式可控入口，并让 `check-only` 回放能直接看出 stable diagnostic 当时是否要求了 CDP
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 副本
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、Windows apply_patch fallback 经验、repo-local stable diagnostic 副本
- 若未使用部分本地资源或上下文，原因：本轮不需要重新展开产品 UI、relay 实现或 jsdom/vitest 深入排查，故未继续读取无关模块源码
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / 状态文档主线，不扩散到 AI 页面行为、settings UI、relay 排序逻辑或共享 preflight schema 大改
- 动态路由 AI 学习只影响运行时推荐，不覆盖用户配置；本轮只允许改验证入口与诊断消费层，不得改业务语义
- 保持 host-friendly external 默认行为兼容：只有显式参数才开启“external 第一步强制检查 CDP”

## 4. 本次禁止事项

- 不把当前主线扩展成真实 browser/self-start 服务调试
- 不把宿主 `host_networking_blocker`、`vite/esbuild spawn EPERM` 或 `apply_patch` 包装器损坏误记成产品回归
- 不重写 shared wrapper 现有 external 预检结构，只做最小增量契约收口

## 5. 本次验收条件

- shared CDP wrapper 存在显式 `RequireExternalCdpPreflight` 开关，并真正驱动第一次 external preflight 的 `RequireCdp`
- AI learning wrapper 能透传该开关，并在 `check-only` 与稳定副本回放中明确显示诊断当时是否要求了 CDP
- no-browser 守护、`tsc --noEmit` 与 learning smoke `check-only` 链路通过

## 6. 本次回滚点

- 回退 `scripts/verify-channel-create-browser-smoke-cdp.ps1`
- 回退 `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- 回退 `scripts/verify-ai-automation-learning-focus.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本契约与状态文案，无 UI 改动
- 受影响后端模块：无
- 受影响前端模块：无产品代码改动；仅更新前端相关 no-browser 守护脚本
- 受影响接口：无业务 API 变化
- 是否影响旧数据：否
- 是否影响旧行为：仅影响 learning smoke wrapper 的 external 诊断入口与 `check-only` 输出；默认 host-friendly external 行为保持兼容

## 8. 实施步骤

1. 复核 automation memory、canonical plan、当前状态文档、前端主线状态与最近 H6 worklog，确认本轮仍应围绕 learning smoke wrapper 的验证入口收口推进。
2. 在共享 `verify-channel-create-browser-smoke-cdp.ps1` 中新增显式 `RequireExternalCdpPreflight` 开关，并把 external 第一步 `RequireCdp` 逻辑改为由 `requireExternalCdpPreflight` 变量统一控制，同时补上 `check-only` 输出。
3. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中透传该开关，补充 `External preflight CDP requirement` 回放行和新的 next-step 命令。
4. 更新 `verify-ai-automation-learning-focus.mjs` 守护新参数与新输出契约。
5. 运行 `verify-ai-automation-learning-focus.mjs`、`tsc --noEmit`、`check-only` 和 `check-only -RequireExternalCdpPreflight`，确认 wrapper 契约、输出和入口全部对齐。
6. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，保证下一轮可直接继承“是否需要显式强制 external CDP 预检”的入口。

## 9. 测试与验证

- 构建命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; . .\\scripts\\use-node-env.ps1; & $env:NODEEXE .\\web\\node_modules\\typescript\\bin\\tsc --noEmit -p .\\web\\tsconfig.json`
- 测试命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; . .\\scripts\\use-node-env.ps1; & $env:NODEEXE .\\scripts\\verify-ai-automation-learning-focus.mjs`
- 专项验证：
  - `$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; & .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
  - `$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\\Local'; & .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
  - `git diff --check -- scripts/verify-channel-create-browser-smoke-cdp.ps1 scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-explicit-external-cdp-preflight-closure.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 10. 风险与兼容性

- 新风险：若下一轮直接用 `check-only -RequireExternalCdpPreflight` 观察 stable 副本，仍只能看到当前已保存诊断的 `requireCdp=false` 历史状态；真正的 `requireCdp=true` 诊断仍需在健康宿主或可达服务上重跑 external 才会落到 stable copy
- 兼容性风险：低；本轮新增开关默认为关闭，不改变现有 host-friendly external 默认路径
- 是否阻塞下一任务：不阻塞；下一轮可以在健康宿主或服务可达环境直接使用 `-RequireExternalCdpPreflight` 获取第一份同时包含服务与 CDP reachability 的 fresh diagnostic

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（learning focus 守护与两次 `check-only`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、Windows apply_patch fallback 经验、repo-local stable diagnostic 副本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 H6 worklog 说明上轮已完成 stable freshness 提示，因此本轮最值得推进的是“第一步 external 是否强制检查 CDP”的显式入口；stable diagnostic 副本证明当前 `requireCdp=false` 的历史快照仍在；主计划与动态路由实施计划限定本轮不得越界到产品逻辑
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未做真实 browser/self-start；继续使用 repo-local `check-only` 收口验证入口
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 external 真实服务仍是 `host_networking_blocker`；`vite/esbuild spawn EPERM` 仍阻塞真实 jsdom/vitest 入口
- 待验证页面清单：健康宿主上的 `AI 自动化` learning external/browser 证据，尤其是 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：需要在可达 backend/frontend 的宿主上实际生成一份 `requireCdp=true` 的 fresh diagnostic，验证 stable copy 不再把 CDP 标成 skipped
- 下一任务前置条件是否满足：满足；下一轮可直接用新参数继续同一条 H6 诊断链
