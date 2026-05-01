# 2026-04-22 Phase G CC Switch 繁中文案收口

## 1. 任务信息

- 任务名称：CC Switch 繁中文案残留英文收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题池优先返工窗口
- 对应 milestone：截图问题池 / CC Switch 中文化与流程型交互收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6 节、第 14 节、第 16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0 节、第 1.2 节、第 1.3 节、第 1.4 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-ccswitch-progressive-help-closure.md`
- 本次任务目标：
  - 修正繁中 `CC Switch` 文案里的无必要英文残留
  - 把静态验证脚本同步收紧，防止后续回退
  - 保持改动只落在同一 UI 主线的小闭环内
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-22-phase-g-ccswitch-progressive-help-closure.md`
  - automation memory `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
  - `web/src/components/modules/navbar/DocModal.tsx`
  - `web/public/locale/zh-Hans.json`
  - `web/public/locale/zh-Hant.json`
  - `scripts/verify-ccswitch-flow.mjs`
- 本次使用的本地 resources / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、上一轮 CC Switch worklog、automation memory、当前 locale 与验证脚本
- 若未使用部分本地资源或上下文，原因：本轮任务已能由现有文档、源码和脚本闭环，不需要额外外部资料
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：未使用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：未使用
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮属于单模块小闭环修复

## 3. 本次硬规则

- 只处理 `CC Switch` 中文化残留，不扩散到其他截图主题
- 必须保持和需求 09、需求 12、需求 19、需求 20、需求 27、需求 42 一致
- 必须有真实代码增量和直接可重复的静态验证

## 4. 本次禁止事项

- 不改 deep-link 语义和后端契约
- 不把浏览器 smoke 未跑结果写成已完成
- 不顺手扩散到其他设置页或渠道页文案

## 5. 本次验收条件

- 繁中 `CC Switch` 不再保留无必要英文示例文案
- 繁中 `CLI 工具` 标签改为中文一致表达
- 静态验证脚本能锁住这两个回归点
- `CC Switch` 静态验证与前端 typecheck 通过

## 6. 本次回滚点

- `web/public/locale/zh-Hant.json`
- `scripts/verify-ccswitch-flow.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 locale 文案，再补验证脚本
- 受影响后端模块：无
- 受影响前端模块：`DocModal` 的 `CC Switch` 文案资源
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：只影响繁中界面显示与验证约束，不影响导入行为

## 8. 实施步骤

1. 复核 `CC Switch` 现有流程实现、上一轮 worklog 与总账要求，确认主流程已存在，剩余缺口集中在繁中文案残留。
2. 修改 `zh-Hant` locale，把 `CLI 工具` 改成 `客戶端工具`，把名称示例从 `例：My Claude` 改成 `例：我的 Claude 設定`。
3. 更新 `scripts/verify-ccswitch-flow.mjs`，增加针对繁中示例与标签的负向断言，避免后续回退。

## 9. 测试与验证

- 构建命令：`D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`D:\gol1\node.exe scripts\verify-ccswitch-flow.mjs`
- 专项验证：`rg -n '"tabCCSwitch"|"ccswitchCliTool"|"ccswitchNamePlaceholder"|"ccswitchImport"' web/public/locale/zh-Hant.json scripts/verify-ccswitch-flow.mjs`

## 10. 风险与兼容性

- 新风险：低；本轮仅调整繁中文案和静态断言
- 兼容性风险：低；不改运行时逻辑
- 是否阻塞下一任务：不阻塞

## 11. 收工记录

- 构建是否通过：通过；前端 typecheck 已通过
- 测试是否通过：通过；`verify-ccswitch-flow.mjs` 已通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：主规划、用户上下文总账、详细工作流、前端主线状态、上一轮 CC Switch worklog、automation memory、当前 locale 与脚本源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - 主规划与总账要求 `CC Switch` 继续保持流程型交互，并清理中文界面无必要英文
  - 上一轮 worklog 说明主流程和帮助提示已完成，当前更适合做中文化残留收口
  - 当前脚本已覆盖主要结构，但未锁定繁中残留英文示例和标签
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：未使用
- 手工 smoke 状态：未执行
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先做静态中文化收口，浏览器级验证留给同主线下一轮
- 待验证页面清单：
  - `DocModal` 的 `CC Switch` 页签桌面视图
  - `DocModal` 的 `CC Switch` 页签 375px 视图
  - `Claude` 进阶模型映射折叠区浏览器级显隐
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮为集中小闭环
- worklog 是否更新：是
- 遗留项：
  - `CC Switch` 浏览器级 smoke 仍未完成
  - 繁中界面其他截图区域是否还有零散英文，仍需同主线继续清理
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

1. 继续留在 Phase G 同主线，优先补 `CC Switch` 页签的浏览器级 smoke 或手工最小回归。
2. 若浏览器 smoke 仍受宿主环境阻塞，则切到同主线相邻的小任务，继续清理截图问题池中的繁中/中文化残留。
3. 完成浏览器级证据后，再把 `CC Switch` 收口状态同步回前端主线总览。
