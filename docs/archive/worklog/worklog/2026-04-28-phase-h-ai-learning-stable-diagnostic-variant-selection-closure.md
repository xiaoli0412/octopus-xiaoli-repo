# 2026-04-28 Phase H AI learning stable diagnostic variant selection closure

## 1. 任务信息

- 任务名称：Phase H6 learning stable diagnostic variant selection closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-stable-cdp-expectation-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable diagnostic 主线，把 repo-local stable diagnostic 从“单一总副本”收口成“按 `requireCdp` 预期分桶并优先命中匹配副本”的入口，同时修复 `check-only` 回放后误入共享 wrapper 的脚本流转问题
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 副本目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、AI 自动化/动态路由实施计划、最近 H6 worklog、`brainstorming` skill 文档、repo-local stable diagnostic 副本
- 若未使用部分本地资源或上下文，原因：本轮不需要重新展开 relay/runtime 产品逻辑，也不需要切回备份、设置或其他前端主线
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 验证入口主线，不扩散到 AI 页面产品逻辑、relay 评分或宿主网络修复主题
- 只改 wrapper、静态守护与状态记录，不做无关 UI/接口行为扩散修改
- 必须形成真实代码增量与可验证闭环，不能只做阅读/总结

## 4. 本次禁止事项

- 不把任务扩成真实 external/browser 服务联调
- 不修改动态路由学习数据模型、AI Profile 行为或设置页业务交互
- 不把 `host_networking_blocker`、`vite/esbuild spawn EPERM` 或宿主目录权限问题误记成产品回归

## 5. 本次验收条件

- external stable diagnostic 会按 `requireCdp` 生成 requirement-specific repo-local 副本
- `check-only` 优先回放与本次命令匹配的 stable 副本，并明确提示是命中匹配副本还是回退到最近可用副本
- `check-only` 回放 stable summary 后直接退出，不再继续误入共享 wrapper
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品数据与 UI，只改 learning smoke wrapper 的 stable diagnostic 选副本逻辑与静态守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 repo-local stable diagnostic 的落盘/回放策略，并修复 `check-only` 控制流

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、环境 next plan 和最近 H6 worklog，确认本轮仍应停留在 learning smoke stable diagnostic 主线。
2. 复查 `verify-ai-automation-learning-browser-smoke.ps1` 当前逻辑，确认它只维护单一 stable copy，并在 `check-only` 回放后仍会继续执行共享 wrapper。
3. 在 wrapper 中新增 stable copy 目录、requirement-specific 变体命名与选副本 helper，并让 external 失败时除 legacy copy 外同步写入 `optional-cdp` 或 `require-cdp` 变体副本。
4. 为兼容已有历史副本，补一层 `Sync-StableExternalPreflightDiagnosticVariantCopyFromLegacy`，在只存在 legacy copy 时自动回填当前可推导出的 requirement-specific 副本。
5. 更新 `check-only` 选择顺序与提示，优先命中匹配副本，并通过 `External preflight stable copy note` 明确说明当前是“命中匹配副本”还是“无匹配副本而回退”。
6. 修复 `check-only` 模式在输出 stable preview 后继续误入共享 wrapper 的问题，改为直接 `exit 0`。
7. 更新 `verify-ai-automation-learning-focus.mjs` 守护上述 helper、变体文件名、新提示文本与 `check-only -> exit 0` 控制流。
8. 运行两次 `check-only`、静态守护、`tsc --noEmit` 与 `git diff --check`，确认 Phase H6 验证链闭环。
9. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，给下一轮留下“在健康宿主上获取 `requireCdp=true` fresh artifact”的明确入口。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-variant-selection-closure.md`

## 10. 风险与兼容性

- 新风险：低；wrapper 只增强 repo-local stable diagnostic 的存储/回放策略，并修复 `check-only` 控制流
- 兼容性风险：低；保留 legacy stable copy 路径，不影响已有文档和人工查阅路径，同时为新 requirement-specific 副本提供兼容回填
- 是否阻塞下一任务：不阻塞；下一轮可直接在健康宿主或可达服务环境执行 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight` 来生成真正的 `requireCdp=true` fresh artifact

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（learning focus 守护、两次 `check-only`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、AI 自动化/动态路由实施计划、最近 H6 worklog、`brainstorming` skill 文档、repo-local stable diagnostic 副本目录
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 H6 worklog 说明上轮已把 stable replay 的 CDP 预期提示补齐，因此本轮最值得推进的是“同一 repo-local artifact 是否能按命令预期选副本、并停止误入共享 wrapper”；用户上下文总账和实施计划限制本轮不得越界到业务逻辑；repo-local stable diagnostic 目录证明当前仍只有 legacy copy，需要先补 requirement-specific 兼容回填
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未做真实 external/browser；继续使用 repo-local `check-only` 收口 stable diagnostic 入口
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 external 真实服务仍是 `host_networking_blocker`；`vite/esbuild spawn EPERM` 仍阻塞真实 jsdom/vitest 入口
- 待验证页面清单：健康宿主上的 `AI 自动化` learning external/browser 证据，尤其是 `-Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight` fresh artifact
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：需要在可达 backend/frontend 的宿主上生成一份真正的 `latest-external-preflight-diagnostic-require-cdp.json`，确认 stable replay 不再回退到 legacy copy；真实 jsdom/vitest 入口仍待宿主 `esbuild spawn EPERM` 解阻后补跑
- 下一任务前置条件是否满足：满足；下一轮可直接用同一条 H6 诊断链继续采集 fresh external/browser 证据
