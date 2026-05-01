# 2026-04-18 Milestone 6 Validation Workflow Closure

## 1. Task Info

- Task: continue Milestone 6 acceptance closure by adding a repository-level validation workflow
- Date: 2026-04-18
- Phase: milestone 6 `Validation And Deployment`
- Goal: reduce dependence on the current local machine for Linux/Docker acceptance by codifying the validation chain in CI

## 2. Inputs Before Start

- canonical doc: [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
- recent runtime worklogs:
  - [2026-04-18-linux-backend-smoke-path.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-18-linux-backend-smoke-path.md)
  - [2026-04-18-backup-replace-prune-preview-closure.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-18-backup-replace-prune-preview-closure.md)
- existing CI baseline: [release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml)
- local constraints at implementation time:
  - Docker unavailable on the current machine
  - current Linux smoke path not runnable in the Windows / Git Bash mixed local thread

## 3. What Was Added

- added [validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml)
- the new validation workflow codifies the Milestone 6 acceptance chain into two jobs:
  - `backend-and-web`
    - `go test ./internal/op -count=1`
    - frontend `pnpm exec tsc --noEmit`
    - Linux backend smoke via [smoke-linux-backend.sh](/D:/GPT-codex/octopus_repo/scripts/smoke-linux-backend.sh)
  - `docker-source-build`
    - Docker source build via [Dockerfile.debian](/D:/GPT-codex/octopus_repo/scripts/dockerfiles/Dockerfile.debian)

## 4. Why This Was Chosen

- local Docker is not installed, so a real local Docker smoke could not be completed in this thread
- local Linux backend smoke is now available as a script, but current execution environment is not Linux / WSL
- moving the acceptance chain into CI lets the repository itself validate:
  - Linux runtime boot and HTTP smoke
  - frontend type integrity
  - Docker source-build integrity

## 5. Docs Updated

- [README.md](/D:/GPT-codex/octopus_repo/README.md) now mentions the repository validation workflow and what it covers

## 6. Validation Status

- local checks run in this thread:
  - `bash -n scripts/smoke-linux-backend.sh`
  - `go test ./internal/op -count=1`
  - `pnpm exec tsc --noEmit` in `web/`
- local blockers still present:
  - Docker command not installed locally
  - Linux smoke intentionally guarded against non-Linux / non-WSL execution in the current mixed shell

## 7. Sub-Agents Used

- `019da0c4-3af0-7a60-912d-02168520f5c2`, `gpt-5.4`
  - read-only review of the minimal CI/workflow shape to close the remaining milestone 6 acceptance chain
  - conclusion adopted: add a separate repository validation workflow rather than overloading existing release/build scripts

## 8. Remaining Gaps After This Round

- the new validation workflow has been added, but has not been observed running in GitHub Actions from this local thread
- browser-level smoke for key frontend surfaces is still pending
- Docker runtime smoke with `compose up` is still separate from pure source-build validation and remains open where environment permits

