# 2026-04-24 Phase G CC Switch 娴忚鍣ㄨ瘉鎹敋鐐硅ˉ榻愪笌瀹夸富 blocker 璁板綍

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛歚CC Switch` 娴忚鍣ㄨ瘉鎹敋鐐硅ˉ榻愪笌 smoke 璺緞鏀跺彛
- 鏃ユ湡锛歚2026-04-24`
- 褰撳墠闃舵锛歚Phase G` 鎴浘浼樺厛 UI 涓荤嚎
- 瀵瑰簲 milestone锛歚P0-B` 鍚屼富绾挎祻瑙堝櫒璇佹嵁闂幆灏濊瘯

## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): `yes`
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?`9.6`銆乣14`銆乣16` 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?`1.0`銆乣1.2`銆乣1.3`銆乣1.4` 鑺?+- 涓婁竴涓浉鍏?worklog锛?+  - `docs/worklog/2026-04-23-phase-g-ccswitch-375px-and-verification-sync.md`
  - `docs/worklog/2026-04-24-phase-g-homepage-layout-browser-smoke-closure.md`
- 鏈浠诲姟鐩爣锛?+  - 缁?`DocModal / CC Switch` 琛ラ綈绋冲畾 `data-testid` 閿氱偣锛岄伩鍏嶆祻瑙堝櫒 smoke 缁х画渚濊禆鑴嗗急鏂囨鍖归厤銆?+  - 鍦ㄧ幇鏈夊叡浜?smoke 璺緞涓婅ˉ `ccswitch` 鍦烘櫙锛屽苟鍚屾鎺㈢储 CLI browser smoke 澶嶇敤璺緞銆?+  - 鑻ユ祻瑙堝櫒瀹夸富閾捐矾浠嶉樆濉烇紝鑷冲皯鐣欎笅鍙鐢ㄨ剼鏈笌鏄庣‘ blocker锛屼笉鎶婃湭璺戦€氱殑娴忚鍣ㄨ瘉鎹鎴愬畬鎴愩€?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-ccswitch-375px-and-verification-sync.md`
  - `docs/worklog/2026-04-24-phase-g-homepage-layout-browser-smoke-closure.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/navbar/DocModal.tsx`
  - `web/src/components/modules/navbar/navbar.tsx`
  - `scripts/verify-channel-create-browser-smoke-cdp.mjs`
  - `scripts/verify-channel-create-browser-smoke.ps1`
  - `scripts/verify-channel-create-browser-smoke.mjs`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-ccswitch-flow.mjs`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細canonical plan銆佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佺幆澧?next plan銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?`CC Switch` worklog銆佷笂涓€杞?homepage browser smoke worklog銆乤utomation memory銆佺幇鏈?channel/home/setting smoke 鑴氭湰婧愮爜
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鍙敹鍙?`CC Switch` 鍚屾睜娴忚鍣ㄨ瘉鎹紝涓嶆墿鏁ｅ埌澶囦唤銆佹ā鍨嬮〉鎴栧悗绔笟鍔￠€昏緫
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細`鍚
- 鏈瀛?agent 浣跨敤妯″瀷锛歚N/A`
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛歚鍚
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔℃槸鍓嶇 / smoke / 璁板綍涓夎€呭己鑰﹀悎鐨勫皬闂幆

## 3. 鏈纭鍒?+
- 鍙鐞?`CC Switch` 娴忚鍣ㄨ瘉鎹浉鍏崇殑 selector銆乻moke 鑴氭湰鍜岃褰曟枃浠讹紝涓嶆敼娣遍摼鍗忚鍜屽悗绔绾︺€?+- 娴忚鍣ㄨ瘉鎹嫢鏈窇閫氾紝蹇呴』鏄庣‘璁板綍瀹夸富 blocker锛屼笉鑳芥妸鑴氭湰瀛樺湪璇鎴愬畬鎴愩€?+- 缁存寔鍚屼竴 `Phase G` screenshot-first 涓荤嚎锛屼笉鎵╂暎鍒颁笉鐩稿共涓婚銆?+
## 4. 鏈绂佹浜嬮」

- 涓嶅洖婊氫粨搴撲腑宸叉湁鐨勬棤鍏宠剰鏀瑰姩銆?+- 涓嶆敼 `ccswitch://` 娣遍摼缁撴瀯銆佸鍏ュ弬鏁板懡鍚嶆垨鍚庣鎺ュ彛璇箟銆?+- 涓嶆妸瀹夸富鐜闂浼鎴?UI 宸查棴鐜€?+
## 5. 鏈楠屾敹鏉′欢

- `DocModal / CC Switch` 鍏峰绋冲畾娴忚鍣ㄩ敋鐐癸紝鍙緵 smoke 鐩存帴瀹氫綅銆?+- `scripts/verify-ccswitch-flow.mjs` 涓?`tsc --noEmit` 淇濇寔閫氳繃銆?+- 娴忚鍣?smoke 鑻ユ垚鍔燂紝闇€瑕嗙洊妗岄潰绔€乫ocus/hover 鍜?`375px`锛涜嫢澶辫触锛岄渶鏄庣‘璁板綍瀹夸富 blocker 涓庡鐜拌瘉鎹€?+
## 6. 鏈鍥炴粴鐐?+
- `web/src/components/modules/navbar/navbar.tsx`
- `web/src/components/modules/navbar/DocModal.tsx`
- `scripts/verify-channel-create-browser-smoke-cdp.mjs`
- `scripts/verify-channel-create-browser-smoke.ps1`
- `scripts/verify-ccswitch-browser-smoke.ps1`
- `scripts/verify-ccswitch-browser-smoke-cli.ps1`
- `scripts/verify-ccswitch-browser-smoke.mjs`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛琛?UI selector锛屽啀鎺?smoke
- 鍙楀奖鍝嶅悗绔ā鍧楋細`鏃燻
- 鍙楀奖鍝嶅墠绔ā鍧楋細`navbar / DocModal / CC Switch`
- 鍙楀奖鍝嶆帴鍙ｏ細`鏃燻
- 鏄惁褰卞搷鏃ф暟鎹細`鍚
- 鏄惁褰卞搷鏃ц涓猴細`鍚锛涗粎澧炲姞 selector 涓庨獙璇佽剼鏈紝涓嶆敼涓氬姟璇箟

## 8. 瀹炴柦姝ラ

1. 澶嶆牳涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?`CC Switch` worklog 涓?automation memory锛岀‘璁ゆ湰杞户缁暀鍦?`Phase G` 鍚屼竴 screenshot-first 姹犮€?+2. 鍦?`navbar / DocModal` 涓负鏂囨。鍏ュ彛銆佸脊绐楁湰浣撱€乣CC Switch` 鍏抽敭鍖哄潡琛ョǔ瀹?`data-testid`銆?+3. 鍦ㄥ叡浜?CDP smoke 涓柊澧?`ccswitch` 鍦烘櫙锛屽苟澧炲姞鐙珛 wrapper 鍏ュ彛銆?+4. 鍥犺嚜鍚姩 CDP 浠嶅崱鍦ㄥ涓?Edge 杩滅▼璋冭瘯鍚姩灞傦紝鍐嶈浆鍚?CLI browser smoke 澶嶇敤璺緞锛屾柊澧?`CC Switch` CLI smoke 鑴氭湰涓?wrapper銆?+5. 楠岃瘉闈欐€佽剼鏈拰 TypeScript锛岄€氳繃鍚庡娴忚鍣ㄩ摼鍋氭渶灏忓鐜帮紝纭褰撳墠瀹夸富瀛樺湪 `Node child_process.spawn(...)=EPERM` blocker銆?+6. 鎶婂凡瀹屾垚澧為噺銆侀樆濉炲師鍥犮€佸鐜板懡浠ゅ拰涓嬩竴杞叆鍙ｅ啓鍥?worklog銆佸墠绔富绾跨姸鎬佷笌 automation memory銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 宸查€氳繃锛?+  - `node scripts/verify-ccswitch-flow.mjs`
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode check-only`
  - `node scripts/verify-ccswitch-browser-smoke.mjs --check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke-cli.ps1 -Mode check-only`
- 宸叉墽琛屼絾澶辫触锛?+  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -EdgeLaunchPreset relaxed -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ccswitch-browser-smoke-cli.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 240 -KeepArtifacts`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cli -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`
  - `@' ... spawn(process.execPath, ['-e', 'console.log("child-ok")']) ... '@ | node -`
  - `D:\gol1\node.exe D:\gol1\node_modules\npm\bin\npx-cli.js -y @playwright/cli -s=octopus-cli-open-direct open http://127.0.0.1:18081 --browser msedge --profile <temp>`
- 涓撻」楠岃瘉缁撹锛?+  - 褰撳墠瀹夸富鐜閲岋紝Node 杩涚▼鍐呮渶灏?`child_process.spawn()` 涔熶細鐩存帴鎶?`EPERM`銆?+  - 鍥犳 `@playwright/cli` 澶辫触涓嶆槸 `CC Switch` 閫昏緫闂锛岃€屾槸鏇翠笂灞傜殑瀹夸富杩涚▼鍒涘缓 blocker銆?+  - CDP 鑷惎鍔ㄩ摼鍚屾牱鍙楀涓?Edge 杩滅▼璋冭瘯鍚姩涓嶇ǔ瀹氬奖鍝嶏紝浠嶆湭鎭㈠鍒板彲浣滀负涓昏瘉鎹殑鐘舵€併€?+
## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏂板 selector 鍜?smoke 鑴氭湰涓嶄細鏀瑰彉涓氬姟琛屼负銆?+- 鍏煎鎬ч闄╋細浣庯紱鍏变韩 CLI wrapper 鏂板浜嗚剼鏈笌鎴愬姛鏍囪鐨勭幆澧冨彉閲忓叆鍙ｏ紝榛樿鍊间繚鎸?channel create 鍘熻涓恒€?+- 鏄惁闃诲涓嬩竴浠诲姟锛歚閮ㄥ垎闃诲`锛涢樆濉炵殑鏄祻瑙堝櫒绾т富璇佹嵁锛屼笉闃诲鍚屾睜 no-browser / 鏂囨 / selector 灏忛棴鐜€?+
## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚閫氳繃锛圱ypeScript noEmit锛塦
- 娴嬭瘯鏄惁閫氳繃锛歚閫氳繃锛坣o-browser / check-only锛夛紱娴忚鍣?self-start 鏈€氳繃`
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佺幆澧?next plan銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?`CC Switch` worklog銆乭omepage browser smoke closure銆乤utomation memory銆佹棦鏈?browser smoke 鑴氭湰婧愮爜
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 涓昏鍒?/ 鐢ㄦ埛涓婁笅鏂囨€昏处 / 鍓嶇涓荤嚎鐘舵€佷竴鑷磋姹?`CC Switch` 缁х画琛ユ祻瑙堝櫒绾?`hover / focus / 375px`銆?+  - homepage browser smoke closure 璇佹槑鍏变韩 smoke 鍦烘櫙鎵╁睍鏂瑰紡鍙鐢紝鍥犳鏈疆鍏堟寜鍚屾ā寮忚ˉ `ccswitch` 鍦烘櫙銆?+  - setting-help / channel-create CLI smoke 婧愮爜鎻愪緵浜嗗彲澶嶇敤鐨勯敭鐩?focus銆佺湡瀹?hover 鍜?`375px` 楠岃瘉鍐欐硶銆?+  - automation memory 鏄庣‘鏈疆鏈€鍊煎緱鎺ㄨ繘鐨勬槸 `CC Switch` 娴忚鍣ㄨ瘉鎹€?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛歚鏈娇鐢╜
- 鎵嬪伐 smoke 鐘舵€侊細`鏈畬鎴恅
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細褰撳墠瀹夸富鐜瀛樺湪 `Node child_process.spawn(...)=EPERM` blocker锛汵ode 杩涚▼鍐呮渶灏忓鐜板凡澶辫触锛屽鑷?`@playwright/cli open` 鏃犳硶鍚姩娴忚鍣ㄤ細璇濄€傚悓鏃讹紝CDP 鑷惎鍔ㄩ摼浠嶅彈瀹夸富 Edge 杩滅▼璋冭瘯鍚姩涓嶇ǔ瀹氬奖鍝嶃€?+- 寰呴獙璇侀〉闈㈡竻鍗曪細
  - `DocModal` 鐨?`CC Switch` 椤电鐪熷疄妗岄潰绔氦浜?+  - `CC Switch` 鐨勯敭鐩?focus / 鐪熷疄 hover tooltip
  - `CC Switch` 鐨?`375px` 绉诲姩绔竷灞€涓庡鍏ユ寜閽彲杈炬€?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞换鍔¤寖鍥村皬銆佷笂涓嬫枃寮鸿€﹀悎
- worklog 鏄惁鏇存柊锛歚鏄痐
- 閬楃暀椤癸細
  - `CC Switch` 娴忚鍣ㄧ骇涓昏瘉鎹粛鏈棴鐜?+  - 瀹夸富鏈?`Node spawn EPERM` 闇€瑕佸崟鐙仮澶嶏紝鍚﹀垯 CLI browser smoke 鏃犳硶鍚姩
  - 瀹夸富鏈?Edge/CDP 鑷惎鍔ㄤ笉绋冲畾浠嶄繚鐣欎负骞宠 blocker
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛歚婊¤冻`锛涘彲缁х画鎺掑涓荤幆澧?blocker锛屾垨鍒囧埌鍚屾睜涓嶄緷璧栨祻瑙堝櫒浼氳瘽鐨勭浉閭讳换鍔?+
## 12. 涓嬩竴杞缓璁?+
- 涓嬩竴杞渶閫傚悎缁х画鎺ㄨ繘锛?+  1. 鍏堝崟鐙鐞嗗涓?`Node spawn EPERM` blocker锛屾仮澶?`@playwright/cli open` 鏈€灏忛摼璺?+  2. 涓€鏃﹀涓绘仮澶嶏紝浼樺厛鐩存帴閲嶈窇 `scripts/verify-ccswitch-browser-smoke.mjs` / `verify-ccswitch-browser-smoke-cli.ps1`
  3. 鑻ュ涓?blocker 鏆傛椂鏃犳硶蹇€熻В闄わ紝鍒欏垏鍥炲悓涓€ `Phase G` screenshot-first 姹犻噷涓嶄緷璧栨祻瑙堝櫒浼氳瘽鐨勭浉閭讳换鍔★紝渚嬪鏇存繁灞備腑鏂囩晫闈㈣嫳鏂囨硠婕忔竻鐞嗘垨 key 绾ц娴?copy / 缁撴瀯鏀跺彛
- 鍚屼富绾垮€欓€夐『搴忥細
  1. 瀹夸富 `Node spawn EPERM` 淇 -> `CC Switch` CLI browser smoke 閲嶈窇
  2. 瀹夸富 Edge/CDP 鑷惎鍔ㄦ仮澶?-> `CC Switch` CDP browser smoke 閲嶈窇
  3. 璁剧疆 / 澶囦唤 / 璇︽儏鎬佹洿娣卞眰涓枃涓绘樉绀烘竻鐞?+  4. key 绾ц娴嬪瓧娈典笌缁撴瀯鏀跺彛
