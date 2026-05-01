# 2026-04-22 Phase F Backup Empty Scope Verification Fix

## Task
- name: Backup empty selective-scope verification fix
- date: 2026-04-22
- phase: Phase F / backup-import-rollback frontend closure
- milestone: Milestone 6 validation and deployment

## Inputs
- canonical: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` sections 11.5, 13, 14, 16
- workflow: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` Phase F
- previous_worklog: `docs/worklog/2026-04-22-phase-f-backup-remaining-tooling-shared-source-closure.md`
- local_resources: automation memory, frontend mainline status doc, `Backup.tsx`, `Backup.test.tsx`, `scripts/verify-backup-component.cjs`, `scripts/verify-backup-logic.mjs`, `scripts/vitest-no-spawn.cjs`
- goal: turn the Backup empty-scope path from false-positive coverage into real coverage, and keep the no-spawn verification chain aligned
- subagents: none; user requested main-thread-only execution

## Guardrails
- stay on Phase F backup/import/rollback
- keep scope inside Backup verification only
- do not change backend contracts or Backup behavior to satisfy a broken test

## Plan
1. inspect the Backup selective-import state flow and the existing empty-scope tests/scripts
2. confirm whether the empty-scope scenario is real or only appears covered
3. fix the Vitest draft and no-spawn component script to actually disable every import scope
4. rerun focused validation and record any remaining environment blockers

## Changes
- found that the old empty-scope scenarios only toggled `Routing` twice, which never produced a zero-scope selective-import state because the other five scopes stayed enabled by default
- added a shared `disableAllImportScopes` helper in `web/src/components/modules/setting/Backup.test.tsx`
- updated the empty selective-import disabled-state test to turn off all six scopes before asserting the disabled import button
- updated the empty selective-rollback fallback test to turn off all six scopes before asserting rollback falls back to full restore
- added a matching `disableAllImportScopes` helper in `scripts/verify-backup-component.cjs`
- wired `verifySelectiveImportDisabledWhenNoScopes` and `verifySelectiveRollbackScopeFallbackToFullRestore` into the no-spawn component verification main flow so those scenarios are actually executed now

## Validation
- passed: `node scripts/verify-backup-component.cjs`
- passed: `node scripts/verify-backup-logic.mjs`
- passed: `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- failed but recorded: `NODE_OPTIONS=--require ./scripts/vitest-no-spawn.cjs D:\gol1\node.exe .\web\node_modules\vitest\vitest.mjs run --config .\web\vitest.config.ts .\web\src\components\modules\setting\Backup.test.tsx .\web\src\components\modules\setting\backup-logic.test.ts --pool threads`
- blocker detail: local Vitest still fails during Vite/esbuild startup with Windows `spawn EPERM`, so the no-spawn script remains the reliable verification path on this host

## Risks And Handoff
- risk: low; this round only fixed verification coverage, not product behavior
- compatibility: no backend, schema, or import contract changes
- manual_smoke: still not run; browser/manual Backup smoke remains the next best evidence target
- next_task: prefer browser/manual Backup smoke if available; otherwise keep harvesting the next smallest same-mainline verification gap
