## 1. Task Info

- Task: add explicit CDP page bootstrap strategy selection for settings-help browser smoke and classify attached-session host failures
- Date: 2026-04-23
- Stage: Phase G screenshot-first settings-help browser evidence closure
- Milestone: shrink the remaining browser-proof gap by separating `json-new` bootstrap behavior from explicit `attached-session` host behavior

## 2. Pre-flight Inputs

- Master plan aligned before coding (yes/no): yes
- Canonical sections: `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` section 9.6 / 14 / 16
- Workflow sections: `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` section 1.0 / 1.2 / 1.3
- Prior related worklogs:
  - `docs/worklog/2026-04-23-phase-g-setting-help-external-cdp-bootstrap-helper.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-local-service-bootstrap.md`
- Resources reviewed:
  - canonical plan
  - user-context requirements and workflow
  - detailed execution workflow
  - frontend mainline status
  - latest automation memory
  - current settings-help smoke wrapper and CDP node script
- Sub-agent usage: none (user explicitly required main-thread-only execution)

## 3. Hard Rules

- Stay on the Phase G settings-help browser-smoke mainline and do not drift into unrelated UI/backend work.
- Keep default `external + cdp` strict behavior intact; new strategy selection must only refine how the Node smoke opens the page.
- Limit changes to the smoke wrapper and CDP Node script; do not touch settings business logic.

## 4. Forbidden

- No weakening of the external/session-reuse contract.
- No silent behavior change that removes the current `auto` fallback semantics.
- No reverting unrelated dirty workspace changes.

## 5. Completion Criteria

- `scripts/verify-setting-help-browser-smoke.ps1` accepts a `-CdpPageBootstrapStrategy` switch and prints it in `check-only` output.
- `scripts/verify-setting-help-browser-smoke-cdp.mjs` receives the strategy and includes it in check-only/result/diagnostic output.
- A real `external + cdp + attached-session` run proves whether the host still blocks before page bootstrap completes.
- Wrapper trace-tail fallback can summarize the explicit attached-session timeout without manual trace reading.

## 6. Changes Applied

- File: `scripts/verify-setting-help-browser-smoke.ps1`
  - Added `-CdpPageBootstrapStrategy` with `auto / json-new / attached-session`.
  - Printed strategy in `check-only` summary.
  - Forwarded `OCTOPUS_UI_SMOKE_CDP_PAGE_BOOTSTRAP_STRATEGY` to the Node smoke for both `check-only` and real runs.
  - Extended trace-tail fallback summary to report `pageStrategy` and recognize the explicit attached-session timeout path.
- File: `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - Added normalized page bootstrap strategy parsing.
  - Added explicit `CdpPageOpenUnavailableError` classification.
  - Split `json-new` opening into a reusable helper and made `openCdpPage()` choose between `auto`, `json-new`, and `attached-session`.
  - Restricted fallback to attached-session so it only happens in `auto` mode.
  - Included `pageStrategy` in `check-only`, result JSON, and diagnostic payloads.

## 7. Verification

- Passed:
  - `node --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -BootstrapExternalCdpSession -SelfStartServices -CdpPageBootstrapStrategy attached-session -CdpUrl http://127.0.0.1:9233 -NodeSmokeTimeoutSeconds 15`
    - wrapper and Node check-only output now show `pageStrategy: attached-session`
  - `try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode external -Driver cdp -SelfStartServices -BootstrapExternalCdpSession -CdpPageBootstrapStrategy attached-session -CdpUrl 'http://127.0.0.1:9233' -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -NodeSmokeTimeoutSeconds 45 } catch { Write-Host $_.Exception.Message; exit 1 }`
    - reproduced explicit attached-session host failure and wrapper now reports:
      - `page_bootstrap_timeout_attached_session`
      - `CdpPageBootstrapPendingTimeout`
      - `page mode: attached-session`
      - `page strategy: attached-session`

## 8. What This Round Proved

- The new strategy switch is wired end-to-end from PowerShell wrapper to the Node CDP smoke.
- The earlier `json-new` timeout is no longer the only visible explanation path.
- Even when we bypass `json/new` and open an explicit attached-session page from the browser websocket, this host still stalls on `Page.enable`.
- That makes the remaining browser-proof blocker narrower and more credible: host-level attached-session bootstrap behavior inside Edge/CDP, not just `json/new` endpoint creation.

## 9. Risks and Compatibility

- Risk: low (smoke-script-only change, default `auto` behavior preserved).
- Compatibility impact: low (no backend/frontend business logic change).
- Remaining blocker: real browser smoke still cannot cross page bootstrap on this host, even with explicit attached-session selection.

## 10. Handoff

- Next run should stay on the same Phase G settings-help browser-smoke mainline.
- Immediate next task: compare whether a longer per-command timeout or a different pre-bootstrap command order changes the explicit attached-session `Page.enable` stall.
- If the same stall persists, record the blocker explicitly in frontend/mainline status as host-level Edge/CDP attached-session bootstrap behavior and then move to the next screenshot-first browser-evidence item in the same pool.

## 11. Final Status

- Result: success
- Need manual intervention: no
