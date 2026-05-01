# 2026-04-27 Bootstrap Admin Default Password And First Login Gate

- Timestamp: 2026-04-27T00:18:02.3118746+08:00
- Mainline: auth/bootstrap credential policy alignment
- Scope: first-login administrator password mechanism, forced password-change gate, docs sync

## Context

- The previous round switched the bootstrap admin fallback to a random password for security.
- Product direction for this round changed: first-start credentials should be predictable as `admin / admin`, but the first login must be forced through a password-change flow before the rest of the management console becomes usable.
- Existing repo context already included most of the required pieces:
  - backend `must_change_password` setting + `force-change-password` endpoint
  - frontend first-login gate screen in `web/src/components/app.tsx`
  - login response/status payloads carrying `must_change_password`

## What changed

- `internal/op/user.go`
  - keeps the fallback bootstrap credential as built-in `admin / admin` when `OCTOPUS_ADMIN_PASSWORD` is not provided
  - preserves force-password-change state for that built-in default path
  - keeps env-provided passwords from being re-exposed via the log hook
- `internal/server/middleware/auth.go`
  - adds a first-login gate: when `must_change_password` is true, authenticated management requests are blocked except for:
    - `GET /api/v1/user/status`
    - `POST /api/v1/user/force-change-password`
- `internal/server/middleware/auth_test.go`
  - adds coverage proving protected routes are blocked during the first-login gate
  - adds coverage proving the two password-change routes remain allowed
- `internal/server/handlers/user_test.go`
  - adds coverage proving `admin / admin` login returns `must_change_password=true`
  - adds coverage proving `force-change-password` clears the gate and allows optional username change
- `web/src/components/app.tsx`
  - aligns first-login copy with the built-in `admin / admin` bootstrap account
- `README.md` / `README_zh.md`
  - updates bootstrap credential documentation to match the implemented behavior

## Verification

- `gofmt -w internal/op/user.go internal/op/user_test.go internal/server/middleware/auth.go internal/server/middleware/auth_test.go internal/server/middleware/cors_test.go internal/server/handlers/user_test.go`
- `go test ./internal/op ./internal/server/middleware -count=1`
  - passed
- `go test ./internal/op ./internal/server/handlers ./internal/server/middleware -count=1`
  - partially blocked: `internal/server/handlers` setup attempted to fetch `github.com/tmaxmax/go-sse@v0.11.0` from `goproxy.cn`, which failed in the current environment DNS path
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - blocked by pre-existing parse errors in `web/src/components/modules/ai-automation/index.tsx`, unrelated to this bootstrap-password change

## Result

- Outcome: success with partial verification blockage outside the changed auth path
- Current behavior now matches the requested product flow: predictable bootstrap login, mandatory first-login password change, optional username rename
