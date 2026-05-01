# 2026-04-22 Phase G Setting Help Browser Smoke PowerShell Entry

## 1. 任务信息

- 任务名称：设置页帮助提示浏览器 smoke 仓库级 PowerShell 入口收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：设置页四卡片帮助提示浏览器验收入口标准化

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3、10.1 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-dual-mode-prep.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-smoke-script-repair.md`
- 本次任务目标：
  - 新增仓库级 `scripts/verify-setting-help-browser-smoke.ps1`，避免真实浏览器 smoke 继续依赖长串临时命令
  - 让入口可自启后端与浏览器，并优先走当前主机更可能成功的 `CDP` 模式
  - 修正 `scripts/verify-setting-help-browser-smoke.mjs` 对 `npx` 路径的硬编码依赖，至少让 CLI 驱动具备“自动发现本机可用入口”的能力
  - 在当前主机上尽量推进真实桌面与 `375px` smoke；若仍失败，必须把阻塞收敛到环境级事实并记录
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-dual-mode-prep.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-smoke-script-repair.md`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - `scripts/smoke-win-backend.ps1`
  - `web/src/components/modules/setting/LLMPrice.tsx`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/src/components/modules/setting/ModelProbe.tsx`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态与前端主线状态、最近两条 Phase G worklog、automation memory、现有 CLI/CDP smoke 脚本、Windows smoke 脚本
- 若未使用部分本地资源或上下文，原因：本轮只收口设置页帮助提示真实浏览器入口，不扩散到备份、渠道、分组或后端业务语义
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，且本轮是同一脚本链路下的强耦合小闭环

## 3. 本次硬规则

- 只处理设置页帮助提示真实 smoke 入口和同主线验证链，不改业务逻辑
- 不把未跑通的浏览器级 smoke 写成已通过
- 本轮必须留下可复跑的仓库级入口，而不是再次停留在临时 PowerShell 命令

## 4. 本次禁止事项

- 不修改设置页四卡片的业务语义
- 不扩散到备份导入、渠道、多 key 或分组主线
- 不回退工作区中与本轮无关的用户已有修改

## 5. 本次验收条件

- 仓库内存在可直接复用的 `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp` 可运行
- `scripts/verify-setting-help-browser-smoke.mjs --check-only` 可运行，且不再依赖不存在的 `D:\gol1\npx.cmd`
- `node scripts/verify-model-probe-help.mjs` 通过
- `node scripts/verify-dynamic-routing-help.mjs` 通过
- `node scripts/verify-circuit-breaker-help.mjs` 通过
- 尝试真实浏览器 smoke，并把成功路径或环境级阻塞写清楚

## 6. 本次回滚点

- `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-setting-help-browser-smoke.mjs`
- `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-powershell-entry.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- automation memory `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 实现范围

- 先改验证入口，再做真实烟测尝试
- 受影响后端模块：无业务改动；仅复用现有 `build/octopus-smoke.exe` 和嵌入静态壳
- 受影响前端模块：无业务改动
- 受影响脚本：
  - `scripts/verify-setting-help-browser-smoke.ps1`
  - `scripts/verify-setting-help-browser-smoke.mjs`
- 是否影响旧数据：否
- 是否影响旧行为：否；仅增强与修复验证入口

## 8. 实施步骤

1. 复核主规划、当前状态、前端主线状态、用户上下文总账、automation memory 和现有 CLI/CDP smoke 脚本，确认本轮继续留在 Phase G 设置页帮助提示同主线。
2. 新增仓库级 `scripts/verify-setting-help-browser-smoke.ps1`，把后端、浏览器和 smoke 执行收口到单一入口。
3. 修正 `scripts/verify-setting-help-browser-smoke.mjs` 的 CLI 驱动，使其自动发现 `node` 和 `npx-cli.js`，不再写死不存在的 `D:\gol1\npx.cmd`。
4. 先跑 `check-only` 与 no-browser 验证，再在当前主机上尝试 `CDP` 和 `CLI` 两种真实 smoke。
5. 把可复跑入口、实跑结果和阻塞原因写回 worklog、前端主线状态和 automation memory。

## 9. 测试与验证

- 已执行：`powershell -ExecutionPolicy Bypass -File scripts/verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp`
- 已执行：`node scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only`
- 已执行：`node scripts/verify-setting-help-browser-smoke.mjs --check-only`
- 已执行：`node scripts/verify-model-probe-help.mjs`
- 已执行：`node scripts/verify-dynamic-routing-help.mjs`
- 已执行：`node scripts/verify-circuit-breaker-help.mjs`
- 已执行：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 已执行：`powershell -ExecutionPolicy Bypass -File scripts/verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp -KeepArtifacts`
- 已执行：`powershell -ExecutionPolicy Bypass -File scripts/verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -KeepArtifacts`

## 10. 风险与兼容性

- 新风险：低；本轮只改验证脚本
- 兼容性风险：低；未改接口、持久化或页面业务逻辑
- 是否阻塞下一任务：部分阻塞；真实浏览器 smoke 仍受当前主机环境限制，但脚本入口和后端静态壳链路已打通

## 11. 收工记录

- 构建是否通过：部分通过
- 测试是否通过：
  - 通过：`verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp`
  - 通过：`verify-setting-help-browser-smoke-cdp.mjs --check-only`
  - 通过：`verify-setting-help-browser-smoke.mjs --check-only`
  - 通过：`verify-model-probe-help.mjs`
  - 通过：`verify-dynamic-routing-help.mjs`
  - 通过：`verify-circuit-breaker-help.mjs`
  - 失败：`tsc --noEmit -p .\web\tsconfig.json`，失败点不在本轮改动，而是工作区既有的 `web/src/components/modules/setting/backup-logic.ts` 与 `web/src/components/modules/setting/Backup.tsx` 大量语法错误
  - 失败：真实浏览器 smoke 仍未形成通过证据
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、当前状态与前端主线状态、最近 Phase G worklog、automation memory、设置页四卡片组件、CLI/CDP smoke 脚本、Windows smoke 脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划与工作流明确本轮必须留在 Phase G 设置页帮助提示浏览器验收同主线
  - 前端主线状态和 automation memory 明确下一步优先项就是“仓库级 PowerShell 入口 + 真实桌面/375px smoke”
  - 现有 CLI/CDP 脚本说明两条驱动都值得复用，但需要仓库级入口把环境准备标准化
  - Windows smoke 脚本提供了后端静态壳与独立日志产物的可复用模式
- 本次是否使用了子 agent 及其结论：无
- 手工 smoke 状态：已实际尝试真实浏览器 smoke，但未形成通过证据
- 手工 smoke 阻塞原因 / 缺少的环境：
  - CLI 驱动仍受本机 `spawn EPERM` 限制，即便改成 `npx-cli.js` 入口也会在本机 `spawn` 层失败
  - CDP 驱动已推进到“后端启动 + Edge 远程调试启动 + 脚本开始驱动页面”的阶段，但 `WebSocket` 会提前关闭，当前主机的 Edge 调试稳定性仍不足
  - 真实静态壳已改为复用后端 `static/out`，不再依赖 `next dev`；因此 `next dev spawn EPERM` 已被验证脚本层绕开
  - 最近一次 CDP 自启产物目录：`%TEMP%/octopus-setting-help-smoke-7c36895ba135406fae9a43521313b7b4`
- 待验证页面清单：设置页 `LLMPrice / DynamicRouting / CircuitBreaker / ModelProbe` 的桌面布局、`375px` 布局、帮助按钮 focus/tooltip
- 遗留项：
  - `scripts/verify-setting-help-browser-smoke.mjs` 里仍残留一组未使用的乱码帮助按钮常量，后续可顺手清理
  - `tsc --noEmit -p .\web\tsconfig.json` 当前被工作区既有 backup 相关语法错误阻断，和本轮设置页 smoke 主线无直接关系
  - `ps1` 当前固定使用 `18081/9222`，若本机这些端口被占用，需要后续加随机端口或占用检测
- 下一任务前置条件是否满足：满足；仓库级入口已存在，下一轮可直接围绕浏览器/CDP 稳定性继续推进，不必重新拼接启动命令

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在同一主线，优先处理 `CDP websocket closed` 的稳定性问题，争取拿到真实桌面与 `375px` 验证证据
- 同主线候选顺序：
  1. 让 `scripts/verify-setting-help-browser-smoke.ps1` 自动避开被占用端口，并增加已有 `9222` 会话复用检测
  2. 调整 `scripts/verify-setting-help-browser-smoke-cdp.mjs` 的 `Target.createTarget / attachToTarget` 或连接保持策略，解决 websocket 提前关闭
  3. 清理 `verify-setting-help-browser-smoke.mjs` 中未使用的乱码帮助按钮常量
  4. 真实 smoke 成功后，把结果回写 `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md` 和 automation memory
