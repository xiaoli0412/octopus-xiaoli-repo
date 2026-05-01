# 2026-04-25 Phase G Backup Post-Import Health Selector Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page follow-up
- Stage: post-import validation selector contract tightening

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- automation memory: `$CODEX_HOME/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`
- `scripts/runtime-win.ps1`

## Plan

- stay on the same Phase G backup mainline
- close one more page-level selector-contract step inside the existing post-import validation panel
- add stable anchors for the post-import validation summary grid and health summary block so later browser smoke does not depend on localized headings
- tighten the repo-local verifier and component assertions around those selectors
- verify with runtime status, repo-local typecheck, and backup component verification on this host

## What Changed

- `web/src/components/modules/setting/Backup.tsx`
  - added `data-testid="backup-post-import-validation-summary-grid"` to the post-import validation summary grid
  - added `data-testid="backup-post-import-health-summary"` to the post-import health summary block
- `web/src/components/modules/setting/Backup.test.tsx`
  - added direct selector assertions for the post-import validation summary grid and health summary in the apply path
  - kept the existing panel-level assertion so the apply path now guards the panel plus its two internal anchor points
- `scripts/verify-backup-component.cjs`
  - replaced the apply-path text wait with selector-based waiting on `backup-post-import-validation-panel`
  - added direct selector assertions for `backup-post-import-validation-summary-grid` and `backup-post-import-health-summary`

## Verification

- Passed: `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- Passed: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Passed: `node scripts/verify-backup-component.cjs`

## Risks / Blockers

- `Backup.test.tsx` still keeps two localized text assertions in the same apply path, so the contract is tighter now but not fully selector-only yet.
- Browser-grade backup evidence is still the next real gap; this round only tightened the page-level anchor contract around the post-import validation area.
- `Backup.tsx` remains a large in-flight file in this workspace, so future rounds should continue keeping the write scope narrow.
- Vitest/Vite/browser smoke remain host-sensitive on this machine, so the stable local proof path is still `runtime status + tsc --noEmit + verify-backup-component.cjs`.

## Result

- Outcome: success
- This round produced a real code increment, moved the post-import validation health area further toward selector-first verification, and left a smaller next step on the same backup page mainline.

## Next Step

1. stay on the same Phase G backup mainline and use `backup-post-import-validation-panel`, `backup-post-import-validation-summary-grid`, and `backup-post-import-health-summary` as stable browser-smoke anchors.
2. if browser evidence remains host-blocked, choose the next smallest `Backup.tsx` page-level cleanup around `backup-compatibility-toggle` or finish removing the remaining localized-text dependency from `Backup.test.tsx` apply assertions.
3. keep `runtime-win.ps1 -Action status`, `tsc --noEmit`, and `verify-backup-component.cjs` as the stable local proof path on this machine.
