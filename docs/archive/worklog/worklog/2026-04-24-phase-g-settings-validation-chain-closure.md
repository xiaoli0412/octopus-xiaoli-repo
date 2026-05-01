# 2026-04-24 Phase G Settings Validation Chain Closure

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛歅hase G 璁剧疆椤?no-browser 楠屾敹閾捐ˉ娲?+- 鏃ユ湡锛歚2026-04-24`
- 褰撳墠闃舵锛歚Phase G` 鎴浘浼樺厛 UI 涓荤嚎
- 瀵瑰簲 milestone锛氳缃〉 `summary-first + help-hint` 楠岃瘉閾炬敹鍙?+
## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): `yes`
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?`9`銆乣14` 鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?`1.0`銆乣1.2`銆乣1.3` 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-24-phase-g-backup-summary-first-closure.md`
- 鏈浠诲姟鐩爣锛氭妸宸茬粡钀藉湴涓旀湰鍦板凡閫氳繃鐨勮缃〉 no-browser 瀹堟姢琛ヨ繘 `validation / release` 宸ヤ綔娴侊紝鏀跺彛鈥滄湰鍦拌剼鏈凡缁裤€丆I 浠嶅彲鑳芥紡妫€鈥濈殑鍚屾睜楠岃瘉缂哄彛銆?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-24-phase-g-circuit-breaker-summary-first-closure.md`
  - `docs/worklog/2026-04-24-phase-g-model-probe-summary-first-batching-closure.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
  - `.github/workflows/validation.yaml`
  - `.github/workflows/release.yaml`
  - `scripts/verify-circuit-breaker-help.mjs`
  - `scripts/verify-model-probe-help.mjs`
  - `scripts/verify-setting-info-logic.mjs`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細涓婅堪鏂囨。銆佷笁鏉¤缃〉 no-browser 鑴氭湰銆乣validation/release` workflow銆乤utomation memory
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鐩爣鏄獙璇侀摼琛ユ礊锛屼笉娑夊強鍐嶆淇敼璁剧疆椤电粍浠剁粨鏋勬垨娴忚鍣?smoke
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細`鍚
- 鏈瀛?agent 浣跨敤妯″瀷锛歚鏈娇鐢╜
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛歚鍚︼紝涓荤嚎绋嬩覆琛屾墽琛宍
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛歚涓嶉€傜敤`
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓嶅垱寤哄瓙 agent锛屼笖鏈疆浠诲姟鍙秹鍙?workflow 涓庤褰曞眰锛屼富绾跨▼涓茶鏇寸ǔ

## 3. 鏈纭鍒?+
- 鍙洿缁?`Phase G / settings / validation-chain` 鎺ㄨ繘锛屼笉鎵╂暎鍒板叾瀹冨姛鑳界嚎銆?+- 涓嶅洖閫€宸叉湁璁剧疆椤?`summary-first` 缁勪欢缁撴瀯锛屽彧淇獙璇侀摼婕忔帴涓庢枃妗ｇ姸鎬併€?+- 鏈疆蹇呴』褰㈡垚鈥渨orkflow 鍙樻洿 + 鐩存帴楠岃瘉 + worklog/memory 鍚屾鈥濈殑闂幆銆?+
## 4. 鏈绂佹浜嬮」

- 涓嶄慨鏀?`CircuitBreaker`銆乣ModelProbe`銆佽缃〉鐗堟湰鍗＄墖杩愯鏃剁粨鏋勩€?+- 涓嶆妸瀹夸富 `spawn EPERM` 鎴?Edge/CDP 闂璇垽鎴愬綋鍓?no-browser 楠岃瘉鍥炲綊銆?+- 涓嶉『鎵嬫墿鏁ｅ埌鏃犲叧鑴氭湰娓呯悊鎴栧叾瀹?CI 閲嶆瀯銆?+
## 5. 鏈楠屾敹鏉′欢

- `node scripts/verify-circuit-breaker-help.mjs` 閫氳繃銆?+- `node scripts/verify-model-probe-help.mjs` 閫氳繃銆?+- `node scripts/verify-setting-info-logic.mjs` 閫氳繃銆?+- `.github/workflows/validation.yaml` 涓?`.github/workflows/release.yaml` 閮芥帴鍏?`ModelProbe` 涓?`SettingInfo` 鐨?no-browser 瀹堟姢銆?+
## 6. 鏈鍥炴粴鐐?+
- `.github/workflows/validation.yaml`
- `.github/workflows/release.yaml`
- `docs/worklog/2026-04-24-phase-g-settings-validation-chain-closure.md`
- `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氫笉鏀?UI锛屼笉鏀逛笟鍔¤涔夛紝鍙ˉ workflow 涓庤褰?+- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細鏃犺繍琛屾椂浠ｇ爜鍙樻洿
- 鍙楀奖鍝嶆帴鍙ｏ細鏃?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紝浠呮彁楂?CI/release 瀵硅缃〉 no-browser 瀹堟姢鐨勮鐩栧害

## 8. 瀹炴柦姝ラ

1. 澶嶆煡 `CircuitBreaker / ModelProbe / SettingInfo` 涓夋潯璁剧疆椤?no-browser 鑴氭湰涓庡綋鍓嶇粍浠跺绾︼紝纭鏈湴鑴氭湰宸茬豢銆佺己鍙ｅ彧鍦?workflow 婕忔帴銆?+2. 鏇存柊 `.github/workflows/validation.yaml` 涓?`.github/workflows/release.yaml`锛屾妸 `verify-model-probe-help.mjs` 涓?`verify-setting-info-logic.mjs` 鎺ュ叆 `Frontend no-browser verification` 姝ラ銆?+3. 鍚屾鏇存柊 worklog 涓?automation memory锛岀暀涓嬩笅涓€杞叆鍙ｃ€?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛氭棤鏂板鏋勫缓锛涙湰杞互 no-browser 鑴氭湰鍜?workflow 妫€鏌ヤ负涓?+- 娴嬭瘯鍛戒护锛?+  - `node scripts/verify-circuit-breaker-help.mjs`
  - `node scripts/verify-model-probe-help.mjs`
  - `node scripts/verify-setting-info-logic.mjs`
- 涓撻」楠岃瘉锛氱‘璁?`validation / release` 閮藉寘鍚?`ModelProbe` 涓?`SettingInfo` 鐨勮缃〉 no-browser 瀹堟姢

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鍙彁楂橀獙璇佽鐩栵紝涓嶆敼涓氬姟琛屼负
- 鍏煎鎬ч闄╋細浣庯紱鑻ュ悗缁缃〉 copy/缁撴瀯缁х画鍙樺寲锛孋I 浼氭洿鏃╂毚闇茶剼鏈绾︽紓绉?+- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛歚鏄紙鏈疆鏃犳柊澧炴瀯寤烘楠わ級`
- 娴嬭瘯鏄惁閫氳繃锛歚鏄痐
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細canonical plan銆佺敤鎴锋€昏处銆佽缁嗗伐浣滄祦銆佺幆澧冭鍒掋€佹渶杩?settings/backup worklog銆乤utomation memory銆佷笁鏉¤缃〉 no-browser 鑴氭湰銆乿alidation/release workflow
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - canonical / workflow / 鐢ㄦ埛鎬昏处锛氱‘璁ゆ湰杞粛搴斿仠鐣欏湪 `Phase G screenshot-first settings pool`锛屼紭鍏堜慨楠岃瘉閾炬紡鎺ヨ繖绉嶄綆椋庨櫓闂幆椤?+  - 鐩稿叧 worklog锛氱‘璁?`CircuitBreaker`銆乣ModelProbe`銆佽缃〉鐗堟湰淇℃伅閮藉凡鍚勮嚜瀹屾垚 no-browser 鏀跺彛锛屼笉搴斿啀鍥炲ご鏀圭粍浠剁粨鏋?+  - automation memory锛氱‘璁ゅ綋鍓嶅涓讳粛瀛樺湪 `spawn EPERM` / CDP blocker锛屾湰杞簲浼樺厛鍋氫笉渚濊禆瀹夸富娴忚鍣ㄧ殑鍚屾睜闂幆
  - 涓夋潯 no-browser 鑴氭湰涓?workflow锛氱‘璁ょ湡瀹炵己鍙ｆ槸 `ModelProbe / SettingInfo` 杩樻湭绾冲叆 `validation / release`
- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛歚鏈娇鐢╜
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛歚涓嶉€傜敤`
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒绾?smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鍙ˉ CI/workflow锛涙祻瑙堝櫒绾ц瘉鎹粛鍙楀涓?`spawn EPERM` 涓?Edge/CDP bootstrap 闂褰卞搷
- 寰呴獙璇侀〉闈㈡竻鍗曪細瀹夸富鎭㈠鍚庯紝缁х画琛ヨ缃〉鐗堟湰淇℃伅銆乣ModelProbe`銆乣CircuitBreaker` 鐨勬祻瑙堝櫒绾?`375px / hover / focus` 璇佹嵁
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘绂佹锛屼笖鏈疆浠诲姟寰堝皬锛屼富绾跨▼涓茶鏇寸ǔ
- worklog 鏄惁鏇存柊锛歚鏄痐
- 閬楃暀椤癸細`DynamicRouting`銆乣Backup` 鍜?settings/browser smoke 浠嶉渶缁х画鍚屾睜鏀跺彛锛沗verify-llm-price-boundary.mjs` 绛夊叾瀹?no-browser 瀹堟姢鏄惁涔熷簲杩涘叆涓婚獙璇侀摼锛屽悗缁彲鍗曠嫭璇勪及
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛歚鏄痐
