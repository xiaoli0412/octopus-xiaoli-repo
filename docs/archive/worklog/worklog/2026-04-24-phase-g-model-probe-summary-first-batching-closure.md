# 2026-04-24 Phase G Model Probe Summary-First Batching Closure

- Master plan aligned before coding (yes/no): `yes`
- Mainline: `Phase G screenshot-first UI closure / setting model probe summary-first batching`
- Summary: stayed in the same Phase G screenshot-first pool and closed an adjacent no-browser settings task that directly matches `需求 72/73/74`: the `ModelProbe` card now keeps the default path compact, moves repeated explanation into `HelpHint`, and avoids rendering the full model accordion list at once by default. The expanded state now renders the first `12` models and reveals more in batches, which keeps the settings page lighter on this host while preserving the current per-model editing flow.
- Verification:
  - `node scripts/verify-model-probe-help.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `node scripts/verify-locale-consistency.mjs`
- Changed files:
  - `web/src/components/modules/setting/ModelProbe.tsx`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/ja.json`
  - `scripts/verify-model-probe-help.mjs`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-model-probe-summary-first-batching-closure.md`
- Result:
  - `ModelProbe` 首屏说明压缩为短句主说明，额外语义改由问号帮助提示承接，不再把整段解释直接铺在卡片正文里。
  - 模型行摘要去掉了每格重复 helper 文本，展开后的默认渲染量收口为 `12` 行，并通过“再显示 N 个模型”继续按批追加，避免同页一次性挂出大批 Accordion 行。
  - 四语 locale 与 no-browser 静态护栏同步更新，避免“实现已变、验证仍按旧文案/旧结构”的回退风险。
- Blocker:
  - 无新增阻塞；浏览器级 settings / screenshot 证据仍受当前宿主 `spawn EPERM` / CDP 问题影响，但本轮任务不依赖该链路。
- Next:
  1. stay in the same Phase G settings pool and apply the same `short copy + summary first + deferred detail` pattern to `CircuitBreaker` or `DynamicRouting`.
  2. if the host `spawn EPERM` blocker gets cleared, rerun settings/browser evidence to confirm the lighter `ModelProbe` structure still behaves well at `375px`.
  3. continue the Chinese copy cleanup pass on deeper settings/detail states only after the remaining high-occupancy settings cards are compressed.
