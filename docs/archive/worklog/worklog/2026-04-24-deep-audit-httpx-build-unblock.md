# 2026-04-24 Deep Audit HTTPX Build Unblock

- Master plan aligned before coding: yes
- Scope: continue admin write-route trust-boundary sweep, then fix the repo-level Go build blocker in `internal/utils/httpx`

## Local resources used

- `AGENTS.md`
- `$CODEX_HOME/automations/octopus/memory.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/worklog/2026-04-24-deep-audit-management-json-body-cap.md`
- `internal/server/handlers/*`, `internal/server/middleware/validate.go`, `internal/utils/httpx/body.go`

## Findings

1. Confirmed no new management-side `RequireJSON()` gaps in the currently registered `/api/v1/*` POST routes beyond the ones already closed in prior runs.
2. Found a concrete repo-level build blocker in `internal/utils/httpx/body.go`: `fmt.Errorf(tooLargeMessage)` trips Go 1.24's non-constant format-string check and prevents clean repository builds/tests once that package is compiled.

## Fixes

- changed `fmt.Errorf(tooLargeMessage)` to `fmt.Errorf("%s", tooLargeMessage)` in `internal/utils/httpx/body.go`
- added `internal/utils/httpx/body_test.go` to lock the helper's invalid-limit, default/custom oversize message, success path, and reader error propagation behavior

## Verification

- `gofmt -w internal/utils/httpx/body.go internal/utils/httpx/body_test.go`
- `. .\\scripts\\use-go-env.ps1; $env:GOCACHE=(Resolve-Path '.\\.tmp-gocache').Path; $env:GOMODCACHE=(Resolve-Path '.\\.tmp-gomodcache').Path; $env:GOPROXY='off'; & $env:GOEXE test ./internal/utils/httpx -count=1`

## Verification not completed

- `. .\\scripts\\use-go-env.ps1; ...; & $env:GOEXE build ./...`
- reason: this session's offline `GOPROXY=off` build could not resolve uncached third-party modules from the local module cache, so repository-wide build confirmation remains environment-blocked rather than code-blocked at the old `httpx` format-string failure

## Residual risks

- AI automation still appears to have a config-source/runtime-consumer disconnect and in-memory-only task runtime snapshots from prior audit findings
- dynamic-routing still lacks an HTTP-level relay-path proof after settings changes
- repo-wide Go green status still needs a network-available or fully primed module-cache rerun

## Next best follow-up

1. Re-run `go build ./...` with reachable module source or warmed cache to confirm the `httpx` blocker is fully cleared.
2. Deep-audit `manual / ai_profile` runtime consumption path in `internal/op/ai_automation*.go` and `web/src/components/modules/ai-automation/index.tsx`.
3. Add an HTTP-level dynamic-routing relay/log regression proving settings changes affect real `/v1/*` traffic as intended.
