# 2026-04-18 Milestone 6 Compose Runtime And Frontend Smoke Closure

## 1. Task Info

- Task: continue Milestone 6 acceptance closure by adding compose-runtime smoke coverage and tightening frontend-shell smoke evidence
- Date: 2026-04-18
- Phase: milestone 6 `Validation And Deployment`
- Goal: move repository validation from `build only` closer to `real runtime acceptance`, while keeping frontend smoke expectations realistic and MD-aligned

## 2. Inputs Before Start

- canonical doc: [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
- existing runtime assets:
  - [scripts/smoke-linux-backend.sh](/D:/GPT-codex/octopus_repo/scripts/smoke-linux-backend.sh)
  - [scripts/smoke-win-backend.ps1](/D:/GPT-codex/octopus_repo/scripts/smoke-win-backend.ps1)
  - [docker-compose.yml](/D:/GPT-codex/octopus_repo/docker-compose.yml)
  - [validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml)
- existing frontend status doc:
  - [FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md)
- sub-agent conclusions reused in this round:
  - main MD audit: runtime verification evidence remained weaker than build/test entrypoints
  - frontend audit: best near-term split is `automatic shell/runtime smoke + manual checklist for complex UI`

## 3. What Was Added Or Tightened

- `docker-compose.yml` is now parameterized for smoke-friendly runtime overrides:
  - `OCTOPUS_PORT`
  - `OCTOPUS_DATA_DIR`
  - `OCTOPUS_CONTAINER_NAME`
  - `host.docker.internal` host-gateway mapping
- added [scripts/smoke-docker-compose.sh](/D:/GPT-codex/octopus_repo/scripts/smoke-docker-compose.sh)
  - builds and starts the repository compose stack
  - waits for `/healthz`
  - verifies `/` returns the expected Octopus shell
  - verifies `/manifest.json` returns the expected app name
  - tears the compose stack down automatically
- tightened both local backend smoke scripts to verify frontend shell assets as part of the same acceptance chain:
  - [scripts/smoke-linux-backend.sh](/D:/GPT-codex/octopus_repo/scripts/smoke-linux-backend.sh)
  - [scripts/smoke-win-backend.ps1](/D:/GPT-codex/octopus_repo/scripts/smoke-win-backend.ps1)
  - both now validate:
    - `/`
    - `/manifest.json`
    - `/healthz`
    - minimal management/gateway flow
- updated [validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml)
  - added `workflow_dispatch`
  - replaced pure Docker source-build job with compose runtime smoke job
- updated docs:
  - [README.md](/D:/GPT-codex/octopus_repo/README.md)
  - [README_zh.md](/D:/GPT-codex/octopus_repo/README_zh.md)
  - [FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md)

## 4. Why This Shape Was Chosen

- the main MD milestone 6 explicitly calls for `Docker build / compose up` style validation, so stopping at source-build only left a visible acceptance gap
- current frontend stack in this thread does not yet have a stable browser automation lane, so the highest-value near-term move was:
  - upgrade existing smoke scripts to prove the frontend shell is actually served
  - document which UI areas still require manual checklist coverage
- this keeps the repo closer to real Linux/server and Docker runtime behavior without pretending that complex backup/log UI flows are fully browser-smoked already

## 5. Validation Status

- local validation still possible in this thread:
  - smoke scripts and CI YAML can be syntax/readability checked
  - frontend-shell verification logic now exists in both local smoke paths
- still blocked locally in this thread:
  - actual Docker runtime execution, because Docker availability depends on the current machine
  - full browser automation for the five main frontend surfaces

## 6. Remaining Gaps After This Round

- GitHub Actions run evidence is still needed to prove the updated validation workflow succeeds remotely
- browser-level smoke for homepage/channel/group/backup/log is still split as:
  - minimal automatic runtime/shell coverage: present
  - complex interaction/manual checklist: still manual
- milestone 6 backup/import still has broader UX gaps beyond this runtime closure, including richer compare/mapping tooling

## 7. Sub-Agents Used

- `019da0cd-3360-7ba3-9d48-d5c7e717b6e2`, `gpt-5.4`
  - read-only main MD milestone audit
  - conclusion adopted: runtime verification evidence remained one of the highest-value remaining milestone 6 gaps
- `019da0cd-33d4-70c1-b698-8382c144ef3a`, `gpt-5.4`
  - read-only frontend smoke audit
  - conclusion adopted: split acceptance into automatic shell/runtime smoke and manual checklist for complex UI surfaces
