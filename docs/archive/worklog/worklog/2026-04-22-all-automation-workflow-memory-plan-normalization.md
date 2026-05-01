# 所有自动化的 workflow / memory / plan 统一记录

> 日期：2026-04-22
>
> 目的：把最新用户上下文、图片优先问题池、`Codex` 执行口径、`Master plan aligned before coding (yes/no):` 前置要求和当前 next plan，同步应用到所有 automation memory 与共享 workflow 文档。

---

## 1. 任务信息

- 任务名称：所有自动化的 workflow / memory / plan 统一
- 日期：2026-04-22
- 当前阶段：跨自动化执行口径统一窗口
- 对应 milestone：自动化链路继承规则收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：第 1 节、第 9 节、第 12 节、第 14 节、第 15 节
- 对应 workflow 章节：第 1 节总原则、第 1.2 节开工前固定动作、第 1.3 节收工前固定动作、第 11 节当前建议施工顺序
- 上一个相关 worklog：[2026-04-22-user-context-mainline-priority-normalization.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-user-context-mainline-priority-normalization.md)
- 本次任务目标：
  - 把共享 workflow 明确成所有自动化都要继承的规则
  - 把当前 next plan 更新为图片问题池优先的统一顺序
  - 把三条 automation memory 同步到同一套入口和下一轮候选顺序
- 本次已盘点本地资源：
  - [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md)
  - [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md)
  - [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md)
  - [ENV_READY_AND_NEXT_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md)
  - [CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md)
  - [FRONTEND_UI_MAINLINE_STATUS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md)
  - [2026-04-22-user-context-mainline-priority-normalization.md](/D:/GPT-codex/octopus_repo/docs/worklog/2026-04-22-user-context-mainline-priority-normalization.md)
  - `C:\Users\李昊桐\.codex\automations\octopus\memory.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、共享 workflow、当前 next plan、当前状态文档、前端主线文档、三个 automation memory
- 若未使用部分本地资源或上下文，原因：本轮目标是先完成跨自动化规则统一，未进入具体代码返工
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否，由主线程直接收口共享规则
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：本轮修改对象高度耦合且均为共享规则载体，主线程直接统一更稳妥

## 3. 本次硬规则

- 所有自动化都必须先读同一套核心文档，再进入编码或验证。
- 所有自动化都必须使用 `Codex` 口径，不再混用 `OpenCode`。
- 当前优先级以图片问题池优先，不允许 memory 继续停留在旧的单线 Phase 叙事。
- 每轮结束后必须给下一轮留下可直接接手的 memory / worklog 入口。

## 4. 本次禁止事项

- 不得只更新 `octopus-2` 而遗漏其它 automation。
- 不得只改 workflow 或只改 memory，造成共享规则继续分裂。
- 不得把这次统一动作写成抽象口号而不落实到具体文件。

## 5. 本次验收条件

- [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 出现跨自动化统一继承规则。
- [ENV_READY_AND_NEXT_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md) 出现当前统一 next plan 和候选任务顺序。
- 三个 automation memory 都写入新的统一入口和下一轮任务顺序。
- 本次 worklog 明确记录这次 cross-automation 统一动作。

## 6. 本次回滚点

- 本轮仅修改文档和 automation memory，可按文件级回退。

## 7. 实现范围

- 先改数据语义还是先改 UI：先改执行语义与计划语义
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：影响后续自动化执行顺序与文档入口，不影响运行时代码

## 8. 实施步骤

1. 重新读取共享 workflow、用户上下文总账、当前 next plan、当前状态文档与三个 automation memory。
2. 更新 workflow 和 next plan，明确所有自动化都要继承的执行前置和当前优先级。
3. 更新三个 automation memory，并新增本 worklog 作为正式记录。

## 9. 测试与验证

- 构建命令：未运行，本轮只修改文档和 memory
- 测试命令：未运行，本轮只修改文档和 memory
- 专项验证：读回共享 workflow、当前 next plan 与三个 automation memory，确认统一规则与优先级一致

## 10. 风险与兼容性

- 新风险：若后续新增文档未同步沿用这套规则，仍可能再度分裂
- 兼容性风险：低，本轮不改代码与接口
- 是否阻塞下一任务：不阻塞，反而为下一轮图片问题修复提供统一入口

## 11. 收工记录

- 构建是否通过：未运行
- 测试是否通过：未运行
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、共享 workflow、当前 next plan、当前状态文档、前端主线文档、三个 automation memory
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账明确了图片问题池与跨自动化继承要求
  - 共享 workflow 需要新增所有自动化统一继承规则
  - 当前 next plan 仍偏旧阶段顺序，需要切换为图片问题池优先
  - 三个 automation memory 口径不一致，需要统一下一轮入口
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：未使用
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮无需运行 UI 或后端 smoke
- 待验证页面清单：下一轮从图片问题池中选出的首个 UI 问题页面
- 若未使用子 agent，原因：共享规则统一任务强耦合，主线程直接修改更稳妥
- worklog 是否更新：是
- 遗留项：
  - 下一轮需要真正进入图片问题池中的一个具体 UI 闭环并做验证
  - 备份导入导出兼容主线仍保留为相邻候选任务
  - 若后续新增 automation，需要沿用本次统一规则追加到 memory
- 下一任务前置条件是否满足：满足，下一轮可直接从图片问题池首个闭环开始
