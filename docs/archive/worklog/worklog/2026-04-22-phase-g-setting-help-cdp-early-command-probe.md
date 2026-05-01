# Phase G setting help CDP early-command probe narrowing

## Scope

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G settings help real browser smoke closure
- Current stage: CDP early page-command probe narrowing
- Core goal: turn the remaining CDP smoke blocker into a precise, repeatable page-command matrix instead of a generic "Edge/CDP unstable" conclusion

## Mini Plan

- Current mainline: settings help real browser smoke
- Current stage: Phase G CDP narrowing
- Core task: add a lightweight first-command probe for page CDP commands and validate whether the blocker is attach/session plumbing or host Edge page-domain execution
- Supporting task: try a direct page websocket path via `/json/new` before falling back to `Target.createTarget`
- Verification: `node --check`, `--check-only`, and one real `self-start + cdp` smoke run
- Done when: the failing stage is narrowed to concrete commands with trace evidence, or the real smoke passes

## Changes Made

- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - added `probeEarlyPageCommands` to compare three earliest page actions: `Runtime.evaluate('1 + 1')`, `Page.navigate('about:blank#probe')`, and `Emulation.setDeviceMetricsOverride(1x1)`
  - emits `cdp-page-probe:summary` and `smoke-page:probe-complete` trace records so every self-start run leaves a compact command matrix in `cdp.trace.log`
  - fixed `sessionMethodFallbacks` so only true browser-level `Target.*` commands omit `sessionId`; page-domain commands now stay on the page session correctly
  - added a `/json/new` direct page-websocket attempt before falling back to `Target.createTarget` + `Target.attachToTarget`

## Verification

- Passed: `node --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
- Passed: `node scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only`
- Failed but narrowed further: `powershell -ExecutionPolicy Bypass -File scripts/verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp -KeepArtifacts`

## What The New Trace Proved

- The prior false blocker (`'<method>' wasn't found`) was caused by sending page-domain methods without `sessionId`; that part is now fixed.
- The direct page-websocket path via `/json/new` is available on this host when requested with the working method combination used in the script.
- Even on the direct `devtools/page/<target>` websocket, the same command matrix appears:
  - `Page.navigate(about:blank#probe)` succeeds quickly
  - `Runtime.evaluate('1 + 1')` times out
  - `Emulation.setDeviceMetricsOverride(1x1)` times out
- This means the remaining blocker is no longer specific to `Target.createTarget` or browser-session attach plumbing.

## Current Blockers

1. The host Edge environment still stalls on very-early page-domain commands that need runtime/emulation, even when using a direct page websocket.
2. `edge.stderr.log` still shows repeated access-denied, network sandbox, and crashpad failures under the temporary Edge profile directory.
3. Real browser smoke therefore still cannot complete the desktop/mobile settings assertions on this host.

## Next Entry

1. Add a small page bootstrap step before the probe (`Page.enable`, `Runtime.enable`, optionally lifecycle enable) to test whether this host only needs explicit domain initialization.
2. If bootstrap does not change the matrix, keep the new probe and direct-page path, and classify the remaining failure as host-specific Edge runtime/emulation stalling.
3. If a healthy external CDP session is available on another host profile, reuse it with the same script to separate environment issues from script logic.

## Final Status

- Result: success
- Need manual intervention: no