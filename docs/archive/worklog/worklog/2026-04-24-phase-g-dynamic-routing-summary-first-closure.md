# 2026-04-24 Phase G Dynamic Routing Summary-First Closure

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first UI closure / settings dynamic-routing summary-first compression`
- Summary: stayed in the same Phase G screenshot-first settings pool and closed the next adjacent no-browser settings compression task after `ModelProbe` and `CircuitBreaker`. `DynamicRouting` now keeps the default path focused on the switch, the core four-card runtime summary, the mode selector, and the budget snapshot. The heavier mode snapshot, key-mix breakdown, and raw summary metadata moved behind an explicit `summary details` accordion so the default settings viewport no longer renders the full summary stack at once.
- Verification:
  - `node scripts/verify-dynamic-routing-help.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-locale-consistency.mjs`
  - attempted but blocked: `node web/node_modules/vitest/vitest.mjs run web/src/components/modules/setting/DynamicRouting.test.tsx --config web/vitest.config.ts`
  - attempted but blocked: `node web/node_modules/vitest/vitest.mjs run web/src/components/modules/setting/DynamicRouting.test.tsx --config web/vitest.config.ts --configLoader runner --pool threads`
- Changed files:
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/DynamicRouting.test.tsx`
  - `scripts/verify-dynamic-routing-help.mjs`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/ja.json`
  - `docs/worklog/2026-04-24-phase-g-dynamic-routing-summary-first-closure.md`
- Result:
  - the default intro block now follows the same `short title + one compact sentence` pattern already used in the adjacent settings cards
  - the runtime tuning switch keeps one short explanation line instead of only a title row, so the default path is still understandable after the summary stack was compressed
  - the summary section now shows only status, enabled channels, failover groups, and current decision by default; mode snapshot, key mix, and raw summary metadata are deferred behind a dedicated accordion
  - the budget snapshot and advanced budget trigger now both keep a short description line, so users can distinguish `quick scan` from `advanced tuning` without reading long helper paragraphs
  - four locales and the no-browser guard were updated together, and the component test was realigned to the new structure even though this host still cannot execute Vitest
- Blocker:
  - host-level `Node child_process.spawn(...)=EPERM` still blocks `vitest/vite/esbuild` startup on this machine; the retry with `--configLoader runner` still failed in Vite config startup on Windows safe real-path handling, so this remains an environment blocker rather than a code regression
  - browser-grade settings evidence remains separately blocked by the known host `spawn EPERM` and Edge/CDP issues, so this round stays on the no-browser closure path
- Next:
  1. stay in the same Phase G settings pool and apply the same `summary first + shorter default copy + deferred detail` pattern to `Backup`, because it is now the clearest remaining same-pool high-occupancy card
  2. once host spawning is stable again, rerun the blocked `DynamicRouting.test.tsx` and the settings/browser smoke tasks to confirm the lighter structure still behaves well under real `375px` and hover/focus interaction
  3. keep treating host-level `spawn EPERM` and Edge/CDP bootstrap as classified blockers instead of spending more rounds re-diagnosing the same environment issue
