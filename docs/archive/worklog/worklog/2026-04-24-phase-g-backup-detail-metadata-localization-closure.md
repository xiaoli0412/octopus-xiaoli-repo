# 2026-04-24 Phase G Backup Detail Metadata Localization Closure

## 1. 任务信息

- 任务名称：备份详情字段与枚举内值中文本地化收口
- 日期：2026-04-24
- 当前阶段：Phase G screenshot-first UI closure
- 对应 milestone：`Phase G settings / backup no-browser contract closure`

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9 节前端收口与中文一致性要求
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3 节

## 2. 本轮问题与目标

- 目标：继续收口备份导入详情内 `contexts / touched_fields / impact / enum` 的中文展示残留，避免中文主显示出现 `routing / fallback / api_keys / primary_model / fallback_model / high` 等内部值。
- 范围：仅 `web/src/components/modules/setting/backup-logic.ts` 的输出语义 + `web/src/components/modules/setting/backup-logic.test.ts` + `scripts/verify-backup-logic.mjs` 的 no-browser 断言。

## 3. 实施内容

1. 在 `backup-logic.ts` 新增最小本地化映射层：
   - context：`routing / api_keys / fallback`
   - field：`primary_model / fallback_model / billing_mode / probe_policy`
   - impact/enum：`high / medium / low / paid / free / manual / auto / per_request / per_token / per_quota / flat / unknown / passive_only / sparse_single / sequential / concurrent`
2. 增加 helper：`localizeKnownToken`、`localizeTokenList`、`formatFieldList`、`formatPolicyValue`。
3. 修改 formatter 输出：
   - `formatContexts` 使用 context 映射
   - `formatModelMappingPreview` 的 `受影响字段` 使用字段映射
   - `formatModelPolicyDiff` 的 `影响级别` 与 `变更字段` 使用本地化映射
   - `formatModelPolicyState` 的 `计费 / 探测` 值使用本地化映射（含 unknown 回退）
4. 同步更新 `backup-logic.test.ts` 与 `scripts/verify-backup-logic.mjs` 中 zh-Hans 期望值。

## 4. 验证

- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- `node scripts/verify-backup-logic.mjs`

## 5. 结果

- 构建：通过
- no-browser 验证：通过
- 说明：未引入后端/接口变更，行为上仅影响非英文 locale 的备份详情文案层。

## 6. 遗留与下步

- 本轮主要闭口了 `contexts / fields / impact / enum` 直出层；仍建议下一轮复查该模块里 `normalization / pricing / skip 原因` 其余细项，并同步到 settings help 文案一致性脚本。

## 7. 使用子协作

- 无（遵循用户不创建子 agent 要求）
