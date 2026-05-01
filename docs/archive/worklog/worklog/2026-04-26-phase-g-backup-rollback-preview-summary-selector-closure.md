# 2026-04-26 Phase G Backup Rollback Preview Summary Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: rollback-preview compatibility-summary selector closure
- Timestamp: 2026-04-26T15:xx:xx+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- recent backup-mainline worklogs:
  - `docs/worklog/2026-04-26-phase-g-backup-rollback-preview-selector-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-history-advanced-pending-selector-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-import-remaining-migration-component-coverage.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- current write scope:
  - `web/src/components/modules/setting/Backup.tsx`
  - `web/src/components/modules/setting/Backup.test.tsx`
  - `scripts/verify-backup-component.cjs`
  - `scripts/verify-backup-component.setting-mock.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- add stable selectors for the rollback-preview compatibility summary row so the preview is no longer only covered by overview text and meta cells
- sync the component test, repo-local verifier, and verifier mock data to the same selector contract
- run static selector scans and diff hygiene checks; record the existing host blocker instead of pretending Node-based verification passed

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added localized rollback summary labels for conflicts, credential rebinds, and preview warnings
  - derived `rollbackRebindCount` and `rollbackWarningCount` from the existing rollback preview state
  - added stable selectors for the new rollback preview summary grid:
    - `backup-rollback-preview-summary-grid`
    - `backup-rollback-preview-summary-conflicts`
    - `backup-rollback-preview-summary-rebinds`
    - `backup-rollback-preview-summary-warnings`
- `web/src/components/modules/setting/Backup.test.tsx`
  - expanded the rollback preview mock to include one conflict, one credential rebind target, and one preview warning
  - asserted the new rollback summary selectors and their `en` locale content
- `scripts/verify-backup-component.cjs`
  - asserted the same rollback summary selectors in the repo-local verifier
- `scripts/verify-backup-component.setting-mock.cjs`
  - aligned the rollback preview mock with one conflict, one credential rebind target, and one preview warning so the verifier can satisfy the new contract once Node execution is available again

## Verification

- Passed static selector scan:
  - `rg -n "rollbackSummaryConflicts|rollbackSummaryRebinds|rollbackSummaryWarnings|backup-rollback-preview-summary-grid|backup-rollback-preview-summary-conflicts|backup-rollback-preview-summary-rebinds|backup-rollback-preview-summary-warnings|credential_rebind_targets: \[\{|preview_warnings: \['provider mismatch'\]" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs scripts/verify-backup-component.setting-mock.cjs`
- Passed targeted diff hygiene check:
  - `git diff --check -- web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs scripts/verify-backup-component.setting-mock.cjs`
  - result still includes the repo's existing LF/CRLF warning on `Backup.tsx`
- Attempted Node-based verification:
  - `node scripts/verify-backup-component.cjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Runtime validation remained host-blocked:
  - both Node-based commands still fail before user code runs with `Assertion failed: ncrypto::CSPRNG(nullptr, 0)` from the current command-runner host path

## Risks / Blockers

- This round only tightened the rollback-preview selector contract; it did not change rollback import semantics or backup business logic
- Node-based verification remains blocked by the host `ncrypto::CSPRNG(nullptr, 0)` startup failure
- `Backup.tsx` still reports the repo's existing LF/CRLF warning under `git diff --check`

## Result

- Outcome: partial success
- This round produced a real code increment on the same backup selector-contract mainline and removed one more rollback-preview coverage gap without widening into backup logic refactors

## Next Step

1. keep the same Phase G backup selector-contract mainline
2. close the next smallest rollback-preview or history-state field-level selector gap, preferably around rows summary or imported-at/manifest metadata if any text-only assertions remain
3. rerun `node scripts/verify-backup-component.cjs` and `node ...tsc --noEmit` only after the host Node/CSPRNG blocker is cleared
