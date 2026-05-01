# Phase G setting help CDP page session narrowing

## Scope

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G settings help real browser smoke closure
- Current stage: CDP self-start stability narrowing
- Core goal: move the real smoke blocker from `Page.enable` to a later, more explainable step and improve traceability and cleanup

## Changes Made

- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - added `OCTOPUS_UI_SMOKE_CDP_TRACE_FILE` support for browser/page websocket, target lifecycle, command send/result/timeout, and final error trace
  - changed page creation flow to `Target.createTarget` plus direct `devtools/page/<target>` websocket connection
  - removed the hard dependency on `Page.enable` and `Runtime.enable` before later page commands
- `scripts/verify-setting-help-browser-smoke.ps1`
  - injects and prints the CDP trace file path
  - moves the temporary Edge profile under the smoke temp directory
  - adds `Stop-ProcessTree` so failed runs clean up spawned processes instead of hanging until the outer timeout

## Verification

- Passed: `D:\gol1\node.exe --check scripts\verify-setting-help-browser-smoke-cdp.mjs`
- Passed: `D:\gol1\node.exe scripts\verify-help-hint-accessible.mjs`
- Passed: `powershell -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp`
- Failed but narrowed: `powershell -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp -KeepArtifacts`

## Result

- Previous blocker: `Page.enable`
- Current blocker: `Emulation.setDeviceMetricsOverride`
- Meaning: browser websocket, target creation, page websocket discovery, page websocket connection, and failure cleanup now pass; the blocker has moved to the first page command stage
- Cleanup result: the wrapper now exits in about 25 seconds with a concrete error instead of hanging until the outer shell timeout

## Current Blockers

1. This host still times out on very-early Edge headless page commands.
2. `edge.stderr.log` still shows repeated access denied, network sandbox, and crashpad related errors.
3. The main blocker now looks host-environment-specific to Edge/CDP page execution, not port resolution, target discovery, or wrapper cleanup.

## Next Entry

1. Add a lighter first-command probe in `verify-setting-help-browser-smoke-cdp.mjs` to compare `Runtime.evaluate('1+1')`, `Page.navigate`, and `Emulation.setDeviceMetricsOverride`.
2. Try `json/new` or more aggressive Edge startup flags to check whether this is specific to the `Target.createTarget` path.
3. If this host keeps failing at the page-command stage, keep the new wrapper/trace path and switch to the same-mainline fallback of reusing an already healthy external CDP session.

## Final Status

- Result: success
- Need manual intervention: no
