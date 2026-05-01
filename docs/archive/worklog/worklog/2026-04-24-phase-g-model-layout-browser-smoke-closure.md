# 2026-04-24 Phase G Model Layout Browser Smoke Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / model layout browser evidence`
- Summary: stayed on the same screenshot-first Phase G pool and closed the missing browser-grade evidence for the model/price layout toggle. Added stable selector anchors to the model page and toolbar layout controls, then reused the host-approved Edge CDP wrapper pattern to verify normal/compact layout switching and `375px` mobile width with seeded model data.
- Verification:
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `pnpm --dir web run build:static`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-model-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
- Changed files:
  - `web/src/components/modules/model/index.tsx`
  - `web/src/components/modules/model/Item.tsx`
  - `web/src/components/modules/toolbar/index.tsx`
  - `scripts/verify-llm-price-boundary.mjs`
  - `scripts/verify-channel-create-browser-smoke-cdp.ps1`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-model-layout-browser-smoke.ps1`
  - `docs/worklog/2026-04-24-phase-g-model-layout-browser-smoke-closure.md`
- Result:
  - The model page now exposes stable browser-smoke anchors for the page root, per-card layout state, and toolbar layout buttons, so the smoke can assert real normal/compact transitions instead of relying on loose text or DOM shape guesses.
  - The CDP smoke seeds deterministic model data through the admin API, enters the model page through persisted nav state, verifies `grid -> list -> grid` switching, and confirms the page remains readable at `375px` without large horizontal overflow.
  - The model layout PowerShell entrypoint reuses the proven CDP wrapper infrastructure already used by the closed channel/group create browser lines on this host.
- Blocker:
  - `auto/json-new` remains diagnostic-only on this host, same as the other CDP-based screenshot-first lines.
- Next:
  1. Move to the next remaining Phase G browser evidence gap from the status docs, with homepage layout/browser evidence as the highest-value adjacent closure.
  2. Keep model layout on the closed path unless future UI changes alter the selector contract or layout-state plumbing.
