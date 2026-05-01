# 2026-04-18 Linux Backend Smoke Path

## 1. Task Info

- Task: close the missing Linux backend smoke path for the mainline runtime acceptance chain
- Date: 2026-04-18
- Phase: main MD phase 6 and milestone 6 follow-up
- Goal: add a reproducible Linux-local smoke path aligned with the existing Windows smoke flow

## 2. Inputs Before Start

- main MD: [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
- existing Windows smoke: [smoke-win-backend.ps1](/D:/GPT-codex/octopus_repo/scripts/smoke-win-backend.ps1)
- existing Linux launcher: [dev-linux.sh](/D:/GPT-codex/octopus_repo/scripts/dev-linux.sh)
- existing build entry points: [build.sh](/D:/GPT-codex/octopus_repo/scripts/build.sh)
- local resources used: current runtime scripts, README, server startup code, sub-agent read-only review
- sub-agents enabled: yes
- sub-agent model: `gpt-5.4`

## 3. What Was Added

- added [smoke-linux-backend.sh](/D:/GPT-codex/octopus_repo/scripts/smoke-linux-backend.sh)
- the Linux smoke path now mirrors the proven Windows-local smoke shape:
  - build a temporary local smoke binary
  - create a temporary SQLite DB and temporary JSON config
  - start a local Python mock upstream
  - start Octopus with `start --config ...`
  - verify `healthz`
  - drive:
    - `/api/v1/user/login`
    - `/api/v1/channel/create`
    - `/api/v1/group/create`
    - `/api/v1/apikey/create`
    - `/v1/chat/completions`
- failure diagnostics were also added so the script prints recent server/mock logs on readiness failure

## 4. Why This Shape Was Chosen

- kept smoke separate from [dev-linux.sh](/D:/GPT-codex/octopus_repo/scripts/dev-linux.sh) because that script is a dev launcher, not a one-shot acceptance script
- kept smoke separate from [build.sh](/D:/GPT-codex/octopus_repo/scripts/build.sh) because that file is a build/release pipeline, not a runtime orchestration tool
- matched the Windows smoke flow to reduce drift between Windows-local and Linux-local acceptance paths

## 5. Docs Updated

- [README.md](/D:/GPT-codex/octopus_repo/README.md) now includes a Linux backend smoke section next to the existing Linux development/build entries

## 6. Validation Status

- `bash -n scripts/smoke-linux-backend.sh`
- targeted script syntax check passed in the current environment
- because the current local thread is a Windows / Git Bash mixed environment rather than a Linux host or WSL runtime, the script now explicitly guards against non-Linux execution and points users to the Windows smoke path instead of producing a misleading readiness failure
- full Linux runtime smoke still depends on executing the script in a Linux-capable environment or compatible WSL/runtime path

## 7. Sub-Agents Used

- `019da0b1-0c38-77c1-91c0-d6c86c192775`, `gpt-5.4`
  - read-only script-placement and runtime-flow review
  - conclusion adopted: keep Linux smoke as a standalone script beside the Windows smoke path, and reuse the same minimal acceptance chain

## 8. Remaining Gaps After This Round

- the repository now has a Linux smoke path, but a real Linux environment execution result is still needed for final acceptance
- Docker runtime smoke remains a separate pending item
- browser-level smoke for the key frontend surfaces is still pending
