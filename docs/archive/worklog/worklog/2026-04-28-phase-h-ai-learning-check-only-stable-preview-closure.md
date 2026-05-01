# 2026-04-28 Phase H AI learning check-only stable preview closure

## 1. 任务信息

- 任务名称：phase h ai learning check-only stable preview closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-copy-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 `check-only` 从“只打印稳定副本路径”收口成“直接回放最近一次 repo-local 稳定诊断摘要”，并顺手清理这条主线残留的 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时补丁文件，减少下一轮接手摩擦
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local 稳定诊断副本 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、using-superpowers / brainstorming skill 文档、Windows codex-run-as-apply-patch fallback 经验
- 若未使用部分本地资源或上下文，原因：未使用外部资料；当前任务边界、验证入口和阻塞分类都可由仓库内文档、脚本与稳定诊断副本直接确定
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / guard / 状态文档主线，不扩散到 AI 页面行为、settings UI、relay 或后端接口
- 本轮必须保留真实代码增量与直接验证，不能只复述环境阻塞
- `vite/esbuild spawn EPERM`、Windows networking/service-provider 初始化失败和默认 `apply_patch` wrapper 失效继续按环境阻塞记录，不误归类为产品回归

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 与 `web/src/components/modules/setting/*` 的用户可见行为
- 不改 shared external preflight schema；本轮只增强 learning wrapper 对现有 repo-local 稳定副本的 `check-only` 消费方式
- 不新增外部依赖或新的长期运行进程

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults` 在稳定副本存在时，会直接打印 `Latest external preflight diagnostic` 摘要，而不是只打印路径
- 稳定副本缺失或损坏时，wrapper 会输出明确的下一步提示
- `verify-ai-automation-learning-focus.mjs` 守住新的 `check-only` 输出契约
- 历史 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时补丁文件清理完成
- `runtime-win status`、`tsc --noEmit`、learning focus guard、learning smoke `check-only` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/worklog/2026-04-28-phase-h-ai-learning-check-only-stable-preview-closure.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改 learning wrapper 的 `check-only` 稳定副本消费与守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强 `check-only` 的诊断复用能力，并清理同主线残留临时文件

## 8. 实施步骤

1. 复核 automation memory、canonical plan、当前状态文档、AI 自动化/动态路由实施计划与最近 H6 worklog，确认本轮仍应围绕 learning smoke wrapper 的验证入口收口推进。
2. 读取 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`，确认当前 repo-local 稳定副本已存在且仍指向 `backend + frontend -> host_networking_blocker`，因此 `check-only` 最值得改成“直接回放摘要”而不是继续扩展 external 失败文案。
3. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Write-StableExternalPreflightDiagnosticPreview`，让 `check-only` 模式优先回放 repo-local 稳定副本；若缺失或损坏，则明确提示“先跑 external 种子”或“直接检查 JSON”。
4. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新 helper、缺失/损坏提示和 `check-only -> Write-StableExternalPreflightDiagnosticPreview` 约束写进静态守护。
5. 清理 repo 根目录残留的 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时补丁文件，收束上一阶段遗留的 workspace 噪音。
6. 运行 `runtime-win status`、focus guard、`tsc --noEmit`、learning smoke `check-only` 和 `git diff --check`，确认新的 `check-only` 输出契约与工作区清理都成立。

## 9. 测试与验证

- 运行：`& .\scripts\runtime-win.ps1 -Action status`
- 运行：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 运行：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 运行：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 运行：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-check-only-stable-preview-closure.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 10. 风险与兼容性

- 新风险：低；只增强 `check-only` 对稳定副本的消费方式，不改 external preflight 结构或产品逻辑
- 兼容性风险：低；稳定副本路径与 external 失败摘要保持兼容，缺失/损坏时仅新增更明确提示
- 是否阻塞下一任务：不阻塞；下一轮可直接通过 `check-only` 回看 repo-local 稳定诊断，再决定是否需要在健康宿主上重跑 external/browser 证据

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 `runtime-win status`、`verify-ai-automation-learning-focus.mjs`、learning smoke `check-only -UseHostFriendlyExternalDefaults` 与 `git diff --check`
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、using-superpowers / brainstorming skill 文档、Windows codex-run-as-apply-patch fallback 经验
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 worklog 说明上一轮已完成 stable copy 路径收口，因此本轮最值得推进的是 `check-only` 的直接消费；实施计划与 canonical plan 限定本轮不得越界；repo-local 稳定诊断副本直接证明当前仍是 `backend + frontend -> host_networking_blocker`；Windows patch fallback 经验保证本轮仍可稳定落 patch
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 wrapper 的 repo-local 诊断复用与工作区清理
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 loopback / networking 初始化仍阻塞真实 external/self-start browser 证据
- 待验证页面清单：`AI 自动化` 学习页 external + cdp 真实 browser 证据仍待健康宿主或可达服务环境补齐
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：是
- 遗留项：
  - 在健康宿主或可达服务环境里，继续优先跑 `& .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 收集真实 browser 证据
  - 若 backend/frontend 已可达但 CDP 仍失败，继续沿 repo-local 稳定副本与 fresh artifact 聚焦 browser/CDP bootstrap，而不是回到 service reachability 分析
  - `vite/esbuild spawn EPERM`、Windows networking/service-provider 初始化问题与默认 `apply_patch` wrapper 失效继续按环境阻塞处理
- 下一任务前置条件是否满足：满足；下一轮可以直接通过 `check-only` 回看稳定副本，再决定是否需要重跑 external 或切换到健康宿主继续同一条 Phase H6 主线
