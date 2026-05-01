# 2026-04-29 Phase H AI learning refresh command closure

## 1. 任务信息

- 任务名称：Phase H6 learning refresh command closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-stable-copy-status-visibility-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable diagnostic 主线，把 `check-only` 从“看完回放后还要自己拼 external 命令”收口成“直接给出与当前调用一致的 preferred refresh command”，减少下一轮因手工重拼命令导致的诊断漂移
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 H6 worklog、`using-superpowers` / `brainstorming` skill 文档
- 若未使用部分本地资源或上下文，原因：本轮不需要重新展开 AI 页面产品逻辑、relay/runtime 评分逻辑或宿主网络修复，只需沿已有 H6 验证入口做 bounded closure
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 验证入口主线，不扩散到 AI 页面产品逻辑、后端 schema 或宿主网络修复主题
- 只改 wrapper、静态守护与状态记录，不做无关 UI/接口行为扩散修改
- 必须形成真实代码增量与可验证闭环，不能只做阅读/总结

## 4. 本次禁止事项

- 不把任务扩成真实 external/browser 服务联调
- 不修改动态路由学习数据模型、AI Profile 行为或设置页业务交互
- 不把 `host_networking_blocker`、`vite/esbuild spawn EPERM` 或沙箱路径差异误记成产品回归

## 5. 本次验收条件

- `check-only` 回放里会直接打印 `External preflight preferred refresh command`
- 推荐命令会沿用当前调用中显式给出的 `-RequireExternalCdpPreflight`、URL、bootstrap、timeout 等 profile 参数
- `next steps` 会优先复用该推荐命令，而不是退回固定模板
- `verify-ai-automation-learning-focus.mjs`、两次 learning smoke `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品数据与 UI，只改 learning smoke wrapper 的 refresh command 生成逻辑与静态守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 repo-local stable diagnostic 回放里的操作建议密度

## 8. 实施步骤

1. 复核 automation memory、主规划、用户上下文总账、详细工作流、当前状态文档、前端主线状态和最近 H6 worklog，确认上一轮已完成 stable copy 状态可见性，本轮最值得推进的是让 replay 结果直接给出当前调用专属的 external 复跑命令。
2. 复查 `verify-ai-automation-learning-browser-smoke.ps1` 当前 `next steps` 仍是固定模板，确认它没有复用本次调用里显式给出的 `RequireExternalCdpPreflight` 或其他 profile 参数。
3. 在 wrapper 中新增命令字符串 helper，根据当前 `PSBoundParameters` 生成 `Mode=external` 的 preferred refresh command，并保留与本次调用一致的 profile。
4. 让 `check-only` 回放和 `Write-ExternalPreflightDiagnosticSummary` 都直接打印 `External preflight preferred refresh command`。
5. 让 `Get-ExternalPreflightNextStepLines` 优先复用该命令；若本次回放已经是 `RequireExternalCdpPreflight=true`，则不再重复追加“再加一次 RequireExternalCdpPreflight”的通用模板提示。
6. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新的 helper、输出字段和 `next steps` 文案纳入静态守护。
7. 运行两次 `check-only`、静态守护、`tsc --noEmit` 与 `git diff --check`，确认 Phase H6 验证链闭环。
8. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，给下一轮留下“直接复制 preferred refresh command 去健康宿主补 fresh artifact”的明确入口。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-29-phase-h-ai-learning-refresh-command-closure.md`

## 10. 风险与兼容性

- 新风险：低；wrapper 只增强 repo-local stable diagnostic 回放时的命令建议，不改变 external 失败分类逻辑
- 兼容性风险：低；原有固定模板 next step 仍保留在没有推荐命令可复用时的兜底路径
- 是否阻塞下一任务：不阻塞；下一轮可直接复制 `preferred refresh command` 到健康宿主或服务可达环境中执行，补 `requireCdp=true` fresh artifact

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（learning focus 守护、两次 `check-only`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 H6 worklog、`using-superpowers` / `brainstorming` skill 文档
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 H6 worklog 说明上一轮已完成 stable copy 可见性，因此本轮最值得推进的是把 replay 后的下一步动作显式化；当前状态文档与前端主线状态确认剩余缺口仍是健康宿主上的真实 external/browser 证据，而不是 wrapper 可消费性不足
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未做真实 external/browser；继续使用 repo-local `check-only` 收口 stable diagnostic 入口
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 external 真实服务仍是 `host_networking_blocker`；`vite/esbuild spawn EPERM` 仍阻塞真实 jsdom/vitest 入口
- 待验证页面清单：健康宿主上的 `AI 自动化` learning external/browser 证据，尤其是通过 replay 结果给出的 `preferred refresh command` 采集真正的 `requireCdp=true` artifact
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：仍需在健康宿主或可达 backend/frontend 的环境中执行 replay 给出的 `preferred refresh command`，补一份真正的 `requireCdp=true` fresh external artifact
- 下一任务前置条件是否满足：满足；下一轮优先继续同一条 Phase H6 诊断链，直接把 replay 提示给出的 refresh command 拿到健康宿主补 fresh external/browser 证据
