package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// OpenAIHandler OpenAI 处理器
type OpenAIHandler struct {
	client  *openai.Client
	model   string
	enabled bool

	// 对话历史管理（简单的内存存储）
	history     map[string][]openai.ChatCompletionMessage
	historyLock sync.RWMutex
	maxHistory  int // 最大历史消息数
}

// NewOpenAIHandler 创建 OpenAI 处理器
func NewOpenAIHandler(apiKey, baseURL, model string) *OpenAIHandler {
	if apiKey == "" {
		log.Println("[OpenAI] API Key 未配置，自动回复功能禁用")
		return &OpenAIHandler{enabled: false}
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		config.BaseURL = baseURL
	}

	return &OpenAIHandler{
		client:     openai.NewClientWithConfig(config),
		model:      model,
		enabled:    true,
		history:    make(map[string][]openai.ChatCompletionMessage),
		maxHistory: 10,
	}
}

// IsEnabled 检查是否启用
func (h *OpenAIHandler) IsEnabled() bool {
	return h.enabled
}

// GenerateReply 生成回复
func (h *OpenAIHandler) GenerateReply(userMessage, userId string) (string, error) {
	if !h.enabled {
		return "", nil
	}

	ctx := context.Background()

	// 获取用户对话历史
	h.historyLock.Lock()
	messages := h.history[userId]

	// 添加系统提示
	if len(messages) == 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你是一个友好的助手，帮助用户回答问题。请用简洁的中文回复。",
		})
	}

	// 添加用户消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	// 调用 OpenAI API
	resp, err := h.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    h.model,
		Messages: messages,
	})
	if err != nil {
		h.historyLock.Unlock()
		return "", fmt.Errorf("OpenAI API 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		h.historyLock.Unlock()
		return "", fmt.Errorf("OpenAI 返回空响应")
	}

	reply := resp.Choices[0].Message.Content

	// 添加助手回复到历史
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: reply,
	})

	// 限制历史消息数量
	if len(messages) > h.maxHistory*2+1 {
		// 保留系统消息 + 最近的消息
		messages = append(messages[:1], messages[len(messages)-h.maxHistory*2:]...)
	}

	// 保存历史
	h.history[userId] = messages
	h.historyLock.Unlock()

	return strings.TrimSpace(reply), nil
}

// ClearHistory 清除用户对话历史
func (h *OpenAIHandler) ClearHistory(userId string) {
	h.historyLock.Lock()
	defer h.historyLock.Unlock()
	delete(h.history, userId)
}

// ClearAllHistory 清除所有对话历史
func (h *OpenAIHandler) ClearAllHistory() {
	h.historyLock.Lock()
	defer h.historyLock.Unlock()
	h.history = make(map[string][]openai.ChatCompletionMessage)
}

// GenerateCustomReply 使用自定义提示生成回复
func (h *OpenAIHandler) GenerateCustomReply(userMessage, systemPrompt string) (string, error) {
	if !h.enabled {
		return "", nil
	}

	ctx := context.Background()

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	resp, err := h.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    h.model,
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI 返回空响应")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
