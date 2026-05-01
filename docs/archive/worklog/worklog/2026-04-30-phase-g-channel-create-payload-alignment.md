# 2026-04-30 Phase G Channel Create Payload Alignment

- Master plan aligned before coding (yes/no): yes
- Mainline: `Phase G screenshot-first UI closure / channel create form contract alignment`
- Current stage: `channel-create 创建链配置传递收口`

## 本轮上下文与本地资源

- canonical / 当前主线文档：`docs/UI_MAINLINE_TASK_2026-04-30.zh-CN.md`
- 前端验收清单：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md`
- 主规划：`docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`
- 当前状态：`docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md`
- 前端主线状态：`docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md`
- 相邻 worklog：
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-group-create-browser-and-advanced-strategy-recovery.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-model-layout-browser-smoke-recovery.md`
  - `docs/archive/worklog/worklog/2026-04-30-phase-g-backup-cli-host-blocker-classification-closure.md`
- 自动化连续性：`$CODEX_HOME/automations/octopus-2/memory.md`

## Skills / 协作说明

- 已按会话要求检查 `using-superpowers`。
- `brainstorming` 仅作为约束核对，不进入设计门禁；本轮属于既有 canonical plan 下的增量重构与验证收口，且用户要求直接执行落地。
- 未使用子 agent；原因：用户明确要求“不要创建子agent，靠主线程解决”。

## 本轮候选任务

1. 复查 `channel-create` 测试链，确认 `channel is disabled` 是否源于配置传递缺口。
2. 若创建链 payload 丢字段，先修创建表单到 API 请求的配置透传，不放宽 verifier。
3. 回归 `channel-create-flow`、`tsc` 与 `check-only`，给下一轮留下更细浏览器复现入口。

## 本轮计划

- 本轮核心任务：修复 `channel-create` 创建提交链丢失 `key_management_mode / key_routing_policy / key source_type / allowed_models` 的配置传递缺口。
- 本轮配套任务：补齐创建默认态显式值，并把 `verify-channel-create-flow.mjs` 收口到真实源码口径。
- 预期验证方式：
  - `node .\scripts\verify-channel-create-flow.mjs`
  - `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only`
  - `git diff --check -- web/src/components/modules/channel/Create.tsx scripts/verify-channel-create-flow.mjs`
- 完成判定标准：
  - 创建 payload 不再只提交 `enabled/channel_key/remark` 的瘦字段版本。
  - `channel-create-flow` 静态守护通过。
  - 类型检查与 `check-only` 继续通过。

## 本轮硬规则

- 只在 `Phase G` 主线内推进，不扩散到无关 UI 改造。
- 写入范围控制在 channel create 提交链和对应 verifier。
- 不回退用户已有改动，不清理无关脏区。

## 本轮禁止事项

- 不改后端健康检查语义。
- 不顺手重做 channel 编辑卡或其它页面 smoke。
- 不通过删除断言来掩盖创建链字段丢失。

## 本轮回滚点

- `web/src/components/modules/channel/Create.tsx`
- `scripts/verify-channel-create-flow.mjs`

## 执行与结果

1. 先按 `UI_MAINLINE_TASK_2026-04-30` 与前端验收清单重新聚焦 `channel-create` 主线，确认这一轮最值得推进的是“配置传递缺失导致测试链/提交链不一致”的小闭环，而不是继续泛化排查宿主问题。
2. 复查 `ChannelForm`、`Create.tsx`、`CardContent.tsx`、`channel.ts` 与后端 `test-models-by-config` handler 后，确认创建表单默认态和提交流程确实存在一处真实缺口：
  - 创建提交只传 `enabled / channel_key / remark`，丢掉了 `key_management_mode`、`key_routing_policy`、`source_type`、`allowed_models`。
  - 这会让“表单里配置过的 key 分类/模型范围/来源类型”在创建落库时丢失，属于 UI 主线明确点名的配置传递不一致。
3. 因此没有继续空转追 `channel is disabled` 现象，而是先把已经确认的主线缺口收口：
  - `Create.tsx` 初始化和 reset 状态现在显式带 `key_management_mode='pooled'`、`key_routing_policy='round_robin'`。
  - 默认 key 行现在显式带 `source_type='unknown'` 与 `allowed_models=''`，不再依赖隐式空值。
  - 创建提交 payload 现在透传 `key_management_mode`、`key_routing_policy`，并对每个 key 一并提交 `source_type` 与 `allowed_models`。
4. 回归时发现 `verify-channel-create-flow.mjs` 里原本有一条已失真的断言，去匹配并不存在于 `Create.tsx` 的 `data-testid="channel-create-dialog"`；该 test id 实际属于运行态 dialog 容器而非当前文件。
5. 没有删除 verifier，而是把它改成对真实源码更有价值的守护：
  - 断言创建默认态包含 `pooled / round_robin / source_type / allowed_models`。
  - 断言 `Create.tsx` 继续通过 `ChannelForm idPrefix="new-channel"` 接入当前创建链。
  - 断言创建提交继续透传本轮补齐的字段。
6. 回归后，`channel-create-flow`、`tsc --noEmit`、`check-only` 与 `git diff --check`（仅剩 LF/CRLF 警告）都通过。

## 验证

- passed `node .\scripts\verify-channel-create-flow.mjs`
- passed `. .\scripts\use-node-env.ps1; & $env:NODEEXE .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`
- passed `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-channel-create-browser-smoke.ps1 -Mode check-only`
- passed `git diff --check -- web/src/components/modules/channel/Create.tsx scripts/verify-channel-create-flow.mjs`

## 本轮变更文件

- `web/src/components/modules/channel/Create.tsx`
- `scripts/verify-channel-create-flow.mjs`
- `docs/archive/worklog/worklog/2026-04-30-phase-g-channel-create-payload-alignment.md`

## 未完成 / 风险 / 阻塞

- 本轮没有拿到新的“运行态实际复现 `channel is disabled`”证据；当前完成的是更前置、更确定的配置传递收口。
- `git diff --check` 仅剩 LF/CRLF 警告，不构成当前主线阻塞。
- 本轮没有改运行态页面结构源码之外的静态产物，也未触发 `build:static` / `web/out -> static/out` 同步；若下一轮继续改 channel create 可见结构或样式，仍需补这条链。

## 下一轮候选任务顺序

1. 延续 `Phase G`，继续跑 `channel-create` 的真实 browser/self-start 交互，优先确认是否还存在“测试时统一落到 disabled”这一运行态现象。
2. 若真实现象仍在，优先在 `channel-create` / `channel detail test key` 这两条直接调用 `test-models-by-config` 的路径上加更细的请求断言或日志，而不是回头重扫全仓。
3. 若运行态已不再出现该现象，则转向相邻页面级 smoke 缺口，继续保持 `Phase G` 小闭环推进。
