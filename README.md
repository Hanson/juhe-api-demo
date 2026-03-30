# 聚合聊天 - PHP 回调服务 Demo

接收企微/个微回调，实现自动回复示例。

## 分支说明

- **master 分支**: PHP 版本（当前分支）
- **golang 分支**: Go 版本（完整功能，包含文件下载、发送图片等）

## 快速开始

### 1. 配置

```bash
cp .env.example .env
```

编辑 `.env`:
```env
PROTOCOL=qiwei
API_URL=http://your-api-server:8080
GUID=your-instance-guid
APP_KEY=your-app-key
APP_SECRET=your-app-secret
CALLBACK_URL=http://你的公网IP:9000/callback
```

### 2. 运行服务

```bash
php -S 0.0.0.0:9000
```

### 3. 测试企微回调

```bash
curl -X POST http://localhost:9000/callback \
  -H "Content-Type: application/json" \
  -d '{
    "guid": "test-guid",
    "notify_type": 11010,
    "data": {
      "content": "123",
      "from_id": "user001",
      "room_id": "0",
      "msg_type": 1
    }
  }'
```

### 4. 测试个微回调

```bash
curl -X POST http://localhost:9000/callback \
  -H "Content-Type: application/json" \
  -d '{
    "guid": "test-guid",
    "notify_type": 1010,
    "data": {
      "content": "123",
      "from_username": "wxid_xxx",
      "room_username": "",
      "msg_type": 1
    }
  }'
```

## 回调格式

### 企微回调结构
```json
{
  "guid": "实例GUID",
  "notify_type": 11010,
  "data": {
    "content": "消息内容",
    "from_id": "发送者ID",
    "room_id": "群ID（群消息时）",
    "msg_type": 1
  }
}
```

### 个微回调结构
```json
{
  "guid": "实例GUID",
  "notify_type": 1010,
  "data": {
    "content": "消息内容",
    "from_username": "wxid_xxx",
    "room_username": "xxx@chatroom",
    "msg_type": 1
  }
}
```

## 自动回复逻辑

企微和个微都有独立的处理文件：
- `reply_qiwei.php` - 企微自动回复
- `reply_weixin.php` - 个微自动回复

当前示例：用户发送 `123` → 回复 `456`

修改自动回复逻辑，编辑对应文件即可。

## 目录结构

```
├── index.php        # 回调入口
├── config.php       # 配置文件
├── api_client.php   # API 客户端（调用 /open/GuidRequest）
├── reply_qiwei.php  # 企微自动回复
├── reply_weixin.php # 个微自动回复
├── .env.example     # 环境变量示例
└── README.md
```

## notify_type 参考

| 类型 | 企微 | 个微 | 说明 |
|------|------|------|------|
| 新消息 | 11010 | 1010 | - |
| 用户登录 | 11003 | 1003 | - |
| 用户登出 | 11004 | 1004 | - |
| 好友变更 | 2131 | - | - |
| 好友申请 | 2132 | - | - |
| 群成员加入 | 1002 | 1046 | - |
| 群成员退出 | 1003 | 1048 | - |
| 群踢人 | 1004 | - | - |
| 群退群 | 1005 | - | - |
| 群创建 | 1006 | - | - |
