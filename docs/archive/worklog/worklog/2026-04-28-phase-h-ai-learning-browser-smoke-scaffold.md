# 2026-04-28 Phase H AI 学习浏览器烟测脚手架收口

## 1. 任务信息

- 任务名称：phase h ai learning browser smoke scaffold
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据铺路

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-settings-learning-runtime-source-alignment.md`
- 本次任务目标：沿着 Phase H6 当前“浏览器证据未闭环”的剩余缺口，为 `AI 自动化` 学习区补稳定 selector、轻量 browser smoke 入口和对应静态守护，减少下一轮再从零拼验证入口的成本
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、DETAILED_EXECUTION_WORKFLOW、Phase H6 连续 worklog、`web/src/components/modules/ai-automation/index.tsx`、`web/src/components/modules/ai-automation/index.test.tsx`、`scripts/verify-ai-automation-learning-focus.mjs`、现有 `backup/settings/home/channel` browser smoke 脚本
- 本次使用的本地 resources / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 browser/no-browser 脚本
- 若未使用部分本地资源或上下文，原因：本轮只做同一 learning 主线下的 smoke 脚手架与 selector 收口，不继续扩展到其它页面或新的 AI task 能力
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 动态路由学习 consumer 与验证主线，不扩散到新的 AI task、后端 schema 或其它页面返工
- 新增 browser smoke 只服务于 `AI 自动化` 学习区，不把 screenshot-first 主线重新发散成大而全的全页 e2e
- 必须同步更新 selector、静态守护、worklog 与 automation memory，避免出现“脚本有了但页面锚点和记录没跟上”的半成品

## 4. 本次禁止事项

- 不改动态路由学习 API 契约
- 不改 `manual / ai_profile` 或 `dynamic_routing_learning_enabled` 的既有运行时语义
- 不因宿主 `vitest/esbuild spawn EPERM` 删除现有 jsdom 测试入口

## 5. 本次验收条件

- `AI 自动化` 学习区补齐稳定的 browser smoke selector，包括 preset、switch、summary、reset 与 states/empty 区
- 新增一条独立的 `AI 自动化` 学习区 browser smoke 脚本与 PowerShell wrapper，至少支持 `check-only`
- `index.test.tsx` 与 `verify-ai-automation-learning-focus.mjs` 同步覆盖新增 selector
- `tsc --noEmit`、`verify-ai-automation-learning-focus.mjs`、新 smoke `--check-only` 与 `git diff --check` 通过

## 6. 本次回滚点

- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/components/modules/ai-automation/index.test.tsx`
- `scripts/verify-ai-automation-learning-focus.mjs`
- `scripts/verify-ai-automation-learning-browser-smoke.mjs`
- `scripts/verify-ai-automation-learning-browser-smoke.ps1`

## 7. 实现范围

- 先改数据语义还是先改 UI：先补学习区 selector 与组件测试，再接 browser smoke 脚本和 wrapper
- 受影响后端模块：无
- 受影响前端模块：`AI 自动化` 学习区 JSX、学习区组件测试、静态验证脚本
- 受影响接口：无新增接口；browser smoke 只复用现有登录、AI config 与 dynamic learning 查询链路
- 是否影响旧数据：否
- 是否影响旧行为：低风险；主要增加 test id 与验证入口，不改变业务交互语义

## 8. 实施步骤

1. 在 `AI 自动化` 学习区补稳定 `data-testid`，覆盖 preset 区、switch 区、secondary summary、reset 按钮和 states/empty 容器。
2. 更新 `index.test.tsx` 与 `verify-ai-automation-learning-focus.mjs`，把新增 selector 纳入回归护栏。
3. 复用现有 CLI browser smoke 模板，新增 `verify-ai-automation-learning-browser-smoke.mjs/.ps1`，先打通 `check-only` 与最小结构校验入口。
4. 运行本轮直接相关验证，补写 worklog、状态记录和 automation memory。

## 9. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-browser-smoke.mjs --check-only`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\runtime-win.ps1 -Action check-only`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\runtime-win.ps1 -Action status`
- 通过：`git diff --check -- web/src/components/modules/ai-automation/index.tsx web/src/components/modules/ai-automation/index.test.tsx scripts/verify-ai-automation-learning-focus.mjs scripts/verify-ai-automation-learning-browser-smoke.mjs scripts/verify-ai-automation-learning-browser-smoke.ps1 docs/worklog/2026-04-28-phase-h-ai-learning-browser-smoke-scaffold.md`
- 未通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 120`，失败点是 `playwright-cli` 启动阶段宿主报 `spawn EPERM`，不是页面 selector 或学习区逻辑回归

## 10. 风险与兼容性

- 新风险：低；主要风险是新 smoke 脚本若依赖过多非 learning 区 DOM，会增加后续维护成本
- 兼容性风险：低；新增 test id 不改变现有用户可见语义
- 是否阻塞下一任务：不阻塞；若脚手架落地，下一轮即可直接补真实 browser pass，而不用再重建入口

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、新 learning smoke 的 `node --check-only` 与 `ps1 check-only`、`runtime-win.ps1 check-only/status`、`git diff --check`；真实 `self-start` 浏览器烟测仍受宿主 `playwright-cli spawn EPERM` 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 `settings/home/channel/backup` smoke 脚本、`runtime-win.ps1`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：automation memory 与连续 Phase H6 worklog 明确上一轮已经把 learning summary consumer 抽象收口完，本轮最值得推进的是浏览器证据脚手架；canonical plan 与 workflow 约束本轮继续停留在 H6 learning 主线；前端主线状态和现有 smoke 脚本提供了可复用的 wrapper/Node smoke 模板；`runtime-win.ps1` 证明当前默认运行策略仍是“项目停驻，按需检查”
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行；仅执行了自动化 smoke 脚手架验证与一次真实 `self-start` 尝试
- 手工 smoke 阻塞原因 / 缺少的环境：真实浏览器路径仍被宿主 `playwright-cli spawn EPERM` 卡住；同时 `runtime-win.ps1 status` 继续报告本机 loopback service-provider 初始化异常，所以这轮不把真实 browser pass 误记为产品回归
- 待验证页面清单：`AI 自动化` 学习区 preset / switch / summary / reset / states 的浏览器级 `375px / click / scroll` 路径
- 若未使用子 agent，原因：用户明确要求主线程串行执行，不创建子 agent
- worklog 是否更新：是
- 遗留项：下一轮需要在同一 Phase H6 主线下继续把新 learning smoke 推进到 host-friendly 的真实浏览器 pass，或在宿主允许时补 `375px / click / scroll` 证据；`vitest/esbuild spawn EPERM` 与 `playwright-cli spawn EPERM` 仍是环境 blocker；默认 `apply_patch` wrapper 仍需靠补齐 `APPDATA/LOCALAPPDATA` 或使用本地 fallback 才能工作
- 下一任务前置条件是否满足：满足；learning 区的 selector、组件测试和 smoke 入口已落地，下一轮可以直接接着做真实 browser pass 或同主线下的轻量 smoke 细化
