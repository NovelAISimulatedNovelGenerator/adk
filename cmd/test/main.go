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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/config"
	"github.com/nvcnvn/adk-golang/pkg/models"
)

var (
	configFile = flag.String("config", "", "配置文件路径，也可通过 ADK_CONFIG 环境变量指定")
	poolName   = flag.String("pool", "", "要测试的模型池名称(如 pool:deepseek_pool)")
	concurrent = flag.Int("concurrent", 10, "并发请求数量")
	timeout    = flag.Int("timeout", 30, "请求超时时间(秒)")
)

// 模拟API服务器，用于测试
func startMockServer(serverID int) *httptest.Server {
	var requestCount int32
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		log.Printf("服务器 %d 收到第 %d 个请求", serverID, count)
		
		// 模拟处理延时
		time.Sleep(100 * time.Millisecond)
		
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":"来自服务器%d的第%d个响应"}}]}`, serverID, count)
	}))
	
	log.Printf("模拟服务器 %d 已启动: %s", serverID, server.URL)
	return server
}

// 测试现有配置中的模型池
func testExistingPool(name string, concurrent int, timeoutSeconds int) {
	registry := models.GetEnhancedRegistry()
	
	log.Printf("获取模型: %s", name)
	model, err := registry.GetModel(name)
	if err != nil {
		log.Fatalf("获取模型失败: %v", err)
	}
	
	log.Printf("成功获取模型: %s", model.Name())
	
	var wg sync.WaitGroup
	var successCount, failCount int32
	
	log.Printf("开始执行 %d 个并发请求...", concurrent)
	startTime := time.Now()
	
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
			defer cancel()
			
			messages := []models.Message{
				{Role: "user", Content: fmt.Sprintf("测试请求 %d", idx)},
			}
			
			log.Printf("发送请求 %d", idx)
			response, err := model.Generate(ctx, messages)
			
			if err != nil {
				log.Printf("请求 %d 失败: %v", idx, err)
				atomic.AddInt32(&failCount, 1)
				return
			}
			
			log.Printf("请求 %d 成功: %s", idx, response)
			atomic.AddInt32(&successCount, 1)
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(startTime)
	
	log.Printf("========测试结果========")
	log.Printf("总请求数: %d", concurrent)
	log.Printf("成功请求: %d", successCount)
	log.Printf("失败请求: %d", failCount)
	log.Printf("总耗时: %v", elapsed)
	log.Printf("平均每请求耗时: %v", elapsed/time.Duration(concurrent))
	log.Printf("======================")
}

// 全局服务器变量，确保测试期间不会关闭
var (
	testServer1 *httptest.Server
	testServer2 *httptest.Server
	testServer3 *httptest.Server
)

// 创建测试用的模型池配置
func createTestPoolConfig() *config.Config {
	// 启动3个模拟服务器
	testServer1 = startMockServer(1)
	testServer2 = startMockServer(2)
	testServer3 = startMockServer(3)
	
	// 创建配置
	cfg := &config.Config{
		ModelAPIPools: map[string]config.ModelPoolConfig{
			"test_pool": {
				Base: "deepseek",
				Endpoints: []config.EndpointConfig{
					{URL: testServer1.URL, APIKey: "test-key-1"},
					{URL: testServer2.URL, APIKey: "test-key-2"},
					{URL: testServer3.URL, APIKey: "test-key-3"},
				},
			},
		},
	}
	
	// 注册模型池
	err := models.RegisterModelPools(cfg)
	if err != nil {
		log.Fatalf("注册模型池失败: %v", err)
	}
	
	return cfg
}

func main() {
	flag.Parse()
	
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("启动模型池测试程序...")
	
	// 在程序结束时关闭测试服务器
	defer func() {
		if testServer1 != nil {
			testServer1.Close()
		}
		if testServer2 != nil {
			testServer2.Close()
		}
		if testServer3 != nil {
			testServer3.Close()
		}
	}()
	
	if *configFile != "" {
		// 从配置文件加载
		configPath := *configFile
		if configPath == "" {
			configPath = os.Getenv("ADK_CONFIG")
			if configPath == "" {
				configPath = "config.yaml"
			}
		}
		
		log.Printf("加载配置: %s", configPath)
		cfg, err := config.Load(configPath)
		if err != nil {
			log.Fatalf("加载配置失败: %v", err)
		}
		
		// 注册模型池
		err = models.RegisterModelPools(cfg)
		if err != nil {
			log.Fatalf("注册模型池失败: %v", err)
		}
		
		log.Println("成功注册配置文件中的模型池")
		
		if *poolName != "" {
			// 测试指定的模型池
			testExistingPool(*poolName, *concurrent, *timeout)
		} else {
			log.Println("请使用 -pool 参数指定要测试的模型池名称")
		}
	} else {
		// 创建测试模型池
		log.Println("创建测试用模型池配置...")
		_ = createTestPoolConfig()
		
		// 测试模型池
		testExistingPool("pool:test_pool", *concurrent, *timeout)
	}
	
	log.Println("测试程序结束")
}
