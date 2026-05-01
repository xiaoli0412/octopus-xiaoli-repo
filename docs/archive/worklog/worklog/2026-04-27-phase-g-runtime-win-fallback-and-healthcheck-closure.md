# 2026-04-27 Phase G Runtime Win Fallback And Healthcheck Closure

## 1. 任务信息

- 任务名称：runtime-win fallback and healthcheck closure
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup browser-smoke host runtime diagnostics
- 对应 milestone：Phase G 当前窗口 backup browser-grade evidence 宿主运行态收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 验收标准 9、11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-backup-browser-wrapper-host-compat-closure.md`
  - `docs/worklog/2026-04-27-phase-g-backup-browser-smoke-script-closure-and-host-blocker-capture.md`
- 本次任务目标：在 backup browser evidence 继续受宿主 Winsock 阻塞时，先把 `scripts/runtime-win.ps1` 修到可稳定诊断 `status / stop / healthcheck`，减少下一轮 self-start / external smoke 的噪音
- 本次已盘点本地资源：AGENTS.md、canonical plan、状态文档、workflow、上一轮 memory、最近 backup/runtime worklogs、`scripts/runtime-win.ps1`、`scripts/use-go-env.ps1`、`scripts/verify-go-env.ps1`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase G backup browser evidence 主线
- 只修共享运行态脚本，不扩散到其他 UI 页面或 AI 自动化主线
- 不改 backup 页面业务逻辑、selector 契约或文案
- 优先把脚本失败收敛成真实宿主阻塞，而不是继续放大脚本自身问题

## 4. 本次禁止事项

- 不对用户现有未提交业务代码做回退
- 不在 `status` 探针不可靠时盲目执行广义停服
- 不把宿主 Winsock 问题误报成 backup 页面回归

## 5. 本次验收条件

- `scripts/runtime-win.ps1 -Action check-only` 能显示真实可用的 Go 路径，不再只依赖 PATH 猜测
- `scripts/runtime-win.ps1 -Action status` 在 `Get-CimInstance` 与 `Get-NetTCPConnection` 不可用时仍能稳定输出 fallback 状态
- `scripts/runtime-win.ps1 -Action stop` 在当前低权限 fallback 模式下不报错
- `scripts/runtime-win.ps1 -Action healthcheck` 不再卡在 Go 环境变量缺失，而是暴露真实服务 / 宿主网络状态

## 6. 本次回滚点

- 仅回退 `scripts/runtime-win.ps1` 即可撤销本轮改动

## 7. 实现范围

- 先改数据语义还是先改 UI：先改脚本回退链与 Go 环境兜底，再回跑最小运行态验证
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响脚本：`scripts/runtime-win.ps1`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强 Windows 宿主下的运行态诊断与 healthcheck 可执行性

## 8. 实施步骤

1. 阅读 `runtime-win.ps1`、`use-go-env.ps1`、`verify-go-env.ps1`，确认当前 `status` 在本机因为缺少 `Get-NetTCPConnection` 会直接失真，`healthcheck` 因 `GOCACHE / LocalAppData` 缺失而不是服务状态失败
2. 在 `runtime-win.ps1` 中新增 `Resolve-GoCommandPath`，优先解析 repo-local `.tools/go/go/bin/go.exe`
3. 在 `runtime-win.ps1` 中新增 `Get-ListeningPortSnapshot`，优先用 `Get-NetTCPConnection`，失败时退到 `.NET IPGlobalProperties.GetActiveTcpListeners()`，至少保持端口监听状态可见
4. 把 `Get-OctopusRepoProcess` 和 `Get-PortStatus` 接到新的 fallback 链，并在 `Show-Status` 中显示端口扫描模式
5. 在 `healthcheck` 路径中新增 workspace 级 Go 缓存目录兜底，自动补齐 `TEMP / TMP / GOCACHE / GOTMPDIR / LOCALAPPDATA`
6. 收紧 fallback 下对 repo-local Go 工具链本身的误识别，避免把 `.tools\go\go\bin\go.exe` 当成运行中的 Octopus 服务
7. 回跑 `check-only / status / stop / healthcheck`，确认脚本失败已收敛到真实宿主 Winsock/service-provider 问题

## 9. 测试与验证

- `& .\scripts\runtime-win.ps1 -Action check-only`
- `& .\scripts\runtime-win.ps1 -Action status`
- `& .\scripts\runtime-win.ps1 -Action stop`
- `& .\scripts\runtime-win.ps1 -Action healthcheck`
- `git diff --check -- scripts/runtime-win.ps1`

## 10. 风险与兼容性

- 兼容性风险：低；改动只在 Windows 运行态脚本，且全部为 fallback / 诊断增强
- 仍存宿主阻塞：`healthcheck` 现已能跑到真实请求阶段，但对 `127.0.0.1:8080/healthz` 的访问仍报 `The requested service provider could not be loaded or initialized`
- 是否阻塞下一任务：部分阻塞；backup browser evidence 仍受宿主 Winsock 问题影响，但当前脚本已能可靠区分“脚本故障”与“宿主网络栈故障”

## 11. 收工记录

- 构建是否通过：未跑全量构建；本轮只做 runtime script fallback 收口
- 测试是否通过：部分通过
  - 通过：`runtime-win.ps1 -Action check-only`
  - 通过：`runtime-win.ps1 -Action status`
  - 通过：`runtime-win.ps1 -Action stop`
  - 失败但已准确定位为宿主阻塞：`runtime-win.ps1 -Action healthcheck`
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、上一轮 memory、recent backup/runtime worklogs、`runtime-win.ps1`、`use-go-env.ps1`、`verify-go-env.ps1`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划 / workflow：确认当前仍应停留在 Phase G backup browser evidence 主线，不扩散
  - 上一轮 memory / worklog：确认 backup browser smoke 的主要噪音已从 wrapper 进入共享运行态诊断脚本
  - `runtime-win.ps1`：确认当前主要缺口是 `Get-NetTCPConnection` 不存在时脚本自身失效，以及 `healthcheck` 对 Go 环境变量过于乐观
  - `use-go-env.ps1` / `verify-go-env.ps1`：提供 repo-local Go 工具链与缓存目录布局，可直接复用到 runtime 脚本
- 手工 smoke 状态：未关闭真实 browser-grade backup evidence
- 手工 smoke 阻塞原因 / 缺少的环境：当前 Windows 宿主对 `127.0.0.1` 的本地 socket / HTTP 请求仍会报 service-provider 初始化失败
- 待验证页面清单：backup 页真实 browser-grade external/self-start smoke、`375px` 页面可读性、history/rollback 真实交互
- worklog 是否更新：是
- 遗留项：
  - backup browser-grade evidence 仍未闭环，主阻塞已收敛为宿主 Winsock / service-provider 问题
  - 若下一轮 `external` 也受同类问题影响，应继续在共享运行态 / self-start 诊断层推进，而不是回到 backup 页面 selector 层空转
  - 仓库内既有 `web/src/components/modules/ai-automation/index.tsx` 全量 `tsc` blocker 仍未处理，非本轮引入
- 下一任务前置条件是否满足：满足；下一轮可继续同一主线，优先尝试 `verify-backup-browser-smoke.ps1 -Mode external`，若仍受同类阻塞，则继续收口 shared runtime / self-start 诊断
