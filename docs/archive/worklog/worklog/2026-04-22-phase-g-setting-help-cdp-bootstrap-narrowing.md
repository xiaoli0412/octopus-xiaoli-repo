# Phase G setting help CDP bootstrap narrowing

## Scope

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G settings help real browser smoke closure
- Current stage: CDP page-domain bootstrap narrowing
- Core goal: verify whether explicit page-domain bootstrap can unlock the remaining settings-help Edge CDP smoke blocker on this host

## Mini Plan

- Current mainline: settings help real browser smoke
- Current stage: Phase G CDP narrowing
- Core task: add explicit page-domain bootstrap before the earliest command probe in the CDP smoke script
- Supporting task: rerun the same `self-start + cdp` smoke path and compare bootstrap results with the existing early-command matrix
- Verification: `node --check`, `--check-only`, and one real `self-start + cdp` smoke run
- Done when: bootstrap either unblocks the smoke or proves with trace evidence that the host still stalls on page/runtime/emulation domains

## Local Resources Used

- [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
- [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)
- [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md)
- [FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md)
- [2026-04-22-phase-g-setting-help-cdp-early-command-probe.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-phase-g-setting-help-cdp-early-command-probe.md)
- [2026-04-22-phase-g-setting-help-cdp-page-session-narrowing.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-phase-g-setting-help-cdp-page-session-narrowing.md)
- `$CODEX_HOME/automations/octopus-2/memory.md`
- [verify-setting-help-browser-smoke-cdp.mjs](/D:/GPT-codex/octopus_repo/scripts/verify-setting-help-browser-smoke-cdp.mjs)
- [verify-setting-help-browser-smoke.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-setting-help-browser-smoke.ps1)

## Why No Sub-Agent

- The user explicitly required the main thread to avoid spawning sub-agents.
- This task touched one tightly scoped script plus one verification loop, so keeping it in the main thread was lower risk.

## Hard Rules

- Stay on the current Phase G settings-help browser-smoke mainline.
- Do not expand into unrelated UI, pricing, backup, or route-target refactors.
- Keep the smoke scope fixed to the same settings cards and help-hint assertions.
- Leave a concrete trace-backed conclusion for the next run.

## Forbidden

- Do not claim browser smoke passed unless the real `self-start + cdp` run succeeds.
- Do not revert unrelated uncommitted work in the dirty workspace.
- Do not broaden this run into wrapper-port, target-discovery, or general Edge cleanup work once the bootstrap result is known.

## Changes Made

- [verify-setting-help-browser-smoke-cdp.mjs](/D:/GPT-codex/octopus_repo/scripts/verify-setting-help-browser-smoke-cdp.mjs)
  - added `bootstrapCdpPageDomain()` before the early command probe
  - explicitly tests `Page.enable`, `Page.setLifecycleEventsEnabled(true)`, and `Runtime.enable`
  - writes `cdp-page-bootstrap:summary` and `smoke-page:bootstrap-complete` trace entries
  - includes `pageBootstrap` in the final JSON payload for successful runs

## Verification

- Passed: `node --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
- Passed: `node scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only`
- Failed but narrowed further: `powershell -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp -KeepArtifacts`

## What The New Trace Proved

- The direct page websocket path still opens correctly via `/json/new`.
- The explicit bootstrap commands all time out on this host:
  - `Page.enable`
  - `Page.setLifecycleEventsEnabled(true)`
  - `Runtime.enable`
- The earlier matrix remains unchanged after bootstrap:
  - `Page.navigate(about:blank#probe)` succeeds quickly
  - `Runtime.evaluate('1 + 1')` still times out
  - `Emulation.setDeviceMetricsOverride(1x1)` still times out
- This means the remaining blocker is not “missing bootstrap”; it is a host-specific stall around Edge page/runtime/emulation domain execution under the current temporary profile and policy environment.

## Current Blockers

1. Real browser smoke still fails on `Emulation.setDeviceMetricsOverride` after bootstrap.
2. Edge stderr still reports access-denied, network sandbox, and crashpad launch errors under the temporary profile directory.
3. The current host therefore still cannot complete the desktop/mobile settings assertions through this self-start Edge CDP path.

## Next Same-Mainline Entry

1. Treat the current script as the canonical narrowed reproducer for this host-specific Edge/CDP failure.
2. On the next run, prefer one adjacent same-mainline experiment instead of redoing bootstrap:
   - try a different self-start browser flag/profile strategy, or
   - reuse a healthy external CDP session to separate host policy from script logic.
3. Keep the trace-backed matrix as the baseline for judging whether the environment changes anything.

## Final Status

- Result: success
- Need manual intervention: no
- Artifact: `C:\Users\李昊桐\AppData\Local\Temp\octopus-setting-help-smoke-da6dc3e0fc8f4ea9af46ffafed685aff\cdp.trace.log`
