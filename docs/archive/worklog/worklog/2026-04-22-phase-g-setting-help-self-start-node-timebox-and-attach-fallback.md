# 2026-04-22 Phase G 设置页帮助提示 self-start 日志收口与 attached-session 回退

## 1. 任务信息

- 任务名称：设置页帮助提示真实浏览器 smoke 的 Node 日志收口与 CDP attached-session 回退
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：设置页四卡片帮助提示浏览器 smoke 阻塞继续收敛

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、10.1 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-powershell-entry.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-smoke-wrapper-port-guard.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-page-session-narrowing.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-bootstrap-narrowing.md`
- 本次任务目标：
  - 让 `scripts/verify-setting-help-browser-smoke.ps1` 在真实 `self-start + cdp` 路径上为 Node smoke 子进程留下独立 stdout/stderr 日志，并且在超时后给出可读错误，而不是继续黑箱等待
  - 清理 `scripts/verify-setting-help-browser-smoke.mjs` 中残留的乱码帮助按钮常量，统一到当前真实 `HelpHint` 默认标签
  - 为 `scripts/verify-setting-help-browser-smoke-cdp.mjs` 增加 `json/new` 页面 websocket 全超时后的 `Target.attachToTarget` 回退，验证阻塞是否只在一种 page 连接路径上出现
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-powershell-entry.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-smoke-wrapper-port-guard.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-page-session-narrowing.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-cdp-bootstrap-narrowing.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、最近设置页 smoke / CDP narrowing worklog、automation memory、当前脚本源码与 trace 产物
- 若未使用部分本地资源或上下文，原因：本轮只收口设置页帮助提示 smoke 同主线，不扩散到备份、渠道、分组、`CC Switch` 或后端业务语义
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，且本轮是单脚本链路的高耦合诊断闭环

## 3. 本次硬规则

- 只处理设置页帮助提示真实 smoke 脚本链，不修改设置页四卡片业务语义
- 不把未跑通的浏览器 smoke 写成已通过
- 必须留下下一轮可直接复跑的日志证据和更窄的阻塞描述

## 4. 本次禁止事项

- 不扩散到设置页其他业务逻辑、备份导入、渠道、多 key、分组或 `CC Switch` 主线
- 不回退工作区中与本轮无关的现有修改
- 不把外部宿主环境问题伪装成代码已全部完成

## 5. 本次验收条件

- `scripts/verify-setting-help-browser-smoke.ps1` 能打印并保留 Node smoke stdout/stderr 日志路径，且支持超时退出
- `scripts/verify-setting-help-browser-smoke.mjs` 不再保留残留乱码帮助按钮常量
- `scripts/verify-setting-help-browser-smoke-cdp.mjs` 在 `json/new` 路径 bootstrap 全超时后，会自动尝试 `attached-session` 回退
- `node --check`、`--check-only` 与 PowerShell `check-only` 通过
- 至少一次真实 `self-start + cdp` 运行能把阻塞从黑箱超时收敛成更窄的 trace 事实

## 6. 本次回滚点

- `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-setting-help-browser-smoke.mjs`
- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-22-phase-g-setting-help-self-start-node-timebox-and-attach-fallback.md`
- automation memory `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 实现范围

- 受影响后端模块：无业务改动；仅复用现有 `build/octopus-smoke.exe` 和嵌入静态壳
- 受影响前端模块：无业务改动；仅清理 smoke 脚本中的帮助按钮默认标签常量
- 受影响脚本：
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 是否影响旧数据：否
- 是否影响旧行为：仅影响验证链可观测性与回退路径，不影响产品运行时行为

## 8. 实施步骤

1. 复核上一轮 wrapper / CDP narrowing 结论，确认这轮继续留在 Phase G 设置页帮助提示真实浏览器 smoke 同主线。
2. 给 `scripts/verify-setting-help-browser-smoke.ps1` 增加 Node 子进程超时等待、stdout/stderr 日志文件和 trace tail 汇总。
3. 清理 `scripts/verify-setting-help-browser-smoke.mjs` 里残留的乱码帮助按钮常量，统一到 `查看帮助`。
4. 给 `scripts/verify-setting-help-browser-smoke-cdp.mjs` 增加 `json/new` 页面 websocket bootstrap 全超时时的 `attached-session` 回退。
5. 运行静态验证与真实 `self-start + cdp`，收敛新的阻塞点。
6. 把结果同步到 worklog、前端主线状态和 automation memory。

## 9. 测试与验证

- 已执行：`D:\gol1\node.exe --check scripts/verify-setting-help-browser-smoke.mjs`
- 已执行：`D:\gol1\node.exe --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 已执行：`D:\gol1\node.exe scripts/verify-setting-help-browser-smoke.mjs --check-only`
- 已执行：`D:\gol1\node.exe scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only`
- 已执行：`powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'check-only' -Driver 'cdp' -NodeSmokeTimeoutSeconds 30 ; exit $LASTEXITCODE"`
- 已执行：`powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'self-start' -Driver 'cdp' -NodeSmokeTimeoutSeconds 45 -KeepArtifacts ; exit $LASTEXITCODE"`
- 已执行：`powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'self-start' -Driver 'cdp' -NodeSmokeTimeoutSeconds 60 -KeepArtifacts ; exit $LASTEXITCODE"`

## 10. 风险与兼容性

- 新风险：低；本轮只改验证脚本和日志收集行为
- 兼容性风险：低；未改接口、持久化或页面业务逻辑
- 是否阻塞下一任务：部分阻塞；真实浏览器 smoke 仍受当前主机 Edge/CDP 宿主环境影响，但阻塞已进一步缩小

## 11. 收工记录

- 构建是否通过：脚本级静态检查通过；本轮未触达业务构建链
- 测试是否通过：
  - 通过：`verify-setting-help-browser-smoke.mjs --check-only`
  - 通过：`verify-setting-help-browser-smoke-cdp.mjs --check-only`
  - 通过：PowerShell wrapper `-Mode check-only -Driver cdp`
  - 失败但继续缩小：两次真实 `self-start + cdp` 运行均未成功完成浏览器 smoke
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、最近设置页 smoke worklog、automation memory、当前脚本源码与 trace 产物
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划与总账要求本轮继续停留在 Phase G 设置页帮助提示真实 smoke 同主线
  - 前几轮 worklog 明确上一轮已修好端口避让与 trace 基础，这轮应继续收紧 Node 子进程可观测性和 CDP page 路径诊断
  - automation memory 说明当前最值得推进的是同主线浏览器证据，不应再切回泛化分析
- 本次是否使用了子 agent 及其结论：无
- 手工 smoke 状态：已实际执行真实 `self-start + cdp`，但未形成通过证据
- 手工 smoke 阻塞原因 / 缺少的环境：
  - 第一轮真实运行证明阻塞不再是黑箱；trace 明确卡在 `json/new` 页面 websocket 上的 `Page.enable / Page.setLifecycleEventsEnabled / Runtime.enable` 连续超时
  - 第二轮真实运行证明新增的 `attached-session` 回退已实际触发；`json/new` 页面 websocket 关闭和 `Target.closeTarget` 都成功完成，新的 attached session 也能建立，但宿主环境下 `attached-session` 上的第一条 `Page.enable` 仍然卡死
  - 因此当前阻塞已从“page 连接路径不确定”进一步缩小到“这台主机的 Edge headless/CDP 对 Page domain 命令整体无响应”，而不是 wrapper、端口、trace、target 创建或 target attach 问题
- 真实产物目录：
  - `C:\Users\李昊桐\AppData\Local\Temp\octopus-setting-help-smoke-a19e9f4497414306ac8873dec07b6ce0`
  - `C:\Users\李昊桐\AppData\Local\Temp\octopus-setting-help-smoke-7ba240a77fae40a8a975839a5cb7e09b`
- 待验证页面清单：设置页 `LLMPrice / DynamicRouting / CircuitBreaker / ModelProbe` 的桌面布局、`375px` 布局、帮助按钮 focus/tooltip
- 遗留项：
  - 目前 Node smoke stdout/stderr 为空，说明阻塞仍在 CDP page-domain 命令阶段，下一轮不需要再排查 CLI 驱动或页面断言逻辑
  - 还未尝试把 `Page.enable` 换成更轻的附着后首命令或改用非 headless / 外部已有 Edge 会话进行对照
  - 前端 typecheck 仍受工作区既有 backup 相关语法错误阻断，本轮没有扩散去处理
- 下一任务前置条件是否满足：满足；wrapper、trace、Node 日志与 attached-session 回退都已落地，下一轮可直接围绕 Page domain 首命令对照实验继续推进

## 12. 下一轮建议

1. 继续留在 Phase G 同主线，优先对 attached-session 路径做更轻的首命令矩阵，例如先比较 `Target.activateTarget`、`Page.getFrameTree`、`Runtime.evaluate('1+1')` 与 `Page.enable` 的响应差异。
2. 若本机 headless Edge 仍然对 page domain 全面无响应，则用同一脚本链改测“外部已打开的非 headless Edge + 远程调试端口”路径，区分宿主 headless 策略问题和浏览器总体 CDP 能力问题。
3. 在拿到真实通过证据前，不把设置页四卡片浏览器 smoke 标记为完成；只把它视为“trace 完整、阻塞已收窄”的进行中状态。

## 13. 最终状态

- 本次结果：成功
- 是否需要人工介入：否

