# 2026-04-28 Phase H AI Learning Stable Diagnostic Copy Closure

## 1. 任务信息

- 任务名称：phase h ai learning stable diagnostic copy closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-external-next-step-guidance-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 external 失败诊断从“仅依赖临时目录 artifact”收口成“repo-local 稳定副本 + wrapper 输出可继承”的闭环，减少下一轮重新定位诊断文件的摩擦
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、`.local-tools/apply_patch.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、Windows apply_patch fallback 经验
- 若未使用部分本地资源或上下文，原因：未使用外部资料；当前任务边界与实现策略都可由仓库内文档、脚本和最近 memory 直接确定
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / guard / 状态文档主线，不扩散到 AI 页面行为、settings UI、relay 或后端接口
- 本轮必须保留真实代码增量与直接验证，不能只记录“环境阻塞”
- `vite/esbuild spawn EPERM`、Windows networking/service-provider 初始化失败和默认 `apply_patch` wrapper 失效继续按环境阻塞记录，不误归类为产品回归

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 与 `web/src/components/modules/setting/*` 的用户可见行为
- 不改 shared external preflight schema；本轮只增强 learning wrapper 对现有诊断 artifact 的继承方式
- 不新增 repo 外部依赖或新的长期运行进程

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1` 提供 repo-local 稳定诊断副本路径：`build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`
- `check-only` 输出和真实 external 失败摘要都能看到这条稳定副本路径
- `verify-ai-automation-learning-focus.mjs` 守住新的稳定副本输出契约
- `tsc --noEmit`、learning focus guard、learning smoke `check-only`、一次按环境阻塞预期失败的 `external` 复跑和 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-copy-closure.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改 learning wrapper 的 external 诊断副本发布与守护输出
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强 external 失败时的 artifact 继承方式与 check-only 可见性

## 8. 实施步骤

1. 复核 automation memory、canonical plan、当前状态文档与最近两份 H6 worklog，确认本轮仍应围绕 external 预检诊断消费链继续收口。
2. 复查 `verify-ai-automation-learning-browser-smoke.ps1`，确认它虽然已经能打印聚合诊断与 next steps，但 external 诊断仍主要落在临时目录，下一轮不够好继承。
3. 在 learning wrapper 中新增 repo root 解析、稳定副本路径 helper 和 `Publish-ExternalPreflightDiagnosticCopy`，让 external 失败时把 `Diagnostic:` 指向的 JSON 额外复制到 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`。
4. 让 `check-only` 也打印稳定副本目标路径，并把 `Write-ExternalPreflightDiagnosticSummary` 扩展为同时显示临时 artifact 路径与稳定副本路径。
5. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新 helper、稳定副本路径常量和输出契约纳入静态守护。
6. 运行 focus guard、`tsc --noEmit`、learning smoke `check-only`、按环境阻塞预期失败的 `external` 复跑与 `git diff --check`，确认稳定副本路径在 check-only 与真实失败摘要里都已出现。

## 9. 测试与验证

- 构建命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 专项验证：
  - `'$env:APPDATA = Join-Path $env:USERPROFILE ''AppData\Roaming''; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE ''AppData\Local''; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults'`
  - `'$env:APPDATA = Join-Path $env:USERPROFILE ''AppData\Roaming''; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE ''AppData\Local''; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults'`（按环境阻塞预期失败，但确认稳定副本路径已输出）
  - `git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-copy-closure.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 10. 风险与兼容性

- 新风险：低；只新增 repo-local 诊断副本写入与 wrapper 输出，不改 external preflight 结构或产品逻辑
- 兼容性风险：低；稳定副本目录位于 `build/` 下，不会影响运行态配置和已有 smoke 行为
- 是否阻塞下一任务：不阻塞；下一轮可直接从 repo-local 稳定副本继续阅读上一次 external 诊断，而不必先找临时目录

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`git diff --check`；真实 `external -UseHostFriendlyExternalDefaults` 已按环境阻塞预期失败，但稳定副本路径已确认出现在失败摘要中
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、Windows apply_patch fallback 经验
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与连续 worklog 说明上一轮已完成聚合诊断和 next-step 输出，因此本轮最值得推进的是 artifact 继承稳定性；canonical plan 与 workflow 限定本轮不得越界；状态文档确认当前主阻塞仍是宿主 reachability，而不是产品行为；Windows patch fallback 经验保证本轮可以在默认 wrapper 失效时继续可靠落 patch
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 repo-local wrapper 的 external 诊断副本闭环
- 手工 smoke 阻塞原因 / 缺少的环境：backend/frontend 在当前宿主上仍先表现为 `host_networking_blocker`，无法进入真实 external browser 证据阶段
- 待验证页面清单：`AI 自动化` 学习页 external + cdp 真实 browser 证据仍待在健康宿主或可达服务环境补齐
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：是
- 遗留项：
  - 在健康宿主或可达服务环境里，优先重跑 `& .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`，并先看 repo-local 稳定副本再决定是否打开新的临时 artifact
  - 若 backend/frontend 已可达但 CDP 仍失败，继续沿同一稳定副本和 fresh artifact 聚焦 browser/CDP bootstrap，而不是回到 service reachability 分析
  - `vite/esbuild spawn EPERM`、Windows networking/service-provider 初始化问题和默认 `apply_patch` wrapper 失效继续按环境阻塞处理
- 下一任务前置条件是否满足：满足；下一轮可以直接从 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json` 与 wrapper 的 next-step 输出继续推进同一条 Phase H6 主线
