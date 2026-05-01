# 2026-04-25 Phase G Backup History Meta Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup history selector contract tightening
- Timestamp: 2026-04-25T20:20:00+08:00

## Context Reused

- AGENTS.md
- docs/LLM-Gateway-Refactor-Plan.zh-CN.md
- docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
- docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md
- docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md
- docs/worklog/2026-04-25-phase-g-backup-selector-uniqueness-closure.md
- docs/worklog/2026-04-25-phase-g-backup-selector-contract-follow-up.md
- web/src/components/modules/setting/Backup.tsx
- web/src/components/modules/setting/Backup.test.tsx
- scripts/verify-backup-component.cjs
- .codex-home-link/automations/octopus-2/memory.md

## Plan

- keep the same Phase G backup mainline
- add a dedicated selector anchor for backup history item meta content
- cover the new anchor in the component test and the repo-local verifier
- finish with typecheck and verifier validation, then record any host/runtime blockers clearly

## What Changed

- web/src/components/modules/setting/Backup.tsx
  - added `backup-history-item-meta` to the snapshot history item header row so the meta block has an explicit stable anchor
- web/src/components/modules/setting/Backup.test.tsx
  - added a direct assertion for `backup-history-item-meta` inside the selected snapshot history item
- scripts/verify-backup-component.cjs
  - added a selector assertion for `backup-history-item-meta` in the snapshot history verification path

## Verification

- pending after edits

## Risks / Blockers

- repo-local Node execution has previously hit the known `ncrypto::CSPRNG` startup issue in this sandbox
- if the verifier still fails to execute locally, the next best path is to keep the same selector-contract mainline and close the next smallest backup-page contract gap without changing ids

## Result

- Outcome: in progress

## Next Step

1. run `tsc --noEmit` for the web workspace
2. run `scripts/verify-backup-component.cjs`
3. if Node runtime is still blocked, record the blocker and continue with the next smallest backup selector-contract closure on the same mainline

