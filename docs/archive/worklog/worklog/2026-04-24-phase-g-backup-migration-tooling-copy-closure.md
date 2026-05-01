# 2026-04-24 Phase G Backup Migration Tooling Copy Closure

## 1. 任务信息

- 任务名称：Phase G 备份页剩余迁移能力中文主文案收口
- 日期：`2026-04-24`
- 当前阶段：`Phase G` 截图优先 UI 主线
- 对应 milestone：备份页中文主文案 no-browser 收口补充闭环

## 2. 开工前输入

- Master plan aligned before coding (yes/no): `yes`
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 `9`、`14` 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 `1.0`、`1.2`、`1.3` 节
- 上一个相关 worklog：`docs/worklog/2026-04-23-phase-g-backup-provider-copy-closure.md`
- 本次任务目标：关闭备份页底部“高级迁移能力待补”区域在简体/繁体中文主显示里残留的 `replace/map`、`remap` 英文技术词，并补齐 no-browser 回归断言。
- 本次已盘点本地资源：
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md`
  - `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
  - `docs/worklog/2026-04-23-phase-g-backup-provider-copy-closure.md`
  - `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档、现有 `verify-backup-component.cjs`、`verify-backup-logic.mjs`、`Backup.tsx`、`backup-logic.ts`
- 若未使用部分本地资源或上下文，原因：未额外读取浏览器 smoke 产物，因为本轮聚焦同主线 no-browser 文案闭环
- 本次是否启用子 agent 与分工边界：`否`
- 本次子 agent 使用模型：`未使用`
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：`否，主线程串行执行`
- 若使用分目录 agent，负责目录与禁止越界范围：`不适用`
- 若未使用子 agent，原因：任务范围很小，且与当前线程上下文和既有脚本强耦合，主线程直接闭环风险更低

## 3. 本次硬规则

- 只处理备份页中文主显示里的英文技术词残留，不扩散到其他设置模块。
- 保留隐藏 English test hooks，不破坏现有 no-browser 脚本锚点。
- 本轮必须形成“代码变更 + no-browser 验证 + worklog + memory”闭环。

## 4. 本次禁止事项

- 不修改备份导入导出后端语义。
- 不触碰与本轮无关的动态路由、首页或模型页逻辑。
- 不把隐藏测试文案直接删掉，避免破坏既有验证链。

## 5. 本次验收条件

- 简体/繁体中文主显示不再直接出现 `replace/map`、`remap`。
- `node scripts/verify-backup-component.cjs` 通过。
- `node scripts/verify-backup-logic.mjs` 通过。
- `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json` 通过。

## 6. 本次回滚点

- `web/src/components/modules/setting/backup-logic.ts` 中新增的中文主文案标准化逻辑
- `scripts/verify-backup-component.cjs`
- `scripts/verify-backup-logic.mjs`

## 7. 实现范围

- 先改数据语义还是先改 UI：先收口前端展示语义，再补脚本
- 受影响后端模块：无
- 受影响前端模块：`web/src/components/modules/setting/backup-logic.ts`
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：仅影响中文主显示文案与 no-browser 校验，不改导入导出行为

## 8. 实施步骤

1. 复查备份页组件与 `backup-logic.ts`，确认英文残留来自“剩余迁移能力”文案而不是组件层可见标题。
2. 在 `backup-logic.ts` 中新增中文主文案标准化处理，仅对 `zh-Hans` / `zh-Hant` 生效，把 `replace/map`、`remap` 替换为当前中文语义。
3. 在 `verify-backup-logic.mjs` 中新增简体/繁体中文断言，并在 `verify-backup-component.cjs` 中从真实渲染可见文本层补上对应 no-browser 守护。

## 9. 测试与验证

- 构建命令：`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-backup-component.cjs`
  - `node scripts/verify-backup-logic.mjs`
- 专项验证：确认简体/繁体中文可见文案不再包含 `replace/map`、`remap`

## 10. 风险与兼容性

- 新风险：低；仅新增文案标准化函数，作用域限定在备份页“剩余迁移能力”条目
- 兼容性风险：低；英文与日文文案未改，隐藏 English test hooks 保持不变
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：`是`
- 测试是否通过：`是`
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、当前状态文档、详细工作流、用户总账、前端主线状态、上一轮 backup worklog、automation memory、现有 backup 相关脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - canonical / workflow / 总账：确认本轮仍属 Phase G screenshot-first 同主线 no-browser 收口，且中文主显示优先级高于泛化优化
  - 前端主线状态 / 上一轮 worklog：确认备份页已有 provider-copy 收口，但“高级迁移能力待补”仍未被明确守护
  - 现有脚本与组件：确认英文残留来自 `backup-logic.ts`，可通过小范围改动闭环
  - automation memory：确认下一轮仍应优先沿同主线继续，不应横跳
- 本次使用了哪些子 agent 及其结论：`未使用`
- 子 agent 分工、负责范围与产出摘要：`不适用`
- 手工 smoke 状态：未执行浏览器级 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮为 no-browser 文案闭环，浏览器级证据不是完成前置条件
- 待验证页面清单：后续若继续备份页截图级验收，可补浏览器级 settings/backup 真实证据
- 若未使用子 agent，原因：任务小且上下文耦合，主线程串行更稳
- worklog 是否更新：`是`
- 遗留项：备份页浏览器级截图证据与设置/创建弹窗同池浏览器验证仍待继续
- 下一任务前置条件是否满足：`是`
