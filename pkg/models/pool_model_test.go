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

package models

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/config"
	"github.com/stretchr/testify/assert"
)

// 测试模型池的负载均衡功能
func TestPoolModelLoadBalancing(t *testing.T) {
	// 创建两个测试服务器，模拟两个API端点
	var server1Count, server2Count int32
	
	// 第一个测试服务器
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&server1Count, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"choices":[{"message":{"role":"assistant","content":"来自服务器1的回复"}}]}`)
	}))
	defer server1.Close()
	
	// 第二个测试服务器
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&server2Count, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"choices":[{"message":{"role":"assistant","content":"来自服务器2的回复"}}]}`)
	}))
	defer server2.Close()
	
	// 创建模型池
	endpoints := []config.EndpointConfig{
		{URL: server1.URL, APIKey: "test-key-1"},
		{URL: server2.URL, APIKey: "test-key-2"},
	}
	
	poolModel, err := NewPoolModel("pool:test_pool", "deepseek", endpoints)
	assert.NoError(t, err)
	assert.NotNil(t, poolModel)
	
	// 发送多个并发请求测试负载均衡
	var wg sync.WaitGroup
	numRequests := 100
	
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			
			messages := []Message{
				{Role: "user", Content: fmt.Sprintf("测试请求 %d", idx)},
			}
			
			_, err := poolModel.Generate(ctx, messages)
			assert.NoError(t, err)
		}(i)
	}
	
	wg.Wait()
	
	// 检查请求是否大致均匀分布到两个服务器
	t.Logf("服务器1请求数: %d, 服务器2请求数: %d", server1Count, server2Count)
	
	// 允许一定误差范围，但请求应该大致均匀分布
	expectedPerServer := numRequests / 2
	tolerance := numRequests / 10 // 10% 容差
	
	assert.InDelta(t, expectedPerServer, server1Count, float64(tolerance), "服务器1应收到大约一半的请求")
	assert.InDelta(t, expectedPerServer, server2Count, float64(tolerance), "服务器2应收到大约一半的请求")
}

// 测试模型池配置解析和注册
func TestRegisterModelPools(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		ModelAPIPools: map[string]config.ModelPoolConfig{
			"test_pool": {
				Base: "deepseek",
				Endpoints: []config.EndpointConfig{
					{URL: "http://api1.example.com", APIKey: "key1"},
					{URL: "http://api2.example.com", APIKey: "key2"},
				},
			},
		},
	}
	
	// 注册模型池
	err := RegisterModelPools(cfg)
	assert.NoError(t, err)
	
	// 检查模型池是否已注册
	registry := GetEnhancedRegistry()
	patterns := registry.ListPatterns()
	
	found := false
	for _, pattern := range patterns {
		if pattern == "^pool:test_pool$" {
			found = true
			break
		}
	}
	
	assert.True(t, found, "模型池应已注册")
	
	// 尝试获取模型池
	model, err := registry.GetModel("pool:test_pool")
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "pool:test_pool", model.Name())
}
