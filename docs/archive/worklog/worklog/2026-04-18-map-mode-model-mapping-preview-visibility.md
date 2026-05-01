# Map Mode Model Mapping Preview Visibility

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 中 backup/import 主线要求，补齐 `map` 模式在 dry-run compatibility 中对 `model_mappings` 的可见性，让用户在 apply 前能看到映射是否真的命中、是否未使用、目标模型是否在当前环境存在。

## 1. 本轮任务信息

- 任务名称：`map` 模式模型映射预览可视化闭环
- 对应主线：backup / import / dry-run compatibility / mapping correction
- 目标聚焦：不扩散到 mapping editor 或 wizard，先把 backend compatibility 与前端 Backup UI 的最小可信预览做实

## 2. 开工前复用的本地资源

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- `docs/worklog/2026-04-18-import-map-mode-minimal-model-mappings.md`
- `internal/model/backup.go`
- `internal/op/backup.go`
- `internal/op/backup_extra.go`
- `internal/op/backup_test.go`
- `web/src/api/endpoints/setting.ts`
- `web/src/components/modules/setting/Backup.tsx`

## 3. 子 agent 与分工

- 子 agent 使用情况：已使用
- 模型：统一使用 `gpt-5.4`
- 子 agent A：只读分析 backend compatibility 中 `map` 模式的结构化映射预览落点
- 子 agent B：只读分析前端 `Backup` 页如何以最小代价展示 mapping preview
- 主线程职责：综合子 agent 结论，完成最终代码修改、测试与 worklog 落地

## 4. 本轮代码改动

### 后端

- 在 `internal/model/backup.go` 新增 `DBImportModelMappingPreview`
- 在 `DBImportCompatibilityReport` 中新增 `model_mapping_previews`
- 在 `DBImportCompatibilitySummary` 中新增以下计数：
  - `model_mapping_previews`
  - `used_model_mappings`
  - `unused_model_mappings`
  - `missing_mapping_targets`
- 调整 `DBImportIncrementalWithOptions`：
  - 保留 `apply scopes` 之后、`apply mappings` 之前的 `originalDump`
  - 导入真正使用 `mapped dump`
  - compatibility 同时使用 `originalDump + mapped dump + model_mappings`
- 在 `internal/op/backup_extra.go` 中拆分预处理：
  - `applyImportScopesToDump`
  - `applyModelMappingsToDump`
- 新增 `buildModelMappingPreviews`
- 新增 `collectDumpModelMappingUsage`
- 新增 `currentStateHasModel`
- 新增 `sortedStringSet`
- mapping preview 现可报告：
  - source model
  - target model
  - contexts
  - touched fields
  - usage count
  - used / unused
  - target exists / missing
  - warnings

### 前端

- 在 `web/src/api/endpoints/setting.ts` 中补齐 compatibility 类型定义
- 在 `web/src/components/modules/setting/Backup.tsx` 中新增：
  - `modelMappingPreviewItems`
  - `Model Mapping Preview` summary card
  - `Used Mappings` summary card
  - `Unused Mappings` summary card
  - `Missing Mapping Targets` summary card
- `map` 模式下新增 `Model mapping previews` 列表展示
- 展示形式继续复用现有 `SummaryCard + CompatibilityList`，未引入新的重交互组件

## 5. 测试补强

- 扩展现有测试 `TestDBImportIncrementalMapModeAppliesModelMappingsToRoutePreviewAndImport`
  - 验证 `model_mapping_previews`
  - 验证 summary 计数
  - 验证 `contexts / touched_fields / usage_count / target_exists`
- 新增测试 `TestDBImportIncrementalMapModeReportsUnusedAndMissingMappingTargets`
  - 验证未使用映射
  - 验证映射目标不存在

## 6. 验证结果

- `gofmt -w internal/model/backup.go internal/op/backup.go internal/op/backup_extra.go internal/op/backup_test.go`
  - 结果：通过
- `go test ./internal/op -count=1`
  - 结果：通过
- `pnpm exec tsc --noEmit`
  - 结果：通过

## 7. 本轮完成度判断

- 已完成：`map` 模式从“能导入”推进到“导入前可解释、可预览、可判断映射质量”
- 已完成：前后端对 `model_mappings` 的结构化反馈闭环
- 仍未完成：
  - 交互式 mapping editor / wizard
  - 更强的 mapping correction UX
  - partial restore 与更细粒度导入选择
  - 更完整的 compare panel

## 8. 与主文档的一致性说明

- 这轮改动直接服务于 backup/import 主线，不偏离 `LLM-Gateway-Refactor-Plan.zh-CN.md`
- 实现重点放在 `dry-run / 差异预览 / 映射修正前可见性`，符合导入前先确认再 apply 的方向
- 本轮没有扩散到额外的重 UI 或无关模块
