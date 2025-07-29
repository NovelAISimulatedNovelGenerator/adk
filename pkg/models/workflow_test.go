package models

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/config"
)

// TestWorkflowPool 模型池测试函数框架
func TestWorkflowPool(t *testing.T) {
	const url = "http://localhost:3000" // 移除结尾的斜杠，避免双斜杠问题
	const apikey = "sk-rac1XoSpt3eESULMNGKxAvBQq2WwcqIoSJMhsg2ubOU6tiJQ"
	// 设置环境变量
	os.Setenv("CUSTOM_API_KEY", apikey)
	os.Setenv("CUSTOM_API_ENDPOINT", url)
	defer func() {
		os.Unsetenv("CUSTOM_API_KEY")
		os.Unsetenv("CUSTOM_API_ENDPOINT")
	}()

	// 创建测试配置
	testConfig := &config.Config{
		ModelAPIPools: map[string]config.ModelPoolConfig{
			"test_pool": {
				Base: "glm-4.5",
				Endpoints: []config.EndpointConfig{
					{URL: url, APIKey: apikey},
				},
			},
		},
	}

	// 注册模型池
	err := RegisterModelPools(testConfig)
	if err != nil {
		t.Fatalf("注册模型池失败: %v", err)
	}

	// 获取模型
	registry := GetEnhancedRegistry()
	model, err := registry.GetModel("pool:test_pool")
	if err != nil {
		t.Fatalf("获取模型失败: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// 测试实际的API端点
	apiEndpoint := url + "/v1/chat/completions"
	t.Logf("正在测试API端点: %s", apiEndpoint)

	// 手动测试HTTP请求
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		t.Fatalf("无法连接到API端点: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	t.Logf("服务器响应状态: %s", resp.Status)
	t.Logf("服务器响应内容: %s", string(body)[:min(len(body), 200)])

	messages := []Message{{Role: "user", Content: "hi"}}
	response, err := model.Generate(ctx, messages)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}
	t.Logf("生成响应: %s", response)

}
