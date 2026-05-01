# 2026-04-22 Phase G CC Switch Progressive Help Closure

## 1. Task Info

- Task: CC Switch progressive help and progress summary closure
- Date: 2026-04-22
- Stage: Phase G screenshot-first UI closure window
- Milestone: screenshot issue pool / CC Switch same-page progressive interaction closure

## 2. Pre-coding Input

- Master plan aligned before coding (yes/no): yes
- Canonical sections: docs/LLM-Gateway-Refactor-Plan.zh-CN.md 9.6, 14, 16
- Workflow sections: docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md 1.0, 1.2, 1.3, 1.4
- Previous related worklog: docs/worklog/2026-04-22-phase-g-channel-create-layered-guidance-closure.md
- Goal:
  - tighten CC Switch into a clearer four-step path
  - add a visible progress summary so users can see what is still missing
  - replace raw app button text with locale-driven labels
- Local resources checked:
  - docs/LLM-Gateway-Refactor-Plan.zh-CN.md
  - docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md
  - docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md
  - docs/CURRENT_STATUS_AND_PLAN.zh-CN.md
  - docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md
  - docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md
  - automation memory C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md
  - web/src/components/modules/navbar/DocModal.tsx
  - web/src/components/common/HelpHint.tsx
  - web/public/locale/en.json
  - web/public/locale/zh-Hans.json
  - web/public/locale/zh-Hant.json
- Sub-agent used: no
- Reason for no sub-agent: user explicitly requested main-thread execution only

## 3. Rules

- only touch CC Switch same-page guidance, help copy, and progressive structure
- do not change deep-link semantics or backend contract
- keep browser/manual smoke marked as pending until actually run

## 4. Acceptance

- CC Switch shows a clear four-step progress summary
- app switch buttons no longer show raw lowercase claude/codex labels
- advanced mapping stays folded behind accordion entry
- no-browser verification and frontend typecheck pass

## 5. Scope

- Frontend only
- Files:
  - web/src/components/modules/navbar/DocModal.tsx
  - web/public/locale/en.json
  - web/public/locale/zh-Hans.json
  - web/public/locale/zh-Hant.json
  - scripts/verify-ccswitch-flow.mjs

## 6. Steps

1. review DocModal and existing CC Switch locale/help coverage
2. add progress summary card and locale-driven app labels
3. add static verification script for this path

## 7. Verification

- node scripts/verify-ccswitch-flow.mjs
- D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json

## 8. Risks

- low risk: UI guidance and copy only
- browser/manual smoke still pending

## 9. Close-out

- build passed: yes
- test passed: yes
- manual smoke: not run
- pending pages:
  - DocModal CC Switch tab desktop
  - DocModal CC Switch tab 375px
  - Claude advanced mapping accordion
- next best task:
  1. browser/manual smoke for CC Switch tab
  2. circuit breaker progressive-help closure
  3. sync this result into frontend mainline summary
