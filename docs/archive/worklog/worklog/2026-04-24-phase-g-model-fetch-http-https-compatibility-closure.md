# 2026-04-24 Phase G Model Fetch HTTP HTTPS Compatibility Closure

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first pool / model fetch http-https compatibility closure`
- Summary: investigated the user-reported `model list fetch` failure and confirmed the hard `https-only` restriction was not in the `/api/v1/channel/fetch-model` handler itself, but in the providers preset validation contract. That validation rejected provider `base_url` values that used plain `http`, which could block local or LAN upstream endpoints from flowing into the channel create path and then into model fetching. The contract is now relaxed to accept absolute `http` or `https` URLs while still rejecting unsupported schemes.
- Verification:
  - `.\scripts\use-go-env.ps1; & $env:GOEXE test .\internal\server\handlers\providers.go .\internal\server\handlers\providers_test.go`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Validation attempts with blocker:
  - `.\scripts\use-go-env.ps1; & $env:GOEXE test .\internal\server\handlers\providers.go .\internal\server\handlers\providers_test.go .\internal\server\handlers\channel.go .\internal\server\handlers\channel_test.go .\internal\server\handlers\test_helper_test.go`
  - blocked by pre-existing compile errors in `internal/op/ai_automation.go` (`sort` unused, missing `aiAutomationFetchModels`, missing `AITaskStartAsync`)
- Changed files:
  - `internal/server/handlers/providers.go`
  - `internal/server/handlers/providers_test.go`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/ja.json`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-model-fetch-http-https-compatibility-closure.md`
- Result:
  - providers payload validation now accepts absolute `http` and `https` base URLs instead of forcing `https-only`.
  - handler tests explicitly cover `http` acceptance and still reject unsupported schemes such as `ftp`.
  - channel form help text now tells users that both `http` and `https` are supported for base URLs, reducing UI-level misdirection.
  - canonical plan, current status, workflow docs, and automation memory now explicitly record that model-fetch/providers paths must allow absolute `http` and `https` URLs.
- Risk:
  - This change intentionally broadens local/LAN compatibility for user-managed upstreams. It does not relax proxy URL validation beyond the existing `http/https/socks/socks5` rules, and it does not affect the separate settings-level API base URL validation, which already accepted both `http` and `https`.
  - Direct package-level `channel` handler regression remains blocked by unrelated AI automation compile failures, so this closure currently relies on validated providers tests plus code-path analysis.
- Next:
  1. if the user still sees model-list fetch failures, inspect the exact upstream error returned by `/api/v1/channel/fetch-model`, because the remaining blocker is more likely upstream reachability, auth, or `/models` compatibility rather than scheme validation.
  2. once the unrelated `internal/op/ai_automation.go` compile blockers are cleared, rerun the direct `channel` handler/package regression path and decide whether the prepared `fetch-model` HTTP test should stay as a maintained handler-level guard.
  3. keep the browser-level `channel create` evidence line as the next UI-pool task once this compatibility fix is merged into the current workspace flow.
