# Import Snapshot History And Targeted Rollback

> 日期：2026-04-18
>
> 目标：按 `docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 的 `11.5.3 / 11.5.4`，把 rollback 从“仅 latest 快照”推进到“可浏览历史快照 + 按指定快照回滚”的最小闭环，并修正快照文件名秒级碰撞风险。

## 1. 任务信息

- 任务名称：快照历史列表与指定快照回滚
- 日期：2026-04-18
- 当前阶段：backup/import 主线持续推进
- 对应 milestone：`11.5.3` 一键回滚到上一 snapshot、快照历史扩展，`11.5.4` 导入后迁移安全兜底

## 2. 开工前输入

- 对应 canonical 章节：`11.5.3 / 11.5.4 / 验收 / 里程碑`
- 上一个相关 worklog：
  - `docs/worklog/2026-04-18-import-selective-scopes-and-rollback-chain-fix.md`
  - `docs/worklog/2026-04-18-import-rollback-latest-snapshot-minimal-closure.md`
- 本次任务目标：
  - 给 rollback 增加历史快照列表能力
  - 支持按指定 `snapshot_name` 回滚，而不是只能 latest
  - 收紧快照路径安全边界，拒绝路径穿越
  - 修复快照文件名只有秒级精度导致连续导入可能覆盖的问题
  - 在 Backup 页面接入最小快照历史 UI
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `internal/model/backup.go`
  - `internal/op/backup.go`
  - `internal/op/backup_test.go`
  - `internal/server/handlers/setting.go`
  - `web/src/api/endpoints/setting.ts`
  - `web/src/components/modules/setting/Backup.tsx`
  - 2026-04-18 已有 import worklog
- 本次是否启用子 agent 与分工边界：是
- 本次子 agent 使用模型：`gpt-5.4`

## 3. 本次硬规则

- 继续只推进 backup/import 主线
- 本地资源优先，复用现有 latest rollback 逻辑，不重做第二套回滚框架
- 子 agent 只做只读分析，主线程负责最终代码、测试、构建与记录
- worklog 与项目代码保持分离

## 4. 本次验收条件

- 后端能列出 import snapshot 历史
- 后端能按指定 `snapshot_name` 回滚
- 回滚目标必须限制在 `import-snapshots/` 目录内，拒绝路径穿越
- 连续导入生成的快照文件名保持唯一
- 前端 Backup 页面能看到快照历史并触发指定回滚
- `go test ./internal/op`
- `pnpm build`

## 5. 实施步骤

1. 在 model 层新增快照列表项结构
2. 在 op 层新增快照历史列举、指定快照加载与指定快照回滚
3. 让 latest rollback 复用统一回滚主干
4. 收紧快照名/路径校验并限定目录边界
5. 将快照文件名从秒级扩展为秒级 + 纳秒后缀
6. 增加快照历史与指定回滚相关回归测试
7. 新增 handler 路由：快照列表、指定快照回滚
8. 前端接入历史列表与指定回滚按钮

## 6. 测试与验证

- 格式化命令：`gofmt -w internal/model/backup.go internal/op/backup.go internal/server/handlers/setting.go internal/op/backup_test.go`
  - 结果：通过
- 测试命令：`go test ./internal/op`
  - 结果：通过
- 前端构建：`pnpm build`
  - 结果：通过
- 新增专项测试：
  - `TestDBListImportSnapshotsReturnsLatestFirstAndMarksLatest`
  - `TestDBRollbackImportSnapshotRestoresSpecifiedHistoricalSnapshot`
  - `TestDBRollbackImportSnapshotRejectsPathTraversal`

## 7. 收工记录

- 本次后端新增：
  - `DBImportSnapshotInfo`
  - `DBListImportSnapshots()`
  - `DBRollbackImportSnapshot(ctx, snapshotName)`
  - `loadImportSnapshotByName()` / `loadImportSnapshotByPath()` / `resolveImportSnapshotPath()`
  - `buildImportSnapshotFilename()`
- 本次 handler 新增：
  - `GET /api/v1/setting/import-snapshots`
  - `POST /api/v1/setting/rollback-import-snapshot`
- 本次前端新增：
  - `useImportSnapshots()`
  - `useRollbackImportSnapshot()`
  - Backup 页面快照历史展开区与逐条回滚按钮
- 本次子 agent：
  - `019d9d59-8a18-7e43-9cd9-e8c2ac8d6639`，模型 `gpt-5.4`
    - 负责前端只读分析，给出历史快照 UI 与 hooks 最小接入方案
  - `019d9cca-2d30-7d92-94b8-ffe48a6d068b`，模型 `gpt-5.4`
    - 负责后端只读分析，给出历史快照列表、指定快照回滚与安全边界建议
- 当前对 MD 的推进结论：
  - `11.5.3`：rollback 已从 latest-only 推进到历史快照可见、可选定目标回滚的最小闭环
  - `11.5.3`：partial restore 仍是 domain 级，不是更细粒度 channel/group/provider 级
  - `11.5.4`：导入安全兜底继续增强，但 rollback preview、历史 diff 与交互式确认仍未完成
- 遗留项：
  - 历史快照当前仍无差异预览与删除能力
  - 指定回滚目前按 `snapshot_name`，还没有可视化 preview / compare
  - 前端仍缺更完整的 rollback history browser 和交互式风险提示
