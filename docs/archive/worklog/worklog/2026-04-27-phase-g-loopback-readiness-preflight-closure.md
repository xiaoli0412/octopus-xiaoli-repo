# 2026-04-27 Phase G Loopback Readiness Preflight Closure

## 1. 任务信息

- 任务名称：loopback readiness preflight closure
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup browser-smoke host runtime diagnostics
- 对应 milestone：Phase G 当前窗口 backup browser-grade evidence 宿主预检收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 备份导入主线 / 验收标准 11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-runtime-win-fallback-and-healthcheck-closure.md`
  - `docs/worklog/2026-04-27-phase-g-backup-external-host-blocker-classification-closure.md`
- 本次任务目标：继续停留在 Phase G backup browser evidence 主线，把 localhost/Winsock 宿主阻塞前移到 `check-only / status / stop` 预检层，减少下一轮 external/self-start 的空转
- 本次已盘点本地资源：AGENTS.md、canonical plan、当前状态文档、详细 workflow、环境 next plan、前端主线状态、上一轮 automation memory、最近 runtime/backup worklogs、`scripts/runtime-win.ps1`、`scripts/verify-backup-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档与脚本，以及本地 skill 文件 `using-superpowers` / `brainstorming`
- 若未使用部分本地资源或上下文，原因：本轮只需要共享 runtime / smoke wrapper 诊断层，不需要扩散到 backup 页面业务逻辑与组件测试
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase G backup browser evidence 主线
- 只修共享 runtime / smoke wrapper 的预检诊断，不修改 backup 页面业务逻辑、selector 契约或 no-browser 验证
- 让 localhost/Winsock 阻塞在 `check-only / status / stop` 阶段就可见，而不是等 external/self-start 失败后再回溯

## 4. 本次禁止事项

- 不扩散到备份页面组件、文案、AI 自动化或其他无关主线
- 不覆盖用户已有未提交业务改动
- 不把宿主 Windows service-provider 问题重新包装成页面层 timeout 或 page regression

## 5. 本次验收条件

- `scripts/runtime-win.ps1 -Action check-only` 直接显示 localhost loopback readiness 结论
- `scripts/runtime-win.ps1 -Action status` 与 `-Action stop` 也显示同一组 localhost 诊断
- `scripts/verify-backup-browser-smoke.ps1 -Mode check-only` 直接显示 backup external/self-start 将受同一宿主阻塞影响
- `scripts/verify-backup-browser-smoke.ps1 -Mode external` 的失败与上述预检结论一致
- `git diff --check -- scripts/runtime-win.ps1 scripts/verify-channel-create-browser-smoke.ps1` 通过

## 6. 本次回滚点

- 仅回退 `scripts/runtime-win.ps1` 与 `scripts/verify-channel-create-browser-smoke.ps1` 即可撤销本轮改动

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享脚本预检输出，再回跑 focused verification
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强本机 localhost/self-start/external 的诊断输出，不改 backup 页面逻辑

## 8. 实施步骤

1. 复核 `runtime-win.ps1`、`verify-backup-browser-smoke.ps1` 与共享 `verify-channel-create-browser-smoke.ps1`，确认当前 external 真实阻塞仍是 localhost/Winsock，而 `check-only` 尚未提前暴露该结论。
2. 在 `scripts/runtime-win.ps1` 中新增 loopback bind/client capability 预检，并把报告接到 `check-only / status / stop` 公共输出层。
3. 在 `scripts/verify-channel-create-browser-smoke.ps1` 中新增同类 localhost 预检，并把报告接到 `check-only` 输出，供 backup wrapper 直接复用。
4. 回跑 `runtime-win.ps1 -Action check-only/status/stop`、`verify-backup-browser-smoke.ps1 -Mode check-only` 与 `-Mode external`，确认预检结论和 external 失败点完全对齐。

## 9. 测试与验证

- 构建命令：未跑全量构建；本轮仅针对共享 runtime / smoke wrapper 诊断脚本
- 测试命令：
  - `& .\scripts\runtime-win.ps1 -Action check-only`
  - `& .\scripts\runtime-win.ps1 -Action status`
  - `& .\scripts\runtime-win.ps1 -Action stop`
  - `& .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
  - `./scripts/use-node-env.ps1; & .\scripts\verify-backup-browser-smoke.ps1 -Mode external -NodePath $env:NODEEXE -NodeSmokeTimeoutSeconds 60`
  - `git diff --check -- scripts/runtime-win.ps1 scripts/verify-channel-create-browser-smoke.ps1`
- 专项验证：localhost loopback bind/client probe 与 external HTTP preflight 失败点一致，均指向 Windows service-provider 初始化失败

## 10. 风险与兼容性

- 新风险：低；仅新增诊断输出与 loopback 预检，不改变 backup browser smoke 的页面行为
- 兼容性风险：低；报告只在共享脚本层输出，本机外其他宿主如果 localhost 正常，会看到“usable for local smoke preflight”而不是阻塞提示
- 是否阻塞下一任务：部分阻塞；backup browser-grade evidence 仍受当前宿主 localhost/Winsock 问题影响，但下一轮已不必再先跑失败才能知道根因

## 11. 收工记录

- 构建是否通过：未跑全量构建；本轮只做 runtime / smoke wrapper 预检收口
- 测试是否通过：部分通过
  - 通过：`runtime-win.ps1 -Action check-only`
  - 通过：`runtime-win.ps1 -Action status`
  - 通过：`runtime-win.ps1 -Action stop`
  - 通过：`verify-backup-browser-smoke.ps1 -Mode check-only`
  - 通过：`git diff --check -- scripts/runtime-win.ps1 scripts/verify-channel-create-browser-smoke.ps1`
  - 失败但已准确定位为宿主阻塞：`verify-backup-browser-smoke.ps1 -Mode external`
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、当前状态文档、workflow、环境 next plan、前端主线状态、上一轮 automation memory、recent runtime/backup worklogs、`runtime-win.ps1`、`verify-backup-browser-smoke.ps1`、`verify-channel-create-browser-smoke.ps1`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划 / workflow / 当前状态文档：确认本轮仍应停留在 Phase G backup browser evidence 主线，不扩散到页面层或 AI 主线
  - 环境 next plan：提醒本机默认运行策略应保持停驻，并优先使用 runtime-win / shared wrapper 做短链预检
  - automation memory / recent worklogs：确认上一轮已经把 external 失败收敛为宿主 localhost/service-provider 阻塞，这轮应继续把诊断前移
  - `runtime-win.ps1`：确认当前低权限 fallback 已可用，适合把 loopback readiness 挂到统一运行态输出
  - `verify-channel-create-browser-smoke.ps1`：确认 backup smoke 复用该脚本，适合在 check-only 层直接补 localhost 预检报告
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：backup 页真实 browser-grade evidence 仍未关闭
- 手工 smoke 阻塞原因 / 缺少的环境：当前 Windows 宿主对 `127.0.0.1` 的 socket bind / TcpClient / HTTP preflight 都报 `无法加载或初始化请求的服务提供程序`
- 待验证页面清单：backup 页真实 browser-grade external/self-start smoke、`375px` 页面可读性、history/rollback 真实交互
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent
- worklog 是否更新：是
- 遗留项：
  - backup browser-grade evidence 仍被当前宿主 localhost/Winsock 层阻塞
  - 当前共享脚本虽然已提前暴露阻塞，但未修复宿主网络栈本身
  - repo-wide `tsc` 仍受既有 `web/src/components/modules/ai-automation/index.tsx` 语法问题阻塞，非本轮引入
- 下一任务前置条件是否满足：满足；下一轮应先看 `runtime-win.ps1 -Action check-only` 或 `verify-backup-browser-smoke.ps1 -Mode check-only` 的 loopback readiness，若仍是 blocked，则继续同一主线下的 shared runtime / host diagnostics，而不是回到 backup selector / 页面层空转；若在可用宿主看到 usable，再继续跑 backup external/self-start 真实 browser evidence
