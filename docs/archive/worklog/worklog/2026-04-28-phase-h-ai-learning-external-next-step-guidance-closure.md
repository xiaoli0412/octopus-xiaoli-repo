# 2026-04-28 Phase H AI Learning External Next-Step Guidance Closure

## 1. 任务信息

- 任务名称：phase h ai learning external next-step guidance closure
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-external-diagnostic-consumer-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning browser smoke 验证主线，把 external 失败摘要从“告诉你卡在哪”继续收口成“直接告诉你下一步该怎么复跑”的命令级入口，减少下一轮还要根据 hints 或 JSON 自己拼命令
- 本次已盘点本地资源：AGENTS.md、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / guard / 状态记录主线，不扩散到 UI、relay、API 或新的浏览器基建主题
- 本轮必须保留真实代码增量与直接验证，不能只继续堆诊断描述
- 外部服务不可达、`vitest/esbuild spawn EPERM`、Windows loopback/service-provider 初始化问题继续按环境阻塞记录，不误归类为产品回归

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 与 `web/src/components/modules/setting/*` 的用户可见行为
- 不改动态路由学习数据模型、评分逻辑或 settings 交互
- 不继续扩展 shared CDP wrapper 的 preflight schema；本轮只消费已有诊断结果并补命令级下一步建议

## 5. 本次验收条件

- external 失败时，learning wrapper 会在 `Latest external preflight diagnostic` 后继续打印 `External preflight next steps`
- 下一步建议至少覆盖：
  - 外部 backend/frontend 本应已启动时的标准 external 复跑命令
  - 需要本机 local service 对照时的 `-SelfStartServices` 复跑命令
- `verify-ai-automation-learning-focus.mjs` 守住新的 wrapper 输出契约
- `tsc --noEmit`、learning focus guard、learning smoke `check-only` 和 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/worklog/2026-04-28-phase-h-ai-learning-external-next-step-guidance-closure.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改 learning wrapper 的 external 失败后续建议输出与守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；`self-start`、产品页面与动态路由学习行为保持不变，只增强 external 失败时的命令级指导

## 8. 实施步骤

1. 复核上一轮 external diagnostic consumer closure，确认当前 external 失败已经能打印聚合诊断，但还没有直接给出下一步复跑命令。
2. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增 `Get-ExternalPreflightNextStepLines`，根据 `failedChecks` 输出下一步建议，并把该段接到 `Write-ExternalPreflightDiagnosticSummary` 后面。
3. 更新 `verify-ai-automation-learning-focus.mjs`，把 `External preflight next steps`、标准 external 复跑命令、`-SelfStartServices` 变体和 CDP 分支提示纳入静态守护。
4. 重新执行 learning focus guard、`tsc --noEmit`、learning smoke `check-only`、`git diff --check`，再复跑一次按环境阻塞预期失败的 `external` 命令，确认新输出真实出现。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 已复跑并按预期失败：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`
  - 当前失败仍为 `backend, frontend -> host_networking_blocker`
  - 但在 rethrow 前已先打印 `External preflight next steps`，明确给出标准 external 复跑命令和 `-SelfStartServices` 对照命令
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-external-next-step-guidance-closure.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 未做成自动断言：尝试把 external 运行时 `Write-Host` 输出重定向到 repo-local 日志文件做二次脚本断言时，被当前宿主命令策略拦截；本轮已通过直接 console 复跑确认新输出真实出现

## 10. 风险与兼容性

- 新风险：低；只增强 external 失败时的指导文案与输出结构
- 兼容性风险：低；不改 external preflight 诊断字段，不改 `self-start` 路径，不改产品行为
- 是否阻塞下一任务：不阻塞；下一轮可直接沿新的命令级建议继续跑 healthy-host / reachable-service 对照

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p .\web\tsconfig.json` 已通过
- 测试是否通过：通过 learning focus guard、learning smoke `check-only -UseHostFriendlyExternalDefaults` 与 `git diff --check`；真实 `external -UseHostFriendlyExternalDefaults` 已按环境阻塞预期失败，但新的 next-step 输出已通过 console 复跑确认生效
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、当前状态文档、前端主线状态、详细工作流、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` 技能文档
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未执行真实 UI 手工点击；本轮聚焦 repo-local wrapper 的 external 失败指导输出
- 自动 smoke 阻塞原因 / 缺少的环境：当前宿主上 backend/frontend 外部可达性仍先失败为 `host_networking_blocker`；`vitest/esbuild spawn EPERM` 仍阻塞 jsdom 入口
- worklog 是否更新：是
- 遗留项：
  - 在健康宿主或已暴露服务的环境里，继续优先用 `-Mode external -UseHostFriendlyExternalDefaults` 采集 learning 页真实 browser 证据
  - 若需要本机 local service 对照，直接在同一命令后追加 `-SelfStartServices`，不要再改 wrapper 默认值
  - 若 backend/frontend 已可达但 CDP 仍失败，继续把 follow-up 聚焦到 browser/CDP bootstrap，而不是回到服务 reachability 排查
- 下一任务前置条件是否满足：满足；下一轮可直接从新的 `External preflight next steps` 命令级入口继续推进同一条 Phase H6 验证主线
