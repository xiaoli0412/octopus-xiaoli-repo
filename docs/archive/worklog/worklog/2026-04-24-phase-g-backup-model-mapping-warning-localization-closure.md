# 2026-04-24 Phase G Backup Model Mapping Warning Localization Closure

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛氬浠芥ā鍨嬫槧灏?warning 涓枃鍖栨敹鍙?+- 鏃ユ湡锛?026-04-24
- 褰撳墠闃舵锛歅hase G 鎴浘浼樺厛 UI 鏀跺彛
- 瀵瑰簲 milestone锛歅hase G settings / backup no-browser contract closure

## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?9 鑺傚墠绔腑鏂囦竴鑷存€т笌 settings 鏀跺彛瑕佹眰
- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?1.0銆?.2銆?.3 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-24-phase-g-backup-context-prefix-localization-closure.md`
- 鏈浠诲姟鐩爣锛氱户缁部 backup 鍚屾睜鏈€灏忛棴鐜紝鏀跺彛妯″瀷鏄犲皠棰勮涓粛鍙兘鐩村嚭鐨勮嫳鏂?warning 鍙ュ瓙锛屽苟淇濇寔 no-browser 楠岃瘉閾惧彲澶嶈窇
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細涓昏鍒掋€佸綋鍓嶇姸鎬併€佹墽琛屽伐浣滄祦銆佸墠绔富绾跨姸鎬併€乤utomation memory銆乥ackup helper / test / verify 鑴氭湰銆佸悗绔?`internal/op/backup_extra.go` 涓殑鐪熷疄 warning 鏂囨
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細閬靛惊鏈疆鈥滀笉瑕佸垱寤哄瓙 agent鈥濈殑鏄庣‘瑕佹眰锛屼笖浠诲姟闆嗕腑鍦ㄥ悓涓€ helper / test / script 閾捐矾

## 3. 鏈纭鍒?+
- 鍙鐞?Phase G backup no-browser 鍚屾睜涓殑鐪熷疄鑻辨枃娉勬紡锛屼笉鎵╂暎鍒板叾浠栭〉闈㈡垨鍚庣 schema
- 鏀瑰姩浠呴檺 shared helper 杈撳嚭璇箟銆侀厤濂楁祴璇曞拰 no-browser 楠岃瘉鑴氭湰
- 鑻辨枃 locale 杈撳嚭淇濇寔涓嶅彉锛屼粎鏀跺彛闈炶嫳鏂?locale 鐨勮嫳鏂?warning 鐩村嚭

## 4. 瀹炴柦鍐呭

1. 鍦?`web/src/components/modules/setting/backup-logic.ts` 涓ˉ鍏呬袱鏉℃ā鍨嬫槧灏?warning 鐨勬渶灏忔湰鍦板寲鏄犲皠锛?+   - `mapped target not found in current environment`
   - `mapping source not referenced by selected import scopes`
2. 鍦?`web/src/components/modules/setting/backup-logic.test.ts` 鏈熬鏂板鐙珛娴嬭瘯鍧楋紝楠岃瘉 zh-Hans 涓嬭繖涓ゆ潯 warning 鐨勬樉绀烘枃妗?+3. 鍦?`scripts/verify-backup-logic.mjs` 杩藉姞鍚屾牱鐨?no-browser 鏂█锛岀‘淇濆涓讳笂鏃犻渶 Vitest worker 涔熻兘鍥炲綊

## 5. 楠岃瘉

- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- `node scripts/verify-backup-logic.mjs`

缁撴灉锛氬潎閫氳繃

## 6. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱浠呭鍔犲凡鐭?warning 鐨勬樉绀哄眰鏈湴鍖栨槧灏?+- 鍏煎鎬ч闄╋細浣庯紱鑻辨枃 locale 缁х画淇濈暀鍚庣鍘熷 warning 鍙ュ瓙锛屼腑鏂?/ 绻佷腑 / 鏃ユ枃鍙敼鍙樺睍绀哄眰
- 鏈疆鏈户缁鐞嗛」锛歚backup-logic.ts` 涓粛娈嬬暀鏃х殑涔辩爜娉ㄩ噴鍧楋紱鐢变簬瀹夸富缂栫爜鍥炴樉瀵艰嚧 `apply_patch` 閽堝璇ュ巻鍙插潡鐨勫尮閰嶄笉绋冲畾锛屾湰杞紭鍏堜繚璇佸姛鑳介棴鐜笌楠岃瘉閫氳繃锛屾竻鐞嗗伐浣滅暀缁欎笅涓€杞悓涓荤嚎鏀跺熬

## 7. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氶€氳繃
- 娴嬭瘯鏄惁閫氳繃锛歚verify-backup-logic.mjs` 閫氳繃
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛岋紱鏈疆鍙渶 no-browser 閾捐矾
- worklog 鏄惁鏇存柊锛氭槸
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒?+
## 8. 涓嬩竴杞缓璁?+
1. 缁х画鐣欏湪 Phase G backup 鍚屾睜锛屼紭鍏堟竻鐞?`backup-logic.ts` 娈嬬暀鐨勪贡鐮佹棫娉ㄩ噴鍧楋紝淇濇寔 helper 鏂囦欢鏁存磥
2. 鑻ョ户缁仛 backup 涓枃鍖栵紝鍙浆鍚?`post-import` 鐨?route / price / alias warning 鏄庣粏鏈湴鍖栵紝浣嗗墠鎻愭槸鍏堢‘璁よ繖浜涙槑缁嗗湪褰撳墠椤甸潰纭疄杩涘叆鍙璺緞
3. 缁х画浼樺厛浣跨敤 `tsc --noEmit` + `node scripts/verify-backup-logic.mjs` 浣滀负鏈満绋冲畾楠岃瘉鍏ュ彛
