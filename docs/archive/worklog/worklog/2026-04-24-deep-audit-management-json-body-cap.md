# 2026-04-24 Deep Audit Management JSON Body Cap

- Master plan aligned before coding: yes
- Scope: management-side JSON request body limits on `/api/v1/*`
- Findings: `RequireJSON()` only enforced content type, so many admin POST JSON handlers decoded unbounded bodies; some `ShouldBindJSON` routes also missed `RequireJSON()` entirely.
- Fixes:
  - added a 2 MiB management-only JSON body cap inside `internal/server/middleware/validate.go`
  - added middleware regression tests in `internal/server/middleware/validate_test.go`
  - attached `RequireJSON()` to `channel/test-models`, `channel/test-models-by-config`, `setting/rollback-import-snapshot`, and `setting/preview-rollback-import-snapshot`
- Verification:
  - `gofmt -w internal/server/middleware/validate.go internal/server/middleware/validate_test.go`
  - `gofmt -w internal/server/handlers/channel.go internal/server/handlers/setting.go`
  - `go test ./internal/server/middleware ./internal/server/handlers -count=1`
- Verification note: tests passed only after switching to offline module cache plus repo-local writable build cache because the default host Go cache/env path was noisy.
- Residual risks:
  - remaining admin write routes still need a pass to confirm whether each body-consuming POST is behind `RequireJSON()` or has its own bounded parser
  - dynamic-routing HTTP-level relay-log audit proof is still missing

---

## 2026-04-24 Follow-up: AI automation remote model discovery cap

- Master plan aligned before coding: yes
- Scope: `internal/op` remote AI model discovery response handling
- Finding: `aiAutomationFetchModelsRemote` accepted third-party `/models` responses without a hard size cap on the success path, so an oversized upstream payload could be buffered and parsed during admin-side model discovery.
- Fixes:
  - added `maxAIAutomationModelDiscoveryResponseBytes` and `decodeRemoteAIAutomationPayload(...)` in `internal/op/ai_automation_models.go`
  - switched remote OpenAI/Anthropic/Gemini parsing to reuse the capped payload reader before JSON unmarshal
  - added regression coverage in `internal/op/ai_automation_test.go` for oversized remote discovery responses
- Verification:
  - `gofmt -w internal/op/ai_automation_models.go internal/op/ai_automation_test.go`
  - `go test ./internal/op -run 'TestAIAutomationFetchModelsPrefersRemoteFreeCandidate|TestDecodeRemoteAIAutomationModelsRejectsOversizedResponse' -count=1`
- Verification note: the host default Go cache path was permission-blocked and the first run timed out while downloading modules; the passing run used repo-local `GOCACHE` plus the existing user module cache with offline `GOPROXY`.
- Residual risks:
  - remaining admin write routes still need a final pass for body-cap consistency
  - dynamic-routing HTTP-level relay-log audit proof is still missing
  - `go env -w` could not persist because the host Go env file is access-restricted in this session
