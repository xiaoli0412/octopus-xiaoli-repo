# 2026-04-24 Phase G Channel Page Browser Smoke Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / channel page browser evidence`
- Summary: stayed on the same screenshot-first Phase G pool and closed the missing page-level browser evidence for the channel page. Added stable selector anchors across the channel list and channel detail path, extended the shared Edge CDP smoke to a dedicated `channel-page` scenario, and verified combined filtering, detail dialog, key expansion, and `375px` readability on this host.
- Verification:
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `node scripts/verify-channel-presentation.mjs`
  - `node scripts/verify-channel-create-flow.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node --check scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `node scripts/verify-channel-create-browser-smoke-cdp.mjs --check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
- Changed files:
  - `web/src/components/modules/channel/index.tsx`
  - `web/src/components/modules/channel/Card.tsx`
  - `web/src/components/modules/channel/CardContent.tsx`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-channel-page-browser-smoke.ps1`
  - `scripts/verify-channel-presentation.mjs`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-channel-page-browser-smoke-closure.md`
- Result:
  - Channel page now exposes stable `data-testid` anchors for the page root, per-card trigger, detail dialog, route-target summary, key filter, key accordion, and key-level model/test sections, so browser smoke no longer depends on brittle text-only selectors.
  - The shared Edge CDP smoke path now supports a `channel-page` scenario that seeds one deterministic channel with two keys plus one route-target override, then validates toolbar combined filtering, detail-dialog opening, key-row expansion, and `375px` width without large horizontal overflow.
  - `scripts/verify-channel-presentation.mjs` now guards the new selector contract in no-browser form, so the page-level browser evidence does not silently drift out of sync with future UI changes.
- Blocker:
  - No page-level blocker remains for the channel page on this host. Remaining channel-line gaps are now limited to `9.1.1` field depth and finer create/edit dialog hover/focus details outside the current page-level closure.
- Next:
  1. Stay on the same Phase G screenshot-first pool and pick the next smallest page-adjacent closure, preferably a remaining Chinese-copy leak or a finer browser interaction gap outside the already-closed channel/home/model/CC Switch lines.
  2. Only revisit `channel-page` smoke internals if a future UI change breaks the selector contract or shared CDP wrapper behavior.

## Follow-up correction (same day)

- A later host rerun after selector-contract tightening did not reproduce the claimed browser-grade pass.
- The code-side contract improved further: toolbar filter controls gained stable test ids, the shared CDP smoke stopped relying on button order and input order, and the channel-page card count no longer over-counts badge/metric descendants as cards.
- Verified green on the code/script side:
  - `node scripts/verify-channel-presentation.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node --check scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `node scripts/verify-channel-create-browser-smoke-cdp.mjs --check-only`
  - `node scripts/verify-channel-create-flow.mjs`
- Host blocker still present on this machine:
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
    - failed before page assertions with `Timed out waiting for http://127.0.0.1:9222/json/version`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\bootstrap-edge-cdp.ps1 -Port 9233 -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -ReadyTimeoutSeconds 30 -StableReadySeconds 3 -OutputJsonPath .codex-tmp\edge-cdp-bootstrap.json`
    - Edge stderr still showed host-level `Access is denied (0x5)` plus crashpad/network sandbox access failures.
- Accurate status after the rerun:
  - `channel-page` selector/script contract is closed.
  - Browser-grade pass is still blocked by host Edge/CDP bootstrap behavior on this machine.
