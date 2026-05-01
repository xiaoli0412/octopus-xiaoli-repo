# 2026-04-25 Phase G Backup Root Assertion And Status Sync

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page follow-up
- Stage: backup page selector contract and summary-doc sync

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- automation memory: `$CODEX_HOME/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `web/src/components/modules/setting/backup-logic.ts`
- `scripts/verify-backup-logic.mjs`

## Plan

- stay on the same Phase G backup mainline
- close one bounded page-level contract step by reusing the stable `backup-page` root selector in more than one backup test path
- sync stale backup blocker wording in status docs so the summary layer matches the now-green no-browser/typecheck proof
- verify with repo-local type/no-browser commands first on this host

## What Changed

- `web/src/components/modules/setting/Backup.test.tsx`
  - added `data-testid="backup-page"` root-anchor assertions to the dry-run/apply path and the replace-preview path so the page-level selector contract is no longer guarded by only one export-only test branch
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - updated the stale backup blocker sentence to reflect that `locale.tsx` is no longer the active blocker and the remaining work is page-detail cleanup plus browser-grade evidence
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - synced the same backup status wording so the frontend summary document matches the current local proof chain

## Verification

- Passed: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Passed: `node scripts/verify-backup-logic.mjs`
- Failed but known-host blocker: `node web/node_modules/vitest/vitest.mjs run web/src/components/modules/setting/Backup.test.tsx --config web/vitest.config.ts`
  - current host still fails during Vite/esbuild startup with `spawn EPERM`

## Risks / Blockers

- `Backup.test.tsx` assertions were tightened, but repo-local component execution still cannot be proven through Vitest on this host because `vite/esbuild` startup hits the known `spawn EPERM` environment blocker
- browser-grade backup evidence is still not closed; the remaining mainline gap is real browser proof for backup page layout/help interactions rather than more helper-only logic churn
- this round used direct PowerShell file replacement because the current host could not execute `apply_patch` successfully; future rounds on this host should keep edits similarly narrow if that tool remains unavailable

## Result

- Outcome: success
- This round produced a real code/doc increment, kept the local type/no-browser proof chain green, and reduced stale handoff drift in the summary docs

## Next Step

1. stay on the same Phase G backup mainline and pursue browser-grade evidence for the backup page using the stable `backup-page` root anchor and existing help/accordion selectors
2. if browser-grade evidence remains blocked by host tooling, choose the next smallest page-level backup contract cleanup in `Backup.tsx` instead of returning to helper-only churn
3. keep treating `spawn EPERM` in Vitest/Vite as a host blocker and prefer repo-local `tsc` plus no-browser scripts first