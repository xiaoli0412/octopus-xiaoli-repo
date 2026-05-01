# 2026-04-25 Deep Audit Import Setting Validation

- Master plan aligned before coding: yes
- Scope: full-repo index refresh plus focused deep audit on admin write boundaries, AI automation trust boundaries, and backup/import setting persistence paths
- Reused local resources:
  - `AGENTS.md`
  - automation memory
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/worklog/2026-04-24-deep-audit-management-update-json-body-guard.md`
  - `docs/worklog/2026-04-24-deep-audit-management-json-body-cap.md`
  - `docs/review/瀹℃煡/2026-04-24-200615-octopus-repo-complete-audit.md`

## Finding

- `DBImportIncrementalWithOptions(...)` imported snapshot `settings` rows through direct upsert/replace helpers without reusing the normal `model.Setting.Validate()` contract.
- Impact: a crafted backup/import payload could persist invalid or unknown setting keys that the normal `/api/v1/setting/set` management path would reject, creating configuration drift and potentially destabilizing runtime behavior after cache refresh.
- Evidence:
  - import path went from `internal/server/handlers/setting.go` to `op.DBImportIncrementalWithOptions(...)`
  - settings branch in `internal/op/backup.go` used `createUpsertSettings(...)` / `replaceSettings(...)` directly
  - no pre-import validation existed for `dump.Settings`

## Fix

- Added `validateImportSettings(...)` in `internal/op/backup.go`.
- The helper now:
  - rejects unknown setting keys not present in `model.DefaultSettings()`
  - reuses `model.Setting.Validate()` for imported settings before any dry-run/apply proceeds
- Added focused regression coverage in `internal/op/backup_test.go` for:
  - invalid known setting value
  - unknown setting key

## Verification

- `gofmt -w internal/op/backup.go internal/op/backup_test.go`
- `go test ./internal/op -run 'TestDBImportIncrementalRejectsInvalidSettingValue|TestDBImportIncrementalRejectsUnknownSettingKey' -count=1`

Verification note:

- `gofmt` passed.
- Focused Go test execution was environment-blocked because this host still lacks the needed module cache in offline mode (`GOPROXY=off` caused module lookup failures for third-party deps). The failure was environmental, not an assertion failure from the new tests.

## Result

- Outcome: partial success
- Manual intervention needed: no immediate manual action required
- Next priority:
  - continue deep audit on remaining import/rollback trust boundaries and post-import validation semantics
  - re-run focused `internal/op` tests once module cache/network is available
  - keep AI Profile runtime-consumption gap as a separate higher-level architectural risk, but do not mix it into low-risk auto-fix scope

## 本轮补充

- 发现并收口了 `SettingSetString` / `SettingSetInt` 的持久化前校验缺口，避免非法设置值先落库再报错。
- 新增 `internal/op/setting_guard_test.go`，覆盖无效枚举值和负数整型值拒绝写入。
- 将 `setupOpTestDB` 补成真实启动顺序，确保 settings cache 在 op 级测试中先初始化。
- 本轮验证已通过 `go test ./internal/op -run ''TestSettingSetStringRejectsInvalidValueBeforePersisting|TestSettingSetIntRejectsNegativeValueBeforePersisting'' -count=1`。
