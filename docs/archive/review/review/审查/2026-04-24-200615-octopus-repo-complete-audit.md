# Findings

## Critical

1. `AI Profile` 切换目前只写设置，不驱动任何实际配置读取或运行时行为，属于“管理面已做、运行面未接入”。
   - 证据：设置页只读写 `config_source_mode` / `active_ai_profile_id`，见 [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:20)、[web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:31)。
   - 证据：后端激活逻辑也只写两个 setting，并把 profile 标成 active，见 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:404)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:407)。
   - 证据：`config_source_mode` / `active_ai_profile_id` 在执行面仅被读取后塞进 AI 任务上下文，没有任何渠道、分组、模型、route-target 或 relay 的消费点，见 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:366)。
   - 佐证：现有测试只证明“不会覆盖手动配置、setting 被切换”，没有证明切换后任何真实读取来源发生变化，见 [internal/op/ai_automation_test.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_test.go:293)、[internal/server/handlers/ai_automation_test.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation_test.go:177)。
   - 影响：文档承诺的“设置页允许在手动配置与 AI 生成方案之间显式切换”当前并未形成运行闭环，用户切换后实际服务行为大概率不变。

2. AI 自动化任务的 `config_snapshot` 只保存在进程内 `sync.Map`，不是持久化任务状态，重启后无法重放或复现同一任务配置。
   - 证据：创建任务时 `ConfigSnapshot` 不入库，只通过 `storeAITaskRuntimeConfig()` 写入进程内 map，见 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:272)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:291)。
   - 证据：执行器通过 `sync.Map` 读取快照，并在 goroutine 结束后删除，见 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:37)、[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:79)、[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:277)、[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:292)。
   - 影响：任务列表、重试、重启恢复、审计复盘只能看到 task 元数据，拿不到当时真正使用的 endpoint / key / model 快照；这与 AI 自动化中心的“任务可回看、可解释”目标不一致。

## High

3. `ai_automation_enabled` 已建模但未见任何执行门控，当前更像未接线开关。
   - 证据：配置读取和更新都包含 `ai_automation_enabled`，见 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:19)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:36)。
   - 证据：`AI Automation` 前端页面没有展示或消费这个开关；页面里真正可切的只有动态学习开关，见 [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:243)、[web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:519)。
   - 证据：任务创建、模型拉取、配置接口和执行器链路没有看到基于 `Enabled` 的拒绝或短路，见 [internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:74)、[internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:117)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:260)、[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:149)。
   - 影响：运维以为能关闭 AI 自动化，其实只是写了一个不会阻止行为的 setting。

4. 顶层状态文档与当前实现严重漂移，仍写“AI 自动化中心代码实现未开始”，会直接误导后续接手和自动化判断。
   - 证据：`CURRENT_STATUS_AND_PLAN.zh-CN.md` 仍写“当前状态：文档规划已立项，代码实现未开始”，见 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:45)、[docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:47)。
   - 证据：实际仓库已经有 AI 自动化前端入口、后端 handler、op、执行器、迁移和测试，见 [web/src/route/config.tsx](/D:/GPT-codex/octopus_repo/web/src/route/config.tsx:30)、[internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:19)、[internal/db/migrate/012.go](/D:/GPT-codex/octopus_repo/internal/db/migrate/012.go:13)。
   - 影响：这不是简单文档落后，而是会让“完成度评估、回归范围、后续排期”持续建立在错误前提上。

5. README 的源码构建说明与当前可行构建路径不一致，容易把用户带到宿主已知失败路径。
   - 证据：README 仍要求 `cd web && pnpm install && pnpm run build && cd ..` 再 `mv web/out static/`，见 [README.md](/D:/GPT-codex/octopus_repo/README.md:88)、[README.md](/D:/GPT-codex/octopus_repo/README.md:90)、[README_zh.md](/D:/GPT-codex/octopus_repo/README_zh.md:87)、[README_zh.md](/D:/GPT-codex/octopus_repo/README_zh.md:89)。
   - 证据：本机直接执行 `next build` 仍然失败于宿主 `spawn EPERM`；真正可用的是仓库自带补丁包装器 `build-web-static.mjs`，见 [scripts/build-web-static.mjs](/D:/GPT-codex/octopus_repo/scripts/build-web-static.mjs:1) 与 [web/next.config.ts](/D:/GPT-codex/octopus_repo/web/next.config.ts:5)。
   - 证据：仓库自己的 Linux 构建脚本已经转到 `pnpm run build:static`，见 [scripts/build.sh](/D:/GPT-codex/octopus_repo/scripts/build.sh:258)；但 Debian Dockerfile 仍在直接跑 `pnpm run build`，见 [scripts/dockerfiles/Dockerfile.debian](/D:/GPT-codex/octopus_repo/scripts/dockerfiles/Dockerfile.debian:7)。
   - 影响：文档、脚本和 Docker 构建口径不统一，发布与本地复现路径存在认知分裂。

## Medium

6. 协议转换边界仍有真实 TODO，当前测试无法证明这些边界场景已完成。
   - 证据：Anthropic `tool_result` 其他结果类型仍是 TODO，见 [internal/transformer/inbound/anthropic/messages.go](/D:/GPT-codex/octopus_repo/internal/transformer/inbound/anthropic/messages.go:162)。
   - 证据：Gemini `json_schema` 转换仍是 TODO，见 [internal/transformer/outbound/gemini/messages.go](/D:/GPT-codex/octopus_repo/internal/transformer/outbound/gemini/messages.go:323)。
   - 证据：统一模型层仍有 `Schema` 与 image-generation tool body 的 TODO，见 [internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:618)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:861)。
   - 影响：主流 chat / response / embedding 主链路已证明可用，但结构化输出、复杂 `tool_result` 和图像工具边角能力仍是部分完成。

7. CI 覆盖面仍小于仓库当前真实测试面，容易出现“本地全绿、CI 只测子集”的假安全感。
   - 证据：CI 只跑 `./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task`，见 [validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:58)、[release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:56)。
   - 证据：仓库实际已有 `internal/client`、`internal/conf`、`internal/helper`、`internal/model`、`internal/server/auth`、`internal/transformer/model`、`internal/transformer/outbound/openai`、`internal/update`、`internal/utils/httpx` 等额外测试文件，这些并不在 CI 显式范围内；虽然 `go test ./...` 本次本机通过，但 CI 仍无法独立证明同等覆盖。
   - 影响：针对非 CI 包的回归更容易在 release 前漏掉。

8. `verify-go-env.ps1 -GoFmt` 仍然无差别递归整个仓库，误扫 `.tools\go` 自带 testdata，导致格式检查结论失真。
   - 证据：脚本直接对仓库根做 `Get-ChildItem -LiteralPath $repoRoot -Recurse -Filter '*.go' -File`，没有排除 `.tools`，见 [scripts/verify-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:87)。
   - 实测：本次运行失败在 `.tools\go\go\src\cmd\compile\internal\syntax\testdata\issue20789.go`，并非仓库业务代码本身。
   - 影响：格式检查不能作为当前工作区 Go 代码卫生的可靠信号。

## Low

9. 当前工作区仍存在基础卫生问题，`git diff --check` 报告两个前端文件 EOF 空行。
   - 证据：`git diff --check` 报 [web/src/components/modules/channel/Form.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/channel/Form.tsx:2188) 和 [web/src/components/modules/group/Editor.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/group/Editor.tsx:1043) 有 `new blank line at EOF`。
   - 影响：不是功能阻塞，但会持续污染 review 与格式检查噪音。

10. `baseline-browser-mapping` 数据过旧告警仍存在，虽然不阻断构建，但说明前端构建依赖存在维护债务。
   - 证据：本次 `build-web-static.mjs` 成功，但输出了多次 `baseline-browser-mapping` 数据过旧告警。

# Completion Assessment

已完成：
- 后端启动链、数据库迁移、鉴权、静态资源回退、`/healthz`、管理 API 主链、relay 主链都是真接线。
- Windows 后端烟测覆盖了 `login -> channel -> group -> apikey -> /v1/chat/completions`，证明最小业务链真实可达。
- 动态路由当前实现不是空壳：模式状态、推荐层、学习状态记录、摘要 API、设置页摘要面和日志审计字段都已接入。
- 备份导入主链、相关 no-browser 校验和组件校验已形成较完整闭环。
- AI 自动化中心的“任务创建、执行、保存 Profile、列表/激活”管理链已实现，不再是文档-only。

部分完成：
- AI 自动化双轨配置主线只完成了“管理面 + 任务面”，未完成“切换后真实消费 AI Profile”的运行闭环。
- 协议转换主路径可用，但高级结构化输出、复杂 `tool_result`、image generation tool body 等边界仍是半实现。
- 前端构建链在仓库包装脚本下可用，但 README、Dockerfile、直接 `next build` 的口径还没完全收口。

未完成：
- `AI Profile` 作为运行时配置来源的真实消费。
- AI 任务快照持久化与重启恢复。
- 直接 `next build` / Vitest 在当前宿主上的无补丁闭环。
- Docker 运行态本机验证。

疑似空实现：
- `ai_automation_enabled` 目前最接近疑似空实现。它已定义、可存储，但本次没有找到任何会真正尊重该开关的执行门控。

总体完成度判断：
- 仓库整体完成度约可评为 `75%~82%`。
- 核心网关、后台管理、备份与动态路由基础能力已经进入“可用并可验证”阶段。
- 当前最大缺口不再是主流程能不能跑，而是若干“看起来已经产品化”的高级能力还停留在管理面或半接线状态。

# Verification Summary

已验证项：
- `git status --short --branch`
- `git branch -a --verbose --no-abbrev`
- `git log --oneline --decorate -n 12`
- `git diff --stat HEAD`
- `git diff --stat origin/dev...HEAD`
- `git diff --check`
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1 -GoTest -GoBuild` 通过
- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 通过
- `D:\gol1\node.exe .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json` 通过
- `D:\gol1\node.exe .\scripts\verify-locale-consistency.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-channel-create-flow.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-channel-presentation.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-dynamic-routing-help.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-group-create-flow.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-llm-price-boundary.mjs` 通过
- `D:\gol1\node.exe --experimental-strip-types .\scripts\verify-backup-logic.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-help-hint-accessible.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-circuit-breaker-help.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-route-target-copy.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-model-probe-help.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-setting-info-logic.mjs` 通过
- `D:\gol1\node.exe .\scripts\verify-backup-component.cjs` 通过
- `D:\gol1\node.exe .\scripts\verify-ccswitch-flow.mjs` 通过
- `D:\gol1\node.exe .\scripts\build-web-static.mjs` 通过

未验证项：
- Docker 相关命令，本机缺少 `docker` 可执行文件。
- 原生 `next build` 全闭环：当前宿主仍报 `spawn EPERM`。
- Vitest 全闭环：当前宿主在 `esbuild/vite` 启动阶段仍报 `spawn EPERM`。
- HTTP 级别的动态路由模式切换到 relay 日志审计字段联动 E2E。
- `AI Profile` 切换后真正改变运行时读取来源的 E2E，因为代码面已显示该消费链不存在。

# Comparison Notes

当前工作区 vs `HEAD`：
- 当前工作区极脏，远超普通审计规模，包含大量已跟踪修改与未跟踪新增。
- 相对 `HEAD` 的主要新增集中在：AI 自动化中心、动态路由、备份导入增强、route-target override、前端设置与渠道/分组 UI、大量测试与验证脚本。
- 本次确认这些差异里相当一部分已是真实现，不应按“未接线草稿”一概处理。

当前分支 vs 稳定基线：
- 当前分支 `feat/erguotou` 相对 `origin/dev` 有大量前移提交，`origin/dev` 没有反向领先提交。
- 这意味着当前审计的“稳定基线”更适合用 `origin/dev`，而不是把问题归因于未同步上游。

代码实现 vs README/docs：
- 一致：README 对动态路由后台任务已收口为 `dynamic summary scan`，与代码一致。
- 不一致：README 仍把源码构建主路径写成 `pnpm run build`，但本机可用主路径已经转到 `build:static` 包装器。
- 不一致：`CURRENT_STATUS_AND_PLAN.zh-CN.md` 仍声称 AI 自动化中心“代码实现未开始”，与当前代码事实明显冲突。
- 不一致：AI 自动化需求文档要求“设置页切换手动配置 / AI Profile 后改变读取来源”，但代码只完成 setting 切换，没有真实消费方。

代码实现 vs 测试/构建/验证覆盖：
- 一致：Go 主链、Windows 烟测、多个 no-browser 校验都能证明核心后端与大量前端契约真实可用。
- 不一致：CI 的 Go 测试范围小于仓库现有测试面，本地 `go test ./...` 才覆盖到完整 Go 测试集合。
- 不一致：Vitest 与原生 `next build` 在当前宿主不可用，但包装构建脚本可用，说明“仓库可构建”和“宿主可直接运行标准命令”不是一回事。

# Top Next Actions

需要优先处理的前三项：
1. 把 `AI Profile` 切换真正接入运行时读取链，或在接入前下调文档承诺并把设置页文案改成“仅记录目标来源”。
2. 把 AI 任务 `config_snapshot` 持久化到数据库，至少保证任务回看、重试和重启后能复现执行上下文。
3. 收口构建与验证口径：修复 `verify-go-env.ps1 -GoFmt` 的扫描范围，统一 README / Dockerfile / 脚本到 `build:static` 现实路径，并明确宿主 `spawn EPERM` 限制。

建议下一步动作：
1. 为 `config_source_mode` 设计真实消费层，最小化实现可以先做“管理界面读取来源切换 + 明确回退手动配置”，再决定是否进入 relay/业务配置读取。
2. 给 AI 任务表增加持久化快照字段，迁移后让 `AITaskGet` 能完整返回当次配置。
3. 更新 `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、README 和相关 AI 文档，删掉“未开始”与过时构建指令。
4. 调整 `.github/workflows/validation.yaml` / `release.yaml`，至少考虑升格到 `go test ./...`，或明确说明为何只跑子集。
5. 为动态路由补一个 HTTP 级 E2E，验证设置变更会反映到 relay log 的 `dynamic_routing_*` 审计字段。

---

## 中文摘要

1. 本次触发时间
- `2026-04-24T20:06:16.248+08:00`

2. 做了哪些检查、运行了哪些命令
- 检查了仓库结构、`git status`、分支、最近提交、`README.md`、`README_zh.md`、AI/动态路由需求文档、自动化记忆。
- 做了 `git diff HEAD`、`git diff origin/dev...HEAD`、`git diff --check`、核心代码链路和关键模块接线核查。
- 运行并通过：`verify-go-env.ps1 -GoTest -GoBuild`、`smoke-win-backend.ps1`、TypeScript `--noEmit`、locale/channel/group/dynamic-routing/backup/help-hint/model-probe/route-target/ccswitch 等 no-browser 脚本、`build-web-static.mjs`。
- 运行但失败：原生 `next build`、Vitest、`verify-go-env.ps1 -GoFmt`、`docker --version`。

3. 修改了哪些文件
- 新增审查报告：`docs/review/审查/2026-04-24-200615-octopus-repo-complete-audit.md`
- 本次未修改业务代码。

4. 发现了什么问题
- `AI Profile` 切换只写 setting，不驱动真实读取来源。
- AI 任务 `config_snapshot` 只存在进程内 `sync.Map`，不持久化，重启后不可复现。
- `ai_automation_enabled` 疑似空开关，未见执行门控。
- `CURRENT_STATUS_AND_PLAN.zh-CN.md` 明显过时，仍写 AI 自动化未开始。
- README / Dockerfile / 实际构建路径口径不统一。
- 协议转换仍有真实 TODO 边界。
- `verify-go-env.ps1 -GoFmt` 误扫 `.tools\go` 导致结论失真。

5. 本次结果是成功、跳过还是失败
- `成功`。完整审计已完成，报告已生成；少数命令因宿主限制失败，但原因已明确记录。

6. 是否需要我手动介入
- `需要`。优先介入 `AI Profile` 真实消费设计、AI 任务快照持久化方案，以及构建/文档口径统一。
