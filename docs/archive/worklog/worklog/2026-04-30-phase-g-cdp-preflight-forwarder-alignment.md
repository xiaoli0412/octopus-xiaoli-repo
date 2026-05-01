# 2026-04-30 Phase G CDP Preflight Forwarder Alignment

## 1. 任务信息

- 任务名称：共享 CDP external preflight 参数透传收口
- 日期：2026-04-30
- 当前阶段：`Phase G screenshot-first UI closure / browser smoke reliability`
- 对应 milestone：里程碑 4 `UI 与易用性`

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9.1 / 9.3 / 9.5 / 9.6 / 12 Phase 7`
- 对应 workflow 章节：`Phase G` UI、移动端、部署与最终验收
- 上一个相关 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-shared-cdp-wrapper-false-positive-closure.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-verification-entrypoint-and-no-browser-contract-recovery.md`
- 本次任务目标：把共享 `verify-channel-create-browser-smoke-cdp.ps1` 已有的 `-RequireExternalCdpPreflight` 能力补齐到仍缺透传的 page-level forwarder，收口同池 external 入口不一致。
- 本次已盘点本地资源：
  - `docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
  - `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/archive/worklog/worklog/README.zh-CN.md`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - `scripts/verify-home-layout-browser-smoke.ps1`
  - `scripts/verify-model-layout-browser-smoke.ps1`
  - `scripts/verify-channel-page-browser-smoke.ps1`
  - `scripts/verify-ccswitch-browser-smoke.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `using-superpowers` 约束核对
  - `brainstorming` 仅作非门禁式约束核对
- 若未使用部分本地资源或上下文，原因：本轮问题已缩小到 smoke wrapper 参数对齐，不需要重新展开业务组件或更早阶段需求稿。
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。

## 3. 本次硬规则

- 只在 `Phase G` browser smoke reliability 主线内推进。
- 只修 wrapper 参数透传和记录口径，不扩散到页面组件或后端逻辑。
- 不为复制版大脚本补整套新 helper；若改动面开始扩散，就在本轮明确截断范围。

## 4. 本次禁止事项

- 不回改已通过验证的页面布局或组件行为。
- 不把 `CLI` wrapper 和 `CDP` wrapper 的参数边界混为一谈。
- 不为了追求“全都统一”而把本轮扩大成多份复制 wrapper 的重构。

## 5. 本次验收条件

- `verify-home-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `verify-model-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `verify-channel-page-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`

以上四个入口都要把 `Explicit external CDP preflight requirement: enabled` 与 `External mode initial CDP preflight: required` 传到共享 wrapper；并同步更新状态文档与 worklog。

## 6. 本次回滚点

- `scripts/verify-home-layout-browser-smoke.ps1`
- `scripts/verify-model-layout-browser-smoke.ps1`
- `scripts/verify-channel-page-browser-smoke.ps1`
- `scripts/verify-ccswitch-browser-smoke.ps1`
- `docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-cdp-preflight-forwarder-alignment.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本入口语义
- 受影响后端模块：无
- 受影响前端模块：无页面源码改动
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只影响 page-level CDP wrapper 的 external/check-only 参数透传与可见说明

## 8. 实施步骤

1. 盘点当前所有 `*browser-smoke*.ps1`，确认哪些页面级 CDP forwarder 已经支持 `-RequireExternalCdpPreflight`，哪些仍缺透传。
2. 只对仍复用共享 `verify-channel-create-browser-smoke-cdp.ps1` 的 forwarder 做最小补丁。
3. 用 `check-only` 证明新参数确实传到了共享 wrapper，而不是只停留在脚本签名层。
4. 把复制版 wrapper 的剩余缺口记录到状态文档，留待下一轮独立处理。

## 9. 测试与验证

- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- `git diff --check -- scripts\verify-home-layout-browser-smoke.ps1 scripts\verify-model-layout-browser-smoke.ps1 scripts\verify-channel-page-browser-smoke.ps1 scripts\verify-ccswitch-browser-smoke.ps1 docs\archive\status\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs\archive\worklog\worklog\2026-04-30-phase-g-cdp-preflight-forwarder-alignment.md`

## 10. 风险与兼容性

- 新风险较低：本轮只新增 switch 透传，不改变 smoke 断言本身。
- 兼容性风险：若后续有人误以为所有 CDP wrapper 都已接入该开关，可能忽略 `group-create-cdp` 与 `settings-help` 两份复制版实现仍未同步。
- 是否阻塞下一任务：不阻塞。下一轮可以直接沿“复制版 wrapper 与 shared helper 继续收口”推进。

## 11. 收工记录

- 构建是否通过：未涉及构建。
- 测试是否通过：通过。四个 page-level forwarder 的 `check-only` 已确认把 external CDP preflight 开关传到共享 wrapper。
- 本次使用了哪些本地资源 / skills / 记忆上下文：见上文。
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `UI_MAINLINE_TASK_2026-04-30`：确认本轮仍应停留在 screenshot-first / browser smoke 主线，不回到页面布局返工。
  - 前端主线状态与上一轮 worklog：确认更高优先级问题已从“wrapper 假阳性”转成“入口参数口径分裂”。
  - 自动化 memory：确认上一轮已修共享 false positive，本轮最合理的相邻任务是把同池 forwarder 继续收口，而不是再追宿主 CDP blocker。
- 本次使用了哪些子 agent 及其结论：无。
- 子 agent 分工、负责范围与产出摘要：无。
- 手工 smoke 状态：未执行；本轮只做 wrapper 入口与 `check-only` 验证。
- 手工 smoke 阻塞原因 / 缺少的环境：真实 external/self-start browser smoke 仍受本机宿主条件影响，本轮无需为参数透传问题重跑整条页面级流程。
- 待验证页面清单：`group-create-cdp`、`settings-help` 复制版 wrapper 的 external CDP preflight 接入仍待处理。
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent。
- worklog 是否更新：yes
- 遗留项：
  - `group-create-cdp` 与 `settings-help` 不走本轮这套共享 forwarder，因此仍未暴露 `-RequireExternalCdpPreflight`。
  - `CLI` wrapper 不需要这条参数，不应误判为缺失。
- 下一任务前置条件是否满足：满足。下一轮可直接从“复制版 CDP wrapper 继续向 shared external-preflight helper 靠拢”接续。

## 12. 执行与结果

1. 先盘点所有 `*browser-smoke*.ps1` 后确认：`home-layout`、`model-layout`、`channel-page`、`ccswitch` 这四个 page-level CDP wrapper 仍只是简单转发到共享 `verify-channel-create-browser-smoke-cdp.ps1`，但没有把共享 wrapper 已支持的 `-RequireExternalCdpPreflight` 透传出来。
2. 这会造成同一 Phase G 池里出现入口口径分裂：learning/shared wrapper 已能显式声明 external 首轮是否必须先做 CDP reachability 预检，而这四个页面只能继续吃隐式默认值。
3. 因此本轮只做最小补丁：给四个 forwarder 增加 `-RequireExternalCdpPreflight` 开关，并在 `$forwardParams` 中透传到共享 wrapper，不改页面逻辑、不改 smoke 断言、不改 backend/self-start 行为。
4. 验证后确认四个 `check-only` 都会打印：
  - `Explicit external CDP preflight requirement: enabled`
  - `External mode initial CDP preflight: required`
   这说明参数已经真正进入共享 wrapper，而不是只停留在入口签名。
5. 同时继续做范围控制：盘点结果显示 `group-create-cdp` 与 `settings-help` 并不是本轮这类简单 forwarder，而是各自复制了一套更早的 CDP wrapper 逻辑。本轮没有把任务扩大成“重构复制版 wrapper”，只把它们明确记录为下一轮同主线候选。

## 13. 验证

- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only -RequireExternalCdpPreflight`

