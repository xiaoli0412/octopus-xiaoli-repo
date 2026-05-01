# 2026-04-24 Phase G Channel Create Browser Wrapper And Smoke Narrowing

## 1. 任务信息

- 任务名称：channel create 浏览器 smoke 宿主层 wrapper 接入与阻塞缩窄
- 日期：2026-04-24
- 当前阶段：Phase G
- 对应 milestone：截图优先返工池 / channel create browser evidence

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`9.1`、`9.7`、`14`
- 对应 workflow 章节：`1.0`、`1.2`、`1.3`、`11.2`、`11.4`
- 上一个相关 worklog：`docs/worklog/2026-04-24-phase-g-channel-create-collapsed-summary-closure.md`
- 本次任务目标：沿同一 Phase G screenshot-first 池，优先尝试解开 `channel create` 浏览器级 smoke 阻塞；若宿主层仍阻塞，则至少补齐稳定入口、收紧脚本脆弱断言，并把真实 blocker 缩窄到更具体的层级
- 本次已盘点本地资源：canonical plan、用户上下文总账、详细工作流、前端主线状态、上一轮 channel create worklog、automation memory、`scripts/verify-setting-help-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke.mjs`、`scripts/verify-channel-create-flow.mjs`、`web/src/components/modules/channel/Form.tsx`
- 本次使用的本地 resources / skills / 记忆上下文：当前线程上下文、现有 screenshot-first worklog、已有 settings help wrapper 路径、channel create 现有 selector contract
- 若未使用部分本地资源或上下文，原因：本轮聚焦 browser smoke 入口与验证链，不展开到 group create 或其它页面
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：由 Codex 主线程串行推进
- 若使用分目录 agent，负责目录与禁止越界范围：无
- 若未使用子 agent，原因：当前关键路径集中在同一组脚本与 channel create 页面，主线程直接收口更快

## 3. 本次硬规则

- 保持任务停留在 `Phase G screenshot-first` 同一主线，不切到无关功能。
- 优先解决浏览器 smoke 入口与验证链路本身的阻塞，不去扩大前端交互范围。
- 若浏览器级验证仍被宿主策略阻塞，必须把 blocker 精确记录到可执行层，而不是停留在“浏览器不行”的模糊描述。

## 4. 本次禁止事项

- 不把未实际跑通的浏览器 smoke 记成完成。
- 不为了通过 smoke 回退现有 `channel create` 结构。
- 不扩散到 group create 或 settings help 之外的新页面。

## 5. 本次验收条件

- `channel create` 有独立的 PowerShell wrapper 入口，可从宿主层启动并收集 stdout/stderr。
- `scripts/verify-channel-create-browser-smoke.mjs` 去掉依赖乱码化中文文本的脆弱断言，改为稳定结构断言。
- 相关 no-browser 验证与类型检查通过。
- 浏览器链路若仍失败，失败点已缩窄到明确层级并写回记录。

## 6. 本次回滚点

- `scripts/verify-channel-create-browser-smoke.mjs`
- `scripts/verify-channel-create-browser-smoke.ps1`
- `scripts/verify-channel-create-flow.mjs`
- `web/public/locale/ja.json`

## 7. 实现范围

- 先改数据语义还是先改 UI：先改 smoke 脚本与验证契约，再试浏览器路径
- 受影响后端模块：无业务代码改动
- 受影响前端模块：无页面结构改动，本轮只校准 `channel create` 的验证脚本与 locale 类型契约
- 受影响接口：浏览器 smoke 的本地服务启动与页面验证入口
- 是否影响旧数据：否
- 是否影响旧行为：否，主要影响自动化验证路径

## 8. 实施步骤

1. 复核主规划、前端主线状态、上一轮 channel create worklog 与 automation memory，确认本轮继续留在 `Phase G screenshot-first` / `channel create browser evidence` 同池。
2. 对照 `settings help` 已有 PowerShell wrapper，新增 `scripts/verify-channel-create-browser-smoke.ps1`，让 `channel create` 浏览器 smoke 也能从宿主层启动后端并调用 Node smoke。
3. 收紧 `scripts/verify-channel-create-browser-smoke.mjs`，把依赖乱码化中文文本的断言改为结构与存在性断言。
4. 同步修正 `scripts/verify-channel-create-flow.mjs` 中与当前实现不一致的过期 no-browser 护栏，并补齐 `ja` locale 的 `modelSearchPlaceholder`，恢复 `tsc` 基线。
5. 运行 no-browser 验证、类型检查和新的 PowerShell wrapper，记录真实通过项与最终阻塞层级。

## 9. 测试与验证

- 构建命令：
  - `node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试命令：
  - `node scripts/verify-channel-create-flow.mjs`
- 专项验证：
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode self-start -NodeSmokeTimeoutSeconds 180 -KeepArtifacts`

## 10. 风险与兼容性

- 新风险：低；本轮主要是 browser smoke 入口与验证脚本收口
- 兼容性风险：低；仅影响自动化验证链，不改业务行为
- 是否阻塞下一任务：部分阻塞；`channel create` 浏览器证据仍未关闭，但阻塞点已经缩窄

## 11. 收工记录

- 构建是否通过：通过；`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`
- 测试是否通过：通过；`node scripts/verify-channel-create-flow.mjs`
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、用户上下文总账、详细工作流、前端主线状态、上一轮 channel create worklog、automation memory、settings help wrapper、channel create 脚本与组件源码
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - settings help wrapper 证明“宿主层起 Node smoke”是本仓库已验证过的路径
  - channel create 上一轮 worklog 证明 selector contract 已经落地，本轮无需再改页面结构
  - 现有 Form 与 locale 源码证明部分 no-browser 断言已落后于当前实现，需要先回收护栏再测浏览器路径
- 本次使用了哪些子 agent 及其结论：未使用
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未做手工点击；已通过 PowerShell wrapper 自动尝试 browser smoke
- 手工 smoke 阻塞原因 / 缺少的环境：Node smoke 进入浏览器步骤后，仍在 `spawn @playwright/cli` 时报 `EPERM`；当前阻塞已从“脚本入口不稳定”缩窄为“Node 进程内部再起 Playwright CLI 子进程被宿主策略拒绝”
- 待验证页面清单：`channel create` dialog 浏览器级 `375px / hover / focus` 证据、随后的 `group create` dialog 浏览器证据
- 若未使用子 agent，原因：当前关键路径集中在同一脚本族和当前页面，主线程串行推进更快
- worklog 是否更新：是
- 遗留项：
  - `channel create` 浏览器 smoke 仍未拿到成功证据
  - PowerShell wrapper 已可用，但 Node 内部 `playwright-cli` 子进程仍被宿主 `EPERM` 拦截
- 下一任务前置条件是否满足：满足；下一轮可直接从“替换 Node 内 `spawn @playwright/cli` 的执行方式或改用另一条宿主允许的浏览器驱动路径”继续，而无需重新摸索当前入口与日志链
