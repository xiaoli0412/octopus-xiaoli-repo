# 2026-04-23 Deep Audit Backup Verification Locale Contract Closure

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氬浠介獙璇佽剼鏈?locale 濂戠害婕傜Щ淇涓庡叡浜鎯呰嫳鏂囧垎闅旂鏀跺彛
- 鏃ユ湡锛歚2026-04-23`
- 褰撳墠闃舵锛歚Phase A` 绋冲畾鎬?楠岃瘉鏀跺彛锛岄檮甯?`Phase F` Backup verification hygiene 淇ˉ
- 瀵瑰簲 milestone锛氶獙璇侀摼璺彲鍥炲綊涓庡彂甯冨墠灏忛闄╀慨澶?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): `yes`
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 涓ǔ瀹氭€с€侀獙璇佷笌澶囦唤瀵煎叆瀵煎嚭鏀跺彛鐩稿叧绔犺妭
- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 鐨?`1.0`銆乣1.2`銆乣1.3`
- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-23-deep-audit-task-race-and-release-toolchain-pin.md`
- 鏈浠诲姟鐩爣锛氫慨澶嶅凡鎺ュ叆 CI 鐨?Backup 楠岃瘉鑴氭湰涓庡綋鍓?locale/UI 濂戠害婕傜Щ锛屽苟琛ヤ笂鍏变韩 helper 鐨勮嫳鏂囪鎯呭垎闅旂涓€鑷存€?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細`AGENTS.md`銆乤utomation memory銆乣docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`銆乣docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`銆佹渶杩?review/worklog銆乣scripts/verify-backup-logic.mjs`銆乣scripts/verify-backup-component.cjs`銆乣web/src/components/modules/setting/backup-logic.ts`銆乣web/src/components/modules/setting/Backup.tsx`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細automation memory銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁嗗伐浣滄祦銆佹渶杩戝鏌ユ姤鍛娿€丳hase F/Phase G 鐩稿叧 backup worklog 閾?+- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鑱氱劍楠岃瘉閾捐矾涓庡叡浜?helper锛屼笉闇€瑕佹墿灞曞埌 Docker/browser 璺緞
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細`鍚
- 鏈瀛?agent 浣跨敤妯″瀷锛歚N/A`
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氫富绾跨▼鎵ц
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛歚N/A`
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏈疆鏄庣‘瑕佹眰涓嶈鍒涘缓瀛?agent锛屼笖鏈淇ˉ鑼冨洿闆嗕腑涓斾綆椋庨櫓

## 3. 鏈纭鍒?+
- 蹇呴』鍏堝鐜?CI 涓湡瀹炲け璐ワ紝鍐嶅喅瀹氳ˉ涓佹柟寮?+- 浼樺厛淇獙璇佽祫浜т笌鍏变韩 helper锛屼笉鏀?Backup 椤甸潰榛樿涓枃琛屼负
- 浠讳綍浠ｇ爜鍙樺姩鍚庨兘蹇呴』璺戞渶灏忓繀瑕侀獙璇侊紝涓嶈兘鎶婃湭鎵ц鐨勫懡浠ゅ啓鎴愰€氳繃

## 4. 鏈绂佹浜嬮」

- 涓嶆敼 Backup 鍚庣濂戠害
- 涓嶆妸鏈疆鎵╁睍鎴愬ぇ鑼冨洿鍓嶇鏂囨鏀归€?+- 涓嶅洖閫€鐢ㄦ埛宸ヤ綔鍖轰腑鐨勬棤鍏宠剰鏀瑰姩

## 5. 鏈楠屾敹鏉′欢

- `node .\scripts\verify-backup-logic.mjs` 閫氳繃
- `node .\scripts\verify-backup-component.cjs` 閫氳繃
- `node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 閫氳繃

## 6. 鏈鍥炴粴鐐?+
- `scripts/verify-backup-logic.mjs`
- `scripts/verify-backup-component.cjs`
- `web/src/components/modules/setting/backup-logic.ts`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛淇獙璇佽剼鏈绾︼紝鍐嶈ˉ鍏变韩 helper 鐨勮嫳鏂囪鎯呰緭鍑?+- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/setting/backup-logic.ts`
- 鍙楀奖鍝嶆帴鍙ｏ細鏃?+- 鏄惁褰卞搷鏃ф暟鎹細`鍚
- 鏄惁褰卞搷鏃ц涓猴細浠呭奖鍝嶉獙璇侀摼璺拰鑻辨枃 locale 涓嬬殑璇︽儏瀛楃涓叉牸寮忥紱榛樿涓枃 UI 琛屼负涓嶅彉

## 8. 瀹炴柦姝ラ

1. 閲嶆柊绱㈠紩骞惰鍙栨湰杞浉鍏虫枃妗ｃ€乺eview銆乵emory 涓?backup 楠岃瘉鑴氭湰锛屽疄鐜板墠鍏堝鐜颁袱鏉″け璐ュ懡浠ゃ€?+2. 灏?`scripts/verify-backup-component.cjs` 娓叉煋鍓嶇殑璁剧疆 store 鏄惧紡鍒囧埌 `en`锛屼娇鑻辨枃鏂█涓庣粍浠堕殣钘忔祴璇曢敋鐐瑰绾﹂噸鏂板榻愩€?+3. 灏?`scripts/verify-backup-logic.mjs` 鏀逛负鏄惧紡浼犲叆 `locale: 'en'`锛屽苟鍚屾瀵归綈褰撳墠 shared helper 鐨?`|` 鍒嗘鏍煎紡銆?+4. 鍦?`web/src/components/modules/setting/backup-logic.ts` 鏂板灞€閮ㄨ嫳鏂囪鎯呭綊涓€鍖?helper锛屼慨澶嶈嫳鏂?locale 涓嬩粛娣峰叆涓枃鍒嗛殧绗?`銆乣 鐨勫叡浜瓧绗︿覆杈撳嚭銆?+5. 澶嶈窇 TypeScript 涓庝袱鏉?backup 楠岃瘉鍛戒护锛岀‘璁ら摼璺仮澶嶃€?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛歚node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 娴嬭瘯鍛戒护锛?+  - `node .\scripts\verify-backup-logic.mjs`
  - `node .\scripts\verify-backup-component.cjs`
- 涓撻」楠岃瘉锛?+  - 鍏堝鐜版棫澶辫触锛屽啀澶嶈窇纭鎭㈠
  - 澶嶆煡 `git diff -- web/src/components/modules/setting/backup-logic.ts scripts/verify-backup-logic.mjs scripts/verify-backup-component.cjs`

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱琛ヤ竵浠呴檺楠岃瘉鑴氭湰涓?shared helper 鐨?locale 杈撳嚭
- 鍏煎鎬ч闄╋細浣庯紱榛樿 `zh-Hans` 椤甸潰涓嶅彉锛屼粎璁╄嫳鏂囬獙璇侀摼鍥炲埌鑷唇鐘舵€?+- 鏄惁闃诲涓嬩竴浠诲姟锛歚鍚

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚閫氳繃`
- 娴嬭瘯鏄惁閫氳繃锛歚閫氳繃`
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆佸綋鍓嶇姸鎬佷笌 workflow 鏂囨。銆佹渶杩戝鏌ユ姤鍛娿€乥ackup 鐩稿叧鍘嗗彶 worklog銆佸綋鍓?backup 鑴氭湰涓庣粍浠舵簮鐮?+- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - automation memory锛氭寚鍑轰笂涓€杞凡缁忕‘璁?backup 鑴氭湰鏄亣闃存€ч樆濉烇紝搴斾紭鍏堜慨楠岃瘉閾?+  - 褰撳墠鐘舵€?workflow锛氱害鏉熸湰杞繀椤诲厛鍏ㄤ粨绱㈠紩銆佸啀娣卞銆佸啀鏈€灏忎慨琛ャ€佸啀钀?worklog/memory
  - backup 鍘嗗彶 worklog锛氱‘璁ょ粍浠朵腑宸插瓨鍦ㄩ殣钘忚嫳鏂囨祴璇曢敋鐐癸紝鍙互閫氳繃鏄惧紡 locale 瀵归綈鑴氭湰鑰屼笉蹇呮敼 UI 榛樿鏂囨
- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛歚鏈娇鐢╜
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛歚N/A`
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛?+- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鏀瑰姩浠呴檺 no-browser 楠岃瘉閾撅紝鏃犻渶棰濆 browser smoke
- 寰呴獙璇侀〉闈㈡竻鍗曪細鐪熷疄娴忚鍣ㄤ笅鐨勮缃〉甯姪鎻愮ず銆乣CC Switch`銆佹笭閬?鍒嗙粍寮圭獥銆乣375px` 甯冨眬浠嶅緟缁х画
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細閬靛惊鏈疆鈥滀笉瑕佸垱寤哄瓙 agent鈥濈殑鏄庣‘瑕佹眰
- worklog 鏄惁鏇存柊锛歚鏄痐
- 閬楃暀椤癸細providers 杩滅婧愪粛渚濊禆 mutable GitHub branch raw URL锛汻EADME 鍗犱綅绗﹀拰 `web/package.json` 鐨?`devs` 纭紪鐮佽矾寰勪粛寰呮敹鍙ｏ紱Phase G 娴忚鍣ㄧ骇璇佹嵁浠嶆湭瀹屾垚
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛歚婊¤冻`
