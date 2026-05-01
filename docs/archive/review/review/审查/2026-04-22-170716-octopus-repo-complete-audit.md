# Octopus Repo Complete Audit

Triggered at: `2026-04-22T17:07:16+08:00`
Workspace: `D:\GPT-codex\octopus_repo`
Branch: `feat/erguotou`
HEAD: `bfa27ae` (`v0.1.3`)
Automation ID: `octopus-repo`

## 1. Findings

### Critical

1. The current workspace cannot compile because the new dynamic-routing relay patch is incomplete.
Evidence: `internal/relay/dynamic.go:82` calls `allowsRacingByPolicy(policy)`, but that function does not exist anywhere under `internal/relay`. Untracked tests also reference missing helpers such as `allowsRacingByDefault`, `isStreamingRequest`, `shouldEscalateToRace`, and `allowsRacingByModel`. `go build ./...`, `go test ./... -count=1`, and `go build -o .\\build\\octopus-smoke-probe.exe .` all fail on `internal\\relay\\dynamic.go:82:53: undefined: allowsRacingByPolicy`. The affected files are workspace-only additions: `internal/relay/dynamic.go`, `internal/relay/budget.go`, `internal/relay/probe.go`, and `internal/relay/relay_test.go` are not in `HEAD`.

### High

2. The race-based dynamic-routing flow is not wired into the live relay path.
Evidence: `effectiveDynamicRoutingTuning`, `runRaceFallback`, `acquireRaceBudgets`, and `recordProbeSuccessOutcomes` only appear in helper files and tests. `internal/relay/relay.go` still runs a simple sequential `iter.Next()` -> single attempt -> next candidate loop with no consecutive-failure accounting, no race launch, and no call into the new helpers. This is an implemented-looking feature that is not reachable from the real request path.

3. `scripts/smoke-win-backend.ps1` can produce a green smoke result without proving that the current source tree built successfully.
Evidence: the build step at `scripts/smoke-win-backend.ps1:105-116` runs `go build` but does not check `$LASTEXITCODE`. A reusable binary already exists at `build/octopus-smoke.exe` from `2026/4/22 08:10:14`. A fresh build of current sources fails, yet `powershell -ExecutionPolicy Bypass -File .\\scripts\\smoke-win-backend.ps1` still completes the runtime flow. The smoke run therefore proves that an existing binary and static assets still work, not that the current Go sources compile.

### Medium

4. Dynamic-routing docs still overstate what is implemented.
Evidence: `docs/DYNAMIC_ROUTING_REQUIREMENTS.md` and `docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md` still describe `incident-safe`, `hybrid`, `shadow-ai`, and `metrics-only` modes. Repository-wide search only finds those names in docs. The concrete implemented task is the daily summary scan in `internal/task/dynamic.go`, registered from `internal/task/init.go:46`.

5. Top-level release documentation is still not publish-ready.
Evidence: `README.md` and `README_zh.md` still contain `<CURRENT_IMAGE>`, `<CURRENT_REPOSITORY_URL>`, `<CURRENT_ONE_CLICK_INSTALL_SCRIPT>`, `<CURRENT_RELEASES_URL>`, and TODO placeholders.

6. Protocol conversion is real on the core path, but several edges are still TODO or placeholder implementations.
Evidence: `internal/transformer/model/model.go:142,208,618-619,861` and `internal/transformer/outbound/gemini/messages.go:321` still contain TODO markers around schema, audio/prediction-style fields, and JSON schema conversion.

### Low

7. `scripts/vitest-no-spawn.cjs` looks like a Windows test workaround, but it does not solve the local Vitest execution gap.
Evidence: it only monkey-patches `child_process.exec` for `net use`. `node .\\web\\node_modules\\vitest\\vitest.mjs run web\\src\\components\\modules\\setting\\backup-logic.test.ts` still fails with `spawn EPERM`.

## 2. Completion Assessment

Overall judgment: the repository is not a fake skeleton. The committed baseline (`bfa27ae` / `v0.1.3`) has a real backend management plane, real static shell delivery, real provider/channel/group flows, and a real gateway runtime. However, the current workspace is a large in-progress patch set, and the newest dynamic-routing relay work is both incomplete and not connected to the live relay path.

Status summary:
- Completed: backend management and gateway mainline, static asset sync, backup frontend static verification, locale parity for the current backup strings, dynamic summary-scan task.
- Partially completed: dynamic-routing phase-1 runtime work, broader protocol-conversion edges, final frontend production-build closure, full verification closure.
- Not completed: doc-promised advanced dynamic-routing modes, README release/install placeholders, current-workspace Go build/test closure.
- Suspected half-implementation: `scripts/vitest-no-spawn.cjs`.

## 3. Verification Summary

Verified:
- Repository structure, git status, recent commits, branch delta, README/docs, validation workflow, and key backend/frontend wiring were reviewed.
- `node .\\web\\node_modules\\typescript\\bin\\tsc --noEmit -p .\\web\\tsconfig.json` passed.
- `node .\\scripts\\verify-backup-logic.mjs` passed.
- `node .\\scripts\\verify-backup-component.cjs` passed.
- `node .\\scripts\\sync-web-static.mjs` passed.
- `powershell -ExecutionPolicy Bypass -File .\\scripts\\smoke-win-backend.ps1` passed against an existing binary.
- Current-workspace Go source build failure was reproduced directly with `go build ./...`, `go test ./... -count=1`, and `go build -o .\\build\\octopus-smoke-probe.exe .`.

Not verified:
- A Windows smoke run against a freshly built binary from the current sources.
- Linux smoke, Docker smoke, and browser/manual acceptance.
- A successful `next build` on a host without the current `spawn EPERM` limitation.
- Generic Vitest execution on this host.

Command snapshot:
- `go build ./...`: failed on `allowsRacingByPolicy` missing.
- `go test ./... -count=1`: failed on the same compile break.
- `go build -o .\\build\\octopus-smoke-probe.exe .`: failed.
- `powershell -ExecutionPolicy Bypass -File .\\scripts\\smoke-win-backend.ps1`: passed with the caveat above.
- `node .\\web\\node_modules\\typescript\\bin\\tsc --noEmit -p .\\web\\tsconfig.json`: passed.
- `node .\\scripts\\verify-backup-logic.mjs`: passed.
- `node .\\scripts\\verify-backup-component.cjs`: passed.
- `node .\\scripts\\sync-web-static.mjs`: passed.
- `node .\\node_modules\\next\\dist\\bin\\next build`: failed after compile with `spawn EPERM`.
- `node .\\web\\node_modules\\vitest\\vitest.mjs run ...`: failed with `spawn EPERM`.
- `pnpm -v`: not runnable.
- `corepack pnpm -v`: failed with Corepack `EPERM`.
- `docker --version`: not runnable.

## 4. Comparison Notes

Current workspace vs `HEAD`:
- The workspace is a very large in-progress change set with about 90 modified tracked files plus many untracked additions.
- The most severe Go build failure is caused by workspace-only files, not by committed `HEAD`.
- Compared with the earlier audit today, the backup/settings frontend parse failure and locale drift no longer reproduce; current `tsc`, backup logic verification, and backup component verification all pass.

Current branch vs stable baseline:
- Current branch is `feat/erguotou` at `bfa27ae` (`v0.1.3`).
- `git rev-list --left-right --count origin/dev...HEAD` returned `0 22`, so the branch is not behind `origin/dev` and is ahead by 22 commits.

Code vs README/docs/task claims:
- Consistent: backend startup, healthcheck, management mainline, backup frontend static validation, and the dynamic summary task are real.
- Inconsistent: README release/install placeholders remain; advanced dynamic-routing modes in docs do not exist in code; the new race/runtime work is not connected to `relay.go`.

Code vs tests/build/verification coverage:
- Frontend static verification is now consistent.
- Backend verification is inconsistent because Go build/test are red while Windows smoke can still be green via an old binary.
- Vitest, local production build completion, Linux smoke, and Docker smoke remain environment-limited.

## 5. Top Next Actions

Top three priority items:
- Repair or shelve the current dynamic-routing relay workspace patch until `go build ./...` and `go test ./...` pass again.
- Decide whether race fallback is actually shipping now; if yes, wire it into `internal/relay/relay.go` and add a real request-path regression test.
- Harden the verification chain so Windows smoke cannot pass on a stale binary when current-source build failed.

Suggested next steps:
- Restore Go compile health first.
- After compile health returns, add one regression test that proves race fallback is reachable from the live request path.
- Re-run verification on a host that has Docker and does not hit the current `spawn EPERM` limitation.
- Replace the README placeholders with real release/install values.

## Chinese Summary

This run completed successfully as an audit, but the current workspace is not releasable. The most important issue is an incomplete dynamic-routing relay patch that breaks `go build` and `go test`. The next most important issue is that the new race fallback helpers are still not connected to `internal/relay/relay.go`, so the feature is not actually reachable from the live request path. The Windows smoke script also needs tightening because it can still pass by reusing an older binary even when the current sources no longer compile.

Checks run this time included repository inspection, git status/history comparison, README/docs/CI review, `go build ./...`, `go test ./... -count=1`, a fresh single-binary probe build, Windows backend smoke, frontend typecheck, backup verification scripts, static asset sync, `next build`, Vitest, and tool-availability checks for pnpm/corepack/docker. No business code was modified. New review artifacts were produced for this run.
