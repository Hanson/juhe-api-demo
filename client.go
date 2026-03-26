package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// Client API 客户端
type Client struct {
	baseURL   string
	guid      string
	appKey    string
	appSecret string
	httpClient *http.Client
	protocol  ProtocolType
}

// NewClient 创建新的 API 客户端
func NewClient(baseURL, guid, appKey, appSecret string, protocol ProtocolType) *Client {
	return &Client{
		baseURL:   baseURL,
		guid:      guid,
		appKey:    appKey,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		protocol: protocol,
	}
}

// GuidRequest 发送 GuidRequest 格式的请求（开放平台标准格式）
func (c *Client) GuidRequest(ctx context.Context, path string, data map[string]interface{}) (string, error) {
	// 添加 guid 到 data
	data["guid"] = c.guid

	// 构建 GuidRequest 请求体
	reqBody := map[string]interface{}{
		"app_key":    c.appKey,
		"app_secret": c.appSecret,
		"path":       path,
		"data":       data,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.baseURL + "/open/GuidRequest"
	log.Printf("[API] POST %s, path=%s", url, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	log.Printf("[API] Response: %s", string(respBody))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("状态码非200: %d", resp.StatusCode)
	}

	// 解析错误码
	errCode := gjson.Get(string(respBody), "error_code").Int()
	if errCode != 0 {
		errMsg := gjson.Get(string(respBody), "error_message").String()
		if errMsg == "" {
			errMsg = gjson.Get(string(respBody), "err_msg").String()
		}
		if errMsg == "" {
			errMsg = "操作失败"
		}
		return "", fmt.Errorf("API错误 [%d]: %s", errCode, errMsg)
	}

	return string(respBody), nil
}

// request 发送 API 请求（直接调用，不走 GuidRequest）
func (c *Client) request(ctx context.Context, path string, body map[string]interface{}) (string, error) {
	// 添加 guid 到请求体
	body["guid"] = c.guid

	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.baseURL + path
	log.Printf("[API] POST %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	log.Printf("[API] Response: %s", string(respBody))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("状态码非200: %d", resp.StatusCode)
	}

	// 解析错误码
	errCode := gjson.Get(string(respBody), "error_code").Int()
	if errCode != 0 {
		errMsg := gjson.Get(string(respBody), "error_message").String()
		if errMsg == "" {
			errMsg = gjson.Get(string(respBody), "err_msg").String()
		}
		if errMsg == "" {
			errMsg = "操作失败"
		}
		return "", fmt.Errorf("API错误 [%d]: %s", errCode, errMsg)
	}

	return string(respBody), nil
}

// ==================== 设置回调 ====================

// SetCallback 设置回调URL
func (c *Client) SetCallback(callbackURL string) error {
	_, err := c.GuidRequest(context.Background(), "/client/set_notify_url", map[string]interface{}{
		"notify_url": callbackURL,
	})
	if err != nil {
		return fmt.Errorf("设置回调失败: %w", err)
	}
	log.Printf("[Client] 设置回调成功: %s", callbackURL)
	return nil
}

// ==================== 发送文本消息 ====================

// SendText 发送文本消息
func (c *Client) SendText(toId, content string) error {
	if c.protocol == ProtocolQiWei {
		// 企微使用 conversation_id
		_, err := c.GuidRequest(context.Background(), "/msg/send_text", map[string]interface{}{
			"conversation_id": toId,
			"content":         content,
		})
		return err
	}
	// 个微使用 to_username
	_, err := c.GuidRequest(context.Background(), "/msg/send_text", map[string]interface{}{
		"to_username": toId,
		"content":     content,
	})
	return err
}

// SendTextWithAt 发送群@消息
func (c *Client) SendTextWithAt(toId, content string, atList []string) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_room_at", map[string]interface{}{
			"conversation_id": toId,
			"content":         content,
			"at_list":         atList,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/send_room_at", map[string]interface{}{
		"to_username": toId,
		"content":     content,
		"at_list":     atList,
	})
	return err
}

// ==================== 发送图片消息 ====================

// SendImage 发送图片消息（需要先上传获取参数）
func (c *Client) SendImage(toId, fileId, aesKey, md5 string, size, width, height uint32) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_image", map[string]interface{}{
			"conversation_id": toId,
			"file_id":         fileId,
			"aes_key":         aesKey,
			"md5":             md5,
			"size":            size,
			"image_width":     width,
			"image_height":    height,
			"is_hd":           false,
		})
		return err
	}
	// 个微
	_, err := c.GuidRequest(context.Background(), "/msg/send_image", map[string]interface{}{
		"to_username":     toId,
		"file_id":         fileId,
		"aes_key":         aesKey,
		"file_md5":        md5,
		"file_size":       size,
		"thumb_width":     width,
		"thumb_height":    height,
		"thumb_file_size": size,
	})
	return err
}

// ==================== 发送文件消息 ====================

// SendFile 发送文件消息
func (c *Client) SendFile(toId, fileId, aesKey, md5, fileName string, size uint32) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_file", map[string]interface{}{
			"conversation_id": toId,
			"file_id":         fileId,
			"aes_key":         aesKey,
			"md5":             md5,
			"size":            size,
			"file_name":       fileName,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/send_file", map[string]interface{}{
		"to_username": toId,
		"file_id":     fileId,
		"aes_key":     aesKey,
		"file_md5":    md5,
		"file_size":   size,
		"file_name":   fileName,
	})
	return err
}

// ==================== 发送视频消息 ====================

// SendVideo 发送视频消息
func (c *Client) SendVideo(toId, fileId, aesKey, md5, fileName string, size, duration, width, height uint32) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_video", map[string]interface{}{
			"conversation_id": toId,
			"file_id":         fileId,
			"aes_key":         aesKey,
			"md5":             md5,
			"size":            size,
			"file_name":       fileName,
			"video_duration":  duration,
			"video_width":     width,
			"video_height":    height,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/send_video", map[string]interface{}{
		"to_username":    toId,
		"file_id":        fileId,
		"aes_key":        aesKey,
		"file_md5":       md5,
		"file_size":      size,
		"video_duration": duration,
		"video_width":    width,
		"video_height":   height,
	})
	return err
}

// ==================== 发送语音消息 ====================

// SendVoice 发送语音消息
func (c *Client) SendVoice(toId, fileId, aesKey, md5 string, size, voiceTime uint32) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_voice", map[string]interface{}{
			"conversation_id": toId,
			"file_id":         fileId,
			"aes_key":         aesKey,
			"md5":             md5,
			"size":            size,
			"voice_time":      voiceTime,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/send_voice", map[string]interface{}{
		"conversation_id": toId,
		"file_id":         fileId,
		"aes_key":         aesKey,
		"md5":             md5,
		"size":            size,
		"voice_time":      voiceTime,
	})
	return err
}

// ==================== 发送小程序消息 ====================

// WeAppInfo 小程序信息
type WeAppInfo struct {
	AppId    string // 小程序AppID
	Title    string // 标题
	Desc     string // 描述
	PagePath string // 页面路径
	UserName string // 小程序原始ID (gh_...)
	AppIcon  string // 图标URL或file_id
	AppName  string // 小程序名称
	FileId   string // 文件ID（CDN上传后获取）
	AesKey   string // AES密钥
	Md5      string // MD5
	Size     uint32 // 文件大小
}

// SendWeApp 发送小程序消息
func (c *Client) SendWeApp(toId string, app *WeAppInfo) error {
	path := "/msg/send_weapp"
	if c.protocol == ProtocolWeiXin {
		path = "/msg/send_mini_app"
	}

	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), path, map[string]interface{}{
			"conversation_id": toId,
			"aes_key":         app.AesKey,
			"appid":           app.AppId,
			"page_path":       app.PagePath,
			"title":           app.Title,
			"md5":             app.Md5,
			"size":            app.Size,
			"username":        app.UserName,
			"appicon":         app.AppIcon,
			"file_id":         app.FileId,
			"appname":         app.AppName,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), path, map[string]interface{}{
		"to_username": toId,
		"aes_key":     app.AesKey,
		"appid":       app.AppId,
		"page_path":   app.PagePath,
		"title":       app.Title,
		"file_md5":    app.Md5,
		"file_size":   app.Size,
		"username":    app.UserName,
		"appicon":     app.AppIcon,
		"file_id":     app.FileId,
		"appname":     app.AppName,
	})
	return err
}

// ==================== 发送链接消息 ====================

// LinkInfo 链接信息
type LinkInfo struct {
	Title       string // 标题
	Description string // 描述
	Url         string // 链接URL
	ImageUrl    string // 图片URL
}

// SendLink 发送链接消息
func (c *Client) SendLink(toId string, link *LinkInfo) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_link", map[string]interface{}{
			"conversation_id": toId,
			"title":           link.Title,
			"description":     link.Description,
			"url":             link.Url,
			"image_url":       link.ImageUrl,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/send_link_card", map[string]interface{}{
		"to_username": toId,
		"title":       link.Title,
		"desc":        link.Description,
		"url":         link.Url,
		"image_url":   link.ImageUrl,
	})
	return err
}

// ==================== 发送名片消息 ====================

// SendCard 发送名片消息
func (c *Client) SendCard(toId, userId string) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_personal_card", map[string]interface{}{
			"conversation_id": toId,
			"user_id":         userId,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/send_share_card", map[string]interface{}{
		"to_username":    toId,
		"share_username": userId,
	})
	return err
}

// ==================== 撤回消息 ====================

// RevokeMessage 撤回消息
func (c *Client) RevokeMessage(toId, msgId string) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/revoke_msg", map[string]interface{}{
			"conversation_id": toId,
			"msgid":           msgId,
		})
		return err
	}
	_, err := c.GuidRequest(context.Background(), "/msg/revoke_msg", map[string]interface{}{
		"to_username":   toId,
		"msg_id":        msgId,
		"client_msg_id": 0,
	})
	return err
}

// ==================== 设置会话已读（个微） ====================

// SetSessionRead 设置会话已读（仅个微）
func (c *Client) SetSessionRead(toUsername string) error {
	if c.protocol == ProtocolWeiXin {
		_, err := c.GuidRequest(context.Background(), "/msg/set_session_read", map[string]interface{}{
			"to_username": toUsername,
		})
		return err
	}
	return nil
}

// ==================== CDN 文件上传/下载 ====================

// C2CUpload C2C文件上传
func (c *Client) C2CUpload(filePath string, fileType uint32) (string, error) {
	return c.GuidRequest(context.Background(), "/cdn/c2c_upload", map[string]interface{}{
		"file_path": filePath,
		"file_type": fileType,
	})
}

// C2CDownload C2C文件下载
func (c *Client) C2CDownload(fileId, aesKey, savePath string, fileType, fileSize uint32) (string, error) {
	return c.GuidRequest(context.Background(), "/cdn/c2c_download", map[string]interface{}{
		"file_id":   fileId,
		"file_type": fileType,
		"aes_key":   aesKey,
		"file_size": fileSize,
		"save_path": savePath,
	})
}

// ==================== 图片URL发送（企微） ====================

// SendImageByURL 通过URL发送图片（企微专用，自动上传）
func (c *Client) SendImageByURL(toId, imageUrl string) error {
	if c.protocol == ProtocolQiWei {
		// 企微支持直接通过URL发送
		_, err := c.GuidRequest(context.Background(), "/msg/send_image", map[string]interface{}{
			"conversation_id": toId,
			"image_url":       imageUrl,
		})
		return err
	}
	return fmt.Errorf("个微暂不支持直接通过URL发送图片，请先上传")
}

// ==================== GIF表情发送（企微） ====================

// SendGifByUrl 通过URL发送GIF（企微专用）
func (c *Client) SendGifByUrl(toId, url, md5 string, width, height uint32) error {
	if c.protocol == ProtocolQiWei {
		_, err := c.GuidRequest(context.Background(), "/msg/send_gif_url", map[string]interface{}{
			"conversation_id": toId,
			"url":             url,
			"md5":             md5,
			"image_width":     width,
			"image_height":    height,
		})
		return err
	}
	return fmt.Errorf("个微暂不支持GIF URL发送")
}
