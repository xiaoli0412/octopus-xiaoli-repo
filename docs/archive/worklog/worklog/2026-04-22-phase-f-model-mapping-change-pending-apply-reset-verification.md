# 2026-04-22 Phase F Model Mapping Change Pending Apply Reset Verification

## 1. 浠诲姟淇℃伅

- 浠诲姟鍚嶇О锛欱ackup 妯″瀷鏄犲皠淇敼鍚?pending apply 澶辨晥楠岃瘉琛ュ己
- 鏃ユ湡锛?026-04-22
- 褰撳墠闃舵锛歅hase F / backup-import-rollback frontend closure
- 瀵瑰簲 milestone锛歁ilestone 6 validation and deployment

## 2. 寮€宸ュ墠杈撳叆

- Master plan aligned before coding (yes/no): yes
- 瀵瑰簲 canonical 绔犺妭锛歚docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 绗?11.5 鑺傘€佺 13 鑺傞噷绋嬬 6銆佺 14 鑺傞獙鏀舵爣鍑嗐€佺 16 鑺傚疄鏂借鍒?+- 瀵瑰簲 workflow 绔犺妭锛歚docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 绗?8 鑺?Phase F銆佺 10 鑺備换鍔℃ā鏉?+- 涓婁竴涓浉鍏?worklog锛歚docs/worklog/2026-04-22-phase-f-import-mode-change-pending-apply-reset-verification.md`
- 鏈浠诲姟鐩爣锛氳ˉ榻?Backup 椤?map 妯″紡涓嬧€滀慨鏀?model mappings 浼氫娇鎸傝捣鐨?Apply This Dry-Run 缁戝畾澶辨晥锛屽苟瑕佹眰閲嶆柊 dry-run鈥濊繖鏉＄粍浠剁骇琛屼负璇佹嵁
- 鏈宸茬洏鐐规湰鍦拌祫婧愶細automation memory銆乧anonical plan銆亀orkflow銆佸墠绔富绾跨姸鎬佹枃妗ｃ€丳hase F worklog 閾俱€乣Backup.tsx`銆乣Backup.test.tsx`銆乣scripts/verify-backup-component.cjs`銆乣scripts/verify-backup-component.setting-mock.cjs`銆乣scripts/verify-backup-logic.mjs`
- 鏈浣跨敤鐨勬湰鍦?resources / skills / 璁板繂涓婁笅鏂囷細automation memory銆乧anonical plan銆亀orkflow銆佸墠绔富绾跨姸鎬佹枃妗ｃ€丳hase F pending-apply 楠岃瘉閾?+- 鑻ユ湭浣跨敤閮ㄥ垎鏈湴璧勬簮鎴栦笂涓嬫枃锛屽師鍥狅細鏈疆缁х画娌跨敤鏃㈡湁 Phase F 涓婁笅鏂囷紝娌℃湁閲嶆柊灞曞紑鏃犲叧鏂囨。鎴栨祻瑙堝櫒 smoke
- 鏈鏄惁鍚敤瀛?agent 涓庡垎宸ヨ竟鐣岋細鍚?+- 鏈瀛?agent 浣跨敤妯″瀷锛氭棤
- 鏈鏄惁鐢?Codex 鑷姩鍖栭摼璺垨鍒嗙洰褰?agent 鍗忎綔鎵ц锛氬惁
- 鑻ヤ娇鐢ㄥ垎鐩綍 agent锛岃礋璐ｇ洰褰曚笌绂佹瓒婄晫鑼冨洿锛氭棤
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛屼笖绂佹鍒涘缓瀛?agent

## 3. 鏈纭鍒?+
- 鍙湪 Phase F backup/import/rollback 涓荤嚎鎺ㄨ繘
- 姣忚疆蹇呴』褰㈡垚鐪熷疄浠ｇ爜澧為噺鍜屽彲鎵ц楠岃瘉缁撴灉
- 涓嶆敼鍚庣瀵煎叆濂戠害锛屼笉鎵╂暎鍒板叾浠栭〉闈?+
## 4. 鏈绂佹浜嬮」

- 涓嶄慨鏀瑰鍏?鍥炴粴鎺ュ彛璇箟
- 涓嶅仛鏃犲叧 UI 娓呯悊
- 涓嶆妸 no-spawn 楠岃瘉璇姤涓烘祻瑙堝櫒绾?e2e

## 5. 鏈楠屾敹鏉′欢

- `Backup.test.tsx` 鏂板 model mappings 淇敼鍚?pending apply 澶辨晥骞堕渶閲嶆柊 dry-run 鐨勬柇瑷€
- `scripts/verify-backup-component.cjs` 瑕嗙洊鍚屼竴琛屼负璺緞
- `scripts/verify-backup-component.setting-mock.cjs` 涓庢柊鏂█淇濇寔涓€鑷?+- `node scripts/verify-backup-component.cjs`
- `node scripts/verify-backup-logic.mjs`
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`

## 6. 鏈鍥炴粴鐐?+
- 鍥為€€ `web/src/components/modules/setting/Backup.test.tsx`
- 鍥為€€ `scripts/verify-backup-component.cjs`
- 鍥為€€ `scripts/verify-backup-component.setting-mock.cjs`

## 7. 瀹炵幇鑼冨洿

- 鍏堟敼鏁版嵁璇箟杩樻槸鍏堟敼 UI锛氫笉鏀逛骇鍝佽涔夛紝鍙ˉ缁勪欢绾ч獙璇佽鐩栦笌楠岃瘉 mock 瀵归綈
- 鍙楀奖鍝嶅悗绔ā鍧楋細鏃?+- 鍙楀奖鍝嶅墠绔ā鍧楋細`web/src/components/modules/setting/Backup.test.tsx`
- 鍙楀奖鍝嶆帴鍙ｏ細鏃?+- 鏄惁褰卞搷鏃ф暟鎹細鍚?+- 鏄惁褰卞搷鏃ц涓猴細鍚︼紝浠呭鍔犳祴璇曚笌楠岃瘉鑴氭湰鏂█

## 8. 瀹炴柦姝ラ

1. 澶嶆煡 `Backup.tsx` 涓?`onModelMappingsTextChange -> resetImportFormState` 鐨勭姸鎬佹竻鐞嗛摼锛岀‘璁や慨鏀?model mappings 浼氭竻绌?pending apply 鏄棦鏈夎涔?+2. 鍦?`Backup.test.tsx` 鏂板 map 妯″紡鏂█锛氶娆?dry-run 浜у嚭 pending apply锛屼慨鏀?mapping 鏂囨湰鍚?pending apply 娑堝け锛屽啀娆?dry-run 浜у嚭鏂扮殑 preview token 涓庢洿鏂板悗鐨?mapping 缁戝畾
3. 鍦?`scripts/verify-backup-component.cjs` 鍚屾鏂板鍚岃矾寰?no-spawn 鏂█
4. 鍦?`scripts/verify-backup-component.setting-mock.cjs` 鍚屾 map dry-run preview token/mock preview target 鐨勬洿鏂板垎鏀?+5. 杩愯缁勪欢楠岃瘉銆侀€昏緫楠岃瘉涓?web typecheck

## 9. 娴嬭瘯涓庨獙璇?+
- 鏋勫缓鍛戒护锛歚D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 娴嬭瘯鍛戒护锛歚node scripts/verify-backup-component.cjs`
- 涓撻」楠岃瘉锛歚node scripts/verify-backup-logic.mjs`

## 10. 椋庨櫓涓庡吋瀹规€?+
- 鏂伴闄╋細浣庯紱鏂板鏂█浼氬 map 妯″紡缁撴灉鍖轰笌楠岃瘉 mock 鏇存晱鎰?+- 鍏煎鎬ч闄╋細浣庯紱鏈疆娌℃湁浜у搧閫昏緫鍜屾帴鍙ｅ彉鏇?+- 鏄惁闃诲涓嬩竴浠诲姟锛氬惁

## 11. 鏀跺伐璁板綍

- 鏋勫缓鏄惁閫氳繃锛氭槸锛宍tsc --noEmit` 閫氳繃
- 娴嬭瘯鏄惁閫氳繃锛氭槸锛宍verify-backup-component` 涓?`verify-backup-logic` 閫氳繃
- 鏈浣跨敤浜嗗摢浜涙湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囷細automation memory銆乧anonical plan銆亀orkflow銆佸墠绔富绾跨姸鎬佹枃妗ｃ€丳hase F worklog 閾俱€丅ackup 缁勪欢涓?no-spawn 楠岃瘉鑴氭湰
- 鏈浣跨敤鐨勬湰鍦拌祫婧?/ skills / 璁板繂涓婁笅鏂囧垎鍒彁渚涗簡浠€涔堢粨璁猴細memory 淇濇寔鏈疆缁х画鍦?Phase F 鍚岄〉灏忛棴鐜紱canonical/workflow 闄愬畾鍙敹鍙?backup/import/rollback锛涘墠绔富绾跨姸鎬佹枃妗ｇ‘璁?Backup 椤典粛鍙户缁ˉ琛屼负璇佹嵁锛涚幇鏈夎剼鏈彁渚涘彲閲嶅鎵ц鐨勯獙鏀跺叆鍙?+- 鏈浣跨敤浜嗗摢浜涘瓙 agent 鍙婂叾缁撹锛氭棤
- 瀛?agent 鍒嗗伐銆佽礋璐ｈ寖鍥翠笌浜у嚭鎽樿锛氭棤
- 鎵嬪伐 smoke 鐘舵€侊細鏈墽琛?+- 鎵嬪伐 smoke 闃诲鍘熷洜 / 缂哄皯鐨勭幆澧冿細褰撳墠绾跨▼浠嶆棤娴忚鍣ㄨ嚜鍔ㄥ寲鍏ュ彛锛屼笖鏈疆鏈€灏忛棴鐜凡鍙敱 no-spawn 缁勪欢楠岃瘉瑕嗙洊
- 寰呴獙璇侀〉闈㈡竻鍗曪細Backup 椤垫祻瑙堝櫒绾ф墜宸?smoke
- 鑻ユ湭浣跨敤瀛?agent锛屽師鍥狅細鐢ㄦ埛瑕佹眰涓荤嚎绋嬫墽琛屼笖绂佹鍒涘缓瀛?agent
- worklog 鏄惁鏇存柊锛氭槸
- 閬楃暀椤癸細娴忚鍣ㄧ骇 Backup smoke 璇佹嵁浠嶇己澶憋紱Vitest 鍦ㄥ綋鍓嶄富鏈轰粛鍙?esbuild spawn EPERM 褰卞搷
- 涓嬩竴浠诲姟鍓嶇疆鏉′欢鏄惁婊¤冻锛氭槸
- 璁板綍鏃堕棿锛?026-04-22T03:01:08+08:00
