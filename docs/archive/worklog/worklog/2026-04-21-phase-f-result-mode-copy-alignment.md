# 2026-04-21 Phase F Result Mode Copy Alignment

## Plan

- current_mainline: Phase F backup/import/rollback frontend closure
- current_stage: 11.5.4 import and rollback preview consistency
- core_task: remove the temporary `selectedImportMode` alias and align the result panel copy with `resultModeLabel/resultModeOption`
- support_task: rewrite this worklog in ASCII because the Windows patch path turned Chinese text into question marks
- completion_criteria: `selectedImportMode` is gone, result copy uses derived labels, and front-end `tsc --noEmit` passes

## Inputs

- canonical: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` chapter 9 plus the active Phase F backup/import closure
- workflow: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- previous_worklog: `docs/worklog/2026-04-21-phase-f-manifest-boolean-state-and-smoke-status-sync.md`
- local_resources: automation memory, `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`, `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`, `docs/worklog/README.zh-CN.md`, previous Phase F worklog, `web/src/components/modules/setting/Backup.tsx`, `git status --short`
- skills_and_context: `using-superpowers`, `brainstorming`, current thread context
- subagents: none
- why_no_subagents: the user explicitly asked to stay on the main thread and this was a tiny single-file closure

## Guardrails

- touch only `web/src/components/modules/setting/Backup.tsx`, this worklog, and automation memory
- do not widen scope beyond Phase F / 11.5.4
- do not change backend contracts, data shape, or route behavior
- finish with targeted front-end validation

## Changes

- removed the temporary `selectedImportMode` alias from `web/src/components/modules/setting/Backup.tsx`
- kept the existing `resultModeOption` lookup and `resultModeLabel` derivation
- changed the visible result copy from hard-coded `Report mode` wording to `{resultModeLabel}: {resultModeOption.label}` with the existing description text
- confirmed the stable Windows patch path for future runs: `D:\gol1\node.exe` -> `%USERPROFILE%\.codex\.sandbox-bin\codex.exe --codex-run-as-apply-patch`

## Validation

- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- `rg -n "selectedImportMode|resultModeLabel|resultModeOption|Report mode|Current preview mode|Current apply mode" web/src/components/modules/setting/Backup.tsx`

## Compatibility And Risks

- no data, schema, or API behavior changed in this round
- no browser/manual smoke rerun in this round
- the default Windows `apply_patch` wrappers are still flaky; the sandbox-bin codex path was the reliable fallback
- the Chinese status doc could be read in UTF-8, but line-based patching against its localized text stayed unreliable in this Windows environment, so this run kept the authoritative closure in code plus worklog/memory

## Handoff

- next_mainline: stay on Phase F / 11.5.4
- next_best_task: keep squeezing tiny `Backup.tsx` consistency closures or, if that stops paying off, move to broader Phase F status-doc sync
- smoke_status: not rerun in this round
- next_task_ready: yes

## Run Result

- status: success
- files_changed: `web/src/components/modules/setting/Backup.tsx`, `docs/worklog/2026-04-21-phase-f-result-mode-copy-alignment.md`
- commands_run: `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`; `rg -n "selectedImportMode|resultModeLabel|resultModeOption|Report mode|Current preview mode|Current apply mode" web/src/components/modules/setting/Backup.tsx`
- open_items: broader Phase F status doc sync is still pending because localized patch matching remained flaky on this host
