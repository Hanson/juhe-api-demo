# 聚合聊天 - 后端开放平台 Demo

这是一个 Go 语言实现的后端开放平台接口 Demo，用于对接 微信 企微协议 SAAS 平台。

## 协议类型说明

本项目支持两种协议：

| 协议 | 配置值 | 说明 | 适用场景 |
|------|--------|------|----------|
| **企微（企业微信）** | `qiwei` | 企业级微信协议 | 企业内部通讯、客户管理、企业群聊 |
| **个微（个人微信）** | `weixin` | 个人微信协议 | 个人号自动化、社群运营、私域流量 |

### 协议差异

| 功能 | 企微（企业微信） | 个微（个人微信） |
|------|-----------------|-----------------|
| 用户标识 | vid / corp_id | wxid |
| 群标识 | chat_id / room_id | chatroom_id |
| 外部联系人 | 支持 | 支持 |
| 企业内部通讯 | 支持 | 不支持 |
| 客户群管理 | 支持 | 支持 |
| 朋友圈 | 支持 | 支持 |
| 小程序发送 | 支持 | 支持 |

## 功能特性

### 通用功能
- ✅ 设置回调地址
- ✅ 接收消息回调（文本、图片、文件、视频、语音）
- ✅ OpenAI 自动回复
- ✅ 接收图片/文件/视频并下载
- ✅ 发送文本消息
- ✅ 发送图片消息
- ✅ 发送小程序消息
- ✅ 发送文件消息
- ✅ 发送链接消息
- ✅ 私有化 CDN 文件上传/下载

### 企微专属功能
- ✅ 企业内部群聊
- ✅ 外部客户管理
- ✅ 企业通讯录
- ✅ 客户群列表

## 项目结构

```
juhe-api-demo/
├── main.go              # 主程序入口
├── config.go            # 配置管理（支持企微/个微协议选择）
├── client.go            # API 客户端
├── models.go            # 数据模型（消息类型常量）
├── message.go           # 消息发送接口
├── download.go          # 文件下载（私有化 CDN，智能选择下载接口）
├── client_download.go   # 客户端下载方法
├── callback.go          # 回调服务器
├── openai_handler.go    # OpenAI 自动回复
├── go.mod               # Go 模块定义
├── .env.example         # 环境变量示例
└── README.md            # 说明文档
```

## 快速开始

### 1. 配置环境变量

复制 `.env.example` 为 `.env` 并填写配置：

```bash
cp .env.example .env
```

#### 企微配置示例

```ini
# 协议类型: qiwei(企微)
PROTOCOL=qiwei

# 后端开放平台 API 配置
API_URL=http://your-api-server:8080
GUID=your-instance-guid

# 开放平台认证配置（必需）
APP_KEY=your-app-key
APP_SECRET=your-app-secret

# 回调服务配置
CALLBACK_PORT=9000
CALLBACK_PATH=/callback
CALLBACK_URL=http://your-public-server:9000/callback

# 私有化 CDN 配置（企微）
CDN_URL_QIWEI=http://your-cdn-server:8081

# OpenAI 配置（可选，用于自动回复）
OPENAI_API_KEY=your-openai-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-3.5-turbo
```

#### 个微配置示例

```ini
# 协议类型: weixin(个微)
PROTOCOL=weixin

# 后端开放平台 API 配置
API_URL=http://your-api-server:8080
GUID=your-instance-guid

# 开放平台认证配置（必需）
APP_KEY=your-app-key
APP_SECRET=your-app-secret

# 回调服务配置
CALLBACK_PORT=9000
CALLBACK_PATH=/callback
CALLBACK_URL=http://your-public-server:9000/callback

# 私有化 CDN 配置（个微）
CDN_URL_WEIXIN=http://your-cdn-server:8082

# OpenAI 配置（可选，用于自动回复）
OPENAI_API_KEY=your-openai-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-3.5-turbo
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 运行服务

```bash
# 启动回调服务（包含自动回复功能）
go run . -action start
```

## 使用示例

### 设置回调地址

```bash
go run . -action set-callback
```

### 发送文本消息

```bash
# 企微：接收者ID 为 vid 或 chat_id
# 个微：接收者ID 为 wxid 或 chatroom_id
go run . -action send-text -to "接收者ID" -content "你好，这是测试消息"
```

### 发送图片消息

```bash
# 图片 URL 需要公网可访问
go run . -action send-image -to "接收者ID" -image-url "https://example.com/image.jpg"
```

### 发送小程序消息

```bash
go run . -action send-weapp \
  -to "接收者ID" \
  -app-id "wx1234567890abcdef" \
  -title "小程序标题" \
  -desc "小程序描述" \
  -page "pages/index/index" \
  -icon "https://example.com/icon.png"
```

### 发送链接消息

```bash
go run . -action send-link \
  -to "接收者ID" \
  -url "https://example.com" \
  -title "链接标题" \
  -desc "链接描述" \
  -icon "https://example.com/icon.png"
```

## 消息类型说明

### 企微消息类型

| 类型 | 值 | 说明 |
|------|-----|------|
| QiWeiMsgTypeText | 1 | 文本消息 |
| QiWeiMsgTypeImage | 2 | 图片消息 |
| QiWeiMsgTypeFile | 3 | 文件消息 |
| QiWeiMsgTypeVideo | 4 | 视频消息 |
| QiWeiMsgTypeVoice | 5 | 语音消息 |
| QiWeiMsgTypeCard | 42 | 名片消息 |
| QiWeiMsgTypeLink | 49 | 链接消息 |

### 个微消息类型

| 类型 | 值 | 说明 |
|------|-----|------|
| WxMsgTypeText | 1 | 文本消息 |
| WxMsgTypeImage | 3 | 图片消息 |
| WxMsgTypeVoice | 34 | 语音消息 |
| WxMsgTypeCard | 37 | 加好友请求 |
| WxMsgTypeCard | 42 | 名片消息 |
| WxMsgTypeVideo | 43 | 视频消息 |
| WxMsgTypeEmoji | 47 | 表情消息 |
| WxMsgTypeLocation | 48 | 位置消息 |
| WxMsgTypeApp | 49 | APP类型消息（文件、链接、小程序等） |
| WxMsgTypeSystem | 10000 | 系统消息 |
| WxMsgTypeRevoke | 10002 | 撤回消息 |

## 私有化 CDN 下载说明

### 智能下载接口选择

本 Demo 根据不同的 `file_id` 格式和协议类型，自动选择正确的下载接口：

| 协议 | file_id 格式 | 下载接口 |
|------|--------------|----------|
| 企微 | `https://wework.qpic.cn/xxx` | 直接 http.Get 下载 |
| 企微 | 小文件（图片、缩略图） | `/cloud/c2c_download` |
| 企微 | 大文件（视频、文件 >25MB） | `/cloud/big_download` |
| 个微 | 所有文件 | `/cloud/c2c_download` |

### 文件类型说明

| fileType | 说明 |
|----------|------|
| 2 | 图片 |
| 3 | 小程序封面 |
| 4 | 视频 |
| 5 | 文件/语音 |

### CDN 信息缓存

CDN 信息通过 `/cdn/get_cdn_info` 接口获取，并缓存 3 小时，避免频繁请求。

## 代码示例

### 在代码中使用客户端

```go
package main

import (
    "log"
)

func main() {
    // 加载配置
    cfg := LoadConfig()

    // 查看当前协议
    log.Printf("当前协议: %s", cfg.ProtocolName())

    // 创建客户端（需要 APP_KEY 和 APP_SECRET 进行认证）
    client := NewClient(cfg.APIURL, cfg.GUID, cfg.AppKey, cfg.AppSecret, cfg.Protocol)

    // 发送文本消息
    err := client.SendText("user_id", "你好！")
    if err != nil {
        log.Fatal(err)
    }

    // 发送小程序
    err = client.SendWeAppSimple(
        "user_id",
        "wx1234567890abcdef",
        "小程序标题",
        "小程序描述",
        "pages/index/index",
        "https://example.com/icon.png",
    )
    if err != nil {
        log.Fatal(err)
    }

    // 发送图片（通过 URL）
    err = client.SendImageByURL("user_id", "https://example.com/image.jpg")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 处理回调消息

回调消息会在 `callback.go` 的 `handleCallback` 方法中处理。你可以根据需要修改处理逻辑：

```go
func (s *CallbackServer) handleTextMessage(msg *CallbackMessage) {
    // 自定义文本消息处理逻辑
    log.Printf("收到文本消息: %s", msg.Content)

    // 自动回复示例
    if s.openaiClient != nil && s.openaiClient.IsEnabled() {
        reply, _ := s.openaiClient.GenerateReply(msg.Content, msg.FromId)
        s.client.SendText(msg.FromId, reply)
    }
}
```

### 下载文件示例

```go
// 创建下载客户端
downloadClient := NewDownloadClient(client, cdnURLQiWei, cdnURLWeiXin, protocol)

// 智能下载文件（自动选择正确的下载接口）
// fileType: 2=图片, 3=小程序封面, 4=视频, 5=文件/语音
err := downloadClient.DownloadFile(fileId, aesKey, fileType, savePath)
if err != nil {
    log.Printf("下载失败: %v", err)
}
```

## 回调消息结构

```json
{
  "msg_id": "消息ID",
  "msg_type": 1,
  "from_id": "发送者ID",
  "to_id": "接收者ID",
  "content": "消息内容",
  "room_id": "群ID（群消息时）",
  "create_time": 1234567890,
  "guid": "实例GUID",
  "file_id": "文件ID（图片/文件消息时）",
  "aes_key": "AES密钥（图片/文件消息时）",
  "file_size": 1024,
  "file_name": "文件名"
}
```

## 企微 vs 个微 ID 格式

| 类型 | 企微格式 | 个微格式 |
|------|----------|----------|
| 用户ID | `1688850535328682` (数字vid) | `wxid_xxxxxxxx` |
| 群ID | `wrxxxxxxxx` (chat_id) | `12345678@chatroom` |
| 企业ID | `1970325032041886` | 无 |

## GuidRequest 请求格式

所有 API 请求都通过 `/open/GuidRequest` 端点发送，请求格式如下：

```json
{
  "app_key": "your-app-key",
  "app_secret": "your-app-secret",
  "path": "/msg/send_text",
  "data": {
    "guid": "your-instance-guid",
    "conversation_id": "user_id",
    "content": "消息内容"
  }
}
```

请求流程：
1. 客户端构建请求数据（包含 `app_key`、`app_secret`、`path`、`data`）
2. 发送 POST 请求到 `/open/GuidRequest`
3. 服务端验证 `app_key` 和 `app_secret`
4. 服务端将请求转发到对应的协议服务器
5. 返回协议服务器的响应

## 注意事项

1. **协议选择**：确保 `PROTOCOL` 配置与实际使用的协议类型一致
2. **认证配置**：`APP_KEY` 和 `APP_SECRET` 是必需的，用于开放平台 GuidRequest 认证
3. **回调地址**：`CALLBACK_URL` 必须是公网可访问的地址，否则无法收到回调
4. **OpenAI**：如果不配置 `OPENAI_API_KEY`，自动回复功能将被禁用
5. **文件上传**：发送图片/文件时，需要提供公网可访问的 URL，系统会先上传到平台获取 `file_id`
6. **ID 格式**：企微和个微的用户ID、群ID格式不同，注意区分
7. **CDN 配置**：企微和个微使用不同的 CDN 服务器，需要分别配置 `CDN_URL_QIWEI` 和 `CDN_URL_WEIXIN`

## 参考文档

- [企微协议 SAAS API 文档](https://wework.apifox.cn/doc-7068784)
- [个微协议 API 文档](https://weixins.apifox.cn/)
- [私有化云存储文档](https://wework.apifox.cn/doc-7176461)

## License

MIT
