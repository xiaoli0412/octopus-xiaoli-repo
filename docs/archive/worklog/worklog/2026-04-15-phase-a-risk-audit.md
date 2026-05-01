# Phase A - 高风险稳定性审计

> 目的：列出当前最适合在 `Phase A` 优先处理的高风险区域，避免后续 `Phase B` 及之后的阶段继续背着半成品推进。

---

## 1. 审计结论

当前仓库的首要问题已经不是“马上编不过”，而是“很多关键能力仍处于半成品状态，如果不先收口，会拖慢后续所有阶段”。

当前高优先风险区域如下。

---

## 2. 高优先风险清单

### 风险 1：advanced failover runtime 仍是 phase-1 语义

受影响区域：

- `internal/relay/relay.go`
- `internal/relay/balancer/*`

表现：

- 已有 `retry_rounds`
- 已有 `failover_window_sec`
- 已有 `race_after_fails`
- 已有 `race_concurrency`

但仍未形成最终裁决规则和完整运行时闭环。

为什么高风险：

- 这会直接阻塞 `Phase B` 的里程碑 1 收口。

---

### 风险 2：route-target 语义还未形成完整闭环

受影响区域：

- `internal/model/channel.go`
- `internal/model/group.go`
- `internal/op/channel.go`
- `internal/op/group.go`

表现：

- 字段和部分运行时逻辑已经有了
- 但还没有形成完整的 `(channel, key, model)` 一致性闭环与验证

为什么高风险：

- 会影响 key fallback、group retry、策略矩阵和后续探测行为。

---

### 风险 3：核心路由行为缺少自动化测试护栏

受影响区域：

- `internal/relay/*`
- `internal/model/*`
- `internal/op/*`

表现：

- 当前仓库整体构建通过
- 但关键路由行为测试仍明显不足

必须补的行为测试：

- 模型白名单
- 严格顺序
- 严格权重
- 随机
- key fallback
- 响应写出后不再重试

为什么高风险：

- 没有这些测试，后续每次修改都可能悄悄改坏核心语义。

---

### 风险 4：备份/导入仍是“可用但不安全”的半成品

受影响区域：

- `internal/op/backup.go`
- `internal/model/backup.go`
- `web/src/components/modules/setting/Backup.tsx`

表现：

- 已有 `checksum`
- 已有 `dry-run`
- 已有部分兼容性报告

但仍缺：

- `snapshot`
- `diff`
- `apply`
- `rollback`
- 默认 secrets 安全语义闭环

为什么高风险：

- 这会直接阻塞 `Phase F`，并且当前导出语义容易带来安全误判。

---

### 风险 5：首页观测只做到基础 token breakdown

受影响区域：

- `web/src/components/modules/home/*`
- `internal/op/stats.go`
- `internal/server/handlers/stats.go`

表现：

- 已有 total
- 已有 by channel
- 已有 by model

仍缺：

- by provider
- probe cost
- breaker 摘要
- probe 状态摘要

为什么高风险：

- 后续做策略与性能调优时会缺观测面。

---

### 风险 6：日志导出尚未完成“AI 排障闭环”验证

受影响区域：

- `internal/op/log.go`
- `internal/server/handlers/log.go`
- `web/src/components/modules/log/*`

表现：

- 已支持导出
- 已有 attempts 结构

但仍缺：

- 导出字段完整性验证
- attempts 链完整性验证
- 与首页统计的一致性验证

为什么高风险：

- 会拖慢 `Phase C` 的收口。

---

### 风险 7：worklog 体系刚落地，后续阶段还缺正式阶段记录

受影响区域：

- `docs/worklog/*`

表现：

- 现在已有 `Phase A` 活跃 worklog
- 也有两个历史后端子任务记录

仍缺：

- `Phase B`
- `Phase C`
- `Phase D`
- `Phase E`
- `Phase F`
- `Phase G`

对应正式阶段记录

为什么高风险：

- 如果不补齐，后续阶段又会回到“做了但没沉淀”的状态。

---

### 风险 8：当前工作区很脏，核心模块改动集中

受影响区域：

- `internal/model/*`
- `internal/op/*`
- `internal/relay/*`
- `web/src/components/modules/*`

表现：

- 当前未提交改动很多
- 且集中在关键业务路径

为什么高风险：

- 很容易把“稳定性修复”和“业务能力推进”混在一起，增加回归成本和排查成本。

---

## 3. 建议处理顺序

### 第一优先级

- 风险 1：advanced failover runtime 半成品
- 风险 2：route-target 闭环不完整
- 风险 3：核心路由行为测试不足

### 第二优先级

- 风险 4：备份/导入安全闭环不足
- 风险 5：首页观测维度不足
- 风险 6：日志导出排障闭环不足

### 第三优先级

- 风险 7：后续阶段 worklog 体系待补齐
- 风险 8：脏工作区下的混改风险

---

## 4. 用法

本文件用于：

- 进入 `Phase B` 前确认当前还剩哪些真正的高风险项
- 每次选下一任务时，先从第一优先级里挑
- 防止后续开发被低优先级 UI 细节带偏

