# 2026-04-29 Phase H AI learning action-class closure

## 1. 任务信息

- 任务名称：Phase H6 learning action-class closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 1.0 / 1.2 / 1.3 / 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-recommended-action-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable replay 主线，把 blocked-host `check-only` 的动作建议进一步压成稳定的 `External preflight recommended action class`，让下一轮可以先按分类判断该继续 replay 还是去健康宿主补证。
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 副本
- 若未使用部分本地资源或上下文，原因：本轮不需要重开 UI、后端 schema、真实 external/browser 联调或宿主网络修复；连续 Phase H6 记录已经把工作面收敛到 replay 输出层
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / stable replay / 状态文档主线，不扩散到 AI 页面交互、settings UI、后端 schema 或宿主网络修复
- 本轮必须形成真实代码增量与直接验证，不能只更新文档或复述阻塞
- 只增强 blocked-host `check-only` 的稳定分类输出与静态守护，不改变 external preflight schema、artifact 命名或真实 external/browser 行为

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*`、`web/src/components/modules/setting/*` 或任何用户可见产品行为
- 不扩大到真实 external/browser 联调，也不尝试修复 `host_networking_blocker`、`vite/esbuild spawn EPERM` 或 Windows socket provider 初始化问题
- 不在 repo-local stable diagnostic 目录中保留任何伪造的 `requireCdp=true` artifact

## 5. 本次验收条件

- `check-only` 输出新增稳定的 `External preflight recommended action class`
- 默认 invocation 命中 fresh matching 副本时，分类输出 `matched_replay_ready`
- `-RequireExternalCdpPreflight` 仍缺 matching 副本时，分类输出 `fallback_only_refresh_required`
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-action-class-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品 UI 或外部数据；只改 wrapper 说明层与静态守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 回放里的稳定分类输出，不改变当前 blocked-host 结论、artifact 选择或 refresh 命令生成方式

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog，确认本轮仍应停留在 repo-local stable replay 输出层而不是扩大到宿主网络或真实 browser。
2. 检查 `verify-ai-automation-learning-browser-smoke.ps1` 与 `verify-ai-automation-learning-focus.mjs`，确认现有链路已经具备 `decision summary + recommended action`，缺口只剩“下一轮能否先按稳定标签做决策”。
3. 在 `scripts/verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Get-StableExternalPreflightRecommendedActionClass`，统一根据 selected/matching 状态、freshness 和 requirement-specific 缺口输出固定分类，并在 summary 输出层新增 `External preflight recommended action class`。
4. 在 `scripts/verify-ai-automation-learning-focus.mjs` 中补上 helper 存在、输出标签与固定分类字面量的静态守护，防止这条分类出口后续无声回退。
5. 运行默认与 `-RequireExternalCdpPreflight` 两次 learning smoke `check-only`、`verify-ai-automation-learning-focus.mjs` 与 `tsc --noEmit`，确认默认 replay 输出 `matched_replay_ready`，require-CDP replay 输出 `fallback_only_refresh_required`。
6. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，给下一轮留下“先看 recommended action class，再决定是否去健康宿主补证”的直接入口。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`

## 10. 风险与兼容性

- 新风险：低；本轮只增加 blocked-host replay 的稳定分类输出与静态守护
- 兼容性风险：低；不改 external preflight schema、不改 artifact 选择逻辑、不改产品运行态行为
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（两次 learning smoke `check-only`、`verify-ai-automation-learning-focus.mjs`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态、最近 Phase H6 worklog、repo-local stable diagnostic 目录
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 Phase H6 worklog 说明上一轮已经把 replay 输出收口到 `recommended action`，因此本轮最值得推进的是再补一个稳定分类出口；repo-local stable diagnostic 目录显示当前真实环境仍只有 parseable `requireCdp=false` 副本，因此最需要的是“保持 replay 还是去补证”的快速分档，而不是继续横向扩写 note
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未做真实 external/browser；继续使用 repo-local `check-only` 和静态守护收口 stable replay 输出层
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 external 真实服务仍是 `host_networking_blocker`；`vite/esbuild spawn EPERM` 仍阻塞真实 jsdom/vitest 入口
- 待验证页面清单：健康宿主上的 `AI 自动化` learning external/browser 证据，尤其是 `requireCdp=true` fresh artifact
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 仍需在健康宿主或可达 backend/frontend 的环境中执行 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`，补一份真正的 `requireCdp=true` stable artifact
  - 当前 blocked-host replay 已同时给出 `decision summary + recommended action + recommended action class`，因此下一轮应优先读取分类标签；若为 `fallback_only_refresh_required`，直接转向健康宿主补 requirement-specific artifact，不再回头扩写 note
- 下一任务前置条件是否满足：满足；下一轮优先继续同一条 Phase H6 诊断链，先读 `External preflight recommended action class`，若仍是 `fallback_only_refresh_required`，就直接去健康宿主执行 preferred refresh command，不再回头重开 UI/runtime 主题
