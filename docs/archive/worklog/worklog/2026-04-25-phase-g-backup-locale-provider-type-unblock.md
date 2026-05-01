# 2026-04-25 Phase G Backup Locale Provider Type Unblock

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page verification recovery
- Stage: backup page contract unblock and locale-provider type recovery

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- automation memory: `C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md`
- recent backup worklogs from `2026-04-24` and `2026-04-25`

## Plan

- stay on the same Phase G backup pool
- stop treating backup helper formatting as the only frontier and repair the page-level TypeScript blocker
- keep verification bounded to repo-local frontend checks that are stable on this host

## What Changed

- `web/src/provider/locale.tsx`
  - added the missing `ja` locale import back into the provider map
  - relaxed the provider-side message object typing from `typeof zh_HansMessages` to a runtime-safe dictionary shape so `NextIntlClientProvider` no longer forces every locale JSON file to be compile-time identical before the targeted locale checks run
- `web/public/locale/zh-Hant.json`
  - filled the missing `aiAutomation.task` keys required by the current UI contract, including quick intents, output/risk/view labels, and diff preview copy
  - cleaned the temporary duplicate keys introduced during the host-encoding-safe patch round
- `web/public/locale/ja.json`
  - filled the same missing `aiAutomation.task` keys for the Japanese locale
  - cleaned the temporary duplicate keys introduced during the host-encoding-safe patch round

## Verification

- Passed: `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
- Passed: `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- Passed: `node scripts/verify-backup-logic.mjs`
- Passed: `node scripts/verify-locale-consistency.mjs`

## Result

- Outcome: success
- The backup page mainline is no longer blocked by `LocaleProvider` message typing or missing `aiAutomation.task` locale structure.

## Remaining Risks / Next Step

1. stay on the same Phase G backup pool and decide whether the next smallest closure is page-level backup detail copy cleanup or browser-grade evidence for the backup page
2. keep `verify-backup-logic.mjs` and `verify-locale-consistency.mjs` in the local proof chain because they are stable on this Windows host
3. if backup page work pauses, the next adjacent blocker on the frontend side is no longer provider typing; it is whichever screenshot-first page still lacks browser-grade evidence or has real contract drift
