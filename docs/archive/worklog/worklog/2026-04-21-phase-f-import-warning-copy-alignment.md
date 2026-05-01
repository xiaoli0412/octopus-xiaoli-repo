# 2026-04-21 Phase F Import Warning Copy Alignment

## Plan

- current_mainline: Phase F backup/import/rollback frontend closure
- current_stage: 11.5.4 import result panel consistency
- core_task: align the derived import warning helper copy with the real dry-run vs apply state
- completion_criteria: risk-signal warning copy follows `Dry-run report` or `Import applied`, and `tsc --noEmit` passes

## Changes

- updated `web/src/components/modules/setting/Backup.tsx` so `buildCompatibilitySignalItems` accepts `importWarningsLabel`
- updated the import result risk-signal call to pass `Dry-run report` or `Import applied` from `resultIsDryRun`
- removed the remaining hardcoded applied-result warning wording drift in the shared helper copy

## Validation

- web TypeScript no-emit check passed for the current Backup.tsx change
- targeted search confirmed the helper now uses importWarningsLabel and the import result call now passes Dry-run report or Import applied

## Risks And Next

- no browser or manual smoke ran in this round; this is compile-safety plus copy alignment only
- Windows default apply_patch wrapper still failed with Access is denied; the working fallback remained %USERPROFILE%/.codex/.sandbox-bin/codex.exe --codex-run-as-apply-patch
- next_best_task: either close one more tiny Backup.tsx wording helper drift, or switch to broader Phase F status-doc sync if same-file yield drops
