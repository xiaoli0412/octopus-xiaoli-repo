# 2026-04-26 Phase G Backup Compatibility Details Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: compatibility-details and replace-prune summary selector coverage
- Timestamp: 2026-04-26T08:13:xx+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/worklog/2026-04-26-phase-g-backup-compatibility-details-selector-closure.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- close the remaining small selector gaps around the compatibility summary, replace-prune summary, and history advanced-pending title
- keep the write scope limited to the backup page, component test, and repo-local verifier
- run a static selector scan and record host blockers instead of pretending the Node runtime is healthy

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added stable selectors for the compatibility summary/title
  - added stable selectors for the replace-prune title/summary
  - added stable selectors for the history advanced-pending title/summary block so the existing warning copy has a real selector target
- `web/src/components/modules/setting/Backup.test.tsx`
  - switched the compatibility summary assertion to `backup-compatibility-summary`
  - switched the history advanced-pending assertion to `backup-advanced-pending-title`
- `scripts/verify-backup-component.cjs`
  - switched compatibility summary / replace-prune summary assertions to stable selectors
  - kept the existing replace-prune section assertions intact

## Verification

- Passed static selector scan:
  - `rg -n "backup-advanced-pending-title|backup-advanced-pending-summary|backup-compatibility-summary|backup-compatibility-title|backup-replace-prune-title|backup-replace-prune-summary|Advanced migration tooling still pending|Detailed diagnostics stay collapsed|Replace-prune preview|Channels to delete|backup-remaining-migration-section-item-rollback-tooling-advanced-pending-text" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs`
- Passed targeted file reads around the touched sections to confirm selector/assertion alignment
- Could not run full Node-based verification in this host session because the repository still hits the existing command-runner `ncrypto::CSPRNG(nullptr, 0)` blocker before user code runs

## Risks / Blockers

- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` and `node scripts/verify-backup-component.cjs` are still blocked by the host runtime before user code runs
- `pwsh -NoProfile -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status` remains unavailable in this shell because `Get-NetTCPConnection` is missing here
- `Backup.tsx` is still a large file because the selector-contract closure work has been incremental; the diff itself is expected, but the file should continue to be watched for accidental drift

## Result

- Outcome: partial success
- This round produced a real code increment on the same Phase G backup selector-contract mainline and closed the next smallest selector gaps without widening into backup business logic

## Next Step

1. rerun `tsc --noEmit` and `node scripts/verify-backup-component.cjs` once the command-runner blocker is cleared
2. continue on the same backup selector-contract mainline and take the next smallest history-state or browser-evidence closure
3. avoid widening into backup business-logic changes until the selector-contract lane is exhausted