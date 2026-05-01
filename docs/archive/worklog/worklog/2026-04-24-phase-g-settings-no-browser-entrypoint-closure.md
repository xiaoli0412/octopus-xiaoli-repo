# 2026-04-24 Phase G Settings No-Browser Entrypoint Closure

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first UI closure / settings no-browser verification entrypoint consolidation`
- Summary: stayed in the same Phase G settings pool and closed a small but recurring workflow gap: the repo already had green local no-browser guards for `Backup / DynamicRouting / CircuitBreaker / ModelProbe / HelpHint / SettingInfo`, but the actual entrypoint was still split across package scripts, raw `node` commands, workflow YAML, and status docs. This round introduced a single `web/package.json` entrypoint, switched both GitHub workflows to that entrypoint, and synchronized the front-end status/workflow docs so the settings pool no longer has a scattered validation contract.
- Local resources used:
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-dynamic-routing-summary-first-closure.md`
  - automation memory `octopus-2`
  - `web/package.json`
  - `.github/workflows/validation.yaml`
  - `.github/workflows/release.yaml`
- Verification:
  - `pnpm --dir web run test:settings-no-browser`
  - `pnpm --dir web exec tsc --noEmit`
  - `rg -n "test:settings-no-browser" web/package.json .github/workflows/validation.yaml .github/workflows/release.yaml docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- Changed files:
  - `web/package.json`
  - `.github/workflows/validation.yaml`
  - `.github/workflows/release.yaml`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-settings-no-browser-entrypoint-closure.md`
- Result:
  - settings no-browser validation now has one stable entrypoint instead of six scattered raw script invocations
  - `DynamicRouting` is no longer tracked as a missing no-browser validation-chain item; its remaining blocker is browser-grade evidence only
  - the workflow doc now points to the consolidated entrypoint, while the remaining malformed legacy command lines are isolated to one old tail block for later cleanup
- Blocker:
  - host-level `Node child_process.spawn(...)=EPERM` still blocks `vitest/vite/esbuild` startup on this machine, so the settings pool remains on the no-browser path for executable verification
  - browser-grade `375px / hover / focus` evidence for settings cards still depends on the known host `spawn EPERM` and Edge/CDP bootstrap blockers
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` still contains one legacy malformed tail block with control characters in the `11.4` list; it no longer blocks execution because the corrected command set is already present above, but it should be cleaned in a later doc-only pass
- Next:
  1. stay in the same Phase G settings pool and decide whether the next highest-value closure is `Backup` browser/no-browser alignment or a doc-only cleanup of the remaining malformed workflow tail block
  2. once host spawning is stable again, rerun the blocked `DynamicRouting.test.tsx` / settings browser smoke tasks so the unified no-browser entrypoint gains matching browser-grade evidence
  3. separately review whether non-settings guards such as `verify-home-layout` / `verify-channel-presentation` / `verify-llm-price-boundary` should also be folded into reusable package-level entrypoints in a later round
