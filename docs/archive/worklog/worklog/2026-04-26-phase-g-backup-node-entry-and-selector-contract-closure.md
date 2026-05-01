# 2026-04-26 Phase G Backup Node Entry And Selector Contract Closure

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: backup verifier contract closure + Windows Node entry recovery hardening
- Timestamp: 2026-04-26T23:58:37+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/worklog/2026-04-26-phase-g-node-env-recovery-and-backup-parse-unblock.md`
- `docs/worklog/2026-04-26-phase-g-backup-import-remaining-migration-component-coverage.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`

## Plan

- stay on the same Phase G backup selector-contract mainline
- fix the remaining rollback-preview summary contract drift first
- keep write scope limited to the backup page, backup logic, and the two repo-local verification scripts
- restore a stable Windows PowerShell Node bootstrap path so backup verification no longer falls back to the broken Codex-bundled node executable

## What Changed

- `scripts/use-node-env.ps1`
  - hardened Node resolution so the bootstrapper skips the broken Codex-bundled `node.exe`
  - prefers an explicitly usable repo-local/system Node candidate and exports both `NODEEXE` and `NODE_BIN`
  - validates candidates with a real inline `node -e` probe instead of only checking file existence
- `web/src/components/modules/setting/Backup.tsx`
  - aligned `MetaGridCell` output with the current verification contract by rendering metadata labels as `label：value`
- `web/src/components/modules/setting/backup-logic.ts`
  - deduplicated generic credential-rebind signals so channel-key/api-key specific signals no longer also emit a redundant generic rebind line
- `web/src/components/modules/setting/Backup.test.tsx`
  - updated compatibility signal assertions to check signal-list content instead of fragile positional indexes
- `scripts/verify-backup-component.cjs`
  - aligned rollback-preview metadata expectations with the current `MetaGridCell` output
  - made map/replace compatibility-signal assertions resilient to ordering by checking the signal-list content
  - fixed replace-prune verifier flow so section selectors are only asserted after the accordion is opened
- `scripts/verify-backup-logic.mjs`
  - switched the script to `jiti`-based loading so it runs under the recovered Node 22 path without failing on direct `.ts` import
  - updated expectations for deduplicated rebind signals and `key`-bearing remaining-migration structures

## Verification

- `./scripts/use-node-env.ps1; node -v`
- `./scripts/use-node-env.ps1; node scripts/verify-backup-component.cjs`
- `./scripts/use-node-env.ps1; node scripts/verify-backup-logic.mjs`
- `./scripts/use-node-env.ps1; node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- `git diff --check -- scripts/use-node-env.ps1 web/src/components/modules/setting/Backup.tsx web/src/components/modules/setting/backup-logic.ts web/src/components/modules/setting/Backup.test.tsx scripts/verify-backup-component.cjs scripts/verify-backup-logic.mjs`

## Risks / Blockers

- `web/src/components/modules/setting/Backup.tsx` still carries the repo's existing LF/CRLF warning under `git diff --check`
- Vitest CLI on this machine still hits a host DNS blocker (`getaddrinfo EAI_FAIL localhost`), so this round used the repo-local Node verification scripts instead
- browser-grade backup-page evidence is still a separate remaining task

## Result

- Outcome: success
- This round restored a stable Windows Node verification path, closed the remaining backup selector-contract drift, and brought the direct backup verification chain back to green without widening into unrelated UI work

## Next Step

1. stay on the same Phase G backup mainline and choose the next smallest closure: browser-grade backup evidence or the lingering `Backup.tsx` line-ending hygiene warning
2. keep routing all Windows PowerShell Node-based checks through `./scripts/use-node-env.ps1`
3. avoid widening into unrelated settings UI cleanup while the backup page still has Phase G evidence tasks left
