# Redacted Snapshot Credential Rebind Targets

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 中 `11.5.2 / 11.5.3 / 11.5.4` 的 backup/import 主线，把 redacted snapshot 导入里原本只有字符串 warning 的“缺失凭据”升级成结构化、可继续处理的 rebind targets。

## 1. 本轮任务信息

- 任务名称：redacted snapshot 缺失凭据重绑定目标结构化
- 对应主线：backup / import / dry-run compatibility / 自动适配 / 缺失 key 占位重绑定
- 本轮聚焦：不做重 wizard，先把后端 contract 与前端展示补齐

## 2. 本轮复用的本地资源

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/worklog/2026-04-18-map-mode-model-mapping-preview-visibility.md`
- `docs/worklog/2026-04-18-import-rollback-preview-minimal-closure.md`
- `internal/model/backup.go`
- `internal/model/channel.go`
- `internal/model/apikey.go`
- `internal/op/backup.go`
- `internal/op/backup_extra.go`
- `internal/op/backup_test.go`
- `web/src/api/endpoints/setting.ts`
- `web/src/components/modules/setting/Backup.tsx`

## 3. 子 agent 与分工

- 子 agent 使用情况：已使用
- 模型：统一使用 `gpt-5.4`
- 子 agent A：只读分析 `11.5.x` 后端剩余缺口，判断下一步最值得写代码的后端项
- 子 agent B：只读分析 Backup UI 当前最值得补的轻量闭环
- 主线程职责：综合结论后直接写代码、跑测试、补 worklog

## 4. 本轮代码改动

### 后端

- 在 `internal/model/backup.go` 新增 `DBImportCredentialRebindTarget`
- 在 `DBImportCompatibilityReport` 中新增 `credential_rebind_targets`
- 在 `DBImportCompatibilitySummary` 中新增：
  - `credential_rebind_targets`
  - `channel_key_rebind_targets`
  - `api_key_rebind_targets`
- 在 `internal/op/backup_extra.go` 中新增：
  - `buildCredentialRebindTargets`
  - `collectSnapshotRouteRefsByChannelAndModel`
  - `summarizeCredentialRouteRefs`
  - `buildAPIKeyRebindContexts`
  - `dedupeSortedStrings`
  - `firstNonEmpty`
- 当前结构化 rebind target 可覆盖：
  - `channel_key`
  - `api_key`
- 当前可返回的信息包括：
  - target type
  - snapshot id
  - channel name
  - key/api-key name
  - source type
  - remark
  - models
  - affected groups
  - contexts
  - warnings

### 前端

- 在 `web/src/api/endpoints/setting.ts` 中补齐 rebind target 类型
- 在 `web/src/components/modules/setting/Backup.tsx` 中新增：
  - `Credential Rebind`
  - `Channel Key Rebind`
  - `API Key Rebind`
  三张 summary card
- 在 compatibility 区新增 `Credential rebind targets` 列表
- 保持 `SummaryCard + CompatibilityList` 的轻量展示方式，不引入新的重交互流程

## 5. 测试补强

- 扩展 `TestDBImportIncrementalDryRunReportsRedactedCredentials`
  - 验证结构化 rebind target summary
  - 验证 channel key / api key 两类 target 都会返回
- 扩展 `TestDBImportIncrementalSkipsEmptyCredentialsOnApply`
  - 验证 apply 后仍返回结构化 rebind target
- 新增 `TestDBImportIncrementalDryRunReportsRedactedCredentialRouteContexts`
  - 验证 channel key 缺失凭据目标会带上模型、受影响 group、route context

## 6. 验证结果

- `gofmt -w internal/model/backup.go internal/op/backup_extra.go internal/op/backup_test.go`
  - 结果：通过
- `go test ./internal/op -count=1`
  - 结果：通过
- `pnpm exec tsc --noEmit`
  - 结果：通过

## 7. 对主文档的推进判断

- 本轮直接推进了 `11.5.2` 中“缺失 key 的占位重绑定”这条要求
- 本轮让 redacted snapshot 从“只会 warning”推进到“后续可继续施工的结构化 contract”
- 本轮也让 dry-run 差异预览更贴近 `11.5.4` 中“真正 apply 前允许用户确认或修正”的方向

## 8. 仍未完成项

- 还没有真正的 placeholder 持久化落库
- 还没有基于 rebind target 的一步式补绑交互
- 还没有把 dry-run 结果与 apply 通过 preview token 强绑定
- 还没有做完整的 conflict wizard / mapping wizard

## 9. 本轮结论

- 当前 backup/import 主线里，redacted snapshot 的最大短板已经不再是“看不见问题”，而是“怎么更进一步完成补绑动作”
- 这轮代码先把后端 contract 与前端展示补齐，为后续做确认条、补绑操作或 preview-token 约束打下基础
