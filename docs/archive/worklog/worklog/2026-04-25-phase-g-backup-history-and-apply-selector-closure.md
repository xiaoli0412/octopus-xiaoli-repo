# 2026-04-25 Phase G Backup History And Apply Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup page interaction selector hardening

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- automation memory: `$CODEX_HOME/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup mainline
- harden one more page-level selector contract around snapshot history open/close behavior
- if time remains, tighten the apply-same-import confirmation path onto stable selectors too
- verify with repo-local runtime status, frontend typecheck, and backup no-browser verification

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added `data-testid="backup-history-trigger"` to the snapshot-history toggle button
  - added `data-testid="backup-history-panel"` to the expanded snapshot-history container
  - added `data-testid="backup-apply-confirm-panel"` to the apply-same-import confirmation block
  - added `data-testid="backup-apply-confirm-switch"` to the apply confirmation switch
- `web/src/components/modules/setting/Backup.test.tsx`
  - switched rollback-history test setup from text-based `Show` matching to `backup-history-trigger`
  - asserted `backup-history-panel` and `backup-apply-confirm-panel` directly
  - added a selector-based helper for the apply confirmation switch so the test no longer depends on the full localized confirmation sentence
- `scripts/verify-backup-component.cjs`
  - switched the history-open path to `backup-history-trigger`
  - added direct assertions for `backup-history-panel` and `backup-apply-confirm-panel`
  - switched apply confirmation in both dry-run/apply and replace flows to `backup-apply-confirm-switch`

## Verification

- Passed: `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- Passed: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Passed: `node scripts/verify-backup-component.cjs`

## Risks / Blockers

- Browser-grade backup evidence is still the next real gap; this round only hardened the page-level selector contract beneath that future smoke path.
- `web/src/components/modules/setting/Backup.tsx` is already a large in-flight file in this workspace, so future rounds should keep edits narrowly scoped and continue avoiding unrelated cleanup.

## Result

- Outcome: success
- This round produced a real selector-contract increment on the same backup mainline and kept the repo-local proof path green.

## Next Step

1. stay on the same Phase G backup mainline and use `backup-page`, `backup-history-trigger`, `backup-history-panel`, `backup-history-list`, `backup-rollback-preview-panel`, and `backup-apply-confirm-panel` as the stable browser-smoke anchors
2. if browser evidence remains host-blocked, pick the next smallest `Backup.tsx` page-level contract cleanup around import summary / compatibility toggles instead of returning to helper-only churn
3. keep `runtime-win.ps1 -Action status`, `tsc --noEmit`, and `verify-backup-component.cjs` as the stable local proof path on this machine
