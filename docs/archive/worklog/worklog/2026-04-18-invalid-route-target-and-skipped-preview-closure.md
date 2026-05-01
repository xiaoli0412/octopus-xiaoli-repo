# Invalid Route Target And Skipped Preview Closure

> 日期：2026-04-18
>
> 目标：继续按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.4` 收口，把 dry-run 导入里的“失效 route-target / 被跳过 route-target preview”从紧凑 route diff 提升为结构化兼容性预览，并接入前端确认闭环。

## 1. 本轮任务信息

- 任务名称：invalid route-target / skipped route-target preview 结构化闭环
- 对应主线：`11.5.4 导入后模拟路由验证 / 迁移差异预览`
- 目标聚焦：
  - 标出导入后失效目标
  - 标出哪些 route-target preview 会被 skip mode 保留现状而跳过
  - 在 apply 前把这些高风险信号纳入确认

## 2. 本轮复用的本地资源

- `AGENTS.md`
- `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 当前线程 handoff 与最近 backup/import worklog
- `internal/model/backup.go`
- `internal/op/backup_extra.go`
- `internal/op/backup_test.go`
- `web/src/api/endpoints/setting.ts`
- `web/src/components/modules/setting/Backup.tsx`
- Windows 本地 `codex-run-as-apply-patch` 补丁落盘经验

## 3. 子 agent 与分工

- 子 agent 使用情况：已使用
- 子 agent 模型：统一使用 `gpt-5.4`
- 本轮子 agent 分工：
  - 后端只读审计：盘点 `route preview candidate` 现有字段，给出最小新增结构与测试建议
  - 前端只读审计：给出 summary 卡、CompatibilityList 位置，以及是否并入 `Apply This Dry-Run` 风险确认
- 主线程职责：综合审计结论后直接完成代码实现、验证、worklog 记录

## 4. 本轮代码改动

### 后端

- 在 `internal/model/backup.go` 新增：
  - `DBImportRoutePreviewIssue`
  - `DBImportCompatibilityReport.InvalidRouteTargets`
  - `DBImportCompatibilityReport.SkippedRoutePreviews`
  - `DBImportCompatibilitySummary.InvalidRouteTargets`
  - `DBImportCompatibilitySummary.SkippedRoutePreviews`
- 在 `internal/op/backup_extra.go` 新增：
  - `buildRoutePreviewIssues()`
  - `classifyRoutePreviewIssue()`
  - `dedupeRoutePreviewIssues()`
- 在 dry-run compatibility 装配阶段：
  - 基于现有 `RoutePreviewDiffs` 生成 `invalid_route_targets`
  - 基于 `skip_mode_preserved_existing_group:*` 生成 `skipped_route_target_previews`
  - 将两项计数写入 summary
  - 将两项提示写入 `route_preview_warnings`

### 目前判定规则

- `invalid_route_target`
  - channel disabled
  - undeclared model
  - missing key
  - missing model
- `skipped_route_target_preview`
  - skip mode 下同名 group 已存在，preview 保留现状而跳过 snapshot route preview

### 前端

- 在 `web/src/api/endpoints/setting.ts` 补齐 compatibility 类型字段：
  - `summary.invalid_route_targets`
  - `summary.skipped_route_target_previews`
  - `summary.route_preview_diffs`
  - `invalid_route_targets`
  - `skipped_route_target_previews`
- 在 `web/src/components/modules/setting/Backup.tsx` 新增：
  - `formatRoutePreviewIssueItem()`
  - `invalidRouteTargetItems`
  - `skippedRoutePreviewItems`
  - summary 卡：
    - `Invalid Route Targets`
    - `Skipped Route Preview`
  - compatibility 列表：
    - `Invalid route targets`
    - `Skipped route-target previews`
- 将两项高风险信号并入 `Apply This Dry-Run` 的确认逻辑：
  - `Invalid route targets detected`
  - `Skipped route-target previews detected`

## 5. 新增测试

- `TestDBImportIncrementalDryRunReportsInvalidRouteTargetForUndeclaredModel`
- `TestDBImportIncrementalDryRunReportsInvalidRouteTargetForMissingKey`
- `TestDBImportIncrementalDryRunReportsSkippedRouteTargetPreviewInSkipMode`

## 6. 验证结果

- `gofmt -w internal/model/backup.go internal/op/backup_extra.go internal/op/backup_test.go`
  - 结果：通过
- `go test ./internal/op -count=1`
  - 结果：通过
- `pnpm exec tsc --noEmit`
  - 结果：通过

## 7. 对主 MD 的直接对应

- 对应 `11.5.4`：
  - 已补“标出导入后失效目标”
  - 已补“标出哪些 route-target 会被跳过”中的结构化 preview 部分
  - 已补“导入应用前允许用户确认”中的风险确认接入

## 8. 本轮结论

- 这轮不是新造一套 route diff 系统，而是在现有 route preview 主线上补了一层更接近用户决策的结构化 issue 视图
- dry-run 现在不仅能看候选链变化，还能明确看到：
  - 哪些 route-target 本身无效
  - 哪些 preview 因 skip mode 被保留现状而未真正展开
- 仍可继续加强的点：
  - 把 `channel_disabled / missing_model / undeclared_model / missing_key` 做成更细的 UI 过滤视图
  - 为 rollback preview 同步补结构化 issue 面板，而不只透出 warnings
  - 将 route preview diff 从字符串列表升级为更强的交互对比面板
