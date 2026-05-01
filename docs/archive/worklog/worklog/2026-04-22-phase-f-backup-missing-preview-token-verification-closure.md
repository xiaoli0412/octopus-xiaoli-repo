# 2026-04-22 Phase F Backup Missing Preview Token Verification Closure

## 1. Task Info

- Task name: Backup dry-run missing preview token verification closure
- Date: 2026-04-22
- Current phase: Phase F / backup-import-rollback frontend closure
- Milestone: Milestone 6 validation and deployment

## 2. Inputs Before Coding

- Canonical plan: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` sections 11.5, 13, 14, 16
- Workflow: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` Phase F and task template section
- Previous related worklog: `docs/worklog/2026-04-22-phase-f-import-mode-change-pending-apply-reset-verification.md`
- Goal: add same-page evidence for the branch where dry-run finishes without a `preview_token`, and make sure pending apply does not survive that result
- Local resources reviewed: automation memory, canonical plan, workflow, frontend mainline status doc, Phase F worklog chain, `Backup.tsx`, `Backup.test.tsx`, `scripts/verify-backup-component.cjs`, `scripts/verify-backup-component.setting-mock.cjs`, `scripts/verify-backup-logic.mjs`
- Subagents used: no, per user instruction to stay on the main thread only

## 3. Plan

1. Recheck the `executeImport` missing-preview-token branch in `Backup.tsx`
2. Add a focused Vitest draft case in `Backup.test.tsx`
3. Mirror the same branch in the no-spawn component verification path
4. Run the existing Backup verification commands and record results

## 4. Changes Made

- `web/src/components/modules/setting/Backup.test.tsx`
  - Added `does not keep pending apply when dry-run omits the preview token`
  - Asserted that the dry-run result still renders rows summary but does not render `Apply This Dry-Run`
  - Asserted that the rerun warning toast is emitted
  - Asserted that `backup.import.dryRunSuccess` is not emitted for this branch
- `scripts/verify-backup-component.setting-mock.cjs`
  - Added a file-name-based dry-run mock branch for `snapshot-missing-preview-token.json` that omits `preview_token`
- `scripts/verify-backup-component.cjs`
  - Added `verifyDryRunMissingPreviewTokenDoesNotLeavePendingApply`
  - Wired it into the main no-spawn verification flow

## 5. Validation

- Passed: `node scripts/verify-backup-component.cjs`
- Passed: `node scripts/verify-backup-logic.mjs`
- Passed: `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`

## 6. Risks And Blockers

- Risk: low; this round only added verification coverage and a matching mock branch
- Blocker: the repository default `apply_patch` wrapper still points at a dead temp path on this host; a working local fallback apply_patch entry was used instead
- Manual smoke: browser/manual Backup smoke is still pending

## 7. Next Suggestions

- Priority 1: run browser/manual Backup smoke when a browser session is available
- Priority 2: if browser/manual smoke is still unavailable, cover the next smallest Backup apply-guard edge case such as `missing_request`
