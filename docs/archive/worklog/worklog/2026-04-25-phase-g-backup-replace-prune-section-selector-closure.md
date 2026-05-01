# 2026-04-25 Phase G Backup Replace-Prune Section Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup page selector contract tightening
- Timestamp: 2026-04-25T19:00:00+08:00

## Context Reused

- AGENTS.md
- docs/LLM-Gateway-Refactor-Plan.zh-CN.md
- docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
- docs/worklog/2026-04-25-phase-g-backup-compatibility-toggle-closure.md
- web/src/components/modules/setting/Backup.tsx
- web/src/components/modules/setting/Backup.test.tsx
- scripts/verify-backup-component.cjs

## Plan

- keep the same Phase G backup selector-contract mainline
- add stable selectors for the replace-prune section cards and item rows
- retarget the replace-mode test and repo-local verifier to the new selectors
- verify with selector scans and record host/runtime blockers clearly

## What Changed

- web/src/components/modules/setting/Backup.tsx
  - added `backup-replace-prune-section-${section.key}` anchors for each replace-prune section card
  - added `backup-replace-prune-section-title-${section.key}` anchors for each replace-prune section title
  - added `backup-replace-prune-section-item-${section.key}-${index}` anchors for each item row
- web/src/components/modules/setting/Backup.test.tsx
  - replaced replace-prune text assertions with stable selector assertions for the channels and apiKeys sections
- scripts/verify-backup-component.cjs
  - replaced replace-prune text assertions with stable selector assertions for the channels and apiKeys sections

## Verification

- Not yet run: repo-local Node verification is still expected to hit the host `ncrypto::CSPRNG(nullptr, 0)` startup failure on this machine
- Pending: `node scripts/verify-backup-component.cjs`
- Pending: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`

## Risks / Blockers

- Host Node runtime may still block script execution before user code runs
- The backup page selector contract is getting denser, so future edits must keep replace-prune selectors distinct from remaining-migration selectors

## Result

- Outcome: partial success
- This round narrows the replace-prune contract to stable selectors and keeps the backup-page contract moving forward

## Next Step

1. rerun the repo-local JS verification once the Node host/runtime issue is available again
2. continue with the next smallest backup-page selector gap only if it does not disturb existing ids
3. keep the same Phase G backup mainline and avoid expanding into unrelated settings UI work
