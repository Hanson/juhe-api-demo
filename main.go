package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 命令行参数
	action := flag.String("action", "start", "操作类型: start, set-callback, send-text, send-image, send-weapp, send-file, send-link")
	toId := flag.String("to", "", "接收者ID（企微:conversation_id, 个微:to_username）")
	content := flag.String("content", "", "消息内容")
	imageURL := flag.String("image-url", "", "图片URL（公网可访问）")
	_ = flag.String("file", "", "本地文件路径（用于上传）") // filePath 保留用于将来扩展
	// 小程序参数
	appId := flag.String("app-id", "", "小程序AppID")
	appTitle := flag.String("title", "", "小程序标题")
	appDesc := flag.String("desc", "", "小程序描述/链接描述")
	appPage := flag.String("page", "", "小程序页面路径")
	appIcon := flag.String("icon", "", "小程序图标URL")
	// 链接参数
	linkUrl := flag.String("url", "", "链接URL")

	flag.Parse()

	// 加载配置
	cfg := LoadConfig()
	if cfg.GUID == "" && *action != "start" {
		log.Fatal("错误: 请在 .env 文件中配置 GUID")
	}

	// 创建客户端
	client := NewClient(cfg.APIURL, cfg.GUID, cfg.AppKey, cfg.AppSecret, cfg.Protocol)

	switch *action {
	case "start":
		// 启动回调服务
		startServer(cfg, client)

	case "set-callback":
		// 设置回调URL
		if err := client.SetCallback(cfg.CallbackURL); err != nil {
			log.Fatalf("设置回调失败: %v", err)
		}
		log.Println("回调设置成功")

	case "send-text":
		// 发送文本消息
		if *toId == "" || *content == "" {
			log.Fatal("错误: 请指定 -to 和 -content 参数")
		}
		if err := client.SendText(*toId, *content); err != nil {
			log.Fatalf("发送文本消息失败: %v", err)
		}
		log.Println("文本消息发送成功")

	case "send-image":
		// 发送图片消息（通过URL，仅企微支持）
		if *toId == "" || *imageURL == "" {
			log.Fatal("错误: 请指定 -to 和 -image-url 参数")
		}
		if err := client.SendImageByURL(*toId, *imageURL); err != nil {
			log.Fatalf("发送图片消息失败: %v", err)
		}
		log.Println("图片消息发送成功")

	case "send-weapp":
		// 发送小程序消息
		if *toId == "" || *appId == "" {
			log.Fatal("错误: 请指定 -to 和 -app-id 参数")
		}
		app := &WeAppInfo{
			AppId:    *appId,
			Title:    *appTitle,
			AppName:  *appDesc, // 使用 desc 作为 appname
			PagePath: *appPage,
			AppIcon:  *appIcon,
		}
		if err := client.SendWeApp(*toId, app); err != nil {
			log.Fatalf("发送小程序消息失败: %v", err)
		}
		log.Println("小程序消息发送成功")

	case "send-link":
		// 发送链接消息
		if *toId == "" || *linkUrl == "" {
			log.Fatal("错误: 请指定 -to 和 -url 参数")
		}
		link := &LinkInfo{
			Title:       *appTitle,
			Description: *appDesc,
			Url:         *linkUrl,
			ImageUrl:    *appIcon,
		}
		if err := client.SendLink(*toId, link); err != nil {
			log.Fatalf("发送链接消息失败: %v", err)
		}
		log.Println("链接消息发送成功")

	default:
		log.Fatalf("未知操作: %s", *action)
	}
}

// startServer 启动回调服务器
func startServer(cfg *Config, client *Client) {
	log.Println("========================================")
	log.Println("    聚合聊天 - 后端开放平台 Demo")
	log.Println("========================================")
	log.Printf("协议类型: %s", cfg.ProtocolName())
	log.Printf("API 地址: %s", cfg.APIURL)
	log.Printf("实例 GUID: %s", cfg.GUID)
	log.Printf("回调地址: %s", cfg.CallbackURL)
	log.Printf("OpenAI: %s", boolStr(cfg.OpenAIAPIKey != "", "已配置", "未配置"))
	log.Printf("私有化CDN: %s", boolStr(cfg.CDNURL() != "", "已配置", "未配置"))
	log.Println("========================================")

	// 创建 OpenAI 处理器
	var openaiClient *OpenAIHandler
	if cfg.OpenAIAPIKey != "" {
		openaiClient = NewOpenAIHandler(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)
		log.Println("[OpenAI] 自动回复功能已启用")
	}

	// 创建下载目录
	if err := os.MkdirAll("./downloads", 0755); err != nil {
		log.Printf("警告: 创建下载目录失败: %v", err)
	}

	// 设置回调
	log.Println("[Callback] 正在设置回调地址...")
	if err := client.SetCallback(cfg.CallbackURL); err != nil {
		log.Printf("警告: 设置回调失败: %v", err)
		log.Println("请确保回调地址可以从公网访问，或手动设置回调")
	}

	// 创建并启动回调服务器
	callbackServer := NewCallbackServer(cfg.CallbackPort, cfg.CallbackPath, cfg.CallbackSecret, client, openaiClient)

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("\n[Server] 正在关闭服务器...")
		os.Exit(0)
	}()

	// 启动服务器
	log.Println("[Server] 回调服务器启动中...")
	if err := callbackServer.Start(); err != nil {
		log.Fatalf("启动回调服务器失败: %v", err)
	}
}

// boolStr 布尔值转字符串
func boolStr(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}
