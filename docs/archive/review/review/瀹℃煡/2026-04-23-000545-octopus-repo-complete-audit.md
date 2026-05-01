# Octopus Repo Complete Audit

Triggered at: `2026-04-23 00:05:45 +08:00`
Workspace: `D:\GPT-codex\octopus_repo`
Branch: `feat/erguotou`
HEAD: `bfa27ae`
Comparison baseline: `origin/dev` (`0 behind / 22 ahead`)
Automation ID: `octopus-repo`

This pass re-read the local automation memory, the latest repo review reports, current `git` status, recent commits, branch delta, README/docs, validation workflow, and the main backend/frontend runtime paths before rerunning verification. The current uncommitted work is concentrated in `internal/op`, `internal/server`, `internal/relay`, backup/settings UI, channel/group editors, scripts, and docs, so those areas were prioritized.

## 1. Findings

### Critical

No confirmed `Critical` findings in this pass. The backend mainline is reachable, `go build ./...` and `go test ./... -count=1` are green, and the Windows backend smoke flow passes end-to-end.

### High

1. Advanced race-based recovery looks implemented but is still not connected to the live relay handler path.

   Evidence:
   - The group model and UI expose `failover_window_sec`, `race_after_fails`, and `race_concurrency` in `internal/model/group.go:21-23` and `web/src/components/modules/group/Editor.tsx:777-781,1021-1029`.
   - The live request path in `internal/relay/relay.go:73,86,142,192,216` now genuinely consumes `RetryRounds`, `RetryDelayMs`, `FailoverWindowSec`, same-channel multi-key fallback via `GetChannelKeyForModelExcept`, and sequential retry attempts.
   - But repository-wide search shows `effectiveDynamicRoutingTuning(...)`, `shouldEscalateToRace(...)`, and `runRaceFallback(...)` only exist in `internal/relay/dynamic.go`, `internal/relay/dynamic_runtime.go`, and tests. They are not called from `internal/relay/relay.go`.
   - That means route-target racing policy, race budgets, and concurrency-based recovery are present as helper/test code but inactive on the live handler path.

   Risk:
   - This is not a total failover regression anymore. Basic retry/fallback is real and tested.
   - The risk is a half-implementation mismatch: advanced race settings and related helper logic create the impression of shipped runtime behavior that the live request path still does not execute.

2. Frontend release/test verification is still blocked on this host by `spawn EPERM`, so local production-build and full Vitest proof are not available.

   Evidence:
   - `Push-Location web; node .\node_modules\next\dist\bin\next build` compiles successfully and completes TypeScript, then ends with `Error: spawn EPERM`.
   - `Push-Location web; node .\node_modules\vitest\vitest.mjs run` fails while loading `web/vitest.config.ts` because Vite/esbuild cannot spawn a subprocess.
   - `node --require .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run --config .\web\vitest.config.ts` still fails with the same `spawn EPERM`.
   - `scripts/vitest-no-spawn.cjs` only patches `child_process.exec` for the `net use` command and is not a working no-spawn fallback for esbuild/Vite startup.

   Risk:
   - The frontend source now type-checks, but this host still cannot produce a final successful `next build` or full Vitest run, so release verification is incomplete in the current environment.

### Medium

3. Dynamic-routing planning docs still overstate the shipped scope.

   Evidence:
   - `docs/DYNAMIC_ROUTING_REQUIREMENTS.md:218-222` and `docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:233-237` still describe `shadow-ai`, `hybrid`, `metrics-only`, and `incident-safe` modes.
   - `docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.md:212-224` and `docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md:227-239` still discuss enabling real hybrid behavior later.
   - The actual shipped path is the daily summary chain: `internal/task/dynamic.go:47` sets `daily_summary_scan_no_runtime_mutation`, `internal/task/dynamic.go:62` reports `dynamic routing health disabled; summary scan skipped`, `internal/server/handlers/stats.go:44,102` serves the summary, and `web/src/components/modules/setting/DynamicRouting.tsx:33,232-281` consumes it.
   - `README.md:178-180` and `README_zh.md:167-169` already describe the smaller, correct scope.

   Risk:
   - Current runtime behavior is narrower than parts of the planning docs imply, which can mislead completion scoring and release expectations.

4. Cross-provider protocol conversion is real on core paths, but several advertised edge cases remain TODO branches without direct proof in this pass.

   Evidence:
   - `internal/transformer/inbound/anthropic/messages.go:162` still says `TODO: support other result types`.
   - `internal/transformer/model/model.go:142,208,618,861` still contains TODO markers around request-model handling, schema, and image-generation request persistence.
   - `internal/transformer/outbound/gemini/messages.go:321` still leaves JSON-schema conversion as TODO.
   - I did not find direct tests targeting those TODO branches during this pass.

   Risk:
   - Main chat flows are fine, but richer `tool_result`, JSON schema, audio/prediction-style, and image-generation edge cases should still be considered partially implemented.

5. Backup/import/rollback is functional on the main path, but the feature is still only partially complete and says so in the UI.

   Evidence:
   - The feature now has real UI and verification for `Replace-prune preview`, `Post-import validation`, and `Rollback preview` in `web/src/components/modules/setting/Backup.tsx:915,925,993` and `web/src/components/modules/setting/Backup.test.tsx:337,455,494`.
   - The same page still renders `Advanced migration tooling still pending` in `web/src/components/modules/setting/Backup.tsx:1008`, and the test suite explicitly asserts that text in `web/src/components/modules/setting/Backup.test.tsx:506`.

   Risk:
   - This is a genuine feature, not an empty shell, but it should still be classified as partially complete rather than fully closed.

6. Release-facing README files are still not ready for external use.

   Evidence:
   - `README.md:45,51-54,71,73,88` still contains `<CURRENT_IMAGE>`, `<CURRENT_REPOSITORY_URL>`, `<CURRENT_ONE_CLICK_INSTALL_SCRIPT>`, `<CURRENT_RELEASES_URL>`, and TODO placeholders.
   - `README_zh.md:45,51-54,70,72,87` contains the same unresolved placeholders.

   Risk:
   - The codebase is much more mature than the top-level install/download docs suggest, but repository-level release readiness is still dragged down by placeholder documentation.

### Low

7. `web/package.json` still contains a non-portable local HTTPS dev path.

   Evidence:
   - `web/package.json:8` hardcodes `/workplace/code/octopus/build/localhost-key.pem` and `/workplace/code/octopus/build/localhost.pem` in the `devs` script.

   Risk:
   - Minor contributor friction and portability debt.
