# Octopus 仓库完整审计

- 审计时间: 2026-04-21 00:46:04 +08:00
- 仓库: `D:\GPT-codex\octopus_repo`
- 当前分支: `feat/erguotou`
- 分支对比: 相对 `origin/dev` ahead `22` / behind `0`
- 工作区状态: `81` 个已跟踪文件改动，`2662` 个未跟踪文件

## 1. Findings

### Critical

1. `Critical` 验证链路会对过期前端产物给出假阳性通过，当前 UI 改动并没有被 runtime smoke 真正覆盖。
   - 证据
   - `internal/server/server.go:69-80` 启动时优先服务 `static/out`
   - `scripts/smoke-linux-backend.sh:101-121` 生成 smoke 配置时把 `static_dir` 固定写成 `${REPO_ROOT}/static/out`
   - `.github/workflows/validation.yaml:52-57` 在 Linux smoke 前只做前端 `tsc --noEmit`，没有执行前端 build，也没有把 `web/out` 同步到 `static/out`
   - 本次实际对比 `git diff --no-index --stat -- web/out static/out` 显示两者仍有 `54 files changed, 244 insertions(+), 586 deletions(-)` 的漂移
   - 影响
   - 当前工作区里像 `Backup.tsx`、`DynamicRouting.tsx` 这类前端改动，即使已经改坏或未完成，CI 和 smoke 也可能继续用旧的 `static/out` 通过
   - 这会直接削弱发布前最关键的回归验证，属于发布阻塞级问题

### High

2. `High` README 宣称的 Windows 构建入口当前是坏的，`scripts/build-win.ps1` 连 `-CheckOnly` 都跑不通。
   - 证据
   - `scripts/build-win.ps1:174` 把 dot-source 和变量赋值拼到了一行
   - 实际运行 `.\scripts\build-win.ps1 -CheckOnly` 直接报错 `The variable '$webDir' cannot be retrieved because it has not been set.`
   - 影响
   - 文档给出的官方 Windows 构建路径当前不可用

3. `High` Windows 开发启动脚本的尾部调用签名和函数定义不一致，主路径即使绕过 Go 环境问题也会在启动阶段失败。
   - 证据
   - `scripts/dev-win.ps1:92-105` 定义的 `Start-WindowProcess` 只接受 `Title`、`WorkingDirectory`、`Command`
   - `scripts/dev-win.ps1:229-235` 实际却以 `-FilePath`、`-Arguments`、`-Environment` 去调用它
   - 影响
   - 文档里的 Windows 本地开发模式不可信

### Medium

4. `Medium` 备份/导入主流程的后端已经真实接线，但前端关键行为仍缺少行为级验证，当前只能证明类型正确，不能证明交互正确。
   - 证据
   - 后端侧已经有较完整测试，覆盖导出默认语义、preview token、replace 预检、rollback preview、route-target override、relay race budget 等
   - `.github/workflows/validation.yaml:52-57` 对前端只做 `pnpm exec tsc --noEmit` 和 Linux smoke
   - `web/src/components/modules/setting/Backup.tsx` 是一个超大状态组件，但仓库里没有对应的浏览器级或组件行为级自动化验证

5. `Medium` `dynamic_routing_summary_scan` 有后端任务实现，但没有真实的 API 或前端消费面，当前更像半接入的后台摘要缓存。
   - 证据
   - `internal/task/dynamic.go:15-125` 定义了 summary 结构、每日任务和内存缓存读写
   - 搜索 `GetDynamicRoutingSummaryScanSummary` 只命中 `internal/task/dynamic.go` 与测试文件，没有命中 handler / API / frontend 消费
   - `web/src/components/modules/setting/DynamicRouting.tsx:158-164` 只有说明文案和预算配置，没有 summary 数据展示

### Low

6. `Low` 备份页自己明确声明还有未接入区域，当前完成度不应按“全部闭环”对外表述。
   - 证据
   - `web/src/components/modules/setting/Backup.tsx:2376-2380` 直接写着 `Still not wired to backend`
   - 列出的未完成项包括完整冲突处理向导、交互式 rollback history browser、rich side-by-side route diff tooling

## 2. Completion Assessment

### 完成度评估

- 仓库整体实现完成度: 约 `84%`
- 当前工作区可发布完成度: 约 `76%`
- 发布信心: `中低`

### 已完成

- 启动链路、`/healthz` 与 `octopus healthcheck` CLI 已真实接线
- 备份/导出/导入/rollback preview 的后端主链路已不是空壳
- route-target override 后端接口与 relay 策略继承已真实接线
- dynamic routing 预算设置已经真实进入 relay 预算获取路径
- handler / relay / backup / task 相关 Go 测试文件已经覆盖不少高价值逻辑

### 部分完成

- 设置页 `Backup` 前端闭环已经很深，但仍存在显式未接线区域
- dynamic routing 后台 summary scan 已实现，但 summary surface 未闭环
- Windows build/dev 脚本已被尝试补齐，但实际入口未闭环
- CI 工作流比上一轮更强，但仍没有锁住“前端产物必须与当前源码一致”这一前提

### 未完成或疑似空实现

- dynamic routing summary 的用户可见消费面
- 备份页的 richer conflict wizard / rollback history browser / rich route diff
- 前端关键设置流程的行为级自动化验证

## 3. Verification Summary

### 已验证项

- 已阅读与核对 `git status`、`git log`、`README.md`、`README_zh.md`、`.github/workflows/validation.yaml`、`cmd/start.go`、`cmd/healthcheck.go`、`internal/server/server.go`、`internal/server/handlers/setting.go`、`internal/server/handlers/route_target.go`、`internal/op/backup.go`、`internal/op/route_target.go`、`internal/op/healthcheck.go`、`internal/relay/dynamic.go`、`internal/relay/budget.go`、`internal/task/init.go`、`internal/task/dynamic.go`、`web/src/api/endpoints/setting.ts`、`web/src/components/modules/setting/Backup.tsx`、`web/src/components/modules/setting/DynamicRouting.tsx`、`scripts/build-win.ps1`、`scripts/dev-win.ps1`、`scripts/smoke-linux-backend.sh`、`scripts/smoke-win-backend.ps1`
- 已执行命令
  - `D:\gol1\node.exe .\node_modules\typescript\bin\tsc --noEmit` 通过
  - `D:\gol1\node.exe .\node_modules\next\dist\bin\next build` 编译通过，但最终失败于 `spawn EPERM`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\verify-go-env.ps1` 失败，原因是本机 Go 可执行文件被系统拒绝执行，报 `Access is denied`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1 -CheckOnly` 失败，原因是脚本自身变量拼接错误
  - `powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1 -CheckOnly` 失败，首先被本机 Go 权限阻塞
  - `git diff --no-index --stat -- web/out static/out` 证实前端导出目录与运行目录存在大范围漂移

### 未验证项

- `go test ./...` 未执行
  - 原因: `C:\Users\李昊桐\AppData\Local\Programs\go\bin\go.exe` 在本机返回 `Access is denied`
- `go build ./...` 未执行
  - 原因同上
- `scripts/smoke-win-backend.ps1` 未完整执行
  - 原因是本机 Go 工具链不可执行
- `scripts/smoke-linux-backend.sh` 未在本机执行
  - 原因是当前环境为 Windows PowerShell，本轮只能做脚本静态核查
- Docker compose smoke 未执行
- 前端浏览器级回归未执行

## 4. Comparison Notes

### 当前工作区与 HEAD

- 工作区相对 `HEAD` 存在 `81` 个已跟踪文件改动
- 未跟踪文件高达 `2662` 个，说明当前工作区非常脏，构建产物、临时文件和审计文件混在一起
- 最近改动重心很明显落在备份/导入/回滚、dynamic routing、route-target override、relay failover/race、Windows / smoke / validation 脚本、设置页和日志/统计前端

### 当前分支与最近稳定基线

- 相对 `origin/dev` 当前分支 ahead `22`，behind `0`
- 提交层面的主分支差距不大，但当前未提交工作区差异远大于提交差异

### 代码实现与 README / docs / 任务说明一致性

- 一致
  - `README` 里关于 `octopus healthcheck` 的 wildcard host 回落到 `127.0.0.1`，代码实现一致
  - backup export 默认全量、显式 `include_secrets=false` 才脱敏，当前前后端一致
  - dynamic-routing 背景任务只是 `dynamic summary scan`、不会自动写回阈值，和代码一致
- 不一致
  - `README` 把 `scripts/build-win.ps1` 作为可用构建入口，但脚本当前实际不可运行
  - `validation.yaml` 的命名和 README 的“acceptance chain”表述会让人误以为已覆盖当前前端运行产物，实际上 smoke 仍可能命中旧 `static/out`

### 代码实现与测试 / 构建 / 验证覆盖一致性

- 一致
  - 备份、route-target、relay budget、dynamic task 等后端改动已经补了不少 Go 测试文件
- 不一致
  - 最复杂的前端设置页只做了类型检查，没有行为级自动化
  - Linux smoke 依赖的前端静态目录不是当前源码即时 build 结果，导致 smoke 的覆盖对象和工作区源码不一致

## 5. Top Next Actions

### 需要优先处理的前三项

1. 修复前端产物验证链
   - 在 CI 和 smoke 前强制执行前端 build，并把 `web/out` 同步到 `static/out`
   - 或直接让 smoke 使用当前构建产物而不是历史目录
   - 同时加一个“`web/out` 与 `static/out` 漂移即失败”的硬性检查

2. 修复 Windows 交付脚本
   - 先修 `scripts/build-win.ps1:174` 的拼接错误
   - 再修 `scripts/dev-win.ps1` 的 `Start-WindowProcess` 调用签名，使其至少能跑到启动阶段
   - 修完后把 `-CheckOnly` 纳入最小回归验证

3. 给 `Backup` 设置页补最小行为级验证
   - 优先覆盖 export、dry-run、Apply Same Import、rollback preview、post-import validation

### 建议下一步动作

- 第一步先修验证链和 Windows 脚本，不要先继续加功能
- 第二步补 `Backup` 页最小浏览器级回归
- 第三步决定 `dynamic_routing_summary_scan` 是要接成真实 summary surface，还是收缩为纯后台维护能力并更新文档口径

## 本次运行中文摘要

1. 本次触发时间
   - 2026-04-21 00:46:04 +08:00

2. 做了哪些检查、运行了哪些命令
   - 检查了仓库结构、git 状态、最近提交、`HEAD` 差异、`origin/dev` 差异、README/docs、核心后端入口、动态路由、备份导入、route-target、CI 工作流和 smoke/build/dev 脚本
   - 运行了 `tsc --noEmit`
   - 运行了 `next build`
   - 运行了 `scripts/verify-go-env.ps1`
   - 运行了 `scripts/build-win.ps1 -CheckOnly`
   - 运行了 `scripts/dev-win.ps1 -CheckOnly`
   - 运行了 `git diff --no-index --stat -- web/out static/out`

3. 修改了哪些文件；如果没有修改，明确写“本次无变更”
   - `docs/review/审查/2026-04-21-004604-octopus-repo-complete-audit.md`
   - `docs/review/审查/2026-04-21-004604-octopus-repo-complete-audit.html`
   - `C:\Users\李昊桐\.codex\automations\octopus-repo\memory.md`

4. 发现了什么问题
   - 最大问题是 CI/smoke 可能验证旧的 `static/out`，而不是当前前端源码对应的产物
   - Windows build 脚本当前坏掉
   - Windows dev 脚本尾部调用也不正确
   - `Backup` 页前端行为验证仍然不足
   - dynamic routing summary scan 仍未形成真实可见闭环

5. 本次结果是成功、跳过还是失败
   - 成功
   - 说明: 审计报告已经完成，但部分 Go / smoke 验证被本机 Go 执行权限阻塞

6. 是否需要我手动介入
   - 需要
   - 优先介入项: 决定先修验证链还是先修 Windows 脚本；同时需要处理本机 Go 工具链 `Access is denied` 问题，否则本地 Go 验证无法闭环