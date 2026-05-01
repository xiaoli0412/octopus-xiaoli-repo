# 2026-04-29 娣卞害瀹℃煡锛欳ORS loopback 杈圭晫鍥炲綊鏀跺彛

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛歞eep audit and low-risk fix for CORS loopback trust boundary regression
- 鏃ユ湡锛?026-04-29
- 褰撳墠闃舵锛歅hase A 瀹夊叏杈圭晫涓庡伐绋嬫敹鍙?+- 瀵瑰簲 milestone锛氬畨鍏ㄨ竟鐣?/ 鍙戝竷闂ㄧ鍥炲綊淇

## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 瀹夊叏杈圭晫銆侀厤缃粯璁ゅ畨鍏ㄣ€侀獙鏀堕棬绂佺浉鍏崇珷鑺?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?1銆?銆?0銆?1 鑺?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-27-deep-audit-jwt-session-binding.md`
- 鏈浠诲姟鐩爣锛氬鏍镐笂涓€杞畬鏁村璁￠噷鏍囧嚭鐨?CORS 杈圭晫鍥炲綊锛屽苟鍦ㄤ綆椋庨櫓鑼冨洿鍐呮仮澶嶁€滀粎 debug 鎴栨樉寮?allowlist 鎵嶅厑璁?loopback origin鈥濈殑绾︽潫锛屽悓鏃惰ˉ瓒虫渶灏忓洖褰掓祴璇?+- 鏈宸茬洏鐐规湰鍦拌祫婧愶細automation memory銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁?workflow銆佷笂涓€杞畬鏁村璁℃姤鍛娿€乣internal/server/middleware/cors.go`銆乣internal/server/middleware/cors_test.go`銆乣internal/conf/debug.go`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細`$CODEX_HOME/automations/octopus/memory.md`銆乣docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`銆乣docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`銆乣docs/review/瀹℃煡/2026-04-28-234549-octopus-repo-complete-audit.md`
- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆鐩爣鏄崟鐐瑰畨鍏ㄥ洖褰掍慨澶嶏紝涓嶉渶瑕佸睍寮€鍒板墠绔?UI 鎴栨洿骞跨殑 relay 涓荤嚎
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭棤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氬惁锛屼粎涓荤嚎绋?+- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛氭棤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰鏈嚜鍔ㄥ寲涓嶈鍒涘缓瀛?agent锛屼笖鏈疆闂灞€閮ㄣ€佷笂涓嬫枃寮鸿€﹀悎銆佷富绾跨▼鍙洿鎺ュ畬鎴?+
## 3. 鏈纭鍒?+
- 浼樺厛淇瀹夊叏杈圭晫鍜屽彂甯冮棬绂侀樆濉烇紝涓嶆墿澶у埌鏃犲叧閲嶆瀯
- 鍙仛灞€閮ㄣ€佸彲楠岃瘉鐨勫皬淇紝骞惰ˉ鏈€灏忓繀瑕佹祴璇?+- 涓嶈鐩栧綋鍓嶅伐浣滃尯鍏朵粬鏈彁浜ゆ敼鍔?+
## 4. 鏈绂佹浜嬮」

- 涓嶆妸 loopback 鏀惧閲嶆柊鍖呰鎴愨€滃紑鍙戜究鍒╂€р€濊€屽拷鐣ラ粯璁ゅ畨鍏ㄨ竟鐣?+- 涓嶄慨鏀规棤鍏?handler / relay / UI 鏂囦欢
- 涓嶅仛鏃犳硶鍦ㄦ湰杞獙璇佺殑璺ㄦā鍧楅噸鏋?+
## 5. 鏈楠屾敹鏉′欢

- `internal/server/middleware` 鐨?CORS 娴嬭瘯閲嶆柊閫氳繃
- 闈?debug 妯″紡涓?loopback origin 涓嶅啀琚棤鏉′欢鏀捐
- 鏄惧紡 allowlist 鐨?loopback 鍦烘櫙浠嶆湁鍥炲綊瑕嗙洊

## 6. 鏈鍥炴粴鐐?+
- `internal/server/middleware/cors.go`
- `internal/server/middleware/cors_test.go`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氬厛鏀瑰悗绔畨鍏ㄨ涔?+- 鍙楀奖鍝嶅悗绔ā鍧楋細`internal/server/middleware`
- 鍙楀奖鍝嶅墠绔ā鍧楋細鏃?+- 鍙楀奖鍝嶆帴鍙ｏ細鎵€鏈夌粡杩囧叏灞€ `Cors()` 涓棿浠剁殑 HTTP 璺敱
- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鏄紝鏀剁揣浜嗛潪 debug 妯″紡涓嬪 loopback origin 鐨勯粯璁ゆ斁琛?+
## 8. 瀹炴柦姝ラ

1. 澶嶆牳 `cors.go`銆乣cors_test.go` 涓?`conf.IsDebug()`锛岀‘璁ゅ綋鍓嶅洖褰掓槸 loopback 鏃犳潯浠舵斁琛岃€屼笉鏄祴璇曡繃鏃躲€?+2. 鎶?`Cors()` 涓殑 loopback 鏀捐鏉′欢鏀跺洖鍒?`conf.IsDebug() && isLocalDevOrigin(origin)`銆?+3. 鏂板闈?debug + 鏄惧紡 allowlist 鐨?loopback 鍥炲綊娴嬭瘯锛岄伩鍏嶈浼ゅ彈鎺у紑鍙戝満鏅€?+
## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛氭湭鍗曠嫭鎵ц鏋勫缓锛涙湰杞负涓棿浠跺眬閮ㄤ慨澶?+- 娴嬭瘯鍛戒护锛?+  - `. .\scripts\use-go-env.ps1; gofmt -w internal\server\middleware\cors.go internal\server\middleware\cors_test.go`
  - 鍦ㄤ粨搴撳唴鏈湴缂撳瓨鐩綍涓嬫墽琛岋細`go test ./internal/server/middleware -run 'TestCors(DeniesLoopbackOutsideDebug|AllowsLoopbackInDebug|AllowsWildcardAllowlist|AllowsConfiguredOriginHost|AllowsConfiguredLoopbackOutsideDebug)$' -count=1`
  - 鍦ㄥ悓鏍风殑鏈湴缂撳瓨/涓存椂鐩綍閰嶇疆涓嬫墽琛岋細`go test ./internal/server/middleware -count=1`
  - `git diff --check -- internal/server/middleware/cors.go internal/server/middleware/cors_test.go`
- 涓撻」楠岃瘉锛氫负缁曡繃褰撳墠 Windows 瀹夸富榛樿 `D:\DevCache\temp` / `D:\DevCache\go-build` 鏉冮檺闂锛屾湰杞妸 `GOCACHE/GOMODCACHE/GOTMPDIR/TMP/TEMP` 閮藉垏鍒颁粨搴撳唴鍙啓鐩綍鍚庡啀璺?Go 娴嬭瘯锛岀‘璁ゅけ璐ュ櫔闊虫潵鑷涓婚粯璁ょ紦瀛樼洰褰曡€屼笉鏄ˉ涓佹湰韬?+
## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱浠呮仮澶嶉粯璁ゆ嫆缁濋潪 debug loopback 璺ㄥ煙
- 鍏煎鎬ч闄╋細浣庡埌涓紱渚濊禆鈥滅敓浜фā寮忎笅 localhost 鑷姩璺ㄥ煙鍙敤鈥濈殑鏈湴璋冭瘯鏂瑰紡浼氳鏀剁揣锛屼絾鏄惧紡 `cors_allow_origins` 浠嶅彲鎭㈠璇ヨ兘鍔?+- 鏄惁闃诲涓嬩竴浠诲姟锛氫笉闃诲锛涘綋鍓嶅凡鎶婅繖涓彂甯冮棬绂佸洖褰掓敹鍙?+
## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氭湭鎵ц鍏ㄩ噺鏋勫缓锛涙湰杞粎鎵ц浜嗗眬閮ㄦ牸寮忓寲涓庝腑闂翠欢娴嬭瘯
- 娴嬭瘯鏄惁閫氳繃锛氶€氳繃锛宍internal/server/middleware` 瀹氬悜涓庡叏鍖呮祴璇曞潎閫氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆佸綋鍓嶇姸鎬佹枃妗ｃ€佽缁?workflow銆佷笂涓€杞畬鏁村璁°€乣cors.go/cors_test.go/debug.go`
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細
  - automation memory锛氱‘璁や笂涓€杞凡鎶?CORS 鍥炲綊鍒椾负褰撳墠宸ヤ綔鍖虹殑鏈€楂樹紭鍏堢骇鍙戝竷闃诲涔嬩竴
  - 褰撳墠鐘舵€?/ workflow锛氱‘璁ゆ湰杞簲浼樺厛鍋氬畨鍏ㄨ竟鐣?+ 灏忎慨 + 鏈€灏忛獙璇?+  - 瀹屾暣瀹¤鎶ュ憡锛氭彁渚涗簡鍏蜂綋鏂囦欢銆佸奖鍝嶅拰鈥滄祴璇曢棬绂佸凡澶辫触鈥濈殑璇佹嵁閾?+  - 浠ｇ爜浜嬪疄锛氱‘璁ゅ绾﹀苟鏈洿鏂帮紝鐪熷疄鍋忓樊鍦?`cors.go`
- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛岋紱鏈疆鏄腑闂翠欢瀹夊叏璇箟淇锛屾棤闇€娴忚鍣?smoke
- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細鏃?+- 寰呴獙璇侀〉闈㈡竻鍗曪細鏃狅紱鏈疆涓烘湇鍔＄涓棿浠朵慨澶?+- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛鏄庣‘瑕佹眰涓嶈鍒涘缓瀛?agent锛屼笖璇ラ棶棰樺眬閮ㄣ€佸彲鐢变富绾跨▼鐩存帴鏀跺彛
- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細`internal/update` 鐨?checksum/signature 鏍￠獙涓?Windows staged replacement 椋庨櫓浠嶆湭澶勭悊锛沗.gitignore` 杩囩獎瀵艰嚧鐨勫伐浣滃尯鍣煶浠嶅湪
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭弧瓒筹紱涓嬩竴杞彲鍥炲埌 `internal/update` 鎴栧叾浠栨湭閴存潈鍏紑鍏ュ彛鐨勯珮椋庨櫓鍖哄煙缁х画娣卞
