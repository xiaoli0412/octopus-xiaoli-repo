# 2026-04-24 Phase G Homepage Layout Browser Smoke Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / homepage layout browser evidence`
- Summary: stayed on the same screenshot-first Phase G pool and closed the missing browser-grade evidence for homepage layout. Added stable browser-smoke anchors across the homepage structure, extended the host-approved Edge CDP smoke to a dedicated `home-layout` scenario with minimal seeded runtime data, and verified desktop layout, default collapsed runtime summary, expanded runtime cards, and `375px` mobile readability on this host.
- Verification:
  - `node scripts/verify-home-layout.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-home-layout-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
- Changed files:
  - `web/src/components/modules/home/index.tsx`
  - `web/src/components/modules/home/total.tsx`
  - `web/src/components/modules/home/token-breakdown.tsx`
  - `web/src/components/modules/home/chart.tsx`
  - `web/src/components/modules/home/rank.tsx`
  - `web/src/components/modules/home/activity.tsx`
  - `scripts/verify-home-layout.mjs`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-home-layout-browser-smoke.ps1`
- Result:
  - Homepage now exposes stable `data-testid` anchors for the page root, main grid, total summary area, breakdown area, chart, rank, activity grid, and runtime detail cards, so browser smoke no longer relies on brittle text-only matching.
  - The shared Edge CDP smoke path now supports a `home-layout` scenario that seeds one minimal channel/group/API key plus one real gateway request, which is enough to exercise homepage total metrics, chart rendering, token breakdown, and activity heatmap without introducing a wider backend task.
  - The browser smoke verifies three key homepage contracts on this host: desktop modules stay visible and proportional, runtime details remain collapsed by default and expand on demand, and `375px` width does not introduce large horizontal overflow.
- Blocker:
  - No homepage browser-evidence blocker remains in the current host path. Remaining homepage gaps are now limited to future copy/layout regressions or deeper data-contract expansion beyond the current estimated/runtime summaries.
- Next:
  1. Move to the next Phase G screenshot-first browser-evidence or localized-copy gap, with `CC Switch` browser evidence and deeper Chinese-page English leaks as the highest-value adjacent items.
  2. Keep homepage on the closed path unless future UI changes alter the selector contract or summary layering.
