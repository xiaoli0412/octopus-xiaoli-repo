# 用户上下文主线优先级标准化记录

> 日期：2026-04-22
>
> 目的：记录本轮把用户多轮上下文、图片补充要求、帮助提示、熔断自定义、`Codex` 执行口径和分目录 agent 可执行边界统一并入主文档体系的落盘动作。

---

## 1. 任务信息

- 任务名称：用户上下文主线优先级与执行口径标准化
- 日期：2026-04-22
- 当前阶段：Phase G 前置文档收口 / Phase F 与 Phase G 交界整理
- 对应 milestone：文档主线与当前优先返工清单收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：第 1 节、第 9 节、第 12 节、第 14 节、第 15 节
- 对应 workflow 章节：第 1 节总原则、第 1.2 节开工前固定动作、第 9 节与当前 UI / 备份相关阶段说明
- 上一个相关 worklog：无直接单一上游，主要继承 2026-04-21 与 2026-04-22 的 Phase F 文档收口记录
- 本次任务目标：
  - 把用户多轮要求统一并入一份需求总账
  - 把帮助提示和熔断增强并入主线优先级
  - 把图片中点名的文档对齐要求写成固定动作
  - 把执行主体统一改成 `Codex`
  - 把 `Master plan aligned before coding (yes/no):` 固定写入模板和规则
- 本次已盘点本地资源：
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`
  - `docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md`
  - `docs/worklog/README.zh-CN.md`
  - `docs/worklog/WORKLOG_TEMPLATE.zh-CN.md`
  - 当前线程上下文与压缩摘要
- 本次使用的本地 resources / skills / 记忆上下文：
  - 使用当前线程压缩摘要恢复上一轮已完成的文档落盘背景与剩余缺口
  - 使用本地文档交叉确认主规划、工作流、状态文档、动态路由文档与 worklog 模板的当前口径
- 若未使用部分本地资源或上下文，原因：未动用外部资料；本轮目标是先收口本地文档体系
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否，先由主线程完成文档口径统一
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：本轮任务高度耦合、主要是同一套文档口径统一，适合主线程一次性收口

## 3. 本次硬规则

- 先改文档，再改代码。
- 不遗漏用户已明确表达过的优先要求。
- 文档中的施工主体统一按 `Codex` 口径描述。
- 所有新增要求都要可被当前自动化链路和分目录 agent 执行。

## 4. 本次禁止事项

- 不得只改单一文档而让主线口径继续分裂。
- 不得继续保留错误的 `OpenCode` 施工主体描述。
- 不得把用户已明确补充的要求继续留在聊天里不落盘。

## 5. 本次验收条件

- 用户上下文总账补充到 40+ 条要求，并涵盖最新补充项。
- `worklog` 目录说明和模板已加入 `Master plan aligned before coding (yes/no):`。
- 详细工作流已加入 `Codex` 口径和分目录 agent 执行边界。
- canonical plan 已去掉错误的 `OpenCode` 施工描述，并补入当前用户上下文优先要求。

## 6. 本次回滚点

- 本轮仅修改文档文件，可按文件级回退。

## 7. 实现范围

- 先改数据语义还是先改 UI：先改文档语义
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：影响后续施工规则与验收口径，不影响运行时功能

## 8. 实施步骤

1. 读取并核对主文档、主工作流、状态文档、动态路由文档和 worklog 模板。
2. 把帮助提示、熔断增强、图片点名文档一致性、`Codex` 施工口径、分目录 agent 可执行边界并入总账和主流程。
3. 更新 worklog 目录说明和模板，固定“主规划对齐后再编码”的前置检查。

## 9. 测试与验证

- 构建命令：未运行，本轮仅改文档
- 测试命令：未运行，本轮仅改文档
- 专项验证：人工复核文档中的关键规则、编号和链接路径

## 10. 风险与兼容性

- 新风险：若后续其它文档仍残留旧口径，可能需要继续清理
- 兼容性风险：低；本轮不改运行时代码
- 是否阻塞下一任务：不阻塞，反而为后续代码返工提供了统一口径

## 11. 收工记录

- 构建是否通过：未运行
- 测试是否通过：未运行
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、工作流、状态文档、动态路由文档、worklog 模板、当前线程压缩摘要
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账已有雏形，但还缺 `Codex` 口径、图片点名文档一致性和模板前置字段
  - 主规划仍残留 `OpenCode` 描述，需要纠正
  - worklog 模板缺少用户明确点名的 `Master plan aligned before coding (yes/no):`
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：未使用
- 手工 smoke 状态：未执行；本轮只收口文档
- 手工 smoke 阻塞原因 / 缺少的环境：无，本轮无需运行
- 待验证页面清单：后续代码返工阶段再对应设置页、渠道页、分组页、首页、`CC Switch`、备份页执行
- 若未使用子 agent，原因：文档统一口径任务强耦合，主线程直接改动效率更高
- worklog 是否更新：是
- 遗留项：
  - 已完成：`CURRENT_STATUS_AND_PLAN.zh-CN.md`、`ENV_READY_AND_NEXT_PLAN.zh-CN.md`、`FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`、`DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md`、`DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md` 的头部主线依赖补齐与 `Codex` 口径对齐
  - 已完成：`USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md` 的截图问题区已升级为“现象、根因、风险、修复动作、自动化验收点”细化结构，便于自动化执行
  - 已完成：`DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 新增截图问题执行模板与硬性验收规则
  - 待后续继续：进入这些文档正文级别的细化条目同步（不仅是头部规则）
  - 文档收口完成后，下一步应进入图片问题和备份导入导出问题的代码修复
- 下一任务前置条件是否满足：满足，后续可以在统一口径下继续推进代码主线
