# Octopus 动态路由需求说明

> 当前范围（2026-04-23 主线更新版）：本文档描述当前仓库内已经接线的动态路由能力，包括 `dynamic_routing_mode`、`dynamic_routing_health_enabled`、五个 `race_*_budget`、relay 内推荐/回退/保守模式层、每日 `dynamic summary scan` 摘要链路，以及设置页与摘要 API 的真实消费闭环；同时包含已经接线生效的 `dynamic_routing_learning_enabled` 本地 AI 学习专线。
>
> 当前实现说明：`shadow-ai`、`hybrid`、`metrics-only`、`strict-mechanism`、`incident-safe` 已有对应设置面、运行时状态与测试。这里的“AI”优先指本地可解释推荐和学习层，不是外接独立大模型服务；外部 AI 自动化中心只负责整理建议和生成 AI Profile，不直接覆盖动态路由。
>
> 执行硬规则：本文档是动态路由子域需求文档，必须同时对齐 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md) 与 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)；若与主规划或用户上下文总账冲突，以两者为准。
>
> 当前执行环境统一按 `Codex` 口径；文档中的自动化、agent 或分目录协作描述必须能在 `Codex` 当前环境下直接落实。

---

## 0. 当前会话补充约束（2026-04-22）

虽然本文档聚焦动态路由子域，但当前会话新增的以下总账要求同样适用：

- 动态路由相关界面在中文场景下必须全中文化，不能泄漏英文枚举值或内部技术状态词。
- 涉及探测、熔断、动态健康的前端入口必须放在设置模块的信息架构内，不得继续混入价格模块。
- 若动态路由或动态健康设置在 UI 中暴露给用户，必须补充帮助提示，并采用渐进展开交互而不是一次性铺开。
- 若采用自动化 agent 或分目录 agent 实施动态路由子任务，必须先定义负责目录、边界和回写主文档的方式。
- 动态路由纳入 `AI 自动化中心 + AI Profile 双轨配置` 主线时，必须作为例外专线处理：动态路由不走普通 AI Profile 覆盖逻辑，只通过设置页手动切换本地机制和 AI 学习。
- `dynamic_routing_learning_enabled` 是动态路由本地 AI 学习开关；关闭后学习数据可以保留，但不能参与推荐排序。

---

## 1. 当前产品目标

当前目标是：

1. 让 failover / race 路径具备可切换的多模式动态路由层。
2. 让并发竞速行为受到全局到 probe 级预算约束。
3. 让 `hybrid` 能在信号可信时采用推荐顺序，信号不足时自动回退到机制路径。
4. 让 `shadow-ai` / `metrics-only` 在不改动真实执行顺序的前提下输出推荐审计。
5. 提供每日动态路由摘要，帮助运维理解当前模式、生效模式与候选池结构。
6. 提供本地 AI 学习闭环，按 `(channel, key, model)` 学习成功率、失败率、延迟、fallback 和 race winner，并只影响运行时推荐。

---

## 2. 当前已交付范围

当前 shipped 范围仅包括以下几类能力：

- 运行模式：`dynamic_routing_mode`
- 运行时健康开关：`dynamic_routing_health_enabled`
- 竞速预算：`race_global_budget`
- 竞速预算：`race_group_budget`
- 竞速预算：`race_channel_budget`
- 竞速预算：`race_key_budget`
- 竞速预算：`race_probe_budget`
- relay 内推荐层与模式状态切换
- relay log 中的动态路由审计字段
- 每日 `dynamic summary scan` 摘要任务
- `stats` API 对动态路由摘要的暴露
- 设置页对动态路由模式、健康开关、预算和摘要的展示

当前已接线范围同时包括：

- 设置项：`dynamic_routing_learning_enabled`
- 学习状态：`DynamicRouteLearningState`
- 管理 API：`GET /api/v1/dynamic-routing/learning`
- 管理 API：`POST /api/v1/dynamic-routing/learning/reset`
- relay 完成后学习写入
- `hybrid` 推荐评分读取学习状态
- 设置页展示学习摘要、学习开关和清空入口

这些能力的目标是约束、观测并在可控边界内调整当前 relay 里的竞速/回退行为。

---

## 3. 运行时行为要求

### 3.1 健康开关

- 当 `dynamic_routing_health_enabled=false` 时，relay 必须退回 group 现有默认竞速参数，不再应用动态健康调节。
- 当 `dynamic_routing_health_enabled=true` 时，relay 可以基于 route-target policy 对 `RaceAfterFails` 与 `RaceConcurrency` 做运行时调节。
- 这些调节只允许影响当前请求的运行时决策，不允许回写用户配置。

### 3.2 运行模式

- `strict-mechanism`：必须使用确定性机制路径，不采用推荐排序，但保留现有 failover / race 机制。
- `metrics-only`：必须计算推荐结果并输出审计，但不能改变真实候选顺序。
- `shadow-ai`：必须生成影子推荐与审计，但不能改变真实候选顺序。
- `hybrid`：在运行时信号可信时采用推荐顺序，信号不足时自动回退到机制路径。
- `incident-safe`：必须进入保守路径，收紧竞速并优先降低事故期间的扩散风险。

### 3.3 竞速预算

- 所有竞速探测都必须受预算限制。
- 预算至少覆盖全局、group、channel、key、probe 五个层级。
- 当预算不足时，系统必须跳过对应竞速候选，而不是突破预算继续探测。

### 3.4 运行时边界

- 动态健康调节只能影响 failover / race 路径中的运行时阈值与并发度。
- 推荐层只能在 `hybrid` 模式且信号可信时改变本次请求的候选顺序。
- 动态健康调节不能持久化写入新的阈值。
- 动态路由审计必须明确记录当前模式、生效模式、决策类型、推荐顺序与回退原因。

### 3.5 本地 AI 学习

- 本地 AI 学习按 `(channel_id, channel_key_id, model_name)` 粒度记录学习状态。
- 学习输入包括成功、失败、状态码、延迟、fallback、race winner 和最近结果时间。
- 学习输出包括 `score` 与 `confidence`，用于辅助 `hybrid` 推荐评分。
- `shadow-ai` 和 `metrics-only` 可以记录学习和审计，但不能改变真实候选顺序。
- `strict-mechanism` 必须保持确定性机制路径。
- `incident-safe` 必须保持保守策略，不允许学习层激进提权。
- 当 `dynamic_routing_learning_enabled=false` 时，学习状态不能参与推荐评分。
- 学习层不得写回 `group_items`，不得覆盖用户 priority，不得永久重排用户配置。

---

## 4. 配置项要求

当前版本动态路由只要求以下配置项：

- `dynamic_routing_mode`：选择当前动态路由模式。
- `dynamic_routing_health_enabled`：启用或关闭运行时动态健康调节。
- `race_global_budget`：全局竞速预算上限。
- `race_group_budget`：单 group 竞速预算上限。
- `race_channel_budget`：单 channel 竞速预算上限。
- `race_key_budget`：单 key 竞速预算上限。
- `race_probe_budget`：单 probe 并发预算上限。
- `dynamic_routing_learning_enabled`：启用或关闭动态路由本地 AI 学习参与推荐。

文档、迁移、后端设置校验、前端设置项必须对这些配置保持一致。

---

## 5. 每日摘要扫描要求

`dynamic summary scan` 的要求如下：

- 按天运行一次。
- 读取动态健康开关状态。
- 读取当前动态路由模式，并给出当前模式 / 生效模式 / 决策基调。
- 汇总 channel 数量、启用 channel 数量、group 数量、failover group 数量。
- 汇总不同 key source type 的数量分布。
- 产生明确的 `basis`，当前 basis 应为 `daily_summary_scan_no_runtime_mutation`。
- 当健康开关关闭时，任务可以跳过扫描，但必须留下可解释的 `skipped` 状态与消息。

该任务的职责仅限于摘要与观测：

- 不持久化改写新的动态阈值。
- 不改写用户路由顺序。
- 不直接写回新的模式决策，只做摘要与观测输出。

---

## 6. 可观测性与接入面要求

当前版本至少需要以下接入面：

- 管理 API 提供动态路由摘要读取入口。
- 设置页提供模式、健康开关与预算配置入口。
- 设置页展示 `dynamic summary scan` 结果。
- 摘要状态、模式、生效模式、决策类型、basis、最近执行信息必须能被 UI 解释。

如果摘要链路存在但没有任何 API 或 UI 消费方，不应计为闭环完成。

---

## 7. 明确非目标

以下内容仍不是动态路由需求，不得写成 shipped 行为或发布门槛：

- 外接独立大模型服务做在线推荐
- 自动持久化阈值调优
- 自动永久重排用户配置路由顺序
- 普通 AI Profile 覆盖动态路由配置

以下内容已从非目标调整为当前正式目标：

- 本地、可解释、按 `(channel, key, model)` 粒度记录的 AI 学习闭环。
- 受 `dynamic_routing_learning_enabled` 控制的运行时学习评分。

---

## 8. 验收标准

以下条件全部满足后，当前版本动态路由才算完成：

1. 默认设置、迁移和校验已包含 `dynamic_routing_mode`、`dynamic_routing_health_enabled` 与全部 `race_*_budget` 配置项。
2. relay 运行时真实消费模式、健康开关与预算配置，而不是只在文档或设置表中存在。
3. `shadow-ai` / `metrics-only` / `strict-mechanism` / `hybrid` / `incident-safe` 均有可解释的运行时差异与测试覆盖。
4. `dynamic summary scan` 每日运行时只产出摘要，不持久化写回新阈值，也不永久改写用户路由顺序。
5. 动态路由摘要可通过管理 API 读取，并在设置页形成真实消费闭环。
6. `dynamic_routing_learning_enabled` 当前已能控制学习状态是否参与推荐排序。
7. 动态路由 AI 学习不写回 `group_items`，不覆盖用户 priority，并有测试证明。

