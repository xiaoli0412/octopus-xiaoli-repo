# Octopus 完整审计报告

审计时间：2026-04-27 13:13:55 +08:00  
仓库：`D:\GPT-codex\octopus_repo`

## 1. Findings

### Critical

1. 前端 no-browser 验证链在当前仓库里是确定性断开的，`validation` / `release` 工作流会调用一个默认无法在 Node 22 下运行的脚本入口。

判断依据：

- [web/package.json](/D:/GPT-codex/octopus_repo/web/package.json:21) 把 `test:setting-info-logic` 写成 `node ../scripts/verify-setting-info-logic.mjs`。
- [web/package.json](/D:/GPT-codex/octopus_repo/web/package.json:22) 和 [web/package.json](/D:/GPT-codex/octopus_repo/web/package.json:23) 又把这个入口串进了 `test:settings-no-browser` 与 `test:screenshot-no-browser`。
- [.github/workflows/validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:46) 与 [.github/workflows/validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:52) 在 CI 中直接跑 `pnpm run test:screenshot-no-browser`。
- [.github/workflows/release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:44) 与 [.github/workflows/release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:50) 也复用了同一链路。
- [scripts/verify-setting-info-logic.mjs](/D:/GPT-codex/octopus_repo/scripts/verify-setting-info-logic.mjs:1) 会动态 `import` [web/src/components/modules/setting/info-logic.ts](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/info-logic.ts:1)。
- 本轮直接用正常的 Node 22 复核：`node scripts/verify-setting-info-logic.mjs` 报 `ERR_UNKNOWN_FILE_EXTENSION: ".ts"`；改成 `node --experimental-strip-types ...` 后通过，输出 `setting-info-logic verification passed`。

影响：

- 这不是“某个边角测试没覆盖”，而是 release gate 里的一条正式验证命令本身就会失败。
- 当前代码与 CI 配置不一致，属于明确的发布阻塞项。

### High

2. AI 自动化里的“受保护动作”是典型的“界面与元数据已接线，但主执行流并未真正接入”的伪实现。

判断依据：

- [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:438) 暴露了 `profile_activate`，文案写的是“允许切换 Profile”。
- [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:439) 同时暴露了 `snapshot_guard`，文案写的是“创建配置快照”。
- 多个 preset 和 one-click 模板把这两个 tool key 当作可授权动作注入任务配置，例如 [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:362) 与 [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:589)。
- 结果区还会展示“已授权 N 项，已实际执行 M 项”，见 [web/src/components/modules/ai-automation/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/ai-automation/index.tsx:1676)。
- 但执行器里，[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:579) 对 `snapshot_guard` 固定写入 `executed: false`，reason 是 `ui_managed_snapshot_only`。
- 同一个执行器里，[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:583) 对 `profile_activate` 也固定写入 `executed: false`，reason 是 `manual_activation_required`。
- 汇总字段更直接：[internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:659) 把 `protected_actions_executed` 硬编码成 `0`。
- 仓库里确实存在手动激活能力，[internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:18) 注册了 `/profiles/:id/activate`，而 [internal/op/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation.go:441) 也实现了 `AIProfileActivate`。问题不在“完全没有功能”，而在“AI task 执行链并不会调用它”。

影响：

- 这类实现最容易误导后续维护者和使用者，以为 AI 自动化已经具备“授权后可执行保护动作”的闭环。
- 从审计目标看，它应被归类为“看起来实现了，但主流程未接入”的高风险不一致。

### Medium

3. 状态文档与当前实现明显漂移，尤其 AI 自动化主线仍被写成“代码实现未开始”，这会直接误导后续交接与验收判断。

判断依据：

- [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:47) 到 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:56) 明确写着“当前状态：文档规划已立项，代码实现未开始”。
- 但真实代码里：
- [web/src/route/config.tsx](/D:/GPT-codex/octopus_repo/web/src/route/config.tsx:21) 已有顶层 `AI Automation` 路由。
- [internal/server/handlers/ai_automation.go](/D:/GPT-codex/octopus_repo/internal/server/handlers/ai_automation.go:18) 已注册一整组 `/api/v1/ai/*` 路由。
- [web/src/components/modules/setting/AIAutomationSource.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/AIAutomationSource.tsx:14) 已有 `manual / ai_profile` 切换卡片。
- [web/src/components/modules/setting/index.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/setting/index.tsx:87) 已把该卡片接入设置页。

影响：

- 这不是普通 README 过时，而是会影响“完成度评估”“验收边界”和后续继续施工方向的状态文档失真。

4. `scripts/verify-go-env.ps1 -GoFmt` 当前会把仓库内的本地 Go 工具链一起扫进去，导致格式检查结果混入与业务代码无关的 `.tools` 文件。

判断依据：

- [scripts/verify-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:98) 到 [scripts/verify-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:115) 使用 `Get-ChildItem -LiteralPath $repoRoot -Recurse -Filter '*.go'` 全仓递归枚举。
- 同一个仓库下确实存在大量本地工具链源码，例如 [D:\GPT-codex\octopus_repo\.tools\go\go\src\archive\tar\common.go](/D:/GPT-codex/octopus_repo/.tools/go/go/src/archive/tar/common.go:1) 以及众多 `.tools\go\go\src\...\testdata` 目录。

影响：

- 这会让本地格式验证出现噪音，难以区分“仓库代码需要 gofmt”还是“工具链镜像目录被误扫”。
- 它不直接影响运行时，但会污染验证结论。

### Low

5. 当前工作区仍有可直接复现的 diff hygiene 问题。

判断依据：

- `git diff --check` 本轮稳定报出 [web/src/components/modules/channel/Form.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/channel/Form.tsx:1380) 的 trailing whitespace。
- 同一命令还报出 [web/src/components/modules/channel/Form.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/channel/Form.tsx:2289) 的 `new blank line at EOF`。

影响：

- 不会影响功能，但会增加 review 噪音，并持续让质量门禁看起来处于脏状态。

## 2. Completion Assessment

总体判断：当前仓库已经不是“骨架阶段”，而是“主流程大多已落地、但验证链和若干 AI 自动化子能力仍未收口”的状态。

完成度评估：

- 已完成：启动/配置/relay 主链路、管理端主体、AI 自动化顶层入口、AI 配置接口、AI Profile 手动激活链路、动态路由学习开关与摘要链路、备份导入导出后端主能力、设置页大部分新卡片接线。
- 部分完成：AI 自动化执行器中的受保护动作、前端 no-browser 验证链、状态文档同步、repo-local 验证脚本的一致性。
- 未完成：当前没有证据证明 AI task 能在“授权后自动激活 profile / 创建快照”；完整前端 CI 链也尚未恢复到可信状态。
- 疑似空实现：AI 自动化中的 `profile_activate` / `snapshot_guard` 执行语义。

我给当前工作区的综合完成度判断是：约 80% 到 85%。核心能力基本齐了，真正拖后腿的是“验证闭环”和“个别看似已做完的 AI 自动化子能力”。

## 3. Verification Summary

已验证项：

- `git status --short --branch`、`git log --oneline --decorate -10`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD` 已完成。
- `go build ./...` 已通过，使用仓库自带脚本 [scripts/verify-go-env.ps1](/D:/GPT-codex/octopus_repo/scripts/verify-go-env.ps1:1) 和 repo-local Go 1.24.4。
- `go test ./...` 已重跑；其中 `internal/op`、`internal/server/middleware`、`internal/task`、`internal/server`、`cmd`、`internal/transformer/...`、`internal/client`、`internal/helper`、`internal/model`、`internal/price`、`internal/update` 通过。
- 使用旁路 Node 22 复核了 `verify-setting-info-logic`：
- `node scripts/verify-setting-info-logic.mjs` 失败，报 `ERR_UNKNOWN_FILE_EXTENSION`。
- `node --experimental-strip-types scripts/verify-setting-info-logic.mjs` 通过，输出 `setting-info-logic verification passed`。
- `git diff --check` 已运行，确认当前只剩 `Form.tsx` 的两处 hygiene 问题。

未验证项：

- `pnpm run test:screenshot-no-browser` 没能在当前宿主完整跑完；一方面 Codex 自带 `node.exe` 会触发 `CSPRNG` 断言，另一方面改走旁路 Node 后 `corepack pnpm` 在本机仍未在超时时间内稳定收敛。
- `pnpm exec tsc --noEmit` 与 `vitest` 没有通过仓库里的 `.cmd` 入口稳定复跑，当前环境报“找不到指定的模块”，更像本机 Node/pnpm 包装链问题，而不是已定位的仓库逻辑问题。
- `go test ./...` 中 `internal/relay` 与 `internal/server/handlers` 的一批用例依旧受本机 `net.Listen(tcp4)` 阻塞，统一报 `socket: The requested service provider could not be loaded or initialized`；这更像宿主网络栈限制，而不是本次可直接归因到代码回归的仓库缺陷。
- `next build` / `build:static` 本轮未重跑到底，因为当前最关键的前端阻塞已由 `verify-setting-info-logic` 入口问题坐实，不需要继续让宿主 Node 工具链噪音稀释结论。

## 4. Comparison Notes

当前工作区与 `HEAD`：

- 当前工作区相对 `HEAD` 仍是超大脏树，`git status` 显示 136 个已修改文件，加上一批新增测试、脚本和文档。
- 这说明本次审计必须同时区分“分支已提交能力”和“工作区待验收能力”。

当前分支与稳定基线：

- 当前分支是 `feat/erguotou`，`HEAD` 为 `bfa27ae`。
- 相对 [origin/dev](/D:/GPT-codex/octopus_repo/.git/refs/remotes/origin/dev)，`HEAD` 本身已有 72 个文件、约 `6001 insertions / 1741 deletions` 的已提交增量，说明当前分支已明显偏离上游稳定基线，并吸收了 Copilot、Antigravity、providers、DocModal 等扩展能力。

代码与 README / docs / 任务说明：

- README 与主实现方向基本一致，AI 自动化、动态路由、备份导入导出、providers、CC Switch 等主线在代码中都能找到真实接线。
- 最明显的不一致在状态文档，尤其是 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:49) 对 AI 自动化状态的描述已经落后于代码。

代码与测试 / 构建 / 验证覆盖：

- 后端构建健康度比前端验证链更好：`go build ./...` 通过，非 socket 依赖的 Go 包测试也大体通过。
- 前端 no-browser 验证链存在明确入口错误，导致测试表面存在但关键命令本身不可信。
- AI 自动化的 profile 手动激活链真实存在，但“受保护动作自动执行”没有接进 task executor，这是典型的“代码表面存在、行为闭环缺失”。

## 5. Top Next Actions

需要优先处理的前三项：

1. 修复 [web/package.json](/D:/GPT-codex/octopus_repo/web/package.json:21) 到 [web/package.json](/D:/GPT-codex/octopus_repo/web/package.json:23) 的 `verify-setting-info-logic` 入口，统一补上 `--experimental-strip-types` 或把脚本改为可直接消费 `.js`，然后重跑 [.github/workflows/validation.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/validation.yaml:46) 与 [.github/workflows/release.yaml](/D:/GPT-codex/octopus_repo/.github/workflows/release.yaml:44) 对应链路。
2. 对 AI 自动化受保护动作做一次明确决策：要么在 [internal/op/ai_automation_executor.go](/D:/GPT-codex/octopus_repo/internal/op/ai_automation_executor.go:579) 接入真正的 gated execution，要么下调 UI 与文档表述，明确它们只是“授权声明 / 手动后续动作”，避免继续伪装成已执行能力。
3. 更新 [docs/CURRENT_STATUS_AND_PLAN.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md:47) 并收敛 repo-local 校验脚本，让 `GoFmt` 排除 `.tools`，同时清掉 [web/src/components/modules/channel/Form.tsx](/D:/GPT-codex/octopus_repo/web/src/components/modules/channel/Form.tsx:1380) 的 whitespace 噪音。

建议下一步动作：

- 先修 no-browser 验证入口，因为这是最硬的 release blocker。
- 修完后在一台网络栈正常的宿主上补跑 `pnpm run test:screenshot-no-browser`、`pnpm exec tsc --noEmit`、`go test ./internal/relay ./internal/server/handlers -count=1`。
- 再决定 AI 自动化受保护动作到底是要做成真执行，还是改成 preview-only 设计。

## 中文摘要

1. 本次触发时间：2026-04-27 13:13:55 +08:00。
2. 做了哪些检查、运行了哪些命令：检查了仓库结构、`git status`、`git log`、`git diff --stat HEAD`、`git diff --stat origin/dev...HEAD`、`git diff --check`；运行了 `scripts/verify-go-env.ps1 -GoBuild`、`scripts/verify-go-env.ps1 -GoTest`；直接用 Node 22 复核了 `verify-setting-info-logic.mjs` 在 plain `node` 与 `--experimental-strip-types` 下的差异；并通过 `rg` / 源码抽查核对了 AI 自动化、动态路由、设置页与文档状态。
3. 修改了哪些文件：新增 [docs/review/审查/2026-04-27-131355-octopus-repo-complete-audit.md](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-27-131355-octopus-repo-complete-audit.md:1)、[docs/review/审查/2026-04-27-131355-octopus-repo-complete-audit.html](/D:/GPT-codex/octopus_repo/docs/review/审查/2026-04-27-131355-octopus-repo-complete-audit.html:1) 和 [C:/Users/李昊桐/.codex/automations/octopus-repo/memory.md](</C:/Users/李昊桐/.codex/automations/octopus-repo/memory.md:1>)。
4. 发现了什么问题：前端 no-browser 验证链存在确定性断链，AI 自动化受保护动作存在“看起来已接线但执行流未接入”的伪实现，状态文档滞后于真实代码，`verify-go-env.ps1 -GoFmt` 会误扫 `.tools`，以及 `Form.tsx` 仍有 whitespace 噪音。
5. 本次结果是成功、跳过还是失败：成功。完整审计已落盘，部分验证项因宿主网络栈 / Node 包装环境限制未能完全跑通，但已明确区分环境问题与仓库问题。
6. 是否需要我手动介入：需要。优先手动介入 no-browser 验证入口修复，以及确认 AI 自动化受保护动作到底要真执行还是改为明确的 preview-only 语义。
