# 2026-04-26 Phase G Backup Import Remaining Migration Component Coverage

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: import-path remaining-migration component/verifier follow-up
- Timestamp: 2026-04-26T02:11:55+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/worklog/2026-04-25-phase-g-backup-import-vs-history-remaining-migration-selector-split.md`
- `docs/worklog/2026-04-25-phase-g-backup-helper-cleanup-and-status-sync.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- add narrow component-test coverage for the import-side `backup-import-remaining-migration-*` selectors that were split from history/rollback in the previous round
- fix the replace-mode verifier/test scope so the apply-confirmation assertion no longer depends on an out-of-scope variable
- run the smallest direct verification available here, and record any host blocker instead of pretending the runtime checks passed

## What Changed

- `web/src/components/modules/setting/Backup.test.tsx`
  - added map-mode assertions that the import-side remaining-migration panel stays collapsed by default, then expands through `backup-import-remaining-migration-trigger` and `backup-import-remaining-migration-section-trigger-0`
  - added replace-mode assertions for the same import-side selector family so the previous import/history split now has component-level regression coverage on both branches
  - declared a local replace-mode `applyButton` handle before asserting confirmation gating, removing the latent out-of-scope reference risk in that test block
- `scripts/verify-backup-component.cjs`
  - updated the replace-mode verifier to explicitly open the import-side remaining-migration panel before checking section-panel selectors
  - added a local replace-mode `applyButton` lookup so the verifier matches the component test contract and no longer relies on a leaked variable assumption

## Verification

- Attempted runtime status check:
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - blocked on this host with `Internal Windows PowerShell error ... 8009001d`
- Passed static contract verification:
  - `rg -n "backup-import-remaining-migration|const applyButton = screen.getByTestId\('backup-apply-same-import-button'\)" web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
  - targeted line reads on both files to confirm the import-side selector family and replace-mode apply-button scope are now present and locally asserted
- Attempted direct Node-based verification, still host-blocked before user code runs:
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-backup-component.cjs`
  - `node --experimental-strip-types scripts/verify-backup-logic.mjs`
  - all still fail with `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`

## Risks / Blockers

- Node startup remains sandbox-host blocked by `ncrypto::CSPRNG(nullptr, 0)`, so this round could only complete static contract verification
- `runtime-win.ps1 -Action status` is additionally blocked by host PowerShell initialization failure, so there is no fresh runtime-status snapshot from this round
- `Backup.tsx` runtime behavior itself was not changed here; this round only tightened regression coverage and verifier assumptions around the already-split selector contract

## Result

- Outcome: partial success
- This round produced a real code increment on the same Phase G backup selector-contract mainline without widening into unrelated backup logic or settings UI churn

## Next Step

1. rerun the targeted Node checks once the host CSPRNG startup blocker is cleared
2. stay on the same backup selector-contract mainline and take the next smallest rollback-preview or browser-evidence selector closure
3. only widen beyond test/verifier coverage if the next blocker is confirmed to be inside `Backup.tsx` itself

## 2026-04-26 Round Update - Import Remaining Migration Selector Contract

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: import-side remaining-migration selector contract follow-up
- Timestamp: 2026-04-26T13:26:07.9305286+08:00

### Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/worklog/2026-04-26-phase-g-backup-import-remaining-migration-component-coverage.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

### Plan

- stay on the same Phase G backup selector-contract mainline
- add stable selectors for the import-side remaining-migration title and summary
- mirror the same contract in the component test and repo-local verifier
- keep the write scope limited to the backup page, the test, and the verifier

### What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added `backup-import-remaining-migration-title`
  - added `backup-import-remaining-migration-summary`
- `web/src/components/modules/setting/Backup.test.tsx`
  - asserted the import-side remaining-migration title and summary in both map-mode and replace-mode flows
- `scripts/verify-backup-component.cjs`
  - asserted the same import-side title and summary selectors in both map-mode and replace-mode flows

### Verification

- Passed static selector scan with `rg -n "backup-import-remaining-migration-title|backup-import-remaining-migration-summary|backup-import-remaining-migration-trigger|backup-advanced-pending-title|backup-advanced-pending-summary" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
- Passed `git diff --check -- web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs` except for the repo's existing LF/CRLF warning on `Backup.tsx`
- Confirmed the touched files now contain the import-side remaining-migration selector contract and the history-side advanced-pending selector contract
- Could not complete Node-based runtime verification because the host still fails before user code runs with `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`

### Risks / Blockers

- Node-based verification is still host-blocked by the same `ncrypto::CSPRNG(nullptr, 0)` startup failure
- `Backup.tsx` still carries the existing LF/CRLF warning when checked by `git diff --check`
- This round only tightened selector contracts; it did not change backup business logic

### Result

- Outcome: partial success
- This round produced a real code increment on the same backup selector-contract mainline and closed the remaining import-side selector drift

### Next Step

1. keep the same backup selector-contract mainline and take the next smallest history-state or rollback-preview closure
2. rerun Node-based verification only after the host CSPRNG blocker is cleared
3. avoid widening into backup business-logic changes until the selector-contract lane is exhausted
