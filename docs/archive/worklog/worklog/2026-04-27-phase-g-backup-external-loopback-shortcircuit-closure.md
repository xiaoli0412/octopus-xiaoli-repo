# 2026-04-27 Phase G Backup External Loopback Shortcircuit Closure

## 1. 任务信息

- 任务名称：backup external loopback shortcircuit closure
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup browser-smoke host runtime diagnostics
- 对应 milestone：Phase G 当前窗口 backup browser-grade evidence 宿主运行态预检收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 验收标准 11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-runtime-healthcheck-loopback-shortcircuit-closure.md`
  - `docs/worklog/2026-04-27-phase-g-backup-external-host-blocker-classification-closure.md`
- 本次任务目标：继续停留在 Phase G backup browser evidence 宿主阻塞主线，把共享 `verify-channel-create-browser-smoke.ps1` 的 `external` 本地 HTTP 预检也收紧到 loopback 诊断层；当 localhost/Winsock 已明确坏掉时，直接短路成宿主阻塞，不再先做 `Wait-Http`
- 本次已盘点本地资源：AGENTS.md、canonical plan、当前状态文档、详细 workflow、环境 next plan、前端主线状态、上一轮 automation memory、最近 runtime/backup worklogs、`scripts/runtime-win.ps1`、`scripts/verify-backup-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档与脚本，以及本地 skill 文件 `using-superpowers` / `brainstorming`
- 若未使用部分本地资源或上下文，原因：本轮只收口共享 browser-smoke wrapper 的 external 预检，不涉及 backup 页面组件、业务接口或 AI 自动化主线
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase G backup browser evidence 主线
- 只修改共享 wrapper `scripts/verify-channel-create-browser-smoke.ps1`
- 不扩散到 backup 页面 selector、前端 UI、AI 自动化或其他无关主线
- `external` 的阻塞分类必须与 `check-only` 和 `runtime-win.ps1` 的 loopback readiness 结论一致

## 4. 本次禁止事项

- 不覆盖用户已有未提交业务改动
- 不把宿主 localhost/Winsock 问题重新包装成页面回归或 backup 脚本逻辑失败
- 不为了这轮小闭环去改 Node smoke 脚本或 backup 组件

## 5. 本次验收条件

- `scripts/verify-backup-browser-smoke.ps1 -Mode check-only` 继续可用并保留 loopback readiness 报告
- `scripts/verify-backup-browser-smoke.ps1 -Mode external` 在当前宿主上直接报宿主级 loopback/service-provider 阻塞
- `external` 不再先进入 `Wait-Http` 再通过 HTTP 错误反推宿主问题
- `git diff --check -- scripts/verify-channel-create-browser-smoke.ps1` 通过

## 6. 本次回滚点

- 仅回退 `scripts/verify-channel-create-browser-smoke.ps1` 即可撤销本轮改动

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享 browser-smoke wrapper 的 external 预检，再回跑 focused verification
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅让 localhost 已明确坏掉时的 external 模式提前失败

## 8. 实施步骤

1. 复核 `verify-backup-browser-smoke.ps1` 与共享 `verify-channel-create-browser-smoke.ps1`，确认 backup external 入口直接复用共享 CLI wrapper。
2. 确认 `check-only` 已能报告 loopback blocked，但 `external` 仍会继续执行两次 `Wait-Http`，带来多余 HTTP 噪音。
3. 在 `scripts/verify-channel-create-browser-smoke.ps1` 中新增 `Assert-LoopbackLocalHttpReady`，对 localhost/127.0.0.1 的 external backend/frontend URL 先做 loopback bind/client 预检。
4. 把 `external` 分支接到该短路预检：进入 `Wait-Http` 前先判断是否已命中 service-provider host blocker。
5. 回跑 backup `check-only` 与 `external`，确认本轮只改变 external 的失败时机和错误质量。
6. 记录本轮附带环境结论：当前会话里 `APPDATA` / `LOCALAPPDATA` 默认为空，会卡 repo-local `apply_patch` wrapper；嵌套 `powershell -File` 在本宿主会报内部加载错误 `8009001d`，应优先直接 `& script.ps1`。

## 9. 测试与验证

- 构建命令：未跑全量构建；本轮仅针对共享 wrapper
- 测试命令：
  - `& .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
  - `.\scripts\use-node-env.ps1; & .\scripts\verify-backup-browser-smoke.ps1 -Mode external -NodePath $env:NODEEXE -NodeSmokeTimeoutSeconds 60`
  - `git diff --check -- scripts/verify-channel-create-browser-smoke.ps1`
- 额外尝试但未采用为正式验证：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-backup-browser-smoke.ps1 -Mode external ...`
  - 上述两条在当前宿主触发内部 PowerShell 加载错误 `8009001d`，因此回退为当前会话内直接 `& script.ps1`
- 专项验证：`external` 现已在 loopback 预检阶段直接短路到宿主阻塞，不再先进入 `Wait-Http` 路径

## 10. 风险与兼容性

- 新风险：低；仅在 localhost loopback 已明确被 service-provider 阻塞时提前失败
- 兼容性风险：低；如果宿主 loopback 正常，`external` 仍按原路径执行 HTTP preflight
- 是否阻塞下一任务：部分阻塞；backup browser-grade external/self-start evidence 仍受当前宿主 localhost/Winsock 问题影响，但下一轮不用再通过 HTTP/Node 链路反推根因

## 11. 收工记录

- 构建是否通过：未跑全量构建；本轮只做 shared wrapper focused 收口
- 测试是否通过：部分通过
  - 通过：`verify-backup-browser-smoke.ps1 -Mode check-only`
  - 失败但已按预期短路为宿主阻塞：`verify-backup-browser-smoke.ps1 -Mode external`
  - 通过：`git diff --check -- scripts/verify-channel-create-browser-smoke.ps1`
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、当前状态文档、workflow、环境 next plan、前端主线状态、上一轮 automation memory、recent runtime/backup worklogs、`scripts/runtime-win.ps1`、`scripts/verify-backup-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke.ps1`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - canonical plan / workflow / 环境 next plan：确认本轮仍应停留在 Phase G backup browser evidence 宿主诊断主线，不扩散到页面层
  - automation memory / recent worklogs：确认上一轮已把 runtime `healthcheck` 收紧到 loopback 预检，这轮应继续补齐 shared browser wrapper 的 external 预检
  - `verify-channel-create-browser-smoke.ps1`：确认真实缺口是 external 仍晚于 loopback 预检才暴露宿主阻塞
  - `verify-backup-browser-smoke.ps1`：确认 backup 入口无需新增独立逻辑，只要复用共享 wrapper 即可同步收口
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：backup 页真实 browser-grade evidence 仍未关闭
- 手工 smoke 阻塞原因 / 缺少的环境：当前 Windows 宿主对 `127.0.0.1` 的 socket bind / TcpClient 初始化都报 `无法加载或初始化请求的服务提供程序`
- 待验证页面清单：backup 页真实 browser-grade external/self-start smoke、`375px` 页面可读性、history/rollback 交互
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：是
- 遗留项：
  - backup browser-grade evidence 仍被当前宿主 localhost/Winsock 层阻塞
  - 当前会话里 `APPDATA` / `LOCALAPPDATA` 为空时，repo-local `apply_patch` wrapper 不能直接工作；需要先补齐环境变量
  - 当前宿主对嵌套 `powershell -File` 存在内部加载错误 `8009001d`，当前轮验证应继续优先使用当前会话内直接 `& script.ps1`
  - repo-wide `tsc` 仍受既有 `web/src/components/modules/ai-automation/index.tsx` 语法问题阻塞，非本轮引入
- 下一任务前置条件是否满足：满足；下一轮若 `runtime-win.ps1 -Action check-only` 与 backup `check-only` 仍显示 loopback blocked，则继续同一主线补 shared runtime / host diagnostics，不回到 backup 页面 selector 层空转；若宿主恢复可用，再优先重试 backup `external`，随后再试 `self-start`
