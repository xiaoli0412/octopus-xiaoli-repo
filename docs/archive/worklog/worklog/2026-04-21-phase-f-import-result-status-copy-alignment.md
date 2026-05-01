# 2026-04-21 Phase F Import Result Status Copy Alignment

## Plan

- current_mainline: Phase F backup/import/rollback frontend closure
- current_stage: 11.5.4 import result panel consistency
- core_task: align the top import-result status wording with the actual returned state
- support_task: leave the next same-mainline entry point in worklog and automation memory
- completion_criteria: badge, title, and top status sentence all use the same dry-run vs applied semantics and `tsc --noEmit` passes

## Changes

- updated `web/src/components/modules/setting/Backup.tsx` so the import result badge now says `Dry-run report` or `Import applied`
- updated `web/src/components/modules/setting/Backup.tsx` so the top status sentence now says `Dry-run report only analyzes...` or `Import applied with the selected mode...`
- updated `web/src/components/modules/setting/Backup.tsx` so the result header title now says `Dry-run report` or `Import applied`

## Validation

- web TypeScript no-emit check passed for the current Backup.tsx changes
- confirmed Backup.tsx now renders Dry-run report and Import applied text in the result panel

## Risks And Next

- no browser or manual smoke ran in this round; this is compile-safety plus wording alignment only
- one same-mainline follow-up still exists in the derived helper copy that says Import result emitted warnings
- next round should stay on Phase F and align that remaining helper copy or switch to same-mainline status-doc sync if Backup.tsx stops yielding safe micro-closures
