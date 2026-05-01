# 2026-04-27 Phase G Windows Winsock Session Env Root Cause

## 1. 任务信息

- 任务名称：windows winsock session env root cause
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup browser-smoke host diagnostics
- 对应 milestone：Phase G 当前窗口 backup browser-grade evidence 宿主阻塞排查

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 验收标准 9、11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-runtime-win-fallback-and-healthcheck-closure.md`
  - `docs/worklog/2026-04-27-phase-g-backup-browser-wrapper-host-compat-closure.md`
- 本次任务目标：排查本机 `The requested service provider could not be loaded or initialized` 的宿主根因，判断是整机 Winsock catalog 损坏还是当前会话环境导致的 provider 初始化失败
- 本次已盘点本地资源：AGENTS.md、canonical plan、状态文档、workflow、上一轮 memory、recent Phase G worklogs、`scripts/runtime-win.ps1`、`scripts/verify-channel-create-browser-smoke.ps1`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase G backup browser evidence 主线
- 只做宿主网络栈 / Winsock / 会话环境排查与最小脚本修补
- 不扩散到无关业务逻辑或 AI 自动化主线

## 4. 本次禁止事项

- 不对系统执行 `winsock reset`、驱动卸载等高风险机器级改动
- 不把当前 agent 会话环境问题误判成仓库代码回归
- 不在结论不清时扩大到 unrelated cleanup

## 5. 本次验收条件

- 明确 `10106 / 无法加载或初始化请求的服务提供程序` 的触发条件
- 明确该问题是否属于整机 Winsock catalog 破坏，还是当前终端会话环境缺失
- 若能在脚本层低风险缓解，则直接落到共享运行态 / smoke wrapper 中
- 通过至少一条同主线验证证明根因判断正确

## 6. 本次回滚点

- 回退 `scripts/runtime-win.ps1` 与 `scripts/verify-channel-create-browser-smoke.ps1` 中的会话环境兜底即可

## 7. 实现范围

- 先改宿主脚本还是先改业务：先排查宿主，会话根因明确后只补脚本会话兜底
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响脚本：`scripts/runtime-win.ps1`、`scripts/verify-channel-create-browser-smoke.ps1`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只补 PowerShell 会话内的核心 Windows 环境变量，降低 agent/自动化环境触发 Winsock provider 初始化失败的概率

## 8. 关键发现

1. `netsh winsock show catalog` 显示的 catalog provider 基本为标准 `mswsock.dll / wshqos.dll / wshtcpip.dll` 条目，未见明显第三方 LSP 污染，因此不像是整机 Winsock catalog 本体破坏。
2. 当前 agent 会话中 `SystemRoot`、`windir`、`APPDATA`、`LOCALAPPDATA`、`COMSPEC` 为空或缺失；同会话内直接 `TcpListener('127.0.0.1', 18081)` 会稳定报 `WSAEPROVIDERFAILEDINIT (10106)`。
3. 只要在同一会话中补回：
   - `SystemRoot=C:\Windows`
   - `windir=C:\Windows`
   - `APPDATA=C:\Users\李昊桐\AppData\Roaming`
   - `LOCALAPPDATA=C:\Users\李昊桐\AppData\Local`
   - `COMSPEC=C:\Windows\System32\cmd.exe`
   则同样的 `TcpListener` 绑定立刻成功，`Invoke-WebRequest` 与 `go healthcheck` 也从 `10106` 转为正常的“连接被拒绝/服务未启动”语义。
4. 这说明当前主阻塞不是整机 Winsock catalog 损坏，而是 **当前自动化/agent PowerShell 会话缺少核心 Windows 环境变量，导致进程内 Winsock provider 无法初始化**。
5. 在把同样的环境兜底补进共享脚本后，backup `self-start` 已不再卡在 Winsock，而是顺利推进到新的真实阻塞：Playwright 找不到 `msedge.exe` 的默认路径。

## 9. 实施步骤

1. 收集 `winsock catalog`、服务状态、路由表和 `ipconfig`，确认问题不像 catalog 被第三方 provider 污染。
2. 用纯 .NET `TcpListener`/`TcpClient` 直接复现 `10106`，绕过仓库代码与 Go/Node 干扰。
3. 检查当前会话环境变量，发现 `SystemRoot / windir / APPDATA / LOCALAPPDATA / COMSPEC` 缺失。
4. 在同一会话中手动补回这些变量，重新执行 `.NET bind` 与 `runtime-win.ps1 -Action healthcheck`，确认 Winsock 错误消失，转为正常 `connect refused`。
5. 将这一结论固化到共享脚本：
   - `runtime-win.ps1` 增加 `Ensure-CoreWindowsEnvironment`
   - `verify-channel-create-browser-smoke.ps1` 增加 `Ensure-CoreWindowsEnvironment`
6. 重新运行同主线验证，确认 backup self-start 不再被 `10106` 阻塞，而是推进到浏览器路径缺失这一下一层真实问题。

## 10. 测试与验证

- `netsh winsock show catalog`
- `Get-Service WinHttpAutoProxySvc, Dhcp, Dnscache, NlaSvc, Tcpip, BFE, mpssvc, Winmgmt`
- `ipconfig /all`
- `route print`
- 未补环境时：`.NET TcpListener('127.0.0.1', 18081)` 复现 `10106`
- 补环境后：`.NET TcpListener('127.0.0.1', 18081)` 通过
- 补环境后：`& .\scripts\runtime-win.ps1 -Action healthcheck` 由 `10106` 收敛为 `connect refused`
- 补环境后：`& .\scripts\verify-backup-browser-smoke.ps1 -Mode self-start -NodePath $env:NODEEXE -NodeSmokeTimeoutSeconds 180` 不再报 Winsock，推进到 `msedge.exe` 路径缺失
- `git diff --check -- scripts/runtime-win.ps1 scripts/verify-channel-create-browser-smoke.ps1`

## 11. 风险与兼容性

- 兼容性风险：低；只为当前 PowerShell 进程补齐标准 Windows 用户态环境变量，不改系统注册表、不改 Winsock catalog、不重置网络栈
- 仍存宿主风险：如果其他自动化入口也运行在缺失同类环境变量的会话中，仍可能复现 `10106`；因此共享脚本兜底是必要的
- 是否阻塞下一任务：不再以 Winsock 阻塞；当前主阻塞已前移为 `msedge.exe` 位置与 Playwright 浏览器解析

## 12. 收工记录

- 构建是否通过：未跑全量构建；本轮只做宿主排查与脚本会话环境收口
- 测试是否通过：通过
  - 通过：定位并复现 `10106` 的会话环境触发条件
  - 通过：补回核心环境变量后，Winsock 错误让位为正常 `connect refused`
  - 通过：backup self-start 不再被 Winsock 阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、上一轮 memory、recent Phase G worklogs、`runtime-win.ps1`、`verify-channel-create-browser-smoke.ps1`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划 / workflow：确认仍应围绕 Phase G backup browser evidence 主线推进
  - 上一轮 worklog / memory：确认此前 `10106` 已成为主阻塞，需要机器级排查
  - `runtime-win.ps1` / shared smoke wrapper：提供可直接承载会话环境兜底的位置
- 手工 smoke 状态：backup self-start 已越过 Winsock 阶段，但尚未形成浏览器级 pass
- 当前真实阻塞：Playwright 默认寻找 `C:\Users\李昊桐\AppData\Local\Microsoft\Edge\Application\msedge.exe`，而本机 Edge 实际位于 `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
- worklog 是否更新：是
- 遗留项：
  - 下一轮应直接修 `Browser` / `BrowserPath` 默认解析，优先指向 `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
  - 若还有其他自动化脚本直接依赖当前会话中的 `SystemRoot / APPDATA / LOCALAPPDATA`，应复用同一兜底函数
- 下一任务前置条件是否满足：满足；下一轮应继续同一主线，优先修浏览器路径解析并重跑 backup self-start
