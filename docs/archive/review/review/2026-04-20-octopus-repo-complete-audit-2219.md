# Octopus 仓库完整审计

- 审计时间：2026-04-20 22:19:57 +08:00
- 审计范围：`D:\GPT-codex\octopus_repo`
- 参考基线：当前工作区 vs `HEAD`，当前分支 `feat/erguotou` vs `origin/dev`
- 复用本地资源：`AGENTS.md`、`README.md`、`README_zh.md`、`docs/CURRENT_STATUS_AND_PLAN.zh-CN.md`、`docs/worklog/`、`docs/review/`、`scripts/`、`web/package.json`

## 1. Findings

### High

1. 前端主线仍可能被“陈旧静态产物”掩盖，当前 backend smoke 不能证明最新 `web/src` 已真实接入发布链路。
   - 证据一：`scripts/smoke-linux-backend.sh:101-121` 固定把 `static/out` 写入运行配置，随后只校验服务端返回的首页和 `manifest.json`。
   - 证据二：`static/static.go:8` 直接 embed `out` 目录；`cmd/start.go:42-49` 之后的服务启动默认消费这条静态资源链路。
   - 证据三：当前工作区里 `web/out` 的时间比 `static/out` 新，本地 smoke 若直接读 `static/out`，就有可能在 `web/src` 已变化但 `static/out` 未同步时仍然通过。
   - 影响：会把“当前源码可发布”误判成“上一次导出的静态产物可服务”。这对最近大量变更的设置页、备份页和日志页尤其危险。

2. 最近改动最重的 Phase F 备份/导入/回滚前端流程，缺少行为级自动化验证；现有验证只能证明类型正确，不能证明关键交互正确。
   - 证据一：`web/package.json:5-10` 只有 `dev`、`build`、`start`、`lint`，没有 `test` 脚本。
   - 证据二：本轮 `rg --files web | rg "(test|spec)\.(ts|tsx|js|jsx)$"` 未找到任何前端测试文件。
   - 证据三：`web/src/components/modules/setting/Backup.tsx` 当前文件体量约 `113665` 字节，承担 export、dry-run、apply-ready、rollback preview、post-import validation 等关键交互，但没有对应前端测试。
   - 证据四：本轮 `D:\gol1\node.exe .\node_modules\typescript\bin\tsc --noEmit` 通过，但 `D:\gol1\node.exe .\node_modules\next\dist\bin\next build` 在编译完成后仍以 `spawn EPERM` 结束，无法在当前环境闭环验证静态导出产物。
   - 影响：最关键的恢复链路现在主要靠后端测试和人工假设兜底，无法证明按钮状态、预览态切换、导出下载、Apply Same Import、回滚预览卡片等真实可用。

### Medium

3. 动态路由“daily summary scan”已经注册定时任务，也在设置页文案中宣称服务于 summary surface，但当前代码里没有真实消费方，属于已实现但未接入主流程的功能段。
   - 证据一：`internal/task/init.go:40-41` 明确把 `dynamic_routing_summary_scan` 注册为“used by the settings/dashboard surface”。
   - 证据二：`internal/task/dynamic.go:15-119` 只在内存里维护 `DynamicRoutingSummaryScanSummary`，并暴露 `GetDynamicRoutingSummaryScanSummary()`。
   - 证据三：本轮全文检索只发现 `GetDynamicRoutingSummaryScanSummary()` 在测试中被调用，没有 handler、API endpoint、stats 聚合或前端查询真正读取这份 summary。
   - 证据四：`web/src/components/modules/setting/DynamicRouting.tsx:162-164` 仍向用户说明“Daily summary scan only updates the dynamic routing health summary surface”，但页面本身没有展示对应 summary 数据。
   - 影响：这不是纯空壳，但属于明显的“后台有任务，前台无消费”。当前只能算部分完成，不能算完整闭环。

4. 当前文档口径存在分裂，仓库级说明里还保留了已失效的环境变量名，容易让后续开发或自动化把配置写错位置。
   - 证据一：`AGENTS.md:238-239` 仍写 `OCTOPUS_DB_TYPE` / `OCTOPUS_DB_CONNSTRING`。
   - 证据二：`README.md:277-278` 与 `README_zh.md:265-266` 已切到 `OCTOPUS_DATABASE_TYPE` / `OCTOPUS_DATABASE_PATH`，且 `internal/conf/config.go:39-43` 使用的是 `OCTOPUS_` 前缀加字段映射，不存在 `DB_CONNSTRING` 这个运行时字段。
   - 影响：这是低层配置文档漂移，不会直接击穿运行时，但会误导新环境搭建和脚本编写。

### Low

5. 本轮没有发现新的“看起来有 UI、实际上完全没后端”的强证据；但历史 worklog 和自动化记忆里有几条旧结论已经过时，如果继续沿用会污染判断。
   - 证据一：当前 `web/src/api/endpoints/setting.ts:441-447` 已把 `include_secrets` 默认值对齐为 `true`。
   - 证据二：当前 `internal/server/handlers/setting.go:111-118` 仍以“未传参即默认全量导出”工作，和 README 口径一致。
   - 证据三：当前 `scripts/smoke-win-backend.ps1:16-18` 语法正常，前一轮记忆里“dot-source 与 `$exePath` 粘连损坏”的结论已失效。
   - 影响：需要清理旧误报，避免把已修复项继续当成当前 blocker。

## 2. Completion Assessment

- 已完成
  - 后端主启动链路、数据库迁移、缓存初始化、HTTP 服务与静态资源回退链路可读且真实接线。
  - 备份导出/导入/回滚后端主链路已真实接入，且 `include_secrets` 默认语义与 README 一致。
  - route-target override 后端接口、缓存刷新与前端编辑入口都已真实接线，不属于伪实现。
  - 日志流、日志导出、token breakdown、healthcheck CLI 都已落入真实代码路径。

- 部分完成
  - 动态路由的 relay 调优、race budget、circuit threshold 真正被消费。
  - 但“daily summary scan -> settings/dashboard summary surface”仍未形成闭环，只能算部分完成。
  - 备份/导入/回滚前端界面结构和后端协议对齐程度较高，但验证闭环不足，只能算部分完成。

- 未完成
  - 前端关键恢复流程的自动化测试覆盖基本为空。
  - 当前工作区没有证据证明最新 `web/src` 一定被同步到最终提供服务的 `static/out` / embed 产物。

- 疑似空实现或未接入
  - `dynamic_routing_summary_scan` 对应的 summary surface。

- 完成度评估
  - 结合当前工作区与已验证主链路，我给当前仓库实现完成度约 `82%`。
  - 后端主功能和大部分接线已经成型。
  - 主要短板在前端关键流程验证闭环，以及一小段动态路由摘要功能没有真正被消费。

## 3. Verification Summary

- 已验证项
  - `git status --short --branch`
  - `git branch -a --verbose --no-abbrev`
  - `git log --oneline --decorate -n 15`
  - `git log --oneline origin/dev..HEAD`
  - `git diff --stat HEAD`
  - 多轮 `rg` / `Get-Content` 交叉核查 `README`、`docs`、`handlers`、`relay`、`task`、`scripts`、`web` 接线关系
  - `D:\gol1\node.exe .\node_modules\typescript\bin\tsc --noEmit`，结果：通过
  - `D:\gol1\node.exe .\node_modules\next\dist\bin\next build`，结果：编译完成后失败，错误为 `spawn EPERM`

- 未验证项
  - `go build ./...` 与 Go 测试矩阵
  - Windows 本地 smoke
  - Docker compose runtime smoke
  - 浏览器级手工验证：backup export、dry-run import、Apply Same Import、rollback preview、post-import validation

- 无法验证的具体原因
  - `C:\Users\李昊桐\AppData\Local\Programs\go\bin\go.exe version` 直接返回 `Access is denied`，导致 Go 构建、Go 测试、`scripts/use-go-env.ps1`、`scripts/smoke-win-backend.ps1` 都无法继续。
  - 当前环境 `docker` 命令不可用，无法本地执行 `scripts/smoke-docker-compose.sh`。
  - `next build` 在当前 Windows 桌面环境结束于 `spawn EPERM`，因此未能在本机完成静态导出闭环验证。

## 4. Comparison Notes

- 当前工作区 vs `HEAD`
  - 当前工作区是大面积 WIP 状态。
  - `git diff --stat HEAD` 显示 81 个已跟踪文件改动，约 `9518` 行新增、`1191` 行删除，另有大量新增测试、脚本、文档和审查文件未提交。
  - 最近未提交改动主要集中在 `backup/import/rollback`、`dynamic routing`、`route target override`、`log export`、`frontend setting UI`、`scripts` 与 `validation` 工作流。

- 当前分支 vs 最近稳定基线
  - 当前分支是 `feat/erguotou`，远端可比稳定基线取 `origin/dev`。
  - `git log origin/dev..HEAD` 显示当前分支已经长期偏离 `origin/dev`，包含 Copilot、Antigravity、provider API、healthcheck、文案与仓库引用等历史增量。
  - `git log HEAD..origin/dev` 本轮为空，说明远端 `origin/dev` 相对当前本地 `HEAD` 没有额外待并入提交。

- 代码实现 vs README / docs / 任务说明
  - 备份导出默认语义现在和 README 对齐，不应再沿用旧报告中“默认导出被前端强制脱敏”的结论。
  - 动态路由 summary surface 的说法在代码层并未闭环，README/页面文案比真实接线更靠前。
  - `README` 已切换到 `OCTOPUS_DATABASE_TYPE` / `OCTOPUS_DATABASE_PATH`，但 `AGENTS.md` 还停留在旧变量名。

- 代码实现 vs 测试 / 构建 / 验证覆盖
  - 后端测试面比上一轮记忆更完整，`validation.yaml` 已覆盖 `internal/op`、`internal/server/handlers`、`internal/server/middleware`、`internal/relay/...`、`internal/task`。
  - 但前端仍只有类型检查，没有行为级测试。
  - backend smoke 消费 `static/out`，不能单独证明当前 `web/src` 已同步到最终服务产物。
  - Docker 运行时 smoke 理论上能覆盖 Dockerfile 中的前端 build，但本轮本地环境无法执行 `docker` 验证。

## 5. Top Next Actions

1. 先收口前端发布验证链路。
   - 要求：让 CI 和本地 smoke 至少有一条路径强制从当前 `web/src` 重新产出并校验 `static/out`，不要只消费工作区中已有的静态目录。

2. 给 Phase F 备份/导入/回滚前端补最小可执行验证。
   - 要求：至少覆盖 export 默认语义、dry-run 报告、Apply Same Import、rollback preview、post-import validation 五条关键交互。

3. 要么真正接入动态路由 summary surface，要么把文案降级成“仅后台汇总”。
   - 要求：不要继续保留“settings/dashboard surface 已接入”的描述，除非补上真实 handler 和前端展示。

### 本次中文摘要

1. 本次触发时间
   - 2026-04-20 22:19:57 +08:00

2. 做了哪些检查、运行了哪些命令
   - 读取了自动化记忆、仓库结构、`git status`、`git branch`、`git log`、`git diff --stat HEAD`
   - 核查了 `README` / `README_zh` / `docs` / `handlers` / `relay` / `task` / `scripts` / `web` 的关键接线
   - 运行了前端类型检查 `D:\gol1\node.exe .\node_modules\typescript\bin\tsc --noEmit`
   - 运行了前端构建 `D:\gol1\node.exe .\node_modules\next\dist\bin\next build`
   - 尝试运行 `go.exe version`、`scripts/smoke-win-backend.ps1`、`docker version`

3. 修改了哪些文件
   - 新增 `docs/review/2026-04-20-octopus-repo-complete-audit-2219.md`
   - 新增 `docs/review/2026-04-20-octopus-repo-complete-audit-2219.html`
   - 更新 `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`

4. 发现了什么问题
   - backend smoke 可能读取陈旧 `static/out`，不能单独证明最新 `web/src` 已进发布链路
   - 备份/导入/回滚前端没有行为级自动化验证
   - 动态路由 summary scan 已注册但没有真实消费方
   - `AGENTS.md` 里的环境变量示例已过时

5. 本次结果是成功、跳过还是失败
   - 成功
   - 审计报告已生成，但部分构建/测试因本机环境限制未能闭环

6. 是否需要我手动介入
   - 需要
   - 需要你提供一个可执行的 Go 工具链环境，或在 CI / 另一台可运行 `go.exe` 与 `docker` 的机器上补跑剩余验证
