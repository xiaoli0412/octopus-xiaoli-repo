# 2026-04-21 Phase F Replace-Prune Component Verification

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛欱ackup 椤甸潰 replace 妯″紡 prune 椋庨櫓缁勪欢绾ч獙璇佽ˉ寮?+- 鏃ユ湡锛?026-04-21
- 褰撳墠闃舵锛歅hase F / Milestone 6 validation closure
- 瀵瑰簲 milestone锛氶噷绋嬬 6 楠岃瘉涓庨儴缃?+
## 2. 寮€宸ュ墠杈撳叆

- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?11.5 鑺傘€佺 13 鑺傞噷绋嬬 6銆佺 14 鑺傞獙鏀舵爣鍑嗐€佺 16 鑺傚疄鏂借鍒?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?8 鑺?Phase F銆佺 10 鑺備换鍔℃ā鏉?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-21-phase-f-backup-component-verification.md`
- 鏈浠诲姟鐩爣锛氬湪宸叉湁 backup 缁勪欢绾ч獙璇侀摼涓婅ˉ榻?`replace` 瀵煎叆璺緞锛岄獙璇?replace-prune 椋庨櫓灞曠ず銆丄pply Same Import 纭瀹堝崼锛屼互鍙?dry-run preview token 缁戝畾琛屼负
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細automation memory銆乧anonical plan銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁嗘墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬佹枃妗ｃ€乣Backup.tsx`銆乣Backup.test.tsx`銆乣scripts/verify-backup-component.cjs`銆乣scripts/verify-backup-component.setting-mock.cjs`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細automation memory銆乣docs/LLM-Gateway-Refactor-Plan.zh-CN.md`銆乣docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`銆乣docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`銆乣docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆浠诲姟宸茬粡鑱氱劍鍒?Phase F backup/import/rollback 鍗曢〉楠岃瘉鏀跺彛锛屼笉闇€瑕佺户缁墿灞曞埌鍏朵粬闃舵 worklog 鎴栨棤鍏虫ā鍧?+- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭棤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓嶈鍒涘缓瀛?agent锛屼笖鏈疆浠诲姟鑼冨洿闆嗕腑鍦ㄥ崟椤甸獙璇侀摼锛屼富绾跨▼澶勭悊鏇寸ǔ濡?+
## 3. 鏈纭鍒?+
- 蹇呴』缁х画鍥寸粫 Phase F backup / import / rollback 涓荤嚎鎺ㄨ繘
- 蹇呴』鐣欎笅鐪熷疄浠ｇ爜澧為噺涓庡彲鎵ц楠岃瘉缁撴灉锛屼笉鑳藉彧鍋氭枃妗ｆ垨鍙ｅご鎬荤粨
- 涓嶆敼鍚庣瀵煎叆/鍥炴粴璇箟锛屽彧琛ュ己鍓嶇楠岃瘉璇佹嵁

## 4. 鏈绂佹浜嬮」

- 涓嶆墿鏁ｅ埌鏃犲叧椤甸潰鎴栧叾浠栦富绾夸换鍔?+- 涓嶆妸鑴氭湰绾ч獙璇佽鎶ユ垚娴忚鍣ㄧ骇 e2e
- 涓嶄负浜嗘祴璇曚究鍒╀慨鏀逛骇鍝佽涔?+
## 5. 鏈楠屾敹鏉′欢

- 浠撳簱閲屽瓨鍦ㄤ竴鏉″彲閲嶅鎵ц鐨?`replace` 妯″紡缁勪欢绾ч獙璇佽矾寰?+- 璇ラ獙璇佽鐩?replace-prune 椋庨櫓灞曠ず銆佹樉寮忕‘璁ゅ畧鍗€乸review token 澶嶇敤
- `node scripts/verify-backup-component.cjs`銆乣node scripts/verify-backup-logic.mjs`銆乣node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 鍏ㄩ儴閫氳繃

## 6. 鏈鍥炴粴鐐?+
- 鍥為€€ `scripts/verify-backup-component.cjs`
- 鍥為€€ `scripts/verify-backup-component.setting-mock.cjs`
- 鍥為€€ `web/src/components/modules/setting/Backup.test.tsx`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氫笉鏀逛笟鍔¤涔夛紝鐩存帴琛ラ獙璇佽剼鏈笌娴嬭瘯瑕嗙洊
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/setting/Backup.test.tsx`
- 鍙楀奖鍝嶆帴鍙ｏ細鏃犳帴鍙ｅ绾﹀彉鏇达紝浠呭鐢ㄧ幇鏈?`DBImportResult` 涓殑 `replace_prune_preview` / `compatibility.replace_pruned_*` 缁撴瀯
- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紝鏂板鐨勬槸楠岃瘉鍏ュ彛鍜屾祴璇?mock

## 8. 瀹炴柦姝ラ

1. 闃呰褰撳墠 Phase F backup 楠岃瘉閾句笌 `Backup.tsx` 涓?replace-prune / Apply Same Import 鐩稿叧瀹炵幇锛岀‘璁ゆ渶灏忓彲楠岃瘉鑺傜偣
2. 鎵╁睍 `scripts/verify-backup-component.setting-mock.cjs`锛岃 `replace` dry-run 杩斿洖缁撴瀯鍖?prune 椋庨櫓涓?credential rebind 鏁版嵁
3. 鎵╁睍 `scripts/verify-backup-component.cjs`锛屾柊澧?replace 妯″紡琛屼负楠岃瘉锛岃鐩栭闄╁睍绀恒€佹樉寮忕‘璁や笌 preview token 澶嶇敤
4. 鎵╁睍 `web/src/components/modules/setting/Backup.test.tsx`锛屾柊澧炶交閲?`Select` mock 涓?replace 妯″紡娴嬭瘯鑽夋锛屽苟淇娴嬭瘯杈呭姪绫诲瀷浣垮墠绔?typecheck 閫氳繃

## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛氭棤棰濆鏋勫缓
- 娴嬭瘯鍛戒护锛?+  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 涓撻」楠岃瘉锛氱‘璁?replace dry-run 浼氬睍绀?`Replace-prune preview`銆乣Channels to delete`銆乣API keys to delete`銆乣Replace mode can remove current project records...`銆乣Replace-prune preview will delete or reset 2 current records.`锛屽苟涓?Apply Same Import 缁х画缁戝畾 `preview-token-replace`

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細缁勪欢楠岃瘉鑴氭湰涓庢祴璇?mock 鐜板湪鏄惧紡渚濊禆 replace-prune 椋庨櫓缁撴瀯锛涘悗缁嫢 Backup 椤垫敼鍔ㄨ缁撴瀯锛岄渶瑕佸悓姝ユ洿鏂伴獙璇佸叆鍙?+- 鍏煎鎬ч闄╋細浣庯紱鏈慨鏀逛骇鍝佽涓猴紝鍙ˉ楠岃瘉涓庢祴璇曡祫浜?+- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚tsc --noEmit` 閫氳繃
- 娴嬭瘯鏄惁閫氳繃锛歚verify-backup-component.cjs` 涓?`verify-backup-logic.mjs` 閫氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆乧anonical plan銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁嗘墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬佹枃妗ｃ€佺幇鏈?backup 鑴氭湰涓庢祴璇曟枃浠?+- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細memory 鏄庣‘涓嬩竴姝ュ簲琛ユ洿楂樺眰 backup 琛屼负璇佹嵁锛沜anonical plan 涓?workflow 闄愬畾鏈疆蹇呴』缁х画鏈嶅姟浜?Phase F锛涘墠绔富绾跨姸鎬佹枃妗ｇ‘璁?backup 椤典粛闇€缁х画寮哄寲楠岃瘉璇佹嵁锛涚幇鏈夎剼鏈笌娴嬭瘯鏂囦欢鎻愪緵浜嗗彲澶嶇敤鐨勫崟杩涚▼楠岃瘉楠ㄦ灦
- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒鎵嬪伐 smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細褰撳墠鐜浠嶅彈 Windows 涓?Vitest/Vite/browser 鑷姩鍖?`spawn EPERM` 闄愬埗锛屾湰杞户缁噰鐢ㄤ粨搴撳唴鍗曡繘绋嬭剼鏈畬鎴愮粍浠剁骇闂幆楠岃瘉
- 寰呴獙璇侀〉闈㈡竻鍗曪細backup 椤电湡瀹炴祻瑙堝櫒涓婁紶/瀵煎叆/鍥炴粴浜や簰銆佹棩蹇楅〉瀹炴椂娴併€侀椤电獎灞忕粏鑺?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔¤仛鐒︿簬鍗曢〉楠岃瘉閾?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細褰撳墠 replace 璺緞宸叉湁缁勪欢绾у彲鎵ц璇佹嵁锛屼絾浠嶄笉鏄祻瑙堝櫒绾х湡瀹炰笂浼?smoke锛涜嫢鐜鎭㈠锛屽簲浼樺厛鎶?`Backup.test.tsx` 鎺ュ洖 Vitest 姝ｅ父杩愯閾惧苟琛ョ湡瀹為〉闈㈢骇 smoke
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭槸锛涗笅涓€杞彲缁х画鍦ㄥ悓涓€涓荤嚎涓嬫帹杩涙祻瑙堝櫒绾?backup/import/rollback smoke锛屾垨琛ュ彟涓€鏉″悓绛夐闄╃骇鍒殑鐪熷疄浜や簰楠岃瘉
