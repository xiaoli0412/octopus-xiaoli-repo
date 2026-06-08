# 开发计划

- [x] 更方便的添加渠道：
  - [x] 先选渠道类型，根据类型自动带入base_url和models
  - [x] 渠道名称添加placeholder
  - [x] base url添加tooltip提示，需要添加/v1结尾
  - [x] 根据渠道自动获取models，显示自动添加模型小按钮
  - [x] 保存按钮左边添加 测试 一级一个 全部测试按钮，和详情tab页的测试功能一样
  - [x] 支持更多渠道，包括：
    - [x] opencode zen
    - [x] 质谱AI
    - [x] 质谱Coding Plan
    - [x] Github Copilot，通过授权码方式（OAuth Device Flow 已实现自动授权与轮询填充）
    - [x] OpenRouter
    - [x] Vercel AI Gateway
    - [x] Antigravity，参考 https://github.com/jenslys/opencode-gemini-auth 里的反代实现，实现网页方式登录（OAuth Web Flow 已实现，支持授权页跳转 + 回调 + 轮询）
  - [x] 部分渠道不是简单的base_url + key模式，所以go中数据存储模型可能要变更
  - [x] 每个渠道尽量先从 base_url + /models 获取可用模型，如果这个接口不支持，尽量从 models.dev网站拉取，如果还没有则默认空，让用户自己填

- [x] 全局设置添加一个设置系统api地址的地方，比如设置为 http://192.168.1.100:3000，默认 http://localhost:3000。然后在全局左侧navbar中底下添加一个 doc 图标，点击后弹框，里面显示一个curl code代码区域，告知用户如何使用当前api。其中curl部分里的 base_url 就是全局设置的，api_key可以点击选择，model可以选择分组名，code添加复制按钮可复制。类型可以选择 OpenAI Chat(/chat/completions)或 OpenAI Responses(/responses) 或 Anthropic（/messages），不同类型会在结尾添加不同的api后缀，参数格式也不一样，具体可参考 @internal/server/handlers/relay.go