# 2026-04-24 Phase G Channel Create CDP Browser Smoke Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / channel create browser evidence`
- Summary: replaced the blocked Node-internal Playwright CLI path with a host-level Edge CDP path for `channel create`, added a dedicated CDP smoke script and PowerShell wrapper, and confirmed the current host can pass the browser smoke when the page bootstrap strategy is pinned to `attached-session`.
- Verification:
  - `node scripts/verify-channel-create-flow.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node --check scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `node scripts/verify-channel-create-browser-smoke-cdp.mjs --check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke-cdp.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke-cdp.ps1 -Mode self-start -Driver cdp -CdpPageBootstrapStrategy attached-session -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke-cdp.ps1 -Mode self-start -Driver cdp -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
- Changed files:
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - `docs/worklog/2026-04-24-phase-g-channel-create-cdp-browser-smoke-closure.md`
- Result:
  - `channel create` browser smoke no longer depends on Node spawning `@playwright/cli`, so the previous `spawn EPERM` blocker is no longer on the critical path.
  - The new host-level CDP wrapper can self-start backend, frontend, and Edge, then drive the existing `channel create` selector contract through open-dialog, first-key-expand, help-hint focus, and `375px` layout assertions.
  - On this host, `auto` still falls into a failing `json-new -> attached-session` bootstrap sequence, but explicit `attached-session` succeeds. The default page bootstrap strategy for this new `channel create` CDP path is now pinned to `attached-session` so the smoke passes out of the box.
  - The legacy `scripts/verify-channel-create-browser-smoke.ps1` entrypoint now defaults to the proven CDP wrapper, while explicit `-Driver cli` remains available as a diagnostic-only fallback path.
- Blocker:
  - The host still shows a CDP page bootstrap weakness on `auto/json-new`, so that strategy should not be reused as the default for this line without a host-specific reason.
- Next:
  1. Reuse the same CDP wrapper pattern for `group create` dialog browser evidence in the same Phase G pool.
  2. If needed later, keep `json-new` as a diagnostic-only contrast path, not the default runtime path, for this host.
