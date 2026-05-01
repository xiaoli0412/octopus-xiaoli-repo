# Phase G - 运行入口与部署链路收口

> 本轮任务聚焦 Linux 服务器运行优先与 Windows 本机试跑，不扩散到无关业务功能。

---

## 1. 任务信息

- 任务名称：运行入口与部署链路收口
- 日期：2026-04-16
- 当前阶段：Phase G
- 对应 milestone：里程碑 6 验证与部署的前置运行链路

## 2. 开工前输入

- 对应 canonical 章节：
  - 第 14 节验收标准
  - 第 13 节里程碑 6 —— 验证与部署
- 对应 workflow 章节：
  - [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md) 第 9 节
- 上一个相关 worklog：
  - `2026-04-15-phase-a-stability-closure.md`
- 本次任务目标：
  - 补齐 Linux 服务器运行入口
  - 补齐 Windows 本机构建与试跑入口
  - 让 README 与真实运行链路一致
- 本次已盘点本地资源：
  - `AGENTS.md`
  - `docs/LLM-Gateway-Refactor-Plan.zh-CN.md`
  - `docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md`
  - `README.md`
  - `README_zh.md`
  - `scripts/build.sh`
  - `scripts/dockerfiles/*`
  - `docker-compose.yml`
  - `internal/conf/config.go`
  - `internal/server/server.go`
  - `internal/server/middleware/static.go`
  - `static/static.go`
- 本次使用的本地 resources / skills / 记忆上下文：
  - 使用当前阶段 worklog、README、现有构建脚本和静态资源链路代码，确认仓库已经具备嵌入式前端与 Docker 二进制准备能力
  - 使用此前 Phase A 自动化验证结论，避免重复质疑已通过的后端/前端基础构建链路
- 若未使用部分本地资源或上下文，原因：
  - 未使用外部资料，因为本轮问题可完全由仓库内现状推导
- 本次是否启用子 agent 与分工边界：
  - 是
  - 子 agent 只读复核 Linux 服务器运行缺口与 Windows 本机脚本建议
  - 主线程负责规则落盘、Go 侧运行链路实现、脚本实现与最终验证
- 本次子 agent 使用模型：
  - `gpt-5.4`

## 3. 本次硬规则

- 不偏离 canonical plan 主线
- 不新增无关业务功能
- 运行入口必须优先服务 Linux 服务器与 Windows 本机试跑
- README、脚本、代码三者必须保持一致

## 4. 本次禁止事项

- 不把部署链路补位扩展成新的产品功能
- 不伪造 Docker / 页面手工 smoke 结果
- 不用危险方式删除仓库外路径

## 5. 本次验收条件

- Go 侧静态资源运行链路测试通过
- 全量 `go test ./...` 通过
- Windows 构建脚本完整跑通
- Linux 开发入口脚本可做最小检查
- README 中英双语与实际入口一致

## 6. 本次回滚点

- 运行入口脚本与 README 可单独回退
- Go 侧静态资源链路可回退到 embed-only 版本

## 7. 实现范围

- 先改数据语义还是先改 UI：不涉及 UI 业务逻辑，优先补运行时与脚本
- 受影响后端模块：
  - `internal/conf/config.go`
  - `internal/server/server.go`
  - `internal/server/middleware/static.go`
- 受影响前端模块：
  - 无
- 受影响接口：
  - 无新增业务 API
- 是否影响旧数据：
  - 否
- 是否影响旧行为：
  - 仅增强静态资源加载优先级与运行入口，不改变业务接口行为

## 8. 实施步骤

1. 补 `server.static_dir` 与本地静态目录优先、embed 兜底逻辑。
2. 补 Windows 本机构建/试跑脚本与 Linux 开发入口脚本。
3. 同步更新 README 与 compose 示例，并做本地验证。

## 9. 测试与验证

- 构建命令：
  - `go build ./...`
- 测试命令：
  - `go test ./internal/server/...`
  - `go test ./...`
- 专项验证：
  - `powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1 -CheckOnly`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1 -CheckOnly`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1`
  - `bash ./scripts/dev-linux.sh --check-only`
  - `bash ./scripts/build.sh help`
  - `bash ./scripts/build.sh linux-server`

## 10. 风险与兼容性

- 新风险：
  - Windows 构建脚本在 PowerShell 下的参数转义容易出错，需实跑验证
  - 完整构建会覆盖 `static/out`，需要确保只操作仓库内目标目录
  - Linux 服务器主产物构建仍依赖 `python3` 执行价格更新步骤，当前环境缺少该命令时会阻塞 `linux-server` 入口
- 兼容性风险：
  - `server.static_dir` 会改变静态资源加载优先级，但不会影响 `/api/*` 与 `/v1/*`
- 是否阻塞下一任务：
  - 否，本轮属于部署验证链路补位，可独立收口

## 11. 收工记录

- 构建是否通过：
  - `go build ./...`：通过
- 测试是否通过：
  - `go test ./internal/server/...`：通过
  - `go test ./...`：通过
  - `powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1 -CheckOnly`：通过
  - `powershell -ExecutionPolicy Bypass -File .\scripts\dev-win.ps1 -CheckOnly`：通过
  - `powershell -ExecutionPolicy Bypass -File .\scripts\build-win.ps1`：通过
  - `bash ./scripts/dev-linux.sh --check-only`：通过
  - `bash ./scripts/build.sh help`：通过
  - `bash ./scripts/build.sh linux-server`：失败，当前环境缺少 `python3`
- 本次使用了哪些本地资源 / skills / 记忆上下文：
  - 使用 canonical plan、workflow、README、现有脚本和静态资源代码确认运行链路缺口
- 本次使用的本地资源 / skills / 记忆上下文分别提供了什么结论：
  - `scripts/build.sh` 与 `scripts/dockerfiles/*` 说明 Linux/Docker 产物链已存在，但缺少直接开发入口
  - `docker-compose.yml` 说明当前 compose 仍是占位级示例，需要收成可直接用的最小版本
  - `static/static.go`、`server.go`、`static.go` 说明服务端适合补本地目录优先、embed 兜底
  - `internal/server/handlers/user.go` 说明当前仓库已有 `/api/v1/user/status` 可作为后续更真实 healthcheck 的候选路径
- 本次使用了哪些子 agent 及其结论：
  - 使用 `gpt-5.4` 子 agent 只读复核 Linux 与 Windows 运行链路缺口
- 子 agent 分工、负责范围与产出摘要：
  - 只读指出：当前最值得补的就是 Linux 开发入口、Windows 本机构建/试跑脚本和 README/compose 对齐
- 手工 smoke 状态：
  - 未执行
- 手工 smoke 阻塞原因 / 缺少的环境：
  - 当前仅补运行链路，不包含浏览器页面人工检查
- 待验证页面清单：
  - Home
  - Channel
  - Group
  - Model
  - Log
  - Setting
- 若未使用子 agent，原因：
  - 不适用
- worklog 是否更新：
  - 是
- 遗留项：
  - 仍需真实 Docker / compose 运行验证
  - 仍需真实页面手工 smoke
  - 若要在当前环境跑通 `./scripts/build.sh linux-server`，还需补齐 `python3`
- 下一任务前置条件是否满足：
  - 运行链路层面基本满足，可继续回到主线业务收口
