# 2026-04-23 Phase G Model Toolbar Search Width Closure

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G` screenshot-first UI closure
- Core task: close the model-page toolbar search-width readability gap without changing search logic
- Context used:
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-model-search-placeholder-contract-closure.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
- Changes made:
  - `web/src/components/modules/toolbar/index.tsx`
    added page-specific expanded search widths for `channel`, `group`, and `model`
    added stable `data-slot` and `data-page` hooks for the expanded search shell
  - `scripts/verify-llm-price-boundary.mjs`
    added assertions for the new width contract
    added a guard that the old fixed `w-20` input implementation is gone
- Verification:
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node scripts/verify-locale-consistency.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Risk: low; no API, locale text, or filtering logic changed
- Blocker: host `Edge/CDP` bootstrap is still blocking browser-level `375px / hover / focus` evidence
- Sub-agents used: none; user asked to keep execution on the main thread
- Next best task:
  1. collect browser-level model/settings `375px` and `hover / focus` evidence if CDP becomes usable
  2. if browser evidence is still blocked, pick the next concrete no-browser entry/layout readability gap in the same pool
