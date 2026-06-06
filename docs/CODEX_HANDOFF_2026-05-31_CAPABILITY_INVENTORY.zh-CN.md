# Codex 交接说明：2026-05-31 能力库存联通修复

## 当前状态

本轮已经完成需求收口与代码级排查，还没有开始正式实现“统一能力库存”主线代码。

当前仓库工作区已有 3 个未提交代码改动，属于上一轮已经开始的热修复，后续实现时应直接接着用，不要回退：

- `scripts/dockerfiles/entrypoint.sh`
  - 已把数据目录可写性判断从 `-w "$DATA_DIR"` 改成走目标运行用户判定函数，修复 root 误判问题。
- `internal/relay/dynamic_runtime.go`
  - 已把 race probe 请求克隆改成调用 `InternalLLMRequest.DeepClone()`。
- `internal/transformer/model/model.go`
  - 已新增 `InternalLLMRequest.DeepClone()` 及其嵌套结构深拷贝辅助函数。

当前还有一个**不要提交进正式版本**的本地文档改动：

- `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md`

## 已确认的根因

### 1. 分组创建 / 渠道 key 限制模型没有打通

根因已确认：

- `internal/op/channel.go` 的 `ChannelLLMList()` 当前只读取：
  - `channel.Model`
  - `channel.CustomModel`
- 它**完全没有**把：
  - `channel.Keys[].AllowedModels`
  - 后续将新增的 `channel.Keys[].request_capabilities`
  纳入模型候选库存。

因此会出现：

- 某个 channel 里 6 个 key 都限制了模型；
- 底层选 key 时能部分识别；
- 但分组创建页 `useModelChannelList()` 读取 `/api/v1/model/channel` 时看不到这些真实可服务模型；
- 最终形成“上面限制了，下面读不到”的断层。

相关文件：

- `internal/op/channel.go`
- `internal/server/handlers/model.go`
- `web/src/api/endpoints/model.ts`
- `web/src/components/modules/group/Editor.tsx`
- `internal/op/group.go`

### 2. API 令牌 supported_models 候选与回显不稳定

根因已确认：

- `web/src/components/modules/setting/APIKey.tsx`
  - 当前候选来源是 `useGroupList()`，直接取 `group.name`
  - 不是统一能力库存
  - 不是“真实可服务模型 / 分组候选”
- `supported_models` 既承担权限过滤，又直接依赖当前候选列表展示
- 当分组较多、候选变化、顺序变化、某些已保存项暂时不在当前候选中时，会出现：
  - 选中困难
  - 保存后再次打开看起来像丢失
  - 其他页面读取时和表单显示不一致

相关文件：

- `web/src/components/modules/setting/APIKey.tsx`
- `web/src/components/modules/navbar/DocModal.tsx`
- `web/src/components/modules/ai-automation/index.tsx`
- `internal/server/handlers/apikey.go`
- `internal/model/apikey.go`
- `internal/op/apikey.go`

## 已确定的实现方向

下次启动后，按下面顺序继续：

### A. 先补数据模型与能力库存层

1. 在 `internal/model/channel.go` 给 `ChannelKey` 增加：
   - `RequestCapabilities string \`json:"request_capabilities"\``
2. 同步扩展：
   - `ChannelKeyAddRequest`
   - `ChannelKeyUpdateRequest`
3. 新增规范化函数，建议仍放在 `internal/model/channel.go`：
   - `NormalizeChannelKeyRequestCapabilities`
   - `ChannelKeyRequestCapabilitiesList`
4. 新增能力库存模型，建议新建文件：
   - `internal/model/capability_inventory.go`
   - 至少定义：
     - `ChannelCapabilityInventory`
     - `CapabilityInventoryItem`
     - `LLMChannel` 扩展字段：
       - `KeyCount`
       - `RequestCapabilities`
       - `InventorySource`

### B. 新增统一能力库存聚合

建议新建：

- `internal/op/capability_inventory.go`

至少实现：

1. `BuildChannelCapabilityInventory(channel model.Channel) []model.CapabilityInventoryItem`
2. `ChannelCanServeModel(channel model.Channel, modelName string) bool`
3. `ChannelLLMList(ctx)` 改走能力库存，而不是裸读 `channel.Model/custom_model`
4. `SelectableCapabilityInventory(ctx)` 或等价接口，供：
   - 分组创建页
   - API 令牌页
   - `DocModal`
   - AI 自动化页
   共用

聚合规则已经确定：

- channel 声明模型 + key 无限制：沿用声明模型
- key 有 `AllowedModels`：这些模型进入真实可服务库存
- channel 声明模型和 key 限制交集为空：该模型不进入候选
- channel 未声明模型但 key 限制了模型：允许进入候选
- `request_capabilities` 进入每个模型的能力摘要

### C. 改分组校验链路

文件：

- `internal/op/group.go`

把 `validateGroupChannelModelTarget()` 从：

- `channel.SupportsModel()`
- `channel.HasConfiguredKeyForModel()`

改成基于统一能力库存判断：

- channel 是否真实可服务该模型
- 至少存在一个匹配的可用 key

### D. 改模型接口

文件：

- `internal/server/handlers/model.go`
- `web/src/api/endpoints/model.ts`

需要做的事：

1. 升级 `/api/v1/model/channel`
   - 保留：
     - `name`
     - `channel_id`
     - `channel_name`
   - 新增：
     - `key_count`
     - `request_capabilities`
     - `inventory_source`
2. 新增统一能力库存接口：
   - `GET /api/v1/model/capability-inventory`
3. 前端 `LLMChannel` 类型同步扩展：
   - `key_count?: number`
   - `request_capabilities?: string[]`
   - `inventory_source?: string`

### E. 改 API 令牌表单

文件：

- `web/src/components/modules/setting/APIKey.tsx`

必须做的改动：

1. 候选来源从 `useGroupList()` 改为统一能力库存接口
2. `supported_models` 保存前统一：
   - trim
   - 去重
   - 稳定排序
3. 表单中拆成两块：
   - 当前候选
   - 已保存但当前候选未命中
4. 增加：
   - 搜索
   - 已选摘要
5. 已保存值即使当前候选缺失，也必须继续展示，不能静默消失

### F. 同步依赖页面

文件：

- `web/src/components/modules/navbar/DocModal.tsx`
- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/components/modules/apikey-dashboard/index.tsx`

要求：

- 不再假设 `supported_models` 总能从当前 `group.name` 列表完整解析
- 要能显示“已保存但当前候选未命中”的模型

## 已确定的测试补充

下次实现后需要补这些测试：

### 后端

- `internal/op/channel_test.go`
  - `ChannelLLMList()` 在多个 key 分别限制模型时返回真实可服务模型
  - channel 未声明模型但 key 限制模型时也能返回
- `internal/op/group_test.go`
  - `validateGroupChannelModelTarget()` 对真实可服务模型通过
  - 对仅声明但无可用 key 的模型拒绝
- `internal/server/handlers/model_test.go`
  - `/api/v1/model/channel` 返回新增字段
  - `/api/v1/model/capability-inventory` 正常返回
- `internal/server/handlers/apikey_test.go` 或 `internal/op/apikey` 测试
  - `supported_models` 规范化、稳定排序
  - 已保存值在候选变化后不会被误清空

### 前端

- `web/src/components/modules/group/...`
  - 分组创建页能看到 key 限制模型导出的真实候选
- `web/src/components/modules/setting/APIKey.tsx`
  - 候选过多可搜索
  - 保存后再次打开稳定回显
  - 已保存但当前候选缺失时仍显示

## 下一次启动后建议的第一步

直接按这个顺序执行：

1. 先实现 `ChannelKey.request_capabilities` 数据模型与规范化
2. 新建 `internal/op/capability_inventory.go`
3. 改 `ChannelLLMList()`
4. 改 `/api/v1/model/channel` 与新增 `/api/v1/model/capability-inventory`
5. 改 `Group` 校验
6. 改 `APIKey` 表单与 `DocModal` / AI 自动化读取
7. 补测试

## 注意事项

- 不要回退当前 `dynamic_runtime.go` / `model.go` / `entrypoint.sh` 的未提交改动。
- 不要把 `docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md` 带进正式提交。
- 当前还没开始实际实现 `request_capabilities`，只是已经完成方案与根因确认。
- 关闭 Codex 前，这份文档就是下一次继续干的入口。

## 2026-06-06 新增要求：上游管理模块收口

用户明确补充：New API / sub2API / OpenAI 兼容这类“深链接上游接入”不能继续作为设置页功能，也不应把所有管理内容堆在渠道页里。

后续主线应按下面口径推进：

- 新增与“渠道 / 分组 / 价格”同级的“上游”管理界面，用于管理多个上游站点；渠道页只保留快速接入入口和已接入渠道的普通管理能力。
- 一个上游站点必须支持多种登录 / 授权方式，包括管理 Token、Access Key、账号密码一次性换 token，以及新旧接口差异下的多路径探测。
- 上游接入必须拉取并展示：服务商模型、价格 / 倍率、订阅、余额、不同 Key、不同分组、每个 Key 的模型限制和请求分类。
- UI 必须分层：上游列表、站点详情、Key、分组、价格、刷新记录等用二级切换，不把所有内容堆在一个界面；手机端必须适配。
- 价格候选在接入结果里只显示 2-3 个，其余通过搜索或滚动查看；详细价格管理应联动同级“价格”页面。
- 拉取到的上游价格必须作为“中转计价 / 真实计价”写入，不覆盖官方价格；价格计算与日志展示应能同时显示官方消耗和真实中转消耗。
- 上游刷新机制必须支持手动刷新和定时刷新，避免高频拉取占性能；刷新结果应同步能力库存、渠道 Key 限制、价格候选和普通渠道展示。
- 每个上游接入成功后，必须能在渠道列表里以普通渠道形式正常显示和管理；更细的上游站点、分组、Key、价格刷新管理放到“上游”同级界面。
- 所有英文协议 / 状态 / 来源标签尽量转换为中文展示，减少中英混排；常驻提示和警告仍应收起，只保留阻断类短提示。
- 底层必须打通，不能只做入口：上游目录、渠道 Key、分组候选、价格页、日志 / 消耗计算、API 参考都应使用同一套真实能力与价格口径。
