# 2026-04-24 Phase G Group Create CDP Browser Smoke Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / group create browser evidence`
- Summary: reused the verified host-level Edge CDP browser-smoke pattern for `group create`, kept the stable `new-group` / `edit-group` testid and dialog anchors, and confirmed the legacy default PowerShell entrypoint now passes end-to-end on this host.
- Verification:
  - `node scripts/verify-group-create-flow.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-group-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
- Changed files:
  - `web/src/components/modules/group/Create.tsx`
  - `web/src/components/modules/group/Card.tsx`
  - `web/src/components/modules/group/Editor.tsx`
  - `scripts/verify-group-create-flow.mjs`
  - `scripts/verify-group-create-browser-smoke-cdp.mjs`
  - `scripts/verify-group-create-browser-smoke-cdp.ps1`
  - `scripts/verify-group-create-browser-smoke.mjs`
  - `scripts/verify-group-create-browser-smoke.ps1`
- Result:
  - `group create` now has stable browser-smoke anchors for the dialog, form, flow card, naming/mode/model/selected sections, and advanced strategy section, so the smoke can assert the real structure instead of relying on loose text matching.
  - The CDP wrapper follows the same host-approved Edge remote-debugging path as `channel create`, with `attached-session` remaining the proven page bootstrap strategy on this machine.
  - The legacy `scripts/verify-group-create-browser-smoke.ps1` entrypoint now defaults to the CDP wrapper, so the main smoke command passes without the earlier `@playwright/cli` spawn path.
- Blocker:
  - `auto/json-new` remains diagnostic-only on this host; keep it as a contrast path, not the default bootstrap strategy.
- Next:
  1. Move to the next Phase G screenshot-first item, starting with the remaining channel/page/browser evidence gaps that still show up in the status docs.
  2. Keep group create on the closed path unless a future regression changes the selector contract.
