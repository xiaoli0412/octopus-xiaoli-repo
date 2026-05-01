# 2026-04-27 Deep Audit Update Extract Link Boundary

## 背景

- 本轮自动深审聚焦自更新链路、部署/容器边界与历史遗留高风险区。
- 复读了 automation memory、近期审查报告、`internal/update/*`、`internal/server/middleware/cors.go`、`internal/server/handlers/providers.go`、`internal/op/backup.go` 后，继续沿“远端输入落盘”方向深挖。

## Findings

1. `internal/update/core.go` 会把 GitHub release ZIP 直接解压到当前可执行文件所在目录。
2. `internal/update/update.go` 原先只校验压缩包条目名本身不能通过 `..` 路径穿越离开目标目录，但没有拒绝“目标目录里已存在的 symlink/junction/reparse point”。
3. 这意味着如果可执行目录下预先存在重解析目录，例如 `linked -> D:\outside`，更新包中的 `linked/payload.bin` 会在逻辑上仍通过 `isPathInDest()`，随后被 `os.OpenFile` 写入到链接指向的目录外位置。

## 处理

- 在 [`internal/update/update.go`](/D:/GPT-codex/octopus_repo/internal/update/update.go:159) 的 `extractFile()` 前增加 `ensureSafeExtractPath()` 保护。
- 新保护会逐级 `Lstat` 目标路径祖先目录，拒绝：
  - 普通 symlink；
  - Windows 上带 `FILE_ATTRIBUTE_REPARSE_POINT` 的特殊链接目录，例如 junction。
- 补充 [`internal/update/update_test.go`](/D:/GPT-codex/octopus_repo/internal/update/update_test.go:57) 中的 Unix symlink 回归测试，以及 [`internal/update/update_test.go`](/D:/GPT-codex/octopus_repo/internal/update/update_test.go:109) 中的 Windows junction 回归测试。

## 验证

- `gofmt -w internal/update/update.go internal/update/update_test.go`
- `go test ./internal/update -count=1`

## 结果

- 本轮确认并修补了一处真实的更新落盘边界缺陷，属于“压缩包路径检查已做，但目标目录链接链路未收口”的典型漏洞。
- 未扩大到更新签名、版本信任或发布流程重构，保持为低风险局部修复。

## 后续建议

- 下一轮可继续深审 `internal/update` 的信任模型，例如 release ZIP 是否需要额外完整性校验或版本回滚保护。
- `providers` 远端刷新仍依赖 GitHub raw 可达性，适合作为下一条 supply-chain/stability 线继续跟进。
