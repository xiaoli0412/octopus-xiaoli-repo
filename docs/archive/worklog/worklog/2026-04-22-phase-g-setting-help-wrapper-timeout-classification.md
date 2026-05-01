# 2026-04-22 Phase G 设置页帮助提示 wrapper 超时分类收口

## 1. 任务信息

- 任务名称：设置页帮助提示真实浏览器 smoke 的 wrapper 超时分类收口
- 日期：2026-04-22
- 当前阶段：Phase G 图片问题优先返工窗口
- 对应 milestone：设置页四卡片帮助提示浏览器 smoke 阻塞继续收敛

## 2. 开工前输入

- Master plan aligned before coding (yes/no): yes
- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 9.6、14、16 节
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 1.0、1.2、1.3 节
- 上一个相关 worklog：`docs/worklog/2026-04-22-phase-g-setting-help-external-cdp-preflight-closure.md`
- 本次任务目标：让 `self-start + cdp` 在 Node 进程被 wrapper 总超时截断时，仍能直接输出结构化 CDP 失败分类，而不是只剩 trace tail
- 本次已盘点本地资源：canonical plan、用户上下文总账、详细工作流、前端主线状态、最近三条 setting-help smoke worklog、automation memory、当前 wrapper 与 CDP smoke 脚本
- 本次是否启用子 agent 与分工边界：否
- 若未使用子 agent，原因：用户明确要求主线程串行推进

## 3. 本次硬规则

- 只处理设置页帮助提示浏览器 smoke 的脚本链，不修改设置页业务逻辑
- 不把真实浏览器 smoke 未通过写成已通过
- 必须留下下一轮可直接复跑、可直接读取的失败分类结果

## 4. 本次禁止事项

- 不扩散到备份导入、渠道、多 key、分组或 `CC Switch` 主线
- 不回退工作区中与本轮无关的已有改动
- 不把 trace 级收窄误写成桌面和移动端真实浏览器已通过

## 5. 本次验收条件

- `scripts/verify-setting-help-browser-smoke-cdp.mjs` 在失败时会写出 `cdp.diagnostic.json`
- `scripts/verify-setting-help-browser-smoke.ps1` 在 CDP 失败时会优先回显诊断摘要
- 当 Node 进程先被 wrapper 总超时截断时，wrapper 仍能从 trace tail 推断出最小失败分类
- `check-only` 和最小 `self-start + cdp` 复现都能拿到一致的脚本行为

## 6. 本次回滚点

- `scripts/verify-setting-help-browser-smoke.ps1`
- `scripts/verify-setting-help-browser-smoke-cdp.mjs`
- `docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- `docs/worklog/2026-04-22-phase-g-setting-help-wrapper-timeout-classification.md`
- automation memory `$CODEX_HOME/automations/octopus-2/memory.md`

## 7. 实现范围

- 受影响后端模块：无
- 受影响前端模块：无
- 受影响脚本：
  - `scripts/verify-setting-help-browser-smoke-cdp.mjs`
  - `scripts/verify-setting-help-browser-smoke.ps1`
- 是否影响旧数据：否
- 是否影响旧行为：仅影响 smoke 验证链的失败产物与诊断输出

## 8. 实施步骤

1. 复核主规划、用户上下文、详细工作流、前端主线状态和最近 worklog，确认本轮继续留在 Phase G 设置页帮助提示 smoke 主线。
2. 给 `scripts/verify-setting-help-browser-smoke-cdp.mjs` 增加 `OCTOPUS_UI_SMOKE_CDP_DIAGNOSTIC_FILE` 支持，并在异常退出时写出结构化诊断 JSON。
3. 给 `scripts/verify-setting-help-browser-smoke.ps1` 增加诊断文件读取和摘要格式化逻辑；若诊断文件尚未生成，则从 trace tail 兜底推断 `page_bootstrap_timeout_preempted`。
4. 修正 PowerShell 脚本新增提示文本为 ASCII，避免 Windows PowerShell 5.1 因非 ASCII 提示文本误解析。
5. 运行静态检查、`check-only` 和最小 `self-start + cdp` 复现，确认新的诊断摘要能直接出现在 wrapper 错误里。

## 9. 测试与验证

- 通过：`D:\gol1\node.exe --check scripts/verify-setting-help-browser-smoke-cdp.mjs`
- 通过：`D:\gol1\node.exe scripts/verify-setting-help-browser-smoke-cdp.mjs --check-only`
- 通过：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-setting-help-browser-smoke.ps1 -Mode check-only -Driver cdp -NodeSmokeTimeoutSeconds 20`
- 通过：`try { & .\scripts\verify-setting-help-browser-smoke.ps1 -Mode self-start -Driver cdp -EdgeLaunchPreset relaxed -NodeSmokeTimeoutSeconds 45 -KeepArtifacts } catch { $_.Exception.Message }`
  - 结果：失败路径里已直接包含
    - `CDP diagnostic classification: page_bootstrap_timeout_preempted`
    - `CDP diagnostic error: CdpPageBootstrapPendingTimeout`
    - `CDP diagnostic page mode: json-new`
    - 下一步提示改为“增加 Node timeout 或拿真实外部 Edge 会话对照”

## 10. 风险与兼容性

- 新风险：低；仅调整 smoke 脚本失败产物与错误摘要
- 兼容性风险：低；未改接口、数据库、页面业务逻辑
- 是否阻塞下一任务：不阻塞；下一轮可直接基于新的分类摘要继续做外部会话对照

## 11. 收工记录

- 构建是否通过：脚本级检查通过
- 测试是否通过：通过；`check-only` 正常，最小 `self-start + cdp` 失败路径已能直接输出结构化 CDP 分类摘要
- 本次使用了哪些本地资源 / skills / 记忆上下文：canonical plan、用户上下文总账、详细工作流、前端主线状态、最近 setting-help smoke worklog、automation memory、当前 wrapper 与 CDP smoke 脚本
- 本次使用了哪些子 agent 及其结论：无
- 手工 smoke 状态：仍未形成真实浏览器通过证据，但失败路径已经从“只有 trace tail”收口到“直接给出分类摘要”
- 当前最小阻塞描述：本机 `self-start + cdp` 在 `json-new` 页面阶段的 `Page.enable -> Page.setLifecycleEventsEnabled -> Runtime.enable` 链持续卡死，wrapper 总超时会先于 Node fallback attached-session 收尾；当前应继续用真实外部 Edge remote debugging 会话做对照
- 遗留项：
  - `cdp.diagnostic.json` 仍主要在 Node 自己退出时写出；若 wrapper 更早杀掉 Node，当前依赖 trace tail 兜底分类
  - 桌面与 `375px` 的真实浏览器通过证据仍未闭环
- 下一任务前置条件是否满足：满足

## 12. 下一轮建议

1. 继续留在 Phase G 同主线，优先拿真实外部 Edge remote debugging 会话跑 `external + cdp` 对照，判断是否仍然是同一类 bootstrap 卡死。
2. 若外部会话也落到相同分类，直接把结论升级为宿主级 Edge/CDP page bootstrap 行为，而不是 self-start/profile 特有问题。
3. 若外部会话能越过 bootstrap，再补第一条真实桌面/`375px` 通过证据，关闭设置页帮助提示浏览器 smoke 缺口。

## 13. 最终状态

- 本次结果：成功
- 是否需要人工介入：否
