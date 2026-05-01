# 2026-04-22 Phase G Channel Create Flow Copy Recovery

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G screenshot-first UI closure
- Current phase: create-channel dialog layered guidance and help copy closure
- Core task: restore one active `flowCopy` definition in `ChannelForm` and remove the temporary broken state
- Support task: tighten `scripts/verify-channel-create-flow.mjs` so it only inspects the active code path

## Done

- Removed the temporary historical `flowCopy` comment blocks from `web/src/components/modules/channel/Form.tsx` and kept a single active `useMemo` definition.
- Preserved the current Simplified Chinese, Traditional Chinese, and English flow copy structure without expanding into locale JSON edits in this round.
- Updated `scripts/verify-channel-create-flow.mjs` to require exactly one active `flowCopy` block and reject matches against `flowCopyLegacy` or commented leftovers.

## Verification

- `node scripts/verify-channel-create-flow.mjs` passed
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` passed

## Risks

- Browser/manual smoke for the channel dialog is still missing.
- The repository worktree is still heavily dirty, so follow-up rounds should remain narrowly scoped.

## Next

1. Continue the remaining create-channel dialog help and copy closure on the same mainline.
2. Then move to `CC Switch` progressive help closure.
3. Add browser-level smoke for the create-channel dialog when the environment is available.
