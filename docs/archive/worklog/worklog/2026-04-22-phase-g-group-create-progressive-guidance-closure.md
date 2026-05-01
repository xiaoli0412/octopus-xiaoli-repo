# 2026-04-22 Phase G Group Create Progressive Guidance Closure

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氬垱寤哄垎缁勫脊绐楁笎杩涘紩瀵间笌楂樼骇绛栫暐鏀跺彛
- 鏃ユ湡锛?026-04-22
- 褰撳墠闃舵锛歅hase G 鍥剧墖闂浼樺厛杩斿伐绐楀彛
- 瀵瑰簲 milestone锛氭埅鍥鹃棶棰樻睜 / 鍒涘缓鍒嗙粍寮圭獥涓昏矾寰勫紩瀵兼敹鍙?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?9.2銆?.3銆?4銆?6 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?1.0銆?.2銆?.3銆?.4 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-22-phase-g-circuit-breaker-progressive-help-closure.md`
- 鏈浠诲姟鐩爣锛?+  - 鎶婂垱寤哄垎缁勫脊绐楄ˉ鎴愨€滀富璺緞寮曞 + 鍩虹绛栫暐鎽樿 + 楂樼骇绛栫暐鎶樺彔 + 绌烘€佷笅涓€姝ヨ鏄庘€濈殑鍙В閲婄粨鏋?+  - 淇鏂板寮曞鏂囨渚濊禆缂哄け locale key 鐨勯闄╋紝閬垮厤涓枃鐣岄潰缁х画娉勬紡鍘熷 key
  - 澧炲姞涓€鏉?no-browser 楠岃瘉鑴氭湰锛屽浐瀹氭湰杞粨鏋勪笌鏂█
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - automation memory `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/group/Editor.tsx`
  - `web/src/components/modules/group/Create.tsx`
  - `web/src/components/modules/group/Card.tsx`
  - `web/src/components/modules/channel/Form.tsx`
  - `scripts/verify-channel-create-flow.mjs`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佸墠绔富绾跨姸鎬併€乤utomation memory銆佸垎缁勫脊绐楃幇鏈夊疄鐜般€佹笭閬撳垱寤洪〉鐨勬笎杩涙枃妗堟ā寮?+- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛岋紝涓旀湰杞槸鍚岄〉灏忛棴鐜?+
## 3. 鏈纭鍒?+
- 鍙鐞嗗垱寤哄垎缁勫脊绐楀悓椤靛紩瀵笺€佺瓥鐣ヨВ閲婂拰绌烘€佽鏄庯紝涓嶆墿鏁ｅ埌鏃犲叧椤甸潰
- 榛樿鐣岄潰蹇呴』淇濇寔绠€娲侊紝楂樼骇绛栫暐缁х画鎶樺彔
- 鏈疆蹇呴』鐣欎笅鍙璺戠殑闈欐€侀獙璇佺粨鏋?+
## 4. 鏈绂佹浜嬮」

- 涓嶆敼鍚庣鍒嗙粍鎺ュ彛鎴栨彁浜ょ粨鏋?+- 涓嶆妸鍒嗙粍楂樼骇绛栫暐閲嶆柊鎽婂洖棣栧睆
- 涓嶅湪鏈仛娴忚鍣?smoke 鐨勬儏鍐典笅瀹ｇО椤甸潰宸插叏閲忛獙鏀?+
## 5. 鏈楠屾敹鏉′欢

- 鍒嗙粍鍒涘缓寮圭獥鍏峰鈥滃垱寤洪『搴忊€濅富璺緞寮曞涓庡熀纭€绛栫暐鎽樿
- 楂樼骇绛栫暐浠嶄綅浜庢姌鍙犲尯锛屼笖鍔犲叆棣栧瓧瓒呮椂銆佷細璇濅繚鎸併€侀噸璇曘€佹晠闅滅獥鍙ｃ€佺珵閫熷弬鏁扮殑甯姪鎻愮ず
- 妯″瀷閫夋嫨鍖哄拰宸查€夋ā鍨嬪尯鍏峰鏄庣‘绌烘€?涓嬩竴姝ヨ鏄?+- `node scripts/verify-group-create-flow.mjs` 閫氳繃
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 閫氳繃

## 6. 鏈鍥炴粴鐐?+
- `web/src/components/modules/group/Editor.tsx`
- `scripts/verify-group-create-flow.mjs`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛纭鍒涘缓鍒嗙粍褰撳墠缁撴瀯鍜屾埅鍥鹃棶棰樺彛寰勶紝鍐嶆敼 UI 寮曞涓庨潤鎬侀獙璇?+- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細鍒涘缓鍒嗙粍寮圭獥
- 鍙楀奖鍝嶆帴鍙ｏ細鏃犳柊澧炴帴鍙?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細浠呭寮哄垱寤哄垎缁勫脊绐楃殑淇℃伅鏋舵瀯涓庤鏄庯紝涓嶆敼鍙樻彁浜よ涔?+
## 8. 瀹炴柦姝ラ

1. 澶嶆煡 `GroupEditor` 鐜扮姸锛岀‘璁ら珮绾х瓥鐣ュ凡鎶樺彔浣嗕富璺緞寮曞涓庣┖鎬佽鏄庝粛缂哄け銆?+2. 鍦?`GroupEditor` 涓姞鍏ラ《閮ㄤ富璺緞寮曞銆佸熀纭€绛栫暐鎽樿銆佹ā寮忚鏄庛€佹ā鍨?宸查€夋ā鍨嬬┖鎬佹彁绀猴紝骞舵妸鏂板璇存槑鏂囨鏀惰繘缁勪欢绾?copy锛岃閬跨己澶?locale key 椋庨櫓銆?+3. 鏂板 `scripts/verify-group-create-flow.mjs`锛屽浐瀹氳繖杞粨鏋勫拰鏂囨寮曠敤鏂█銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛歚D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 娴嬭瘯鍛戒护锛歚node scripts/verify-group-create-flow.mjs`
- 涓撻」楠岃瘉锛氳鍥?`GroupEditor.tsx` 鍜屾柊鑴氭湰锛岀‘璁ゆ棤鏂板 `group.form.*` 婕?key 渚濊禆

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏈疆鍙寮哄垎缁勫垱寤哄脊绐楃殑瑙ｉ噴灞傚拰闈欐€佹牎楠?+- 鍏煎鎬ч闄╋細浣庯紱鏈敼鎺ュ彛鍜屾寔涔呭寲鏁版嵁
- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氶€氳繃
- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細涓昏鍒掋€佺敤鎴蜂笂涓嬫枃鎬昏处銆佸墠绔富绾跨姸鎬併€乤utomation memory銆佸垎缁勫垱寤洪〉涓庢笭閬撳垱寤洪〉鐜版湁瀹炵幇
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - 鐢ㄦ埛涓婁笅鏂囨€昏处鏄庣‘瑕佹眰鍒涘缓鍒嗙粍榛樿绠€娲併€侀珮绾ф姌鍙犮€佺┖鎬佹湁涓嬩竴姝ュ紩瀵?+  - 娓犻亾鍒涘缓椤电幇鏈?`useMemo` 娓愯繘鏂囨妯″紡鍙鐢ㄥ埌鍚岀被寮圭獥锛岄€傚悎瑙勯伩褰撳墠 locale key 缂哄彛
  - 鍓嶇涓荤嚎鐘舵€佽〃鏄庡垎缁勯〉缁撴瀯宸插熀鏈畬鎴愶紝浣嗘埅鍥鹃棶棰樻睜閲岀殑鍒涘缓鍒嗙粍鍙В閲婃€т粛寰呮敹鍙?+- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒绾?smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鍏堜互 no-browser 楠岃瘉鍜?typecheck 鏀跺彛锛屾祻瑙堝櫒绾ф闈?375px 鍥炲綊浠嶉渶鍚庣画缁х画瀹夋帓
- 寰呴獙璇侀〉闈㈡竻鍗曪細鍒涘缓鍒嗙粍寮圭獥妗岄潰甯冨眬銆?75px 甯冨眬銆侀珮绾х瓥鐣ュ睍寮€鍚庣殑鎺掔増銆佹ā鍨嬬┖鎬佷笌宸查€夋ā鍨嬬┖鎬佸彲璇绘€?+- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細
  - 鍒涘缓鍒嗙粍寮圭獥娴忚鍣ㄧ骇 smoke 浠嶆湭瀹屾垚
  - 褰撳墠 `zh-Hant` 缁х画娌跨敤涓?`zh-Hans` 涓€鑷寸殑缁勪欢绾?copy锛岃嫢鍚庣画瑕佽ˉ瀹屽叏绻佷腑鎺緸锛屽彲鍦ㄥ悓涓荤嚎涓嬬户缁粏鍖?+  - 鍏跺畠 Phase G 椤甸潰浠嶆湁娴忚鍣ㄧ骇 smoke 寰呰ˉ
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒?+
## 12. 涓嬩竴杞缓璁?+
- 浼樺厛椤?1锛氭墽琛屽垱寤哄垎缁勫脊绐楁祻瑙堝櫒绾?smoke锛岃鐩栨闈㈢涓?`375px`銆侀珮绾х瓥鐣ュ睍寮€鍜岀┖鎬佽鏄?+- 浼樺厛椤?2锛氳嫢娴忚鍣ㄧ幆澧冧粛涓嶅彲鐢紝缁х画鍚屼富绾夸笅涓€涓缃〉/鎴浘闂闂幆锛屼絾浼樺厛閫夋嫨宸叉湁闈欐€侀獙璇佸叆鍙ｇ殑椤甸潰
- 浼樺厛椤?3锛氬湪鍚屼富绾夸笅琛?`CC Switch` 鎴栬缃〉鐔旀柇妯″潡鐨勬祻瑙堝櫒绾?smoke 璇佹嵁
