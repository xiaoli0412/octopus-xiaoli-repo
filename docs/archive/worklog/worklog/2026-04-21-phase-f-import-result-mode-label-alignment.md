# 2026-04-21 Phase F Import Result Mode Label Alignment

- phase: Phase F
- milestone: 11.5.4
- trigger_time: 2026-04-21 00:02 +08:00
- focus: keep the import result mode label in `web/src/components/modules/setting/Backup.tsx` aligned with returned result-state wording instead of fixed apply wording

## Plan

- current_mainline: Phase F backup/import/rollback frontend closure
- current_stage: 11.5.4 import result consistency closure
- candidate_tasks:
  - remove one more result-state wording drift inside `Backup.tsx`
  - keep scope inside `Backup.tsx` only
  - verify with front-end type check
- core_task: stop the import result detail line from always saying apply semantics when the returned result is still a dry-run report
- support_task: keep the current result-mode source (`effectiveMode`) compile-safe while avoiding wider JSX reshuffles
- completion_criteria: result detail wording no longer implies apply-only semantics for dry-run reports and `tsc --noEmit` stays green

## Changes

- updated [Backup.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx) so the result-mode derivation now lives in `resultModeOption` and remains tied to `effectiveMode`
- updated [Backup.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx) so the visible detail label now uses neutral `Report mode` wording instead of fixed `Current apply mode`
- kept the rendered mode description sourced from the returned result mode path through `selectedImportMode = resultModeOption` to avoid widening this round

## Validation

- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- `rg -n "resultModeOption|selectedImportMode|resultModeLabel|Report mode" web/src/components/modules/setting/Backup.tsx`

## Findings

- the remaining real drift in this block was wording, not state reset; the form-reset chain still clears stale result data correctly before any new import request
- Windows `apply_patch` remained fragile in this session for multiline JSX hunks that contain escaped quotes; the stable progress path was to land only tiny hunks through direct `codex.exe --codex-run-as-apply-patch`
- the final visible label change was intentionally narrowed to `Report mode` because the last fully dynamic JSX hunk kept triggering the same Windows patch parser issue

## Risks

- no browser/manual smoke ran in this round, so this is still compile-safety plus wording alignment only
- `Backup.tsx` remains a very large active file; further edits should keep using tiny hunks on the same mainline
- one small follow-up remains: replace the still-neutral `Report mode` label with the already-derived `resultModeLabel` once the local Windows patch path accepts that exact JSX hunk cleanly

## Next

- stay on `Phase F / 11.5.4`
- next best task: finish the last label swap from `Report mode` to `{resultModeLabel}` in [Backup.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx)
- fallback next task: if the Windows patch parser still rejects that JSX hunk, sync this wording status into broader Phase F status docs and then return to the next tiny `Backup.tsx` consistency closure
