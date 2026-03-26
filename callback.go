package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CallbackServer 回调服务器
type CallbackServer struct {
	port         string
	path         string
	secret       string
	client       *Client
	openaiClient *OpenAIHandler
}

// NewCallbackServer 创建回调服务器
func NewCallbackServer(port, path, secret string, client *Client, openaiClient *OpenAIHandler) *CallbackServer {
	return &CallbackServer{
		port:         port,
		path:         path,
		secret:       secret,
		client:       client,
		openaiClient: openaiClient,
	}
}

// Start 启动回调服务器
func (s *CallbackServer) Start() error {
    gin.SetMode(gin.ReleaseMode)
    r := gin.New()
    r.Use(gin.Recovery())

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // 回调处理
    r.POST(s.path, s.handleCallback)

    log.Printf("[Callback] 回调服务器启动，监听端口 %s", s.port)
    log.Printf("[Callback] 回调地址: http://localhost:%s%s", s.port, s.path)

    return r.Run(":" + s.port)
}

// handleCallback 处理回调消息
func (s *CallbackServer) handleCallback(c *gin.Context) {
    var msg CallbackMessage
    if err := c.ShouldBindJSON(&msg); err != nil {
        log.Printf("[Callback] 解析消息失败: %v", err)
        c.JSON(http.StatusOK, gin.H{"code": 400, "message": "invalid request"})
        return
    }

    // 记录收到的消息
    log.Printf("[Callback] 收到消息: MsgId=%s, MsgType=%d, FromId=%s, RoomId=%s",
        msg.MsgId, msg.MsgType, msg.FromId, msg.RoomId)

    // 根据协议类型分别处理消息
    if AppConfig.Protocol == ProtocolWeiXin {
        // 个微消息类型处理
        switch msg.MsgType {
        case WxMsgTypeText:
            go s.handleTextMessage(&msg)
        case WxMsgTypeImage:
            go s.handleImageMessage(&msg)
        case WxMsgTypeVoice:
            go s.handleVoiceMessage(&msg)
        case WxMsgTypeVideo:
            go s.handleVideoMessage(&msg)
        case WxMsgTypeCard:
            log.Printf("[Callback] 名片消息: %s", msg.Content)
        case WxMsgTypeEmoji:
            log.Printf("[Callback] 表情消息")
        case WxMsgTypeLocation:
            log.Printf("[Callback] 位置消息: %s", msg.Content)
        case WxMsgTypeApp:
            // APP 类型消息（个微），包含文件、链接、小程序等
            log.Printf("[Callback] APP类型消息: FileId=%s", msg.FileId)
            go s.handleFileMessage(&msg)
        case WxMsgTypeSystem:
            log.Printf("[Callback] 系统消息: %s", msg.Content)
        case WxMsgTypeRevoke:
            log.Printf("[Callback] 撤回消息")
        default:
            log.Printf("[Callback] 个微未处理的消息类型: %d", msg.MsgType)
        }
    } else {
        // 企微消息类型处理
        switch msg.MsgType {
        case QiWeiMsgTypeText:
            go s.handleTextMessage(&msg)
        case QiWeiMsgTypeImage:
            go s.handleImageMessage(&msg)
        case QiWeiMsgTypeFile:
            go s.handleFileMessage(&msg)
        case QiWeiMsgTypeVideo:
            go s.handleVideoMessage(&msg)
        case QiWeiMsgTypeVoice:
            go s.handleVoiceMessage(&msg)
        case QiWeiMsgTypeCard:
            log.Printf("[Callback] 名片消息: %s", msg.Content)
        case QiWeiMsgTypeLink:
            log.Printf("[Callback] 链接/小程序消息: %s", msg.Content)
        default:
            log.Printf("[Callback] 企微未处理的消息类型: %d", msg.MsgType)
        }
    }

    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// handleTextMessage 处理文本消息
func (s *CallbackServer) handleTextMessage(msg *CallbackMessage) {
    log.Printf("[Callback] 文本消息: %s", msg.Content)

    // 如果配置了 OpenAI，则进行自动回复
    if s.openaiClient != nil && s.openaiClient.IsEnabled() {
        reply, err := s.openaiClient.GenerateReply(msg.Content, msg.FromId)
        if err != nil {
            log.Printf("[Callback] OpenAI 生成回复失败: %v", err)
            return
        }

        if reply != "" {
            // 发送回复
            toId := msg.FromId
            if msg.RoomId != "" {
                toId = msg.RoomId
            }

            if err := s.client.SendText(toId, reply); err != nil {
                log.Printf("[Callback] 发送回复失败: %v", err)
            } else {
                log.Printf("[Callback] 已发送 AI 回复: %s", reply)
            }
        }
    }
}

// handleImageMessage 处理图片消息
func (s *CallbackServer) handleImageMessage(msg *CallbackMessage) {
    log.Printf("[Callback] 图片消息: FileId=%s, Size=%d", msg.FileId, msg.FileSize)

    if msg.FileId == "" {
        log.Printf("[Callback] 图片消息缺少 FileId，跳过下载")
        return
    }

    // 检查 FileId 是否已经是 URL（企微某些情况下直接返回 URL）
    if isFileIdUrl(msg.FileId) {
        log.Printf("[Callback] FileId 已经是 URL，无需通过 CDN 下载: %s", msg.FileId)
        return
    }

    // 下载图片示例（根据协议类型自动选择下载接口）
    if AppConfig.HasCDN() {
        savePath := "./downloads/" + msg.FileId + ".jpg"
        downloadClient := NewDownloadClient(s.client, AppConfig.CDNURLQiWei, AppConfig.CDNURLWeiXin, AppConfig.Protocol)
        // fileType: 2=图片
        if err := downloadClient.DownloadFile(msg.FileId, msg.AesKey, 2, savePath); err != nil {
            log.Printf("[Callback] 下载图片失败: %v", err)
        } else {
            log.Printf("[Callback] 图片已保存到: %s", savePath)
        }
    }
}

// handleFileMessage 处理文件消息
func (s *CallbackServer) handleFileMessage(msg *CallbackMessage) {
    log.Printf("[Callback] 文件消息: FileId=%s, FileName=%s, Size=%d",
        msg.FileId, msg.FileName, msg.FileSize)

    // 如果 FileId 为空，跳过下载
    if msg.FileId == "" {
        log.Printf("[Callback] 文件消息没有 FileId，跳过下载")
        return
    }

    // 如果 FileId 已经是 URL，跳过 CDN 下载（可直接访问）
    if isFileIdUrl(msg.FileId) {
        log.Printf("[Callback] FileId 已是 URL: %s，跳过 CDN 下载", msg.FileId)
        return
    }

    // 下载文件示例（根据协议类型自动选择下载接口）
    if AppConfig.HasCDN() {
        savePath := "./downloads/" + msg.FileName
        downloadClient := NewDownloadClient(s.client, AppConfig.CDNURLQiWei, AppConfig.CDNURLWeiXin, AppConfig.Protocol)
        // fileType: 5=文件
        if err := downloadClient.DownloadFile(msg.FileId, msg.AesKey, 5, savePath); err != nil {
            log.Printf("[Callback] 下载文件失败: %v", err)
        } else {
            log.Printf("[Callback] 文件已保存到: %s", savePath)
        }
    }
}

// handleVideoMessage 处理视频消息
func (s *CallbackServer) handleVideoMessage(msg *CallbackMessage) {
    log.Printf("[Callback] 视频消息: FileId=%s, Size=%d", msg.FileId, msg.FileSize)

    // 如果 FileId 为空，跳过下载
    if msg.FileId == "" {
        log.Printf("[Callback] 视频消息没有 FileId，跳过下载")
        return
    }

    // 如果 FileId 已经是 URL，跳过 CDN 下载（可直接访问）
    if isFileIdUrl(msg.FileId) {
        log.Printf("[Callback] FileId 已是 URL: %s，跳过 CDN 下载", msg.FileId)
        return
    }

    // 下载视频示例（根据协议类型自动选择下载接口）
    if AppConfig.HasCDN() {
        savePath := "./downloads/" + msg.FileId + ".mp4"
        downloadClient := NewDownloadClient(s.client, AppConfig.CDNURLQiWei, AppConfig.CDNURLWeiXin, AppConfig.Protocol)
        // fileType: 4=视频
        if err := downloadClient.DownloadFile(msg.FileId, msg.AesKey, 4, savePath); err != nil {
            log.Printf("[Callback] 下载视频失败: %v", err)
        } else {
            log.Printf("[Callback] 视频已保存到: %s", savePath)
        }
    }
}

// handleVoiceMessage 处理语音消息
func (s *CallbackServer) handleVoiceMessage(msg *CallbackMessage) {
    log.Printf("[Callback] 语音消息: FileId=%s, Size=%d", msg.FileId, msg.FileSize)

    // 如果 FileId 为空，跳过下载
    if msg.FileId == "" {
        log.Printf("[Callback] 语音消息没有 FileId，跳过下载")
        return
    }

    // 如果 FileId 已经是 URL，跳过 CDN 下载（可直接访问）
    if isFileIdUrl(msg.FileId) {
        log.Printf("[Callback] FileId 已是 URL: %s，跳过 CDN 下载", msg.FileId)
        return
    }

    // 下载语音示例（根据协议类型自动选择下载接口）
    if AppConfig.HasCDN() {
        savePath := "./downloads/" + msg.FileId + ".amr"
        downloadClient := NewDownloadClient(s.client, AppConfig.CDNURLQiWei, AppConfig.CDNURLWeiXin, AppConfig.Protocol)
        // fileType: 5=语音
        if err := downloadClient.DownloadFile(msg.FileId, msg.AesKey, 5, savePath); err != nil {
            log.Printf("[Callback] 下载语音失败: %v", err)
        } else {
            log.Printf("[Callback] 语音已保存到: %s", savePath)
        }
    }
}
