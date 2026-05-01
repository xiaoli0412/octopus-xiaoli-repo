# Worklog 使用说明

> 本目录是 Octopus 项目的施工记录目录。
>
> 它不是普通笔记目录，而是 [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md) 与 [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 的执行落地点。

执行规则：

- 本目录所有 worklog 都必须先服从 `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`。
- 如果 worklog 与主规划冲突，以主规划为准。
- 如果实现必须偏离主规划，先更新主规划，再更新 worklog，再改代码，最后补验证记录。
- 当前执行环境统一按 `Codex` 口径记录，不再使用 `OpenCode` 口径描述当前施工流程。
- worklog 中写出的步骤必须能被当前 `Codex` 主线程、自动化链路以及按目录切分的子 agent / 分目录 agent 直接执行；若某一步只适用于特定工具或特定环境，必须写清限制条件与替代路径。

---

## 1. 目录用途

本目录用于记录：

- 每个阶段正式开工前的任务输入
- 每次改动对应的硬规则、禁止事项、验收标准
- 每次收工后的构建、测试、兼容性、回滚点
- 当前阶段的遗留项与阻塞项
- 本轮使用了哪些本地资源、skills、记忆上下文与子 agent，以及它们分别产出了什么结论

没有 worklog 的任务，不算正式进入施工状态。

---

## 2. 命名规则

文件命名统一使用：

- `YYYY-MM-DD-<phase-or-task>.md`

推荐命名示例：

- `2026-04-15-phase-a-stability-closure.md`
- `2026-04-15-phase-b-milestone-1-gap-closure.md`
- `2026-04-15-phase-c-observability-closure.md`
- `2026-04-15-backend-task-channel-routing-order.md`

规则：

- 阶段级文档优先使用 `phase-*`
- 子任务文档优先使用 `backend-task-*`、`frontend-task-*`、`import-task-*`
- 一个任务一个文件，不要把多个不相干主题塞进同一份 worklog

---

## 3. 正式施工前必须做什么

每次开始任何任务前，必须先完成：

1. 阅读 canonical plan 对应章节
2. 阅读 [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 中对应 Phase
3. 阅读当前阶段已有 worklog
4. 盘点本轮直接可复用的本地资源与上下文来源，至少包括：仓库内 MD、当前阶段 worklog、现有脚本、测试命令、当前线程上下文、本地 skills、可复用记忆结论
5. 将本轮使用的本地资源、skills 与记忆上下文写入 worklog；若暂未使用某类资源，也要说明原因
6. 判断本次任务是否适合拆成相互独立的子任务；若适合，应先明确子 agent 分工、范围与交付物；若未使用子 agent，也要在 worklog 中说明原因
7. 若当前环境允许并可正常调度，子 agent 默认统一使用 `gpt-5.4`；若无法使用，必须在 worklog 中写明阻塞原因与回退方案
8. 明确本次任务的 7 个输入：
   - 任务名
   - 对应 canonical 章节
   - 对应 milestone
   - 本次硬规则
   - 本次禁止事项
   - 本次验收条件
   - 本次回滚点
9. 在 worklog 开工区显式填写：`Master plan aligned before coding (yes/no):`

如果这 7 项以及 `Master plan aligned before coding (yes/no):` 没有写进 worklog，不允许开始正式编码。

---

## 4. 收工前必须补什么

每次任务结束前，worklog 必须补齐：

1. 本次影响的硬规则
2. 本次兼容性风险
3. 本次是否影响旧数据
4. 本次是否影响旧接口或旧 UI
5. 本次回滚点
6. 本次构建/测试结果
7. 本次使用了哪些本地资源 / skills / 记忆上下文，以及它们分别提供了什么结论
8. 本次是否使用了子 agent；若使用，分别负责了什么范围、产出了什么结论；若未使用，也要说明原因
9. 手工 smoke 状态、阻塞原因、缺少的环境，以及待验证页面清单
10. 本次遗留项
11. 是否满足进入下一任务的前置条件

补充记录要求：

- 正式施工前，必须先记录本轮实际使用的本地资源与上下文来源，不能只写泛化描述
- 若使用了子 agent，必须记录其分工、负责范围、交付物与最终采用结论
- 若使用了子 agent，必须记录使用模型；默认应为 `gpt-5.4`，若不是，必须说明原因
- 若手工 smoke 尚未完成，必须明确记录阻塞原因、缺少的环境和未验证页面，不能写成模糊遗留项
- 若未使用子 agent，也必须写明原因，例如任务过小、上下文强耦合、当前环境不可用等
- 若某项施工步骤计划交给 `Codex` 自动化或分目录 agent 执行，也必须在 worklog 中写清其负责目录、负责问题和不可越界范围

---

## 5. 当前阶段顺序

当前仓库施工顺序固定为：

1. `Phase A`：工程稳定性收口
2. `Phase B`：里程碑 1 可用性核心收口
3. `Phase C`：里程碑 2 可观测性收口
4. `Phase D`：里程碑 3 策略与性能收口
5. `Phase E`：里程碑 5 价格归一化收口
6. `Phase F`：里程碑 6 备份/导入/迁移适配收口
7. `Phase G`：UI、移动端、部署与最终验收

不允许跳过前序阶段直接进入后序阶段主施工。

---

## 6. 推荐文件

本目录至少应长期存在以下文件：

- [README.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/worklog/README.zh-CN.md)
- [WORKLOG_TEMPLATE.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/worklog/WORKLOG_TEMPLATE.zh-CN.md)
- 当前活跃阶段的 worklog
