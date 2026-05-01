# 2026-04-30 Phase G Shared Root CDP Forward Args Alignment

## 1. 任务信息

- 任务名称：共享根 browser smoke wrapper 的 `Driver=cdp` 高级参数透传收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.4 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-browser-wrapper-inventory-guard.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-root-browser-alias-alignment.md`
- 本次任务目标：把共享根 `verify-channel-create-browser-smoke.ps1` 的 `Driver=cdp` 路径从“只转发基础参数”收口到与 page-level forwarder 一致的高级 CDP 参数面，避免 shared root 顶层命令继续报 `RequireExternalCdpPreflight` 等未知参数。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-copied-cdp-wrapper-preflight-closure.md`
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - `scripts/verify-browser-smoke-wrapper-alignment.mjs`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory：确认上一轮已经收口 shared root alias 与 specialized root guard，当前最小连续缺口是 shared root `Driver=cdp` 能力面没有跟上 page-level forwarder。
  - `using-superpowers`：按要求核对技能边界。
  - `brainstorming`：仅作流程边界核对；用户已明确要求在既有主线内直接推进代码，本轮不进入设计审批门禁。
- 若未使用部分本地资源或上下文，原因：本轮只处理 shared root wrapper 参数透传，不涉及页面组件、后端业务实现或真实 live rerun。
- 本次是否启用子 agent 与分工边界：否。
- 本次子 agent 使用模型：无。
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否。
- 若使用分目录 agent，负责目录与禁止越界范围：无。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、范围集中，主线程即可闭环。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修改共享根 wrapper 的 `Driver=cdp` 参数面、静态 guard 与状态记录，不改页面 smoke `mjs` 与前端业务组件。
- 必须保留已有 `-BrowserPath` 兼容性和 shared CDP wrapper 的实际执行逻辑。

## 4. 本次禁止事项

- 不扩散到真实 browser/CDP live rerun。
- 不顺手重构 page-level forwarder 或 specialized root 的执行流。
- 不改变共享 `verify-channel-create-browser-smoke-cdp.ps1` 的 runtime 行为和 host-blocker 分类逻辑。

## 5. 本次验收条件

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -BootstrapExternalCdpSession -CdpPageBootstrapStrategy auto -CdpBootstrapCommandOrder runtime-page-lifecycle -CdpPort 9333 -CdpUrl http://127.0.0.1:9333`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -BrowserPath $env:COMSPEC -RequireExternalCdpPreflight`
- `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- `git diff --check -- scripts\verify-channel-create-browser-smoke.ps1 scripts\verify-browser-smoke-wrapper-alignment.mjs docs\archive\status\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-shared-root-cdp-forward-args-alignment.md`

完成标准：共享根 `Driver=cdp` 直接接受并下沉 shared CDP wrapper 的高级参数；旧 `-BrowserPath` 兼容仍成立；alignment guard 与 screenshot 聚合入口继续通过。

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke.ps1`
- `scripts/verify-browser-smoke-wrapper-alignment.mjs`
- `docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-root-cdp-forward-args-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享根 wrapper 参数语义。
- 受影响后端模块：无。
- 受影响前端模块：无页面源码改动；仅调整 smoke wrapper guard 与文档记录。
- 受影响接口：无运行态 API 变化。
- 是否影响旧数据：否。
- 是否影响旧行为：只影响共享根 `Driver=cdp` 的公开参数面；旧 `-BrowserPath` 与既有执行逻辑保持兼容。

## 8. 实施步骤

1. 复核 automation memory、当前主线文档、运行态策略文档和相邻 Phase G wrapper worklog，确认本轮仍应停留在 browser smoke reliability 主线，不回到页面布局。
2. 用 `check-only` 直接验证共享根 `verify-channel-create-browser-smoke.ps1 -Driver cdp` 当前不接受 `-RequireExternalCdpPreflight`，把缺口从“推测”收敛为可复现参数错误。
3. 用 `apply_patch` 最小修改 `scripts/verify-channel-create-browser-smoke.ps1`：
   - 参数区新增 `CdpPort / CdpUrl / CdpCommandTimeoutMs / EdgeLaunchPreset / EdgeProfileStrategy / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder / BootstrapExternalCdpSession / RequireExternalCdpPreflight / SelfStartServices`；
   - `Driver=cdp` 的 `$forwardParams` 同步透传这些参数到底层 shared CDP wrapper；
   - 保留 `Browser -> BrowserPath` 的兼容映射。
4. 同步升级 `scripts/verify-browser-smoke-wrapper-alignment.mjs`，把 shared root `Driver=cdp` 的高级参数面也纳入静态守门，避免后续再次无声回退。
5. 运行 shared-root `check-only`、alignment guard、screenshot 聚合入口与 `git diff --check`，确认本轮是实质能力收口而不是只补文档。

## 9. 测试与验证

- 构建命令：未涉及构建。
- 测试命令：
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -BootstrapExternalCdpSession -CdpPageBootstrapStrategy auto -CdpBootstrapCommandOrder runtime-page-lifecycle -CdpPort 9333 -CdpUrl http://127.0.0.1:9333`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -BrowserPath $env:COMSPEC -RequireExternalCdpPreflight`
  - `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
  - `node .\scripts\run-frontend-verification-suite.mjs screenshot`
- 专项验证：
  - 本轮先观察到 pre-fix 失败：`A parameter cannot be found that matches parameter name 'RequireExternalCdpPreflight'.`
  - 修复后 `check-only` 摘要里能直接看到 `Explicit external CDP preflight requirement: enabled`、`External mode initial CDP preflight: required`，以及自定义 `CdpUrl / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder` 已真实下沉到 shared CDP wrapper。

## 10. 风险与兼容性

- 新风险：较低；修改集中在 shared root 参数透传层，不改底层 CDP 执行逻辑。
- 兼容性风险：低；旧 `-BrowserPath` 调用继续通过，新增参数只让 shared root 顶层能力与 page-level forwarder 对齐。
- 是否阻塞下一任务：不阻塞；下一轮可继续回到真实 browser/CDP 证据与宿主 blocker 分类。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。shared-root `Driver=cdp` 三组 `check-only`、alignment guard、screenshot 聚合入口与 `git diff --check` 均通过。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke reliability 主线，不扩散到页面布局返工。
  - `DETAILED_EXECUTION_WORKFLOW`、`ENV_READY_AND_NEXT_PLAN` 与 `USER_CONTEXT_REQUIREMENTS`：确认当前执行口径仍是默认停驻、按需 `check-only/self-start/external`，并要求把主规划、用户上下文和 worklog 同步对齐。
  - 相邻 wrapper worklog 与 automation memory：确认 thin forwarder、shared root alias 和 specialized root guard 已收口，当前最小连续缺口是 shared root `Driver=cdp` 高级参数面没有跟上。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行人工操作；本轮只跑 repo-local `check-only` 与 no-browser 聚合入口。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 browser/CDP 证据仍需健康宿主或外部会话；本轮不需要重开 live path。
- 待验证页面清单：下一轮优先在健康宿主上继续验证 `backup`、`ccswitch`、`group-create`、`home-layout`、`model-layout`、`channel-page`、`settings-help` 与 `ai-learning` 的真实 browser/CDP 证据。
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务小、范围集中，主线程即可闭环。
- worklog 是否更新：yes
- 遗留项：
  - 当前 browser smoke 家族的公开参数面已继续向 shared root `Driver=cdp` 收口；剩余主线工作重新回到真实 browser/CDP 证据与宿主 blocker。
  - 若后续 shared root 再增加新的 CDP 公共参数，也必须在同一轮同步更新 `verify-browser-smoke-wrapper-alignment.mjs`。
- 下一任务前置条件是否满足：满足。下一轮可直接沿同一 Phase G 主线继续真实 browser/CDP 证据或宿主 blocker 分类。

## 12. 执行与结果

1. 复核 automation memory、当前主线文档与相邻 wrapper worklog 后确认：thin forwarder、shared root alias 与 specialized root guard 都已收口，但共享根 `verify-channel-create-browser-smoke.ps1` 的 `Driver=cdp` 路径仍没有暴露 shared CDP wrapper 的高级参数面，是当前最小且连续的结构缺口。
2. 先用 pre-fix `check-only` 直接复现了这个缺口：统一根命令在 `-Driver cdp -RequireExternalCdpPreflight` 下会立即报 `NamedParameterNotFound`，说明 page-level forwarder 已支持的 external-preflight / bootstrap 参数并没有在 shared root 顶层同样成立。
3. 因此本轮只对 shared root 做最小补丁：
   - `verify-channel-create-browser-smoke.ps1` 的 `Driver=cdp` 现在会显式接受并 forward `CdpPort / CdpUrl / CdpCommandTimeoutMs / EdgeLaunchPreset / EdgeProfileStrategy / CdpPageBootstrapStrategy / CdpBootstrapCommandOrder / BootstrapExternalCdpSession / RequireExternalCdpPreflight / SelfStartServices`；
   - `verify-browser-smoke-wrapper-alignment.mjs` 也同步升级为要求 shared root 保留这组 `cdp` 参数面和对应透传，不再只检查 alias 与基础 `Browser -> BrowserPath` 映射。
4. 验证结果显示：
   - shared-root `-Driver cdp -RequireExternalCdpPreflight` 已能正常输出 shared CDP wrapper 的 `check-only` 摘要；
   - 带 `-BootstrapExternalCdpSession -CdpPageBootstrapStrategy auto -CdpBootstrapCommandOrder runtime-page-lifecycle -CdpUrl http://127.0.0.1:9333` 的调用也会正确反映到摘要里，说明参数不是“被接受但没下沉”；
   - 旧 `-BrowserPath` 兼容仍然成立；
   - alignment guard 与 screenshot 聚合入口继续通过，说明本轮没有打断既有 repo-local 验证编排。

## 13. 验证

- passed `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- observed expected pre-fix parameter gap from `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -BootstrapExternalCdpSession -CdpPageBootstrapStrategy auto -CdpBootstrapCommandOrder runtime-page-lifecycle -CdpPort 9333 -CdpUrl http://127.0.0.1:9333`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only -Driver cdp -BrowserPath $env:COMSPEC -RequireExternalCdpPreflight`
- passed `node .\scripts\verify-browser-smoke-wrapper-alignment.mjs`
- passed `node .\scripts\run-frontend-verification-suite.mjs screenshot`
