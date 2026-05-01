# 2026-04-21 Phase F Rollback Domain Copy Alignment

## Plan

- current_mainline: Phase F backup/import/rollback frontend closure
- current_stage: 11.5.4 import and rollback preview consistency
- core_task: align the remaining migration-tooling footer copy in `web/src/components/modules/setting/Backup.tsx` with the current selective rollback flow
- support_task: add matching assertions in the Backup component test and the node-based verification script
- completion_criteria: the footer no longer claims a plain partial-restore gap, the new rollback-domain wording is asserted in tests, and the web typecheck / backup verification chain passes

## Inputs

- canonical: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` section `11.5.4`
- workflow: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` Phase F section
- previous_worklogs:
  - `docs/worklog/2026-04-21-phase-f-result-mode-copy-alignment.md`
  - `docs/worklog/2026-04-21-phase-f-selective-import-apply-scope-verification.md`
- local_resources: automation memory, current Phase F worklog chain, `web/src/components/modules/setting/Backup.tsx`, `web/src/components/modules/setting/Backup.test.tsx`, `scripts/verify-backup-component.cjs`, `git status --short`
- skills_and_context: `using-superpowers`, `brainstorming`, current automation thread instructions
- subagents: none
- why_no_subagents: user explicitly requested main-thread execution and this task is a single-file UI copy alignment with local test coverage

## Guardrails

- touch only `web/src/components/modules/setting/Backup.tsx`, `web/src/components/modules/setting/Backup.test.tsx`, `scripts/verify-backup-component.cjs`, this worklog, and automation memory
- stay on `Phase F / 11.5.4`
- do not change backend contracts, route behavior, or rollback payload semantics
- keep the change limited to copy alignment plus direct verification

## Expected Change

- replace the stale partial-restore wording with rollback-domain wording that matches the current selective-scope override flow
- keep the footer honest about what is still pending, without underclaiming the current capability
- assert the new footer copy in both the Vitest draft and the node verification script

## Validation

- `node scripts/verify-backup-component.cjs`
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- targeted `rg` checks for the new footer copy

## Risks

- low product risk because only copy and assertions change
- the worktree is already dirty, so this run should stay tightly scoped
- browser/manual smoke is still not part of this micro-closure

## Next

- if this footer copy closes cleanly, the next Phase F micro-step should either tighten another Backup footer/microcopy mismatch or return to broader Phase F status-doc sync

## Run Result

- status: success
- files_changed: `web/src/components/modules/setting/Backup.tsx`, `web/src/components/modules/setting/Backup.test.tsx`, `scripts/verify-backup-component.cjs`, `docs/worklog/2026-04-21-phase-f-rollback-domain-copy-alignment.md`, `C:/Users/李昊桐/.codex/automations/octopus-2/memory.md`
- commands_run: `node scripts/verify-backup-component.cjs`; `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`; `rg -n -C 2 "granular rollback-domain editing|finer-grained partial restore|Rollback scope override" web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs docs/worklog/2026-04-21-phase-f-rollback-domain-copy-alignment.md`
- open_items: browser/manual smoke is still pending, but it does not block this copy-and-assertion closure