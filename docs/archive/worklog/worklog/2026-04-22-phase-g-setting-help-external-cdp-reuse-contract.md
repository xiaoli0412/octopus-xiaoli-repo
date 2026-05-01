# 2026-04-22 Phase G 设置页帮助提示 external CDP 复用契约收口

## 1. 任务信息

- 任务名称：设置页帮助提示真实浏览器 smoke 的 external CDP 复用契约收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：设置页四卡片帮助提示浏览器 smoke 阻塞继续收敛

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、10.1 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-setting-help-headed-bootstrap-classification.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-cdp-fallback.md`
- 本次任务目标：
  - 把 `scripts/verify-setting-help-browser-smoke.ps1` 的 `external + cdp` 模式收紧成真正的“只复用外部后端/前端/CDP 会话”
  - 避免外部模式在缺少 CDP 端点时偷偷代启 Edge，影响“external 会话 vs self-start 临时 profile”的对照结论
  - 把该契约同步回前端主线状态与 automation memory，给下一轮留下明确入口
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-headed-bootstrap-classification.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-cdp-fallback.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，且本轮是高耦合单脚本契约收口

## 3. 本次硬规则

- 只处理设置页帮助提示真实 smoke 的 wrapper 契约，不修改设置页四卡片业务逻辑
- 不把 external 模式的“缺端点自动代启”继续保留成隐式行为
- 必须留下可复跑、可解释、可对照的 external 模式入口

## 4. 本次禁止事项

- 不扩散到设置页其他功能、备份导入、渠道、多 key、分组或 `CC Switch` 主线
- 不回退工作区中与本轮无关的现有修改
- 不把受控失败写成真实浏览器 smoke 已通过

## 5. 本次验收条件

- `scripts/verify-setting-help-browser-smoke.ps1` 在 `-Mode external -Driver cdp` 下不再因为缺少 CDP 端点而代启 Edge
- `check-only` 输出能明确提示 external CDP 模式依赖外部已开的 remote debugging 会话
- 前端主线状态同步写明 external 模式契约已收紧

## 6. 本次回滚点

- `scripts/verify-setting-help-browser-smoke.ps1`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-reuse-contract.md`
- automation memory `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 实施步骤

1. 复核主规划、用户上下文、详细工作流、前端主线状态和最近同主线 worklog，确认本轮继续停留在 Phase G 设置页帮助提示 smoke 收口主线。
2. 对照现有脚本实现，确认 `external + cdp` 模式仍会在缺少 CDP 端点时代启 Edge，与 worklog 和 memory 中“复用外部会话”的下一步目标不一致。
3. 修改 `scripts/verify-setting-help-browser-smoke.ps1`：
   - `check-only` 输出增加 external CDP 依赖说明
   - `external + cdp` 模式在缺少 CDP 端点时直接报错并给出下一步提示
   - 只有 `self-start + cdp` 才允许代启 Edge
4. 同步更新 `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`，写明 external 模式契约已收紧。
5. 跑最小验证并记录本轮结果与下一轮入口。

## 8. 测试与验证

- 通过：`powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'check-only' -Driver 'cdp' -CdpUrl 'http://127.0.0.1:9999' -NodeSmokeTimeoutSeconds 15 ; exit $LASTEXITCODE"`
  - 结果：通过，`check-only` 继续正常输出，并保留了 `edge-cdp expects an external browser with remote debugging enabled` 的 Node 侧说明
- 未完成：`external + cdp` 的“前后端存在、CDP 缺失”受控失败验证
  - 原因：本轮多次尝试通过单条命令临时拉起本地后端或 mock 服务后再调用 wrapper，但当前环境对较长的嵌套命令触发了 policy block，导致没能在本轮拿到更精确的第二条受控失败证据

## 9. 风险与兼容性

- 新风险：低；本轮只改验证脚本和文档同步
- 兼容性风险：低；未改接口、数据库、前端业务逻辑
- 当前阻塞：真实浏览器 smoke 仍未闭环；但 external 模式与 self-start 模式的职责边界已经收紧，后续对照结果会更可信

## 10. 收工记录

- 构建是否通过：未涉及构建
- 测试是否通过：部分通过；`check-only` 通过，外部模式的更精确受控失败验证因命令策略阻塞未完成
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、最近设置页 smoke worklog、automation memory、当前脚本源码
- 本次使用了哪些子 agent 及其结论：无
- worklog 是否更新：是
- 遗留项：
  - 仍需拿一条真实外部 Edge CDP 会话去跑 `external + cdp`，确认当前宿主问题是否只限 self-start 临时 profile
  - `desktop` 与 `375px` 的真实浏览器通过证据仍未闭环
- 下一任务前置条件是否满足：满足；下一轮可直接复用本轮收紧后的 external 契约，跑真实外部会话对照

## 11. 最终状态

- 本次结果：成功
- 是否需要人工介入：否
