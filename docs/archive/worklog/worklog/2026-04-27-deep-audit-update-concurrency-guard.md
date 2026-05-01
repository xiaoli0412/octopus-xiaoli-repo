# 2026-04-27 深度审查：管理端自更新并发保护

- Task: deep audit and low-risk fix for management self-update concurrency
- Canonical refs:
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `$CODEX_HOME/automations/octopus/memory.md`
- Milestone / Phase: Phase A stability and security hardening
- Master plan aligned before coding (yes/no): yes

## 本轮直接复用的本地资源

- `$CODEX_HOME/automations/octopus/memory.md`：承接上一轮对 `internal/update` 解压边界的审查结论，继续补自更新链路稳定性。
- `docs/review/审查/2026-04-27-131355-octopus-repo-complete-audit.md`：核对前一轮高风险结论，避免重复报告已被当前代码修复的 AI 自动化问题。
- `internal/update/core.go`、`internal/update/update.go`、`internal/server/handlers/update.go` 及现有 `internal/update/update_test.go`：作为本轮更新链路事实依据。

## Findings 与处理

1. `internal/update/core.go`
   - 问题：管理端 `POST /api/v1/update` 调用 `UpdateCore()` 时没有任何并发互斥；多个管理员或重复点击可并发进入同一下载、解压、重启链路。
   - 影响：可能出现同一可执行目录被交叉写入、重复解压、竞争重启，导致自更新行为不确定，严重时可能把管理面卡在半更新状态。
   - 证据：`UpdateCore()` 原先直接执行 `getDownloadFilename -> doRequestWithFallback -> unzip -> go restartExecutable`，没有 `sync/atomic`、锁或 in-flight guard；处理器 `internal/server/handlers/update.go` 也会把所有失败统一当作 500，无法区分“真正异常”和“已有更新进行中”。
   - 处理：已修复。新增全局原子并发保护，重复触发时返回 `update.ErrUpdateInProgress`；处理器将其映射为 `409 Conflict`。同时补了更新包和处理器回归测试。

2. `internal/update/core.go`
   - 问题：Windows 上若替换进程启动失败，旧进程仍会无条件执行 `os.Exit(0)`。
   - 影响：更新失败不仅不会完成切换，还会把当前服务直接停掉，形成明确的管理面可用性故障。
   - 证据：`restartExecutable()` 的 Windows 分支原先在 `cmd.Start()` 报错后只记录日志，随后仍继续执行 `os.Exit(0)`。
   - 处理：已修复。现在只有在替换进程成功拉起后才退出当前进程；若启动失败，会保留当前进程并释放更新中的并发标记。

## 本轮改动文件

- `internal/update/core.go`
- `internal/update/update_test.go`
- `internal/server/handlers/update.go`
- `internal/server/handlers/update_test.go`

## 验证

- 已执行：`gofmt -w internal/update/core.go internal/update/update_test.go internal/server/handlers/update.go internal/server/handlers/update_test.go`
- 已执行：`go test ./internal/update ./internal/server/handlers -run "TestUpdate(CoreRejectsConcurrentExecution|FuncReturnsConflictWhenUpdateAlreadyRunning|FuncReturnsInternalServerErrorForUnexpectedFailure|RestartExecutableWindowsDoesNotExitWhenRestartFails)" -count=1`
- 已执行：`git diff --check -- internal/update/core.go internal/update/update_test.go internal/server/handlers/update.go internal/server/handlers/update_test.go docs/worklog/2026-04-27-deep-audit-update-concurrency-guard.md`

## 剩余风险与下一轮建议

- 自更新链仍缺少归档完整性/签名校验，当前只做了体积限制与解压边界保护，供应链完整性仍是后续高优先级主题。
- `internal/server/handlers/providers.go` 仍依赖 GitHub raw 获取 pinned providers 列表，虽然已有限时、重定向限制和 embedded fallback，但供应链/可用性设计仍适合继续深审。
- 导入/回滚链与后台定时任务本轮完成了全量扫描，但尚未对 `internal/task` 的运行时行为做更细的竞态级验证，下一轮可继续补该区域。
