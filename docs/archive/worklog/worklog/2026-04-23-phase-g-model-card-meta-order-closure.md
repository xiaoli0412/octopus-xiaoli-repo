# 2026-04-23 Phase G Model Card Meta Order Closure

## 1. 任务信息

- 任务名称：模型卡片元信息顺序收紧收口
- 日期：2026-04-23
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：模型/价格区域普通布局与紧凑布局双模式收口后的信息层级收紧

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9`、`14`、`16` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3`、`1.4`、`9` 节
- 上一个相关 worklog：`docs/worklog/2026-04-23-phase-g-model-layout-runtime-closure.md`
- 本次任务目标：
  - 在同一 `Phase G screenshot-first` 主线下继续收紧模型卡片普通布局的信息顺序
  - 把 `规范名称 / 计费模式 / 官方价格` 收到统一的中文 meta 信息带，减少上下跳读
  - 同步 `verify-llm-price-boundary.mjs`，给新结构加 no-browser 回归护栏
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-model-layout-runtime-closure.md`
  - `$CODEX_HOME/automations/octopus-2/memory.md`
  - `web/src/components/modules/model/Item.tsx`
  - `scripts/verify-llm-price-boundary.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、当前状态文档、上一轮模型 worklog、automation memory、模型模块源码与验证脚本
- 若未使用部分本地资源或上下文，原因：本轮只处理模型卡片信息层级与 no-browser 验证，不扩散到浏览器证据或其它页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮是同一模块内的紧耦合小闭环

## 3. 本次硬规则

- 只处理模型卡片信息顺序与验证脚本，不扩散到其它设置页或后端主题
- 保持现有 `普通 / 紧凑` 双布局入口不变，只收紧普通布局信息层级
- 必须留下可复跑的 no-browser 断言，不能只改视觉结构

## 4. 本次禁止事项

- 不把浏览器级 `375px` 证据误记为完成
- 不改模型接口、价格同步逻辑或探测语义
- 不顺手清理无关页面的文案或布局

## 5. 本次验收条件

- `ModelItem` 普通布局将 `规范名称 / 计费模式 / 官方价格` 收口到统一 meta 信息带
- 紧凑布局继续保留同一套 meta 信息，不回退现有布局切换语义
- `node scripts/verify-llm-price-boundary.mjs` 通过并覆盖新结构
- `node scripts/verify-locale-consistency.mjs` 通过
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过

## 6. 本次回滚点

- `web/src/components/modules/model/Item.tsx`
- `scripts/verify-llm-price-boundary.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- `docs/worklog/2026-04-23-phase-g-model-card-meta-order-closure.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改前端卡片信息层级，再补脚本断言
- 受影响后端模块：无
- 受影响前端模块：模型卡片组件、模型/价格 no-browser 验证脚本、状态文档
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅收紧信息顺序与展示层级，不改变模型编辑、删除、布局切换或数据语义

## 8. 实施步骤

1. 复核主规划、用户上下文总账、前端主线状态、上一轮模型 worklog 和 automation memory，确认当前最值得推进的是模型卡片信息顺序收紧。
2. 更新 `web/src/components/modules/model/Item.tsx`，把普通布局的 `规范名称 / 计费模式 / 官方价格` 统一收口到同一条 meta 信息带，并保持紧凑布局共用同一套元信息来源。
3. 更新 `scripts/verify-llm-price-boundary.mjs`，新增对 meta 信息带和顺序的 no-browser 断言。
4. 同步前端主线状态文档、当前状态文档与本轮 worklog。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-llm-price-boundary.mjs`
  - `node scripts/verify-locale-consistency.mjs`
- 专项验证：读回 `web/src/components/modules/model/Item.tsx` 与 `scripts/verify-llm-price-boundary.mjs`，确认 meta 信息带与顺序断言均已落地

## 10. 风险与兼容性

- 新风险：低；本轮只调整模型卡片信息层级与静态断言
- 兼容性风险：低；未改接口、未改数据结构、未改路由语义
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、当前状态文档、上一轮模型 worklog、automation memory、模型模块源码与验证脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 用户上下文总账与前端主线状态都确认模型/价格区仍在同一 screenshot-first 池，且当前优先级是收紧布局和中文主显示
  - 上一轮模型 worklog 已经收口了布局状态传递，因此本轮最合适的是继续收紧信息顺序，而不是重做布局
  - automation memory 说明浏览器级证据仍受宿主 CDP 阻塞，因此继续推进 no-browser 小闭环是当前最稳妥路径
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：宿主 `Edge/CDP` bootstrap 仍阻塞浏览器级 `375px / hover / focus` 证据，本轮继续以 no-browser 验证收口
- 待验证页面清单：模型页浏览器级 `普通 / 紧凑` 切换、模型页 `375px` 布局、设置页相关卡片 hover/focus 证据
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮为模型模块单点收口任务
- worklog 是否更新：是
- 遗留项：
  - 模型/价格区域浏览器级 `375px` 证据仍未补齐
  - 若后续发现新的中文界面英文主显示泄漏，仍需在同池继续做 no-browser 小闭环
  - 宿主 `Edge/CDP` 阻塞仍未解除
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

- 下一轮最适合继续推进：继续留在 `Phase G` 同一 screenshot-first 池，优先尝试模型页或设置页的浏览器级 `375px / hover / focus` 证据；若宿主浏览器链路仍阻塞，则转向 channel/group create dialog 的同池可验证小闭环。
- 同主线候选顺序：
  1. 模型页浏览器级 `普通 / 紧凑` 与 `375px` 证据
  2. 设置页 help-hint 浏览器级 `hover / focus / 375px` 证据
  3. channel create dialog 浏览器级 `375px` / help-hint 证据
  4. group create dialog 浏览器级 `375px` / help-hint 证据
