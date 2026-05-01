# Octopus 完整审计报告

- 本次触发时间：2026-04-29 17:51:55 +08:00
- 仓库路径：`D:\GPT-codex\octopus_repo`
- 审计分支：`feat/erguotou`
- 审计提交：`bfa27aecbec14329a43550c0d5409aefea257af0`
- 对比基线：`origin/dev`

## 1. Findings

1. `High` AI Profile 的 rich typed payload 仍未形成需求文档承诺的“主消费合同 / 后端校验”闭环。
   判断依据：需求文档要求结果区输出结构化 Profile，并明确 typed payload 是主消费合同，且验收应包含后端可校验能力，见 [docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:68)、[docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:97)、[docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:183)。当前执行器确实会生成 `grouping_suggestions`、`candidate_channel_model_mappings`、`price_items`、`classification`、`issues`、`suggested_actions` 等 rich domain payload，见 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:1346)、[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:1370)。但运行时真正消费 typed payload 的主路径只会从 `typed_config / config / runtime` 里抽取 `base_url`、`api_key`、`channel_type`、`model`、`use_local_default` 这些执行配置字段，见 [internal/op/ai_automation_profile_typed.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_profile_typed.go:347)、[internal/op/ai_automation_profile_typed.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_profile_typed.go:402)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:837)、[internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:893)。而 `grouping_suggestions / price_items / classification / suggested_actions / blocking_activation` 这类 richer 字段，本轮检索只确认被生成、测试和 UI 预览使用，未发现真实 downstream consumer 或 validator；前端也主要是优先展示 `domain_payload` 与迁移状态，见 [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:229)、[web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:2515)。
   影响：AI Profile 不是空实现，但当前更接近“结构化预览 + AI 配置切换”而非“结构化建议已被后端主链真实消费和校验”。这应归类为核心特性部分完成，而不是完成闭环。

2. `Medium` [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:459) 的验收状态描述已经明显过时，会直接误导后续接手和自动化审计。
   判断依据：文档仍写“当前环境下也无法直接复跑”，并把原因归结为 “本机 `go` 不在 PATH” 与 “本机 `pnpm` 未正确装好链接”，见 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:459)、[docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:461)。同一文档的结论表还把 `go test / web build / docker 验收` 写成“未完成 / 没有收口证据”，见 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:511)。但本轮在当前工作区已实际通过 `go build ./...`、`go test ./... -count=1`、`node web/node_modules/typescript/bin/tsc --noEmit -p web/tsconfig.json`、`node scripts/build-web-static.mjs`、`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 和整组 repo-local no-browser 校验。文档对当前可复跑性和验证闭环的结论已经不再成立。
   影响：这不是运行时 bug，但会让后续实现、审计和交接建立在错误前提上，持续制造“仓库还没验证过”的假判断。

3. `Medium` `use-go-env.ps1` 与接手文档承诺“补齐工作区缓存目录”，但实际不会纠正继承下来的坏缓存路径，文档给出的标准入口可能直接导致假失败。
   判断依据：接手文档明确把 `. .\scripts\use-go-env.ps1` 描述成“为当前 PowerShell 会话解析 Go 工具链并补齐工作区缓存目录”，见 [AGENTS.md](/D:/GPT-codex/octopus_repo/AGENTS.md:215)、[AGENTS.md](/D:/GPT-codex/octopus_repo/AGENTS.md:216)。但脚本本身只在 `GOCACHE` 或 `GOTMPDIR` 为空时才赋值，见 [scripts/use-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/use-go-env.ps1:229)、[scripts/use-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/use-go-env.ps1:234)，并没有像 [scripts/verify-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:76) 那样按“当前值不可写则回退到工作区目录”的方式处理，也没有兜底 `GOMODCACHE`。本轮第一次按文档直接执行 `. .\scripts\use-go-env.ps1; go build ./...` 时，构建立刻被继承的 `D:\DevCache\go-mod` / `D:\DevCache\go-build` 不可写路径打断，报错为 `Access is denied`；显式把 `GOMODCACHE/GOCACHE/GOTMPDIR/TEMP/TMP` 指到仓库内目录后，`go build ./...` 与 `go test ./... -count=1` 即全部通过。
   影响：这是官方开发入口与真实行为不一致的问题，会把“宿主缓存脏了”的情况误报成“当前仓库构建坏了”，直接降低 Windows 本地验证的可信度。

4. `Medium` 协议转换主链可用，但高级边界仍有真实 TODO，当前只能判定为部分完成。
   判断依据：`internal/transformer/model/model.go` 在 audio / prediction / JSON schema / image-generation 等位置仍保留 TODO，见 [internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:142)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:208)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:618)、[internal/transformer/model/model.go](/D:/GPT-codex/octopus_repo/internal/transformer/model/model.go:861)。`internal/utils/tokenizer/tokenizer.go` 也还留着“更多模型” TODO，见 [internal/utils/tokenizer/tokenizer.go](/D:/GPT-codex/octopus_repo/internal/utils/tokenizer/tokenizer.go:8)。本轮 `go test ./... -count=1` 已证明核心聊天/嵌入主链未坏，但没有额外证据表明这些 TODO 分支已被完整覆盖。
   影响：主流链路目前是真可用的，但 richer provider edge cases 仍应视作部分完成。若对外继续使用“无缝协议转换”口径，应补测试或收窄说明。

5. `Low` `.gitignore` 过窄，当前工作区噪音已经明显影响审计与回归判断。
   判断依据：当前 `git status --short` 共 `346` 行；而 `.gitignore` 只显式覆盖 `data`、`build`、`static/out/*`、`.tools/`、`.tmp*` 和 `CON` 等少量路径，见 [/.gitignore](/D:/GPT-codex/octopus_repo/.gitignore:1)、[/.gitignore](/D:/GPT-codex/octopus_repo/.gitignore:18)。本轮状态里仍包含大量 cache、logs、临时 patch、工具产物与文档中间件。
   影响：这不会直接破坏运行时，但会提高误提交本地产物、误判真实 diff 和漏看业务回归的概率。

6. `Low` `config_health` 与 `config_health_check` 的命名在需求文档与代码之间仍有轻微漂移。
   判断依据：中文需求文档的类型列表仍保留 `config_health`，见 [docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:83)，但同文档后文与实施计划已经把 canonical 名称定为 `config_health_check`，见 [docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md:179) 和 [docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md:262)。代码常量也统一使用 `config_health_check`，见 [internal/model/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/model/ai_automation.go:56)。
   影响：问题级别不高，但会持续制造检索、文档维护和测试命名噪音。

## 2. Completion Assessment

总体判断：

- 核心网关主线完成度约 `88%-90%`。启动链、管理 API、中继 API、Go 构建与测试、前端类型检查、静态导出、Windows 后端最小 smoke 与 repo-local no-browser 验证链都已真实可达。
- 最近扩展能力完成度约 `76%-82%`。AI 自动化、动态路由学习、备份增强和 richer provider support 都不是空壳，但其中多处仍停留在“主体已接线、rich 语义或边界未闭环”的状态。

已完成：

- `cmd/start.go -> server -> handlers -> relay` 主链真实可达；本轮 `go build ./...` 与 `go test ./... -count=1` 均通过。
- 前端 `tsc --noEmit`、`build-web-static.mjs`、整组 repo-local no-browser 校验和 Windows backend smoke 已形成真实验证闭环。
- AI 自动化任务、历史列表、Profile 持久化、typed payload 回填、任务恢复和设置页 `manual / ai_profile` 切换均已接线，不再属于“未实现”。
- 动态路由学习的设置页开关、样本摘要、重置、AI 自动化页学习区和运行时评分开关已经形成 UI/API/relay 闭环。
- 备份导出、dry-run、`incremental / map / merge / replace / skip` 导入模式、结构化兼容性报告与 rollback preview 主链已经是可运行实现。

部分完成：

- AI Profile rich typed payload 仍缺少文档承诺的主消费闭环与后端 validator。
- 协议转换高级边界仍保留 TODO，当前应视为“主链完成、边角部分完成”。
- 备份高级迁移工具仍显式处于 partial 状态，UI 与文档也明确写了 `Advanced migration tooling still pending`。
- Windows 会话级 Go 启动入口 `use-go-env.ps1` 不能稳定收敛坏缓存环境，导致官方文档入口有假失败风险。

未完成：

- richer AI Profile 域的真实 downstream consumer / validator。
- 备份高级迁移的 conflict wizard、完整 mapping editor、交互式 diff 与更细粒度 partial restore。
- 当前宿主上的 Vitest 本机证明、Linux smoke、Docker compose smoke 与人工前端 checklist。

疑似空实现：

- 本轮未再确认到“UI 已声明但主流程完全未接入”的纯空实现。
- 多 AI lane 控制台虽然只是前端 fan-out，但这一点已被需求文档显式承认，不再属于伪接线。

## 3. Verification Summary

已验证项：

- Git / 版本基线：`git status --short --branch`、`git branch -a --verbose --no-abbrev`、`git log --oneline --decorate -n 20`、`git rev-list --left-right --count origin/dev...HEAD`、`git diff --shortstat HEAD`、`git diff --stat origin/dev...HEAD`。
- Go 构建与测试：在显式把 `GOMODCACHE/GOCACHE/GOTMPDIR/TEMP/TMP` 指向仓库可写目录后，`go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1` 与 `go test ./... -count=1` 全部通过。
- 前端类型与静态导出：`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`node .\scripts\build-web-static.mjs` 通过。
- Windows 后端运行态 smoke：`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1` 通过，覆盖 `/`、`/manifest.json`、`/healthz`、登录、建 channel/group/apikey 和 `/v1/chat/completions` 转发。
- repo-local no-browser 校验：`verify-locale-consistency`、`verify-home-layout`、`verify-channel-create-flow`、`verify-channel-presentation`、`verify-group-create-flow`、`verify-llm-price-boundary`、`verify-backup-logic`、`verify-backup-component`、`verify-help-hint-accessible`、`verify-dynamic-routing-help`、`verify-circuit-breaker-help`、`verify-model-probe-help`、`verify-setting-info-logic`、`verify-ai-config-profile-summary`、`verify-ai-automation-learning-focus`、`verify-ccswitch-flow`、`verify-route-target-copy` 全部通过。
- `runtime-win.ps1 -Action status` 已验证脚本可正常读取“当前本机默认停驻、8080 未监听”的事实；`runtime-win.ps1 -Action healthcheck` 在服务未启动时正确失败为连接被拒绝，而不是脚本内部崩溃。

未验证项：

- Vitest：`node .\web\node_modules\vitest\vitest.mjs run --config .\web\vitest.config.ts` 仍在装载 `vitest.config.ts` 阶段被宿主 `esbuild spawn EPERM` 卡住，无法在本机证明真实单测结果。
- Linux smoke：`wsl -l -v` 当前返回 `Wsl/EnumerateDistros/Service/E_ACCESSDENIED`，本机不能直接复跑 Linux / WSL 路径。
- Docker compose smoke：`docker --version` 当前显示命令不可用，本机无法执行 `scripts/smoke-docker-compose.sh`。
- 人工前端 checklist：`docs/MANUAL_FRONTEND_CHECKLIST.zh-CN.md` 本轮未做人眼验收。

需要特别说明的验证事实：

- 按 AGENTS 文档直接执行 `. .\scripts\use-go-env.ps1; go build ./...` 的第一轮结果并没有证明源码有错，失败原因是继承到不可写的 `D:\DevCache\go-mod` / `D:\DevCache\go-build` 缓存路径。这一宿主问题在切换到仓库内缓存目录后已经被消除。

## 4. Comparison Notes

当前工作区与 `HEAD`：

- 工作区相对 `HEAD` 仍然非常脏，`git status --short` 共 `346` 行，但这轮没有确认到“当前工作区已经打断 Go build / Go test / tsc / static build”的硬回归。
- 主要风险来自业务改动与大量本地产物混在同一工作区，增大误判 diff 的概率，而不是当前源码已经进入不可构建状态。

当前分支与主分支 / 最近稳定基线：

- `origin/dev...HEAD = 0 behind / 22 ahead`，当前分支没有落后 `origin/dev`。
- 这说明风险不在“分支没同步基线”，而在“当前工作区累积了大批未提交改动和审查产物”。

代码实现与 README / docs / 任务说明：

- 动态路由学习的 README 口径与代码实现已基本一致：README 说明“daily background dynamic-routing task does not persist new thresholds or reorder user routing”，见 [README.md](/D:/GPT-codex/octopus_repo/README.md:175)；需求文档也明确 `dynamic_routing_learning_enabled` 只影响推荐评分，不写回用户 priority，见 [docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:22)、[docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:107)、[docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md:187)。这条主线本轮没有发现新的实现漂移。
- 备份高级迁移仍然明确 partial，代码、UI 和 docs 已基本对齐，不再属于新的不一致问题。
- 真正仍然不一致的是 `CURRENT_STATUS_AND_PLAN.zh-CN.md` 对当前可复跑性和验证状态的判断，以及 AI Profile rich payload 被文档写成“主消费合同”，但当前实现主要仍停留在 preview/config-only consumer。

代码实现与测试 / 构建 / 验证覆盖：

- GitHub Actions 已把 `pnpm run test`、`pnpm exec tsc --noEmit`、`pnpm run build:static`、定向 Go 测试、`go test ./internal/update -count=1`、Linux backend smoke 与 Docker compose smoke 都纳入 gate，见 [/.github/workflows/validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:46)、[/.github/workflows/validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:65) 和 [/.github/workflows/release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:46)、[/.github/workflows/release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:63)。
- 本轮在当前宿主上实际证明了其中大部分链路，但 Vitest/Linux/Docker 仍受环境限制未能同机完成。这意味着“仓库声明的验证闭环”总体是成型的，只是本机无法把所有 gate 都本地复现。

## 5. Top Next Actions

需要优先处理的前三项：

1. 为 AI Profile rich typed payload 选择明确闭环方向。
   要求：要么为 `grouping / channel_recognition / price_recognition / model_classification / config_health_check` 至少各接上一条真实 downstream consumer 或 validator；要么收缩需求文档和 UI 口径，把它明确降级为“结构化预览 + 审计记录”，不要继续写成 typed payload 主消费合同已闭环。

2. 修正 Windows Go 会话启动入口与文档承诺不一致的问题。
   要求：让 `scripts/use-go-env.ps1` 像 `scripts/verify-go-env.ps1` 一样，能够在继承坏 `GOMODCACHE/GOCACHE/GOTMPDIR` 时自动回退到仓库可写目录；同时更新 AGENTS 文档，避免把一个会假失败的入口继续写成标准起手式。

3. 补协议转换高级边界的直接证明。
   要求：优先围绕 `audio / prediction / json_schema / image-generation / complex tool_result` 这些 TODO 分支补单测或集成测试；若短期不支持，就同步收窄 README / docs 的“无缝协议转换”口径。

建议下一步动作：

- 在处理完上述三项后，立即更新 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:459) 的过时结论，把“当前无法复跑 / 无验收证据”的描述替换为当前真实验证状态。
- 顺手扩充 `.gitignore`，把本地缓存、日志、patch、中间产物和 automation 输出进一步收敛，降低后续审计噪音。
- 当可用 Linux / Docker 或 CI 环境就绪时，补跑 Vitest、`smoke-linux-backend.sh` 与 `smoke-docker-compose.sh`，把当前仍受宿主限制的未验证项补齐。

## 中文摘要

1. 本次触发时间：`2026-04-29 17:51:55 +08:00`
2. 做了哪些检查、运行了哪些命令：读取 automation memory、git 状态/分支/最近提交/与 `origin/dev` 差异，抽查 README、workflow、AI 自动化、动态路由、备份、transformer、脚本与现有审查产物；运行了 `. .\scripts\use-go-env.ps1; go build ./...`、带仓库内缓存的 `go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task ./internal/update -count=1`、`go test ./... -count=1`、`node .\web\node_modules\typescript\bin\tsc --noEmit -p .\web\tsconfig.json`、`node .\scripts\build-web-static.mjs`、`powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\smoke-win-backend.ps1`、17 条 repo-local no-browser 校验脚本，以及 `vitest`、`runtime-win.ps1 -Action {status,healthcheck}`、`docker --version`、`wsl -l -v` 试跑以确认环境限制。
3. 修改了哪些文件：新增 [docs/review/审查/2026-04-29-175155-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-29-175155-octopus-repo-complete-audit.md:1)、`docs/review/审查/2026-04-29-175155-octopus-repo-complete-audit.html`，并更新 automation memory；除审查产物外，本次无业务变更。`build-web-static.mjs` 额外刷新了被忽略的 `static/out` 产物。
4. 发现了什么问题：最重要的问题是 AI Profile rich typed payload 仍缺少真实 downstream consumer / validator；其次是 `CURRENT_STATUS_AND_PLAN.zh-CN.md` 已经过时、`use-go-env.ps1` 会让文档标准入口出现假失败，以及协议转换高级边界仍保留 TODO；另外 `.gitignore` 过窄与 `config_health` 命名漂移属于低优先级持续噪音。
5. 本次结果是成功、跳过还是失败：`成功`。完整审计已完成并落盘；当前工作区的 Go 构建/测试、`tsc`、静态导出、Windows 后端 smoke 和 repo-local no-browser 验证链均已通过。
6. 是否需要我手动介入：`需要`。优先需要拍板的是 AI Profile rich payload 到底要做真实 consumer / validator，还是收缩需求与 UI 口径；其次建议尽快确认是否接受修补 `use-go-env.ps1` 与更新 `CURRENT_STATUS_AND_PLAN.zh-CN.md` 的过时结论。
