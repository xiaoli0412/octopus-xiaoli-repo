# 2026-04-30 Phase G Model Probe Contract And Static Sync Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / setting model-probe contract recovery`
- Current stage: `设置页 Model Probe 契约与静态产物收口`

## 本轮上下文与本地资源

- 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 前端验收清单：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
- canonical plan：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-channel-create-payload-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`
  - `docs/archive/worklog/worklog/2026-04-24-phase-g-model-probe-summary-first-batching-closure.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 Phase G 主线下的增量契约修复与验证收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求主线程串行推进。

## 本轮候选任务

1. 复跑 `channel-create` 真实 self-start smoke，确认上一轮留下的 `channel is disabled` 风险是否仍存在。
2. 若 `channel-create` 已恢复，则切到同阶段设置页相邻缺口，收口 `Model Probe` 的卡内滚动/搜索与 no-browser 契约漂移。
3. 若设置页源码发生变更，同步补 `build:static` 与 `web/out -> static/out`，避免运行页滞后于源码。

## 本轮计划

- 本轮核心任务：先确认 `channel-create` 真实运行态是否仍异常；若已恢复，则完成 `Model Probe` 当前实现、状态文档与 no-browser 守护的统一契约收口。
- 本轮配套任务：对设置页改动补 `build:static`，同步 `web/out -> static/out`。
- 预期验证方式：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
  - `node .\scripts\verify-model-probe-help.mjs`
  - 等价 `settings-no-browser` Node 守护链
  - `node .\scripts\build-web-static.mjs`
  - `git diff --check -- web/src/components/modules/setting/ModelProbe.tsx scripts/verify-model-probe-help.mjs docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 完成判定标准：
  - `channel-create` 真实 self-start 不再复现上一轮担心的运行态失败。
  - `Model Probe` 重新具备默认路径摘要、稳定搜索/滚动锚点与 `12` 条分批默认值。
  - no-browser 守护、静态构建与静态产物同步完成。

## 本轮硬规则

- 只在 `Phase G` 主线内推进，不扩散到无关 UI 改造。
- 若 `channel-create` 已恢复，则不继续空转追同一现象，而是切到同阶段相邻小闭环。
- 对设置页只动 `ModelProbe` 与其守护/记录文件，不重开其他卡片结构返工。

## 执行与结果

1. 先复核 automation memory、当前主线文档、前端验收清单、前端主线状态和上一轮 `channel-create` worklog，确认当前最佳入口仍是 `Phase G screenshot-first UI closure`。
2. 按计划先跑真实 `channel-create` self-start 烟测，结果通过；本轮未再复现上一轮遗留的 `channel is disabled` 风险，因此没有继续在同一路径空转排查。
3. 转入同阶段相邻的设置页 `Model Probe` 小闭环后，发现一处真实契约漂移：
  - `scripts/verify-model-probe-help.mjs` 已回退到要求 `defaultPathTitle/defaultPathHint` 存在于组件源码的旧断言；
  - 但当前 `ModelProbe.tsx` 实际已经不再渲染 `defaultPath` 摘要，导致源码、locale、状态文档和 no-browser 守护之间重新失配；
  - 同时当前组件的默认分批值也退回到 `6`，与 `2026-04-24` 的状态文档口径 `12` 不一致。
4. 没有只改脚本掩盖漂移，而是把组件和守护一起重新收口到同一口径：
  - `ModelProbe.tsx` 恢复 `defaultPathTitle/defaultPathDesc/defaultPathHint` 摘要区，并把默认策略 badge 放回首屏；
  - 为搜索框、默认路径区、展开开关、卡内滚动区、模型列表、折叠态与空态补稳定 `data-testid`；
  - 卡内滚动区补 `overscroll-contain`，继续保持“卡内滚动，不拉长整页”的验收要求；
  - `DEFAULT_VISIBLE_MODEL_COUNT` 调整回 `12`，与前端状态文档恢复一致；
  - `verify-model-probe-help.mjs` 同步改成验证当前真实契约，而不是只盯着过时结构。
5. 因本轮改动属于真实页面源码变更，继续补跑了 `build:static`，确认 `web/out` 已重新导出并同步到 `static/out`。
6. 在验证统一入口时又发现宿主命令层问题：旧 worklog 里使用的 `D:\gol1\corepack.cmd` 在当前机器不存在；改用当前 Node 安装里的 `corepack` 入口后，`pnpm run test:settings-no-browser` 仍会因为本机 `C:\Users\李昊桐\AppData\Local\pnpm\config\rc` 的 `EPERM` 与 `pnpm` 自调用 PATH 解析失败而中断。
7. 没把这个宿主问题当成代码 blocker，而是直接改用等价的逐条 Node 守护链，按 `test:settings-no-browser` 的实际内容把 `backup/dynamic-routing/circuit-breaker/model-probe/setting-info/ai-config-profile/ai-learning` 全部跑绿，保住本轮验证闭环。

## 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-model-probe-help.mjs`
- passed `. .\scripts\use-node-env.ps1; foreach (...) { node verify-backup-logic / verify-backup-component / verify-help-hint-accessible / verify-dynamic-routing-help / verify-circuit-breaker-help / verify-model-probe-help / verify-setting-info-logic / verify-ai-config-profile-summary / verify-ai-automation-learning-focus }`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\build-web-static.mjs`
- passed `git diff --check -- web/src/components/modules/setting/ModelProbe.tsx scripts/verify-model-probe-help.mjs docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- observed host command blocker `pnpm run test:settings-no-browser` via Node corepack path still fails on this machine because `pnpm` self-invocation hits `C:\Users\李昊桐\AppData\Local\pnpm\config\rc` `EPERM` and then loses the nested `pnpm` command on PATH`

## 本轮变更文件

- `web/src/components/modules/setting/ModelProbe.tsx`
- `scripts/verify-model-probe-help.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-model-probe-contract-and-static-sync-closure.md`
- `web/out`（构建产物）
- `static/out`（同步后的嵌入产物）

## 未完成 / 风险 / 阻塞

- `Model Probe` 的浏览器级桌面端与 `375px` 真实手感证据还没在本轮补跑；当前完成的是源码、no-browser 契约和静态产物闭环。
- 宿主上的 `pnpm` 统一脚本入口仍受 `pnpm config rc EPERM` 与嵌套命令 PATH 解析问题影响；当前已用等价 Node 守护链绕过，不构成本轮代码 blocker，但后续若要恢复统一入口，需单独修宿主环境。
- 本轮没有扩散到其他设置卡片；如果下一轮继续 settings，同样应优先挑一个小闭环，而不是重开整页返工。

## 下一轮候选任务顺序

1. 延续 `Phase G`，补设置页 `Model Probe` 的浏览器级桌面端 / `375px` 手感证据，重点确认默认路径摘要、卡内滚动与搜索输入在真实页面下仍稳定。
2. 若 settings 浏览器证据仍受宿主阻塞，则留在同主线修统一验证入口，优先解决 `pnpm run test:settings-no-browser` 的宿主命令链，而不是再改业务组件。
3. 若 settings 证据闭环已够，再切到同阶段相邻页面的剩余截图优先缺口，但继续保持“一轮一小闭环”。
