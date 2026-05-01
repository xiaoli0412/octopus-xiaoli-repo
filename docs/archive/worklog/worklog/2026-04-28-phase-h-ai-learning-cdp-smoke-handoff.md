# 2026-04-28 Phase H AI Learning CDP Smoke Handoff

## 1. 任务信息

- 任务名称：phase h ai learning cdp smoke handoff
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-browser-smoke-scaffold.md`
- 本次任务目标：把 `AI 自动化` 学习区从 `playwright-cli` 专用 smoke 入口收口到仓库现有共享 `CDP self-start` 包装链，并把宿主 loopback 阻塞从误导性的“找不到空闲端口”改成准确的 host networking blocker
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、DETAILED_EXECUTION_WORKFLOW、Phase H6 连续 worklog、`scripts/verify-channel-create-browser-smoke-cdp.mjs`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-home-layout-browser-smoke.ps1`、`scripts/verify-channel-page-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行执行，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 浏览器证据主线，不扩散到新的 AI task 能力、后端 schema 或设置页新交互
- 只改 learning smoke 场景、共享 smoke wrapper 和对应静态守护，不改业务页面语义
- 验证要优先区分“产品回归”与“宿主阻塞”，不能继续留下误导性的环境报错

## 4. 本次禁止事项

- 不改 `AI 自动化` 页面学习区 UI 语义
- 不改动态路由学习 API 契约
- 不新增新的浏览器自动化基建分支，优先复用现有 `channel/home/settings` 的 CDP 骨架

## 5. 本次验收条件

- `scripts/verify-ai-automation-learning-browser-smoke.ps1` 改为默认接入共享 CDP wrapper，而不是只走 CLI
- 共享 `scripts/verify-channel-create-browser-smoke-cdp.mjs` 新增 `ai-learning` 场景，能覆盖 learning 页加载、样本出现、reset、switch 和 `375px` 宽度断言
- 共享 `scripts/verify-channel-create-browser-smoke-cdp.ps1` 的 Node 解析跳过 Codex 自带 node，并在 loopback service-provider 失败时给出准确 host blocker
- `verify-ai-automation-learning-focus.mjs` 能守住上述入口与场景契约

## 6. 本次改动

- `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - 新增 `aiLearningSmoke*` 种子常量
  - 新增 `ensureAILearningSeedData()`：打开 `dynamic_routing_learning_enabled`、创建最小 channel/group/apikey、打一次真实网关请求，给 learning query 产出样本
  - 新增 `waitForAILearningPageReady()`、`waitForAILearningDataLoaded()`、`waitForAILearningResetState()`、`waitForEvaluation()`
  - `smokeScenario` 分发新增 `ai-learning`
  - `navStorage.activeItem` 新增 `ai`
  - 新增 learning 场景断言：desktop 壳、preset 选择、reset 后空态、switch 切换、mobile `375px`
- `scripts/verify-ai-automation-learning-browser-smoke.ps1`
  - 从 CLI wrapper 改为默认走共享 CDP wrapper
  - 默认透传 `CdpPageBootstrapStrategy=attached-session` 与 `CdpBootstrapCommandOrder=runtime-page-lifecycle`
  - 设置 `OCTOPUS_UI_SMOKE_SCENARIO=ai-learning`
- `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - 对齐 CLI wrapper 的稳定 Node 解析逻辑，跳过 Codex 自带 `node.exe`
  - 新增 Windows service-provider 错误识别，把 loopback bind/probe 失败明确标成 host networking blocker
- `scripts/verify-ai-automation-learning-focus.mjs`
  - 新增对 `ai-learning` CDP wrapper 和共享场景分发的静态守护

## 7. 测试与验证

- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only`
- 通过：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 通过：`git diff --check -- scripts/verify-channel-create-browser-smoke-cdp.mjs scripts/verify-channel-create-browser-smoke-cdp.ps1 scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs`
- 未通过但已正确分类：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
  - 当前失败点已不再是假“端口占用”，而是准确输出：`Loopback TCP probing on 127.0.0.1:18081 is blocked by Windows service-provider initialization`
  - 该失败属于宿主 loopback / service-provider 阻塞，不是 learning smoke 场景回归

## 8. 风险与兼容性

- 新风险：低；主要是共享 CDP 场景脚本继续变长，但本轮没有改业务代码
- 兼容性风险：低；learning smoke 改动只影响验证入口，不影响用户可见行为
- 是否阻塞下一任务：部分阻塞浏览器级真实 pass，但不阻塞同主线继续收口。当前已具备稳定 `check-only`、准确 host blocker 和可复用 `ai-learning` 场景骨架

## 9. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only`、`git diff --check`
- 手工 smoke 状态：未执行
- 自动 smoke 阻塞原因 / 缺少环境：Windows 本机 loopback service-provider 初始化仍然失败，导致 `self-start` 无法完成本地端口 bind；该问题已在共享 CDP wrapper 中被准确分类
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` skill 文档、现有 home/channel/settings smoke 脚本
- worklog 是否更新：是
- 遗留项：
  - 继续尝试在 host-friendly 条件下跑通 `ai-learning self-start + cdp`
  - 若宿主 loopback 仍不可用，则优先复用新的 `ai-learning` 场景做 `external + cdp` 对照，不再回到旧 CLI-only 路径
  - 当前仍遗留 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时文件，因宿主策略阻止直接删除，下一轮若环境允许需优先清理
- 下一任务前置条件是否满足：满足；下一轮可以直接从 `verify-ai-automation-learning-browser-smoke.ps1` 与 `ai-learning` CDP 场景继续推进真实浏览器证据
