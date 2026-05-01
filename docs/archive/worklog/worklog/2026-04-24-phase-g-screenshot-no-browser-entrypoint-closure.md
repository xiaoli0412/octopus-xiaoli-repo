# 2026-04-24 Phase G Screenshot No-Browser Entrypoint Closure

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first UI closure / no-browser verification entrypoint consolidation`
- Summary: stayed in the same screenshot-first Phase G pool and folded the scattered no-browser guards for `Home / Channel / Group / Model / Route Target Overrides / CC Switch + settings no-browser` into one `web/package.json` entrypoint: `test:screenshot-no-browser`. The validation/release workflows now call that single entrypoint, and the workflow/status docs have been synchronized so the repo no longer keeps a split contract where settings had a unified entrypoint but the rest of the screenshot-first no-browser chain still lived as raw command lists.
- Local resources used:
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-settings-no-browser-entrypoint-closure.md`
  - `docs/worklog/2026-04-24-phase-g-settings-no-browser-green-restoration.md`
  - automation memory `octopus-2`
  - `web/package.json`
  - `.github/workflows/validation.yaml`
  - `.github/workflows/release.yaml`
  - screenshot-first no-browser scripts under `scripts/verify-*.mjs|cjs`
- Verification:
  - `rg -n "test:screenshot-no-browser|test:settings-no-browser|validation / release|Frontend no-browser verification" web/package.json .github/workflows/validation.yaml .github/workflows/release.yaml docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `$env:COREPACK_HOME='D:\GPT-codex\octopus_repo\.tools\corepack'; D:\gol1\corepack.cmd pnpm --dir web run test:screenshot-no-browser`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `git diff --check -- web/package.json .github/workflows/validation.yaml .github/workflows/release.yaml docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- Changed files:
  - `web/package.json`
  - `.github/workflows/validation.yaml`
  - `.github/workflows/release.yaml`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-screenshot-no-browser-entrypoint-closure.md`
- Result:
  - screenshot-first no-browser validation now has one stable package entrypoint instead of a mixed set of raw workflow command chains
  - validation/release now cover `locale consistency + Home / Channel / Group / Model / Route Target Overrides + settings no-browser + CC Switch` in one pass
  - the first implementation of the new entrypoint exposed a host-side nested-`pnpm` portability issue; it was fixed in the same round by rewriting the entrypoint to a pure `node ... && ...` command chain, and the final entrypoint now runs green on this machine
- Blocker:
  - `scripts/runtime-win.ps1 -Action status` still hits `Get-CimInstance Win32_Process` access denial on this host, so runtime-status probing remains a host permission blocker rather than a code-side problem
  - browser-grade `375px / hover / focus` evidence outside the already-closed settings/help and homepage/model pools still depends on the known host browser/CDP constraints
- Next:
  1. stay in the same Phase G screenshot-first pool and decide whether the next best closure is browser-evidence recovery for `Channel / Group / CC Switch`, or another no-browser/doc cleanup that reduces drift without reopening stable UI code
  2. if host browser constraints remain, keep preferring same-pool validation-chain/documentation closures that make future Phase G passes cheaper
  3. if runtime status probing is needed again, either harden `runtime-win.ps1` for low-privilege fallback or continue recording the current permission blocker instead of treating it as task failure
