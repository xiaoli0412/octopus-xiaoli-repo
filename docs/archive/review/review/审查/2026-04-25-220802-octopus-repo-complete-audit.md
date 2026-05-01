# Octopus 完整审计报告

- 仓库：`D:\GPT-codex\octopus_repo`
- 触发时间：`2026-04-25T22:08:02+08:00`
- 审计范围：当前工作区、`HEAD`、`origin/dev`、README/docs、可运行验证链

## 1. Findings

### `Critical` 当前工作区的备份设置页已经回归成占位实现，且文件结构本身疑似损坏

证据：

- `web/src/components/modules/setting/Backup.tsx:472-475` 定义 `SettingBackup()` 后直接返回 `TODO` 占位。
- `web/src/components/modules/setting/Backup.tsx:478` 开始又出现游离 JSX，说明当前文件不是单纯“功能没写完”，而是处于截断/拼接后的坏状态。
- `web/src/components/modules/setting/index.tsx:19,98-99` 仍在真实设置页中懒加载 `SettingBackup`，问题处于主路径。
- `web/src/components/modules/setting/Backup.test.tsx:362,491,602` 等测试仍在断言完整备份页节点，但当前组件根本无法满足这些断言。
- `git diff -- web/src/components/modules/setting/Backup.tsx` 显示这是当前工作区相对 `HEAD` 的未提交回归，不是已提交基线问题。

判断：这是直接的发布阻塞项。当前工作区如果进入构建/发布，设置页备份区域高概率不可用，甚至可能直接导致前端编译失败。

### `High` `config_source_mode / active_ai_profile_id` 已有管理面与设置项，但没有接入真实运行时选择逻辑

证据：

- `internal/op/ai_automation.go:20,29` 读取并返回 `active_ai_profile_id` / `config_source_mode`。
- `internal/op/ai_automation.go:431-442` 的 `AIProfileActivate()` 只是在 setting 表里写入 `ai_profile + profile id`。
- `internal/op/ai_automation_executor.go:439-440` 只是把这两个值放进 AI 任务上下文 payload。
- `web/src/components/modules/setting/AIAutomationSource.tsx:20-31` 已经给用户暴露了“手动 / AI Profile”切换入口。
- 对比可见，真正进入 relay 运行时的只有动态路由学习开关：`internal/relay/dynamic_mode.go:278-306` 会把学习状态加进候选评分。

判断：这是“看起来实现了但主流程未接入”的典型问题。当前 `config_source_mode / active_ai_profile_id` 更像管理面元数据，而不是已经闭环的 runtime selector。

### `High` AI 自动化主线没有通过自己的目标测试，远端模型发现和异步任务执行都不稳定

证据：

- `internal/op/ai_automation_models.go:18-41` 明确先走远端模型发现，再回退本地缓存。
- `internal/op/ai_automation_test.go:21` 的 `TestAIAutomationFetchModelsPrefersRemoteFreeCandidate` 在实际运行中失败，断言点在 `internal/op/ai_automation_test.go:60`，实际结果落回 `local_model_cache`。
- `internal/op/ai_automation.go:276-317` 创建任务后立即 `AITaskStartAsync(task.ID)`。
- `internal/op/ai_automation_executor.go:125-136` 使用 goroutine + 超时上下文异步执行任务。
- `internal/op/ai_automation_test.go:174,274,393` 的多个任务测试都停在 `pending`/`running`，没有按预期完成。

判断：AI 自动化不是空实现，但当前完成度仍处于“功能已铺开、正确性未收口”的阶段，而且已经有真实回归证据。

### `High` 备份/导入/回滚后端测试大面积失败，根因集中在测试基线被启动初始化逻辑污染

证据链：

- `internal/op/op_test.go:13-28` 的 `setupOpTestDB()` 每次都会设置 bootstrap admin 环境变量，然后调用 `InitCache()`。
- `internal/op/cache.go:9-17` 的 `InitCache()` 会先跑 `UserInit()`，再刷新 settings/channel/group 等缓存。
- `internal/op/user.go:86-100` 的 `UserInit()` 在空库时会自动创建首个管理员。
- 因此大量备份测试的“空库 / 仅存在测试自己创建的数据”的假设被破坏，实际红点也符合这个模式：
  - `internal/op/backup_test.go:20` 的 `TestDBExportAllIncludesSecretsByDefault` 实际运行时导出了两个用户，期望只看到 `backup-admin`。
  - `internal/op/backup_test.go:87` 的 `TestDBExportAllExcludesInternalAuthTokenSecret` 在创建同 key setting 时撞到 `UNIQUE constraint failed: settings.key`。
  - `internal/op/backup_test.go:2024` 的回滚测试在 `UserVerify(before import)` 就失败。
  - `internal/op/backup_test.go:2223` 的回滚预览统计期望 `users=1`，但初始化默认管理员会把这个计数抬高。

判断：当前 `backup` 主线不是“没有测试”，而是测试基线本身已经被新的启动初始化语义改坏，导致验证体系失真。这也是高优先级修复项。

### `Medium` 文档状态与当前实现明显不一致，接手判断会被误导

证据：

- `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:47-49` 仍写着“AI 自动化中心主线状态：文档规划已立项，代码实现未开始”。
- 但当前工作区实际上已经存在 AI 自动化后端路由 `internal/server/handlers/ai_automation.go:19-31`，也有顶层前端入口 `web/src/route/config.tsx:21-30`。

判断：这不是普通文档过期，而是会直接误导后续接手和审计判断的状态漂移。

### `Medium` 本机可运行验证链存在真实环境阻塞，导致部分承诺的构建/烟测无法在当前宿主机复现

证据：

- `node -v` 可以运行，但一旦执行脚本就触发 `Assertion failed: ncrypto::CSPRNG(nullptr, 0)`，所以 `pnpm` / `corepack` / `tsc` / repo-local Node 验证脚本都无法启动。
- `go build ./...` 在补充本地 `GOCACHE` 后仍因 DNS 无法解析 `proxy.golang.org` 失败，缺失依赖拉取证据。
- `docker --version` 不存在，当前宿主机没有 Docker。
- `bash --version` 直接报 `couldn't create signal pipe`，仓库里的 Linux shell smoke 不能在本机直接跑。

判断：这不是代码缺陷本身，但它让 README / CI / worklog 中承诺的本地验证路径，在这台机器上无法作为当前轮的可用证据链。

## 2. Completion Assessment

### 完成度评估

- 当前工作区整体完成度：约 `70%`。
- 已完成：基础后端启动链、管理/relay 路由注册、动态路由学习信号接入、AI 自动化中心的模型/任务/Profile 数据结构与接口、healthcheck CLI、设置页 AI 入口、validation workflow 骨架。
- 部分完成：AI 自动化端到端正确性、AI Profile 双轨配置、备份/导入/回滚全链路、前端设置页主线、CI 所承诺的本地等价验证路径。
- 未完成：`config_source_mode / active_ai_profile_id` 的真实运行时消费闭环、当前工作区下的备份页前端实现收口、当前主机上的完整 smoke 复现。
- 疑似空实现或伪接线：`config_source_mode / active_ai_profile_id` 当前只影响设置和任务上下文，不影响真实 relay 选择行为。

### 已完成 / 部分完成 / 未完成 / 疑似空实现

- 已完成：`dynamic_routing_learning_enabled` 的运行时接线，证据见 `internal/relay/dynamic_mode.go:278-306`。
- 已完成：AI 自动化接口注册和顶层导航接入，证据见 `internal/server/handlers/ai_automation.go:19-31`、`web/src/route/config.tsx:21-30`。
- 部分完成：AI 自动化任务执行、远端模型发现、Profile 保存与激活，已有代码但测试不绿。
- 部分完成：备份/导入/回滚后端能力，已有大规模实现和测试，但当前测试基线失真、前端工作区回归。
- 未完成：当前工作区备份页主实现收口。
- 疑似空实现：`config_source_mode / active_ai_profile_id` runtime consumer。

## 3. Verification Summary

### 已验证项

- 仓库状态：`git status --short --branch`、`git diff --stat`、`git diff --name-only`、`git log --oneline --decorate -n 12`。
- 分支对比：`git merge-base HEAD origin/dev`、`git log origin/dev..HEAD`、`git diff origin/dev...HEAD --stat`。
- 核心入口与接线：`main.go`、`cmd/start.go`、`internal/server/server.go`、`internal/server/router/router.go`、`internal/task/init.go`、`internal/relay/relay.go`。
- 关键新能力接线：`internal/server/handlers/ai_automation.go`、`internal/server/handlers/dynamic_routing.go`、`internal/op/ai_automation*.go`、`internal/op/dynamic_route_learning.go`、`internal/relay/dynamic_mode.go`、`web/src/route/config.tsx`、`web/src/components/modules/ai-automation/index.tsx`、`web/src/components/modules/setting/AIAutomationSource.tsx`。
- 当前工作区关键回归：`git diff -- web/src/components/modules/setting/Backup.tsx`、静态阅读 `Backup.tsx`。
- 可运行验证：
  - `go version` 通过。
  - `node -v` 通过。
  - `go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1` 部分执行成功，其中 `internal/server/middleware` 与 `internal/relay/balancer` 通过。
  - `internal/op` 真实跑出了多条失败断言。

### 未验证项

- 前端 `tsc --noEmit`、`build:static`、repo-local Node 验证脚本：被当前宿主机 Node `CSPRNG` 启动错误阻塞。
- `go build ./...` 全量构建：被当前宿主机 DNS 无法解析 `proxy.golang.org` 阻塞。
- `scripts/smoke-linux-backend.sh`、`scripts/smoke-docker-compose.sh`：本机缺少可用 `bash` / `docker`，未执行。
- 浏览器级 smoke / Docker 运行态 smoke：未验证。

### 运行过的关键命令

- `git status --short --branch`
- `git diff --stat`
- `git log --oneline --decorate -n 12`
- `git merge-base HEAD origin/dev`
- `go version`
- `node -v`
- `go build ./...`
- `go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1`
- 多轮 `Get-Content` / `Select-String` / `rg` 对核心代码、测试和文档进行交叉审查

## 4. Comparison Notes

### 当前工作区 vs `HEAD`

- 当前工作区非常脏：`133` 个 tracked 文件修改，且有大量未跟踪文件。
- 一批关键能力尚未进入 `HEAD`，例如 `internal/op/ai_automation.go`、`internal/op/ai_automation_executor.go`、`internal/relay/dynamic_mode.go` 在 `git show HEAD:...` 下直接不存在，说明它们仍停留在工作区增量中。
- `Backup.tsx` 的坏状态属于当前工作区问题，不是已提交基线问题。

### 当前分支 vs 主分支 / 稳定基线

- `git merge-base HEAD origin/dev` 返回的就是 `origin/dev` 当前提交，说明 `origin/dev` 是这次最接近的稳定基线。
- `HEAD` 相比 `origin/dev` 已经引入 Copilot、Antigravity、providers、healthcheck CLI 等能力；但本轮真正风险更高的是“工作区相对 `HEAD` 的未提交扩展和回归”。

### 代码实现 vs README / docs / 任务说明

- README 当前 `build:static -> static/out` 的源码构建口径已经和仓库实现对齐，这一项本轮没有发现偏差。
- 但 `docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:49` 仍声称 AI 自动化“代码实现未开始”，与当前工作区实际实现严重不一致。

### 代码实现 vs 测试 / 构建 / 验证覆盖

- `.github/workflows/validation.yaml:48,55,58` 仍要求跑 `pnpm exec tsc --noEmit`、`pnpm run build:static`、`go build ./...`、`go test ./internal/op ...`。
- 当前宿主机上，Node 脚本链和 Docker/bash smoke 无法复现；而已能运行的 `internal/op` 测试又出现了真实红点，说明“有验证配置”不等于“当前实现已被验证通过”。
- 备份页前端测试仍假定完整组件存在，但当前工作区 `Backup.tsx` 已明显不满足这个前提。

## 5. Top Next Actions

### 需要优先处理的前三项

1. 修复 `web/src/components/modules/setting/Backup.tsx` 的当前工作区回归，恢复为可编译、非占位、可渲染的真实备份页实现，然后立即重跑 `Backup.test.tsx` 与相关 no-browser 验证。
2. 决定 `config_source_mode / active_ai_profile_id` 的产品语义。
   - 如果它们应该改变真实运行时来源，就把消费点接进 relay/channel/group/model 主路径并补测试。
   - 如果它们不应该影响运行时，就收窄 UI 和文档表述，避免制造“切换已生效”的假象。
3. 统一修复 AI 自动化与备份后端测试。
   - 先修测试基线被 `InitCache()/UserInit()` 污染的问题。
   - 再修 AI 自动化远端模型发现与异步任务执行稳定性。
   - 最后在健康宿主机或 CI 上补齐 `go build`、`go test`、`tsc`、`build:static`、smoke 证据。

### 建议下一步动作

- 先把当前工作区当作“未收口候选版本”处理，不要直接以此发布。
- 以 `origin/dev` 作为稳定基线、以 `HEAD` 作为当前分支已提交基线，先清理工作区中尚未稳定的主路径回归，再继续扩展功能。
- 修完 `Backup.tsx` 和后端测试基线后，再重新跑一次完整审计，确认当前工作区是否重新回到“可验证状态”。

## 中文摘要

1. 本次触发时间
   - `2026-04-25T22:08:02+08:00`
2. 做了哪些检查、运行了哪些命令
   - 检查了仓库结构、`git status`、`git diff --stat`、最近提交、`origin/dev` 基线、核心 README/docs、主入口、路由注册、AI 自动化、动态路由、备份/导入/回滚、设置页与测试文件。
   - 运行了 `go version`、`node -v`、`go build ./...`、`go test ./internal/op ./internal/server/handlers ./internal/server/middleware ./internal/relay/... ./internal/task -count=1`。
   - 尝试了前端 `pnpm/corepack/node+tsc` 入口，但被当前宿主机 Node 启动错误阻塞。
3. 修改了哪些文件
   - 新增 `docs/review/审查/2026-04-25-220802-octopus-repo-complete-audit.md`
   - 新增 `docs/review/审查/2026-04-25-220802-octopus-repo-complete-audit.html`
4. 发现了什么问题
   - 当前工作区 `Backup.tsx` 已回归成 `TODO` 占位并伴随游离 JSX，属于主路径阻塞。
   - `config_source_mode / active_ai_profile_id` 已有 UI 和 setting，但没有接进真实运行时消费链。
   - AI 自动化目标测试失败，远端模型发现和异步任务执行不稳定。
   - 备份/导入/回滚测试大量失败，根因集中在 `InitCache()/UserInit()` 改变了测试初始库假设。
   - 当前宿主机 Node/Docker/bash/Go 网络环境阻塞了部分验证链。
5. 本次结果是成功、跳过还是失败
   - `成功`。审计任务本身已完成，但代码与环境层面存在阻塞项和未验证项。
6. 是否需要我手动介入
   - `需要`。至少需要人工确认并优先处理 `Backup.tsx` 回归、AI Profile 切换语义、以及后端测试基线修复策略。
