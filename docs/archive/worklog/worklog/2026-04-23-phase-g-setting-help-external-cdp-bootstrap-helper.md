## 1. Task Info

- Task: add an explicit external Edge CDP bootstrap helper path to the settings-help browser smoke wrapper
- Date: 2026-04-23
- Stage: Phase G screenshot-first settings-help browser evidence closure
- Milestone: shrink the external CDP proof gap without weakening the strict external/session-reuse contract

## 2. Pre-flight Inputs

- Master plan aligned before coding (yes/no): yes
- Canonical sections: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` section 9.6 / 14 / 16
- Workflow sections: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` section 1.0 / 1.2 / 1.3
- Prior related worklogs:
  - `docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-reuse-contract.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-local-service-bootstrap.md`
- Resources reviewed:
  - canonical plan
  - user-context requirements and workflow
  - detailed execution workflow
  - environment readiness plan
  - frontend mainline status
  - latest automation memory
  - smoke wrapper and `scripts/bootstrap-edge-cdp.ps1`
- Sub-agent usage: none (user explicitly required main-thread-only execution)

## 3. Hard Rules

- Keep default `external + cdp` strict: it must still fail when no external endpoint exists.
- Only add an explicit helper path; do not silently auto-launch Edge for external mode.
- Stay on the settings-help browser-smoke mainline and do not drift into unrelated UI or backend work.

## 4. Forbidden

- No weakening of the external/session-reuse contract.
- No changes to settings business logic or unrelated frontend modules.
- No reverting unrelated dirty workspace changes.

## 5. Completion Criteria

- `scripts/verify-setting-help-browser-smoke.ps1` supports an explicit switch that bootstraps a localhost external Edge CDP endpoint via `scripts/bootstrap-edge-cdp.ps1`.
- `check-only` shows this helper path clearly.
- `external + cdp + -BootstrapExternalCdpSession` reaches Node/CDP smoke instead of failing at `/json/version`.
- Plain `external + cdp` without the new switch still fails strictly at external endpoint preflight.

## 6. Changes Applied

- File: `scripts/verify-setting-help-browser-smoke.ps1`
- Added switch: `-BootstrapExternalCdpSession`.
- Added localhost-only helper plumbing:
  - `Test-LocalhostBaseUrl`
  - `Invoke-ExternalCdpBootstrapHelper`
- Wired external mode so the new helper is only used when explicitly requested.
- Kept default external behavior unchanged.
- Printed helper status in `check-only` output and explained that it reuses `scripts/bootstrap-edge-cdp.ps1`.
- Fixed helper stdout handling so the bootstrap script can still print its JSON summary while the wrapper reliably parses the saved summary file.

## 7. Verification

- Passed:
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -BootstrapExternalCdpSession -SelfStartServices -CdpUrl http://127.0.0.1:9233 -NodeSmokeTimeoutSeconds 15`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\bootstrap-edge-cdp.ps1 -Port 9233 -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -ReadyTimeoutSeconds 30 -StableReadySeconds 3 -OutputJsonPath .codex-tmp\edge-cdp-bootstrap.json`
    - result: `started-new-endpoint`, `/json/version` reachable on this host
  - `try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode external -Driver cdp -SelfStartServices -BootstrapExternalCdpSession -CdpUrl 'http://127.0.0.1:9233' -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -NodeSmokeTimeoutSeconds 15 -KeepArtifacts } catch { Write-Host $_.Exception.Message; exit 1 }`
    - reached external CDP preflight successfully and advanced into Node/CDP smoke
  - `try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode external -Driver cdp -SelfStartServices -BootstrapExternalCdpSession -CdpUrl 'http://127.0.0.1:9233' -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -NodeSmokeTimeoutSeconds 45 -KeepArtifacts } catch { Write-Host $_.Exception.Message; exit 1 }`
    - produced structured classification: `page_bootstrap_timeout_preempted / CdpPageBootstrapPendingTimeout / pageMode=json-new`
  - `try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode external -Driver cdp -SelfStartServices -CdpUrl 'http://127.0.0.1:9234' -NodeSmokeTimeoutSeconds 15 } catch { Write-Host $_.Exception.Message; exit 1 }`
    - confirmed default strict behavior remains: missing endpoint still fails at external preflight

## 8. What This Round Proved

- The host no longer blocks progress at the older 鈥渕anual external Edge endpoint unreachable鈥?stage when we use the explicit helper path.
- External mode can now bootstrap a localhost Edge remote debugging endpoint in a controlled, opt-in way while preserving the default strict contract.
- Once the endpoint exists, external mode reproduces the same `json-new -> Page.enable timeout` classification seen in self-start runs.
- The remaining blocker is therefore narrower and more credible: host-level Edge/CDP page bootstrap behavior, not simple endpoint availability.

## 9. Risks and Compatibility

- Risk: low (wrapper-only change, default behavior preserved).
- Compatibility impact: low (no backend/frontend business logic change).
- Remaining blocker: real browser smoke still does not pass; the confirmed external path now fails at the same page bootstrap stage as self-start.

## 10. Handoff

- Next run should stay on the same Phase G mainline.
- Immediate next task: compare whether a longer Node timeout or a different page-opening strategy changes the external `json-new` bootstrap classification.
- If the same classification persists, record the blocker explicitly as host-level Edge/CDP page bootstrap behavior and use it as the current browser-evidence gate status instead of continuing to relitigate endpoint startup.

## 11. Final Status

- Result: success
- Need manual intervention: no
