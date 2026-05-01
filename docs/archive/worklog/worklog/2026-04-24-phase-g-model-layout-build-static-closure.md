# 2026-04-24 Phase G Model Layout Build Static Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / model layout browser evidence`
- Summary: stayed on the same Phase G browser-evidence line and closed the missing static-build verification leg after the model layout browser smoke. The host shell still could not call `pnpm` directly and Next 16 build workers were failing on this Windows host with `spawn EPERM`, so the build chain was narrowed into a repo-owned wrapper that runs `tsc --noEmit` first, then calls Next build with a host-specific worker compatibility patch, and finally syncs `web/out` into `static/out`.
- Verification:
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node ./.codex-tmp/corepack/v1/pnpm/10.33.1/dist/pnpm.cjs --dir web run build:static`
- Changed files:
  - `scripts/build-web-static.mjs`
  - `web/next.config.ts`
  - `web/package.json`
  - `web/src/components/modules/home/total.tsx`
  - `docs/worklog/2026-04-24-phase-g-model-layout-build-static-closure.md`
- Result:
  - `build:static` now runs through a repo-owned wrapper instead of relying on a host PATH-level `pnpm` contract.
  - The wrapper preserves a real TypeScript preflight, skips only Next's internal type-check worker for that build invocation, patches the build worker path so this Windows host no longer fails on `spawn EPERM` / thread clone issues, and then syncs `web/out` into `static/out`.
  - While unblocking this chain, a real syntax regression in `home/total.tsx` was also fixed so `tsc --noEmit` and the static build could both complete.
- Blocker:
  - `baseline-browser-mapping` still prints stale-data warnings during build, but they are non-blocking and do not prevent static export or sync.
- Next:
  1. Return to the remaining Phase G screenshot-first browser evidence pool, with homepage browser evidence as the best adjacent closure.
  2. Keep the build wrapper path as the stable Windows verification entry unless upstream Next or host policy changes remove the worker incompatibility.
