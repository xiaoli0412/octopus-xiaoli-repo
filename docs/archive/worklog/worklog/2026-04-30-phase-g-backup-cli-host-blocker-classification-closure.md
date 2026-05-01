# 2026-04-30 Phase G Backup CLI Host Blocker Classification Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / backup page browser evidence`
- Current stage: `backup 页面 browser smoke 宿主阻塞分类收口`

## 本轮上下文与本地资源

- canonical / 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 需求清单：`docs/PLAN.md`
- 主规划：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 当前状态：`docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`
  - `docs/archive/worklog/worklog/2026-04-27-phase-g-backup-browser-wrapper-host-compat-closure.md`
  - `docs/archive/worklog/worklog/2026-04-27-phase-g-backup-browser-smoke-script-closure-and-host-blocker-capture.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 canonical plan 下的增量重构与验证收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求“不要创建子agent，靠主线程解决”。

## 本轮候选任务

1. 运行 backup 页面 browser smoke，确认当前断点是页面契约还是共享 wrapper。
2. 若断点在共享 wrapper，做最小修复，让失败语义能准确区分宿主 blocker 与页面回归。
3. 回归 backup 页 no-browser + check-only + self-start，留下下一轮明确入口。

## 本轮计划

- 本轮核心任务：把 backup 页面 CLI browser smoke 的泛化失败改成明确的 host-level `spawn EPERM` 分类。
- 本轮配套任务：回归 `backup-component` 与 `check-only`，确认页面契约未回退。
- 预期验证方式：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
  - `node .\scripts\verify-backup-component.cjs`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
  - `git diff --check -- scripts\verify-channel-create-browser-smoke.ps1`
- 完成判定标准：backup `self-start` smoke 若不能通过，至少必须输出明确的 `spawn EPERM` 宿主 blocker，而不是误导性的 success-marker 失败或 channel-create 标签噪音。

## 本轮硬规则

- 只在 `Phase G` 主线内推进，不扩散到无关 UI 改造。
- 写入范围控制在共享 CLI smoke wrapper 与必要记录。
- 不回退用户已有改动，不清理无关脏区。

## 本轮禁止事项

- 不改 backup 页面业务逻辑。
- 不重写 Playwright CLI smoke 基建。
- 不把宿主 `spawn EPERM` 误记成 backup 页面回归。

## 本轮回滚点

- 仅回滚 `scripts/verify-channel-create-browser-smoke.ps1` 本轮补丁即可撤销行为变化。

## 执行与结果

1. 先跑 backup focused 验证，确认 `verify-backup-browser-smoke.ps1 -Mode check-only` 与 `node .\scripts\verify-backup-component.cjs` 均通过，说明 backup 页面 selector / no-browser 契约仍正常。
2. 真实重跑 `verify-backup-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180` 后，确认失败并不在 backup 页面 DOM 或 import/rollback 逻辑，而是在 `scripts/verify-backup-browser-smoke.mjs` 内调用 Playwright CLI 时直接触发 Node 子进程 `spawn EPERM`。
3. 对共享 `scripts/verify-channel-create-browser-smoke.ps1` 做最小收口：
   - `Assert-NodeSmokeSucceeded` 新增 `SmokeLabel` 参数，不再把所有 CLI smoke 失败都写成 `channel create`。
   - 当 stdout/stderr 同时满足 `playwright-cli start` 和 `spawn EPERM` 特征时，直接抛出明确的 host-level blocker 结论。
   - 临时目录名改为跟随当前页面标签，避免 backup 等页面继续落到 `octopus-channel-create-smoke-*` 噪音目录。
4. 修复后重跑 backup `self-start` smoke，确认失败信息已收口为“Node backup is blocked by a host-level child-process 'spawn EPERM' failure while launching Playwright CLI”，可直接作为宿主 blocker 交接，而不是页面回归信号。

## 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- passed `node .\scripts\verify-backup-component.cjs`
- observed expected host blocker `spawn EPERM` from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180`
- passed `git diff --check -- scripts\verify-channel-create-browser-smoke.ps1`

## 本轮变更文件

- `scripts/verify-channel-create-browser-smoke.ps1`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-backup-cli-host-blocker-classification-closure.md`

## 未完成 / 风险 / 阻塞

- backup 页真实 browser-grade pass 仍未关闭；当前宿主在 Playwright CLI 拉起阶段稳定触发 `spawn EPERM`。
- 本轮没有改 backup 页面源码，因此未触发 `build:static` / `web/out -> static/out` 同步动作；后续若再改页面源码，仍需按 `UI_MAINLINE_TASK_2026-04-30` 的要求同步运行态产物。
- 共享 wrapper 的新分类只解决“失败语义误导”问题，不解决宿主权限/缓存/安全软件导致的 `spawn EPERM` 根因。

## 下一轮候选任务顺序

1. 延续 `Phase G`，在可用宿主上直接复跑 `verify-backup-browser-smoke.ps1 -Mode external` 或 `self-start`，用当前明确 blocker 分类继续取证。
2. 若仍留在本机，避免再把 backup 页面当成业务回归排查；优先切换到同主线下不依赖 Playwright CLI 的相邻页面闭环，或改走已验证可用的 CDP 路径页面。
3. 若继续补 backup browser evidence，应优先研究是否能像 channel/home/model 一样切到现有 CDP wrapper，而不是重复围绕 CLI `spawn EPERM` 空转。
