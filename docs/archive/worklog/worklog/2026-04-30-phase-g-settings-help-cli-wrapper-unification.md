# 2026-04-30 Phase G Settings Help CLI Wrapper Unification

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / browser smoke reliability`
- Current stage: `shared wrapper entry alignment`

## 本轮上下文与本地资源

- 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-cli-wrapper-unification.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-cdp-wrapper-false-positive-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-probe-browser-smoke-closure-and-settings-host-blocker-classification.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 Phase G 主线下的验证入口收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求主线程串行推进，不创建子 agent。

## 本轮候选任务

1. 继续找出仍保留复制版 CLI wrapper 的 page-level smoke 入口，并收口到共享 wrapper。
2. 若 `settings-help` CLI 入口可安全收口，则统一到共享 `verify-channel-create-browser-smoke.ps1`。
3. 保持 settings 现有 `Model Probe` browser smoke 契约不回退，只修 wrapper，不重开页面组件实现。

## 本轮计划

- 本轮核心任务：把 `settings-help` 的 CLI PowerShell wrapper 从自复制实现收口到共享 CLI wrapper。
- 本轮配套任务：验证 `check-only` 与真实 `self-start -Driver cli` 是否仍稳定落到预期宿主 blocker。
- 预期验证方式：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cli`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`
  - PowerShell parser validation for `scripts\verify-setting-help-browser-smoke.ps1`
  - `git diff --check -- scripts\verify-setting-help-browser-smoke.ps1 docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-settings-help-cli-wrapper-unification.md`
- 完成判定标准：
  - `settings-help` 不再保留复制版 CLI wrapper 逻辑；
  - CLI `check-only` 通过；
  - CLI `self-start` 若仍失败，必须稳定收口到共享 wrapper 的宿主 blocker 语义。

## 本轮硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 不改 `ModelProbe.tsx` 或 settings 业务组件，只改 smoke wrapper 入口。
- 不扩大到无关页面或后端代码。

## 执行与结果

1. 先复核 automation memory、当前主线文档、前端状态和最近相邻 worklog，确认上一轮已经把 `settings-help` 的 `Model Probe` browser smoke 契约补齐，当前更值得推进的是 wrapper 结构一致性，而不是重复改页面实现。
2. 盘点同池脚本后确认：`backup`、`ccswitch`、`group-create` 的 CLI 路径都已 forward 到共享 `verify-channel-create-browser-smoke.ps1`，而 `settings-help` 仍保留一整份复制版 CLI 包装逻辑，是当前最小且连续的收口点。
3. 继续核对共享 CLI wrapper 的调用契约，确认它本来就是“PowerShell wrapper 自启服务，Node smoke 走 `--external` 模式”；`settings-help` 现有 CLI 路径也是同样语义，因此改成 forwarder 不会改变页面验证行为，只会消除复制逻辑。
4. 直接将 `scripts/verify-setting-help-browser-smoke.ps1` 重写成纯 forwarder：
   - `Driver=cdp` 继续 forward 到共享 `verify-channel-create-browser-smoke-cdp.ps1`，保留 `settings-help` 原有的 `CdpPageBootstrapStrategy=auto` 等默认参数；
   - `Driver=cli` 改为和 `backup / ccswitch / group-create` 一样 forward 到共享 `verify-channel-create-browser-smoke.ps1`；
   - 通过环境变量注入 `scripts/verify-setting-help-browser-smoke.mjs`、`setting-help-browser-smoke passed` 与 `settings help` label。
5. 验证结果表明收口生效：
   - `check-only` 默认 CDP 路径继续通过；
   - `check-only -Driver cli` 继续通过，并显示共享 wrapper 的统一 loopback 预检摘要；
   - `self-start -Driver cli` 在本机稳定失败于共享 wrapper 的 Playwright CLI `spawn EPERM` host blocker，不再依赖 settings 自己那份复制版分类逻辑。
6. 因此本轮结论是：`settings-help` 当前剩余的 CLI browser gap 已和 `backup / ccswitch / group-create` 一样，明确收敛为宿主执行环境问题；后续无需再维护一份单独的 settings CLI wrapper 实现。

## 验证

- passed PowerShell parser validation for `scripts\verify-setting-help-browser-smoke.ps1`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cli`
- observed expected host blocker `spawn EPERM` from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180`

## 本轮变更文件

- `scripts/verify-setting-help-browser-smoke.ps1`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-settings-help-cli-wrapper-unification.md`

## 未完成 / 风险 / 阻塞

- 当前宿主上的 `settings-help` CLI browser smoke 仍会稳定卡在 Playwright CLI 子进程 `spawn EPERM`，这属于宿主 blocker，不是页面回归。
- 本轮没有重跑 `Driver=cdp self-start`，因为该路径的宿主 blocker 已在上一轮明确为 `page_bootstrap_timeout_attached_session`，本轮重点仅是 CLI wrapper 结构收口。

## 下一轮候选任务顺序

1. 继续 Phase G，盘点是否还有其它 page-level smoke 入口保留复制版 CLI wrapper，优先收口到共享链。
2. 若 wrapper 收口已基本完成，则切回健康宿主/外部会话复跑一条真实 page-level browser smoke，验证当前剩余问题确实都只是宿主 blocker。
3. 若仍留在本机，不再反复重开 settings 页面实现；把焦点保持在 shared wrapper 结构一致性和 host blocker 分类上。
