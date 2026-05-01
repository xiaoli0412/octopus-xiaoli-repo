# 2026-04-27 Phase H Node Env Validation Unblock

## 1. 任务信息

- 任务名称：phase h node env validation unblock
- 日期：2026-04-27
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H5/H6 consumer contract 验证环境收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / 验收标准 19-21
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-27-phase-h-ai-config-profile-summary-contract-closure.md`
- 本次任务目标：修复当前 Codex 自动化上下文中的 Node 执行阻塞，让前一轮 Phase H 前端 contract 改动重新可跑共享验证链
- 本次已盘点本地资源：AGENTS.md、canonical plan、CURRENT_STATUS_AND_PLAN、DETAILED_EXECUTION_WORKFLOW、最近 Phase H worklog、automation memory、`scripts/use-node-env.ps1`、`scripts/runtime-win.ps1`、`scripts/verify-backup-browser-smoke.ps1`、`scripts/vitest-no-spawn.cjs`、`web/vitest.config.ts`
- 本次使用的本地 resources / skills / 记忆上下文：上述文档与脚本，以及 session 开头读取的 `using-superpowers` / `brainstorming` skill 文件
- 若未使用部分本地资源或上下文，原因：本轮聚焦共享 Node 验证入口，不扩散到更多 AI Profile consumer 或浏览器 smoke
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：不适用
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若使用分目录 agent，负责目录与禁止越界范围：不适用
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H5/H6 主线，只修验证环境，不改 AI Profile 业务 contract
- 优先复用已有 Windows 环境回退模式，避免引入新的验证入口分叉
- 必须把“已恢复的验证项”和“仍阻塞的验证项”明确分开记录

## 4. 本次禁止事项

- 不扩散到 AI 自动化页、设置页或后端模型逻辑
- 不把临时命令当作最终方案，必须收敛到共享脚本入口
- 不创建子 agent

## 5. 本次验收条件

- `scripts/use-node-env.ps1` 能在当前自动化上下文中补齐 Node 运行所需的核心 Windows 环境变量
- 通过共享入口恢复 `tsc --noEmit`、locale consistency 与至少一条现有 no-browser 验证
- 明确记录 `vitest/esbuild spawn EPERM` 的剩余阻塞，不把它误记成业务代码失败

## 6. 本次回滚点

- `scripts/use-node-env.ps1`
- `docs/worklog/2026-04-27-phase-h-node-env-validation-unblock.md`
- `C:\Users\李昊桐\.codex\automations\octopus-2\memory.md`

## 7. 实现范围

- 先改数据语义还是先改 UI：都不改；先收口共享验证环境脚本
- 受影响后端模块：无
- 受影响前端模块：无业务代码改动
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：低风险；仅增强 Node 脚本会话环境兜底

## 8. 实施步骤

1. 复核 automation memory、最近 Phase H worklog 和共享脚本，确认当前主阻塞已从 AI 业务 contract 转为 Node 执行环境。
2. 复现当前会话中的 `node` 失败，确认 Codex 内置 `node.exe` 会因缺失 `SystemRoot` / `windir` / `APPDATA` / `LOCALAPPDATA` / `ProgramData` / `COMSPEC` 触发 `CSPRNG(nullptr, 0)`；同时确认直跑用户安装的 `node.exe` 仍是宿主级 `Access is denied`。
3. 复用 `runtime-win.ps1` 和 browser smoke 包装器里的已有经验，为 `scripts/use-node-env.ps1` 增加 `Ensure-CoreWindowsEnvironment` 回退逻辑。
4. 通过共享脚本重新运行 `tsc --noEmit`、`verify-locale-consistency.mjs` 和 `verify-setting-info-logic.mjs`，确认 Phase H 前端验证链恢复。
5. 重新尝试 `vitest`，确认剩余问题已经收敛为 `vite/esbuild` 在加载 `vitest.config.ts` 时的 `spawn EPERM`，而不再是 Node 启动失败。

## 9. 测试与验证

- 构建命令：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：`. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-locale-consistency.mjs`
- 专项验证：
  - `. .\scripts\use-node-env.ps1`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-setting-info-logic.mjs`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE -r .\scripts\vitest-no-spawn.cjs .\web\node_modules\vitest\vitest.mjs run .\web\src\components\modules\setting\AIAutomationSource.test.tsx .\web\src\components\modules\ai-automation\index.test.tsx --config .\web\vitest.config.ts`
  - `git diff --check -- scripts/use-node-env.ps1`

## 10. 风险与兼容性

- 新风险：`use-node-env.ps1` 现在会主动补齐核心 Windows 环境变量；这属于低风险会话级修复，但后续若其它脚本依赖“变量为空”的异常态，需要留意行为差异
- 兼容性风险：低；不改业务配置、不改 API、不改数据库
- 是否阻塞下一任务：不阻塞；前一轮 Phase H 的共享前端静态验证已恢复，剩余只卡在 `vitest/esbuild spawn EPERM`

## 11. 收工记录

- 构建是否通过：通过；`tsc --noEmit -p web/tsconfig.json` 已恢复通过
- 测试是否通过：部分通过
  - 通过：`verify-locale-consistency.mjs`
  - 通过：`verify-setting-info-logic.mjs`
  - 通过：`git diff --check -- scripts/use-node-env.ps1`
  - 未通过：`vitest` 仍在 `vite/esbuild` 加载 `web/vitest.config.ts` 阶段报 `spawn EPERM`
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、workflow、当前状态文档、automation memory、最近 Phase H worklog、`using-superpowers` / `brainstorming` skill 文档、共享 Node/Windows 脚本
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：明确上一轮真正优先项是先恢复 Node 验证环境，而不是继续扩 AI Profile consumer
  - `runtime-win.ps1` / browser smoke wrapper：提供可复用的 Windows 核心环境兜底模式
  - `web/vitest.config.ts` / `scripts/vitest-no-spawn.cjs`：确认剩余阻塞点位于 `esbuild` worker spawn，而不是 Node 主进程初始化
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：不适用
- 手工 smoke 状态：未执行浏览器 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮不需要浏览器；剩余验证阻塞集中在 `vitest/esbuild spawn EPERM`
- 待验证页面清单：设置页 `AIAutomationSource` 与 AI 自动化页顶部摘要卡片的 `vitest` 场景
- 若未使用子 agent，原因：用户明确禁止创建子 agent，且本轮为单脚本小闭环
- worklog 是否更新：已更新
- 遗留项：
  - `vitest` 仍无法在当前宿主通过 `esbuild` worker 启动
  - 用户安装版 `C:\Users\李昊桐\AppData\Local\Programs\nodejs\node.exe` 仍是 `Access is denied`，共享脚本当前稳定回退到 `D:\gptm\live-2d\node\node.exe`
- 下一任务前置条件是否满足：满足；下一轮可直接继续处理 `vitest/esbuild spawn EPERM`，或在保留共享 Node 验证链的前提下回到更小的 Phase H consumer contract 任务
