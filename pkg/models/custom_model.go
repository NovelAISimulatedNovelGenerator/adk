// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 自定义模型集成。该实现提供了一个通用的HTTP API接口，
// 支持符合OpenAI ChatCompletion API格式的任意第三方模型服务。
// 可用于集成各种兼容OpenAI API格式的模型服务。

package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// CustomModel 实现 Model 接口，用于自定义的第三方模型服务
type CustomModel struct {
	BaseModel
	apiKey     string
	endpoint   string
	client     *http.Client
	actualModel string // 发送给API的实际模型名称
}

// customChatMessage 表示聊天消息格式（兼容OpenAI格式）
type customChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// customChatRequest 表示聊天完成请求结构
type customChatRequest struct {
	Model    string              `json:"model"`
	Messages []customChatMessage `json:"messages"`
}

// customChatResponse 表示聊天完成响应结构
type customChatResponse struct {
	Choices []struct {
		Message customChatMessage `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewCustomModel 创建新的自定义模型实例
func NewCustomModel(modelName string) (Model, error) {
	apiKey := os.Getenv("CUSTOM_API_KEY")
	if apiKey == "" {
		return nil, errors.New("CUSTOM_API_KEY 环境变量未设置")
	}

	endpoint := os.Getenv("CUSTOM_API_ENDPOINT")
	if endpoint == "" {
		return nil, errors.New("CUSTOM_API_ENDPOINT 环境变量未设置")
	}

	return &CustomModel{
		BaseModel:   BaseModel{name: modelName},
		apiKey:      apiKey,
		endpoint:    endpoint,
		client:      &http.Client{},
		actualModel: modelName, // 默认使用相同的模型名称
	}, nil
}

// NewCustomModelWithActualName 创建新的自定义模型实例，允许指定实际模型名称
func NewCustomModelWithActualName(instanceName, actualModelName, apiKey, endpoint string) (Model, error) {
	if apiKey == "" {
		return nil, errors.New("未提供API密钥")
	}

	if endpoint == "" {
		return nil, errors.New("未提供API端点")
	}

	return &CustomModel{
		BaseModel:   BaseModel{name: instanceName},
		apiKey:      apiKey,
		endpoint:    endpoint,
		client:      &http.Client{},
		actualModel: actualModelName, // 使用指定的实际模型名称
	}, nil
}

// Generate 实现 Model 接口的生成方法
func (m *CustomModel) Generate(ctx context.Context, messages []Message) (string, error) {
	// 转换为自定义聊天消息格式
	chatMsgs := make([]customChatMessage, len(messages))
	for i, msg := range messages {
		chatMsgs[i] = customChatMessage{Role: msg.Role, Content: msg.Content}
	}

	reqBody, err := json.Marshal(customChatRequest{
		Model:    m.actualModel, // 使用实际模型名称，而不是实例名称
		Messages: chatMsgs,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/chat/completions", m.endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("自定义模型API错误: %s - %s", resp.Status, string(body))
	}

	var result customChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Error.Message) > 0 {
		return "", errors.New(result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("自定义模型响应中没有选择项")
	}

	return result.Choices[0].Message.Content, nil
}

// GenerateStream 实现 Model 接口的流式生成方法
func (m *CustomModel) GenerateStream(ctx context.Context, messages []Message) (chan StreamedResponse, error) {
	//TODO:
	return nil, errors.New("自定义模型暂不支持流式生成")
}

// init 函数注册自定义模型模式
func init() {
	registry := GetEnhancedRegistry()

	// 注册 custom 模式
	err := registry.RegisterPattern("^custom$", func(modelName string) (Model, error) {
		return NewCustomModel(modelName)
	})
	if err != nil {
		panic(fmt.Sprintf("注册自定义模型模式失败: %v", err))
	}

	// 注册 custom:* 模式，支持自定义模型名称
	err = registry.RegisterPattern("^custom:.+$", func(modelName string) (Model, error) {
		return NewCustomModel(modelName)
	})
	if err != nil {
		panic(fmt.Sprintf("注册自定义模型通配符模式失败: %v", err))
	}
}
