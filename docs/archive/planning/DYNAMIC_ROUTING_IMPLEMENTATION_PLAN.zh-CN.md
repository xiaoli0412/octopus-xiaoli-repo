# Octopus 动态路由实施方案

> 当前范围（2026-04-23 主线更新版）：本文档覆盖当前已接线的动态路由实现主线，即 `dynamic_routing_mode`、`dynamic_routing_health_enabled`、五个 `race_*_budget`、relay 内推荐/回退/保守模式层、每日 `dynamic summary scan`、动态路由摘要 API，以及设置页对应的配置与摘要展示；同时包含已接线的 `dynamic_routing_learning_enabled` 本地 AI 学习专线。
>
> 当前实现说明：`shadow-ai`、`hybrid`、`metrics-only`、`strict-mechanism`、`incident-safe` 已有对应实现与测试。这里的“AI”优先指本地可解释推荐与学习层，而不是外接独立大模型服务；外部 AI 自动化中心只负责生成建议和 AI Profile，不直接覆盖动态路由。
>
> 执行硬规则：本文档的实现步骤必须对齐 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md)、[DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 与 [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)；冲突时以前述文档为准。
>
> 当前执行环境统一按 `Codex` 口径；若采用自动化 agent 或分目录 agent 并行实施，必须先定义负责目录和越界限制，再进入编码。

---

## 0. 当前会话实施补充（2026-04-22）

在动态路由实施层，当前会话新增以下执行要求：

- 动态路由子任务开工前同样需要完成主规划与用户上下文总账对齐，并在 worklog 记录 `Master plan aligned before coding (yes/no):`。
- 如果任务拆给分目录 agent，需要在实施计划或 worklog 中明确目录边界和禁止越界范围。
- 动态路由相关 UI 改动要同步满足“设置归位、帮助提示、渐进展开、全中文化”的当前优先规则。
- 动态路由实施结论必须回写主文档和 worklog，不能只停留在子域计划文档里。
- 动态路由是 AI 自动化主线中的例外专线，不走普通 AI Profile 覆盖逻辑；设置页必须单独提供本地机制与 AI 学习的手动开关。

---

## 1. 当前目标交付物

当前实施主线补齐以下五项能力：

1. 动态路由模式、健康开关与预算配置项落库、校验、前后端接线。
2. relay 在 failover / race 路径中真实消费动态路由模式、健康开关与 `race_*_budget`。
3. `hybrid` / `shadow-ai` / `metrics-only` / `strict-mechanism` / `incident-safe` 形成可解释的运行时差异。
4. 每日 `dynamic summary scan` 生成可解释摘要，但不做 runtime mutation。
5. 摘要结果通过管理 API 和设置页形成真实可见闭环。
6. 本地 AI 学习状态已按 `(channel, key, model)` 学习并参与允许模式下的运行时推荐。

---

## 2. 模块拆分

### 2.1 设置与迁移模块

职责：

- 为 `dynamic_routing_mode`、`dynamic_routing_health_enabled` 和五个 `race_*_budget` 提供默认值。
- 在后端设置校验中保证这些配置为合法布尔值或非负整数。
- 在前端设置常量与设置页中暴露对应项。

### 2.2 relay 运行时调节模块

职责：

- 在健康开关开启时，根据 route-target policy 调整 `RaceAfterFails` 与 `RaceConcurrency`。
- 在健康开关关闭时，回退到 group 原始默认值。
- 保持调节行为仅限于当前请求运行时，不写回配置。

### 2.3 模式决策与审计模块

职责：

- 解析 `dynamic_routing_mode`。
- 生成推荐候选顺序与置信度。
- 决定当前请求采用推荐、影子记录、仅指标、纯机制或事故安全路径。
- 把模式、生效模式、决策类型、回退原因与推荐顺序写入 relay log 审计字段。

### 2.4 竞速预算模块

职责：

- 为全局、group、channel、key、probe 维度分别加预算约束。
- 在预算耗尽时拒绝继续竞速候选。
- 与现有 failover / race fallback 路径直接集成。

### 2.5 每日摘要扫描模块

职责：

- 每日运行一次 `dynamic summary scan`。
- 汇总当前模式、生效模式与决策基调。
- 汇总 channels、groups、failover groups 和 key source type 分布。
- 输出 `status`、`message`、`basis`、时间戳等摘要字段。
- 明确保证 `basis` 为 `daily_summary_scan_no_runtime_mutation`。

### 2.6 摘要 API 与设置页展示模块

职责：

- 通过管理 API 暴露动态路由摘要。
- 在设置页展示模式、健康开关、预算输入和摘要面板。
- 对 `skipped`、`ok`、`error` 等状态做可读化解释。

### 2.7 本地 AI 学习模块

职责：

- 新增 `dynamic_routing_learning_enabled` 设置项。
- 新增 `DynamicRouteLearningState`，按 `(channel_id, channel_key_id, model_name)` 保存学习状态。
- relay attempt 完成后记录成功、失败、状态码、延迟、fallback 与 race winner。
- 计算 `score` 与 `confidence`，并在 `hybrid` 评分中作为辅助信号。
- 在 `shadow-ai` 与 `metrics-only` 中只记录和审计，不改变真实顺序。
- 在 `strict-mechanism` 中保持确定性机制路径。
- 关闭学习开关后保留数据，但不参与评分。
- 明确禁止写回 `group_items`、priority、渠道配置或用户显式顺序。

---

## 3. 当前请求与任务流

### 3.1 relay 请求路径

1. 请求进入 relay。
2. 读取 `dynamic_routing_mode` 与 `dynamic_routing_health_enabled`。
3. 构建推荐候选顺序与置信度。
4. 若当前模式为 `hybrid` 且信号可信，则采用推荐顺序；否则按模式要求走影子记录、仅指标、纯机制或事故安全路径。
5. 在健康开关开启时，根据 route-target policy 计算运行时 `RaceAfterFails` / `RaceConcurrency`。
6. 进入 failover / race 路径前申请 `race_*_budget`。
7. 若预算允许则执行竞速探测；若预算不足则跳过该候选。
8. 将模式决策写入 relay log 审计字段。
9. 如果 `dynamic_routing_learning_enabled=true`，在尝试完成后更新本地学习状态。
10. 下一次评分时，`hybrid` 可读取学习状态辅助推荐排序，但不得写回用户配置。

### 3.2 后台摘要路径

1. 定时任务每天触发一次 `dynamic summary scan`。
2. 读取当前模式与健康开关；关闭时仍要写入可解释摘要。
3. 健康开关开启时读取 channels 与 groups。
4. 生成摘要并缓存在内存中。
5. 管理 API 读取该摘要。
6. 设置页消费该摘要并展示给用户。

---

## 4. 当前配置项清单

当前实施范围只包含以下配置项：

- `dynamic_routing_mode`
- `dynamic_routing_health_enabled`
- `race_global_budget`
- `race_group_budget`
- `race_channel_budget`
- `race_key_budget`
- `race_probe_budget`

当前已接线项同时包括：

- `dynamic_routing_learning_enabled`

仍然不新增外接模型服务地址作为动态路由依赖；动态路由学习必须是本地可解释学习。

---

## 5. 当前阶段划分

### Phase A

补齐设置默认值、迁移与校验，确保动态路由模式、动态健康开关和预算配置真实存在。

### Phase B

把模式层、健康开关和预算接入 relay 的 failover / race 运行时路径。

### Phase C

实现 relay 审计与每日 `dynamic summary scan`，只做摘要产出，不做 runtime mutation。

### Phase D

通过管理 API 与设置页把摘要和配置接出来，形成用户可见闭环。

### Phase E

本地 AI 学习状态和 `dynamic_routing_learning_enabled` 开关已经接线，学习分数已接入 `hybrid` 推荐评分，并确保不覆盖用户配置。

---

## 6. 后续 backlog 与已提升目标

以下内容仍然只是 backlog，不作为当前动态路由实施承诺：

- 外接独立大模型服务
- 自动持久化阈值调优
- 普通 AI Profile 覆盖动态路由配置

以下内容已提升为当前正式目标：

- 本地 AI 学习闭环。
- `dynamic_routing_learning_enabled` 设置开关。
- `(channel, key, model)` 粒度学习状态。
- 学习分数参与 `hybrid` 推荐评分。
- 学习推荐不写回 `group_items`，不覆盖 priority。

如果后续重新推进这些方向，必须基于新的代码现实重新开计划，而不是直接恢复旧文案。

---

## 7. 当前最低发布门槛

当前阶段发布前，至少必须满足：

1. `dynamic_routing_mode`、`dynamic_routing_health_enabled` 与全部 `race_*_budget` 已完成默认值、迁移、设置校验与前端常量接线。
2. relay 运行时真实消费这些配置，并在预算耗尽时产生可解释的跳过行为。
3. 五种动态路由模式都已形成实际可验证的运行时差异。
4. `dynamic summary scan` 已注册为每日任务，且明确只生成摘要，不持久化改写阈值或用户路由顺序。
5. 管理 API 能读取动态路由摘要。
6. 设置页能展示模式、健康开关、预算设置和摘要结果。
7. 设置页当前已能控制 `dynamic_routing_learning_enabled`。
8. 学习状态当前已可查询和清空，并且关闭开关后不参与排序。
9. 已有测试证明学习推荐不写回 `group_items`，不覆盖 priority。

