# 1. Findings

## Critical

### 1. `internal/utils/httpx` breaks full-repo Go validation
- Evidence: `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild -GoTest` failed in this audit at `internal/utils/httpx/body.go:22` with `non-constant format string in call to fmt.Errorf`.
- Code reference: `fmt.Errorf(tooLargeMessage)` in `internal/utils/httpx/body.go:22`.
- This is not dead code. `ReadLimitedBody` is used by `internal/helper/fetch.go` and multiple outbound adapters including OpenAI, Gemini, Anthropic, Copilot, and Antigravity.
- Impact: `go build ./...` and `go test ./...` are not green, so the current workspace is not release-clean.

## High

### 2. `manual / ai_profile` switching looks surface-level and is not proven to drive runtime behavior
- Evidence: `internal/op/ai_automation.go` reads and writes `config_source_mode` and `active_ai_profile_id`.
- Evidence: `web/src/components/modules/setting/AIAutomationSource.tsx` only updates the setting value.
- Evidence: `web/src/components/modules/ai-automation/index.tsx` still builds effective runtime config from local/manual values and explicit `config_snapshot`, rather than resolving an active AI profile into runtime config.
- Evidence: `internal/op/ai_automation_executor.go` only includes `config_source_mode` and `active_ai_profile_id` in the AI context payload; no runtime consumer was found that changes execution config based on those values.
- Evidence: tests such as `TestAIProfileActivateHandlerSwitchesSettingsOnly` and `TestAIProfileActivateDoesNotOverwriteManualConfig` prove setting changes, but do not prove runtime source switching.
- Conclusion: this is a classic "looks implemented but main flow not wired" finding.

### 3. AI task config snapshots are in-memory only and are lost across restart
- Evidence: `internal/op/ai_automation_executor.go` uses `sync.Map` (`aiTaskRuntimeConfig`) to store runtime snapshots.
- Evidence: `AITaskCreate` stores snapshots with `storeAITaskRuntimeConfig` after creating the DB row, but `AITask` has no persisted config snapshot fields.
- Evidence: task startup deletes the in-memory snapshot on completion.
- Impact: if the process restarts while a task is pending or running, the database still has a task record but the execution snapshot is gone.
- Conclusion: the AI task pipeline works as a single-process async flow, but not as a durable recoverable task system.

## Medium

### 4. The audit target is a very dirty workspace, not just `HEAD`
- Evidence: this run observed `130` tracked status entries and `187` untracked status entries from `git status --short`.
- Evidence: `git diff --stat HEAD -- . ':(exclude).next' ':(exclude)node_modules' ':(exclude).tools'` shows 130 tracked files changed relative to `HEAD`.
- Evidence: `git rev-list --left-right --count origin/dev...HEAD` returned `0 22`, so committed `HEAD` is ahead of `origin/dev`, but the workspace also contains a large amount of uncommitted work.
- Impact: conclusions must distinguish committed baseline from working-tree candidate implementation.

### 5. Dynamic routing is genuinely wired in the current workspace, but runtime E2E proof is still incomplete
- Positive evidence: docs and README now explicitly describe the daily task as a `dynamic summary scan` that does not mutate runtime config.
- Positive evidence: `internal/server/handlers/stats_test.go` verifies the summary endpoint and `daily_summary_scan_no_runtime_mutation` basis.
- Positive evidence: this audit also passed `verify-dynamic-routing-help.mjs` and backend smoke.
- Gap: I did not find an HTTP-level end-to-end test that flips `dynamic_routing_mode`, sends a real relay request, and proves the live relay path plus audit fields change accordingly.

### 6. Some status docs are stale versus current code reality
- Evidence: `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` still says the AI automation center implementation has not started.
- But the current workspace already contains AI automation models, handlers, executor logic, frontend pages, and tests.
- Impact: planning docs can mislead future work by understating current implementation and hiding the real remaining gaps.

### 7. Validation coverage is much better, but still does not prove all key promises
- Passed this run: backend smoke, TypeScript noEmit, static web build, locale consistency, backup logic, backup component, dynamic-routing help, circuit-breaker help, and setting-info logic.
- Still unproven: real runtime consumption of active AI profiles, task restart durability, and dynamic-routing HTTP behavior changes under live relay traffic.

## Low

### 8. Repo hygiene still needs cleanup before release
- `git diff --check` still reports extra blank lines at EOF in `web/src/components/modules/channel/Form.tsx:2188` and `web/src/components/modules/group/Editor.tsx:1043`.
- `build-web-static.mjs` passed, but emitted repeated `baseline-browser-mapping` staleness warnings.

# 2. Completion Assessment

## Overall
- The current workspace is not an empty shell. Major functionality is real: backend main flow, dynamic routing, backup/import, AI automation basics, and a large amount of frontend work are all present.
- The repo is also not fully release-ready. One full-repo Go build blocker remains, and two high-risk AI automation gaps remain around runtime profile switching and task snapshot durability.

## Completed
- Backend main path is reachable and real. This run passed the Windows backend smoke flow, including `/healthz`, frontend shell, login, create channel/group/apikey, and `/v1/chat/completions`.
- Dynamic routing is genuinely implemented in this workspace, not just documented.
- Backup/import is real and includes replace/merge/map, preview token flow, rollback snapshot preview, and post-import validation.
- Frontend static export is working. TypeScript noEmit passed and `build-web-static.mjs` successfully exported and synced `static/out`.
- AI automation basic flow is real enough to create tasks, update progress, call AI, and save `AIProfile` results.

## Partially Completed
- `manual / ai_profile` configuration switching is only partially completed because the runtime consumer is not proven.
- AI task execution is only partially completed because config snapshots are not durable across restart.
- Dynamic routing verification is only partially completed because runtime HTTP-level E2E proof is still missing.
- Documentation sync is only partially completed because some status docs are stale.

## Not Completed
- Full-repo `go build ./...` and `go test ./...` are not green.
- Release-clean repo hygiene is not complete.
- Active AI profile driven runtime config switching is not demonstrated end-to-end.

## Suspected Surface-Level Implementations
- `config_source_mode` and `active_ai_profile_id` are the strongest surface-level candidates: the settings exist and can be switched, but no real runtime consumer was found.

# 3. Verification Summary

## Commands Run
- `git status --short --branch`
- `git branch -vv --all`
- `git log --oneline --decorate -n 12`
- `git tag --sort=-creatordate | Select-Object -First 10`
- `git rev-list --left-right --count origin/dev...HEAD`
- `git diff --stat HEAD -- . ':(exclude).next' ':(exclude)node_modules' ':(exclude).tools'`
- `git diff --check`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoBuild -GoTest`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- `D:\gol1\node.exe .\scripts\build-web-static.mjs`
- `D:\gol1\node.exe .\scripts\verify-locale-consistency.mjs`
- `D:\gol1\node.exe .\scripts\verify-backup-logic.mjs`
- `D:\gol1\node.exe .\scripts\verify-backup-component.cjs`
- `D:\gol1\node.exe .\scripts\verify-dynamic-routing-help.mjs`
- `D:\gol1\node.exe .\scripts\verify-circuit-breaker-help.mjs`
- `D:\gol1\node.exe .\scripts\verify-setting-info-logic.mjs`

## Verified Items
- Windows backend smoke passed.
- Frontend typecheck passed.
- Frontend static export passed.
- Locale consistency script passed.
- Backup logic and backup component scripts passed.
- Dynamic routing help script passed.
- Circuit breaker help script passed.
- Setting info logic script passed.
- Dynamic routing summary endpoint and no-mutation basis are backed by tests.

## Unverified Items
- Runtime consumption of `active_ai_profile_id`.
- AI task restart/recovery behavior.
- HTTP-level dynamic-routing mode change affecting live relay behavior and audit fields.
- Docker and broader environment validation were not rerun in this pass.

# 4. Comparison Notes

## Working Tree vs `HEAD`
- `HEAD` is `bfa27ae` on `feat/erguotou` and also tagged `v0.1.3`.
- The current working tree contains a large amount of tracked and untracked change beyond `HEAD`.
- This audit therefore applies primarily to the current workspace candidate implementation, not only to committed `HEAD`.

## Current Branch vs Stable Baseline
- `HEAD` is ahead of `origin/dev` by 22 commits and behind by 0 commits.
- However, the real audit target is broader than committed branch state because the workspace is dirty.

## Code vs README / Docs / Task Notes
- Dynamic-routing docs are much closer to code reality now, especially around the summary-only daily scan.
- AI automation status docs are partially stale and still understate the amount of implementation already present.
- Task notes for `manual / ai_profile` switching promise more than the runtime wiring currently proves.

## Code vs Tests / Build / Validation Coverage
- Coverage is significantly improved for backend smoke, frontend static build, backup/import scripts, and dynamic-routing/settings no-browser checks.
- Coverage is still insufficient to prove three important promises: runtime profile switching, task durability across restart, and live dynamic-routing behavior changes.
- Full-repo Go validation still fails, which outweighs many local green checks.

# 5. Top Next Actions

## Top Three Priorities
1. Fix `internal/utils/httpx/body.go:22` and restore green `go build ./...` and `go test ./...`.
2. Add a real runtime consumer for `config_source_mode / active_ai_profile_id`, then prove `manual -> ai_profile -> manual` behavior end-to-end.
3. Persist AI task config snapshots or add an equivalent restart-safe recovery mechanism.

## Suggested Next Actions
- Cleanly separate intended implementation from temporary workspace artifacts before the next release decision.
- After fixing the Go build blocker, rerun full Go validation and clean up `git diff --check` EOF issues.
- Add an integration test for active AI profile based runtime config resolution.
- Add an HTTP-level dynamic-routing E2E that proves mode changes affect real relay behavior and audit fields.
- Update stale status docs so planning reflects the real current implementation state.

# Run Summary
- Trigger time: `2026-04-24T15:34:52.3702701+08:00`
- Checks performed: repo state, branch/tag/log review, baseline diffs, Go validation, backend smoke, frontend typecheck, frontend static build, locale/backup/dynamic-routing/circuit-breaker/setting-info scripts.
- Files modified in this audit run: final markdown and HTML reports under `docs/review/瀹℃煡/`.
- Main issues found: full-repo Go build blocker, runtime AI profile switching not wired, in-memory-only AI task snapshots, dirty workspace, incomplete dynamic-routing E2E proof, stale status docs, and minor repo hygiene issues.
- Result: success.
- Manual intervention needed: yes, especially to confirm the intended working-tree boundary and prioritize the next fix wave.
