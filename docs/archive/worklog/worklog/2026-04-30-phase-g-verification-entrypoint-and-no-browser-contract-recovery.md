# 2026-04-30 Phase G Verification Entrypoint And No-Browser Contract Recovery

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / verification entrypoint recovery`
- Current stage: `设置页统一验证入口恢复，并顺手收口同池 no-browser 契约漂移`

## 本轮上下文与本地资源

- 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 前端验收清单：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
- canonical plan：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-probe-contract-and-static-sync-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-channel-create-payload-alignment.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 Phase G 主线下的验证编排修复与小范围契约收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求主线程串行推进，不创建子 agent。

## 本轮候选任务

1. 修复 `test:settings-no-browser` 的宿主敏感命令链，避免每次都卡在嵌套 `pnpm/corepack`。
2. 若统一入口恢复，再把 `test:screenshot-no-browser` 跑起来，暴露并收口同池相邻页面的真实 no-browser 漂移。
3. 若仍有时间，继续只在同一 screenshot-first 池内推进一个额外小闭环，不扩散到无关页面或后端。

## 本轮计划

- 本轮核心任务：恢复 `settings-no-browser` 与 `screenshot-no-browser` 的统一可执行入口，不再依赖宿主敏感的嵌套 `pnpm run ...` 自调用。
- 本轮配套任务：对统一入口恢复后暴露出的首页、渠道卡片与模型页 no-browser 契约漂移做最小修复，保证主入口重新跑绿。
- 预期验证方式：
  - `. .\scripts\use-node-env.ps1; pnpm --dir web run test:settings-no-browser`
  - `. .\scripts\use-node-env.ps1; pnpm --dir web run test:screenshot-no-browser`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
  - `git diff --check -- ...`
- 完成判定标准：
  - `settings-no-browser` 不再卡在嵌套 `pnpm` / `corepack` 入口。
  - `screenshot-no-browser` 能在当前主机重新跑通。
  - 本轮收口只停留在 Phase G 同池验证入口与相邻 no-browser 契约，不扩散到无关主题。

## 本轮硬规则

- 只在 `Phase G` screenshot-first 主线内推进，不扩散到后端、数据库或无关页面返工。
- 优先修验证编排层，再做最小页面契约收口；不为了让脚本通过而放宽门禁。
- 不回退用户已有脏区，不清理无关改动，不顺手大改设置页业务组件。

## 本轮执行与结果

1. 先复核 automation memory、当前主线文档、前端验收清单、前端主线状态和上一轮 `Model Probe` worklog，确认当前最值得推进的是“恢复统一验证入口”，而不是继续在设置页业务组件上空转。
2. 读 `web/package.json`、`scripts/use-node-env.ps1` 和现有 worklog 后确认根因有两层：
  - `test:settings-no-browser` 与 `test:screenshot-no-browser` 通过嵌套 `pnpm run ...` 自调用放大了 Windows 宿主上的 `pnpm config rc EPERM` 风险；
  - 当前 Node/`corepack` 默认又试图使用宿主不可写的 `D:\DevCache\node\corepack`，使得统一入口即便外层命令能起，也会在缓存层再次卡住。
3. 没有继续接受“逐条 Node 命令替代统一入口”的长期漂移，而是直接恢复仓库内统一入口：
  - 在 [`scripts/use-node-env.ps1`](D:/GPT-codex/octopus_repo/scripts/use-node-env.ps1) 里补上 repo-local `COREPACK_HOME / PNPM_HOME / XDG_CONFIG_HOME / npm_config_userconfig`，并定义显式 `pnpm` PowerShell 函数，固定走 `node + corepack.js pnpm`，绕开坏掉的 shim 和不可写宿主缓存。
  - 在 [`scripts/run-frontend-verification-suite.mjs`](D:/GPT-codex/octopus_repo/scripts/run-frontend-verification-suite.mjs) 新增仓库级聚合执行器，用单次 Node 进程串行驱动 `settings` / `screenshot` 两套 no-browser 套件。
  - 将 [`web/package.json`](D:/GPT-codex/octopus_repo/web/package.json) 里的 `test:settings-no-browser` 与 `test:screenshot-no-browser` 从嵌套 `pnpm run ...` 改成调用该聚合执行器。
  - 在 [`.gitignore`](D:/GPT-codex/octopus_repo/.gitignore) 补上 `.tmp-tooling/`，避免 repo-local `corepack/pnpm` 缓存污染工作区。
4. 统一入口恢复后，`settings-no-browser` 已在当前主机直接跑绿；随后继续按“第六步衔接下一轮”的要求让 `screenshot-no-browser` 往下跑，顺手收口三处真实相邻漂移：
  - 首页：[`web/src/components/modules/home/index.tsx`](D:/GPT-codex/octopus_repo/web/src/components/modules/home/index.tsx)、[`PageWrapper.tsx`](D:/GPT-codex/octopus_repo/web/src/components/common/PageWrapper.tsx) 与 [`scripts/verify-home-layout.mjs`](D:/GPT-codex/octopus_repo/scripts/verify-home-layout.mjs) 重新对齐到“首页单栏例外 + 仍保留 browser smoke 锚点”的当前主线口径；补回 `home-page / home-main-grid / home-total-section / home-stats-chart-section / home-rank-section / home-activity-section` 等稳定锚点，而没有把首页强拉回旧双栏。
  - 渠道页卡片层：[`web/src/components/modules/channel/Card.tsx`](D:/GPT-codex/octopus_repo/web/src/components/modules/channel/Card.tsx) 与 [`web/src/components/modules/channel/index.tsx`](D:/GPT-codex/octopus_repo/web/src/components/modules/channel/index.tsx) 恢复 key 数、模式/策略 badge、`channel-page` 根锚点、`channel-card-*` / `channel-detail-dialog-*` 等 `channel-presentation` 合同；只动总页卡片层，不扩散到详情逻辑重写。
  - 模型页：[`scripts/verify-llm-price-boundary.mjs`](D:/GPT-codex/octopus_repo/scripts/verify-llm-price-boundary.mjs) 把旧的两栏断言修正为当前 `UI_MAINLINE_TASK_2026-04-30` 规定的桌面端三栏口径，避免 verifier 落后于当前主线文档。
5. 上述三步都以“统一入口恢复后真实跑出来的失败”为驱动，没有额外全仓扫描，也没有为了多做改动而扩散到不相关页面。

## 验证

- passed `. .\scripts\use-node-env.ps1; pnpm --dir web run test:settings-no-browser`
- passed `. .\scripts\use-node-env.ps1; pnpm --dir web run test:screenshot-no-browser`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-home-layout.mjs`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-channel-presentation.mjs`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-llm-price-boundary.mjs`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- passed `git diff --check -- scripts/use-node-env.ps1 scripts/run-frontend-verification-suite.mjs web/package.json .gitignore web/src/components/common/PageWrapper.tsx web/src/components/modules/home/index.tsx web/src/components/modules/home/total.tsx web/src/components/modules/home/activity.tsx web/src/components/modules/home/chart.tsx web/src/components/modules/home/rank.tsx scripts/verify-home-layout.mjs web/src/components/modules/channel/Card.tsx web/src/components/modules/channel/index.tsx scripts/verify-llm-price-boundary.mjs`

## 本轮变更文件

- `scripts/use-node-env.ps1`
- `scripts/run-frontend-verification-suite.mjs`
- `web/package.json`
- `.gitignore`
- `web/src/components/common/PageWrapper.tsx`
- `web/src/components/modules/home/index.tsx`
- `web/src/components/modules/home/total.tsx`
- `web/src/components/modules/home/activity.tsx`
- `web/src/components/modules/home/chart.tsx`
- `web/src/components/modules/home/rank.tsx`
- `scripts/verify-home-layout.mjs`
- `web/src/components/modules/channel/Card.tsx`
- `web/src/components/modules/channel/index.tsx`
- `scripts/verify-llm-price-boundary.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-verification-entrypoint-and-no-browser-contract-recovery.md`

## 未完成 / 风险 / 阻塞

- 本轮恢复的是 no-browser 聚合入口与同池静态契约，不等于浏览器级桌面端 / `375px` 手感证据已经全部补齐；手工 checklist 与 page-level browser smoke 仍需后续继续做。
- `use-node-env.ps1` 现在把 `corepack/pnpm` 缓存收口到仓库内 `.tmp-tooling/`；这是为当前宿主环境稳定性做的最小工程化修复，不影响生产运行，但后续若团队有统一 Node shell 入口，可以考虑把这一层抽成更通用的 shared helper。
- 统一入口恢复后，后续同池若再出现失败，更可能是页面契约真实漂移，而不是命令链本身失效；下一轮应继续保持“先跑统一入口，再修单一闭环失败项”的顺序。

## 下一轮候选任务顺序

1. 继续 `Phase G`，优先补 browser-grade 手感证据，尤其 settings / model / channel 相邻页面的 `375px / hover / focus` 缺口。
2. 若同池 no-browser 聚合再次暴露新的单点契约漂移，保持本轮做法：只修当前失败项，不重开整页返工。
3. 若浏览器级 smoke 继续受宿主 Edge/CDP 或 Playwright CLI 阻塞，则沿已有 wrapper 路径收口宿主分类，不要再回退到手动逐条执行 no-browser 脚本的临时做法。
