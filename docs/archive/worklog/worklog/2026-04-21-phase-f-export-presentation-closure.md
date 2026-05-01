# 2026-04-21 Phase F Export Presentation Closure

## Plan

- current_mainline: Phase F backup/import/rollback frontend closure
- current_stage: 11.5.4 export snapshot presentation consistency
- core_task: centralize the backup export snapshot copy into a shared helper and keep the UI/test contract aligned
- support_task: extend the source-level and component-level verification paths so full vs redacted export states stay observable
- completion_criteria: export helper covers full and redacted states, Backup page renders from the helper, and the targeted verification commands pass

## Changes

- added `getExportSnapshotPresentation()` in `web/src/components/modules/setting/backup-logic.ts`
- rewired `web/src/components/modules/setting/Backup.tsx` to render export summary, warning copy, badges, and toggle label from the shared helper
- added export-state assertions to `web/src/components/modules/setting/backup-logic.test.ts`
- extended `scripts/verify-backup-logic.mjs` to cover full and redacted export presentation states
- added a UI-level export-state smoke in `web/src/components/modules/setting/Backup.test.tsx`
- taught `scripts/verify-backup-component.cjs` to use the select mock through `@/components/ui/select` and to assert the map-mode preview rows without duplicate-text ambiguity

## Validation

- `node --experimental-strip-types scripts/verify-backup-logic.mjs`
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- `node scripts/verify-backup-component.cjs`

## Compatibility And Risks

- no backend contract or data shape changed
- `backup-logic.ts` now owns both export states, so future wording changes only need one helper update
- the component verification still depends on the local select mock, but the real UI contract remains unchanged

## Handoff

- next_mainline: stay on Phase F / 11.5.4 until backup/import/rollback smoke evidence is stronger
- next_best_task: run a minimal browser/manual smoke for export toggle, dry-run, Apply Same Import, and rollback preview if the environment allows it
- smoke_status: source-level and component-level verification passed in this round
- next_task_ready: yes

## Run Result

- status: success
- files_changed: `web/src/components/modules/setting/backup-logic.ts`, `web/src/components/modules/setting/Backup.tsx`, `web/src/components/modules/setting/backup-logic.test.ts`, `web/src/components/modules/setting/Backup.test.tsx`, `scripts/verify-backup-logic.mjs`, `scripts/verify-backup-component.cjs`
- commands_run: `node --experimental-strip-types scripts/verify-backup-logic.mjs`; `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`; `node scripts/verify-backup-component.cjs`; `node -e "const fs=require('node:fs'); JSON.parse(fs.readFileSync('web/package.json','utf8')); console.log('package-json-ok')"`; `git status --short`
- open_items: browser/manual smoke is still the next useful evidence layer for the backup flow
