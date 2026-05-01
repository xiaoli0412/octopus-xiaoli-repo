# 2026-04-24 Phase G Channel Create Browser Smoke Hookup

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first UI closure / channel create browser smoke hookup`
- Summary: kept the same screenshot-first Phase G pool, added stable browser-smoke selectors for the channel create trigger and dialog, and landed a dedicated `channel create` browser smoke script that reuses the same authenticated Playwright CLI flow used by the settings help smoke line.
- Verification:
  - `node scripts/verify-channel-create-flow.mjs`
  - `node scripts/verify-channel-presentation.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-channel-create-browser-smoke.mjs --self-start` -> blocked by host `spawn EPERM` when Node tries to spawn backend / Playwright CLI subprocesses
  - `node scripts/verify-channel-create-browser-smoke.mjs --external` with `OCTOPUS_UI_SMOKE_FRONTEND_URL=http://127.0.0.1:8080` and `OCTOPUS_UI_SMOKE_BACKEND_URL=http://127.0.0.1:8080` -> blocked by the same host `spawn EPERM` on Playwright CLI launch
  - `D:\gol1\node.exe .\node_modules\next\dist\bin\next build` (in `web/`) -> compile phase succeeded, later failed with host `spawn EPERM`
  - `node scripts/sync-web-static.mjs`
  - `Invoke-WebRequest http://127.0.0.1:8080/healthz`
- Changed files:
  - `web/src/components/ui/morphing-dialog.tsx`
  - `web/src/components/modules/toolbar/index.tsx`
  - `web/src/components/modules/channel/Create.tsx`
  - `web/src/components/modules/channel/Form.tsx`
  - `scripts/verify-channel-create-flow.mjs`
  - `scripts/verify-channel-create-browser-smoke.mjs`
- Result:
  - Stable `data-testid` / `data-slot` anchors now exist for `channel create` browser smoke: toolbar trigger, dialog root, flow card, basic section, key section, key filter, key trigger, and key primary area.
  - A dedicated browser smoke script is now ready to verify `channel create` dialog open, first key expand, help-hint hover, and `375px` layout once the host allows Node child-process spawning.
  - No-browser verification and typecheck stayed green after the selector hookup.
- Blocker:
  - Current host environment rejects Node child-process launch with `spawn EPERM`, which blocks both the new `channel create` smoke script and a fresh `next build` completion path.
  - Existing Octopus service at `http://127.0.0.1:8080` is healthy, so the remaining blocker is not page logic or backend health; it is host-level subprocess policy.
- Next:
  1. reuse the new selector contract and rerun `scripts/verify-channel-create-browser-smoke.mjs` from a host path that permits Playwright CLI spawning, or drive it through an approved PowerShell/browser wrapper path.
  2. once `channel create` browser evidence is captured, reuse the same pattern for `group create` dialog browser smoke in the same Phase G pool.
  3. if host subprocess policy stays blocked, continue with the next same-pool no-browser `channel create` readability / copy closure instead of leaving the line idle.
