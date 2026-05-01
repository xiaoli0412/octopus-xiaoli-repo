# 2026-04-27 深度审查：更新重启顺序与 shutdown 空指针保护

## 背景

- 本轮自动深审承接上一轮 `internal/update` 路径审查，继续聚焦自更新失败链路和运行时收口。
- 直接复核了 `internal/update/core.go`、`internal/utils/shutdown/shutdown.go`、相关测试与上一轮审查结论，优先处理能在低风险范围内直接修复的发布阻塞项。

## Findings 与处理

1. `internal/update/core.go`
   - 问题：Windows 自更新重启路径先调用 `shutdown.Shutdown()`，再尝试拉起替换进程。
   - 影响：如果替换进程启动失败，当前进程会先执行关闭回调，导致服务提前进入关闭态；同时上一轮还已证明 `shutdown` 未初始化时这里会直接 panic。
   - 证据：原实现中 `restartExecutable()` 进入后立即调用 `shutdown.Shutdown()`，随后才在 Windows 分支里调用 `startReplacementProcess(...)`。
   - 处理：已修复。现在 Windows 分支先尝试启动替换进程，只有成功后才执行 `shutdown.Shutdown()` 与 `exitProcess(0)`；非 Windows 分支仍保持 `exec` 前关闭资源。

2. `internal/utils/shutdown/shutdown.go`
   - 问题：`Shutdown()` 默认假定 `ilog` 已初始化，未初始化时会在错误日志或完成日志处触发空指针 panic。
   - 影响：不仅自更新失败路径会受影响，任何在测试或早期初始化阶段直接调用 `shutdown.Shutdown()` 的代码都会有崩溃风险。
   - 证据：原实现无 `ilog == nil` 判断，直接调用 `ilog.Errorf(...)` / `ilog.Infof(...)`。
   - 处理：已修复。`Shutdown()` 现在在写日志前检查 `ilog != nil`，同时保留已注册关闭函数的执行顺序。

## 本轮改动文件

- `internal/update/core.go`
- `internal/update/update_test.go`
- `internal/utils/shutdown/shutdown.go`
- `internal/utils/shutdown/shutdown_test.go`

## 验证

- 已执行：`gofmt -w internal/update/core.go internal/update/update_test.go internal/utils/shutdown/shutdown.go internal/utils/shutdown/shutdown_test.go`
- 已执行：`go test ./internal/update ./internal/utils/shutdown -run 'Test(RestartExecutableWindowsDoesNotExitWhenRestartFails|RestartExecutableWindowsShutsDownAfterReplacementStarts|ShutdownWithoutLoggerDoesNotPanic|UnzipFailsWhenDirectoryEntryCollidesWithExistingFile|EnsureSafeExtractPathRejectsSymlinkAncestor|UnzipRejectsExistingSymlinkInDestination)' -count=1`
- 已执行：`git diff --check -- internal/update/core.go internal/update/update_test.go internal/utils/shutdown/shutdown.go internal/utils/shutdown/shutdown_test.go`
- 未执行：`go test ./internal/server/handlers -run TestCancelAITaskHandlerMarksTaskCanceled -count=20` 本轮因离线/网络受限环境触发 `github.com/tmaxmax/go-sse` 依赖下载失败，无法在当前宿主复验历史 finding。

## 剩余风险与下一轮建议

- `internal/op/ai_automation_executor.go` 仍使用 `context.Background()` 派生异步任务上下文，`AITaskCancel()` 也仍只改数据库状态，不会真正取消后台 goroutine；这条后台生命周期风险仍应作为下一轮高优先级项继续处理。
- 自更新链路仍缺少归档完整性/签名校验，供应链完整性依旧是更新路径剩余的最高风险点。
- `docker-compose.yml` 与 `scripts/dockerfiles/Dockerfile.debian` 仍未收紧到 `read_only/tmpfs/cap_drop/security_opt` 等最小权限运行面，容器逃逸缓解与误写防护仍有加固空间。
