# 2026-04-26 Phase G Node Env Recovery And Backup Parse Unblock

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract tightening
- Stage: Node verification environment recovery + backup component parse/type unblock
- Timestamp: 2026-04-26T17:10:00+08:00

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- recent backup-mainline worklogs:
  - `docs/worklog/2026-04-26-phase-g-backup-history-size-selector-closure.md`
  - `docs/worklog/2026-04-26-phase-g-backup-rollback-preview-summary-selector-closure.md`
  - `docs/worklog/2026-04-25-phase-g-backup-import-vs-history-remaining-migration-selector-split.md`
- automation memory: `.codex-home-link/automations/octopus-2/memory.md`

## Plan

- treat the reported Node blocker as this round's primary task
- verify whether Node itself is broken or whether the failure actually comes from a bad executable path / current user code parse errors
- restore a repo-local Node entrypoint that always uses the working system Node installation
- unblock the backup page from JSX/type parse failures so Node-based verification can return to real code-level output

## What Changed

- `scripts/use-node-env.ps1`
  - added a lightweight repo-local Node environment bootstrapper
  - normalizes `APPDATA` when the host session does not provide it
  - pins the active session to the working system Node install at `C:\Users\李昊桐\AppData\Local\Programs\nodejs\node.exe`
  - exports `NODEEXE` and aliases `node` / `node.exe` so later PowerShell verification commands stop drifting into the broken Codex-bundled executable path
- `web/src/components/modules/setting/Backup.tsx`
  - repaired JSX structure breaks that were previously causing `tsc` and `verify-backup-component.cjs` to fail in parse stage
  - restored the backup page to a compilable state under the current `web/tsconfig.json`
  - aligned post-import health summary English labels with the existing verifier contract: `Health-check targets` and `Passed`
  - split import/history remaining-migration section ordering so rollback history once again starts with the rollback-tooling section expected by current tests/verifier
- `web/src/components/modules/setting/backup-logic.ts`
  - added stable `key` fields to `RemainingMigrationToolingSection` entries so section/item selectors no longer depend on fragile positional assumptions
- `scripts/verify-backup-component.setting-mock.cjs`
  - aligned the mock post-import health summary with the current verifier expectation (`targets: 3`, `passed: 2`) so backup verification now exercises real component behavior instead of failing on stale mock data

## Verification

- Confirmed working system Node directly:
  - `C:\Users\李昊桐\AppData\Local\Programs\nodejs\node.exe -v`
  - `C:\Users\李昊桐\AppData\Local\Programs\nodejs\node.exe -e "console.log('node-ok')"`
- Confirmed repo-local Node bootstrap works:
  - `.\scripts\use-node-env.ps1; node -v; Write-Host ('NODEEXE=' + $env:NODEEXE)`
  - result pins the session to `C:\Users\李昊桐\AppData\Local\Programs\nodejs\node.exe` and reports `v22.19.0`
- TypeScript verification now passes through the Node environment gate and compiles successfully:
  - `.\scripts\use-node-env.ps1; node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Backup verifier also now runs through Node and reaches repo-level assertion output instead of host CSPRNG startup failure:
  - `.\scripts\use-node-env.ps1; node scripts/verify-backup-component.cjs`
  - current remaining failure is a backup selector/content contract assertion: `Expected rollback preview scope summary selector content`
- Diff hygiene:
  - `git diff --check -- scripts/use-node-env.ps1 web/src/components/modules/setting/backup-logic.ts scripts/verify-backup-component.setting-mock.cjs web/src/components/modules/setting/Backup.tsx`
  - still reports the repo's existing line-ending warning on `Backup.tsx`

## Risks / Blockers

- The previous blocker description in memory was too broad: Node itself is not blocked on this host when we force the working system install. The real remaining issue is backup verifier contract drift inside current repo code.
- `rg` from the Codex-bundled WindowsApps path is still unreliable on this host; for shell searches, prefer explicit working binaries or PowerShell file reads when necessary.
- `Backup.tsx` still carries the repo's existing LF/CRLF warning under `git diff --check`.
- One backup verifier assertion still fails on rollback preview scope summary content, so this round restored Node execution and compile verification but did not fully green the backup verifier.

## Result

- Outcome: success
- This round removed the host-level Node verification blocker, introduced a reusable repo-local Node bootstrap entrypoint, and pushed the remaining issue down to a normal backup-page contract mismatch.

## Next Step

1. stay on the same Phase G backup selector-contract mainline
2. close the remaining rollback-preview scope summary mismatch in `Backup.tsx` / `verify-backup-component.cjs`
3. route future Windows PowerShell Node-based checks through `.\scripts\use-node-env.ps1` before invoking `node ...`
