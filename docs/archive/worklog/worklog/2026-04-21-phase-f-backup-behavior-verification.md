# 2026-04-21 Phase F Backup Behavior Verification

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氬浠藉鍏ラ〉琛屼负绾ч獙璇佹渶灏忛棴鐜?+- 鏃ユ湡锛?026-04-21
- 褰撳墠闃舵锛歅hase F / Milestone 6 validation closure
- 瀵瑰簲 milestone锛氶噷绋嬬 6 楠岃瘉涓庨儴缃?+
## 2. 寮€宸ュ墠杈撳叆

- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?11.5 鑺傘€佺 13 鑺傞噷绋嬬 6銆佺 14 鑺傞獙鏀舵爣鍑嗐€佺 16 鑺傚疄鏂借鍒?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?8 鑺?Phase F銆佺 10 鑺備换鍔℃ā鏉?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-21-phase-f-frontend-validation-source-sync.md`
- 鏈浠诲姟鐩爣锛氳ˉ涓婅缃〉澶囦唤/瀵煎叆/鍥炴粴涓绘祦绋嬬殑婧愮爜绾ц涓洪獙璇佽瘉鎹紝閬垮厤缁х画鍙潬鏂囨瀵归綈鎴栭潤鎬佸澹?smoke
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細automation memory銆乧anonical plan銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁嗘墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬佹枃妗ｃ€乣web/src/components/modules/setting/Backup.tsx`銆乣web/src/api/endpoints/setting.ts`銆乣web/package.json`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細璇诲彇浜?`using-superpowers` 涓?`brainstorming` 鎶€鑳借鏄庯紱澶嶇敤浜?automation memory 涓€滀笅涓€姝ュ簲琛?Phase F 琛屼负绾ч獙璇佲€濈殑缁撹锛涘鐢ㄤ簡鍓嶇涓荤嚎鏂囨。涓叧浜?backup 椤甸潰浠嶇己鐪熷疄琛屼负楠屾敹鐨勮褰?+- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈户缁睍寮€鏇存棭鐨勬枃妗堝瀷 Phase F worklog锛屽洜涓烘湰杞富浠诲姟宸茬粡鍒囧埌琛屼负绾ч獙璇佽ˉ璇侊紝涓嶅啀闇€瑕侀噸澶嶆壂鏂囨宸紓
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭棤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓嶅垱寤哄瓙 agent

## 3. 鏈纭鍒?+
- 蹇呴』鍥寸粫 Phase F 澶囦唤/瀵煎叆/鍥炴粴涓荤嚎鎺ㄨ繘
- 蹇呴』钀戒笅鐪熷疄浠ｇ爜澧為噺鍜屽彲鎵ц楠岃瘉锛屼笉鑳藉彧鍐欒鍒掓垨鎬荤粨
- 涓嶅紩鍏ヤ笌褰撳墠涓荤嚎鏃犲叧鐨勬柊妗嗘灦鎴栧ぇ鑼冨洿閲嶆瀯

## 4. 鏈绂佹浜嬮」

- 涓嶆墿鏁ｅ埌鍏朵粬椤甸潰鎴栨棤鍏?UI 娓呯悊
- 涓嶆敼鍚庣瀵煎叆璇箟
- 涓嶆妸婧愮爜绾ц涓洪獙璇佽鎶ユ垚瀹屾暣娴忚鍣?e2e 楠屾敹

## 5. 鏈楠屾敹鏉′欢

- 澶囦唤椤垫牳蹇冪姸鎬?椋庨櫓鎺ㄥ閫昏緫浠庣粍浠跺唴鎶界鍒板彲澶嶇敤妯″潡
- 瀛樺湪涓€鏉″彲閲嶅鎵ц鐨勬簮鐮佺骇楠岃瘉鍛戒护锛岃鐩?dry-run 缁戝畾銆丄pply Same Import 闂ㄦ銆乺ollback/import 椋庨櫓姹囨€汇€乸ost-import summary 鎺ㄥ
- 鍓嶇 `tsc --noEmit` 浠嶇劧閫氳繃

## 6. 鏈鍥炴粴鐐?+
- 鍒犻櫎 `web/src/components/modules/setting/backup-logic.ts`
- 鍥為€€ `web/src/components/modules/setting/Backup.tsx`
- 鍥為€€ `scripts/verify-backup-logic.mjs`
- 鍥為€€ `web/package.json` 涓殑 `test:backup-logic`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鎶界函閫昏緫锛屽啀鍥炴帴 UI锛屽啀琛ラ獙璇佽剼鏈?+- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/setting/Backup.tsx`
- 鍙楀奖鍝嶆帴鍙ｏ細鏃犳帴鍙ｅ绾﹀彉鍖栵紝浠呭鐢ㄧ幇鏈?`DBImportResult` / rollback preview 鏁版嵁缁撴瀯
- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紝涓昏鏄妸鐜版湁琛屼负鍒ゆ柇鏀舵暃鍒扮函鍑芥暟骞惰ˉ楠岃瘉鍏ュ彛

## 8. 瀹炴柦姝ラ

1. 浠?`Backup.tsx` 鎻愬彇鍏煎鎬ц鏁般€侀闄╂瑙堛€丄pply Same Import 瀹堝崼銆佺粨鏋滃睍绀烘枃妗堛€乸ost-import summary 鎺ㄥ鍒?`backup-logic.ts`
2. 鐢ㄦ娊绂诲悗鐨勭函鍑芥暟鏇挎崲缁勪欢鍐呭搴旀帹瀵煎拰瀹堝崼閫昏緫锛岄檷浣庣姸鎬佹紓绉婚闄?+3. 鏂板 `scripts/verify-backup-logic.mjs`锛岀敤 Node + `--experimental-strip-types` 鐩存帴鏍￠獙鏍稿績琛屼负璺緞
4. 鍦?`web/package.json` 鏆撮湶 `test:backup-logic` 鍛戒护
5. 杩愯婧愮爜绾ч獙璇佸拰鍓嶇 typecheck锛岀‘璁ゆ湰杞棴鐜垚绔?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛氭棤棰濆鏋勫缓
- 娴嬭瘯鍛戒护锛歚node --experimental-strip-types scripts/verify-backup-logic.mjs`
- 涓撻」楠岃瘉锛歚node web/node_modules/typescript/bin/tsc --noEmit`
- 鐜缁曡璇存槑锛氬綋鍓嶇幆澧?`pnpm` 涓嶅湪 PATH锛宍corepack pnpm` 鍙堜細鍥犺闂?`AppData\Local\node\corepack\lastKnownGood.json` 瑙﹀彂 `EPERM`锛屽洜姝ゆ湰杞洿鎺ヤ娇鐢?Node 鍜屼粨搴撳唴鏈湴 TypeScript 鍏ュ彛瀹屾垚楠岃瘉

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細鏂板鐨勬簮鐮佺骇楠岃瘉浼氬鏂囨鍜岀姸鎬佹帹瀵兼洿鏁忔劅锛屽悗缁敼 backup 娴佺▼鏃堕渶瑕佸悓姝ユ洿鏂伴獙璇佽剼鏈?+- 鍏煎鎬ч闄╋細浣庯紱缁勪欢琛屼负涓嶅彉锛屽彧鏄妸閫昏緫鎶界骞堕泦涓鐢?+- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氭湰杞湭璺戝畬鏁村墠绔?build锛屼絾 `tsc --noEmit` 閫氳繃
- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃锛宍node --experimental-strip-types scripts/verify-backup-logic.mjs` 杈撳嚭 `backup-logic verification passed`
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆乧anonical plan銆佽缁嗘墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬佹枃妗ｃ€乣using-superpowers`銆乣brainstorming`
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細memory 鏄庣‘瑕佹眰浼樺厛琛?Phase F 琛屼负绾ч獙璇侊紱canonical plan 涓?workflow 闄愬畾浜嗘湰杞繀椤绘湇鍔′簬澶囦唤/瀵煎叆/鍥炴粴涓荤嚎锛涘墠绔富绾挎枃妗ｆ槑纭簡 backup 椤典粛缂虹湡瀹炶涓鸿瘉鎹紱鎶€鑳借鏄庢弧瓒充細璇濊捣濮嬪拰瀹炵幇鍓嶆祦绋嬭姹?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒鎵嬪伐 smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆浼樺厛琛ユ簮鐮佺骇琛屼负楠岃瘉锛屼笉娑夊強娴忚鍣ㄧ骇涓婁紶/鍥炴粴浜や簰
- 寰呴獙璇侀〉闈㈡竻鍗曪細backup 椤电湡瀹炴祻瑙堝櫒浜や簰銆佹棩蹇楅〉瀹炴椂娴併€侀椤电獎灞忕粏鑺?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔¤寖鍥撮泦涓?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細褰撳墠楠岃瘉浠嶄笉鏄祻瑙堝櫒绾?e2e锛涜繕闇€瑕佽ˉ涓€灞傜湡瀹?backup/import/rollback 浜や簰楠屾敹锛屼紭鍏堣鐩栨枃浠朵笂浼犮€乨ry-run -> Apply Same Import銆乻napshot rollback preview/restore
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭槸锛屽綋鍓嶅凡鍏峰鍙鐢ㄧ殑婧愮爜绾ц涓洪獙璇佸叆鍙ｏ紝涓嬩竴杞彲鐩存帴缁х画琛ユ祻瑙堝櫒绾ф垨鏇撮珮灞傜殑 Phase F 琛屼负璇佹嵁
