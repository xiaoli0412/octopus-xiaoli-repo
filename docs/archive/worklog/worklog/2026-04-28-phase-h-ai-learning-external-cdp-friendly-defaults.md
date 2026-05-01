# 2026-04-28 Phase H AI Learning External CDP Friendly Defaults

## 1. 任务信息

- 任务名称：phase h ai learning external cdp friendly defaults
- 日期：2026-04-28
- 当前阶段：Phase H AI 自动化中心与配置 Profile 双轨主线
- 对应 milestone：Phase H6 动态路由 AI 学习浏览器级证据入口收口

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` Phase H / H6 / 动态路由学习只影响运行时推荐、不覆盖用户配置
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 9.5 Phase H
- 上一个相关 worklog：`docs/worklog/2026-04-28-phase-h-ai-learning-cdp-smoke-handoff.md`
- 本次任务目标：在不改业务 UI 的前提下，把 `AI 自动化` learning smoke 的 `external + cdp` 复跑入口收口成更短、更稳定的 host-friendly 默认组合，减少下一轮继续拼长参数与误用旧 wrapper 的成本
- 本次已盘点本地资源：AGENTS.md、automation memory、canonical plan、CURRENT_STATUS_AND_PLAN、FRONTEND_UI_MAINLINE_STATUS、DETAILED_EXECUTION_WORKFLOW、USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW、Phase H6 连续 worklog、`scripts/verify-ai-automation-learning-browser-smoke.ps1`、`scripts/verify-channel-create-browser-smoke-cdp.ps1`、`scripts/verify-ai-automation-learning-focus.mjs`、`scripts/runtime-win.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：`using-superpowers`、`brainstorming`、automation memory、Phase H6 worklog 连续记录
- 若未使用部分本地资源或上下文，原因：本轮只做验证入口收口，不需要再次展开业务页面与后端实现细节
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 本次是否由 Codex 自动化链路或分目录 agent 协作执行：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进，不创建子 agent

## 3. 本次硬规则

- 继续停留在 Phase H6 learning smoke 主线，不扩散到 AI 页面业务逻辑、动态路由 API 或新的浏览器基建分支
- 只改 wrapper、静态守护和记录文件，保持风险集中且可回退
- `external + cdp` 默认严格语义仍保留；新的 host-friendly 组合必须通过显式开关启用，不能悄悄改掉原有模式

## 4. 本次禁止事项

- 不改 `web/src/components/modules/ai-automation/*` 用户可见行为
- 不改动态路由学习数据结构与设置页交互
- 不把当前宿主 loopback/service-provider 阻塞误写成产品缺陷

## 5. 本次验收条件

- `scripts/verify-ai-automation-learning-browser-smoke.ps1` 新增显式 host-friendly external 开关，并自动注入已知稳定的 `CdpUrl / timeout / preset / profile / external bootstrap` 组合
- `scripts/verify-ai-automation-learning-focus.mjs` 对上述新入口和默认参数有静态守护
- `learning smoke check-only` 能直接展示新的 host-friendly external 组合
- `tsc --noEmit`、守护脚本和 `git diff --check` 通过

## 6. 本次回滚点

- 删除 `UseHostFriendlyExternalDefaults` 开关和对应守护即可回到上一轮 wrapper 语义

## 7. 实现范围

- 先改数据语义还是先改 UI：不改业务数据与 UI，只改验证 wrapper
- 受影响后端模块：无
- 受影响前端模块：无
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：否；默认 strict `external + cdp` 语义仍保留

## 8. 实施步骤

1. 复核上一轮 Phase H6 worklog、shared CDP wrapper 和 learning smoke wrapper，确认最小改动点只在入口层
2. 给 `verify-ai-automation-learning-browser-smoke.ps1` 增加 `UseHostFriendlyExternalDefaults`，把已知稳定的 `external + cdp` 参数组装成显式快捷入口
3. 同步更新 `verify-ai-automation-learning-focus.mjs`，锁定新开关与默认组合，避免后续漂移
4. 运行 check-only、静态守护、`tsc --noEmit` 与 `git diff --check`

## 9. 测试与验证

- 构建命令：`$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- 测试命令：
  - `$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; . .\scripts\use-node-env.ps1; & $env:NODEEXE .\scripts\verify-ai-automation-learning-focus.mjs`
  - `$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'; $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'; & .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode check-only -UseHostFriendlyExternalDefaults`
  - `git diff --check -- scripts/verify-ai-automation-learning-browser-smoke.ps1 scripts/verify-ai-automation-learning-focus.mjs docs/worklog/2026-04-28-phase-h-ai-learning-external-cdp-friendly-defaults.md docs/CURRENT_STATUS_AND_PLAN.zh-CN.md docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 专项验证：check-only 输出已明确显示 `CdpUrl=http://127.0.0.1:9233`、`NodeSmokeTimeoutSeconds=70`、`CdpCommandTimeoutMs=30000`、`EdgeLaunchPreset=relaxed`、`EdgeProfileStrategy=workspace-fixed`，并且自动打开 `BootstrapExternalCdpSession + SelfStartServices`

## 10. 风险与兼容性

- 新风险：低；仅新增 wrapper 组合入口和守护断言
- 兼容性风险：低；默认 external/self-start 行为未被覆盖，只是新增显式快捷开关
- 是否阻塞下一任务：不阻塞；下一轮可以直接复用更短的 host-friendly external 命令继续推进 learning browser 证据

## 11. 收工记录

- 构建是否通过：通过
- 测试是否通过：通过 `verify-ai-automation-learning-focus.mjs`、learning smoke `check-only -UseHostFriendlyExternalDefaults`、`tsc --noEmit`
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、当前状态文档、前端主线状态、Phase H6 连续 worklog、`using-superpowers` / `brainstorming` 技能文档
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - automation memory：确认上一轮已经把 learning smoke 接到共享 CDP wrapper，本轮应继续压缩 external 复跑成本而不是重回 CLI-only
  - CURRENT_STATUS / FRONTEND_UI_MAINLINE_STATUS：确认当前剩余缺口仍是浏览器级证据和宿主阻塞，而不是页面逻辑回归
  - Phase H6 worklog：确认 `9233 + relaxed + workspace-fixed + external bootstrap` 是同类 smoke 里最可复用的 host-friendly 组合
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行真实 browser pass
- 手工 smoke 阻塞原因 / 缺少的环境：当前 Windows 本机 loopback/service-provider 阻塞仍存在，真实 self-start pass 仍不可用；本轮只收口了更易复跑的 external 入口
- 待验证页面清单：`AI 自动化` learning 区的真实 external/browser 证据、必要时和 settings learning 卡做同宿主对照
- 若未使用子 agent，原因：用户明确要求主线程串行推进
- worklog 是否更新：是
- 遗留项：
  - 继续在 host-friendly 条件下优先尝试 `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults`
  - 若仍需进一步对照，再显式覆盖 `CdpPageBootstrapStrategy` 或 `CdpBootstrapCommandOrder`，而不是重新手拼整串参数
  - 当前 `.tmp_phase_h_ai_learning_cdp_*.patch` 临时补丁文件仍在，下一轮若宿主允许删除应优先清理
- 下一任务前置条件是否满足：满足；下一轮可以直接用新的 host-friendly external 入口继续推进 AI learning browser-grade 证据
