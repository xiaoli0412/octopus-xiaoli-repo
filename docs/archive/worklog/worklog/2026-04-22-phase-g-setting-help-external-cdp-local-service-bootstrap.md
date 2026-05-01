# 2026-04-22 Phase G setting-help external CDP local service bootstrap

## 1. Task Info

- Task: add an optional local backend/frontend bootstrap path for `external + cdp` smoke runs
- Date: 2026-04-22
- Stage: Phase G image-priority settings-help browser-smoke closure
- Milestone: keep narrowing the remaining settings-help browser proof gap

## 2. Pre-flight Inputs

- Master plan aligned before coding (yes/no): yes
- Canonical sections: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` section 9.6 / 14 / 16
- Workflow sections: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` section 1.0 / 1.2 / 1.3
- Prior related worklogs:
  - `docs/worklog/2026-04-22-phase-g-setting-help-wrapper-timeout-classification.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-preflight-closure.md`
- Resources reviewed:
  - canonical plan
  - user-context requirements and workflow
  - detailed execution workflow
  - environment readiness plan
  - frontend mainline status
  - latest automation memory
  - wrapper and CDP smoke scripts
- Sub-agent usage: none (user explicitly requested main-thread-only execution)

## 3. Hard Rules

- Only touch the smoke wrapper path; do not change settings business logic.
- Keep `external + cdp` as strict external-CDP reuse, with no hidden Edge auto-launch.
- Leave a concrete, repeatable command entry for the next-run comparison.

## 4. Forbidden

- No expansion into backup/import, channel, multi-key, group, or `CC Switch` tracks.
- No reverting unrelated dirty workspace changes.
- Do not claim browser smoke pass while the external CDP endpoint is still unavailable.

## 5. Completion Criteria

- `scripts/verify-setting-help-browser-smoke.ps1` supports `-SelfStartServices`.
- `external + cdp + -SelfStartServices` starts local backend/frontend first, then validates the external CDP endpoint.
- `check-only` clearly shows whether external local service bootstrap is enabled.
- One run proves the expected order: local services first, external CDP preflight second.

## 6. Changes Applied

- File: `scripts/verify-setting-help-browser-smoke.ps1`
- Added new switch: `-SelfStartServices`.
- Added unified gate: `bootstrapLocalServices = self-start || (external && SelfStartServices)`.
- Kept CDP auto-port/session reuse logic scoped to `self-start` only.
- Updated `check-only` summary to show the external local-service bootstrap option.
- Moved external CDP preflight to run after service readiness checks, so the `-SelfStartServices` path is truly usable.

## 7. Verification

- Passed:
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -NodeSmokeTimeoutSeconds 15`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -SelfStartServices -NodeSmokeTimeoutSeconds 15`
  - `try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode external -Driver cdp -SelfStartServices -CdpUrl 'http://127.0.0.1:9233' -NodeSmokeTimeoutSeconds 15 -KeepArtifacts } catch { $_.Exception.Message }`
    - expected order observed: `Starting local backend and frontend for external CDP smoke` then `Preflighting external Edge CDP endpoint`
    - final error still expected: external CDP endpoint not reachable at `http://127.0.0.1:9233`
- Additional contrast run:
  - `try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode external -Driver cdp -CdpUrl 'http://127.0.0.1:9233' -NodeSmokeTimeoutSeconds 15 } catch { $_.Exception.Message }`
    - failed at `Verifying external backend and frontend` with `Timed out waiting for http://127.0.0.1:18081/healthz`
    - this confirms `-SelfStartServices` meaningfully changes run order
- Still blocked externally:
  - direct attempts to expose `http://127.0.0.1:9233/json/version` from a manually launched Edge session remained unreachable on this host

## 8. Risks and Blockers

- Risk: low (script-only behavior change).
- Compatibility impact: low (no backend/frontend business logic change).
- Current blocker: on this host, attempts to launch an external Edge session with `--remote-debugging-port=9233` still did not expose `/json/version`; external session proof is not closed yet.

## 9. Handoff

- Next run should stay on this exact mainline and solve the external Edge `/json/version` availability first.
- Once the endpoint is reachable, run `external + cdp + -SelfStartServices` immediately and compare bootstrap classification against current self-start results.
- If the same bootstrap/runtime stage fails externally, record it as host-level Edge/CDP behavior, not self-start-profile-only behavior.

## 10. Final Status

- Result: success
- Need manual intervention: no
