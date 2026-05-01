# Import Rollback Latest Snapshot Minimal Closure

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.2 / 11.5.3 / 11.5.4`，把当前缺失的 rollback 主线推进成“导入前自动保存 snapshot + 一键回滚到最近一次 snapshot”的最小可用闭环。

## 1. 任务信息

- 任务名称：最近一次导入快照回滚最小闭环
- 日期：2026-04-18
- 当前阶段：backup/import 主线继续推进
- 对应 milestone：`11.5.2` 回滚、`11.5.3` 一键回滚到上一个 snapshot、`11.5.4` 保留并增强 rollback 能力

## 2. 开工前输入

- 对应 canonical 章节：`11.5.2 / 11.5.3 / 11.5.4 / 14.12`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-18-import-identity-normalization-tests-and-validation.md`
  - `docs/worklog/2026-04-18-import-map-mode-minimal-model-mappings.md`
- 本次任务目标：
  - 导入 apply 前自动导出 full snapshot
  - 落盘到数据库同目录下的 `import-snapshots/`
  - 记录 latest metadata
  - 增加“回滚最近一次导入快照”的后端能力、API 入口与最小前端入口
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `internal/db/db.go`
  - `internal/server/handlers/setting.go`
  - `web/src/api/endpoints/setting.ts`
  - `web/src/components/modules/setting/Backup.tsx`
- 本次是否启用子 agent 与分工边界：本轮主要由主线程直接推进；子 agent 只读核对 rollback 可复用资产与最小改动面
- 本次子 agent 使用模型：`gpt-5.4`

## 3. 本次硬规则

- 只做 backup/import 主线的 rollback，不扩散到无关模块
- 本地资源优先，复用现有 `DBExportAll` 与导入链，不重造第二套备份体系
- 子 agent 默认统一使用 `gpt-5.4`
- 先做最小可用闭环，不先上历史浏览器、复杂向导或多快照管理页

## 4. 本次验收条件

- 非 dry-run 导入前会自动保存 pre-import snapshot
- snapshot 文件可落盘并可重新读取成 `DBDump`
- 可以回滚到最近一次 snapshot
- rollback 后结果能恢复到导入前状态，而不是只做普通 replace import
- 管理端存在可调用 API 入口
- 前端存在最小按钮入口
- `go test ./internal/op`
- `go test ./internal/relay/...`
- `pnpm build`

## 5. 实施步骤

1. 增加 rollback 结果结构与快照 metadata 结构
2. 复用 `DBExportAll`，在 apply 前导出 full snapshot 并落盘
3. 记录 `latest-import-snapshot.json`
4. 增加最近一次 snapshot 的读取与 rollback 函数
5. 为 rollback 增加全量重置核心表的事务逻辑，避免仅 replace 导入导致“新增资源残留”
6. 补专项测试
7. 增加 handler API 入口与前端最小按钮
8. 运行 Go 与前端验证

## 6. 测试与验证

- 格式化命令：`gofmt -w internal/model/backup.go internal/db/db.go internal/op/backup.go internal/op/backup_test.go internal/server/handlers/setting.go`
  - 结果：通过
- 测试命令：`go test ./internal/op`
  - 结果：通过
- 测试命令：`go test ./internal/server/handlers/...`
  - 结果：当前路径无测试文件，非失败
- 测试命令：`go test ./internal/relay/...`
  - 结果：通过
- 前端验证：`pnpm build`
  - 结果：通过
- 新增专项测试：
  - `TestDBImportIncrementalApplySavesPreImportSnapshot`
  - `TestDBRollbackLatestImportSnapshotRestoresPreviousState`

## 7. 收工记录

- 本轮后端新增：
  - 导入前自动保存 snapshot
  - `import-snapshots/` 目录与 latest metadata 文件
  - 最近一次快照回滚函数
  - rollback 专用全量清理核心表逻辑
- 本轮 API 新增：
  - `POST /api/v1/setting/rollback-latest-import`
- 本轮前端新增：
  - `useRollbackLatestImportSnapshot`
  - Backup 页面最小 rollback 按钮与说明
- 本轮已经覆盖的 rollback 语义：
  - apply 前自动存档
  - 最近一次回滚
  - 回滚后恢复到导入前的核心配置状态
- 本轮仍未完成：
  - 历史快照列表
  - 多快照选择回滚
  - 回滚前差异预览
  - 与 `map`/partial restore 联动的交互式回滚向导

## 8. 子 agent 与本地资源记录

- 本次子 agent 默认统一使用 `gpt-5.4`
- 子 agent 只读职责：核对 rollback 是否已有落盘目录、恢复逻辑或可复用脚本，并给出最小改动面建议
- 主线程职责：吸收结论后直接完成 rollback 后端实现、测试、API 入口与最小前端接线

## 9. 风险与遗留项

- 当前 rollback 是“最近一次 snapshot”的最小闭环，还不是完整历史管理系统
- 非 sqlite 数据库当前落盘目录采用 `data/import-snapshots/` 的保守兜底，后续可继续细化
- rollback 目前是管理端动作，尚未增加更细的确认向导和风险提示
- `11.5.x` 仍未完成的主线还包括：
  - 交互式 mapping editor / conflict wizard
  - partial restore / selective import
  - 更细粒度 provider/channel/group 映射

