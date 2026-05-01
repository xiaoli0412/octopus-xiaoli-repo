# 2026-04-25 Phase G Backup Selector Contract Uniqueness Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup page selector contract tightening
- Timestamp: 2026-04-25T17:05:00+08:00

## Context Reused

- AGENTS.md
- docs/LLM-Gateway-Refactor-Plan.zh-CN.md
- docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
- docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md
- docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md
- docs/worklog/2026-04-25-phase-g-backup-replace-prune-selector-closure.md
- .codex-tmp/edit-memory/memory.md
- web/src/components/modules/setting/Backup.tsx
- web/src/components/modules/setting/Backup.test.tsx
- scripts/verify-backup-component.cjs

## Plan

- keep the same Phase G backup selector-contract mainline
- remove any remaining selector reuse between replace-prune and remaining-migration section 0
- add explicit test and verifier coverage for the unique selector contract
- finish with a text-level verification pass and record any host/runtime blockers clearly

## What Changed

- web/src/components/modules/setting/Backup.tsx
  - kept `backup-replace-prune-panel` / `backup-replace-prune-trigger` dedicated to the replace-prune preview block
  - kept `backup-remaining-migration-section-trigger-0` / `backup-remaining-migration-section-panel-0` dedicated to the first remaining-migration section
- web/src/components/modules/setting/Backup.test.tsx
  - added explicit assertions that the replace-prune trigger still exists alongside the remaining-migration section 0 selector pair
- scripts/verify-backup-component.cjs
  - added explicit assertions that replace-prune selectors and remaining-migration selectors all coexist in replace mode without reuse

## Verification

- Performed selector scans on `Backup.tsx`, `Backup.test.tsx`, and `scripts/verify-backup-component.cjs`
- Confirmed `backup-replace-prune-trigger` and `backup-replace-prune-panel` are still dedicated to the replace-prune preview block
- Confirmed `backup-remaining-migration-section-trigger-0` and `backup-remaining-migration-section-panel-0` are still dedicated to the first remaining-migration section
- Could not run repo-local JS execution end to end because the host Node runtime still has the known `ncrypto::CSPRNG` startup failure in this sandbox

## Risks / Blockers

- Repo-local Node execution remains blocked on this host, so the new assertions have not been executed yet
- This round stayed intentionally narrow to selector uniqueness and did not widen the backup UI surface

## Result

- Outcome: partial success
- This round restored the intended selector separation and added coverage so future regressions are easier to catch

## Next Step

1. rerun repo-local JS verification once the Node host/runtime issue is available again
2. if the backup selector contract stays stable, move to the next Phase G backup-page evidence gap only if it does not disturb the current ids
3. keep the same mainline and avoid expanding into unrelated settings UI work