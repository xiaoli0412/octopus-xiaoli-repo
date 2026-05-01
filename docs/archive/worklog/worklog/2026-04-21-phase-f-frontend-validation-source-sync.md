# 2026-04-21 Phase F Frontend Validation Source Sync

## 1. 任务信息

- 任务名称：前端验证链路源码同步收口
- 日期：2026-04-21
- 当前阶段：Milestone 6 validation closure / Phase F frontend verification credibility
- 对应 milestone：里程碑 6 验证与部署

## 2. 开工前输入

- 对应 canonical 章节：`docs/LLM-Gateway-Refactor-Plan.zh-CN.md` 第 13 节里程碑 6、第 14 节验收标准、第 16 节实施规则
- 对应 workflow 章节：`docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md` 第 9 节与第 10 节
- 上一个相关 worklog：`docs/worklog/2026-04-21-phase-f-manifest-boolean-state-and-smoke-status-sync.md`
- 本次任务目标：让 CI 和本地 smoke 校验当前 `web/src` 导出的静态产物，而不是隐式依赖可能陈旧的 `static/out`
- 本次已盘点本地资源：automation memory、canonical plan、详细执行工作流、前端主线状态文档、现有 validation workflow、Linux/Windows smoke 脚本、`web/package.json`、`static/static.go`、`scripts/build.sh`、`scripts/build-win.ps1`
- 本次使用的本地 resources / skills / 记忆上下文：读取了 `using-superpowers` 与 `brainstorming` 技能说明；复用了 automation memory 中“优先修复前端验证链路”的结论；复用了前端主线状态文档中关于 smoke 边界的记录
- 若未使用部分本地资源或上下文，原因：未继续展开更早的 Phase F 文案型 worklog，因为本轮主任务已切换到验证链路收口，旧文案闭环不再是当前最高优先级
- 本次是否启用子 agent 与分工边界：否
- 本次子 agent 使用模型：无
- 若未使用子 agent，原因：任务范围集中在少量脚本和 workflow 文件，且用户明确要求不要创建子 agent

## 3. 本次硬规则

- 必须围绕里程碑 6 的验证主线推进，不能扩散到无关 UI 微调
- 验证链路必须尽量证明当前源码而不是历史静态产物
- 若实现与文档口径发生变化，必须同步更新状态记录和 worklog

## 4. 本次禁止事项

- 不改主业务逻辑
- 不回退用户现有修改
- 不把 smoke 描述成完整前端行为级自动验收

## 5. 本次验收条件

- 仓库存在一条可重复执行的 `web/src -> web/out -> static/out` 同步命令
- CI 在后端 build 和 smoke 前先执行前端 typecheck 与静态同步
- Linux / Windows smoke 都能优先使用较新的源码导出静态目录
- 至少完成与本轮改动直接相关的最小命令级验证

## 6. 本次回滚点

- 删除 `scripts/sync-web-static.mjs`
- 回退 `web/package.json`
- 回退 `.github/workflows/validation.yaml`
- 回退 smoke 脚本中的静态目录选择逻辑

## 7. 实现范围

- 先改数据语义还是先改 UI：先改验证脚本与 workflow，再同步文档
- 受影响后端模块：无主业务后端改动，仅 smoke 启动配置使用的静态目录来源变化
- 受影响前端模块：`web/package.json` 构建脚本
- 受影响接口：无
- 是否影响旧数据：否
- 是否影响旧行为：会让验证更偏向当前源码产物，减少旧 `static/out` 被误验的概率

## 8. 实施步骤

1. 新增静态同步脚本，并在前端包脚本里暴露统一命令
2. 调整 GitHub Actions validation workflow，让前端 typecheck 与静态同步先于后端验证
3. 调整 Linux / Windows smoke，优先选择更新的源码导出目录并输出实际使用路径
4. 更新前端主线状态文档，记录新的验证链路口径
5. 收敛构建脚本中的静态同步入口，避免 `build.sh` 与 Windows 构建脚本继续各自漂移

## 9. 测试与验证

- 构建命令：`node scripts/sync-web-static.mjs`
- 测试命令：`node -e "const fs=require('node:fs'); JSON.parse(fs.readFileSync('web/package.json','utf8')); console.log('package-json-ok')"`
- 专项验证：`Select-String -Path '.github/workflows/validation.yaml','scripts/smoke-linux-backend.sh','scripts/smoke-win-backend.ps1','docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md' -Pattern 'build:static|SMOKE_STATIC_DIR|Resolve-SmokeStaticDir|当前 `web/src` 导出的静态壳|pnpm run build:static'`
- 构建链补充验证：`Select-String -Path 'scripts/build.sh','scripts/build-win.ps1','scripts/sync-web-static.mjs' -Pattern 'build:static|sync-web-static|Invoke-FrontendStaticSync'`

## 10. 风险与兼容性

- 新风险：如果 `web/out` 是旧产物但时间戳更新，smoke 仍可能优先吃到它；不过这比无条件吃仓库内旧 `static/out` 更接近当前源码链路
- 兼容性风险：低，主要是验证脚本路径选择变化
- 是否阻塞下一任务：否

## 11. 收工记录

- 构建是否通过：已完成本轮最小命令级验证；未在本机完整重跑 Go 源码构建
- 测试是否通过：与本轮改动直接相关的脚本/配置检查通过
- 本次使用了哪些本地资源 / skills / 记忆上下文：automation memory、canonical plan、详细执行工作流、前端主线状态文档、validation workflow、smoke 脚本、构建脚本、`using-superpowers`、`brainstorming`
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：memory 指向“前端验证链路优先”；canonical plan 与 workflow 约束了里程碑 6 的主线和记录方式；前端主线状态文档暴露了当前 smoke 仍可能只验静态壳；现有构建脚本暴露了第二条分叉同步逻辑，所以本轮一并收敛；技能说明用于满足会话起始与实现前检查要求
- 本次使用了哪些子 agent 及其结论：无
- 子 agent 分工、负责范围与产出摘要：无
- 手工 smoke 状态：未执行浏览器手工 smoke
- 手工 smoke 阻塞原因 / 缺少的环境：本轮先收口验证脚本链，未启动浏览器自动化或完整本地 Go 重编译环境
- 待验证页面清单：设置页备份/导入复杂流、日志页实时流、首页窄屏细节
- 若未使用子 agent，原因：用户明确要求主线程执行，且本轮任务紧耦合、改动范围小
- worklog 是否更新：是
- 遗留项：还缺真正的行为级前端验证，尤其是 Phase F 备份/导入/回滚流程；本机仍未完成完整 Go 源码重编译链验证
- 下一任务前置条件是否满足：是，验证链路可信度已提升，适合继续补行为级验证
