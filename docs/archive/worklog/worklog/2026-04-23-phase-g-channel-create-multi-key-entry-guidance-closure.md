# 2026-04-23 Phase G 娓犻亾鍒涘缓澶?Key 杈撳叆寮曞鏀跺彛

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氬垱寤烘笭閬撳脊绐楀 Key 杈撳叆浣嶅紩瀵兼敹鍙?+- 鏃ユ湡锛?026-04-23
- 褰撳墠闃舵锛歅hase G 鎴浘浼樺厛 UI 涓荤嚎
- 瀵瑰簲 milestone锛歅0 娓犻亾鍒涘缓寮圭獥澶?Key 鍙悊瑙ｆ€цˉ寮?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?`9.1`銆乣9.1.1`銆乣14`銆乣16` 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?`1.2`銆乣1.3`銆乣1.4`銆乣11.2` 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-22-phase-g-channel-create-layered-guidance-closure.md`
- 鏈浠诲姟鐩爣锛?+  - 鎶婂垱寤烘笭閬撳脊绐楀 Key 鍖洪粯璁ゆ姌鍙犳€佽ˉ鎴愨€滃厛灞曞紑鍗＄墖锛屽啀鍦ㄧ涓€涓緭鍏ユ濉啓鐪熷疄 API 瀵嗛挜鈥濈殑鏄庣‘寮曞銆?+  - 淇濇寔澶?Key 榛樿鎶樺彔鍜屽師鏈夋繁缁垮渾瑙掗鏍间笉鍙橈紝涓嶆墿鏁ｅ埌鍏朵粬鎴浘涓婚銆?+  - 璁╂棤娴忚鍣ㄩ獙璇佽剼鏈浐瀹氳繖鏉″紩瀵奸摼锛岄伩鍏嶄笅杞洖閫€鎴愨€滅湅寰楄澶?Key锛屼絾浠嶄笉鐭ラ亾 key 濉摢鈥濄€?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - automation memory `C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/channel/Form.tsx`
  - `scripts/verify-channel-create-flow.mjs`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `web/public/locale/en.json`
  - `web/public/locale/ja.json`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?channel create worklog銆乤utomation memory銆佸綋鍓嶈〃鍗曚笌鑴氭湰婧愮爜
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞槸鍚屼竴妯″潡鐨勫皬闂幆

## 3. 鏈纭鍒?+
- 鍙鐞嗗垱寤烘笭閬撳脊绐椾腑澶?Key 杈撳叆浣嶇殑鍙悊瑙ｆ€э紝涓嶆敼鍚庣濂戠害鍜屼繚瀛樿涔夈€?+- 榛樿鐣岄潰缁х画淇濇寔鎶樺彔涓庣畝娲侊紝涓嶉噸鏂版憡寮€楂樼骇瀛楁銆?+- 鏈疆蹇呴』鐣欎笅鍙噸澶嶆墽琛岀殑鏃犳祻瑙堝櫒楠岃瘉锛屼笉鎶婃枃妗堝井璋冨啓鎴愭湭楠岃瘉瀹屾垚銆?+
## 4. 鏈绂佹浜嬮」

- 涓嶆墿鏁ｅ埌鍒嗙粍寮圭獥銆佹ā鍨嬩环鏍煎尯銆佸浠介〉鎴?`CC Switch`銆?+- 涓嶅洖婊氬伐浣滃尯涓凡鏈夌殑鏃犲叧淇敼銆?+- 涓嶆妸鏈墽琛岀殑鐪熷疄娴忚鍣?smoke 璁颁负宸插畬鎴愩€?+
## 5. 鏈楠屾敹鏉′欢

- 澶?Key 鎬诲尯鍑虹幇鈥滃厛灞曞紑鍗＄墖锛屽啀鍦ㄧ涓€涓緭鍏ユ濉湡瀹?API 瀵嗛挜鈥濈殑鏄庣‘鎻愮ず銆?+- 灞曞紑鍚庣殑瀵嗛挜杈撳叆鍖哄嚭鐜颁竴鏉′紭鍏堢骇鏇撮珮鐨?lead hint锛屾槑纭湡瀹?API 瀵嗛挜搴斿厛濉啓銆?+- `scripts/verify-channel-create-flow.mjs`銆乣scripts/verify-locale-consistency.mjs` 鍜?`web` TypeScript 妫€鏌ラ€氳繃銆?+
## 6. 鏈鍥炴粴鐐?+
- `web/src/components/modules/channel/Form.tsx`
- `scripts/verify-channel-create-flow.mjs`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `web/public/locale/en.json`
- `web/public/locale/ja.json`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛琛?UI 寮曞锛屽啀鍚屾鑴氭湰鍜?locale
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細鍒涘缓娓犻亾 / 缂栬緫娓犻亾寮圭獥涓殑澶?Key 鍖?+- 鍙楀奖鍝嶆帴鍙ｏ細鏃?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細浠呭寮哄 Key 杈撳叆寮曞锛屼笉鏀瑰彉淇濆瓨缁撴瀯鎴栧瓧娈靛惈涔?+
## 8. 瀹炴柦姝ラ

1. 澶嶆牳 `ChannelForm` 褰撳墠澶?Key 缁撴瀯锛岀‘璁も€滃 Key 瀛樺湪浣嗙敤鎴蜂粛鍙兘涓嶇煡閬?key 濉摢鈥濈殑缂哄彛杩樺湪銆?+2. 鍦ㄥ Key 鍖烘爣棰樹笅鏂板鎬绘彁绀猴紝鏄庣‘鈥滃厛灞曞紑鍗＄墖锛屽啀鍦ㄧ涓€涓緭鍏ユ濉湡瀹?API 瀵嗛挜鈥濄€?+3. 鍦ㄥ睍寮€鍚庣殑 `keyValue` 杈撳叆妗嗗墠鏂板 lead hint锛屾槑纭湡瀹炲瘑閽ヤ紭鍏堝～鍐欓『搴忋€?+4. 鏇存柊鍥涜 locale锛屽苟鍚屾澧炲己 `verify-channel-create-flow.mjs` 瀵规柊鎻愮ず鐨勬柇瑷€銆?+5. 杩愯鏈疆涓夋潯鐩存帴楠岃瘉锛岀‘璁ゆ棤娴忚鍣ㄩ棴鐜垚绔嬨€?+
## 9. 娴嬭瘯涓庨獙璇?+
- `node scripts/verify-channel-create-flow.mjs`
- `node scripts/verify-locale-consistency.mjs`
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏈疆鍙柊澧炲紩瀵兼枃妗堝拰鑴氭湰鏂█
- 鍏煎鎬ч闄╋細浣庯紱鏈慨鏀?API銆佷繚瀛樻暟鎹拰寮圭獥浜や簰涓昏矾寰?+- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氶€氳繃锛圱ypeScript noEmit锛?+- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?channel create worklog銆乤utomation memory銆乣ChannelForm` 涓庨獙璇佽剼鏈簮鐮?+- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 鐢ㄦ埛涓婁笅鏂囨€昏处涓殑闇€姹?`46`銆乣47` 浠嶈姹傚悓娓犻亾澶?Key 蹇呴』璁┾€滃摢涓湴鏂瑰～ key鈥濅竴鐪肩湅鎳傘€?+  - 鍓嶇涓荤嚎鐘舵€佸拰涓婁竴杞?channel create worklog 璇存槑澶?Key 缁撴瀯宸茬粡鎴愬瀷锛屾湰杞渶閫傚悎琛ヨ緭鍏ヤ綅寮曞锛岃€屼笉鏄噸鍐欑粨鏋勩€?+  - automation memory 鎻愰啋娴忚鍣ㄧ骇璇佹嵁浠嶅彈瀹夸富闄愬埗锛屽洜姝ゆ湰杞紭鍏堟敹鍙ｅ彲楠岃瘉鐨?no-browser 灏忓垏鐗囥€?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛岀湡瀹炴祻瑙堝櫒 smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆浼樺厛鍋氭簮鐮佷笌闈欐€侀獙璇侀棴鐜紱娴忚鍣ㄧ骇 `375px` 涓庣湡瀹炵偣鍑昏矾寰勪粛寰呭悓姹犲悗缁ˉ璇?+- 寰呴獙璇侀〉闈㈡竻鍗曪細鍒涘缓娓犻亾寮圭獥妗岄潰绔€乣375px` 涓嬪 Key 灞曞紑/鏀惰捣銆佺湡瀹炶緭鍏ヨ矾寰?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細
  - 鍒涘缓娓犻亾寮圭獥娴忚鍣ㄧ骇 `375px` 璇佹嵁浠嶆湭琛ラ綈
  - 鍚屼竴鎴浘姹犱腑鐨?channel/group create dialog 娴忚鍣ㄨ瘉鎹粛鍙户缁帹杩?+- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒?+
## 12. 涓嬩竴杞缓璁?+
- 涓嬩竴杞渶閫傚悎缁х画鎺ㄨ繘锛氬湪鍚屼竴 `Phase G` 涓荤嚎涓嬶紝浼樺厛琛ュ垱寤烘笭閬撳脊绐楃湡瀹炴祻瑙堝櫒 `375px` 璇佹嵁锛涜嫢娴忚鍣ㄧ骇楠岃瘉浠嶅彈闃伙紝鍒欏垏鍒板悓姹犵殑 group create dialog 鎴?help-hint hover/focus 鏀跺彛
- 鍚屼富绾垮€欓€夐『搴忥細
  1. 鍒涘缓娓犻亾寮圭獥娴忚鍣ㄧ骇 `375px` / 澶?Key 灞曞紑璇佹嵁
  2. 鍒涘缓鍒嗙粍寮圭獥娴忚鍣ㄧ骇璇佹嵁
  3. help-hint hover/focus 娴忚鍣ㄦ垨 no-browser 琛ュ己
  4. 鍚屾睜鍓╀綑涓枃鍖栦笌甯冨眬鍘嬬缉闂
