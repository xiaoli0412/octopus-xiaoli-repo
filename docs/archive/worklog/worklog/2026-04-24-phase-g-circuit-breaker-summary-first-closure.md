# 2026-04-24 Phase G Circuit Breaker Summary-First Closure

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first UI closure / settings circuit breaker summary-first compression`
- Summary: stayed in the same Phase G screenshot-first settings pool after the `ModelProbe` closure and moved one adjacent high-occupancy card forward. `CircuitBreaker` now follows the same `short copy + summary first + details only when needed` rule more strictly: the default path keeps one short sentence, summary cards no longer repeat helper paragraphs in the card body, and the recovery flow moved behind its own accordion so the detailed three-step explanation no longer occupies the default settings viewport.
- Verification:
  - `node scripts/verify-circuit-breaker-help.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-locale-consistency.mjs`
- Changed files:
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/ja.json`
  - `scripts/verify-circuit-breaker-help.mjs`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-circuit-breaker-summary-first-closure.md`
- Result:
  - `CircuitBreaker` 默认路径从“标题单行 + 多块正文解释”收口为“短句主说明 + 摘要优先”，让用户先看当前推荐区间和运行时状态，再决定是否细调。
  - 四张运行时摘要卡片去掉了卡片正文里的重复 helper 段落，改成标题旁 `HelpHint`，减少默认 DOM 和文字密度，同时保留可解释性。
  - 恢复逻辑不再默认铺开在正文区，而是收进单独的 `Recovery Flow` 折叠层；默认只显示一句“恢复自动处理”的短说明，三步细节仅在排障时展开。
  - 高级熔断策略标题下补了一行短说明，并把四语 locale 与 no-browser 静态护栏同步到新结构，避免“组件已压缩但验证仍盯旧 DOM”的回退风险。
- Blocker:
  - 无新增代码阻塞。
  - 浏览器级 settings screenshot / `375px` 证据仍受当前宿主 `spawn EPERM` 与 Edge/CDP bootstrap 问题影响，因此本轮仍以 no-browser 结构闭环为准。
- Next:
  1. stay in the same Phase G settings pool and apply the same `short copy + summary first + deferred detail` pattern to `DynamicRouting`, because it is still one of the remaining high-occupancy settings cards in the same performance problem set.
  2. after that, inspect `Backup` for the next same-pool closure around default copy length, summary hierarchy, and whether deeper detail blocks can move behind explicit expansion.
  3. once host browser spawning is stable again, rerun the blocked settings/browser smoke tasks to confirm `CircuitBreaker` and `ModelProbe` still behave well at `375px` and under real hover/focus interaction.
