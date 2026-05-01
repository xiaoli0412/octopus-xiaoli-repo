# 2026-04-22 Phase G 设置页帮助提示 smoke 的 headed 预设对照与 bootstrap 分类收口

## 1. 任务信息

- 任务名称：设置页帮助提示真实浏览器 smoke 的 headed 预设对照与 attached-session bootstrap 分类收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：设置页四卡片帮助提示浏览器 smoke 阻塞继续收敛

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、10.1 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-setting-help-self-start-node-timebox-and-attach-fallback.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-bootstrap-narrowing.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-early-command-probe.md`
- 本次任务目标：
  - 为 PowerShell wrapper 增加 `headed-relaxed` 自启动预设，验证当前阻塞是否仅发生在 headless 自启动模式
  - 让 CDP smoke 在 fallback attached-session bootstrap 仍然全超时时提前结构化失败，而不是继续挂到外层 Node 超时
  - 把最新对照结论同步到前端主线状态与 automation memory，给下一轮留下明确入口
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-self-start-node-timebox-and-attach-fallback.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-bootstrap-narrowing.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-early-command-probe.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、最近设置页 smoke narrowing worklog、automation memory、当前脚本源码与 trace 产物
- 若未使用部分本地资源或上下文，原因：本轮继续限制在设置页帮助提示 smoke 同主线，不扩散到备份、渠道、分组、`CC Switch` 或后端业务逻辑
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，且本轮是单脚本链路的高耦合验证闭环

## 3. 本次硬规则

- 只处理设置页帮助提示真实 smoke 的验证脚本链，不修改设置页四卡片业务逻辑
- 不把未通过的真实浏览器 smoke 写成通过
- 必须留下可复跑、可对照、可分类的失败证据，而不是继续停留在“超时但不确定为什么”

## 4. 本次禁止事项

- 不扩散到设置页其他功能、备份导入、渠道、多 key、分组或 `CC Switch` 主线
- 不回退工作区中与本轮无关的现有修改
- 不把宿主环境阻塞伪装成脚本已经完全解决

## 5. 本次验收条件

- `scripts/verify-setting-help-browser-smoke.ps1` 支持 `headed-relaxed` 预设并把预设信息传给 Node smoke
- `scripts/verify-setting-help-browser-smoke-cdp.mjs` 在 fallback attached-session bootstrap 全超时时提前抛出结构化错误
- `node --check` 与 `check-only` 验证通过
- 至少一轮真实 `self-start + cdp` 验证能把错误从外层 Node 超时收敛为明确的 bootstrap 分类结果

## 6. 本次回滚点

- `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-22-phase-g-setting-help-headed-bootstrap-classification.md`
- automation memory `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本与运行态分类，不涉及 UI 业务行为
- 受影响后端模块：无业务改动；仅复用现有 `build/octopus-smoke.exe` 作为自启动 backend
- 受影响前端模块：无业务改动
- 受影响接口：无产品接口改动；仅复用登录与设置页页面加载路径
- 是否影响旧数据：否
- 是否影响旧行为：仅影响 smoke 验证链的启动预设、错误分类与日志可读性

## 8. 实施步骤

1. 复核主规划、用户上下文、详细工作流、前端主线状态和最近同主线 worklog，确认本轮继续留在 Phase G 设置页帮助提示 smoke 收口主线。
2. 给 `scripts/verify-setting-help-browser-smoke.ps1` 增加 `headed-relaxed` 预设、窗口样式控制和预设信息透传。
3. 给 `scripts/verify-setting-help-browser-smoke-cdp.mjs` 增加更早的 bootstrap 失败分类，在 fallback attached-session 也全超时时直接抛出 `CdpPageBootstrapUnavailableError`。
4. 跑 `node --check`、`check-only`、`self-start + cdp`，并对比 `relaxed` 与 `headed-relaxed` 的真实结果。
5. 把最新结论同步到前端主线状态与 automation memory。

## 9. 测试与验证

- 构建命令：`D:\gol1\node.exe --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 测试命令：
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'check-only' -Driver 'cdp' -EdgeLaunchPreset 'relaxed' -NodeSmokeTimeoutSeconds 30 ; exit $LASTEXITCODE"`
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'check-only' -Driver 'cdp' -EdgeLaunchPreset 'headed-relaxed' -NodeSmokeTimeoutSeconds 30 ; exit $LASTEXITCODE"`
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'self-start' -Driver 'cdp' -EdgeLaunchPreset 'relaxed' -NodeSmokeTimeoutSeconds 75 -KeepArtifacts ; exit $LASTEXITCODE"`
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'self-start' -Driver 'cdp' -EdgeLaunchPreset 'headed-relaxed' -NodeSmokeTimeoutSeconds 90 -KeepArtifacts ; exit $LASTEXITCODE"`
  - `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'self-start' -Driver 'cdp' -EdgeLaunchPreset 'relaxed' -NodeSmokeTimeoutSeconds 95 -KeepArtifacts ; exit $LASTEXITCODE"`
- 专项验证：
  - `check-only` 输出现在会包含 `edgeLaunchPreset / edgeProfileStrategy`
  - `relaxed` 自启动仍失败，但失败已从“外层 Node 超时”收敛为 `CdpPageBootstrapUnavailableError`
  - `headed-relaxed` 自启动与 `relaxed` 一样在 attached-session bootstrap 阶段卡死，证明当前阻塞不只限于 headless 自启动

## 10. 风险与兼容性

- 新风险：低；本轮只改验证脚本与错误分类逻辑
- 兼容性风险：低；未改接口、数据库、前端业务逻辑
- 是否阻塞下一任务：部分阻塞；真实浏览器 smoke 仍未通过，但阻塞已被稳定分类为 Edge/CDP page bootstrap 宿主问题

## 11. 收工记录

- 构建是否通过：脚本级静态检查通过
- 测试是否通过：
  - 通过：`node --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - 通过：两条 `check-only` 验证
  - 失败但已收敛：三条真实 `self-start + cdp` 验证未通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、最近设置页 smoke worklog、automation memory、当前脚本源码与 trace 产物
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划与用户上下文要求本轮继续停留在 Phase G 设置页帮助提示真实 smoke 同主线，不得切换到其他主题
  - 详细工作流要求每轮补齐小计划、验证、worklog 与 memory
  - 上一轮 worklog 与 memory 明确指出下一步应比较更轻命令与外部/非 headless 会话差异，本轮据此优先加入 headed 预设并提前分类 bootstrap 失败
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：已实际执行多轮真实 `self-start + cdp`；均未形成通过证据
- 手工 smoke 阻塞原因 / 缺少的环境：
  - `relaxed` 与 `headed-relaxed` 都会在 `json-new` 页面 bootstrap 失败后落到 `attached-session` 回退
  - 回退后的 `Page.enable / Page.setLifecycleEventsEnabled / Runtime.enable` 仍全部超时
  - 说明当前主机的 Edge/CDP 对 page bootstrap 域整体无响应，不只是 headless 自启动策略问题
- 待验证页面清单：设置页 `LLMPrice / DynamicRouting / CircuitBreaker / ModelProbe` 的桌面布局、`375px` 布局、帮助按钮 focus/tooltip
- 若未使用子 agent，原因：用户明确要求主线程串行推进
- worklog 是否更新：是
- 遗留项：
  - 仍缺“外部已打开 CDP 会话”对照验证，来确认是否仅限自启动临时 profile
  - wrapper 中 `Target.closeTarget` 在 fallback 错误后的 `No target with given id found` 仍只是被 trace 记录，暂不影响主结论
  - 最终浏览器通过证据仍未闭环
- 下一任务前置条件是否满足：满足；下一轮可直接以本轮新增的结构化失败分类为基线，切到“外部已有 CDP 会话”同主线对照验证

## 12. 最终状态

- 本次结果：成功
- 是否需要人工介入：否
