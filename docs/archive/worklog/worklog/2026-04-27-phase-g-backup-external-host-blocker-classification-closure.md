# 2026-04-27 Phase G Backup External Host Blocker Classification Closure

## 1. 任务信息

- 任务名称：backup external host-blocker classification closure
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup browser-smoke host runtime diagnostics
- 对应 milestone：Phase G 当前窗口 backup browser-grade external preflight 收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase 7 UI / 备份导入主线 / 验收标准 11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-runtime-win-fallback-and-healthcheck-closure.md`
  - `docs/worklog/2026-04-27-phase-g-backup-browser-wrapper-host-compat-closure.md`
- 本次任务目标：沿同一 Phase G backup browser evidence 主线继续推进，优先尝试 `external` 路径；若宿主仍阻塞，则把共享 CLI wrapper 的 external 目标 URL 与 localhost 错误分类收紧到可直接复用的状态
- 本次已盘点本地资源：AGENTS.md、canonical plan、当前状态文档、详细工作流、前端主线状态、上一轮 automation memory、最近 backup/runtime worklogs、`scripts/verify-backup-browser-smoke.ps1`、`scripts/verify-backup-browser-smoke.mjs`、`scripts/verify-channel-create-browser-smoke.ps1`
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase G backup browser evidence 主线
- 只修共享 CLI wrapper，不扩散到 backup 页面业务逻辑、selector 契约或 AI 自动化主线
- 若 external 仍命中 localhost / Winsock 阻塞，必须把结果收敛成宿主网络栈问题，而不是回到 backup 页面层空转
- 保持本轮验证最小化且直接相关，只围绕 `check-only` 和 `external` 预检链路

## 4. 本次禁止事项

- 不修改 backup 页面组件、文案或 no-browser 验证脚本
- 不覆盖用户现有未提交业务改动
- 不把 localhost `service-provider` 故障继续包装成普通 timeout

## 5. 本次验收条件

- `verify-backup-browser-smoke.ps1 -Mode check-only` 在未显式传 `FrontendUrl` 时，能显示 external 默认复用 backend-served 静态壳的真实 URL
- `verify-backup-browser-smoke.ps1 -Mode external` 在当前宿主失败时，直接给出 localhost `service-provider` 分类，而不是只报模糊 timeout
- `git diff --check -- scripts/verify-channel-create-browser-smoke.ps1` 通过

## 6. 本次回滚点

- 仅回退 `scripts/verify-channel-create-browser-smoke.ps1` 即可撤销本轮改动

## 7. 实现范围

- 先改数据语义还是先改 UI：先改共享 wrapper 的 external 目标 URL 选择与 HTTP 预检错误分类，再回跑 focused verification
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响脚本：`scripts/verify-channel-create-browser-smoke.ps1`
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅收紧 external CLI smoke 的默认 URL 与宿主错误诊断输出

## 8. 实施步骤

1. 复核 `verify-backup-browser-smoke.*` 与共享 `verify-channel-create-browser-smoke.ps1`，确认 backup external 入口实际复用同一 CLI wrapper
2. 重新执行 `backup check-only` 与 `backup external`，确认当前主要噪音已收敛到两处：
   - external 默认仍盲猜 `http://127.0.0.1:3101`
   - localhost `service-provider` 故障被 `Wait-Http` 吃掉并最终只表现为 timeout
3. 在共享 wrapper 中新增：
   - `Get-ErrorMessageChain`
   - `Test-IsLocalUrl`
   - `Test-IsServiceProviderFailureMessage`
4. 收紧 `Wait-Http`：
   - 保留最后一次失败原因
   - 对 localhost `service-provider` 故障直接早失败并输出“宿主网络阻塞”分类说明
   - 非该类错误超时后也会带出最后失败信息
5. 调整 external 模式的 frontend URL 选择：若未显式传 `FrontendUrl`，则默认复用 backend URL，用于 backend-served 静态壳 smoke；仅在显式指定时才指向独立前端服务
6. 回跑 `backup check-only` 与 `backup external`，确认输出已经对齐本轮目标

## 9. 测试与验证

- `& .\scripts\verify-backup-browser-smoke.ps1 -Mode check-only`
- `./scripts/use-node-env.ps1; & .\scripts\verify-backup-browser-smoke.ps1 -Mode external -NodePath $env:NODEEXE -NodeSmokeTimeoutSeconds 60`
- `git diff --check -- scripts/verify-channel-create-browser-smoke.ps1`

## 10. 风险与兼容性

- 兼容性风险：低；改动仅在共享 CLI wrapper，且保留显式 `-FrontendUrl` 优先
- 行为变化：external 模式若未传 `FrontendUrl`，现在默认指向 backend-served 静态壳；如果后续目标是独立前端 dev server，必须显式传 `-FrontendUrl`
- 仍存宿主阻塞：`http://127.0.0.1:18081/healthz` 在当前 Windows 宿主仍会报 `The requested service provider could not be loaded or initialized`
- 是否阻塞下一任务：部分阻塞；backup browser-grade evidence 仍受宿主 localhost/Winsock 问题影响，但 wrapper 现在已能更准确地暴露阻塞而非误导到页面层

## 11. 收工记录

- 构建是否通过：未跑全量构建；本轮只做共享 CLI wrapper 的 external 诊断收口
- 测试是否通过：部分通过
  - 通过：`verify-backup-browser-smoke.ps1 -Mode check-only`
  - 通过：`git diff --check -- scripts/verify-channel-create-browser-smoke.ps1`
  - 失败但已准确定位为宿主阻塞：`verify-backup-browser-smoke.ps1 -Mode external`
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、当前状态文档、workflow、前端主线状态、上一轮 automation memory、recent backup/runtime worklogs、`verify-backup-browser-smoke.*`、`verify-channel-create-browser-smoke.ps1`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划 / workflow：确认本轮应继续停留在 Phase G backup browser evidence 主线，不扩散
  - automation memory / recent worklogs：确认上一轮已把阻塞范围收敛到共享 runtime / localhost 问题，这轮应继续缩小到 external 预检层
  - `verify-backup-browser-smoke.*`：确认 backup external 入口不需要新脚本基础设施，而是直接复用共享 CLI wrapper
  - `verify-channel-create-browser-smoke.ps1`：确认真正需要收口的是 external 默认 URL 与 `Wait-Http` 的宿主错误分类，而不是 backup 页面自身契约
- 手工 smoke 状态：未关闭真实 browser-grade backup evidence
- 手工 smoke 阻塞原因 / 缺少的环境：当前 Windows 宿主对 `127.0.0.1:18081` 的 HTTP 预检仍会报 `service-provider` 初始化失败
- 待验证页面清单：backup 页真实 browser-grade external/self-start smoke、`375px` 页面可读性、history/rollback 真实交互
- worklog 是否更新：是
- 遗留项：
  - backup browser-grade evidence 仍未闭环，当前主阻塞继续指向宿主 localhost / Winsock 层
  - repo-wide `tsc` 仍受既有 `web/src/components/modules/ai-automation/index.tsx` 语法问题阻塞，非本轮引入
- 下一任务前置条件是否满足：满足；下一轮可继续同一主线，优先在可正常处理 localhost HTTP 的宿主上复跑 `verify-backup-browser-smoke.ps1 -Mode external`，若仍命中同类问题，则继续推进 shared runtime / localhost 诊断，而不是回到 backup selector 层
