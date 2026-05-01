# 2026-04-27 Phase G Backup Browser Wrapper Host Compatibility Closure

## 1. 任务信息

- 任务名称：backup browser smoke wrapper host-compat closure
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup page browser-smoke entry tightening
- 对应 milestone：Phase G 当前窗口 backup browser-grade evidence 预备收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 备份导入主线 / 验收标准 11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-backup-history-rollback-metadata-attribute-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-node-entry-and-selector-contract-closure.md`
- 本次任务目标：优先推进 backup 页 browser-grade smoke，如果宿主受阻，则先收口共享 browser-smoke wrapper 的 Node / port / check-only 契约，减少下一轮复跑噪音
- 本次已盘点本地资源：AGENTS.md、canonical plan、状态文档、workflow、用户上下文总账、前端主线状态、上一轮 memory、最近 backup worklogs、`scripts/verify-backup-browser-smoke.ps1`、`scripts/verify-backup-browser-smoke.mjs`、`scripts/verify-channel-create-browser-smoke.ps1`、`Backup.tsx`、`Backup.test.tsx`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase G backup browser evidence 主线
- 只修 backup 复用 smoke wrapper 的宿主兼容性与验证入口，不扩大到无关页面
- 保持 backup 页面与 no-browser 合同不回退
- 优先复用 `./scripts/use-node-env.ps1` 的 Node 路径

## 4. 本次禁止事项

- 不扩散到其他设置页或 AI 自动化主线
- 不修改备份业务逻辑与页面文案契约
- 不在宿主 socket/service-provider 明确阻塞时继续盲目重试同一启动路径

## 5. 本次验收条件

- `verify-backup-browser-smoke.ps1 -Mode check-only` 在未预热环境下也能解析到可运行 Node，并输出准确的 check-only 摘要
- 共享 wrapper 不再把空字符串候选传给 Node 解析逻辑
- 共享 wrapper 的 `check-only` 结果与真实 smoke 会使用的 URL / backend binary / Node 路径一致
- `node scripts/verify-backup-component.cjs` 通过
- `node scripts/verify-backup-logic.mjs` 通过

## 6. 本次回滚点

- 仅回退 `scripts/verify-channel-create-browser-smoke.ps1` 即可撤销本轮改动

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享 smoke wrapper 的环境解析与 host fallback，再回跑 focused verification
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响脚本：`scripts/verify-channel-create-browser-smoke.ps1`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；只增强 wrapper 的 Node 选择、check-only 环境透传和端口探测弹性

## 8. 实施步骤

1. 阅读 backup browser smoke 入口与共享 `verify-channel-create-browser-smoke.ps1`，确认 backup smoke 本身已复用 channel-create 的 CLI wrapper
2. 直接跑 `check-only` 与 `self-start`，把问题收敛到两个脚本级缺口：未预热环境下误选 Codex 自带 Node、以及自举路径遇到宿主 service-provider/socket 错误后噪音过大
3. 在共享 wrapper 中新增可运行 Node 解析：跳过 Codex bundled node，优先使用显式 `NodePath` / `NODEEXE` / `NODE_BIN` / 常用本地 Node 路径
4. 把 `Resolve-FreeTcpPort` 搜索窗口从 20 扩到 200，并让 `Test-TcpPortAvailable` 在本机 service-provider 绑定异常时不再把端口误判为已占用
5. 把 `OCTOPUS_UI_SMOKE_FRONTEND_URL`、`OCTOPUS_UI_SMOKE_BACKEND_URL`、`OCTOPUS_UI_SMOKE_NODE`、`OCTOPUS_UI_SMOKE_BACKEND_BIN`、`OCTOPUS_UI_SMOKE_NPX_SCRIPT` 的环境透传前移到 `check-only` 之前，保证诊断输出和真实 smoke 使用同一组参数
6. 再次运行 `backup check-only` 与 backup focused no-browser 验证，确认脚本收口成立；保留 `self-start` 的宿主 socket 阻塞结论，不继续空转

## 9. 测试与验证

- `& .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- `./scripts/use-node-env.ps1; & .\scripts\verify-backup-browser-smoke.ps1 -Mode self-start -NodePath $env:NODEEXE -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
- `./scripts/use-node-env.ps1; node scripts/verify-backup-component.cjs`
- `./scripts/use-node-env.ps1; node scripts/verify-backup-logic.mjs`
- `git diff --check -- scripts/verify-channel-create-browser-smoke.ps1`

## 10. 风险与兼容性

- 新风险：共享 wrapper 现在会主动跳过 Codex bundled node；如果未来该 bundled node 恢复可用，需要重新评估是否保留该过滤规则
- 兼容性风险：低；改动仅在脚本层，且保持显式 `NodePath` 优先
- 宿主阻塞：当前 Windows 宿主对 `127.0.0.1:18081` 的绑定和 `Invoke-WebRequest` 访问都报 `The requested service provider could not be loaded or initialized`，因此 `self-start` / `external` 的真实 browser smoke 仍未闭环
- 是否阻塞下一任务：部分阻塞；backup browser-grade evidence 仍受宿主网络栈问题影响，但 wrapper 本身已更容易在可用宿主上复跑

## 11. 收工记录

- 构建是否通过：未跑全量构建；本轮只做 backup 共享 smoke wrapper 收口和 backup focused no-browser 验证
- 测试是否通过：部分通过
  - 通过：`verify-backup-browser-smoke.ps1 -Mode check-only`
  - 通过：`node scripts/verify-backup-component.cjs`
  - 通过：`node scripts/verify-backup-logic.mjs`
  - 失败且已定位为宿主阻塞：`verify-backup-browser-smoke.ps1 -Mode self-start`
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、状态文档、workflow、用户上下文总账、前端主线状态、上一轮 memory、backup worklogs、`verify-backup-browser-smoke.ps1`、`verify-backup-browser-smoke.mjs`、`verify-channel-create-browser-smoke.ps1`、`Backup.tsx`、`Backup.test.tsx`、本地 skill 文件 `using-superpowers` / `brainstorming` / `playwright`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划 / workflow / 用户总账：确认当前仍应优先停留在 Phase G backup browser evidence 主线
  - 上一轮 memory / worklog：确认应先尝试 browser-grade evidence，再在阻塞时回退到同主线脚本收口
  - `verify-backup-browser-smoke.*`：确认 backup smoke 已复用共享 CLI wrapper，不需新造基础设施
  - `verify-channel-create-browser-smoke.ps1`：确认真实问题在共享 wrapper 的 Node 解析与宿主启动兼容性，而不是 backup 页面自身契约
- 手工 smoke 状态：未关闭真实 browser-grade backup evidence
- 手工 smoke 阻塞原因 / 缺少的环境：本机对 `127.0.0.1:18081` 的 socket / HTTP 调用均返回 service-provider 初始化失败，导致自举服务无法稳定监听
- 待验证页面清单：backup 页真实 browser-grade self-start 或 external smoke、`375px` 页面可读性、history/rollback 交互在可用宿主上的真实验证
- worklog 是否更新：是
- 遗留项：
  - 共享 wrapper 已收口，但 backup browser-grade self-start 仍被宿主 Winsock / service-provider 问题阻塞
  - `runtime-win.ps1 -Action status` 在本机缺少 `Get-NetTCPConnection`，当前不能作为可靠状态探针
  - `web/src/components/modules/ai-automation/index.tsx` 的既有全量 `tsc` blocker 仍未处理
- 下一任务前置条件是否满足：满足；下一轮应优先在可用宿主或已有运行实例上尝试 `backup -Mode external`，若仍受同类阻塞，再转向统一运行态脚本的 service-provider 兼容修补
