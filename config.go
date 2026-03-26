package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// ProtocolType 协议类型
type ProtocolType string

const (
	// ProtocolQiWei 企微协议（企业微信）
	ProtocolQiWei ProtocolType = "qiwei"
	// ProtocolWeiXin 个微协议（个人微信）
	ProtocolWeiXin ProtocolType = "weixin"
)

// Config 应用配置
type Config struct {
	// 协议类型
	Protocol ProtocolType // 企微或个微

	// API 配置
	APIURL    string // 后端开放平台 API 地址
	GUID      string // 实例 GUID
	AppKey    string // 应用 AppKey（开放平台认证）
	AppSecret string // 应用 AppSecret（开放平台认证）

	// 回调配置
	CallbackPort   string // 回调服务端口
	CallbackPath   string // 回调路径
	CallbackURL    string // 完整的回调 URL（公网可访问）
	CallbackSecret string // 回调密钥（可选）

	// OpenAI 配置
	OpenAIAPIKey  string // OpenAI API Key
	OpenAIBaseURL string // OpenAI API Base URL（可选，用于代理）
	OpenAIModel   string // 使用的模型

	// CDN 配置（私有化云存储）
	CDNURLQiWei  string // 企微 CDN 服务地址（端口 34789）
	CDNURLWeiXin string // 个微 CDN 服务地址（端口 35789）
}

// AppConfig 全局配置实例
var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 尝试加载 .env 文件
	_ = godotenv.Load(".env")

	// 解析协议类型
	protocol := ProtocolQiWei // 默认企微
	if p := os.Getenv("PROTOCOL"); p != "" {
		if p == "weixin" || p == "wx" {
			protocol = ProtocolWeiXin
		}
	}

	cfg := &Config{
		Protocol:       protocol,
		APIURL:         getEnvOrDefault("API_URL", "http://localhost:8080"),
		GUID:           getEnvOrDefault("GUID", ""),
		AppKey:         getEnvOrDefault("APP_KEY", ""),
		AppSecret:      getEnvOrDefault("APP_SECRET", ""),
		CallbackPort:   getEnvOrDefault("CALLBACK_PORT", "9000"),
		CallbackPath:   getEnvOrDefault("CALLBACK_PATH", "/callback"),
		CallbackURL:    getEnvOrDefault("CALLBACK_URL", "http://localhost:9000/callback"),
		CallbackSecret: getEnvOrDefault("CALLBACK_SECRET", ""),
		OpenAIAPIKey:   getEnvOrDefault("OPENAI_API_KEY", ""),
		OpenAIBaseURL:  getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:    getEnvOrDefault("OPENAI_MODEL", "gpt-3.5-turbo"),
		CDNURLQiWei:    getEnvOrDefault("CDN_URL_QIWEI", "http://8.138.182.93:34789"),
		CDNURLWeiXin:   getEnvOrDefault("CDN_URL_WEIXIN", "http://8.138.182.93:35789"),
	}

	// 验证必要配置
	if cfg.GUID == "" {
		log.Println("警告: GUID 未设置，请在 .env 文件中配置 GUID")
	}
	if cfg.AppKey == "" {
		log.Println("警告: APP_KEY 未设置，请在 .env 文件中配置 APP_KEY")
	}
	if cfg.AppSecret == "" {
		log.Println("警告: APP_SECRET 未设置，请在 .env 文件中配置 APP_SECRET")
	}

	AppConfig = cfg
	return cfg
}

// IsQiWei 是否为企微协议
func (c *Config) IsQiWei() bool {
	return c.Protocol == ProtocolQiWei
}

// IsWeiXin 是否为个微协议
func (c *Config) IsWeiXin() bool {
	return c.Protocol == ProtocolWeiXin
}

// ProtocolName 获取协议名称
func (c *Config) ProtocolName() string {
	if c.IsQiWei() {
		return "企微（企业微信）"
	}
	return "个微（个人微信）"
}

// CDNURL 获取当前协议对应的 CDN URL
func (c *Config) CDNURL() string {
	if c.IsQiWei() {
		return c.CDNURLQiWei
	}
	return c.CDNURLWeiXin
}

// HasCDN 是否配置了 CDN
func (c *Config) HasCDN() bool {
	return c.CDNURLQiWei != "" || c.CDNURLWeiXin != ""
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
