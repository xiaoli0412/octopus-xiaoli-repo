# 2026-04-25 Phase G Backup Import Vs History Remaining Migration Selector Split

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: backup page selector uniqueness follow-up
- Timestamp: 2026-04-25T22:49:16+08:00

## Context Reused

- AGENTS.md
- docs/LLM-Gateway-Refactor-Plan.zh-CN.md
- docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
- docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md
- docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md
- docs/worklog/2026-04-25-phase-g-backup-selector-uniqueness-closure.md
- docs/worklog/2026-04-25-phase-g-backup-remaining-migration-section-selector-closure.md
- web/src/components/modules/setting/Backup.tsx
- web/src/components/modules/setting/Backup.test.tsx
- scripts/verify-backup-component.cjs

## Plan

- stay on the same Phase G backup selector-contract mainline
- remove selector reuse between import-result remaining-migration controls and history/rollback remaining-migration controls
- retarget repo-local verifier references so the import path uses its own selector prefix while rollback keeps the old stable ids
- finish with static selector scans and record the still-active host verification blocker

## What Changed

- web/src/components/modules/setting/Backup.tsx
  - split the import-result remaining-migration controls onto a dedicated selector family: ackup-import-remaining-migration-*
  - kept the history/rollback remaining-migration controls on the existing ackup-remaining-migration-* ids so rollback-path tests remain stable
  - split the state holders too, so import-side and history-side expansion no longer share the same toggle state
- scripts/verify-backup-component.cjs
  - updated the zh-Hans import-path and replace-mode import-path assertions to target the new ackup-import-remaining-migration-* selectors
  - kept rollback preview assertions on the original ackup-remaining-migration-* selectors

## Verification

- Passed static selector scan on Backup.tsx: duplicate remaining-migration selectors removed; only ackup-apply-same-import-button still appears twice in static text scan because of DOM query string + rendered button id
- Passed static grep alignment on Backup.tsx and erify-backup-component.cjs for the new import/history selector split
- Blocked runtime verification:
  - 
ode web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json
  - 
ode scripts/verify-backup-component.cjs
  - both still fail before user code runs on this host with Assertion failed: ncrypto::CSPRNG(nullptr, 0)

## Risks / Blockers

- Node startup remains host-blocked in this sandbox, so this round could only complete static contract verification
- Backup.test.tsx still references rollback-path remaining-migration selectors only; that is intentional after the split, but import-path selector coverage in the component test is still a good next follow-up when runtime execution is available again

## Result

- Outcome: partial success
- This round produced a real code increment that narrows selector ambiguity on the backup page without widening scope beyond the current Phase G backup mainline

## Next Step

1. rerun 	sc --noEmit and erify-backup-component.cjs as soon as a host with working Node entropy/CSPRNG startup is available
2. if runtime validation is still blocked, add one narrow Backup.test.tsx import-path assertion for ackup-import-remaining-migration-* to mirror the verifier split
3. keep the rollback-path selector ids unchanged unless a documented contract change is needed
