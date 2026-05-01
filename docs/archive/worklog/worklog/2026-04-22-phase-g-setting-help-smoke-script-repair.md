# 2026-04-22 Phase G Setting Help Smoke Script Repair

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氳缃〉甯姪鎻愮ず缁熶竴 smoke 鑴氭湰淇涓庤寖鍥磋ˉ榻?+- 鏃ユ湡锛?026-04-22
- 褰撳墠闃舵锛歅hase G 鍥剧墖闂浼樺厛杩斿伐绐楀彛
- 瀵瑰簲 milestone锛氳缃〉甯姪鎻愮ず娴忚鍣ㄩ獙鏀舵敹鍙?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?9.6銆?4銆?6 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?10銆?1 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-dual-mode-prep.md`
- 鏈浠诲姟鐩爣锛?+  - 淇 `scripts/verify-setting-help-browser-smoke.mjs` 涓凡鎹熷潖鐨勫瓧绗︿覆/鏂█锛屾仮澶嶈剼鏈彲鎵ц鎬?+  - 鎶婅缃〉缁熶竴 smoke 鐨勮鐩栬寖鍥翠粠涓夊潡鍗＄墖琛ュ埌鍥涘潡鍗＄墖锛岀撼鍏?`LLMPrice`
  - 鍦ㄥ綋鍓嶄富鏈洪檺鍒朵笅锛屽敖閲忔帹杩涚湡瀹炴祻瑙堝櫒 smoke锛屽苟鏄庣‘鍓╀綑闃诲鐐?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `docs/worklog/2026-04-22-phase-g-llm-price-boundary-closure.md`
  - `docs/worklog/2026-04-22-phase-g-setting-help-browser-smoke-dual-mode-prep.md`
  - `scripts/verify-setting-help-browser-smoke.mjs`
  - `scripts/verify-model-probe-help.mjs`
  - `scripts/verify-llm-price-boundary.mjs`
  - `web/src/components/modules/setting/LLMPrice.tsx`
  - `web/src/components/modules/setting/ModelProbe.tsx`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佸墠绔富绾跨姸鎬併€佹墜宸ラ獙鏀舵竻鍗曘€乤utomation memory銆佽缃〉鍥涘潡鍗＄墖瀹炵幇涓庣幇鏈夎剼鏈?+- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鍙鐞嗚缃〉 smoke 鑴氭湰锛屼笉鎵╂暎鍒板叾瀹冮〉闈㈡垨鍚庣璇箟
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭湭浣跨敤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氬惁
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬩覆琛屾帹杩涳紝涓旀湰杞槸鍗曡剼鏈皬闂幆

## 3. 鏈纭鍒?+
- 鍙慨璁剧疆椤电粺涓€ smoke 鑴氭湰鍜屽悓涓荤嚎楠岃瘉鍏ュ彛锛屼笉鏀逛笟鍔￠€昏緫
- 涓嶆妸鏈疄闄呰窇閫氱殑鐪熷疄娴忚鍣?smoke 鍐欐垚宸查€氳繃
- 蹇呴』鐣欎笅鍙璺戠殑闈欐€侀獙璇佺粨鏋滀笌鏄庣‘闃诲鎻忚堪

## 4. 鏈绂佹浜嬮」

- 涓嶄慨鏀瑰悗绔矾鐢便€佺啍鏂€佹帰娴嬬殑鐪熷疄琛屼负
- 涓嶆墿鏁ｅ埌澶囦唤銆佹笭閬撱€佸垎缁勭瓑鏃犲叧椤甸潰
- 涓嶅洖閫€宸ヤ綔鍖洪噷涓庢湰杞棤鍏崇殑宸叉湁淇敼

## 5. 鏈楠屾敹鏉′欢

- `scripts/verify-setting-help-browser-smoke.mjs` 鎭㈠鍙墽琛?+- 缁熶竴 smoke 鑴氭湰瑕嗙洊 `LLMPrice / DynamicRouting / CircuitBreaker / ModelProbe`
- `node scripts/verify-setting-help-browser-smoke.mjs --check-only` 閫氳繃
- `node scripts/verify-model-probe-help.mjs` 閫氳繃
- `node scripts/verify-llm-price-boundary.mjs` 閫氳繃
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 閫氳繃

## 6. 鏈鍥炴粴鐐?+
- `scripts/verify-setting-help-browser-smoke.mjs`
- `docs/worklog/2026-04-22-phase-g-setting-help-smoke-script-repair.md`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛淇?smoke 鑴氭湰锛屽啀楠岃瘉鐜版湁 UI
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細鏃犱笟鍔￠€昏緫鏀瑰姩锛涘彧璇昏缃〉鍥涘潡鍗＄墖
- 鍙楀奖鍝嶆帴鍙ｏ細鏃犳柊澧炴帴鍙?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紱浠呬慨澶嶅拰澧炲己楠岃瘉鑴氭湰

## 8. 瀹炴柦姝ラ

1. 澶嶆牳璁剧疆椤?smoke 涓荤嚎銆佸墠杞?worklog 鍜?automation memory锛岀‘璁ゆ湰杞紭鍏堜慨鑴氭湰鑰屼笉鏄户缁┖杞湪鑷惎閾句笂銆?+2. 閲嶅啓 `scripts/verify-setting-help-browser-smoke.mjs`锛屼慨鎺夌紪鐮?瀛楃涓叉崯鍧忥紝骞舵妸缁熶竴妫€鏌ヨ寖鍥磋ˉ鍒?`LLMPrice`銆?+3. 杩愯 `check-only`銆佹棤娴忚鍣ㄩ獙璇佸拰鍓嶇 typecheck锛涢殢鍚庡皾璇曠敤 PowerShell 鎵嬪伐涓茶仈鍓嶅悗绔笌 Playwright CLI 鍋氱湡瀹?smoke銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 宸叉墽琛岋細`node scripts/verify-setting-help-browser-smoke.mjs --check-only`
- 宸叉墽琛岋細`node scripts/verify-model-probe-help.mjs`
- 宸叉墽琛岋細`node scripts/verify-llm-price-boundary.mjs`
- 宸叉墽琛岋細`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 宸叉墽琛岋細PowerShell 鍚姩 `build/octopus-smoke.exe` + `next dev` 鐨勮繍琛屾€佸噯澶?+- 宸叉墽琛岋細PowerShell + `playwright-cli` 鐨勭湡瀹?smoke 灏濊瘯

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱浠呮敼楠岃瘉鑴氭湰
- 鍏煎鎬ч闄╋細浣庯紱鏈敼鎺ュ彛銆佹寔涔呭寲鎴栧墠绔笟鍔℃祦绋?+- 鏄惁闃诲涓嬩竴浠诲姟锛氶儴鍒嗛樆濉烇紱鑴氭湰宸叉仮澶嶏紝浣嗗綋鍓嶄富鏈轰笅鐪熷疄娴忚鍣?smoke 浠嶅彈 PowerShell 鍐呰仈 JS 杞箟鍜屽涓荤瓥鐣ラ檺鍒?+
## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氶€氳繃
- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃锛坄check-only`銆乣verify-model-probe-help`銆乣verify-llm-price-boundary`銆佸墠绔?`tsc`锛?+- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佸墠绔富绾跨姸鎬併€佹墜宸ラ獙鏀舵竻鍗曘€乤utomation memory銆佽缃〉鍥涘潡鍗＄墖瀹炵幇銆佺幇鏈?smoke 鑴氭湰
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - automation memory 鏄庣‘鎸囧嚭涓嬩竴浼樺厛椤逛粛鏄缃〉甯姪鎻愮ず娴忚鍣?smoke
  - 浠锋牸缁存姢涓庢ā鍨嬫帰娴嬬殑鍓嶈疆 worklog 鏄庣‘杩欎袱鍧楀崱鐗囧簲褰㈡垚鎴愬闂幆锛屼笉搴斿啀鍚勮嚜娓哥
  - 鎵嬪伐楠屾敹娓呭崟瑕佹眰瑕嗙洊妗岄潰鍜?`375px`
- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細宸插皾璇曠湡瀹炴祻瑙堝櫒 smoke锛屼絾鏈舰鎴愰€氳繃璇佹嵁
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細
  - Node 鍐?`spawn` 鍦ㄦ湰鏈鸿Е鍙?`EPERM`锛屼笉閫傚悎缁х画渚濊禆 JS 鑷惎瀛愯繘绋嬮摼
  - PowerShell 鍙惎鍔ㄥ墠鍚庣鏈嶅姟锛屼絾鍐呰仈 `playwright-cli eval` 鐨勮秴闀?JS 瀛楃涓蹭粛瀹规槗琚綋鍓嶅涓荤瓥鐣?杞箟灞傛墦鏂?+- 寰呴獙璇侀〉闈㈡竻鍗曪細璁剧疆椤?`LLMPrice / DynamicRouting / CircuitBreaker / ModelProbe` 鐨勬闈㈠竷灞€銆乣375px` 甯冨眬銆佸府鍔╂寜閽?hover/focus
- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細
  - 杩樻病鏈夋妸 PowerShell 鐗堢湡瀹?smoke 涓叉垚浠撳簱鍐呯嫭绔嬭剼鏈叆鍙?+  - 鐪熷疄娴忚鍣?smoke 浠嶉渶鎷嗘垚鏇寸煭鐨?PowerShell 姝ラ鎴栧崟鐙?`.ps1` 鏂囦欢鍐嶈窇
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒筹紱鐜版湁 `mjs` 宸叉仮澶嶅彲鎵ц锛屽彲鍦ㄤ笅涓€杞户缁ˉ PowerShell 鍏ュ彛鎴栨媶姝ュ疄璺?+
## 12. 涓嬩竴杞缓璁?+
- 涓嬩竴杞渶閫傚悎缁х画鎺ㄨ繘锛氭妸鏈疆鎵嬪姩 PowerShell smoke 灏濊瘯鏀跺彛鎴愪粨搴撳唴鐙珛 `.ps1` 鑴氭湰锛岄伩鍏嶇户缁敤瓒呴暱鍐呰仈鍛戒护
- 鍚屼富绾垮€欓€夐『搴忥細
  1. 鏂板 `scripts/verify-setting-help-browser-smoke.ps1`锛屾妸鐪熷疄 smoke 鎷嗘垚鐭楠?+  2. 鐢ㄨ `.ps1` 鑴氭湰璺戦€氳缃〉鍥涘潡鍗＄墖鐨勬闈笌 `375px` 娴忚鍣?smoke
  3. 鎶婄湡瀹?smoke 缁撴灉鍐欏洖 `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md` 鍜?automation memory
