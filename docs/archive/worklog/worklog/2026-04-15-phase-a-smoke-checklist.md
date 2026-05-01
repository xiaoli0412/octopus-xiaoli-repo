# Phase A - 核心页面冒烟清单

> 目的：为 `Phase A` 提供最小可重复的页面级人工检查清单。
>
> 使用前先执行：
>
> - `powershell -ExecutionPolicy Bypass -File .\scripts\phase-a-check.ps1`

---

## 1. 检查范围

本清单覆盖以下核心模块：

- Home
- Channel
- Group
- Model
- Log
- Setting

对应前端路由配置来源：

- `web/src/route/config.tsx`

---

## 2. Home 页面

对应模块：

- `web/src/components/modules/home`

检查项：

- 页面可正常渲染
- 总览卡片无明显空白或报错
- `TokenBreakdown` 组件可正常显示
- 首页没有因统计字段缺失导致的明显崩溃

通过标准：

- 页面能稳定打开
- 无阻塞性交互错误

---

## 3. Channel 页面

对应模块：

- `web/src/components/modules/channel`

检查项：

- 渠道列表可正常渲染
- 搜索、排序、筛选不报错
- 单个渠道卡片可正常展开编辑
- `key_management_mode`、`key_routing_policy` 相关表单项可显示
- 渠道创建入口可正常打开

通过标准：

- 列表可见
- 编辑表单无阻塞性报错
- 新增表单字段可正常展示

---

## 4. Group 页面

对应模块：

- `web/src/components/modules/group`

检查项：

- 分组列表可正常渲染
- 搜索、排序、筛选不报错
- 分组卡片编辑器可正常打开
- `retry_rounds`、`retry_delay_ms`、`failover_window_sec`、`race_after_fails`、`race_concurrency` 相关配置可显示

通过标准：

- 页面能稳定渲染
- 高级故障转移配置项能正常显示

---

## 5. Model 页面

对应模块：

- `web/src/components/modules/model`

检查项：

- 模型列表可正常渲染
- 搜索、排序、筛选不报错
- 单个模型编辑浮层可正常打开
- 官方价格与网关价格相关字段不会导致页面崩溃

通过标准：

- 模型卡片正常展示
- 编辑浮层正常显示

---

## 6. Log 页面

对应模块：

- `web/src/components/modules/log`

检查项：

- 日志列表可正常渲染
- SSE 或初始日志加载不会导致页面直接报错
- 导出入口可显示
- attempts 相关展示不会因字段缺失而崩溃

通过标准：

- 页面正常打开
- 日志区域可见
- 导出相关 UI 可见

---

## 7. Setting 页面

对应模块：

- `web/src/components/modules/setting`

重点子模块：

- `Backup.tsx`
- `LLMPrice.tsx`
- `CircuitBreaker.tsx`
- `Log.tsx`

检查项：

- 设置页主内容可正常渲染
- 备份/导入区域可正常显示
- 价格、日志、熔断器配置区域可正常显示
- 不出现因新增字段导致的明显布局错乱

通过标准：

- 设置页可见
- 备份/导入区块无阻塞性渲染问题

---

## 8. 移动端基础检查

最小检查宽度：

- `375px`

检查项：

- Home
- Channel
- Group
- Model
- Log
- Setting

通过标准：

- 页面无明显横向溢出
- 关键按钮和表单仍可操作

---

## 9. 记录方式

每次手工冒烟完成后，回填到当前活跃阶段 worklog：

- 是否通过
- 哪个页面失败
- 失败表现
- 是否阻塞进入下一阶段

