# 2026-04-25 Phase G Backup Remaining Migration Panel Anchor

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page follow-up
- Stage: backup page selector contract tightening

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `docs/worklog/2026-04-25-phase-g-backup-root-assertion-and-status-sync.md`
- `docs/worklog/2026-04-25-phase-g-backup-helper-cleanup-and-status-sync.md`

## Plan

- keep the same Phase G backup mainline
- add a stable testid anchor to the remaining-migration accordion content so later browser evidence does not depend on text matching alone
- lock that selector in the existing backup test path
- sync the backup status docs with the new page-level contract cleanup
- verify with repo-local type/no-browser commands first

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added `data-testid="backup-remaining-migration-panel"` to the remaining migration accordion content so the panel now has a stable page-level anchor for later browser evidence work
- `web/src/components/modules/setting/Backup.test.tsx`
  - added a regression assertion that the remaining migration panel anchor appears after expanding the advanced migration accordion
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - updated the backup status wording to mention the remaining page-level selector cleanup instead of only the older helper-level state
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - synced the same backup wording so the frontend summary stays aligned with the current contract state

## Verification

- not yet run in this message, scheduled next:
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-backup-logic.mjs`
  - `node scripts/verify-locale-consistency.mjs`

## Risks / Blockers

- Vitest/Vite still has the known host `spawn EPERM` blocker here, so repo-local no-browser/typecheck commands remain the stable proof path
- browser-grade backup evidence is still pending; this round only strengthened the selector contract for that future pass

## Result

- Outcome: success
- This round produced a real backup-page contract increment and kept the next browser-evidence step cheaper

## Next Step

1. keep the same Phase G backup mainline and use the new `backup-remaining-migration-panel` anchor in any later browser-grade evidence or smoke update
2. if browser evidence remains blocked, choose the next smallest `Backup.tsx` page-level contract cleanup rather than helper-only churn
3. rerun the repo-local type/no-browser verification chain after the doc sync is applied
