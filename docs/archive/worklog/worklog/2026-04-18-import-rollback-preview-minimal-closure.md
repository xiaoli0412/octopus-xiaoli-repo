# Import Rollback Preview Minimal Closure

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.4`，把 rollback 从“可选快照并执行”推进到“回滚前可预览高价值差异摘要”的最小闭环。

## 1. 任务信息

- 任务名称：rollback preview / snapshot diff 最小闭环
- 日期：2026-04-18
- 当前阶段：backup/import 主线持续推进
- 对应 milestone：`11.5.4` 迁移差异预览、导入/回滚前风险提示

## 2. 开工前输入

- 对应 canonical 章节：`11.5.3 / 11.5.4 / 验收 / 里程碑`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-18-import-snapshot-history-and-targeted-rollback.md`
  - `docs/worklog/2026-04-18-import-selective-scopes-and-rollback-chain-fix.md`
- 本次任务目标：
  - 新增按指定 snapshot 的 rollback preview API
  - 复用现有 compatibility / route preview diff 逻辑，输出高价值回滚前摘要
  - 在 Backup 页面历史快照列表中加入 Preview 按钮和最小预览区域
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `internal/model/backup.go`
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `internal/server/handlers/setting.go`
  - `web/src/api/endpoints/setting.ts`
  - `web/src/components/modules/setting/Backup.tsx`
- 本次是否启用子 agent 与分工边界：本轮子 agent 槽位受限，主线程直接推进；仍复用前序 `gpt-5.4` 子 agent 对 rollback/history 与前端接入方案的只读结论

## 3. 本次硬规则

- 继续只做 backup/import 主线
- 复用现有 compatibility / route preview / rows summary 逻辑，不新造第二套 diff 引擎
- 先做最小 preview 摘要，不上复杂对话框或交互式 compare
- 记录单独写入 `docs/worklog/`

## 4. 本次验收条件

- 后端支持按 `snapshot_name` 获取 rollback preview
- preview 返回 manifest、rows summary、compatibility、preview warnings
- 前端历史快照列表支持 Preview 按钮
- Preview 区域能展示 rows、warnings、missing providers/models、route diff 摘要、base URL mismatch
- `go test ./internal/op`
- `pnpm build`

## 5. 实施步骤

1. 在 model 层新增 `DBRollbackPreviewResult`
2. 在 op 层新增 `DBPreviewRollbackImportSnapshot()`
3. 复用 `buildImportCompatibility()` 生成 preview compatibility
4. 增加 `rows_summary` 与 `preview_warnings` 辅助构建函数
5. 新增 handler：`POST /api/v1/setting/preview-rollback-import-snapshot`
6. 增加专项回归测试，验证 preview 的高价值字段与 warning 输出
7. 前端 `setting.ts` 增加 preview hook
8. Backup 页面历史列表增加 Preview 按钮和预览摘要区

## 6. 测试与验证

- 格式化命令：`gofmt -w internal/model/backup.go internal/op/backup.go internal/server/handlers/setting.go internal/op/backup_test.go`
  - 结果：通过
- 测试命令：`go test ./internal/op`
  - 结果：通过
- 前端构建：`pnpm build`
  - 结果：通过
- 新增专项测试：
  - `TestDBPreviewRollbackImportSnapshotBuildsCompatibilityAndRowsSummary`

## 7. 收工记录

- 本次后端新增：
  - `DBRollbackPreviewResult`
  - `DBPreviewRollbackImportSnapshot()`
  - `buildDumpRowsSummary()`
  - `buildRollbackPreviewWarnings()`
  - `POST /api/v1/setting/preview-rollback-import-snapshot`
- 本次前端新增：
  - `usePreviewRollbackImportSnapshot()`
  - 历史快照行内 `Preview` 按钮
  - rollback preview 摘要区
- 本次推进后的 MD 结论：
  - `11.5.3`：rollback 历史选择闭环已具备
  - `11.5.4`：rollback preview 已具备最小摘要能力，不再只有“执行回滚”单动作
  - `11.5.4`：仍未完成交互式 diff、映射确认向导、回滚前完整 compare 视图
- 遗留项：
  - rollback preview 目前是摘要，不是完整可交互 compare
  - 历史快照没有删除/清理策略 UI
  - 更细粒度 partial restore 仍停留在 domain 级 scopes
