# 2026-04-26 Phase G Backup History Size Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: snapshot-history size selector contract closure
- Timestamp: 2026-04-26T16:00:00+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- recent backup-mainline worklogs:
  - `docs/worklog/2026-04-26-phase-g-backup-import-remaining-migration-component-coverage.md`
  - `docs/worklog/2026-04-26-phase-g-backup-rollback-preview-summary-selector-closure.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- current write scope:
  - `web/src/components/modules/setting/Backup.tsx`

## Plan

- stay on the same Phase G backup selector-contract mainline
- close a real implementation/test drift instead of adding a new speculative selector
- align the snapshot-history card with the already-existing component-test and repo-local verifier contract for `backup-history-item-size`
- keep verification to static selector scans and diff hygiene because Node startup is still host-blocked here

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added the missing `backup-history-item-size` node inside each snapshot-history card
  - reused the existing `formatFileSize` helper so the card now emits the same `Size：2 KB` contract already asserted by `Backup.test.tsx` and `scripts/verify-backup-component.cjs`
  - kept the fallback explicit with `Size：Unknown` when `size_bytes` is absent, matching the rest of the snapshot metadata style

## Verification

- Passed static selector scan:
  - `rg -n "backup-history-item-size|formatFileSize\(snapshot\.size_bytes\)|snapshot\.size_bytes !== undefined" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
- Passed targeted touched-file read to confirm the node now exists exactly where the history meta block renders
- Passed targeted diff hygiene check except for the repo's existing line-ending warning:
  - `git diff --check -- web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
  - result still reports the repo's existing `LF will be replaced by CRLF` warning on `Backup.tsx`
- Did not rerun Node-based verification in this round because the host remains blocked before user code runs with `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`

## Risks / Blockers

- Node-based checks remain blocked by the host `ncrypto::CSPRNG(nullptr, 0)` startup failure, so this round only completed static contract verification
- `Backup.tsx` still reports the repo's existing LF/CRLF warning under `git diff --check`
- this round intentionally stayed inside the selector-contract lane and did not change backup import/rollback business logic

## Result

- Outcome: success
- This round produced a small but real code increment on the same backup selector-contract mainline and removed an implementation/test drift that was already present in the repository

## Next Step

1. keep the same Phase G backup selector-contract mainline
2. close the next smallest snapshot-history or rollback-preview field-level selector drift that is already asserted elsewhere or visible as text-only metadata
3. rerun `node scripts/verify-backup-component.cjs` and `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` only after the host Node/CSPRNG blocker is cleared
