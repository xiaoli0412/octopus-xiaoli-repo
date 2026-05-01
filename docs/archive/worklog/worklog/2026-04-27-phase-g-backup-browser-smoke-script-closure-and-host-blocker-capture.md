# 2026-04-27 Phase G Backup Browser Smoke Script Closure And Host Blocker Capture

## 1. 任务信息

- 任务名称：backup browser smoke script closure and host-blocker capture
- 日期：2026-04-27
- 当前阶段：Phase G screenshot-first UI closure / backup page browser evidence closure
- 对应 milestone：Phase G 当前窗口 backup browser-grade evidence 收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 中 Phase 7 UI / 备份导入主线 / 验收标准 11、12、15
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1、8、9、11 节
- 上一个相关 worklog：
  - `docs/worklog/2026-04-27-phase-g-backup-history-rollback-metadata-attribute-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-node-entry-and-selector-contract-closure.md`
- 本次任务目标：把 backup browser smoke 从“脚手架已落地”推进到“脚本逻辑收口 + 真实阻塞点清晰可复现”，并同步清掉 `backup-component` 中残留的 selector 漂移
- 本次已盘点本地资源：AGENTS.md、canonical plan、前端主线状态、workflow、环境 next plan、最近 backup worklogs、automation memory、`Backup.tsx`、`Backup.test.tsx`、`internal/op/backup.go`、`internal/server/handlers/user.go`、`internal/server/middleware/auth.go`、`scripts/verify-backup-browser-smoke.mjs`、`scripts/verify-backup-browser-smoke.ps1`、`scripts/verify-backup-component.cjs`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档与文件；session 开头读取了 `using-superpowers` 与 `brainstorming` 本地 skill 文件；automation memory 用于继承上一轮 backup selector-contract 和 browser smoke 铺路状态
- 若未使用部分本地资源或上下文，原因：本轮不需要扩展到其他设置卡片、AI 自动化主线或全局构建清理，因为目标只锁定 backup browser smoke
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：用户明确要求不要创建子 agent，且本轮任务为 backup 验证链的主线程强耦合收口

## 3. 本次硬规则

- 继续停留在 Phase G backup 主线，不扩散到其他页面或 AI 自动化模块
- 优先让 backup browser smoke 与 no-browser verifier 使用同一套 selector / 数据合同
- external 模式不得偷偷修改外部管理员密码
- Windows 上所有 Node 验证继续统一走 `./scripts/use-node-env.ps1`

## 4. 本次禁止事项

- 不改 backup 业务语义和后端导入/回滚逻辑
- 不处理与 backup 无关的全局 `tsc` / Vitest / AI automation 问题
- 不通过 destructive git 命令清工作区

## 5. 本次验收条件

- `scripts/verify-backup-component.cjs` 中残留的 `Run Import` 文本选择器切到稳定 `backup-import-button`
- `scripts/verify-backup-browser-smoke.mjs` 收口响应 envelope、首次登录 gate、上传文件注入与等待逻辑
- `./scripts/use-node-env.ps1; node scripts/verify-backup-component.cjs` 通过
- `./scripts/use-node-env.ps1; node scripts/verify-backup-logic.mjs` 通过
- browser smoke 至少能产出真实、单点、可复现的 host blocker，而不是继续停留在脚手架未收口状态

## 6. 本次回滚点

- 仅回退 `scripts/verify-backup-component.cjs` 与 `scripts/verify-backup-browser-smoke.mjs` 的本轮改动即可

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 verifier / smoke 脚本合同，再回到运行态验证
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：只影响 browser smoke 对现有 `/api/v1/user/login`、`/api/v1/user/force-change-password`、`/api/v1/setting/import-snapshots`、`/api/v1/setting/export` 的调用方式
- 是否影响旧数据：否
- 是否影响旧行为：否；只影响本地验证链

## 8. 实施步骤

1. 复读 canonical / workflow / recent worklog / memory，并确认本轮仍应锁定 Phase G backup browser smoke 收口
2. 把 `scripts/verify-backup-component.cjs` 的残余 `Run Import` 文本选择器改成 `screen.getByTestId('backup-import-button')`
3. 修 `scripts/verify-backup-browser-smoke.mjs`：
   - 登录与管理 API 统一解包 `resp.Success.data`
   - `self-start` 可自动处理 bootstrap admin 首次改密，`external` 模式则显式报阻塞，不自动改外部管理员密码
   - 通过 base64 注入 `File`，删除对浏览器端再 fetch 导出的脆弱依赖
   - 等待 backup 页时主动滚动 settings 容器，而不是只滚 `window`
   - 自启动时如果前后端子进程早退，直接把 stdout/stderr 写入报错
4. 跑 no-browser focused verification，并尝试真实 browser smoke
5. 若 browser smoke 仍失败，则把失败点收敛为明确宿主 blocker 并写回 worklog / memory

## 9. 测试与验证

- 构建命令：
  - `./scripts/use-node-env.ps1; node --check scripts/verify-backup-browser-smoke.mjs`
- 测试命令：
  - `./scripts/use-node-env.ps1; node scripts/verify-backup-component.cjs`
  - `./scripts/use-node-env.ps1; node scripts/verify-backup-logic.mjs`
- 专项验证：
  - `./scripts/use-node-env.ps1; node scripts/verify-backup-browser-smoke.mjs --self-start`
  - `./scripts/use-node-env.ps1; $env:OCTOPUS_UI_SMOKE_FRONTEND_URL='http://127.0.0.1:18141'; $env:OCTOPUS_UI_SMOKE_BACKEND_URL='http://127.0.0.1:18141'; node scripts/verify-backup-browser-smoke.mjs --external`
  - `& .\build\octopus-smoke.exe start --config .\.tmp\backup-smoke-external\config.json`
  - `& .\build\octopus-smoke.exe start --config .\.tmp\backup-smoke-external\config-0.0.0.0.json`
  - `git diff --check -- scripts/verify-backup-component.cjs scripts/verify-backup-browser-smoke.mjs`

## 10. 风险与兼容性

- 新风险：`self-start` 模式会在 bootstrap admin 使用默认 `admin/admin` 时自动改成 smoke 专用密码；但该行为现在只限 `self-start`，`external` 已显式禁止自动改外部环境密码
- 兼容性风险：低；改动只在本地验证脚本，不影响产品功能
- 是否阻塞下一任务：部分阻塞；backup browser-grade evidence 仍被宿主运行态问题卡住，但脚本逻辑本身已基本收口

## 11. 收工记录

- 构建是否通过：通过；`node --check scripts/verify-backup-browser-smoke.mjs` 通过
- 测试是否通过：部分通过；`verify-backup-component.cjs`、`verify-backup-logic.mjs` 通过，browser smoke 仍被宿主运行态阻塞
- 本次使用了哪些本地资源 / skills / 记忆上下文：AGENTS.md、canonical plan、前端主线状态、workflow、环境 next plan、recent backup worklogs、automation memory、`internal/op/backup.go`、`internal/server/handlers/user.go`、`internal/server/middleware/auth.go`、`scripts/verify-backup-browser-smoke.mjs`、`scripts/verify-backup-component.cjs`、本地 skill 文件 `using-superpowers` / `brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - canonical / workflow / 前端主线状态：确认当前仍应优先闭合 Phase G backup browser evidence，而不是扩散到其它主题
  - `internal/op/backup.go`：再次确认只有真实 apply 才会写 import snapshot history，dry-run 不会写历史
  - `user.go` / `auth.go` / `middleware/auth.go`：确认首次登录改密 gate 的真实行为，并据此把 self-start / external 行为分开处理
  - automation memory：确认上轮已经完成 selector / metadata contract 收口，这轮可直接转入 browser smoke 逻辑收口
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：已尝试 browser-grade smoke，但未跑通
- 手工 smoke 阻塞原因 / 缺少的环境：
  - `node scripts/verify-backup-browser-smoke.mjs --self-start` 当前在本机会于 `spawnProcess(startBackend)` 阶段直接报 `spawn EPERM`
  - 前台直接执行 `build/octopus-smoke.exe start --config ...` 时，后端会成功读配置和静态目录，但在 `net.Listen` 阶段报 `socket: The requested service provider could not be loaded or initialized`，无论 host 是 `127.0.0.1` 还是 `0.0.0.0`
  - 因此 browser smoke 当前剩余 blocker 已明确收敛为宿主机级进程/网络提供程序问题，而不是 backup 页面逻辑未收口
- 待验证页面清单：backup 页真实 browser-grade pass（dry-run -> apply -> history -> rollback preview -> 375px）
- 若未使用子 agent，原因：用户明确禁止，且任务为 backup 验证链单点收口
- worklog 是否更新：是
- 遗留项：
  - 宿主 `spawn EPERM` 阻塞 Node 自启动 backend/frontend 进程
  - 宿主 `net.Listen` / socket provider 初始化异常阻塞前台二进制起服
  - `Backup.tsx` 的既有 LF/CRLF warning 仍未处理
  - `web/src/components/modules/ai-automation/index.tsx` 的既有 `tsc` 阻塞仍是外部问题
- 下一任务前置条件是否满足：满足；下一轮可继续同一主线，从宿主运行态 blocker 旁路或诊断收口继续推进 browser evidence
