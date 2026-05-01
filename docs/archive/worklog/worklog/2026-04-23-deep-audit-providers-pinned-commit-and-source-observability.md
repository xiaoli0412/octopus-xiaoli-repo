# 2026-04-23 Deep Audit Providers Pinned Commit And Source Observability

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛歱roviders 杩滅▼婧愬浐瀹?commit SHA 涓庢潵婧愬彲瑙傛祴鎬ф敹鍙?+- 鏃ユ湡锛?026-04-23
- 褰撳墠闃舵锛歅hase A / security and release-readiness deep audit
- 瀵瑰簲 milestone锛歨igh-risk trust-boundary closure

## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 涓畨鍏ㄨ竟鐣屻€佸彲鍥炲綊楠岃瘉涓庡彂甯冨彲鎺ф€х浉鍏宠姹?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 涓€滃厛瀵归綈璁″垝锛屽啀鍋氬皬骞呬慨澶嶏紝鍐嶈ˉ楠岃瘉涓庤惤鐩樷€?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-23-deep-audit-release-validation-gate.md`
- 鏈浠诲姟鐩爣锛氬叧闂?providers 杩滅▼鍒嗘敮婕傜Щ椋庨櫓锛屽苟琛ユ渶灏忓彲瑙傛祴鎬э紝渚夸簬鍚庣画鎺掗殰纭褰撳墠鍝嶅簲鏉ヨ嚜 pinned remote 杩樻槸 embedded fallback
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細`AGENTS.md`銆乤utomation memory銆乣docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`銆乣docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`銆佹棦鏈?providers 瀹℃煡璁板綍銆乣internal/server/handlers/providers.go`銆乣internal/server/handlers/providers_test.go`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細浠撳簱鍐?worklog 涓?automation memory 鎻愪緵浜?providers branch raw URL 椋庨櫓鐨勫欢缁笂涓嬫枃锛涚幇鏈?providers 娴嬭瘯鎻愪緵浜嗗皬淇彲钀界偣锛涘綋鍓嶇嚎绋嬩笂涓嬫枃鏄庣‘浜嗘湰杞喅绛栭噰鐢?`commit SHA`
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈户缁睍寮€鍓嶇/browser 璧勬簮锛屽洜涓烘湰杞洰鏍囪仛鐒﹀悗绔?trust-boundary 灏忎慨
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭湭浣跨敤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氬惁
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛氫笉閫傜敤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細浠诲姟寰堝皬锛屼笖鐩存帴渚濊禆褰撳墠 providers 瀹℃煡涓婁笅鏂囷紝涓荤嚎绋嬪畬鎴愭洿绋冲Ε

## 3. 鏈纭鍒?+
- 涓嶆墿澶т负 release asset 璁捐锛屽彧鍋氫綆椋庨櫓 `commit SHA` 鏀跺彛
- 涓嶇牬鍧忕幇鏈?`/api/v1/providers` JSON 鍝嶅簲缁撴瀯
- 鎵€鏈夋敼鍔ㄥ繀椤绘湁瀹氬悜鍥炲綊娴嬭瘯

## 4. 鏈绂佹浜嬮」

- 涓嶉噸鏋?providers 鏁翠綋缂撳瓨/璺敱閫昏緫
- 涓嶈Е纰版棤鍏?dirty workspace 鏂囦欢
- 涓嶆妸鏈疆鎵╁睍鎴愭柊鐨勮缃」鎴栧墠绔ぇ鏀?+
## 5. 鏈楠屾敹鏉′欢

- `providersGitHubURL` 涓嶅啀鎸囧悜 `refs/heads/...`
- 瀛樺湪鍥炲綊娴嬭瘯闃叉 future drift
- `/api/v1/providers` 鍙湪涓嶆敼 body 濂戠害鐨勫墠鎻愪笅鏆撮湶鏉ユ簮涓?pinned commit 淇℃伅
- `go test ./internal/server/handlers -count=1` 閫氳繃

## 6. 鏈鍥炴粴鐐?+
- 鍥炴粴 `internal/server/handlers/providers.go`
- 鍥炴粴 `internal/server/handlers/providers_test.go`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鏀瑰悗绔?trust-boundary 璇箟
- 鍙楀奖鍝嶅悗绔ā鍧楋細`internal/server/handlers/providers.go`銆乣internal/server/handlers/providers_test.go`
- 鍙楀奖鍝嶅墠绔ā鍧楋細鏃犵洿鎺ヤ唬鐮佹敼鍔?+- 鍙楀奖鍝嶆帴鍙ｏ細`GET /api/v1/providers`锛堜粎鏂板鍝嶅簲澶达紝涓嶆敼鍝嶅簲浣擄級
- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細杞诲井锛屾柊澧炲彧璇诲搷搴斿ご锛涘師 body 琛屼负涓嶅彉

## 8. 瀹炴柦姝ラ

1. 灏?providers 杩滅▼婧愪粠 mutable branch raw URL 鍥哄畾鍒板綋鍓嶇‘璁ょ殑 commit SHA銆?+2. 娣诲姞鍥炲綊娴嬭瘯锛岀‘淇?URL 涓嶈兘鍐嶅洖閫€鍒?`refs/heads/...`銆?+3. 鍦?`/api/v1/providers` 鍝嶅簲澶磋ˉ `X-Octopus-Providers-Source` 涓?`X-Octopus-Providers-Commit`锛岀敤浜庡垽鏂懡涓?remote 杩樻槸 embedded锛屽苟璁板綍 pinned SHA銆?+4. 瀵归綈鐜版湁娴嬭瘯璋冪敤鐐瑰苟鎵ц handlers 瀹氬悜鍥炲綊銆?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛氭湭鍗曠嫭鎵ц `go build ./...`锛屾湰杞粎瑙﹀強 handlers 灏忚寖鍥?+- 娴嬭瘯鍛戒护锛歚go test ./internal/server/handlers -count=1`
- 涓撻」楠岃瘉锛歚gofmt -w internal/server/handlers/providers.go internal/server/handlers/providers_test.go`

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細鏃犳柊澧為珮椋庨櫓锛涗粎鏂板鍝嶅簲澶达紝鐞嗚涓婃瀬浣庢鐜囧奖鍝嶄緷璧栦弗鏍?header 鐧藉悕鍗曠殑澶栭儴瀹㈡埛绔?+- 鍏煎鎬ч闄╋細`/api/v1/providers` body 鏈彉锛屽吋瀹规€ч闄╀綆
- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氭湭鎵ц鍏ㄩ噺鏋勫缓锛沨andlers 瀹氬悜娴嬭瘯閫氳繃
- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃锛宍go test ./internal/server/handlers -count=1` 閫氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆佹棦鏈?providers worklog銆佸綋鍓嶇嚎绋嬪喅绛栥€佺幇鏈?handlers 娴嬭瘯
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細memory/worklog 纭涓婅疆 residual risk 浠嶆槸 mutable branch source锛涘綋鍓嶇嚎绋嬬‘璁ら€夋嫨 `commit SHA` 鑰屼笉鏄?`release asset`锛涚幇鏈夋祴璇曞憡璇夋垜鍙互鍦?handlers 灞備綆椋庨櫓琛ュ洖褰?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭湭浣跨敤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氫笉閫傜敤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛屾祻瑙堝櫒鎴?Docker smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏈疆鏀瑰姩浠呬负 handlers 灏忎慨锛屾棤闇€鎵╁ぇ鍒版祻瑙堝櫒/Docker锛涘綋鍓嶄富鏈轰篃涓嶉€傚悎琛?Linux/Docker 璺緞
- 寰呴獙璇侀〉闈㈡竻鍗曪細鏃犳柊澧為〉闈㈠緟楠岃瘉
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細浠诲姟鑼冨洿灏忋€佸己渚濊禆褰撳墠涓婁笅鏂囷紝涓荤嚎绋嬫洿楂樻晥涓旈伩鍏嶅苟鍙戝紑閿€
- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細providers 浠嶄緷璧?GitHub 鍙揪鎬э紝鍙槸宸蹭粠 mutable branch 鏀剁揣鍒?immutable commit锛涜嫢鍚庣画闇€瑕佹洿姝ｅ紡娌荤悊锛屽彲鍐嶈瘎浼?release asset
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭槸锛宲roviders mutable-source 椋庨櫓宸蹭粠 branch drift 鏀剁揣锛屼笅涓€杞彲杞悜 README/release 鍗犱綅绗︽垨 multipart import resource audit
