# 2026-04-25 Phase G Backup Panel Selector Follow-up

- Master plan aligned before coding: yes
- Mainline: Phase G screenshot-first UI closure / backup page selector-contract follow-up
- Stage: backup page panel-level selector closure and repo-local proof tightening

## Context Reused

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
- automation memory: `$CODEX_HOME/automations/octopus-2/memory.md`
- recent backup worklogs from `2026-04-25`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/components/modules/setting/Backup.test.tsx`
- `scripts/verify-backup-component.cjs`

## Plan

- stay on the same Phase G backup selector-contract mainline
- pick the next smallest panel-level selector gap from `Backup.tsx`
- keep the write scope inside backup page verification only
- verify with repo-local typecheck and backup verification, then probe Vitest only if the host allows it

## Candidate Tasks Considered

1. Add missing component-test coverage for `backup-pending-apply-panel`
2. Add missing repo-local assertions for parent panel selectors already present in `Backup.tsx`
3. Expand to browser-grade backup smoke using existing page anchors
4. Touch backup logic copy or import behavior

## Chosen Task

- Core task: close the smallest remaining panel-level selector gap by locking `backup-pending-apply-panel` in the component test flows
- Companion task: tighten `scripts/verify-backup-component.cjs` with `backup-remaining-migration-panel` and `backup-replace-prune-panel`

## 鏈纭鍒?+
- 缁х画娌跨敤褰撳墠 Phase G backup selector-contract 涓荤嚎
- 鍙ˉ椤甸潰绾?panel selector 濂戠害锛屼笉鎵╁埌澶囦唤涓氬姟璇箟
- 淇濇寔 selector-first銆乺epo-local proof-first 鐨勯獙璇佹柟寮?+
## 鏈绂佹浜嬮」

- 涓嶆敼瀵煎叆/瀵煎嚭/鍥炴粴閫昏緫
- 涓嶆敼 locale 閫昏緫涓庡府鍔╂枃妗?+- 涓嶆墿鍒版祻瑙堝櫒绾?smoke 淇
- 涓嶈Е纰伴潪澶囦唤椤甸潰鏂囦欢

## 鏈楠屾敹鏉′欢

- `Backup.test.tsx` 鍦?dry-run/apply銆乵ap銆乺eplace 涓夋潯娴侀噷绋冲畾鏂█ `backup-pending-apply-panel`
- `scripts/verify-backup-component.cjs` 绋冲畾鏂█ `backup-remaining-migration-panel` 涓?`backup-replace-prune-panel`
- `tsc --noEmit` 涓?`verify-backup-component.cjs` 閫氳繃

## 鏈鍥炴粴鐐?+
- 濡傛柊澧?selector 鏂█璇垽锛屽洖閫€鏈疆鏂板鏂█鍗冲彲

## 瀹炵幇鑼冨洿

- 鍏堣ˉ缁勪欢娴嬭瘯閲岀殑 panel-level selector 鏂█
- 鍐嶈ˉ repo-local verification script 閲岀殑鐖跺眰 panel selector 鏂█
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/setting/Backup.test.tsx`
- 鍙楀奖鍝嶈剼鏈細`scripts/verify-backup-component.cjs`
- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚?+
## 瀹炴柦姝ラ

1. 瀵圭収 `Backup.tsx`銆乣Backup.test.tsx`銆乣verify-backup-component.cjs` 鐩樼偣鐜版湁 `backup-*` selector 瑕嗙洊銆?+2. 鍦ㄧ粍浠舵祴璇曠殑 incremental / map / replace 涓夋潯璺緞涓ˉ涓?`backup-pending-apply-panel` 鏂█銆?+3. 鍦?repo-local verification script 涓ˉ涓?`backup-remaining-migration-panel` 涓?`backup-replace-prune-panel` 鐨勬柇瑷€銆?+4. 杩愯 `tsc --noEmit` 涓?`verify-backup-component.cjs`銆?+5. 棰濆璇曡窇鏈€灏?Vitest 鍗曟枃浠跺叆鍙ｏ紝璁板綍瀹夸富绾х粨鏋滐紝涓嶆妸瀹冧綔涓烘湰杞€氳繃鍓嶆彁銆?+
## 瀹為檯瀹屾垚

- 鍦?`Backup.test.tsx` 鐨?dry-run/apply銆乵ap銆乺eplace 涓夋潯宸插畬鎴愰妫€鐨勮矾寰勪腑锛岃ˉ涓婁簡 `backup-pending-apply-panel` 鐨勬樉寮忔柇瑷€锛岄伩鍏嶅彧閿佷綇 ready badge 鑰屾湭閿佷綇鐖剁骇 panel 瀹瑰櫒銆?+- 鍦?`scripts/verify-backup-component.cjs` 涓ˉ涓婁簡锛?+  - 鎵撳紑鍓╀綑杩佺Щ鑳藉姏鎶樺彔鍖哄悗鐨?`backup-remaining-migration-panel` 鏂█
  - replace 妯″紡棰勬鍚庣殑 `backup-replace-prune-panel` 鏂█
- 鏈疆鏈敼鍔?`Backup.tsx` 涓氬姟閫昏緫锛屼篃鏈敼鍔ㄥ浠?copy / locale / import behavior銆?+
## 娴嬭瘯涓庨獙璇?+
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- `node scripts/verify-backup-component.cjs`
- `node -r .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run web/src/components/modules/setting/Backup.test.tsx`

## 楠岃瘉缁撴灉

- `tsc --noEmit`锛氶€氳繃
- `verify-backup-component.cjs`锛氶€氳繃
- 鍗曟枃浠?Vitest锛氭湭閫氳繃锛屽涓讳粛鎶?`spawn EPERM`锛屽睘浜庡凡鐭ョ幆澧冮樆濉烇紝涓嶆槸鏈疆鏂板 selector 鏂█瀵艰嚧鐨勫洖褰?+
## 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細鏂板 panel selector 鏂█浼氭洿鏃╂毚闇查〉闈㈢粨鏋勫洖閫€锛屼絾灞炰簬浣庨闄┿€佹鍚戞姢鏍?+- 鍏煎鎬ч闄╋細浣庯紝浠呭奖鍝嶆祴璇曚笌 repo-local 楠岃瘉鑴氭湰
- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氭湰杞湭璺?build
- 娴嬭瘯鏄惁閫氳繃锛歚tsc` 涓?repo-local backup verification 閫氳繃锛沄itest 鍗曟枃浠跺彈瀹夸富 `spawn EPERM` 闃诲
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細AGENTS.md銆丩LM-Gateway-Refactor-Plan.zh-CN.md銆丆URRENT_STATUS_AND_PLAN.zh-CN.md銆丗RONTEND_UI_MAINLINE_STATUS.zh-CN.md銆丏ETAILED_EXECUTION_WORKFLOW.zh-CN.md銆乤utomation memory銆佸悓鏃?backup worklogs銆乣Backup.tsx`銆乣Backup.test.tsx`銆乣verify-backup-component.cjs`
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細纭褰撳墠涓荤嚎浠嶆槸 Phase G backup selector-contract 鏀跺彛锛屼笖鏈満鏈€绋冲畾鐨勮瘉鏄庨摼浠嶆槸 `tsc --noEmit + verify-backup-component.cjs`
- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭湭浣跨敤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭湭浣跨敤瀛?agent锛屽師鍥犳槸鏈疆浠诲姟鍐欎綔鐢ㄥ煙寰堝皬锛屼笖鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬩覆琛屾帹杩?+- 鎵嬪伐 smoke 鐘舵€?/ 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鏈仛娴忚鍣?smoke锛沄itest 鐩存帴杩愯浠嶅彈瀹夸富 `spawn EPERM` 闃诲

## 閬楃暀椤?+
- `Backup.test.tsx` 涓?repo-local script 涔嬪鐨?Vitest 姝ｅ紡璺戦€氫粛鍙楀涓?`spawn EPERM` 褰卞搷
- backup 椤甸潰鏄惁杩樺瓨鍦ㄦ洿灏忕殑 panel-level selector 鍙岄噸瑕嗙洊缂哄彛锛岄渶瑕佸湪涓嬩竴杞户缁洏鐐?+- browser-grade 澶囦唤椤佃瘉鎹粛鏈ˉ榻?+
## 涓嬩竴杞缓璁?+
1. 缁х画鐣欏湪鍚屼竴鏉?Phase G backup selector-contract 涓荤嚎锛屼紭鍏堟壘涓嬩竴涓皻鏈鈥滅粍浠舵祴璇?+ repo-local script鈥濆弻閲嶅浐鍖栫殑 panel/root selector銆?+2. 濡傛灉椤甸潰绾?selector 缂哄彛宸茬粡鍩烘湰娓呭畬锛屽氨杞悜 backup 椤甸潰娴忚鍣ㄧ骇璇佹嵁锛屼紭鍏堝鐢ㄧ幇鏈?`backup-page / backup-pending-apply-panel / backup-history-panel / backup-remaining-migration-panel` 閿氱偣銆?+3. 缁х画鎶?`tsc --noEmit + verify-backup-component.cjs` 浣滀负杩欏彴鏈哄櫒涓婄殑绋冲畾 proof path锛沄itest 浠嶅彧浣滀负闄勫姞鎺㈡祴锛屼笉浣滀负涓婚€氳繃渚濇嵁銆?+
