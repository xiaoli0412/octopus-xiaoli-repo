# 2026-04-29 Phase H AI learning stable freshness threshold closure

## 1. 任务信息

- 任务名称：Phase H6 learning stable freshness threshold closure
- 日期：2026-04-29
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由 AI 学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-29-phase-h-ai-learning-refresh-command-closure.md`
- 本次任务目标：继续停留在同一条 Phase H6 learning smoke stable diagnostic 主线，把 `check-only` 从“显示 checked at / age，由接手人自己判断是否过旧”收口成“直接给出 fresh/stale 结论”，减少下一轮健康宿主 external 复跑前的人工换算成本
- 本次已盘点本地资源：`AGENTS.md`、automation memory、`docs/LLM-Gateway-Refactor-Plan.zh-CN.md`、`docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`、`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`、`docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、最近 Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、repo-local stable diagnostic 目录
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 H6 worklog、`using-superpowers` / `brainstorming` skill 文档、repo-local stable diagnostic 副本
- 若未使用部分本地资源或上下文，原因：本轮不需要重新展开产品 UI、后端 schema 或宿主网络修复；现有 H6 连续 worklog 已明确把剩余空间收敛为 stable replay 可消费性
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：Codex 主线程自动化链路串行执行
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke wrapper / guard / 状态文档主线，不扩散到 AI 页面行为、settings UI、relay/runtime 学习逻辑或宿主网络修复主题
- 本轮必须形成真实代码增量与直接验证，不能只复述环境阻塞或只更新文档
- 只增强 stable replay 的 freshness 可读性，不改 external preflight schema 与失败分类语义

## 4. 本次禁止事项

- 不修改 `web/src/components/modules/ai-automation/*` 或 `web/src/components/modules/setting/*` 的用户可见行为
- 不扩大到真实 external/browser 联调
- 不把 `host_networking_blocker`、`vite/esbuild spawn EPERM` 或宿主 socket provider 初始化失败误记成产品回归

## 5. 本次验收条件

- `verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only` 会在 stable copy 状态行中打印 `fresh against 24h threshold` 或 `stale against Nh threshold`
- 可通过 `-StableDiagnosticFreshnessThresholdHours` 临时收紧阈值，以便明确打出 stale 分支
- `verify-ai-automation-learning-focus.mjs` 覆盖新增 helper、参数与输出契约
- 默认阈值与 `1h` 阈值两次 `check-only`、`tsc --noEmit` 与 `git diff --check` 通过

## 6. 本次回滚点

- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-29-phase-h-ai-learning-stable-freshness-threshold-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：不改产品数据与 UI，只改 learning smoke wrapper 的 stable replay freshness 结论与静态守护
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 `check-only` 回放里的 freshness 结论，不改变 external 失败分类与推荐命令逻辑

## 8. 实施步骤

1. 复核 automation memory、canonical plan、用户上下文总账、详细工作流、当前状态文档、前端主线状态和最近 H6 worklog，确认上一轮已完成 preferred refresh command，本轮最值得推进的是把 `age` 再收口成可直接消费的 fresh/stale 结论。
2. 复查 repo-local `build/verify-ai-automation-learning/*` 副本，确认当前稳定副本时间仍在 24 小时内，适合用默认阈值与压缩阈值对照验证 fresh/stale 两条分支。
3. 在 `verify-ai-automation-learning-browser-smoke.ps1` 中新增 `StableDiagnosticFreshnessThresholdHours` 参数和 freshness helper，基于 `checkedAt` 与阈值计算 stable copy 的 `fresh/stale` 标签。
4. 让 `matching / alternate / legacy stable diagnostic copy status` 行统一追加 freshness 结论，默认按 24 小时阈值输出。
5. 同步更新 `verify-ai-automation-learning-focus.mjs`，把新增 helper、参数和输出字符串纳入静态守护。
6. 运行默认阈值与 `1h` 阈值两次 learning smoke `check-only`，确认同一 stable copy 能分别打出 `fresh against 24h threshold` 与 `stale against 1h threshold`。
7. 更新当前状态文档、前端主线状态、automation memory 与本 worklog，给下一轮留下“先看 freshness 结论，再决定是否去健康宿主刷新”的明确入口。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults -StableDiagnosticFreshnessThresholdHours 1`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/worklog/2026-04-29-phase-h-ai-learning-stable-freshness-threshold-closure.md`

## 10. 风险与兼容性

- 新风险：低；fresh/stale 只作用于 repo-local stable replay 的说明层，不改变 fresh artifact 本身，也不影响 external 失败分类
- 兼容性风险：低；若 `checkedAt` 不可解析，wrapper 仍会回退成只打印 `checked at / age` 或更基础的状态标签，不会阻断 `check-only`
- 是否阻塞下一任务：不阻塞；下一轮可以先根据 freshness 结论判断当前 stable copy 是否已值得刷新，再决定是否去健康宿主执行 external/browser 复跑

## 11. 收工记录

- 构建是否通过：通过（前端 `tsc --noEmit`）
- 测试是否通过：通过（learning focus 守护、默认阈值与 `1h` 阈值两次 `check-only`、`git diff --check`）
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、主规划、用户上下文总账、详细工作流、环境 next plan、当前状态文档、前端主线状态、最近 H6 worklog、`using-superpowers` / `brainstorming` skill 文档、repo-local stable diagnostic 目录
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与最近 H6 worklog 说明上一轮已完成 preferred refresh command，因此本轮最值得推进的是把 `age` 升级成 freshness 结论；stable diagnostic 目录显示当前 optional/legacy 副本仍在、`requireCdp=true` 变体仍缺失，适合做 freshness 对照验证；状态文档与前端主线状态确认当前剩余主缺口仍是健康宿主上的真实 external/browser 证据
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：未做真实 external/browser；继续使用 repo-local `check-only` 收口 stable diagnostic 入口
- 手工 smoke 阻塞原因 / 缺少的环境：当前宿主 external 真实服务仍是 `host_networking_blocker`；`vite/esbuild spawn EPERM` 仍阻塞真实 jsdom/vitest 入口
- 待验证页面清单：健康宿主上的 `AI 自动化` learning external/browser 证据，尤其是 `requireCdp=true` fresh artifact
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：已更新
- 遗留项：
  - 仍需在健康宿主或可达 backend/frontend 的环境中执行 replay 给出的 `preferred refresh command`，补一份真正的 `requireCdp=true` fresh external artifact
  - 当前 optional/legacy 副本在默认 `24h` 阈值下仍显示 fresh，因此下一轮如果距离上次 external 更久，可先用默认 freshness 结论判断是否需要刷新，再决定是否上健康宿主
- 下一任务前置条件是否满足：满足；下一轮优先继续同一条 Phase H6 诊断链，先看 stable replay freshness 结论，再决定是否去健康宿主补 fresh external/browser 证据
