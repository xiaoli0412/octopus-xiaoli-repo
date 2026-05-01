# 2026-04-25 Phase G Backup Helper Cleanup And Status Sync

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup detail follow-up
- Stage: backup helper-chain cleanup and status sync

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/README.zh-CN.md`
- automation memory: `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- `web/src/components/modules/setting/backup-logic.ts`
- `scripts/verify-backup-logic.mjs`

## Plan

- keep the same Phase G backup pool
- close the smallest remaining helper cleanup by removing the duplicate policy-warning pattern table and consolidating dispatch logic
- verify with repo-local `tsc --noEmit` and `verify-backup-logic.mjs`
- sync the remaining status through worklog even if the summary docs on this Windows host still hit an `apply_patch` transport issue

## What Changed

- `web/src/components/modules/setting/backup-logic.ts`
  - removed the now-redundant duplicate `BACKUP_POLICY_WARNING_PATTERNS` constant
  - introduced a typed `BackupPolicyWarningPattern` definition for the active localized warning matcher table
  - replaced the old cast-heavy branch chain with a shared `localizePolicyWarning()` helper so warning localization stays explicit and easier to maintain

## Verification

- Passed: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Passed: `node scripts/verify-backup-logic.mjs`

## Risks / Blockers

- The targeted code path is green, but `Backup.tsx` page-level contract cleanup is still pending as the next bounded backup-page task.
- Updating `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md` and `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md` hit a host-side `apply_patch` transport/parser issue on this Windows machine when patching those Chinese markdown files through the local wrapper. This did not block code progress or validation, but the stale blocker sentence in those summary docs still needs a follow-up patch.
- Vitest was not rerun because the existing host-level `spawn EPERM` issue remains unchanged.

## Result

- Outcome: success
- This round produced a real backup-helper code increment, kept the verification chain green, and left a narrower next entry point than before.

## Next Step

1. patch the two stale summary-doc blocker sentences once the markdown `apply_patch` transport issue is bypassed cleanly on this host
2. stay on the same backup mainline and take the next smallest page-level closure in `Backup.tsx`, preferably a visible contract cleanup rather than a broad refactor
3. keep `tsc --noEmit` and `node scripts/verify-backup-logic.mjs` as the stable local proof path here
