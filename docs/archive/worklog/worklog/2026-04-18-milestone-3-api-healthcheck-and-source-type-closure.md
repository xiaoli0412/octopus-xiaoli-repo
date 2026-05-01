# 2026-04-18 Milestone 3 API Healthcheck And Source Type Closure

## 1. Task Info

- Task: milestone 3 API closure, healthcheck unification, and `source_type` cleanup
- Date: 2026-04-18
- Phase: Milestone 3 `Strategy And Performance`
- Milestone: 3

## 2. Inputs Before Start

- Canonical sections: `6.3`, `6.4`, `7.1.2`, `10.2.1`, milestone 3
- Related workflow sections: milestone 3 and phase 4 sections in the main MD
- Previous related worklog: [2026-04-18-milestone-3-runtime-budget-and-dynamic-health.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-18-milestone-3-runtime-budget-and-dynamic-health.md)
- Goals for this round:
  - verify and close the latest `source_type` changes
  - make `channel/group` handlers return proper `400/404` for config errors
  - unify `test-models-by-config` with the shared healthcheck pipeline
- Local resources used: `AGENTS.md`, the main MD, the previous milestone 3 worklog, existing `relay/op/model` tests, and thread handoff context
- Sub-agents enabled: yes
- Sub-agent model: `gpt-5.4`

## 3. Hard Rules

- Dynamic adjustment may tune thresholds only, never user priority order.
- `paid/metered` and `free/public` differences must affect both runtime behavior and test endpoints.
- `unknown` must default to conservative behavior.
- Auxiliary notes must stay under `docs/worklog/`.
- Sub-agents use `gpt-5.4`, and the main thread owns final integration.

## 4. Do Not Do

- Do not revert user changes.
- Do not drift away from milestone 3 mainline work.
- Do not treat unknown `source_type` values as aggressive-race eligible by default.

## 5. Acceptance Criteria

- `source_type` normalization, validation, and conservative default behavior are consistent.
- `channel/group` handlers no longer collapse business config errors into `500`.
- `test-models-by-config` uses the shared healthcheck logic.
- Related Go tests pass.

## 6. Rollback Points

- `internal/model/channel.go`
- `internal/model/backup.go`
- `internal/op/healthcheck.go`
- `internal/relay/dynamic.go`
- `internal/relay/relay.go`
- `internal/server/handlers/channel.go`
- `internal/server/handlers/group.go`

## 7. Scope

- Backend or UI first: backend first
- Backend modules touched:
  - `internal/model`
  - `internal/op`
  - `internal/relay`
  - `internal/server/handlers`
- Frontend modules touched in main thread: `web/src/api/endpoints/channel.ts`, `web/src/components/modules/channel/Form.tsx`, `web/src/components/modules/channel/CardContent.tsx`
- APIs affected:
  - `/api/v1/channel/create`
  - `/api/v1/channel/update`
  - `/api/v1/channel/enable`
  - `/api/v1/channel/delete/:id`
  - `/api/v1/channel/fetch-model`
  - `/api/v1/channel/test-models`
  - `/api/v1/channel/test-models-by-config`
  - `/api/v1/group/create`
  - `/api/v1/group/update`
  - `/api/v1/group/delete/:id`
- Old data impact: yes, old `source_type` aliases and dirty values are interpreted more conservatively
- Old behavior impact: yes, `unknown` no longer defaults to aggressive racing

## 8. Steps

1. Fix and tighten `source_type` normalization and validity checks.
2. Align milestone 3 default racing behavior so `unknown` and `private/internal` stay conservative.
3. Add policy metadata to healthcheck results.
4. Route `test-models-by-config` through `CheckChannelModelHealthForConfig()`.
5. Fix `channel/group` handler error mapping and `match_regex` pre-validation.
6. Run `gofmt` and package tests.
7. Add the daily dynamic-routing adjustment task skeleton required by milestone 3.

## 9. Testing And Verification

- Formatting:
  - `gofmt -w internal/model/channel.go internal/model/channel_test.go internal/model/backup.go internal/op/healthcheck.go internal/relay/dynamic.go internal/relay/relay.go internal/relay/relay_test.go internal/server/handlers/channel.go internal/server/handlers/group.go`
- Tests:
  - `go test ./internal/model/...`
  - `go test ./internal/op/...`
  - `go test ./internal/relay/...`
  - `go test ./internal/server/...`
  - `go test ./internal/task/...`
  - `pnpm exec tsc --noEmit` in `web/`
- Result:
  - after this quality sweep, `go test ./internal/model/... ./internal/op/... ./internal/server/handlers ./internal/relay/... ./internal/task/...` passed
  - after this quality sweep, `pnpm exec tsc --noEmit` in `web/` passed

## 10. Risks And Compatibility

- New risk: callers that implicitly relied on `unknown => race allowed` now become more conservative.
- Compatibility risk: frontend config-test UI is now aligned on request shape and canonical `source_type` input, but browser-level smoke is still pending.
- Blocks next task: no

## 11. Finish Notes

- Build status: full project build not run in this worklog
- Test status: passed for the targeted Go packages
- Main completed changes:
  - corrected the runtime distinction between `classified` and `pooled` key selection so pooled channels use a shared key pool for the declared/shared model set
  - added model-layer regression tests that explicitly prove `classified` and `pooled` produce different eligible key sets
  - aligned healthcheck proxy behavior with relay semantics, including system proxy fallback when `proxy=true` and `channel_proxy` is empty
  - applied `channel.custom_header` to healthcheck requests to avoid relay-vs-healthcheck false negatives
  - normalized group-name handling at handler entry so blank names fail fast and trimmed names persist canonically
  - rebuilt broken handler test scaffolding and restored runnable `channel/group` handler regression tests
  - tightened frontend channel form input by switching `source_type` to canonical controlled values and fixing the multi-`base_url` fetch-model button gate
  - surfaced light policy metadata (`source_type` / `billing_mode` / `probe_policy` / `policy_basis`) in the channel config-test summary
  - introduced canonical `source_type` constants:
    - `unknown`
    - `public/free`
    - `paid/metered`
    - `private/internal`
  - added effective fallback behavior for dirty `source_type` values
  - made `unknown` and `private/internal` conservative by default for race behavior
  - added `source_type`, `billing_mode`, `probe_policy`, and `policy_basis` to healthcheck results
  - unified config-based model testing with the shared healthcheck path
  - fixed handler-side `400/404/500` error classification for `channel/group`
  - added `channel.match_regex` pre-validation in handlers
  - narrowed `fetch-model` to use `ChannelFetchModelRequest`
  - aligned frontend `fetch-model` request shape with the narrowed backend contract to avoid a real form regression
  - added minimum handler-level regression tests for milestone 3 API closure
  - added a daily dynamic-routing adjustment background task skeleton and wired it into the task scheduler
  - added task-level tests for the daily adjustment summary and disabled-state skip behavior
- Local resources used and what they contributed:
  - `AGENTS.md`: local-resource priority and sub-agent usage rules
  - main MD: milestone 3 hard rules for dynamic adjustment, paid/free behavior, and budgets
  - previous worklog: confirmed runtime budget and dynamic health were already landed
  - handoff summary: identified unverified `source_type` changes and handler/API closure as the next mainline work
- Sub-agents used and conclusions:
  - `019d9daf-f828-73c3-a571-8bbb4efbb984`, `gpt-5.4`
    - read-only analysis for handler/API closure gaps
    - conclusion adopted: fix error mapping, regex validation, and config-test endpoint visibility
  - `019d9daf-f8c9-7863-8dd5-bac0d3a80938`, `gpt-5.4`
    - read-only analysis for cost-aware probe and test-endpoint gaps
    - conclusion adopted: unify `test-models-by-config` with healthcheck and expose policy metadata
  - `019d9e3c-3cfd-7312-8861-f7c095f50d61`, `gpt-5.4`
    - assigned frontend request-body alignment for config-test calls
    - main thread did not wait for that result before finishing backend closure
- Manual smoke status: not fully run
- Manual smoke blockers: full frontend-backend interactive environment was not part of this pass
- Pages still needing verification:
  - channel form config test
  - channel card single-key test
- Worklog updated: yes
- Encoding note: the first version of this worklog was corrupted by a bad text-encoding write path and was rewritten into a stable readable version in the same file.
- Additional follow-up completed after the quality sweep:
  - normalized `LLMInfo` policy fields (`billing_mode`, `probe_policy`, `probe_interval_seconds`, `probe_concurrency_limit`) on create/update and added regression tests for normalization and invalid-value rejection
  - converted model create/edit UI from free-text policy inputs to controlled canonical options for `billing_mode` and `probe_policy`
  - established a stable model-default policy baseline so route-target work can build on `model default > channel/key inheritance` instead of raw free-form values

- Remaining items:
  - full browser smoke for channel form, channel card single-key test, grouped model picker, and model create/edit overlays is still pending
  - explicit route-target-level persistent overrides for `(channel, key, model)` are still not modeled as first-class CRUD data; current progress has stabilized the model-default layer but not the full override layer
  - the daily dynamic-adjustment task exists as a skeleton summary scan, but not as a full persisted tuning pipeline yet
  - probe cost is still recent-window observation, not full persistent accounting
- Next task prerequisites met: yes

## 12. Follow-Up On Backup / Import / Route-Target Preview

- Follow-up date: 2026-04-18
- Follow-up scope: close the `route-target explicit override > model default > channel/key inheritance` preview chain for backup/import/rollback preview, without drifting away from the main MD.

### What Was Closed In This Follow-Up

- `DBImportRoutePreviewCandidate` now carries first-class route-target policy preview fields for:
  - `probe_interval_seconds`
  - `probe_concurrency_limit`
  - `billing_mode_basis`
  - `probe_policy_basis`
  - `probe_interval_basis`
  - `probe_concurrency_basis`
- import dry-run route preview no longer reuses current-channel keys blindly when a snapshot channel already exists by name. It now keeps snapshot key intent and maps key IDs by credential where possible.
- import dry-run route preview now applies snapshot `route_target_overrides` onto the imported-preview candidate chain, so preview reflects the effective post-import route-target policy instead of only the current runtime cache state.
- current-preview now also resolves route-target policy using the loaded import-state view, so model-default policy and existing persisted overrides show up consistently in before-candidates.
- route preview diff signatures now include route-target policy fields, so a pure policy change on the same `(channel, key, model)` candidate still produces a visible diff.
- frontend backup UI now aligns its route-preview candidate type with the backend field names and shows:
  - interval
  - concurrency
  - finer per-field policy basis details
- rollback preview no longer relies on a separate fallback-only compact route diff row; full formatted route diff text is surfaced through rollback preview warnings.

### Validation Run In This Follow-Up

- Go tests:
  - `go test ./internal/op -run TestDBImportIncrementalDryRunRoutePreviewDiffDetectsRouteTargetOverridePolicyChanges`
  - `go test ./internal/op/...`
  - `go test ./internal/op/... ./internal/relay/... ./internal/db/... ./internal/model/...`
- Frontend lint:
  - `pnpm exec eslint src/components/modules/setting/Backup.tsx src/api/endpoints/setting.ts`
- Result:
  - all targeted Go tests passed
  - backend package sweep for `internal/op / relay / db / model` passed
  - targeted frontend eslint passed after removing the last unused helper warning

### Sub-Agents Used In This Follow-Up

- `019d9efe-f38d-7793-9495-a72ce96c1482`, `gpt-5.4`
  - owned frontend-only changes in `web/src/api/endpoints/setting.ts` and `web/src/components/modules/setting/Backup.tsx`
  - conclusion adopted with main-thread field-name alignment and rollback-preview cleanup
- `019d9f0a-a53c-7661-b538-3b1d2480bdeb`, `gpt-5.4`
  - assigned backend-only route-target preview closure in `internal/model/backup.go`, `internal/op/backup_extra.go`, `internal/op/backup_test.go`
  - main thread completed and integrated the backend closure directly after local verification remained on the critical path

### What Still Remains After This Follow-Up

- route-target override persistence and runtime precedence are now visible in backup/import preview, but full first-class CRUD/API/UI management for `(channel, key, model)` overrides is still not complete.
- rollback preview is now richer, but it is still rendered as formatted text rather than a dedicated structured compare panel.
- full browser smoke on Linux/server runtime and a Windows local run is still pending.

## 13. Follow-Up On Route-Target Override CRUD/API/UI

- Follow-up date: 2026-04-18
- Follow-up scope: move `RouteTargetOverride` from runtime-only / preview-only capability into a minimal first-class managed object with backend API closure and a first frontend entry point.

### What Was Closed In This Follow-Up

- backend route-target override management API was added with repository-consistent management paths:
  - `GET /api/v1/route-target/list`
  - `POST /api/v1/route-target/upsert`
  - `POST /api/v1/route-target/delete`
- request models were made explicit in `internal/model/route_target.go` for:
  - `RouteTargetOverrideUpsertRequest`
  - `RouteTargetOverrideDeleteRequest`
- op-layer delete support was completed with exact triple-key deletion on `(channel_id, channel_key_id, model_name)`.
- handler-level regression tests were added for:
  - successful upsert
  - successful list
  - `404` on deleting a missing override
- frontend now has a first management entry point inside `ChannelForm` advanced settings:
  - summary block in advanced routing settings
  - `Manage overrides` dialog
  - minimal `list + upsert + delete` workflow for persisted route-target overrides on an existing channel
- frontend channel API client now exposes:
  - `useRouteTargetOverrideList()`
  - `useUpsertRouteTargetOverride()`
  - `useDeleteRouteTargetOverride()`

### Validation Run In This Follow-Up

- Go tests:
  - `go test ./internal/server/handlers -run RouteTarget`
  - `go test ./internal/op -run RouteTarget`
  - `go test ./internal/server/handlers ./internal/op/... ./internal/relay/... ./internal/model/...`
- Frontend checks:
  - `pnpm exec tsc --noEmit`
  - `pnpm exec eslint src/api/endpoints/channel.ts src/components/modules/channel/Form.tsx`
- Result:
  - targeted and package-level Go tests passed
  - frontend TypeScript build check passed
  - frontend eslint reported only pre-existing warnings in `ChannelForm`, with no blocking errors

### Sub-Agents Used In This Follow-Up

- `019d9f29-fc16-7731-9bd5-3592af65d399`, `gpt-5.4`
  - read-only backend API design review
  - conclusion adopted: keep repository-style `/list + /upsert + /delete` management endpoints
- `019d9f29-fba4-7240-b039-d9fd497d1323`, `gpt-5.4`
  - read-only frontend placement review
  - conclusion adopted: first UI entry point should live in `ChannelForm` advanced settings instead of a new standalone page

### What Still Remains After This Follow-Up

- the new frontend management UI is intentionally minimal and dialog-based; it is not yet a full structured route-target policy panel.
- route-target override management is only surfaced in channel edit flow, not yet in channel view routing summary or a dedicated route-target management surface.
- full browser smoke on Linux/server runtime and a Windows local run is still pending.

## 14. Follow-Up On Route-Target View Visibility And Channel-Scoped Reads

- Follow-up date: 2026-04-18
- Follow-up scope: make route-target override visibility available in channel view mode, and stop relying on frontend-side filtering over a global override list when the current screen only needs one channel.

### What Was Closed In This Follow-Up

- backend route-target list API now supports `channel_id` query filtering on:
  - `GET /api/v1/route-target/list?channel_id=<id>`
- op-layer support was added for channel-scoped reads through `RouteTargetOverrideListByChannel()`.
- handler-level regression coverage was extended to confirm channel-scoped list filtering returns only the requested channel rows.
- frontend channel API hook now supports channel-scoped route-target reads by parameterizing `useRouteTargetOverrideList(channelId?)`.
- channel edit form no longer pulls the full override list and filters client-side when editing an existing channel; it now requests only the current channel’s overrides.
- channel view mode now shows route-target override visibility directly inside the existing routing section of `CardContent`, including:
  - override count
  - first few override rows
  - model name
  - key id
  - billing mode
  - probe policy
  - interval / concurrency summary

### Validation Run In This Follow-Up

- Go tests:
  - `go test ./internal/server/handlers -run RouteTarget`
  - `go test ./internal/op -run RouteTarget`
  - `go test ./internal/server/handlers ./internal/op/... ./internal/relay/... ./internal/model/...`
- Frontend checks:
  - `pnpm exec tsc --noEmit`
  - `pnpm exec eslint src/api/endpoints/channel.ts src/components/modules/channel/Form.tsx src/components/modules/channel/CardContent.tsx`
- Result:
  - targeted and package-level Go tests passed
  - frontend TypeScript build check passed
  - eslint only reported pre-existing warnings in `ChannelForm` and one unused import warning in `CardContent`, with no blocking errors

### Sub-Agents Used In This Follow-Up

- `019d9f36-efd2-77d1-abad-8ea6284754ca`, `gpt-5.4`
  - read-only review for the best view-mode placement of route-target override visibility
  - conclusion adopted: keep view-mode visibility inside the existing routing section of `CardContent`, and prefer channel-scoped list reads instead of global frontend filtering

### What Still Remains After This Follow-Up

- route-target view mode currently exposes a compact textual summary, not a dedicated structured policy compare panel.
- key labels in the view-mode summary still use the persisted key id directly; a richer label based on key remark / masked key can still improve readability.
- full browser smoke on Linux/server runtime and a Windows local run is still pending.

## 15. Follow-Up On Route-Target Readability And Run-Prep Findings

- Follow-up date: 2026-04-18
- Follow-up scope: finish the key-label readability pass for route-target visibility, and document the immediate Linux/Windows run-prep conclusions without drifting away from code mainline.

### What Was Closed In This Follow-Up

- route-target key labels are now unified across channel module surfaces with one shared frontend helper in `web/src/components/modules/channel/key-label.ts`.
- the shared label rule is now explicit and consistent:
  - `remark` first
  - masked key second
  - `#id` fallback
- channel view-mode route-target summary now uses the same readable key labels as edit-mode.
- channel edit-mode route-target summary and existing-override list now also use the shared key-label helper instead of raw `channel_key_id`.

### Validation Run In This Follow-Up

- Frontend checks:
  - `pnpm exec tsc --noEmit`
  - `pnpm exec eslint src/components/modules/channel/key-label.ts src/components/modules/channel/Form.tsx src/components/modules/channel/CardContent.tsx`
- Backend checks:
  - `go test ./internal/server/handlers ./internal/op/... ./internal/relay/... ./internal/model/...`
- Result:
  - frontend TypeScript build check passed
  - backend package checks passed
  - eslint still reports pre-existing warnings in `ChannelForm`, but no blocking errors

### Sub-Agents Used In This Follow-Up

- `019d9f3e-26ff-7880-9319-f0cbf8402d38`, `gpt-5.4`
  - read-only review of existing key-label display patterns in frontend
  - conclusion adopted: standardize on `remark > masked key > #id` and factor that rule into a shared helper
- `019d9f3e-2779-7a81-951c-4e99375abf02`, `gpt-5.4`
  - read-only review of Linux server run prerequisites and Windows local shortest validation path
  - conclusion recorded for next mainline step: Linux runtime is blocked more by missing execution/acceptance closure than by a proven startup-code bug; Windows already has a short path through `scripts/dev-win.ps1`

### What Still Remains After This Follow-Up

- route-target visibility is now readable, but still compact and text-first; a structured policy panel remains the next UX step.
- Linux server runtime verification still needs a real execution pass and an explicit acceptance checklist.
- Windows local runtime verification still needs an actual end-to-end run, even though the shortest script path is already identified.

## 16. Follow-Up On Runtime Reliability, Full Backup User Restore, And Group Stale Cleanup

- Follow-up date: 2026-04-18
- Follow-up scope: close the immediate reliability and usability gaps that were blocking the main MD from real runnable quality, without broad rewrites.

### What Was Closed In This Follow-Up

- startup config loading was hardened in `cmd/start.go`:
  - `start` now uses `PreRunE`
  - configuration load failures now abort startup instead of being ignored
  - duplicate `UserInit()` execution was removed because `InitCache()` already owns that path
- server startup reliability was tightened in `internal/server/server.go`:
  - `/healthz` now exists as a stable HTTP health endpoint
  - `server.Start()` now returns early when `ListenAndServe()` fails immediately
  - static asset serving now prefers a verified local directory and otherwise falls back to embedded assets cleanly
- Docker and local run preparation were aligned toward real runtime use:
  - `docker-compose.yml` now uses the CLI healthcheck against `http://127.0.0.1:8080/healthz`
  - `scripts/dev-linux.sh` now validates Go / Node.js / pnpm minimum versions instead of only checking command presence
  - Linux script Go requirement was aligned to `1.24.4+`, matching the current Windows helper scripts
- backup/import model integrity was repaired after a broken patch path corrupted `internal/model/backup.go`:
  - the advanced backup schema was restored as valid Go code
  - `DBDump.Users` is now part of the full snapshot model
- full-backup / full-import user restore usability was closed further:
  - `DBExportAll()` now exports users along with routing/config data
  - import/rollback paths keep users in the replace snapshot flow
  - post-import validation now calls full `InitCache()` so user/settings/channel/group/apiKey/llm/route-target/stats caches are refreshed together
  - `UserInit()` now reloads the first persisted user explicitly instead of accidentally reusing a stale in-memory primary key
- the stale group-item cleanup line was preserved and validated instead of weakened:
  - channel updates still reject new group bindings when the channel/model has no configured usable key
  - when channel keys or allowed-model coverage shrink, stale `group_items` and related `route_target_overrides` are removed
  - relay regression tests that intentionally simulate historical dirty rows were updated to inject those rows directly at the DB layer, so runtime stale-row tolerance remains tested without reopening invalid management writes

### Validation Run In This Follow-Up

- Formatting:
  - `gofmt -w cmd/start.go cmd/healthcheck.go internal/server/server.go internal/model/channel.go internal/model/backup.go internal/op/channel.go internal/op/channel_test.go internal/op/group.go internal/op/group_test.go internal/op/backup.go internal/op/backup_extra.go internal/op/backup_test.go internal/op/cache.go internal/op/op_test.go internal/op/user.go internal/relay/relay_more_test.go`
- Targeted Go tests:
  - `go test ./internal/op -run TestDBExportAllIncludesSecretsByDefault -count=1`
  - `go test ./internal/op -run TestDBRollbackLatestImportSnapshotRestoresPreviousState -count=1`
  - `go test ./internal/op -run TestChannelUpdateRemovesStaleGroupItemsWhenModelsShrink -count=1`
  - `go test ./internal/op -run TestLLMDeleteRemovesReferencedGroupItems -count=1`
  - `go test ./internal/op -run TestLLMBatchDeleteRemovesReferencedGroupItems -count=1`
  - `go test ./internal/op -run TestGroupItemBatchAddDeduplicatesAndAppendsPriority -count=1`
  - `go test ./internal/op -run TestGroupUpdateRefreshesNameMapAndItems -count=1`
  - `go test ./internal/relay -run TestHandlerSkipsStaleGroupItemWhenChannelDoesNotDeclareModel -count=1`
  - `go test ./internal/relay -run TestRunRaceFallbackRecordsUnavailableCandidates -count=1`
- Package and repository sweeps:
  - `go test ./cmd ./internal/server/... ./internal/op/... ./internal/model/... -count=1`
  - `go test ./... -count=1`
- Frontend check:
  - `pnpm exec tsc --noEmit`
- Result:
  - full Go repository test sweep passed
  - frontend TypeScript build check passed

### Sub-Agents Used In This Follow-Up

- `019d9f70-9d4f-7de1-9271-7c0ea11e57a7`, `gpt-5.4`
  - read-only milestone/MD status audit
  - conclusion adopted: route-target CRUD/preview closure is already beyond the old milestone text, while Linux/Windows runtime acceptance and some structured UX still remain open
- `019d9f70-9dc7-75c2-a23f-631de560b4ad`, `gpt-5.4`
  - read-only backup/import audit
  - conclusion adopted: the real remaining usability gap was not only user rows in the dump, but also cache rewarming after import/rollback, especially `userCache`

### What Still Remains After This Follow-Up

- Linux server runtime still needs a real manual execution pass and acceptance checklist, even though startup/healthcheck/script preparation is now much stronger.
- Windows local runtime still needs an actual end-to-end run, not just script preparation.
- Docker runtime still needs a real container smoke run after these healthcheck/startup fixes.
- route-target management and visibility are functional but still text-first; a richer structured policy panel remains a later UI task.

## 17. Follow-Up On Windows Local Runtime Smoke And Config BOM Compatibility

- Follow-up date: 2026-04-18
- Follow-up scope: convert the remaining Windows local runtime status from "prepared" into "actually smoke-tested", while fixing a real Windows config-file compatibility issue discovered during the run.

### What Was Closed In This Follow-Up

- `internal/conf/config.go` now supports UTF-8 BOM config files when `viper.ReadInConfig()` rejects them.
  - this fixes a real Windows-local failure mode where PowerShell-written JSON config files could fail to parse before the server even started
  - the fallback only activates when the file already failed normal parsing and a BOM prefix is actually present
- a regression test was added in `internal/conf/config_test.go` to prove BOM-prefixed config files load correctly and populate `AppConfig`
- `scripts/dev-linux.sh` version comparison was hardened to use numeric component comparison instead of relying on `sort -V`
  - this avoids false failures in mixed Windows/MSYS/Git Bash environments where `sort -V` behavior is not reliable enough for the current script path
- Windows local backend runtime was smoke-tested with a real local process and a mock upstream instead of stopping at package tests:
  - built a temporary local `octopus-smoke.exe`
  - started the backend with a temporary SQLite DB and temporary JSON config
  - confirmed `/healthz` returned `{"status":"ok"}`
  - confirmed admin login succeeded through `/api/v1/user/login`
  - created a minimal channel through `/api/v1/channel/create`
  - created a minimal matching group through `/api/v1/group/create`
  - created a gateway API key through `/api/v1/apikey/create`
  - completed a real `/v1/chat/completions` request against a local mock upstream and received a successful chat-completion response
- a reusable Windows smoke script was added as `scripts/smoke-win-backend.ps1` so the same local runtime path can be repeated without manually reassembling the flow

### Validation Run In This Follow-Up

- Go checks:
  - `gofmt -w internal/conf/config.go internal/conf/config_test.go`
  - `go test ./internal/conf -count=1`
  - `go test ./internal/conf ./cmd ./internal/server/... ./internal/op/... ./internal/model/... -count=1`
- Linux prep check:
  - `bash scripts/dev-linux.sh --backend-only --check-only`
- Windows local smoke:
  - built `build/octopus-smoke.exe`
  - launched backend with a temporary BOM-prefixed JSON config and temporary SQLite DB
  - drove `healthz -> user/login -> channel/create -> group/create -> apikey/create -> /v1/chat/completions`
  - used a local Python mock upstream server to avoid external network and real-provider dependency
- Result:
  - BOM-config regression test passed
  - package checks passed for the touched backend areas
  - Linux script preflight now passes in the current mixed environment
  - Windows local backend smoke passed end-to-end

### Sub-Agents Used In This Follow-Up

- `019d9fec-548e-7b60-ac80-c2c51212be0c`, `gpt-5.4`
  - read-only analysis of the shortest real HTTP smoke path
  - conclusion adopted: the minimal closure is `user/login -> channel/create -> group/create -> apikey/create -> /v1/chat/completions`, with `group.name == request.model` as the key pitfall
- `019d9fec-5362-7781-9c33-15839c3e6c91`, `gpt-5.4`
  - read-only Docker text-chain review
  - conclusion adopted: current compose/runtime path is close, but local source-to-image closure is still not self-contained because the Dockerfiles package prebuilt binaries rather than building them from source

### What Still Remains After This Follow-Up

- Linux server runtime still needs a real server-side execution pass, not just preflight and code-level readiness.
- Docker runtime still needs an actual smoke run in an environment with Docker installed.
- current Dockerfiles are still packaging prebuilt binaries rather than performing a self-contained source build; compose therefore validates the published image path more than the local source tree.
- browser-level smoke on the pending frontend surfaces is still outstanding.

## 18. Follow-Up On Linux Binary Delivery And Source-Build Docker Closure

- Follow-up date: 2026-04-18
- Follow-up scope: improve the Linux/Docker runtime mainline without drifting into a full deployment rewrite, by closing the highest-value gaps discovered during the latest runtime audit.

### What Was Closed In This Follow-Up

- `scripts/build.sh` now exposes a lightweight Linux delivery path:
  - added `linux-binary`
  - this path builds only `build/bin/octopus-linux-x86_64`
  - it no longer forces the frontend build / price-update pipeline when the real need is only a Linux server binary
- `scripts/build.sh` now also exposes a local source-image entry point:
  - added `docker-image`
  - this path is intended to build `octopus:local` directly from the repository source tree when Docker is available
- `scripts/build.sh` now has a split environment-prep model:
  - full `prepare_environment()` still serves the heavier release/build flows
  - new `prepare_lightweight_environment()` keeps `linux-binary` on a Go-only dependency path
- Docker source-build closure was materially improved:
  - `scripts/dockerfiles/Dockerfile.debian` was converted into a multi-stage build
  - frontend export is now built inside the image build path
  - the Go binary is now built inside the image build path
  - the final runtime layer still uses the existing `entrypoint.sh` and runtime contract, so the change stays low-drift relative to the current repo structure
- local Docker build context hygiene was improved with a new `.dockerignore`
  - this avoids sending `web/node_modules`, local build outputs, and other noisy directories into the Docker build context
- repository-facing runtime docs were updated in `README.md`
  - compose instructions now reflect that the repo compose path builds from local source
  - the new `linux-binary` path is documented as the shortest Linux server binary delivery route

### Validation Run In This Follow-Up

- Go checks:
  - `go test ./cmd ./internal/server/... ./internal/op/... ./internal/model/... ./internal/conf -count=1`
- Script checks:
  - `bash scripts/build.sh help`
  - `bash scripts/build.sh linux-binary`
- Result:
  - backend package checks passed for the touched areas
  - `linux-binary` completed successfully in the current environment and produced `build/bin/octopus-linux-x86_64`
  - direct Docker smoke is still blocked on the current machine because Docker is not installed, but the repo-side source-build path is now much closer to self-contained

### Sub-Agents Used In This Follow-Up

- `019da02d-7779-7971-b003-a08a3e3faf9a`, `gpt-5.4`
  - read-only Linux/Docker delivery-chain audit
  - conclusion adopted: the biggest remaining gap was not Linux binary compilation itself, but the lack of a self-contained source-to-image build path; the recommended smallest fix was a multi-stage Dockerfile plus clearer build entry points

### What Still Remains After This Follow-Up

- Linux server runtime still needs a real server-side execution pass and acceptance checklist on an actual Linux environment.
- Docker runtime still needs an actual `docker build` / `docker compose up` smoke run on a machine with Docker installed.
- `README_zh.md` still needs a clean encoding-safe sync for the new Linux/Docker wording.
- browser-level smoke on the pending frontend surfaces is still outstanding.

- Quality-sweep sub-agents used and conclusions:
  - `019d9e5f-ac51-7211-a196-ba0ee0f5d0df`, `gpt-5.4`
    - audited `classified` vs `pooled` runtime semantics
    - conclusion adopted: split runtime eligibility at `internal/model/channel.go` and replace pooled tests that were asserting the wrong behavior
  - `019d9e5f-acdf-7ee3-9a3c-dd9872043aab`, `gpt-5.4`
    - audited healthcheck vs relay proxy/header consistency
    - conclusion adopted: align system-proxy fallback semantics and apply `custom_header` to healthcheck requests
  - `019d9e5f-ae1f-79b3-8946-03f7c8810eaf`, `gpt-5.4`
    - audited group-name validation, `source_type` UI/handler consistency, and fetch-model UI gate
    - conclusion adopted: canonicalize group-name handling, constrain `source_type` UI input, and fix the multi-`base_url` model-fetch gate
