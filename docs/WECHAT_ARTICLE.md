# 为什么我要自己搭一个 LLM API 网关？——Octopus 使用指南

> 本文面向有一定技术基础、已在使用 Claude、GPT 或其他 AI API 的开发者。

---

## 先说痛点：你是不是也遇到过这些？

**场景一**：你同时买了 Claude API、GPT-4o API，还薅了 Gemini 的免费额度，但每个工具的 Key 要分别填到不同的工具里——Claude Code 一个 Key，Cursor 一个 Key，自己写的脚本又是另一个 Key。  
改一次 Key，要去四五个地方改。

**场景二**：API 调用量不小，但 provider 偶尔限速（429）、偶尔不可用（5xx），你的程序直接失败，没有自动重试，也没有 fallback。

**场景三**：你给团队里的其他人也开了 API 权限，但根本不知道谁用了多少、用在了哪，费用完全是黑盒。

**场景四**：想接 GitHub Copilot 的后端能力，或者用 Google Gemini Code Assist 的免费额度，但这些服务走 OAuth，不是标准 API Key，和你现有工具完全不兼容。

**场景五**：OpenClaw真好用，但是也真不安全，我的key很可能在我不知道的情况下被泄露出去，导致费用炸表。

---

## Octopus 解决的就是这些问题

Octopus 是一个**为个人和小团队设计的 LLM API 聚合网关**。它的核心思路很简单：

> 所有 AI 服务的请求，都走同一个入口。由 Octopus 来负责路由、负载均衡、协议转换和记录。

你只需要告诉所有工具一个地址：`http://your-server:1088`，和一个统一的 API Key。背后接了多少个服务商、哪个 Key 在用、用了多少，完全由 Octopus 管理。

---

## 核心能力一览

### 1. 多渠道聚合，一个接口走天下

支持 OpenAI、Anthropic（Claude）、Google Gemini、火山引擎（豆包）等主流服务商，以及：

- **GitHub Copilot**：通过官方 OAuth Device Flow，用你的 GitHub 账号直接接入 Copilot 后端
- **Google Gemini Code Assist（Antigravity）**：通过 Google OAuth，使用免费的个人 Code Assist 额度，包含 `gemini-2.5-flash`、`gemini-2.5-pro` 等模型，零成本

接入后，所有工具统一走 OpenAI Chat 格式（`/v1/chat/completions`）调用，无需改代码。

### 2. 负载均衡 + 熔断器

同一个模型可以配置多个渠道，支持：

| 策略 | 适用场景 |
|------|---------|
| 轮询（Round Robin） | 多 Key 平摊请求 |
| 随机（Random） | 简单分流 |
| 故障转移（Failover） | 主备切换 |
| 加权（Weighted） | 按配额比例分配 |

内置**熔断器**：某个渠道连续失败后自动切走，冷却期结束后尝试恢复，不需要你手动干预。

### 3. 协议转换，真正的万能适配器

| 入站格式 | 出站格式 |
|---------|---------|
| OpenAI Chat | OpenAI / Anthropic / Gemini / 火山引擎 |
| OpenAI Responses | ← 同上 |
| Anthropic Messages | ← 同上 |
| Embeddings | OpenAI Embeddings |

你的工具说 OpenAI，后端其实是 Claude——Octopus 在中间自动翻译，双方都不知道。

### 4. 精细的 API Key 管理

- 给不同的人/工具发不同的 Key
- 每个 Key 可以设置**允许使用的模型范围**、**费用上限**、**有效期**
- 实时看到每个 Key 的用量和费用，随时停用

### 5. 统计与日志

- 按模型、按渠道、按 API Key 统计用量和费用
- 完整的请求日志，可按时间/模型/状态过滤
- 支持小时/天级别的历史趋势图

### 6. 一键导入 CC Switch

如果你用 Claude Code、Codex 或 Gemini CLI，可以在 API 文档页面直接生成导入链接，一键把 Octopus 配置为这些工具的后端。

---

## 📸 界面预览

### 🖥️ 桌面端

**首页**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-home.png" alt="首页" width="600">

**渠道**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-channel.png" alt="渠道" width="600">

**分组**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-group.png" alt="分组" width="600">

**价格**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-price.png" alt="价格" width="600">

**日志**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-log.png" alt="日志" width="600">

**设置**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-setting.png" alt="设置" width="600">

**Curl 示例**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-api-curl.png" alt="Curl 示例" width="600">

**CC Switch**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/desktop-api-cc.png" alt="CC Switch" width="600">

### 📱 移动端

**首页**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-home.png" alt="移动端首页" width="300">

**渠道**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-channel.png" alt="移动端渠道" width="300">

**分组**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-group.png" alt="移动端分组" width="300">

**价格**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-price.png" alt="移动端价格" width="300">

**日志**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-log.png" alt="移动端日志" width="300">

**设置**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-setting.png" alt="移动端设置" width="300">

**Curl 示例**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-api-curl.png" alt="移动端 Curl" width="300">

**CC Switch**

<img src="https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/web/public/screenshot/mobile-api-cc.png" alt="移动端 CC Switch" width="300">

## 快速上手

### Docker（最简单）

```bash
docker run -d \
  --name octopus \
  -v /path/to/data:/app/data \
  -p 1088:1088 \
  ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.16.4
```

浏览器打开 `http://localhost:1088`，默认账号密码 `admin` / `admin`。

### Docker Compose

```bash
wget https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/refs/heads/main/docker-compose.yml
docker compose up -d
```

---

## 实际使用场景

### 场景：Claude Code + 免费 Gemini 额度

1. 在 Octopus 添加 **Antigravity** 渠道 → 用 Google 账号完成 OAuth → 自动获取 Gemini Code Assist 可用模型
2. 添加 **Anthropic** 渠道 → 填入 Claude API Key
3. 创建一个分组 `coding`，包含上面两个渠道，策略设为 Failover（Claude 优先，Gemini 兜底）
4. 在 Claude Code 里填入 Octopus 地址和 API Key，模型选 `coding`，或者使用`Octopus`页面上的导入`CC Switch`的方法
5. 完成。Claude 正常时走 Claude，限速时自动用免费 Gemini 兜底

### 场景：给团队统一发 Key

1. 在 Octopus 为每个人创建独立 API Key，设置各自的费用上限
2. 所有人统一用 `http://octopus:1088` 这个地址
3. 管理员在统计面板看谁用了多少，月底一目了然

---

## 开源 & 部署

- GitHub：https://github.com/xiaoli0412/octopus-xiaoli-repo
- 上游：https://github.com/bestruirui/octopus（感谢原作者 bestruirui）
- 支持平台：Linux / macOS / Windows，Docker 支持 amd64 / arm64 / armv7

本分支（`main`）在上游基础上额外支持：
- GitHub Copilot OAuth 接入
- Google Gemini Code Assist（Antigravity）OAuth 接入
- Providers 预设库（20+ 供应商开箱即用）
- CC Switch 一键导入
- 渠道模型测试（创建前验证可用性）

---

## 最后

如果你符合以下任意一条，Octopus 值得你花 5 分钟部署试用：

- 同时使用 2 个以上 AI 服务商
- 需要给自己或团队统一管理 API 消费
- 想用免费的 GitHub Copilot 或 Gemini Code Assist 作为备用模型
- 希望在不改代码、不换工具的情况下切换底层模型

**部署一次，受益长期。**
