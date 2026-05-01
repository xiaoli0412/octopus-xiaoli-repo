# 2026-04-18 Backup Replace-Prune Preview Closure

## 1. Task Info

- Task: close backup/import `replace` prune preview parity with actual apply behavior
- Date: 2026-04-18
- Phase: main MD phase 6 `Backup / Import / Migration Adaptation`
- Milestone focus: full backup/import usability and dry-run fidelity

## 2. Inputs Before Start

- Canonical doc: [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
- Prior backup/import code: [backup.go](/D:/GPT-codex/octopus_repo/internal/op/backup.go), [backup_extra.go](/D:/GPT-codex/octopus_repo/internal/op/backup_extra.go), [backup.go model](/D:/GPT-codex/octopus_repo/internal/model/backup.go)
- Existing UI: [Backup.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx)
- Existing tests: [backup_test.go](/D:/GPT-codex/octopus_repo/internal/op/backup_test.go)
- Local resources used: main MD, current worktree context, existing backup tests, AGENTS rules, previous worklog chain
- Sub-agents enabled: yes
- Sub-agent model: `gpt-5.4`

## 3. Goal

- make dry-run and rollback preview explicitly show which top-level objects `replace` will prune
- keep preview semantics aligned with actual apply semantics, including selective import scopes and secret-bearing snapshot rules
- expose the new preview in the backup UI without creating a fake or lossy summary

## 4. What Was Completed

- backend compatibility report now exposes structured replace-prune data for:
  - `channels`
  - `groups`
  - `settings`
  - `llm_infos`
  - `api_keys`
- backend also keeps per-domain summary counters for the same prune categories
- dry-run compatibility generation now respects `ImportScopes`, so disabled scopes do not falsely report prune impact
- API key prune preview now follows real apply semantics:
  - only active for `replace`
  - only active when API key scope is enabled
  - only active when `Manifest.ContainsSecrets == true`
  - keep-set is based on filtered importable plaintext API keys, not raw snapshot rows
- rollback preview now also carries the same replace-prune structure, so users can see which current objects disappear after rollback-to-snapshot
- frontend backup UI now renders a dedicated replace-prune section for both:
  - import dry-run / apply result
  - rollback preview

## 5. Files Changed

- [internal/model/backup.go](/D:/GPT-codex/octopus_repo/internal/model/backup.go)
- [internal/op/backup.go](/D:/GPT-codex/octopus_repo/internal/op/backup.go)
- [internal/op/backup_extra.go](/D:/GPT-codex/octopus_repo/internal/op/backup_extra.go)
- [internal/op/backup_test.go](/D:/GPT-codex/octopus_repo/internal/op/backup_test.go)
- [web/src/api/endpoints/setting.ts](/D:/GPT-codex/octopus_repo/web/src/api/endpoints/setting.ts)
- [web/src/components/modules/setting/Backup.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/Backup.tsx)

## 6. Tests And Verification

- `go test ./internal/op -count=1`
- `pnpm exec tsc --noEmit` in `web/`
- frontend sub-agent also ran:
  - `pnpm exec eslint src/api/endpoints/setting.ts src/components/modules/setting/Backup.tsx`

## 7. Sub-Agents Used

- `019da090-11fa-7903-a3eb-2e997e2e97e5`, `gpt-5.4`
  - owned frontend-only work in `web/src/api/endpoints/setting.ts` and `web/src/components/modules/setting/Backup.tsx`
  - result adopted and integrated into main-thread backend closure
- `019da09d-5d31-79e2-a54f-8c9ddee561b5`, `gpt-5.4`
  - read-only audit of replace preview semantics
  - key conclusion adopted: prune preview must honor `ImportScopes`, `ContainsSecrets`, and filtered API key keep-set semantics

## 8. Remaining Mainline Gaps After This Round

- backup/import preview is now closer to actual `replace`, but browser-level smoke for the backup page is still pending
- Linux real server runtime acceptance and Docker runtime smoke are still not closed in this round
- richer structured compare UX for rollback/import preview is still a later UI step beyond this closure

