# 2026-04-21 Phase F Backup Component Verification

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛欱ackup 瀵煎叆/鍥炴粴缁勪欢绾ц涓洪獙璇佹敹鍙?+- 鏃ユ湡锛?026-04-21
- 褰撳墠闃舵锛歁ilestone 6 validation closure / Phase F backup-import-rollback verification
- 瀵瑰簲 milestone锛氶噷绋嬬 6 楠岃瘉涓庨儴缃?+
## 2. 寮€宸ュ墠杈撳叆

- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?11.5 鑺傘€佺 13 鑺傞噷绋嬬 6銆佺 14 鑺傞獙鏀舵爣鍑嗐€佺 16 鑺傚疄鏂借鍒?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?8 鑺?Phase F銆佺 10 鑺備换鍔℃ā鏉?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-21-phase-f-backup-behavior-verification.md`
- 鏈浠诲姟鐩爣锛氭妸 Backup 椤电殑楠岃瘉浠庢簮绾х函鍑芥暟鏍￠獙鎺ㄨ繘鍒扮粍浠剁骇琛屼负璇佹嵁锛岃鐩?dry-run銆丄pply Same Import 椋庨櫓纭銆乸ost-import summary銆乻napshot rollback preview
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細automation memory銆乧anonical plan銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁嗘墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬佹枃妗ｃ€乣web/src/components/modules/setting/Backup.tsx`銆乣backup-logic.ts`銆佺幇鏈?`backup-logic.test.ts`銆乣DynamicRouting.test.tsx`銆乣web/package.json`銆乣web/vitest.config.ts`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細璇诲彇浜?`using-superpowers` 涓?`brainstorming` 鎶€鑳借鏄庯紱澶嶇敤浜?automation memory 涓€滀紭鍏堣ˉ Backup 琛屼负绾ч獙璇佲€濈殑涓嬩竴姝ョ粨璁猴紱澶嶇敤浜嗗墠绔富绾跨姸鎬佹枃妗ｉ噷瀵?Backup 椤典粛缂虹湡瀹炶涓洪獙璇佺殑鍒ゆ柇
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈户缁睍寮€鏇存棭鐨?copy alignment worklog锛屽洜涓鸿繖浜涚粨璁哄凡缁忚鍚庣画鐘舵€佹枃妗ｅ拰 automation memory 鍚告敹锛屾湰杞彧鑱氱劍琛屼负楠岃瘉闂幆
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭棤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓嶈鍒涘缓瀛?agent锛屼笖鏈疆浠诲姟闆嗕腑鍦ㄥ崟涓〉闈笌楠岃瘉鑴氭湰

## 3. 鏈纭鍒?+
- 蹇呴』缁х画鍥寸粫 Phase F backup / import / rollback 涓荤嚎鎺ㄨ繘
- 蹇呴』鐣欎笅鐪熷疄浠ｇ爜澧為噺鍜屽彲鎵ц楠岃瘉鍏ュ彛锛屼笉鑳藉彧鍐欐€荤粨
- 涓嶅緱鎶婃湰鏈哄彈闄愮殑 Vitest/娴忚鍣ㄧ幆澧冮樆濉炴墿鏁ｆ垚鏁翠釜涓荤嚎鍋滄粸

## 4. 鏈绂佹浜嬮」

- 涓嶆墿鏁ｅ埌鏃犲叧 UI 娓呯悊鎴栨枃妗堝井璋?+- 涓嶆敼鍚庣瀵煎叆/鍥炴粴璇箟
- 涓嶆妸鍗曡繘绋嬭剼鏈獙璇佽鎶ヤ负娴忚鍣ㄧ骇 e2e

## 5. 鏈楠屾敹鏉′欢

- 浠撳簱涓柊澧炰竴鏉″彲閲嶅鎵ц鐨?Backup 缁勪欢绾ч獙璇佸懡浠?+- 琛屼负楠岃瘉瑕嗙洊 dry-run -> Apply Same Import -> post-import summary锛屼互鍙?snapshot history -> rollback preview
- 鍓嶇 `tsc --noEmit` 缁х画閫氳繃
- 鍘熸湁 `backup-logic` 婧愮骇楠岃瘉缁х画閫氳繃

## 6. 鏈鍥炴粴鐐?+
- 鍒犻櫎 `web/src/components/modules/setting/Backup.test.tsx`
- 鍒犻櫎 `scripts/verify-backup-component.cjs`
- 鍒犻櫎 `scripts/verify-backup-component.*-mock.cjs`
- 鍥為€€ `web/package.json` 涓柊澧炵殑 `test:backup-component`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛琛ユ祴璇曚笌楠岃瘉鍏ュ彛锛屼笉鏀逛富涓氬姟 UI 璇箟
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/setting/Backup.test.tsx`
- 鍙楀奖鍝嶆帴鍙ｏ細鏃犳帴鍙ｅ绾﹀彉鏇达紱鍙秷璐圭幇鏈?`DBImportResult`銆乺ollback preview 缁撴瀯
- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紝鏂板鐨勬槸楠岃瘉鍏ュ彛鍜屾祴璇曡祫浜?+
## 8. 瀹炴柦姝ラ

1. 闃呰 `Backup.tsx`銆乣backup-logic.ts`銆佺幇鏈夋祴璇曚笌 Phase F 鐩稿叧 worklog锛岀‘璁ょ粍浠跺眰鏈€鍊煎緱瑕嗙洊鐨勮涓鸿妭鐐?+2. 鏂板 `web/src/components/modules/setting/Backup.test.tsx`锛岀敤 Vitest 椋庢牸鎻忚堪 dry-run銆丄pply Same Import銆乺ollback preview 涓ゆ潯涓昏矾寰?+3. 鐢变簬鏈満 Vitest/Vite 鍦?Windows 涓嬫寔缁Е鍙?`spawn EPERM`锛屾敼涓烘柊澧炲崟杩涚▼缁勪欢楠岃瘉鑴氭湰 `scripts/verify-backup-component.cjs`
4. 涓哄崟杩涚▼鑴氭湰琛ユ渶灏?mock 妯″潡锛?+   - `scripts/verify-backup-component.next-intl-mock.cjs`
   - `scripts/verify-backup-component.toast-mock.cjs`
   - `scripts/verify-backup-component.setting-mock.cjs`
5. 鍦?`web/package.json` 涓柊澧?`test:backup-component` 鍏ュ彛锛屽浐瀹氳繖鏉￠獙璇侀摼

## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛氭棤棰濆鏋勫缓
- 娴嬭瘯鍛戒护锛?+  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 涓撻」楠岃瘉锛?+  - 楠岃瘉 dry-run 棣栨璋冪敤浼氫骇鍑?preview token 骞惰姹?Apply Same Import 椋庨櫓纭
  - 楠岃瘉 Apply Same Import 浼氬鐢?dry-run 缁戝畾鐨?preview token
  - 楠岃瘉 apply 涔嬪悗鑳界湅鍒?post-import validation / health check 鍖哄潡
  - 楠岃瘉 snapshot history 鐨?Preview 浼氬睍绀?rollback preview 闈㈡澘

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細鏂板浜嗕竴鏉¤剼鏈骇缁勪欢楠岃瘉閾撅紝鍚庣画濡傛灉 Backup 椤典緷璧栨柊澧?provider/hook锛岄渶瑕佸悓姝ユ洿鏂?mock 妯″潡
- 鍏煎鎬ч闄╋細浣庯紱鏂板鏂囦欢涓嶆敼鍙樹骇鍝佽涓猴紝鍙鍔犻獙璇佽祫浜?+- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚tsc --noEmit` 閫氳繃
- 娴嬭瘯鏄惁閫氳繃锛歚verify-backup-component.cjs` 涓?`verify-backup-logic.mjs` 閫氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆乧anonical plan銆佽缁嗘墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬佹枃妗ｃ€乣Backup.tsx`銆佸凡鏈?backup logic 楠岃瘉鏂囦欢銆乣using-superpowers`銆乣brainstorming`
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細memory 鏄庣‘瑕佹眰浼樺厛琛?Backup 琛屼负楠岃瘉锛沜anonical plan 涓?workflow 闄愬畾鏈疆蹇呴』鏈嶅姟浜?Phase F锛涘墠绔富绾跨姸鎬佹枃妗ｇ‘璁?Backup 椤佃繕缂虹湡瀹炶涓虹骇璇佹嵁锛涘凡鏈?`backup-logic` 楠岃瘉鎻愪緵浜嗘簮绾у熀纭€锛屼笉闇€瑕侀噸澶嶉€犺疆瀛?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒鎵嬪伐 smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈満 Vitest/Vite 涓庢祻瑙堝櫒鑷姩鍖栭摼鍧囧彈 Windows `spawn EPERM` 鐜闄愬埗锛涙湰杞敼鐢ㄤ粨搴撳唴鍗曡繘绋嬭剼鏈畬鎴愮粍浠剁骇闂幆楠岃瘉
- 寰呴獙璇侀〉闈㈡竻鍗曪細娴忚鍣ㄥ唴鐪熷疄鏂囦欢涓婁紶涓庡洖婊氭寜閽氦浜掋€佹棩蹇楅〉瀹炴椂娴併€侀椤电獎灞忕粏鑺?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔″己鑰﹀悎浜庡崟涓〉闈㈤獙璇?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細`Backup.test.tsx` 褰撳墠浣滀负姝ｅ紡娴嬭瘯鑽夌淇濈暀锛屼絾鏈満鏈€缁堝彲鎵ц閾句粛渚濊禆鍗曡繘绋?`verify-backup-component.cjs` 鑰屼笉鏄?Vitest锛涘悗缁嫢鐜鎭㈠锛屽簲浼樺厛鎶婅繖浠芥祴璇曟帴鍥?Vitest 姝ｅ父杩愯閾?+- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭槸锛孊ackup 椤靛凡缁忓叿澶囩粍浠剁骇鍙墽琛岄獙璇佸叆鍙ｏ紝涓嬩竴杞彲缁х画琛?export 榛樿璇箟鎴栨洿鐪熷疄鐨勬祻瑙堝櫒绾?smoke
