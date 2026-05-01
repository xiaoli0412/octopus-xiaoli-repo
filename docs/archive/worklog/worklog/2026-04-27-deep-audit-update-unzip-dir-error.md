# 2026-04-27 深度审查：自更新解压目录创建错误收口

- Task: deep audit and low-risk fix for self-update unzip directory error handling
- Canonical refs:
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `$CODEX_HOME/automations/octopus/memory.md`
- Milestone / Phase: Phase A stability and update-path hardening
- Master plan aligned before coding (yes/no): yes

## 本轮直接复用的本地资源

- `$CODEX_HOME/automations/octopus/memory.md`：承接上一轮 `internal/update` 并发保护与重启失败保活结论，继续沿更新链深入，而不是重复扫已落地修复。
- `docs/review/审查/2026-04-27-131355-octopus-repo-complete-audit.md`：复核上一轮高优先级结论，确认 `verify-setting-info-logic` 在仓库约定 Node 22 下已可直接运行，避免继续把过时结论当成当前 blocker。
- `internal/update/update.go`、`internal/update/update_test.go`：作为本轮更新解压路径事实依据。

## Findings 与处理

1. `internal/update/update.go`
   - 问题：`unzip()` 遍历 zip 目录项时，对 `os.MkdirAll(fpath, os.ModePerm)` 的返回值完全忽略。
   - 影响：如果更新目标目录中预先存在与 zip 目录项同名的普通文件，或目录创建本身失败，更新流程会静默跳过目录创建错误并继续执行，可能把“半失败解压”推进为“更新成功”，留下不完整二进制/资源状态。
   - 证据：目录分支原先仅调用 `os.MkdirAll(...)` 后 `continue`，无错误检查；本轮新增回归用例可稳定复现“目标路径已有阻塞文件”时的失败场景。
   - 处理：已修复。现在目录项创建失败会立刻返回错误并终止解压，避免更新链误报成功。

2. 审查结论校正
   - 问题：上一轮审查把 `web/package.json` 中 `test:setting-info-logic` 记成 plain `node` 下必现断链。
   - 影响：如果继续沿用该结论，会让后续自动化重复追逐一个已经被当前实现收口的问题，稀释真正的高风险排查。
   - 证据：在仓库自带 `scripts/use-node-env.ps1` 提供的 Node 22.14.0 环境下，直接执行 `scripts/verify-setting-info-logic.mjs` 已输出 `setting-info-logic verification passed`。
   - 处理：本轮未改该链路代码，但已在 worklog / memory 中校正，不再把它列为当前明确缺陷。

## 本轮改动文件

- `internal/update/update.go`
- `internal/update/update_test.go`

## 验证

- 已执行：`gofmt -w internal/update/update.go internal/update/update_test.go`
- 已执行：`go test ./internal/update -run 'Test(UnzipFailsWhenDirectoryEntryCollidesWithExistingFile|UnzipRejectsExistingSymlinkInDestination|EnsureSafeExtractPathRejectsSymlinkAncestor|ReadUpdateResponseBodyRejectsOversizedSuccessPayload|UpdateCoreRejectsConcurrentExecution)' -count=1`
- 已执行：`git diff --check -- internal/update/update.go internal/update/update_test.go`
- 已执行：`. ./scripts/use-node-env.ps1; & $env:NODEEXE scripts/verify-setting-info-logic.mjs`

## 剩余风险与下一轮建议

- 自更新链仍缺少归档完整性/签名校验；在尺寸限制、解压边界、并发保护、目录创建失败收口之后，供应链完整性仍是更新路径最值得优先继续深审的点。
- Docker 运行面仍未收紧到 `read_only/tmpfs/cap-drop` 级别，且 compose 文件仍以默认可写根文件系统运行，适合下一轮继续检查最小权限运行面。
- `providers` 远端刷新虽然已有 pinned commit 和 embedded fallback，但信任边界仍依赖 GitHub raw 获取，后续仍值得继续审查可用性与来源完整性。
