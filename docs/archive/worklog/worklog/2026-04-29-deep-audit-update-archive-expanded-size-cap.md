# 2026-04-29 深度审查：自更新归档展开体积上限

- Task: deep audit and low-risk fix for self-update archive expansion limits
- Canonical refs:
  - `$CODEX_HOME/automations/octopus/memory.md`
  - `docs/review/审查/2026-04-29-175155-octopus-repo-complete-audit.md`
  - `docs/worklog/2026-04-27-deep-audit-update-extract-link-boundary.md`
  - `docs/worklog/2026-04-27-deep-audit-update-concurrency-guard.md`
- Milestone / Phase: Phase A stability and update-path hardening
- Master plan aligned before coding (yes/no): yes

## 本轮直接复用的本地资源

- `automation memory`：沿用上一轮已经确认的更新链遗留风险，继续聚焦 `internal/update`，避免重复扫低价值区域。
- `docs/review/审查/2026-04-29-175155-octopus-repo-complete-audit.md`：确认“checksum/signature 缺失、Windows 直接落地替换”仍是高优先级未解项，本轮只处理可局部收口的资源耗尽边界。
- `internal/update/update.go`、`internal/update/update_test.go`：承接之前已收口的响应体限制、路径穿越、symlink/junction、目录碰撞与并发保护测试，继续补解压资源边界。

## Findings 与处理

1. `internal/update/update.go`
   - 问题：更新链虽然已经限制下载 ZIP 体积为 `128 MiB`，但解压阶段没有对总展开量做任何上限控制。
   - 影响：攻击者一旦能够影响 release 归档内容，或者上游归档异常膨胀，就可以让管理端在 `POST /api/v1/update` 时把大量压缩数据展开到磁盘，造成磁盘耗尽、更新时间异常拉长，甚至把可执行目录打到半更新状态。这属于典型的压缩包资源放大风险。
   - 证据：`UpdateCore()` 会直接将下载回来的 release ZIP 传给 `unzip(data, filepath.Dir(execPath))`；原 `unzip()` 只校验路径边界和链接边界，没有累计 `zip.File.UncompressedSize64`，也没有在 `io.Copy` 阶段限制总写入量。
   - 处理：已修复。新增 `maxUpdateExpandedBytes = 512 << 20`，在解压前预校验所有 ZIP 条目的累计未压缩体积，并在实际 `io.Copy` 时继续用剩余额度二次兜底，防止篡改元数据或异常条目把磁盘写爆。

2. `internal/update/update_test.go`
   - 问题：此前回归测试已覆盖响应体过大、目录冲突、symlink/junction、文件/目录碰撞等边界，但没有任何测试保护“单文件展开过大”与“多文件累计展开过大”两类场景。
   - 影响：后续如果有人去掉或放宽更新包解压限制，CI 不会第一时间拦住这类资源耗尽回归。
   - 证据：原测试文件没有针对 `UncompressedSize64` 或多条目累计体积的断言。
   - 处理：已补最小必要回归测试，分别覆盖单条目超过上限和多条目累计超过上限时必须拒绝解压，且目标目录不应留下落地文件。

## 本轮改动文件

- `internal/update/update.go`
- `internal/update/update_test.go`
- `docs/worklog/2026-04-29-deep-audit-update-archive-expanded-size-cap.md`

## 验证

- 已执行：`gofmt -w internal/update/update.go internal/update/update_test.go`
- 已执行：`git diff --check -- internal/update/update.go internal/update/update_test.go`
- 已执行：`. .\scripts\use-go-env.ps1; $env:GOPROXY='https://proxy.golang.org,direct'; $env:GOSUMDB='sum.golang.org'; go test ./internal/update -run 'Test(ReadUpdateResponseBodyRejectsOversizedSuccessPayload|ReadUpdateResponseBodyRejectsErrorStatusWithBodySnippet|ReadUpdateResponseBodyAcceptsSmallSuccessPayload|UpdateCoreRejectsConcurrentExecution|RestartExecutableWindowsDoesNotExitWhenRestartFails|RestartExecutableWindowsShutsDownAfterReplacementStarts|EnsureSafeExtractPathRejectsSymlinkAncestor|UnzipRejectsExistingSymlinkInDestination|UnzipRejectsSymlinkArchiveEntry|UnzipFailsWhenDirectoryEntryCollidesWithExistingFile|UnzipRejectsEntryLargerThanExpandedLimit|UnzipRejectsArchiveWhoseExpandedSizeExceedsLimit|UnzipDoesNotReplaceExistingDirectoryWithFile)' -count=1`
- 未执行：`go test ./...`。本轮改动局限于 `internal/update`，因此只做直接相关的最小必要回归。

## 剩余风险与下一轮建议

- 自更新链仍然没有 checksum/signature 验证，供应链完整性依旧是更新路径当前最高风险问题。
- Windows 自更新仍然直接在 live executable 目录解压替换，没有 staged replacement / rollback 目录，文件锁与部分替换风险仍在。
- `scripts/use-go-env.ps1` 虽已能修复缓存目录，但本轮运行时仍暴露出会继承坏的 `GOPROXY` 值；若要提升 Windows 本地可复跑性，下一轮可考虑把代理环境兜底也纳入脚本或验证入口。
