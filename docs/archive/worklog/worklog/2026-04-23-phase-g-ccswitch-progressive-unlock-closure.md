# 2026-04-23 Phase G CC Switch 娓愯繘瑙ｉ攣鏀跺彛

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛歚CC Switch` 鍚岄〉娓愯繘瑙ｉ攣鏀跺彛
- 鏃ユ湡锛歚2026-04-23`
- 褰撳墠闃舵锛歚Phase G` 鍥剧墖闂浼樺厛杩斿伐绐楀彛
- 瀵瑰簲 milestone锛氭埅鍥鹃棶棰樻睜 / `CC Switch` 娴佺▼鍨嬩氦浜掗棴鐜?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): `yes`
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?`9.6`銆乣14`銆乣16` 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?`1.0`銆乣1.2`銆乣1.3`銆乣1.4` 鑺?+- 涓婁竴涓浉鍏?worklog锛?+  - `docs/worklog/2026-04-22-phase-g-ccswitch-progressive-help-closure.md`
  - `docs/worklog/2026-04-22-phase-g-ccswitch-traditional-chinese-copy-closure.md`
  - `docs/worklog/2026-04-23-phase-g-setting-help-attached-session-timeout-order-contrast.md`
- 鏈浠诲姟鐩爣锛?+  - 鎶?`CC Switch` 浠庘€滃彧鏈夋楠ゆ枃妗堚€濇帹杩涘埌鈥滃瓧娈电湡瀹炴寜姝ラ瑙ｉ攣鈥濈殑鐘舵€併€?+  - 璁╂祦绋嬬湡姝ｆ敹绱т负鈥滃厛閫夊鎴风 -> 鍐嶉€?API 瀵嗛挜 -> 鍐嶉€変富妯″瀷 -> 鍐嶇‘璁ゅ悕绉?-> 鏈€鍚庢寜闇€灞曞紑 Claude 楂樼骇鏄犲皠鈥濄€?+  - 鐢ㄩ潤鎬佽剼鏈攣浣忔柊鐨勬笎杩涜В閿佺害鏉燂紝骞跺悓姝ヨˉ榻愬洓璇枃妗堛€?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - automation memory `C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/navbar/DocModal.tsx`
  - `scripts/verify-ccswitch-flow.mjs`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/ja.json`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?automation memory銆佸綋鍓?`DocModal` 涓?`CC Switch` 闈欐€侀獙璇佽剼鏈?+- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鍙鐞?`CC Switch` 鍚岄〉浜や簰闂幆锛屼笉闇€瑕佺户缁睍寮€鍚庣鎴栧浠戒富绾?+- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細`鍚
- 鏈瀛?agent 浣跨敤妯″瀷锛歚鏃燻
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛歚鍚
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛歚鏃燻
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞槸鍗曠粍浠?+ 鍗曡剼鏈?+ locale 鐨勫皬闂幆

## 3. 鏈纭鍒?+
- 鍙鐞?`CC Switch` 鍚岄〉娴佺▼瑙ｉ攣锛屼笉鏀规繁閾惧弬鏁拌涔夛紝涓嶆墿鏁ｅ埌鍏朵粬鎴浘涓婚銆?+- 蹇呴』瀵归綈闇€姹?`09`銆乣12`銆乣19`銆乣20`銆乣27`銆乣32`銆乣42`銆乣43`銆?+- 蹇呴』鐣欎笅鍙噸澶嶆墽琛岀殑鏃犳祻瑙堝櫒楠岃瘉锛屼笉鎶婄函鏂囨璋冩暣鍐欐垚闂幆瀹屾垚銆?+
## 4. 鏈绂佹浜嬮」

- 涓嶆敼 `ccswitch://` 娣遍摼缁撴瀯涓庡弬鏁板懡鍚嶃€?+- 涓嶆敼鍚庣鎺ュ彛銆佹潈闄愭ā鍨嬫垨 API key 杩囨护璇箟銆?+- 涓嶆妸鏈墽琛岀殑娴忚鍣ㄧ骇 smoke 浼鎴愬凡瀹屾垚銆?+
## 5. 鏈楠屾敹鏉′欢

- 鏈€?API 瀵嗛挜鏃讹紝涓绘ā鍨嬮€夋嫨浣嶅浜庨攣瀹氱姸鎬佸苟鏄剧ず閿佸畾鎻愮ず銆?+- 鏈€変富妯″瀷鏃讹紝鍚嶇О杈撳叆鍖轰笉鐩存帴灞曞紑锛岃€屾槸鏄剧ず涓嬩竴姝ユ彁绀恒€?+- Claude 鐨勯珮绾ф槧灏勫彧鏈夊湪涓绘ā鍨嬪拰鍚嶇О閮藉氨缁悗鎵嶅厑璁稿睍寮€銆?+- `node scripts/verify-ccswitch-flow.mjs`銆乣node scripts/verify-locale-consistency.mjs` 涓?`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 閫氳繃銆?+
## 6. 鏈鍥炴粴鐐?+
- `web/src/components/modules/navbar/DocModal.tsx`
- `scripts/verify-ccswitch-flow.mjs`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/ja.json`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鏀?UI 娓愯繘瑙ｉ攣锛屽啀琛ヨ剼鏈拰 locale 绾︽潫
- 鍙楀奖鍝嶅悗绔ā鍧楋細`鏃燻
- 鍙楀奖鍝嶅墠绔ā鍧楋細`DocModal / CC Switch` 寮圭獥
- 鍙楀奖鍝嶆帴鍙ｏ細`鏃燻
- 鏄惁褰卞搷鏃ф暟鎹細`鍚
- 鏄惁褰卞搷鏃ц涓猴細`浠呮敹绱у瓧娈靛睍绀洪『搴忎笌閿佸畾鎻愮ず锛屼笉鏀瑰彉鏈€缁堝鍏ラ摼鎺ヨ涔塦

## 8. 瀹炴柦姝ラ

1. 澶嶆牳 canonical銆佺敤鎴蜂笂涓嬫枃鎬昏处銆佸墠绔富绾跨姸鎬佷笌涓婅疆 memory锛岀‘璁?`CC Switch` 浠嶆槸 Phase G screenshot-first 姹犵殑涓嬩竴椤广€?+2. 鍦?`DocModal.tsx` 涓柊澧?`ccswitchCanChooseModel / ccswitchCanConfirmName / ccswitchCanOpenAdvanced` 涓夊眰闂ㄦ帶锛屾妸涓绘ā鍨嬨€佸悕绉板拰 Claude 楂樼骇鏄犲皠鏀逛负鐪熷疄鎸夋楠よВ閿併€?+3. 涓洪攣瀹氭€佽ˉ榻愬洓璇彁绀烘枃妗堬紝骞跺寮?`scripts/verify-ccswitch-flow.mjs`锛屾妸闂ㄦ帶閫昏緫鍜屾柊澧?locale key 鍥哄寲涓嬫潵銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛?+  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 娴嬭瘯鍛戒护锛?+  - `node scripts/verify-ccswitch-flow.mjs`
  - `node scripts/verify-locale-consistency.mjs`
- 涓撻」楠岃瘉锛?+  - `DocModal` 涓?`CC Switch` 鐨勪富妯″瀷閫夋嫨浣嶅甫 `disabled={!ccswitchCanChooseModel}`
  - 鍚嶇О鍖哄煙鎸?`ccswitchCanConfirmName` 鍦ㄨ緭鍏ユ鍜岄攣瀹氭彁绀轰箣闂村垏鎹?+  - Claude 楂樼骇鏄犲皠鎸?`ccswitchCanOpenAdvanced` 鍦?Accordion 涓庨攣瀹氭彁绀轰箣闂村垏鎹?+
## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏈疆鍙皟鏁村墠绔笎杩涘睍绀洪『搴忓拰甯姪鏂囨
- 鍏煎鎬ч闄╋細浣庯紱娣遍摼鍙傛暟鍜?API key 杩囨护閫昏緫鏈敼
- 鏄惁闃诲涓嬩竴浠诲姟锛歚鍚

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚閫氳繃`
- 娴嬭瘯鏄惁閫氳繃锛歚閫氳繃`
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€乤utomation memory銆乣DocModal` 瀹炵幇銆乣verify-ccswitch-flow.mjs`
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 鐢ㄦ埛涓婁笅鏂囨€昏处鏄庣‘瑕佹眰 `CC Switch` 蹇呴』浠庡崐鎴愬搧琛ㄥ崟鏀跺彛涓烘祦绋嬪瀷浜や簰銆?+  - 鍓嶇涓荤嚎鐘舵€佷笌 automation memory 鏄庣‘鎸囧嚭褰撳墠 browser-smoke 涓荤嚎鏆傛椂鍙楀涓?Edge/CDP 闃诲锛屽洜姝ら€傚悎鍏堝仛鍚屼富绾跨殑 no-browser 缁撴瀯闂幆銆?+  - 鐜版湁 `DocModal` 宸叉湁姝ラ鎽樿涓庢繁閾鹃€昏緫锛屽洜姝ゆ湰杞渶鍚堥€傜殑鏄敹绱у瓧娈佃В閿侀『搴忥紝鑰屼笉鏄噸鍋氭暣涓ā鍧椼€?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛歚鏃燻
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛歚鏃燻
- 鎵嬪伐 smoke 鐘舵€侊細`鏈墽琛屾祻瑙堝櫒绾ф墜宸?smoke`
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細褰撳墠 screenshot-first 娴忚鍣ㄥ疄璇佷粛鍙楁湰鏈?Edge/CDP bootstrap 闃诲锛宍CC Switch` 鐨?hover/focus 涓?`375px` 璇佹嵁闇€绛夊悓涓荤嚎娴忚鍣ㄨ矾寰勬仮澶嶅悗鍐嶈ˉ
- 寰呴獙璇侀〉闈㈡竻鍗曪細
  - `DocModal` 鐨?`CC Switch` 椤电妗岄潰绔湡瀹炰氦浜?+  - `DocModal` 鐨?`CC Switch` 椤电 `375px` 瑙嗗浘
  - `CC Switch` 閿佸畾鎬佷笌楂樼骇鏄犲皠灞曞紑鐨?hover/focus 鐪熷疄娴忚鍣ㄨ瘉鎹?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔¤寖鍥村皬銆佷笂涓嬫枃寮鸿€﹀悎
- worklog 鏄惁鏇存柊锛歚鏄痐
- 閬楃暀椤癸細
  - `CC Switch` 娴忚鍣ㄧ骇 smoke 涓?`375px` 鎴浘璇佹嵁浠嶆湭琛?+  - 娓犻亾鍒涘缓 / 鍒嗙粍鍒涘缓寮圭獥涓?help-hint hover/focus 浠嶅湪鍚屼竴 screenshot-first 鍊欓€夋睜涓?+- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛歚婊¤冻`

## 12. 涓嬩竴杞缓璁?+
- 涓嬩竴杞渶閫傚悎缁х画鎺ㄨ繘锛歚CC Switch` 鐨勬祻瑙堝櫒绾?hover/focus 涓?`375px` 璇佹嵁锛涜嫢瀹夸富娴忚鍣ㄩ樆濉炰粛鏈В闄わ紝鍒欏垏鎹㈠埌鍚屾睜鐨勬笭閬?鍒嗙粍鍒涘缓寮圭獥娴忚鍣ㄨ瘉鎹垨缁х画鍙潤鎬侀棴鐜殑 `help-hint hover/focus` 鏂█寮哄寲
- 鍚屼富绾垮€欓€夐『搴忥細
  1. `CC Switch` 娴忚鍣ㄧ骇 smoke / `375px` 瀹炶瘉
  2. 娓犻亾鍒涘缓寮圭獥娴忚鍣ㄧ骇璇佹嵁
  3. 鍒嗙粍鍒涘缓寮圭獥娴忚鍣ㄧ骇璇佹嵁
  4. `help-hint` hover/focus 娴忚鍣ㄧ骇璇佹嵁
