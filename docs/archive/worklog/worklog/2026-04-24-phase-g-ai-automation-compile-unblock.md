# 2026-04-24 Phase G AI Automation Compile And Recommendation Contract Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G verification-chain unblock with Phase H AI automation contract closure`
- Summary: confirmed the old AI automation compile blocker no longer blocks package-level verification in the current workspace, then added minimal op/handler regression coverage for AI model selection and task config snapshot behavior. The new tests exposed a real contract drift in the default `free_success_latency` recommendation path, so the `free_likely` heuristic was tightened to stop treating `per_token` models with zero-filled prices as free-like candidates.
- Verification:
  - `./scripts/use-go-env.ps1; & $env:GOFMTEXE -w ./internal/op/ai_automation.go ./internal/op/ai_automation_models.go ./internal/op/ai_automation_test.go ./internal/server/handlers/ai_automation_test.go`
  - `$env:GOCACHE=<abs .codex-tmp/go-build>; $env:GOTMPDIR=<abs .codex-tmp/go-tmp>; ./scripts/use-go-env.ps1; & $env:GOEXE test ./internal/op`
  - `$env:GOCACHE=<abs .codex-tmp/go-build>; $env:GOTMPDIR=<abs .codex-tmp/go-tmp>; ./scripts/use-go-env.ps1; & $env:GOEXE test ./internal/server/handlers`
  - `$env:GOCACHE=<abs .codex-tmp/go-build>; $env:GOTMPDIR=<abs .codex-tmp/go-tmp>; ./scripts/use-go-env.ps1; & $env:GOEXE test ./internal/op ./internal/server/handlers ./internal/model ./internal/relay/...`
- Changed files:
  - `internal/op/ai_automation.go`
  - `internal/op/ai_automation_models.go`
  - `internal/op/ai_automation_test.go`
  - `internal/server/handlers/ai_automation_test.go`
  - `docs/worklog/2026-04-24-phase-g-ai-automation-compile-unblock.md`
- Result:
  - AI automation no longer blocks `internal/op` or `internal/server/handlers` verification in this workspace.
  - Remote model discovery now respects the documented `free_success_latency` policy more strictly.
  - Task-level AI config snapshots are covered so they can drive a task without mutating global AI automation settings.
  - Handler coverage now checks that `/api/v1/ai/models/fetch` returns the expected recommended free candidate.
- Risk:
  - Recommendation ordering for AI automation tasks is now slightly stricter and depends on `billing_mode=free` or truly unknown zero-cost metadata. This matches the current AI automation docs more closely and does not affect business routing.
- Next:
  1. Return to the Phase G screenshot-first browser-evidence pool and continue the remaining browser smoke closure work.
  2. If Phase H backend work continues next, expand AI automation handler/task contract coverage before moving deeper into frontend AI automation UI.
