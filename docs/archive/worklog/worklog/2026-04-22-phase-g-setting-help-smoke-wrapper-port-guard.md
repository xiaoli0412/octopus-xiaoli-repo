# 2026-04-22 Phase G Setting Help Smoke Wrapper Port Guard

## Scope

- Master plan aligned before coding (yes/no): yes
- Mainline: Phase G 璁剧疆椤靛府鍔╂彁绀烘祻瑙堝櫒 smoke 鏀跺彛
- Current stage: 娴忚鍣?smoke 鍏ュ彛绋冲畾鍖栦笌鐪熷疄鎵ц闃诲鏀舵暃
- Goal: 淇缁熶竴 smoke 鑴氭湰閲屼細褰卞搷鐪熷疄鎵ц鐨勫府鍔╂寜閽€夋嫨鍣ㄦ紓绉伙紝骞剁粰 PowerShell wrapper 琛ヤ笂 self-start 绔彛鍐茬獊淇濇姢涓庢棭閫€鎶ラ敊锛岄伩鍏嶅啀娆″仠鐣欏湪鏃犱俊鎭秴鏃?+
## This Round Plan

1. 澶嶆牳璁剧疆椤靛洓鍧楀崱鐗囩殑甯姪鎻愮ず smoke 鍏ュ彛涓?`HelpHint` 榛樿 `aria-label` 鏄惁涓€鑷淬€?+2. 淇缁熶竴 smoke 鑴氭湰涓殑甯姪鎸夐挳閫夋嫨鍣ㄦ紓绉伙紝骞惰ˉ涓€鏉¤剼鏈骇鏂█閿佷綇璇ョ害鏉熴€?+3. 澶嶈窇 `check-only` 涓?wrapper 鑷惎閾捐矾锛岀‘璁ゆ柊鐨勯樆濉炵偣琚敹鏁涘埌鍙В閲婅寖鍥淬€?+
## Changes Made

1. 淇浜嗙粺涓€ Playwright smoke 鑴氭湰鐨勫府鍔╂寜閽€夋嫨鍣ㄧǔ瀹氭€с€?+   - 鏂囦欢锛歚scripts/verify-setting-help-browser-smoke.mjs`
   - 鍋氭硶锛氭柊澧?`stableHelpButtonLabel` / `stableHelpButtonSelector`锛屾妸鐪熸鍙備笌鍗＄墖妫€娴嬩笌 focus 妫€鏌ョ殑閫夋嫨鍣ㄥ浐瀹氫负 `鏌ョ湅甯姪`锛屼笉鍐嶅彈鏂囦欢鍐呭巻鍙蹭贡鐮侀粯璁ゅ€煎奖鍝嶃€?+2. 鍔犲己浜嗗府鍔╂彁绀哄彲璁块棶鎬ч獙璇佽剼鏈殑璺ㄨ剼鏈竴鑷存€ф柇瑷€銆?+   - 鏂囦欢锛歚scripts/verify-help-hint-accessible.mjs`
   - 鍋氭硶锛氶櫎浜嗘鏌?`HelpHint.tsx` 鏈韩锛屼篃鏄惧紡妫€鏌?`verify-setting-help-browser-smoke.mjs` 涓?`verify-setting-help-browser-smoke-cdp.mjs` 鐨勯粯璁ゅ府鍔╂爣绛句笌褰撳墠缁勪欢淇濇寔涓€鑷淬€?+3. 涓?PowerShell wrapper 琛ヤ笂 self-start 绔彛鍐茬獊淇濇姢涓庢洿鏃╃殑澶辫触鏆撮湶銆?+   - 鏂囦欢锛歚scripts/verify-setting-help-browser-smoke.ps1`
   - 鍋氭硶锛氭柊澧?`Test-TcpPortAvailable`銆乣Resolve-FreeTcpPort`锛岃 wrapper 鍦ㄨ嚜鍚ā寮忎笅閬囧埌榛樿绔彛鍗犵敤鏃惰嚜鍔ㄥ皾璇曠浉閭荤┖闂茬鍙ｃ€?+   - 鍋氭硶锛氬悗绔惎鍔ㄥ悗鍏堢瓑寰呯害 1.2 绉掑苟妫€鏌ユ槸鍚︽彁鍓嶉€€鍑猴紱鑻ュ凡閫€鍑猴紝鍒欑洿鎺ヨ鍙栨棩蹇楀苟鎶涘嚭閿欒锛岃€屼笉鏄户缁崱鍦?`Wait-Http` 閲屽舰鎴愰粦绠辫秴鏃躲€?+4. 娓呯悊浜嗘湰杞皟璇曟椂浜х敓鐨勪复鏃?`apply_patch` 鎺㈤拡鏂囦欢锛屾病鏈夋妸璋冭瘯鍨冨溇鐣欏湪宸ヤ綔鍖恒€?+
## Verification

- `D:\gol1\node.exe --check scripts\verify-setting-help-browser-smoke.mjs`
- `D:\gol1\node.exe scripts\verify-setting-help-browser-smoke.mjs --check-only`
- `D:\gol1\node.exe scripts\verify-help-hint-accessible.mjs`
- `D:\gol1\node.exe scripts\verify-setting-help-browser-smoke-cdp.mjs --check-only`
- `powershell -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp`
- `powershell -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp`

## Results

- 缁熶竴 smoke 鑴氭湰涓嶅啀渚濊禆閿欒鐨勫府鍔╂寜閽粯璁ゅ€硷紝鍜?`HelpHint` 鐨勫綋鍓嶅疄鐜伴噸鏂板榻愩€?+- wrapper 鐨?`self-start` 閾捐矾涓嶅啀琚?`127.0.0.1:18081` 绔彛鍗犵敤鐩存帴鎵撴柇涓烘棤淇℃伅瓒呮椂锛涙湰杞凡纭鍘熷厛鐨勭鍙ｅ啿绐佽璺ㄨ繃銆?+- 鏈€鏂颁竴娆?`self-start + cdp` 涓細
  - 鍚庣浜岃繘鍒跺凡鎴愬姛鍚姩锛屽苟鍐欏嚭 `Program started, press Ctrl+C to exit`
  - Edge 宸叉垚鍔熸毚闇?`DevTools listening on ws://127.0.0.1:9222/...`
  - 鏂扮殑闃诲鐐瑰凡缁忎粠鈥滃叆鍙ｈ剼鏈?绔彛鍐茬獊鈥濇敹鏁涗负鈥淐DP smoke 鍦ㄥ綋鍓嶇幆澧冮噷娌℃湁鍦ㄨ秴鏃舵椂闂村唴瀹屾垚杩斿洖鈥?+
## Current Blockers

1. 鏈€鏂颁竴娆?wrapper 鑷惎閾捐矾缁撴潫鍚庯紝`http://127.0.0.1:18081/healthz` 宸蹭笉鍙揪锛岃鏄庡悗绔繘绋嬪湪 wrapper 缁撴潫鍓嶅悗娌℃湁绋冲畾瀛樻椿鍒版渶缁堥獙鏀讹紝浠嶉渶涓嬩竴杞簿纭畾浣嶈繘绋嬬敓鍛藉懆鏈熸垨绛夊緟閾捐矾銆?+2. Edge 鏃ュ織鏄剧ず褰撳墠涓绘満瀵?profile/cache/sandbox 鐩綍瀛樺湪澶氭潯 `鎷掔粷璁块棶 (0x5)` 涓?crashpad/network sandbox 鐩稿叧閿欒锛涜櫧鐒?CDP 宸茬洃鍚垚鍔燂紝浣嗚繖浠嶆槸鏈湴娴忚鍣ㄨ繍琛岀幆澧冮闄┿€?+3. 鏈疆娌℃湁鎷垮埌鐪熸鐨勬闈㈢涓?`375px` 娴忚鍣?smoke 鎴愬姛缁撴灉锛屽洜姝よ缃〉鍥涘潡鍗＄墖鐨勭湡瀹炴祻瑙堝櫒绾ч獙鏀朵粛涓嶈兘璁颁负瀹屾垚銆?+
## Next Entry

1. 缁х画鐣欏湪 Phase G 鍚屼富绾匡紝浼樺厛绮剧‘瀹氫綅 `self-start + cdp` 鏈繑鍥炴椂鐨勫疄闄呭仠鐣欐楠ゃ€?+2. 寤鸿涓嬩竴杞厛鍦?wrapper 閲屾妸璋冪敤 Node CDP 鑴氭湰鐨?stdout/stderr 鍗曠嫭钀芥棩蹇楋紝鍖哄垎鏄崱鍦ㄧ櫥褰曘€乴ocalStorage 娉ㄥ叆銆侀〉闈㈠姞杞斤紝杩樻槸鍗″湪甯姪鎻愮ず鏂█銆?+3. 鑻ユ湰鏈?Edge 娌欑鎷掔粷璁块棶鎸佺画瀵艰嚧涓嶇ǔ瀹氾紝鍙湪鍚屼富绾夸笅鍒囧洖澶栭儴宸插紑鍚?`--remote-debugging-port=9222` 鐨勬祻瑙堝櫒浼氳瘽锛屽啀璺戯細
   - `D:\gol1\node.exe scripts\verify-setting-help-browser-smoke-cdp.mjs`

