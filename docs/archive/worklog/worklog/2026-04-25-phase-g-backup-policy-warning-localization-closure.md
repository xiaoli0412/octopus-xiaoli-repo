# 2026-04-25 Phase G Backup Policy Warning Localization Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup detail localization
- Stage: backup helper-chain detail rendering closure

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- automation memory: `C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md`
- recent worklogs in `docs/worklog/2026-04-24-phase-g-backup-*.md`

## Plan

- keep the same Phase G backup no-browser chain
- close one small visible leak: localized sentence-level model-policy warnings in backup details
- verify with the stable no-browser helper path first

## What Changed

- `web/src/components/modules/setting/backup-logic.ts`
  - switched model-policy warning localization to a clean helper path for:
    - `billing_mode changed from ... to ...`
    - `probe_policy changed from ... to ...`
    - `probe_interval changed from ... to ...`
    - `probe_concurrency changed from ... to ...`
    - `model:... concurrent probe/race may increase cost`
  - kept English output unchanged while restoring readable zh-Hans / zh-Hant / ja sentence rendering
- `scripts/verify-backup-logic.mjs`
  - updated the no-browser assertion path to lock the corrected zh-Hans warning rendering
  - added one English-path assertion so this helper chain no longer only proves the Chinese branch
- `web/src/components/modules/setting/Backup.tsx`
  - repaired one local string-template corruption around the backup text block so the file no longer fails at the earlier parser-broken lines in this area

## Verification

- Passed: `node scripts/verify-backup-logic.mjs`
- Failed but investigated: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`

## Current Blockers / Risks

- `Backup.tsx` still has a larger pre-existing type/integration drift unrelated to this helper-only closure:
  - missing `TEXT` / `HELP_TEXT` / `HiddenTestText`
  - prop contract drift like `englishTitle`
  - missing export summary fields
- `web/src/provider/locale.tsx` still has a pre-existing AI automation locale shape mismatch
- `backup-logic.ts` still reports one remaining TS spread typing complaint in the warning-pattern dispatch path and should be cleaned next round together with helper cleanup
- Vitest was not rerun; host still has known worker instability (`spawn EPERM`)

## Result

- Outcome: partial success
- This round still produced a real code increment and a green no-browser verification result on the targeted backup helper chain

## Next Step

1. stay on the same Phase G backup pool and finish the tiny `backup-logic.ts` TS helper typing cleanup so helper-only verification is type-clean
2. then decide whether to keep closing backup helper detail leaks or switch to the larger `Backup.tsx` contract-repair task as a separate bounded closure
3. keep using repo-local no-browser verifiers first on this host
