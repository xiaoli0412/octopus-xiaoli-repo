# 2026-04-28 Phase H AI learning stable diagnostic freshness closure

## 1. 任务信息

- 任务名称：phase h ai learning stable diagnostic freshness closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-check-only-stable-preview-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 repo-local stable diagnostic 的 `check-only` 回放从“能看摘要”再收口到“能判断是不是旧结果”，避免下一轮把历史 external 失败副本误当成当前 live probe 结果
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`、`docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local 稳定诊断副本 `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、using-superpowers / brainstorming skill 文档、repo-local stable diagnostic 副本
- 若未使用部分本地资源或上下文，原因：未使用外部资料；当前任务边界、阻塞分类和最小改动点都能由仓库内文档、脚本与 stable diagnostic 直接确定
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / guard / 状态文档主线，不扩散到 AI 页面行为、settings UI、relay 或 shared preflight schema
- 本轮必须保留真实代码增量与直接验证，不能只复述环境阻塞
- `vite/esbuild spawn EPERM`、Windows networking/service-provider 初始化失败和默认 `apply_patch` wrapper 失效继续按环境阻塞记录，不误归类为产品回归

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 与 `web/src/components/modules/setting/*` 的用户可见行为
- 不改 shared external preflight 的 JSON schema；本轮只增强 learning wrapper 对已有 stable diagnostic 的消费提示
- 不新增外部依赖或新的长期运行进程

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults` 在 stable diagnostic 存在时，会额外打印 `diagnostic source / checked at / diagnostic age`
- `check-only` 回放会明确提示这是“最近一次 external 失败保存结果”，不是 live probe
- `verify-ai-automation-learning-focus.mjs` 守住新增的 helper 与输出契约
- `tsc --noEmit`、learning focus guard、learning smoke `check-only` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-freshness-closure.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改 learning wrapper 的 stable diagnostic 回放元信息与守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强 `check-only`/诊断摘要的可读性与交接稳定性

## 8. 实施步骤

1. 复核 automation memory、canonical plan、当前状态文档、前端主线状态与最近 H6 worklog，确认本轮仍应围绕 learning smoke wrapper 的验证入口收口推进。
2. 读取 repo-local `build/verify-ai-automation-learning/latest-external-preflight-diagnostic.json`，确认现有副本已含 `checkedAt`，且当前稳定分类仍为 `backend + frontend -> host_networking_blocker`，因此最值得推进的是“提示这份摘要的时间/来源”，而不是继续扩展失败分类。
3. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增 stable diagnostic 元信息 helper，给摘要补上 `diagnostic source / checked at / diagnostic age`，并在 `check-only` 稳定副本回放时明确提示“这不是 live probe”。
4. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新增 helper 和输出契约纳入静态守护。
5. 更新当前状态文档、前端主线状态与本 worklog，保证下一轮可直接继承本轮结论。
6. 运行 learning focus guard、`tsc --noEmit`、learning smoke `check-only` 与 `git diff --check`，确认新增元信息输出稳定可用。

## 9. 测试与验证

- 运行：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 运行：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 运行：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 运行：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-stable-diagnostic-freshness-closure.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 10. 风险与兼容性

- 新风险：低；只增强 stable diagnostic 回放的元信息与提示，不改 external preflight 结构或产品逻辑
- 兼容性风险：低；若 `checkedAt` 无法解析，wrapper 仍会回退为只打印原始诊断摘要，不影响现有失败分类
- 是否阻塞下一任务：不阻塞；下一轮可以先用 `check-only` 判断 stable diagnostic 是否已过旧，再决定是否需要在健康宿主上重跑 external/browser 证据

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only -UseHostFriendlyExternalDefaults` 与 `git diff --check`
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、详细工作流、AI 自动化/动态路由实施计划、Phase H6 连续 worklog、using-superpowers / brainstorming skill 文档、repo-local stable diagnostic 副本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 H6 worklog 说明上轮已完成 stable preview 回放，因此本轮最值得推进的是“减少误读旧副本”的提示；repo-local stable diagnostic 直接证明根 JSON 已有 `checkedAt` 可复用；canonical plan 与实施计划限定本轮不得越界到产品 UI 或后端行为
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 wrapper 的 stable diagnostic 元信息与交接可读性
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 loopback / networking 初始化仍阻塞真实 external/self-start browser 证据
- 待验证页面清单：`AI 自动化` 学习页 external + cdp 真实 browser 证据仍待健康宿主或可达服务环境补齐
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：是
- 遗留项：
  - 在健康宿主或可达服务环境里，继续优先跑 `& .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults` 收集真实 browser 证据
  - 若 backend/frontend 已可达但 CDP 仍失败，继续沿 stable copy 与 fresh artifact 聚焦 browser/CDP bootstrap，而不是回到 service reachability 分析
  - `vite/esbuild spawn EPERM`、Windows networking/service-provider 初始化问题与默认 `apply_patch` wrapper 失效继续按环境阻塞处理
- 下一任务前置条件是否满足：满足；下一轮可以先用 `check-only` 回放里新增的 `source / checked at / age` 判断是否需要重跑 external，而不必先重新摸索当前 stable diagnostic 的新旧程度
