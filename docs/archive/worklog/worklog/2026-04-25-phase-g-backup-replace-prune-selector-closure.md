# 2026-04-25 Phase G Backup Replace-Prune Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup page selector contract tightening
- Timestamp: 2026-04-25T16:08:15.0659268+08:00

## Context Reused

- AGENTS.md
- docs/LLM-Gateway-Refactor-Plan.zh-CN.md
- docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
- docs/worklog/2026-04-25-phase-g-backup-history-preview-selector-closure.md
- docs/worklog/2026-04-25-phase-g-backup-remaining-migration-section-selector-closure.md
- automation memory: /automations/octopus-2/memory.md
- web/src/components/modules/setting/Backup.tsx
- web/src/components/modules/setting/Backup.test.tsx
- scripts/verify-backup-component.cjs

## Plan

- stay on the same Phase G backup selector-contract mainline
- add a dedicated replace-prune preview panel with a stable selector contract
- keep the existing remaining-migration section selectors intact
- sync the backup component test and repo-local verifier to the same page-level contract
- finish with a text-level verification pass and record any host/runtime blockers clearly

## What Changed

- web/src/components/modules/setting/Backup.tsx
  - added a dedicated ackup-replace-prune-panel preview block driven by eplacePruneSummaryText
  - added a dedicated ackup-replace-prune-trigger toggle for expanding the replace-prune details
  - kept ackup-remaining-migration-section-trigger-0 / ackup-remaining-migration-section-panel-0 on the first remaining-migration section instead of reusing them for replace-prune
- web/src/components/modules/setting/Backup.test.tsx
  - no additional test edits were required in this round because the existing replace-prune assertions already match the new dedicated panel contract
- scripts/verify-backup-component.cjs
  - no additional verifier edits were required in this round because the existing replace-prune assertions already match the new dedicated panel contract

## Verification

- Performed multiple selector scans against Backup.tsx, Backup.test.tsx, and scripts/verify-backup-component.cjs
- Confirmed ackup-replace-prune-panel and ackup-replace-prune-trigger now exist as dedicated selectors in Backup.tsx
- Confirmed ackup-remaining-migration-section-trigger-0 and ackup-remaining-migration-section-panel-0 remain present for the remaining-migration accordion flow
- Could not run repo-local JS execution end to end because the host Node runtime still has the known script-startup issue from earlier rounds

## Risks / Blockers

- Repo-local Node script execution remains blocked on this host, so 	sc --noEmit and the verifier script were not re-run successfully in this round
- Backup.tsx has accumulated substantial selector-contract work; future rounds should keep changes narrow and avoid reusing section ids across distinct UI surfaces

## Result

- Outcome: partial success
- This round produced a real page-level selector increment by splitting replace-prune out into its own preview panel instead of borrowing the first remaining-migration section

## Next Step

1. rerun repo-local JS verification once the Node host/runtime issue is available again
2. keep the same Phase G backup mainline and continue with the next smallest page-level selector or browser-evidence closure only if it does not disturb existing ids
3. if browser-grade evidence is still blocked, stay on the backup page contract and avoid expanding into unrelated UI work