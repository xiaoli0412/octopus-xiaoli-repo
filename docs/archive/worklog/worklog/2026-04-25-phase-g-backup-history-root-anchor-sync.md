# 2026-04-25 Phase G Backup History Root Anchor Sync

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup history / rollback selector sync and status docs alignment

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- add one more page-level anchor to the history / rollback path without touching backup business logic
- keep the status docs aligned with the now-green backup selector contract summary
- verify with selector scans and record any host/runtime blockers clearly

## What Changed

- `web/src/components/modules/setting/Backup.test.tsx`
  - added `backup-page` root coverage to the history / rollback path so the page root is guarded beyond the export smoke flow
- `scripts/verify-backup-component.cjs`
  - added the same `backup-page` root coverage to the history / rollback path so the repo-local verifier matches the component test contract
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - updated the backup summary bullet to mention the dry-run/apply root anchor coverage and the current `backup-page` selector contract state
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - updated the backup status bullet to keep the same summary aligned with the current selector contract state

## Verification

- Passed: `node -v`
- Passed: `rg -n "backup-page" web\src\components\modules\setting\Backup.test.tsx scripts\verify-backup-component.cjs docs\CURRENT_STATUS_AND_PLAN.zh-CN.md docs\FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- Failed: `node scripts\verify-backup-component.cjs`
- Failed: `node web\node_modules\typescript\bin\tsc --noEmit -p web\tsconfig.json`

## Risks / Blockers

- Repo-local Node script execution still hits the host `ncrypto::CSPRNG(nullptr, 0)` startup failure, so the stable proof path on this machine remains selector scans plus text-level validation.
- The backup page selector contract is now denser, so future edits must keep `backup-page` / history / rollback / remaining-migration anchors distinct.

## Result

- Outcome: partial success
- This round produced a real page-level selector increment and kept the summary docs aligned, but host Node execution is still blocked.

## Next Step

1. keep the same Phase G backup selector-contract mainline and look for the next smallest page-level selector gap, preferably one already exercised by `Backup.test.tsx`
2. if Node verification becomes available, rerun `tsc --noEmit` and `scripts/verify-backup-component.cjs` before widening scope
3. continue avoiding selector-id reuse across backup preview surfaces
