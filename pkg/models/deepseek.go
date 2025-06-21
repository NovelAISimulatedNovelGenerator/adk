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
//
// DeepSeek 模型集成。该实现基于 DeepSeek Chat Completion HTTP API，
// 语法基本与 OpenAI ChatCompletion API 保持一致（如有差异请根据官方文档调整）。
// 目前仅实现非流式 Generate，流式接口返回未实现错误。

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

const (
	defaultDeepSeekAPIEndpoint = "https://api.deepseek.com/v1"
)

// DeepSeekModel implements the Model interface for DeepSeek Chat models.
type DeepSeekModel struct {
	BaseModel
	apiKey   string
	endpoint string
	client   *http.Client
}

// deepSeekChatMessage mirrors the OpenAI/DeepSeek chat message format.
type deepSeekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// deepSeekChatRequest represents a chat completion request.
type deepSeekChatRequest struct {
	Model    string                `json:"model"`
	Messages []deepSeekChatMessage `json:"messages"`
}

// deepSeekChatResponse mirrors the expected response structure.
type deepSeekChatResponse struct {
	Choices []struct {
		Message deepSeekChatMessage `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewDeepSeekModel creates a new DeepSeekModel.
func NewDeepSeekModel(modelName string) (Model, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, errors.New("DEEPSEEK_API_KEY environment variable not set")
	}

	endpoint := os.Getenv("DEEPSEEK_API_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultDeepSeekAPIEndpoint
	}

	return &DeepSeekModel{
		BaseModel: BaseModel{name: modelName},
		apiKey:    apiKey,
		endpoint:  endpoint,
		client:    &http.Client{},
	}, nil
}

// Generate implements the Model interface (non-streaming).
func (m *DeepSeekModel) Generate(ctx context.Context, messages []Message) (string, error) {
	// Convert to DeepSeek chat messages format.
	chatMsgs := make([]deepSeekChatMessage, len(messages))
	for i, msg := range messages {
		chatMsgs[i] = deepSeekChatMessage{Role: msg.Role, Content: msg.Content}
	}

	reqBody, err := json.Marshal(deepSeekChatRequest{
		Model:    m.name,
		Messages: chatMsgs,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/chat/completions", m.endpoint)
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
		return "", fmt.Errorf("DeepSeek API error: %s - %s", resp.Status, string(body))
	}

	var result deepSeekChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Error.Message) > 0 {
		return "", errors.New(result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no choices in DeepSeek response")
	}

	return result.Choices[0].Message.Content, nil
}

// GenerateStream returns not implemented.
func (m *DeepSeekModel) GenerateStream(ctx context.Context, messages []Message) (chan StreamedResponse, error) {
	return nil, errors.New("DeepSeek streaming not implemented")
}

// Register DeepSeek patterns at init.
func init() {
	registry := GetEnhancedRegistry()
	patterns := []string{`deepseek-.*`}
	for _, p := range patterns {
		if err := registry.RegisterPattern(p, func(modelName string) (Model, error) {
			return NewDeepSeekModel(modelName)
		}); err != nil {
			fmt.Printf("Error registering DeepSeek pattern %s: %v\n", p, err)
		}
	}
}
