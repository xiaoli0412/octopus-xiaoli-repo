# 2026-04-25 deep audit bootstrap password log contract

## Background

- This run rebuilt a repository-wide index and then re-checked management request boundaries, bootstrap auth flow, container entrypoint privilege drop, and update/download paths.
- The audit confirmed a concrete contract drift between README and code: when OCTOPUS_ADMIN_PASSWORD is unset, internal/op/user.go generates a random bootstrap admin password but the implementation no longer exposes that password once, which can lock a fresh instance out of the admin UI.

## Findings

1. esolveBootstrapAdminCredentials() generates a random password when no bootstrap password env var is set.
2. UserInit() immediately hashes and persists that generated password.
3. Without a one-time disclosure path, operators cannot recover the plaintext from storage and first-login access becomes unavailable.
4. The configured-password branch should stay redacted so env-provided secrets are not re-exposed in logs.

## Fix

- Added logBootstrapAdminCredentials and ootstrapPasswordForLog in internal/op/user.go.
- Auto-generated bootstrap passwords are now logged once for first-login recovery, matching the documented contract.
- Explicitly configured bootstrap passwords are redacted before reaching the log hook.

## Verification

- gofmt -w internal/op/user.go internal/op/user_test.go`n+- . .\\scripts\\use-go-env.ps1; =(Resolve-Path '.\\.tmp-gocache').Path; if (Test-Path '.\\.tmp-gomodcache') { =(Resolve-Path '.\\.tmp-gomodcache').Path }; ='off'; &  test ./internal/op -run 'TestUserInitUsesConfiguredBootstrapCredentials|TestUserInitGeneratesBootstrapPasswordWhenUnset|TestUserInitDoesNotExposeConfiguredBootstrapPasswordToLogHook' -count=1`n+
## Residual risk

- This run did not rerun repository-wide go test ./..., go build ./..., or frontend verification.
- internal/relay, web, and dynamic-routing runtime paths were only pattern-scanned this round and still need deeper review in later runs.
