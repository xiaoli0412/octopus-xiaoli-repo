# 2026-04-23 娣卞害瀹℃煡锛氬浠藉揩鐓х鍙烽摼鎺ヨ竟鐣屽姞鍥?+
## 鍩烘湰淇℃伅

- 鏃堕棿锛?026-04-23 08:57:03 +08:00
- 涓荤嚎锛氬畨鍏ㄦ紡娲炰笌鏉冮檺杈圭晫娣卞
- Master plan aligned before coding: yes

## 鏈疆澶嶇敤鐨勬湰鍦拌祫婧?+
- `AGENTS.md`锛氱‘璁よ嚜鍔ㄥ寲鎵ц杈圭晫銆佽緭鍑鸿姹傘€侀伩鍏嶈鐩栫敤鎴风幇鏈夋敼鍔ㄣ€?+- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`锛氱‘璁ゅ綋鍓嶄富绾夸粛浠ラ珮椋庨櫓璺緞瀹℃煡涓庡皬骞呬慨澶嶄负鍏堛€?+- `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`锛氱‘璁ゆ瘡杞渶鍏堝榻?canonical plan锛屽啀鍋?worklog / memory 鍥炲啓銆?+- `C:\Users\鏉庢槉妗怽.codex\automations\octopus\memory.md`锛氱户鎵夸笂涓€杞?residual risks锛屼紭鍏堝鏌?providers trust boundary 涓庡浠?鍥炴粴鍏ュ彛銆?+- `internal/server/handlers/providers.go`銆乣internal/server/handlers/setting.go`銆乣internal/op/backup.go`銆乣internal/op/backup_test.go`锛氭湰杞富瑕佸鏌ヤ笌淇瀵硅薄銆?+
## 鏈疆瑕嗙洊鑼冨洿

- 鍏ㄤ粨绱㈠紩閲嶅缓锛歚cmd`銆乣internal`銆乣web`銆乣scripts`銆乣.github`銆乣docs`
- 娣卞妯″潡锛?+  - `internal/server/handlers/providers.go`
  - `internal/helper/fetch.go`
  - `internal/client/http.go`
  - `internal/server/handlers/setting.go`
  - `internal/op/backup.go`
  - `internal/op/log.go`
  - `scripts/dockerfiles/Dockerfile.debian`
  - `scripts/dockerfiles/entrypoint.sh`

## Findings

### 1. 澶囦唤蹇収鍥炴粴璺緞瀛樺湪绗﹀彿閾炬帴瓒婄晫璇诲彇绐楀彛

- 鏂囦欢锛歚internal/op/backup.go`
- 浣嶇疆锛歚resolveImportSnapshotPath`
- 闂鎻忚堪锛氬師瀹炵幇鍙 `filepath.Abs(...)` 缁撴灉鍋氣€滄槸鍚︿綅浜庡揩鐓х洰褰曞墠缂€涓嬧€濈殑瀛楃涓插垽鏂紝娌℃湁瑙ｆ瀽绗﹀彿閾炬帴鏈€缁堢洰鏍囥€?+- 褰卞搷锛氬鏋?`import-snapshots` 鐩綍鍐呭嚭鐜版寚鍚戠洰褰曞 JSON 鏂囦欢鐨?symlink锛宍DBRollbackImportSnapshot` / `DBPreviewRollbackImportSnapshot` 浼氭妸瀹冨綋鎴愬悎娉曞揩鐓ц鍙栵紝绐佺牬鈥滀粎鍏佽璇诲彇蹇収鐩綍鍐呮枃浠垛€濈殑杈圭晫銆?+- 璇佹嵁锛氳皟鐢ㄩ摼鏄?`DBRollbackImportSnapshot` -> `loadImportSnapshotByName` -> `loadImportSnapshotByPath` -> `resolveImportSnapshotPath` -> `os.ReadFile(resolvedPath)`锛涙棫瀹炵幇瀵?`resolvedPath` 浣跨敤 `filepath.Abs` 鍚庣洿鎺ュ仛鍓嶇紑姣旇緝锛屾湭璋冪敤 `filepath.EvalSymlinks`銆?+- 澶勭悊鐘舵€侊細宸蹭慨澶嶃€?+
## 鏈疆宸插疄鏂界殑灏忓箙淇

1. 鍦?`internal/op/backup.go` 鐨?`resolveImportSnapshotPath` 涓柊澧炵湡瀹炶矾寰勮В鏋愶細
   - 瀵瑰揩鐓х洰褰曞拰鐩爣蹇収璺緞鍒嗗埆鎵ц `filepath.EvalSymlinks`銆?+   - 鑻ョ洰鏍囪矾寰勮В鏋愬埌鐩綍澶栵紝鍒欐嫆缁濊鍙栥€?+   - 鑻ョ洰鏍囨枃浠朵笉瀛樺湪锛屼粎淇濈暀鍘熸湁 `os.IsNotExist` 瀹归敊锛屼笉鏀瑰彉姝ｅ父涓嶅瓨鍦ㄨ矾寰勭殑閿欒璇箟銆?+
2. 鍦?`internal/op/backup_test.go` 涓柊澧炲洖褰掓祴璇?`TestDBRollbackImportSnapshotRejectsSymlinkOutsideSnapshotDir`锛?+   - 鍦ㄥ揩鐓х洰褰曞唴鍒涘缓鎸囧悜涓存椂鐩綍澶栭儴鏂囦欢鐨?symlink銆?+   - 鏂█鍥炴粴鎺ュ彛鎷掔粷璇ヨ矾寰勶紝骞惰繑鍥?`snapshot path is outside import snapshot directory`銆?+   - Windows 鏃犵鍙烽摼鎺ユ潈闄愭椂浣跨敤 `t.Skipf(...)`锛岄伩鍏嶆妸鐜闄愬埗璇涓哄姛鑳藉け璐ャ€?+
## 楠岃瘉

- 宸叉墽琛岋細`gofmt -w internal/op/backup.go internal/op/backup_test.go`
- 宸叉墽琛岋細`go test ./internal/op -count=1`
  - 璇存槑锛氶€氳繃浠撳簱鏈湴 `.tools/` 娉ㄥ叆 `GOCACHE/GOTMPDIR/TMP/TEMP`锛岃閬块粯璁?Windows 缂撳瓨鐩綍鏉冮檺闂銆?+- 缁撴灉锛氶€氳繃銆?+
## 鏈墽琛岄獙璇佷笌鍘熷洜

- `go test ./...`锛氭湰杞慨鏀逛粎瑙﹀強 `internal/op` 鐨勫揩鐓ц矾寰勬牎楠屼笌瀹氬悜娴嬭瘯锛屾湭鎵╁ぇ鍒板叏浠撳洖褰掞紝浠ラ伩鍏嶅湪澶ч噺鐢ㄦ埛杩涜涓敼鍔ㄤ笂寮曞叆鏃犲叧鍣煶銆?+- Linux / Docker 杩愯鏃堕獙璇侊細褰撳墠涓绘満浠嶆槸 Windows 鐜锛屾湰杞湭鏂板瀹瑰櫒杩愯鏃朵唬鐮侊紝鏈噸澶嶆墽琛屻€?+
## 褰撳墠鍓╀綑椋庨櫓

1. `internal/server/handlers/providers.go` 浠嶄緷璧栧彲鍙樼殑 GitHub branch raw URL 浣滀负 providers 杩滅▼鏉ユ簮锛岃櫧鐒跺凡闃绘柇 redirect锛屼絾渚涘簲閾捐緭鍏ヤ粛闈炰笉鍙彉璧勬簮銆?+2. `scripts/dockerfiles/Dockerfile.debian` 鏈€缁堥暅鍍忎粛浠?root 鍚姩锛屽啀鐢?entrypoint 闄嶆潈锛涜櫧鐒惰繍琛屾湡鍙檷鏉冿紝浣嗛粯璁ら暅鍍忓眰绾т粛涓嶆槸鏈€灏忔潈闄愯捣鐐广€?+3. 娴忚鍣ㄧ骇 UI / 鎵嬪伐鎴浘璇佹嵁浠嶆湭鍦ㄦ湰杞ˉ鍏咃紝鏈疆鑱氱劍鍚庣淇′换杈圭晫銆?+
## 涓嬩竴杞缓璁紭鍏堥」

1. 缁х画瀹?providers 杩滅▼婧愶紝璇勪及鏄惁鍒囨崲鍒板浐瀹?commit SHA 鎴?release asset銆?+2. 澶嶅 `setting/import` 涓婁紶璺緞鐨勫ぇ鍖呬綋 / multipart 琛屼负锛岀‘璁や笉瀛樺湪棰濆鐨勮祫婧愭斁澶х獥鍙ｃ€?+3. 鑻ュ悗绔珮椋庨櫓灏忎慨鏆傛椂鏀跺彛锛屽啀鍥炲埌 Phase G 鐨勬祻瑙堝櫒绾ц瘉鎹棴鐜€?+
## 鏈疆缁撴灉

- 缁撴灉锛氭垚鍔?+- 鏄惁闇€瑕佷汉宸ヤ粙鍏ワ細鏆備笉闇€瑕?+
