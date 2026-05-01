# 2026-04-22 Phase G Setting Help Browser Smoke CDP Fallback

## Scope

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G settings help browser-smoke closure
- Goal: add an Edge CDP fallback smoke entry so the settings cards can still be verified when Playwright self-start is blocked by host policy

## What Changed

1. Added `scripts/verify-setting-help-browser-smoke-cdp.mjs`.
2. Kept the scope on the same four setting cards:
   - `LLMPrice`
   - `DynamicRouting`
   - `CircuitBreaker`
   - `ModelProbe`
3. Kept the fallback checks aligned with the existing smoke intent:
   - admin login token bootstrap
   - localStorage injection for auth/nav/locale
   - desktop card presence
   - help-button focus and tooltip visibility
   - `375px` mobile width and overflow guard

## Verification

- `D:\gol1\node.exe --check scripts/verify-setting-help-browser-smoke.mjs`
- `D:\gol1\node.exe --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
- `D:\gol1\node.exe scripts/verify-setting-help-browser-smoke.mjs --check-only`
- `D:\gol1\node.exe scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only`
- `D:\gol1\node.exe scripts/verify-model-probe-help.mjs`

## Blockers

- `scripts/verify-setting-help-browser-smoke.mjs --self-start` is still blocked here by `node:child_process.spawn` returning `EPERM`.
- Playwright is not available locally, and `npx @playwright/cli` also hits npm-cache `EPERM` on this host.
- Long-lived background process launch through this shell remains flaky under host policy, so the real browser run still needs an external process path.

## Next Entry

1. Start backend in a separate shell:
   - `$env:OCTOPUS_ADMIN_USERNAME='admin'; $env:OCTOPUS_ADMIN_PASSWORD='admin'; & 'D:\GPT-codex\octopus_repo\build\octopus-smoke.exe' start`
2. Start Edge remote debugging in a separate shell:
   - `msedge.exe --remote-debugging-port=9222 --user-data-dir=%TEMP%\octopus-edge-cdp about:blank`
3. Run the real settings smoke:
   - `D:\gol1\node.exe scripts\verify-setting-help-browser-smoke-cdp.mjs`
