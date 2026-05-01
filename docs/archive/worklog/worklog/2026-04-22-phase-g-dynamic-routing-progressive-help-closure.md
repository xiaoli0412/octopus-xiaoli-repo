# 2026-04-22 Phase G Dynamic Routing Progressive Help Closure

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氳缃〉鍔ㄦ€佽矾鐢辨笎杩涘府鍔╀笌楂樼骇棰勭畻鎶樺彔鏀跺彛
- 鏃ユ湡锛?026-04-22
- 褰撳墠闃舵锛歅hase G 鍥剧墖闂浼樺厛杩斿伐绐楀彛
- 瀵瑰簲 milestone锛氭埅鍥鹃棶棰樻睜 / 璁剧疆椤垫帰娴嬩笌鍔ㄦ€佸仴搴峰綊浣嶆敹鍙?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?9 鑺傘€佺 14 鑺傘€佺 16 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?1.0 鑺傘€佺 1.2 鑺傘€佺 1.3 鑺傘€佺 1.4 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-22-phase-g-circuit-breaker-progressive-help-closure.md`
- 鏈浠诲姟鐩爣锛?+  - 鎶婅缃〉鍔ㄦ€佽矾鐢变粠鈥滃紑鍏?+ 鍏ㄩ噺棰勭畻瀛楁鐩撮摵鈥濇敹鍙ｆ垚鈥滈粯璁ょ畝娲?+ 鎽樿浼樺厛 + 楂樼骇棰勭畻鎶樺彔鈥濄€?+  - 缁欒繍琛屾椂璋冧紭寮€鍏冲拰鍚勫眰绾ч绠楄ˉ榻愬満鏅寲甯姪鎻愮ず锛岀户缁惤瀹炩€滄帰娴嬩笌妫€娴嬪綊浣嶅埌璁剧疆椤碘€濈殑涓荤嚎銆?+  - 鐣欎笅鍙噸澶嶆墽琛岀殑鏃犳祻瑙堝櫒楠岃瘉鑴氭湰锛屽苟璁板綍褰撳墠 Vitest 鐜闃诲銆?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - automation memory `C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/DynamicRouting.test.tsx`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `scripts/vitest-no-spawn.cjs`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€佷笂涓€杞?automation memory銆佹湰鍦?`frontend-design` / `brainstorming` / `using-superpowers`
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鍙仛鐒﹁缃〉鍔ㄦ€佽矾鐢卞悓椤靛皬闂幆锛屼笉灞曞紑澶囦唤瀵煎叆鎴栧悗绔姩鎬佽矾鐢卞疄鐜?+- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭棤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氬惁
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛氭棤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞槸鍗曟ā鍧楀墠绔井闂幆

## 3. 鏈纭鍒?+
- 鍙鐞嗚缃〉鍔ㄦ€佽矾鐢卞崱鐗囩殑缁撴瀯銆佸府鍔╂彁绀恒€乴ocale 鍜岄獙璇侊紝涓嶆墿鏁ｅ埌鍚庣杩愯鏃惰涔夈€?+- 蹇呴』瀵归綈闇€姹?03銆?4銆?6銆?2銆?9銆?0銆?7銆?2锛屼笉鍐嶆妸鎺㈡祴/鍔ㄦ€佸仴搴峰叆鍙ｆ贩鍥炰环鏍奸〉銆?+- 蹇呴』鐣欎笅鍙噸澶嶆墽琛岀殑鏃犳祻瑙堝櫒楠岃瘉锛屼笉鎶婃湭璺戦€氱殑娴忚鍣ㄦ垨 Vitest 缁撴灉鍐欐垚閫氳繃銆?+
## 4. 鏈绂佹浜嬮」

- 涓嶄慨鏀瑰姩鎬佽矾鐢卞悗绔瓥鐣ャ€佺粺璁℃帴鍙ｆ垨鎸佷箙鍖栫粨鏋勩€?+- 涓嶆妸澶囦唤銆乣CC Switch`銆佹笭閬撻〉绛夊叾浠栨埅鍥句富棰樻贩鍏ユ湰杞敼鍔ㄣ€?+- 涓嶆帺鐩栧綋鍓?`vitest` 鐨?`spawn EPERM` 鐜闂銆?+
## 5. 鏈楠屾敹鏉′欢

- 璁剧疆椤靛姩鎬佽矾鐢遍粯璁ゅ彧灞曠ず璇存槑銆佸紑鍏炽€佹憳瑕佸拰棰勭畻姒傝锛岄珮绾ч绠楅」杩涘叆鎶樺彔鍖恒€?+- 涓夊 locale 閮借ˉ榻愭柊鐨勫府鍔╂枃妗堬紝绻佷腑涓嶅啀淇濈暀 `Failover / Key / Probe` 杩欑被涓绘樉绀鸿瘝銆?+- `node scripts/verify-dynamic-routing-help.mjs` 閫氳繃锛屽墠绔?`tsc --noEmit` 閫氳繃銆?+
## 6. 鏈鍥炴粴鐐?+
- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/components/modules/setting/DynamicRouting.test.tsx`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `scripts/verify-dynamic-routing-help.mjs`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鏀?UI 缁撴瀯鍜?locale锛屽啀琛ラ獙璇佽剼鏈笌娴嬭瘯鏂█
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細璁剧疆椤靛姩鎬佽矾鐢卞崱鐗?+- 鍙楀奖鍝嶆帴鍙ｏ細鏃?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細浠呰皟鏁磋缃〉鐨勫睍绀虹粨鏋勫拰璇存槑鏂瑰紡锛屼笉鏀瑰彉鍚庣瀹為檯淇濆瓨閿€?+
## 8. 瀹炴柦姝ラ

1. 澶嶆牳鍔ㄦ€佽矾鐢卞崱鐗囧綋鍓嶅疄鐜板拰鐢ㄦ埛涓婁笅鏂囦腑鈥滄帰娴?妫€娴嬪綊浣嶃€侀粯璁ょ畝娲併€佸府鍔╂彁绀衡€濈殑纭姹傦紝纭褰撳墠缂哄彛鏄瓧娈电洿閾轰笌鏂囨涓嶈冻銆?+2. 閲嶆瀯 `SettingDynamicRouting`锛氬鍔犻粯璁よ矾寰勮鏄庛€佽繍琛屾椂璋冧紭寮€鍏冲府鍔┿€佹憳瑕佷紭鍏堝竷灞€銆侀绠楁瑙堝崱鐗囧拰鈥滈珮绾х珵閫熼绠椻€濇姌鍙犲尯銆?+3. 鏇存柊涓嫳绻佷笁濂?locale锛屽苟琛?`scripts/verify-dynamic-routing-help.mjs`锛涘悓姝ユ敹绱?`DynamicRouting.test.tsx` 鏂█鍒版柊缁撴瀯銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛歚D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 娴嬭瘯鍛戒护锛歚node scripts/verify-dynamic-routing-help.mjs`
- 涓撻」楠岃瘉锛?+  - 灏濊瘯鎵ц `D:\gol1\node.exe .\web\node_modules\vitest\vitest.mjs run --config .\web\vitest.config.ts .\web\src\components\modules\setting\DynamicRouting.test.tsx --pool threads`
  - 缁撴灉浠嶅湪鍔犺浇 `vitest.config.ts` 闃舵瑙﹀彂 `spawn EPERM`锛屼笌浠撳簱鏃㈡湁 Vite/Vitest 鐜闃诲涓€鑷达紝鏈疆鏈兘鍦ㄨ鐜涓嬫嬁鍒?Vitest 閫氳繃缁撴灉

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏈疆鍙皟鏁磋缃〉灞曠ず缁撴瀯鍜屽府鍔╂彁绀?+- 鍏煎鎬ч闄╋細浣庯紱鏈慨鏀规帴鍙ｅ拰鎸佷箙鍖栧瓧娈?+- 鏄惁闃诲涓嬩竴浠诲姟锛氫笉闃诲锛屼絾娴忚鍣ㄧ骇 smoke 鍜?Vitest 鐜鎭㈠鍚庤繕搴旇ˉ鏇撮珮灞傞獙璇佽瘉鎹?+
## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氶€氳繃锛坄tsc --noEmit`锛?+- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃锛坄verify-dynamic-routing-help.mjs`锛夛紱Vitest 鍥犵幆澧?`spawn EPERM` 鏈€氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€乤utomation memory銆佽缃〉鐜版湁瀹炵幇銆乣frontend-design`銆乣brainstorming`銆乣using-superpowers`
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 鐢ㄦ埛涓婁笅鏂囨€昏处鏄庣‘杩欒疆蹇呴』缁х画钀藉疄鈥滄帰娴?妫€娴嬪綊浣嶅埌璁剧疆椤碘€濃€滈粯璁ょ畝娲?+ 甯姪鎻愮ず + 娓愯繘灞曞紑鈥濄€?+  - 鍓嶇涓荤嚎鐘舵€佸拰涓婅疆 memory 鎸囧悜璁剧疆椤典粛鏈夊悓涓荤嚎鍙敹鍙ｉ」锛屼紭鍏堜簬鍥炲埌鏃犲叧椤甸潰銆?+  - `CircuitBreaker.tsx` 鎻愪緵浜嗗彲浠ュ鐢ㄧ殑鈥滄憳瑕佷紭鍏?+ 楂樼骇鎶樺彔鈥濈粨鏋勫弬鑰冦€?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒绾ф墜宸?smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鍏堜互鍚岄〉闈欐€侀獙璇佸拰 typecheck 鏀跺彛锛涙祻瑙堝櫒绾?smoke 浠嶅緟缁熶竴瀹夋帓
- 寰呴獙璇侀〉闈㈡竻鍗曪細璁剧疆椤靛姩鎬佽矾鐢卞崱鐗囨闈㈠竷灞€銆?75px 甯冨眬銆侀珮绾ч绠楀睍寮€涓庡府鍔╂彁绀?hover/focus
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞负鍗曟ā鍧楀井闂幆
- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細
  - `DynamicRouting.test.tsx` 宸插榻愭柊缁撴瀯锛屼絾鏈満 Vite/Vitest 浠嶅洜 `spawn EPERM` 鏃犳硶瀹屾垚鎵ц
  - 璁剧疆椤靛姩鎬佽矾鐢辨祻瑙堝櫒绾?smoke 灏氭湭琛ラ綈
  - 鍓嶇涓荤嚎鐘舵€佹枃妗ｅ悗缁彲琛ヤ竴鏉♀€滆缃〉鍔ㄦ€佽矾鐢卞凡杩涘叆榛樿绠€娲?+ 楂樼骇棰勭畻鎶樺彔鈥濈殑鐘舵€佽鏄?+- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒?+
## 12. 涓嬩竴杞缓璁?+
- 涓嬩竴杞渶閫傚悎缁х画鎺ㄨ繘锛氳缃〉鍔ㄦ€佽矾鐢辨祻瑙堝櫒绾?smoke锛屾垨缁х画鍚屼富绾跨殑璁剧疆椤垫帰娴?甯姪鎻愮ず鏀跺彛
- 鍚屼富绾垮€欓€夐『搴忥細
  1. 璁剧疆椤靛姩鎬佽矾鐢辨闈笌 375px 娴忚鍣ㄧ骇 smoke
  2. 缁х画鏀跺彛璁剧疆椤垫ā鍨嬫帰娴嬬瓥鐣ョ殑娴忚鍣ㄧ骇璇佹嵁锛屽舰鎴愪笌鍔ㄦ€佽矾鐢?鐔旀柇涓€鑷寸殑璁剧疆椤甸棴鐜?+  3. 鍥炲埌 `CC Switch` 鎴栫啍鏂崱鐗囩殑娴忚鍣ㄧ骇 smoke 娓呭熬
