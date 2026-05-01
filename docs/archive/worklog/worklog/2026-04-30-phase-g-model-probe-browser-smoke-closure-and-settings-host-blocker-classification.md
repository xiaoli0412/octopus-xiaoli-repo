# 2026-04-30 Phase G Model Probe Browser Smoke Closure And Settings Host Blocker Classification

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / setting model-probe browser evidence`
- Current stage: `Model Probe browser-grade smoke contract closure + settings host blocker classification`

## 本轮上下文与本地资源

- 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 前端验收清单：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
- canonical plan：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-probe-contract-and-static-sync-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-verification-entrypoint-and-no-browser-contract-recovery.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-backup-cli-host-blocker-classification-closure.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 Phase G 主线下的验证脚本收口与宿主 blocker 分类，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求主线程串行推进，不创建子 agent。

## 本轮候选任务

1. 给 settings browser smoke 补上 `Model Probe` 的真实交互断言，而不是只验证卡片存在。
2. 若真实 browser smoke 仍被宿主拦住，先把失败语义收口成明确 host blocker，再停止空转。
3. 保持 `test:settings-no-browser`、`verify-model-probe-help.mjs` 与 `tsc --noEmit` 继续为绿，不破坏已恢复的统一入口。

## 本轮计划

- 本轮核心任务：让 settings browser smoke 真实验证 `Model Probe` 的默认折叠、搜索、分批与卡内滚动。
- 本轮配套任务：若 browser smoke 仍因宿主问题失败，收口 settings wrapper 的 Node/Playwright host blocker 分类。
- 预期验证方式：
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-setting-help-browser-smoke.mjs --check-only`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-setting-help-browser-smoke-cdp.mjs --check-only`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-model-probe-help.mjs`
  - `. .\scripts\use-node-env.ps1; pnpm --dir web run test:settings-no-browser`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`
  - `git diff --check -- scripts\verify-setting-help-browser-smoke.ps1 scripts\verify-setting-help-browser-smoke.mjs scripts\verify-setting-help-browser-smoke-cdp.mjs`
- 完成判定标准：
  - settings browser smoke 已包含 `Model Probe` 关键交互断言；
  - 当前宿主若仍失败，失败信息必须明确指向 `CDP bootstrap blocker` 或 `Playwright CLI spawn EPERM`，不再误报 success marker 漂移。

## 本轮硬规则

- 只在 `Phase G` screenshot-first settings 池内推进，不扩散到无关页面返工。
- 优先改验证脚本，不重开 `ModelProbe.tsx` 业务结构。
- 本轮不因宿主阻塞回到宽泛分析；失败必须被明确分类后再收束。

## 执行与结果

1. 先复核 automation memory、当前主线文档、前端验收清单、前端主线状态和最近两份 Phase G worklog，确认本轮最佳入口仍是“settings browser-grade 证据补洞”，不是再改业务卡片结构。
2. 核对现有 `scripts/verify-setting-help-browser-smoke.mjs` 与 `...cdp.mjs` 后确认它们对 `Model Probe` 只验证“卡片存在 + help hint 可聚焦 + 375px 不溢出”，没有验证当前最关键的真实交互：
   - 默认折叠占位是否仍在；
   - 搜索是否能按 canonical name 命中；
   - `show more` 是否按 `12` 条分批；
   - 长列表是否仍留在卡内滚动区。
3. 没有回到页面源码做结构改动，而是直接收紧 smoke：
   - 在 `scripts/verify-setting-help-browser-smoke.mjs` 中新增模型 seed、后台 `/api/v1/model/create` 数据准备、`Model Probe` 默认折叠/展开/搜索/empty/show-more/scroll-region 断言；
   - 在 `scripts/verify-setting-help-browser-smoke-cdp.mjs` 做同口径补齐，保证 CLI 与 CDP 两条 settings browser smoke 不再只有一条覆盖真实交互；
   - 新增断言仍严格复用现有 `setting-model-probe-*` testid，不改页面业务组件。
4. 回归轻量链后，`check-only`、`verify-model-probe-help.mjs`、`pnpm --dir web run test:settings-no-browser` 与 `tsc --noEmit` 全部通过，说明本轮脚本增强没有破坏已恢复的 no-browser 统一入口。
5. 继续跑真实 settings browser smoke 时又暴露两个宿主级问题：
   - CDP `self-start` 仍像上一轮一样停在 `json-new -> attached-session` 双路径都无法完成 `Page.enable / Page.setLifecycleEventsEnabled / Runtime.enable` 的 bootstrap blocker；
   - CLI `self-start` 则先暴露出 settings PowerShell wrapper 误选 Codex bundled Node，导致找不到 `npx-cli.js`；修复 Node 选择优先级后，CLI 真实失败稳定收敛为 Playwright CLI 子进程 `spawn EPERM`。
6. 因此本轮又补了同主线的一个必要配套修复：
   - `scripts/verify-setting-help-browser-smoke.ps1` 现在会优先选有 `node_modules/npm/bin/npx-cli.js` 的可用 Node 路径，而不是误用 Codex bundled Node；
   - 同时把 CLI 路径里的 `spawn EPERM` 收口成明确的 host-level blocker，和 backup/shared wrapper 的口径保持一致，不再误报成“did not emit expected success marker”。
7. 最终结论是：`Model Probe` 的 browser smoke 契约已经补齐，但当前宿主上的真实 browser evidence 仍分别受 `CDP page bootstrap blocker` 与 `Playwright CLI spawn EPERM` 阻塞；这两个阻塞现在都能被明确分类，不会再被误记为页面回归或 success-marker 漂移。

## 验证

- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-setting-help-browser-smoke.mjs --check-only`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-setting-help-browser-smoke-cdp.mjs --check-only`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-model-probe-help.mjs`
- passed `. .\scripts\use-node-env.ps1; pnpm --dir web run test:settings-no-browser`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only`
- observed expected host blocker `page_bootstrap_timeout_attached_session / CdpPageBootstrapUnavailableError` from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- observed and then fixed Node selection drift for CLI mode; after the wrapper fix, observed expected host blocker `spawn EPERM` from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`
- passed PowerShell parser validation for `scripts\verify-setting-help-browser-smoke.ps1`
- passed `git diff --check -- scripts\verify-setting-help-browser-smoke.ps1 scripts\verify-setting-help-browser-smoke.mjs scripts\verify-setting-help-browser-smoke-cdp.mjs`

## 本轮变更文件

- `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-setting-help-browser-smoke.mjs`
- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-model-probe-browser-smoke-closure-and-settings-host-blocker-classification.md`

## 未完成 / 风险 / 阻塞

- 当前宿主仍无法给出 settings `Model Probe` 的真实 browser-grade green pass：CDP 路径稳定卡在 page bootstrap，CLI 路径稳定卡在 Playwright CLI `spawn EPERM`。
- 本轮补的是 smoke 契约与 blocker 分类，不是页面源码；因此未触发 `build:static` 与 `web/out -> static/out` 同步，这不影响本轮目标。
- 若下一轮继续追 settings browser evidence，必须优先换宿主或沿已有 host blocker 分类继续取证，不要回头再把 `Model Probe` 业务结构当成问题源。

## 下一轮候选任务顺序

1. 若有更健康宿主，直接复跑 `verify-setting-help-browser-smoke.ps1` 的 `self-start` 或 `external` 路径，验证本轮新增的 `Model Probe` 真实交互断言是否整体通过。
2. 若仍在本机，避免再改 `ModelProbe.tsx`；优先把 settings browser evidence 视为宿主问题跟踪项，转向同主线下其它不依赖当前 blocker 的 page-level browser gap。
3. 如果下一轮仍留在 settings 池，优先沿共享 wrapper 思路继续压实 host blocker 分类，而不是重新拆散成零散手工脚本验证。
