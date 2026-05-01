# 2026-04-25 Phase G Backup Page Selector Contract Follow-up

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page follow-up
- Stage: backup page selector contract and browser-evidence prep

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- automation memory: `$CODEX_HOME/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`
- `scripts/verify-backup-logic.mjs`

## Plan

- stay on the same Phase G backup mainline
- add one more selector-contract assertion to the backup component verification script so `backup-page` is guarded in the map/replace verification path too
- run repo-local typecheck and backup verification again
- sync the status docs / worklog / automation memory so the next round can move directly to browser-grade evidence on the same selector contract

## What Changed

- `scripts/verify-backup-component.cjs`
  - added a `backup-page` root-anchor assertion to the map/replace verification path so the stable page selector is now guarded across export, dry-run/apply, and map/replace flows

## Verification

- Passed: `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- Passed: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Passed: `node scripts/verify-backup-component.cjs`

## Risks / Blockers

- Browser-grade backup evidence is still the next real gap.
- Vitest/Vite/browser smoke remain host-sensitive on this machine, so the reliable local proof path is still the repo-local no-browser verification chain.

## Result

- Outcome: success
- This round produced a real page-level contract increment, kept the local verification chain green, and left a narrower browser-evidence next step.

## Next Step

1. stay on the same Phase G backup mainline and use the stable `backup-page` root anchor to pursue browser-grade evidence.
2. if browser evidence remains host-blocked, pick the next smallest `Backup.tsx` page-level contract cleanup instead of returning to helper-only backup localization.
3. keep `runtime-win.ps1 -Action status`, `tsc --noEmit`, and `verify-backup-component.cjs` as the stable local proof path on this machine.