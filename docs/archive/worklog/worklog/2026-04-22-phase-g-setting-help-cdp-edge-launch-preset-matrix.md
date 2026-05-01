# Phase G setting help CDP Edge launch preset matrix

## Scope

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G settings help real browser smoke closure
- Current stage: self-start Edge environment preset narrowing
- Core goal: convert the next host-side experiment into a reusable wrapper capability instead of another one-off command trial

## Mini Plan

- Current mainline: settings help real browser smoke
- Current stage: Phase G CDP narrowing
- Core task: add explicit Edge launch preset and profile strategy switches to the PowerShell smoke wrapper
- Supporting task: rerun the same `self-start + cdp` path with one stricter same-mainline experiment (`relaxed + workspace-fixed`) and compare the trace/stderr outcome with the existing baseline
- Verification: PowerShell script parse check, `-Mode check-only`, and one real `self-start + cdp` smoke run
- Done when: the wrapper can express alternate host-side launch experiments and the first preset trial leaves a trace-backed conclusion

## Local Resources Used

- [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
- [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)
- [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md)
- [ENV_READY_AND_NEXT_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md)
- [FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md)
- [2026-04-22-phase-g-setting-help-cdp-bootstrap-narrowing.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-phase-g-setting-help-cdp-bootstrap-narrowing.md)
- [2026-04-22-phase-g-setting-help-cdp-early-command-probe.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-phase-g-setting-help-cdp-early-command-probe.md)
- `$CODEX_HOME/automations/octopus-2/memory.md`
- [verify-setting-help-browser-smoke.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-setting-help-browser-smoke.ps1)
- [verify-setting-help-browser-smoke-cdp.mjs](/D:/GPT-codex/octopus_repo/scripts/verify-setting-help-browser-smoke-cdp.mjs)

## Why No Sub-Agent

- The user explicitly required main-thread execution only.
- This round touched one wrapper script and one same-mainline verification loop, so serial work kept the write scope tight.

## Hard Rules

- Stay on the same Phase G settings-help browser-smoke mainline.
- Do not expand into unrelated UI, backup, channel, or route-target work.
- Keep the smoke assertions fixed; only vary the host-side Edge launch environment.
- Leave a concrete next-entry recommendation based on real trace evidence.

## Forbidden

- Do not claim browser smoke passed unless the real `self-start + cdp` run completes.
- Do not overwrite unrelated workspace changes.
- Do not repeat the already-closed bootstrap-only experiment.

## Changes Made

- [verify-setting-help-browser-smoke.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-setting-help-browser-smoke.ps1)
  - added `-EdgeLaunchPreset` with `default` and `relaxed` options
  - added `-EdgeProfileStrategy` with `temp-random` and `workspace-fixed` options
  - extracted `Resolve-EdgeProfile` so self-start runs can reuse a stable profile tree when needed
  - extracted `Get-EdgeLaunchArguments` so Edge flag bundles become explicit, reviewable presets instead of inline one-off arguments
  - prints the active preset/profile choice in `check-only` and self-start runs for easier artifact comparison

## Verification

- Passed: `powershell -NoProfile -ExecutionPolicy Bypass -Command "& '.\scripts\verify-setting-help-browser-smoke.ps1' -Mode 'check-only' -Driver 'cdp' -EdgeLaunchPreset 'relaxed' -EdgeProfileStrategy 'workspace-fixed' ; exit $LASTEXITCODE"`
- Passed: PowerShell parse check for `scripts/verify-setting-help-browser-smoke.ps1`
- Failed but narrowed further: `powershell -NoProfile -ExecutionPolicy Bypass -Command "& .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp -CdpPort 9232 -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -KeepArtifacts *> $log"`

## What The New Experiment Proved

- The wrapper can now replay host-side Edge experiments without editing the script again.
- The first same-mainline environment trial (`relaxed + workspace-fixed`) still produces the exact same CDP command matrix as the previous baseline:
  - `Page.enable` times out
  - `Page.setLifecycleEventsEnabled(true)` times out
  - `Runtime.enable` times out
  - `Runtime.evaluate('1 + 1')` times out
  - `Page.navigate(about:blank#probe)` succeeds
  - `Emulation.setDeviceMetricsOverride(1x1)` times out
- Moving the profile from `%TEMP%` to a stable workspace path and using the more permissive flag bundle did not unblock page/runtime/emulation execution on this host.

## Current Blockers

1. Real browser smoke still fails on `Emulation.setDeviceMetricsOverride` after the new launch preset experiment.
2. Edge stderr still reports the same host-specific access-denied, network sandbox, and crashpad failures, now against the workspace-fixed profile path instead of `%TEMP%`.
3. The blocker therefore remains host-level rather than wrapper-argument-level for the tested preset/profile combination.

## Artifacts

- Failed self-start run artifact: `C:\Users\鏉庢槉妗怽AppData\Local\Temp\octopus-setting-help-smoke-3d0a587c196e416989c9bf41d542850a`
- Stable profile path used in this round: `D:\GPT-codex\octopus_repo\.tools\verify-setting-help-browser-smoke\edge-profile\relaxed`

## Next Same-Mainline Entry

1. Keep the new preset/profile switches as the canonical way to express future Edge self-start experiments.
2. Next, prefer one adjacent experiment that does not duplicate this result, for example:
   - reuse an already-open healthy external CDP session with the same Node script, or
   - add one more narrowly scoped preset aimed at disabling profile/network writes even harder, if and only if it changes a concrete stderr category.
3. If the next run still reproduces the same matrix, treat external CDP reuse as the higher-value path because wrapper-side flag tuning will have shown diminishing returns.

## Final Status

- Result: success
- Need manual intervention: no
