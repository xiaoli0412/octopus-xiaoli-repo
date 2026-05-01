# 2026-04-25 Phase G Backup Remaining Migration Section Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page follow-up
- Stage: backup page selector contract tightening
- Timestamp: 2026-04-25T15:31:26+08:00

## Context Reused

- AGENTS.md
- docs/LLM-Gateway-Refactor-Plan.zh-CN.md
- docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
- docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md
- web/src/components/modules/setting/Backup.tsx
- web/src/components/modules/setting/Backup.test.tsx
- scripts/verify-backup-component.cjs
- latest backup selector-contract worklogs from 2026-04-25
- automation memory: $env:CODEX_HOME/automations/octopus-2/memory.md

## Plan

- keep the same Phase G backup mainline
- add stable selector anchors to the nested remaining-migration accordion sections so browser evidence no longer needs to click by text for inner expansion
- retarget the existing backup test and repo-local verification script to the new section selectors
- verify with repo-local type/no-browser commands if the host allows it; otherwise record the host blocker clearly

## What Changed

- web/src/components/modules/setting/Backup.tsx
  - added stable data-testid anchors for the nested remaining-migration section items, their triggers, and their panels
- web/src/components/modules/setting/Backup.test.tsx
  - retargeted the advanced migration test to ackup-remaining-migration-section-trigger-0 and ackup-remaining-migration-section-panel-0
- scripts/verify-backup-component.cjs
  - retargeted the repo-local verification path to click and assert the same nested section selectors

## Verification

- Attempted: 
ode web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json
- Attempted: 
ode scripts/verify-backup-component.cjs
- Blocked on this host by the same sandbox Node startup failure (Assertion failed: ncrypto::CSPRNG(nullptr, 0))

## Risks / Blockers

- Repo-local Node validation still fails before user code runs on this machine, so the proof path remains blocked by host/runtime startup rather than by the backup-page changes themselves
- The nested remaining-migration section selectors are now stable, but browser-grade evidence still needs a host that can actually start Node/Vitest successfully

## Result

- Outcome: partial success
- This round produced a real backup-page selector increment and retargeted both the component test and the repo-local verification script, but host Node startup still blocks the requested verification commands

## Next Step

1. keep the same Phase G backup mainline and use the new nested remaining-migration section selectors in any later browser-grade evidence or smoke update
2. if the host keeps blocking Node startup, continue with the next smallest backup-page selector contract cleanup that does not require new runtime proof
3. when a healthier host is available, rerun 	sc --noEmit and erify-backup-component.cjs immediately
