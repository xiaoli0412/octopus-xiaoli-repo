# 2026-04-23 AI 自动化中心主线规划

## Context

- Task: 将 `AI 自动化中心 + AI Profile 双轨配置 + 动态路由 AI 学习` 并入项目主线。
- Master plan aligned before coding (yes/no): yes
- Scope: 文档与主线同步，不改产品代码，不改数据库，不触发迁移。
- Local resources used: `AGENTS.md`、canonical plan、用户需求总账、详细工作流、动态路由文档、前端主线状态、环境 next plan、当前线程上下文。

## Decisions

- 新增顶层 `AI 自动化` 栏目作为所有 AI 自动化任务的统一入口。
- 用户手动配置和 AI 生成 Profile 必须双轨并存。
- AI 生成分组、渠道识别、价格识别、模型归类等结果保存为 AI Profile，不覆盖手动配置。
- 设置页负责手动切换 `manual` 与 `ai_profile`。
- 动态路由不走普通 AI Profile 覆盖体系，单独保留本地机制与 AI 学习开关。
- 动态路由 AI 学习只影响运行时推荐，不写回 `group_items`，不覆盖 priority。

## Documentation Changes

- 新增 `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`。
- 新增 `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.md`。
- 新增 `docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`。
- 新增 `docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.md`。
- 计划同步更新 canonical plan、用户需求总账、详细工作流、当前状态、前端主线状态、环境 next plan 和动态路由文档。

## Implementation Notes

- 本轮不实现后端模型、API、前端页面或动态学习算法。
- 后续实现必须先对齐新增四份 AI 自动化文档。
- 第一阶段默认只允许 AI Profile 保存、预览、激活和回退，不开放静默覆盖写入。

## Validation Plan

- 搜索确认新增文档存在。
- 搜索确认主线文档包含 `AI 自动化`、`AI Profile`、`dynamic_routing_learning_enabled`、`不覆盖用户配置`。
- 搜索确认动态路由文档不再把在线学习列为非目标。

## Status

- Result: completed
- Verification: 文档落盘与关键术语搜索已完成；动态路由文档中“在线学习非目标”旧口径已清理。
- Next: 进入 Phase H2/H3 前，先按新增 AI 自动化文档设计后端模型、设置项、迁移和 API。
