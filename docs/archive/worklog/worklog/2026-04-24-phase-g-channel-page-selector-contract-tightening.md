# 2026-04-24 Phase G 娓犻亾鎬婚〉 selector 鍚堝悓鏀剁揣涓庡涓诲璺戜慨姝?+
## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氭笭閬撴€婚〉 browser smoke selector 鍚堝悓鏀剁揣涓庡涓诲璺戜慨姝?+- 鏃ユ湡锛歚2026-04-24`
- 褰撳墠闃舵锛歚Phase G screenshot-first UI closure`
- 瀵瑰簲 milestone锛歚Phase G / 閲岀▼纰?4 UI 涓庢槗鐢ㄦ€

## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?`9.1 / 9.7 / 12 Phase 7`
- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 涓?screenshot-first / Phase G 鎵ц鍙ｅ緞
- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-24-phase-g-channel-page-browser-smoke-closure.md`
- 鏈浠诲姟鐩爣锛氫笉鎵╂暎椤甸潰缁撴瀯锛屽彧鏀剁揣 `channel-page` browser smoke 鐨?selector 鍚堝悓骞跺鏍稿綋鍓嶅涓讳笂鐨勭湡瀹炴祻瑙堝櫒缁撴灉
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細canonical plan銆乣CURRENT_STATUS_AND_PLAN`銆乣FRONTEND_UI_MAINLINE_STATUS`銆乣ENV_READY_AND_NEXT_PLAN`銆佷笂涓€浠?channel-page worklog銆乣scripts/verify-channel-page-browser-smoke.ps1`銆乣scripts/verify-channel-create-browser-smoke-cdp.mjs`銆乣scripts/verify-channel-presentation.mjs`銆乣web/src/components/modules/toolbar/index.tsx`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓婅堪鏂囨。銆佽剼鏈笌褰撳墠 automation memory
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鍙鐞嗘笭閬撴€婚〉鍚屾睜鑴氭湰鍚堝悓涓庡涓诲璺戯紝涓嶅睍寮€鍒?group/backup/AI 鑷姩鍖栦富绾?+- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氫笉閫傜敤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氭槸锛宍octopus-2` 涓荤嚎绋嬭嚜鍔ㄥ寲
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛氫笉閫傜敤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔¤寖鍥村皬銆佷笂涓嬫枃寮鸿€﹀悎

## 3. 鏈纭鍒?+
- 鍙仠鐣欏湪 `Phase G screenshot-first` 鍚屼竴涓荤嚎锛屼笉鎵╂暎鍒板叾瀹冮〉闈㈡垨鍚庣閫昏緫
- 浼樺厛淇楠岃瘉鍚堝悓婕傜Щ锛屽啀澶嶈窇鐪熷疄瀹夸富楠岃瘉
- 涓嶆妸瀹夸富 Edge/CDP 鍚姩澶辫触浼鎴愰〉闈㈢骇閫氳繃

## 4. 鏈绂佹浜嬮」

- 涓嶆敼娓犻亾椤典笟鍔¤涔変笌浜や簰缁撴瀯
- 涓嶉噸鍐欏叡浜?wrapper 鐨勫ぇ娈靛惎鍔ㄩ€昏緫
- 涓嶆妸宸叉湁鈥滃凡闂幆鈥濇枃妗ｇ粨璁虹户缁師鏍峰杩颁负浜嬪疄

## 5. 鏈楠屾敹鏉′欢

- 娓犻亾鎬婚〉宸ュ叿鏍忕瓫閫変笌鍏变韩 CDP smoke 浣跨敤绋冲畾 selector
- `channel-presentation` no-browser 鎶ゆ爮涓?TypeScript 妫€鏌ラ€氳繃
- 鐪熷疄娴忚鍣ㄥ璺戠粨鏋滆鏄庣‘璁板綍涓洪€氳繃鎴栧涓婚樆濉烇紝涓嶈兘鍋滃湪妯＄硦鐘舵€?+
## 6. 鏈鍥炴粴鐐?+
- 鍥炴粴 `web/src/components/modules/toolbar/index.tsx`
- 鍥炴粴 `scripts/verify-channel-create-browser-smoke-cdp.mjs`
- 鍥炴粴 `scripts/verify-channel-presentation.mjs`
- 鍥炴粴鏈鏂板/琛ュ厖鐨?worklog 璁板綍

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鏀?UI 娴嬭瘯閿氱偣锛屽啀鏀瑰叡浜?smoke 鑴氭湰
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/toolbar/index.tsx`
- 鍙楀奖鍝嶆帴鍙ｏ細鏃?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紝涓昏褰卞搷楠岃瘉閫夋嫨鍣ㄤ笌 browser smoke 绋冲畾鎬?+
## 8. 瀹炴柦姝ラ

1. 澶嶆牳 `channel-page` 鍦烘櫙涓庢笭閬撻〉宸ュ叿鏍?璇︽儏鍖虹幇鐘讹紝纭浠嶅瓨鍦ㄤ綅缃瀷 selector 涓庡崱鐗囪璁℃暟婕傜Щ銆?+2. 涓烘笭閬撻〉绛涢€夊叆鍙ｈˉ绋冲畾 `data-testid`锛屽苟鎶婂叡浜?`channel-page` smoke 鍒囧埌绋冲畾 selector锛屽悓鏃朵慨姝ｅ崱鐗囪鏁伴€昏緫銆?+3. 杩愯 no-browser / typecheck / shared-script 妫€鏌ワ紝鍐嶅璺?`channel-page self-start + cdp` 涓庣嫭绔?`bootstrap-edge-cdp.ps1`锛岀‘璁ょ湡瀹為樆濉炵偣銆?+4. 鎶娾€滀唬鐮佸悎鍚屽凡鏀跺彛銆佸涓?browser pass 浠嶅彈闃诲鈥濈殑缁撹鍐欏洖 worklog锛屼緵涓嬩竴杞洿鎺ユ帴缁€?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛歚node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 娴嬭瘯鍛戒护锛?+  - `powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode check-only`
  - `node scripts/verify-channel-presentation.mjs`
  - `node --check scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `node scripts/verify-channel-create-browser-smoke-cdp.mjs --check-only`
  - `node scripts/verify-channel-create-flow.mjs`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-page-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -CdpBootstrapCommandOrder runtime-page-lifecycle -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\bootstrap-edge-cdp.ps1 -Port 9233 -EdgeLaunchPreset relaxed -EdgeProfileStrategy workspace-fixed -ReadyTimeoutSeconds 30 -StableReadySeconds 3 -OutputJsonPath .codex-tmp\edge-cdp-bootstrap.json`
- 涓撻」楠岃瘉锛氱‘璁?`channel-page` 鍦烘櫙涓嶅啀渚濊禆 `button:nth-of-type(...)` 鍜岃緭鍏ラ『搴忥紝涓斿涓诲け璐ョ偣鏄庣‘鍋滅暀鍦?Edge/CDP 鍚姩灞?+
## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏈疆鍙敹绱ч獙璇佸悎鍚?+- 鍏煎鎬ч闄╋細浣庯紱鏂板鐨勬槸娴嬭瘯閿氱偣锛屼笉鏀瑰彉鐢ㄦ埛浜や簰璇箟
- 鏄惁闃诲涓嬩竴浠诲姟锛氫笉闃诲鍚屾睜 no-browser/鏂囨浠诲姟锛屼絾闃诲娓犻亾鎬婚〉 browser-grade pass 鍦ㄥ綋鍓嶅涓讳笂鐨勬敹鍙?+
## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚閫氳繃锛圱ypeScript noEmit锛塦
- 娴嬭瘯鏄惁閫氳繃锛歚閮ㄥ垎閫氳繃锛涗唬鐮?鑴氭湰渚ч獙璇侀€氳繃锛岀湡瀹?browser smoke 浠嶅彈瀹夸富闃诲`
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細canonical plan銆侀樁娈电姸鎬佹枃妗ｃ€佷笂涓€浠?channel-page worklog銆佺浉鍏?smoke 鑴氭湰銆乤utomation memory
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 涓昏鍒掍笌闃舵鐘舵€佹枃妗ｏ細纭鏈疆浠嶅簲鐣欏湪 `Phase G screenshot-first` 椤甸潰绾ц瘉鎹睜
  - 涓婁竴浠?channel-page worklog锛氬彂鐜扳€滃凡闂幆鈥濈粨璁洪渶瑕佸璺戞牎姝?+  - 鐩稿叧鑴氭湰涓庨〉闈唬鐮侊細纭闂涓昏鍦?selector 鍚堝悓婕傜Щ涓庡涓?Edge 鍚姩灞傦紝鑰屼笉鏄笟鍔￠〉闈㈢粨鏋?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細`鏈仛浜哄伐鐐瑰嚮锛涘凡鎵ц鑷姩 browser smoke 澶嶈窇`
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細褰撳墠瀹夸富 Edge 鑷惎鍔ㄤ粛鎶?`鎷掔粷璁块棶 (0x5)`锛宍channel-page self-start + cdp` 鍋滃湪 `json/version` 瓒呮椂涔嬪墠锛屾棤娉曟嬁鍒伴〉闈㈡柇瑷€绾ц瘉鎹?+- 寰呴獙璇侀〉闈㈡竻鍗曪細娓犻亾鎬婚〉 `channel-page` 鐪熷疄 browser-grade pass銆佹笭閬撳垱寤?缂栬緫寮圭獥鏇寸粏 hover/focus 缁嗚妭
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛屼笖鏈疆浠诲姟杈圭晫灏?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細
  - 褰撳墠涓枃鐘舵€佸ぇ鏂囨。閲屼粛鏈夋棫鐨勨€滄笭閬撴€婚〉宸插湪鏈満闂幆鈥濊娉曪紝鍚庣画搴斿湪鍙畨鍏ㄧ紪杈戞椂琛ヤ竴杞悓姝ヤ慨姝?+  - 娓犻亾鎬婚〉鐪熷疄 browser-grade pass 浠嶉渶鍦ㄥ彲鐢ㄥ涓绘垨澶栭儴宸插惎鍔?CDP 浼氳瘽閲岃ˉ榻?+- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒筹紱涓嬩竴杞彲缁х画鐣欏湪鍚屼竴 `Phase G` 姹犻噷琛ュ涓绘棤鍏崇殑灏忛棴鐜紝鎴栧湪鍙敤瀹夸富涓婄洿鎺ヨˉ璺戞笭閬撴€婚〉 browser pass
