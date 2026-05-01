# 2026-04-22 Phase G Model Probe Progressive Guidance Closure

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氳缃〉妯″瀷鎺㈡祴妯″潡娓愯繘寮忚鏄庝笌楂樼骇缁嗚皟鏀跺彛
- 鏃ユ湡锛?026-04-22
- 褰撳墠闃舵锛歅hase G 鍥剧墖闂浼樺厛杩斿伐绐楀彛
- 瀵瑰簲 milestone锛氭埅鍥鹃棶棰樻睜 / 璁剧疆椤垫帰娴嬩笌妫€娴嬪綊浣嶆敹鍙?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?9.6銆?4銆?6 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?1.0銆?.2銆?.3銆?.4 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-22-phase-g-dynamic-routing-progressive-help-closure.md`
- 鏈浠诲姟鐩爣锛?+  - 鎶婅缃〉妯″瀷鎺㈡祴浠庘€滄寜妯″瀷骞抽摵涓夊瓧娈碘€濇敹鍙ｆ垚鈥滈粯璁よ鏄?+ 鎼滅储 + 鎽樿鍗＄墖 + 绛栫暐鎸囧紩 + 楂樼骇缁嗚皟灞曞紑鈥?+  - 缁х画钀藉疄鈥滄帰娴嬩笌妫€娴嬪綊浣嶅埌璁剧疆椤点€佷环鏍奸〉涓嶆壙杞芥帰娴嬪叆鍙ｂ€濈殑涓荤嚎瑕佹眰
  - 琛ヤ竴鏉?no-browser 楠岃瘉鑴氭湰锛屽浐瀹氭湰杞粨鏋勪笌 locale 鏂█
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - automation memory `C:\Users\鏉庢槉妗怽.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/setting/ModelProbe.tsx`
  - `web/src/components/modules/setting/DynamicRouting.tsx`
  - `web/src/components/modules/setting/CircuitBreaker.tsx`
  - `web/public/locale/en.json`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€乤utomation memory銆佽缃〉鐜版湁瀹炵幇
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鍙仛鐒﹁缃〉妯″瀷鎺㈡祴鍚岄〉灏忛棴鐜紝涓嶅睍寮€澶囦唤銆佹笭閬撻〉鎴栧悗绔瓥鐣ュ疄鐜?+- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭湭浣跨敤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氬惁
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛氭湭浣跨敤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞槸鍗曟ā鍧椼€佸皬鑼冨洿銆佸己鑰﹀悎 UI 闂幆

## 3. 鏈纭鍒?+
- 鍙鐞嗚缃〉妯″瀷鎺㈡祴妯″潡锛屼笉鎵╂暎鍒颁环鏍奸〉鎴栧悗绔ā鍨嬫帴鍙ｈ涔?+- 蹇呴』淇濇寔鈥滄帰娴嬩笌妫€娴嬪綊浣嶅埌璁剧疆椤碘€濈殑鍙ｅ緞锛屼笉鎶婁环鏍间笌鎺㈡祴鍐嶆娣锋斁
- 蹇呴』鐣欎笅鍙璺戠殑鏃犳祻瑙堝櫒楠岃瘉璇佹嵁

## 4. 鏈绂佹浜嬮」

- 涓嶄慨鏀瑰悗绔帰娴嬬瓥鐣ヨ涔夈€佹寔涔呭寲缁撴瀯鍜屾帴鍙ｅ瓧娈?+- 涓嶆妸鏈墽琛岀殑娴忚鍣ㄧ骇 smoke 鍐欐垚宸插畬鎴?+- 涓嶅洖鍒版棤鍏抽〉闈㈠仛娉涘寲 UI 娓呯悊

## 5. 鏈楠屾敹鏉′欢

- 璁剧疆椤垫ā鍨嬫帰娴嬮粯璁ゅ厛灞曠ず璇存槑銆佹悳绱€佹憳瑕佸崱鐗囦笌绛栫暐鎸囧紩
- 鍗曟ā鍨嬬粏璋冧繚鐣欏湪灞曞紑闈㈡澘涓紝涓嶅啀榛樿閾烘弧鏁撮〉
- 涓夎 locale 琛ラ綈鏂板璇存槑銆佹憳瑕佷笌鎸囧紩鏂囨
- `node scripts/verify-model-probe-help.mjs` 閫氳繃
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 閫氳繃

## 6. 鏈鍥炴粴鐐?+
- `web/src/components/modules/setting/ModelProbe.tsx`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/zh-Hant.json`
- `scripts/verify-model-probe-help.mjs`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鏀硅缃〉缁撴瀯涓庢枃妗堬紝鍐嶈ˉ楠岃瘉鑴氭湰
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細璁剧疆椤垫ā鍨嬫帰娴嬫ā鍧椾笌 locale
- 鍙楀奖鍝嶆帴鍙ｏ細鏃犳柊澧炴帴鍙ｏ紱缁х画澶嶇敤妯″瀷鍒楄〃涓庢ā鍨嬫洿鏂版帴鍙?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細浠呭寮鸿缃〉灞曠ず缁撴瀯鍜屽府鍔╄鏄庯紝涓嶆敼鍙樻ā鍨嬫帰娴嬪瓧娈典繚瀛橀€昏緫

## 8. 瀹炴柦姝ラ

1. 澶嶆牳 `ModelProbe.tsx` 褰撳墠缁撴瀯銆佺幇鏈?locale 涓庤缃〉鐩搁偦妯″潡鐨勬笎杩涘紡缁撴瀯鍙傝€冦€?+2. 閲嶆瀯 `SettingModelProbe` 涓庡崟妯″瀷灞曞紑鍖猴紝鍔犲叆榛樿璺緞璇存槑銆佹憳瑕佸崱鐗囥€佺瓥鐣ユ寚寮曞拰楂樼骇缁嗚皟灞傘€?+3. 琛ラ綈涓夎 locale锛屾柊澧?`scripts/verify-model-probe-help.mjs`锛屽苟鎵ц no-browser 楠岃瘉涓庡墠绔?typecheck銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛歚D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 娴嬭瘯鍛戒护锛歚node scripts/verify-model-probe-help.mjs`
- 涓撻」楠岃瘉锛氳鍙?`ModelProbe.tsx` 涓庝笁璇?`modelProbe` locale 鐗囨锛岀‘璁ゆ柊澧炵粨鏋勩€佽鏄庢枃妗堝拰鎽樿鏂█钀藉湴

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏈疆鍙皟鏁磋缃〉鎺㈡祴妯″潡鐨勫睍绀虹粨鏋勪笌璇存槑灞?+- 鍏煎鎬ч闄╋細浣庯紱鏈敼鎺ュ彛涓庢寔涔呭寲鏁版嵁
- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氶€氳繃
- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佽缁嗗伐浣滄祦銆佸墠绔富绾跨姸鎬併€乤utomation memory銆佽缃〉鐜版湁瀹炵幇
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 鐢ㄦ埛涓婁笅鏂囨€昏处鏄庣‘瑕佹眰鎺㈡祴涓庢娴嬩粠浠锋牸鍖哄煙瀹屽叏鍓ョ锛屽苟鍦ㄨ缃〉浠ユ洿浜烘€у寲缁撴瀯鍛堢幇
  - 鍔ㄦ€佽矾鐢变笌鐔旀柇妯″潡宸插舰鎴愨€滈粯璁よ鏄?+ 鎽樿浼樺厛 + 楂樼骇灞曞紑鈥濈殑鍙鐢ㄦ牱寮?+  - automation memory 鎸囧悜鍚屼竴鏉¤缃〉鍥剧墖闂涓荤嚎锛屽簲缁х画鍦ㄨ缃〉鐩搁偦妯″潡涓婂舰鎴愰棴鐜?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒绾?smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鍏堜互 no-browser 楠岃瘉鍜?typecheck 鏀跺彛锛涙祻瑙堝櫒绾ч〉闈㈠洖褰掍粛闇€鍦ㄥ悓涓荤嚎涓嬬户缁畨鎺?+- 寰呴獙璇侀〉闈㈡竻鍗曪細璁剧疆椤垫ā鍨嬫帰娴嬪崱鐗囨闈㈠竷灞€銆乣375px` 甯冨眬銆佸府鍔╂彁绀?hover/focus銆佸崟妯″瀷灞曞紑涓庝繚瀛樹氦浜?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞负鍗曢〉灏忛棴鐜?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細
  - 璁剧疆椤垫ā鍨嬫帰娴嬪崱鐗囩殑娴忚鍣ㄧ骇 smoke 浠嶆湭鎵ц
  - 褰撳墠杩樻病鏈夐〉闈㈢骇璇佹嵁璇佹槑甯姪鎻愮ず鍦?hover/focus 涓嬬殑瀹為檯鍙鎬?+  - 鍓嶇涓荤嚎鐘舵€佹枃妗ｆ湰杞湭鍗曠嫭鏀瑰啓锛屼粛闇€鍦ㄥ悗缁悓涓荤嚎鏀跺彛鏃剁粺涓€琛ヨ繘娴忚鍣ㄧ骇璇佹嵁鐘舵€?+- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒?+
## 12. 涓嬩竴杞缓璁?+
- 涓嬩竴杞渶閫傚悎缁х画鎺ㄨ繘锛氳缃〉妯″瀷鎺㈡祴鍗＄墖鐨勬闈?/ `375px` 娴忚鍣ㄧ骇 smoke 涓庡府鍔╂彁绀?hover/focus 鍥炲綊
- 鍚屼富绾垮€欓€夐『搴忥細
  1. 璁剧疆椤垫ā鍨嬫帰娴嬪崱鐗囨祻瑙堝櫒绾?smoke
  2. 璁剧疆椤靛姩鎬佽矾鐢变笌鐔旀柇鍗＄墖娴忚鍣ㄧ骇 smoke 娓呭熬
  3. 鑻ラ〉闈㈢骇楠岃瘉浠嶅彈闃伙紝鍒欑户缁ˉ璁剧疆椤电浉閭绘ā鍧楃殑鏇撮珮灞?no-browser 璇佹嵁
