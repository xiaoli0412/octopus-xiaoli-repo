# 2026-04-25 Phase G Backup Post-Validation Summary Sync

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page contract follow-up
- Stage: backup post-import validation summary and replace-prune state sync
- Timestamp: 2026-04-25T22:43:51+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/backup-logic.ts`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup page mainline
- pick one bounded page-level cleanup inside `Backup.tsx`
- align the post-import validation summary grid with the existing shared `backup-logic.ts` summary shape
- keep replace-prune detail state from leaking across file changes / new previews / apply
- verify on the existing repo-local proof path if the host allows Node and nested PowerShell to start

## Chosen Task

- Core task: sync the `Backup.tsx` post-import validation summary grid with the current `getPostImportValidationSummary()` output
- Companion task: reset `showReplacePruneDetails` when the selected file changes, when a new import preview starts, and after applying the same import

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added localized labels for the summary fields that already exist in `backup-logic.ts` but were not surfaced in the page grid:
    - `disabled channels`
    - `stale items removed`
    - `price-rule warnings`
    - `alias mappings`
    - `alias warnings`
  - expanded `backup-post-import-validation-summary-grid` so the page now renders all current post-import summary counters instead of only the older four-field subset
  - reset `showReplacePruneDetails` when a new file is selected, when a new import preview result is received, and after `handleApplySameImport()` succeeds, so replace-mode expanded state does not leak into the next preview/apply cycle

## Verification

- Text-level verification performed:
  - `git diff -- 'web/src/components/modules/setting/Backup.tsx'`
  - targeted file reads around the updated label block, import handlers, and post-import summary grid
- Attempted runtime / repo-local verification:
  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-backup-component.cjs`
- Host blockers encountered during verification:
  - nested PowerShell hit `8009001d`
  - every attempted Node startup in this sandbox hit `Assertion failed: ncrypto::CSPRNG(nullptr, 0)` before script execution

## Risks / Blockers

- The code increment is small and local to `Backup.tsx`, but repo-local execution proof could not be completed because the host currently blocks both nested PowerShell and Node startup
- `Backup.test.tsx` and `scripts/verify-backup-component.cjs` were not updated in this round, so the newly surfaced post-import counters are currently a page-level consistency fix rather than a newly asserted selector contract
- The file remains large and in-flight in this workspace; future rounds should continue choosing one bounded backup-page closure at a time

## Result

- Outcome: partial success
- This round produced a real `Backup.tsx` code increment on the active Phase G mainline, narrowed one page-vs-helper inconsistency, and left a direct next entry even though host-level verification was blocked

## Next Step

1. rerun `tsc --noEmit` and `verify-backup-component.cjs` once the host Node/CSPRNG issue is no longer blocking execution
2. stay on the same `Backup.tsx` mainline and add the smallest matching verifier/component assertions for the newly surfaced post-import summary fields
3. if execution remains host-blocked, continue with another bounded backup-page contract cleanup that can still be text-verified without widening scope
