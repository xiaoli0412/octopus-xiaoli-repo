# Dry-Run To Apply Same Import Closure

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 backup/import 主线，把当前已经具备的 `dry-run / 差异预览 / 映射修正 / 回滚` 能力，补成一个轻量的“看完 dry-run 后可直接 apply 同一份导入请求”的闭环。

## 1. 本轮任务信息

- 任务名称：`dry-run -> Apply same import` 轻量确认闭环
- 对应主线：backup / import / dry-run / apply before confirm
- 目标聚焦：不引入重 wizard，不改变后端协议，直接复用当前 dry-run 结果和导入参数

## 2. 本轮复用的本地资源

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/worklog/2026-04-18-map-mode-model-mapping-preview-visibility.md`
- `docs/worklog/2026-04-18-redacted-snapshot-credential-rebind-targets.md`
- `web/src/components/modules/setting/Backup.tsx`
- `web/src/api/endpoints/setting.ts`

## 3. 子 agent 与分工

- 子 agent 使用情况：已使用
- 模型：统一使用 `gpt-5.4`
- 子 agent：只读分析当前 Backup UI 中最小但高价值的下一步闭环
- 主线程职责：综合只读建议后直接写前端代码与验证

## 4. 本轮代码改动

### 前端

- 在 `web/src/components/modules/setting/Backup.tsx` 新增 `PreparedImportRequest`
- 在 dry-run 成功后，缓存当前导入请求：
  - file
  - mode
  - modelMappings
  - importScopes
  - fileName
  - mappingCount
  - scopeLabels
- 新增 `executeImport()` 复用导入逻辑，统一 dry-run 与 apply
- 在结果区新增 `Apply This Dry-Run` 面板
- 面板可展示：
  - 本次将 apply 的文件
  - mode
  - mappings 数量
  - scopes
- 新增 apply 风险聚合信号：
  - replace mode
  - replace-prune preview
  - compatibility conflicts
  - route conflicts
  - credential rebind
  - missing mapping targets
  - base URL mismatches
  - schema mismatches
- 当存在高风险信号时，要求显式确认后才能执行 `Apply Same Import`

## 5. 本轮为什么符合主文档

- 这一步没有偏离 backup/import 主线
- 它直接服务于 `apply 前允许用户确认或修正映射` 的方向
- 它复用了现有 dry-run 和 compatibility 结果，没有额外发明一套重交互流程

## 6. 验证结果

- `pnpm exec tsc --noEmit`
  - 结果：通过

## 7. 本轮结论

- 当前 Backup 页已从“只能看到 dry-run 结果”推进到“可基于同一份 dry-run 结果直接 apply”
- 这让 import 主线更接近真正可用的操作闭环，而不是停留在报告层
- 仍未完成：
  - preview token / checksum 级别的 server-side 绑定
  - 更强的交互式 conflict wizard
  - 更细粒度 partial restore
