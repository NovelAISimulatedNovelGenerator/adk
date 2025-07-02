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
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/nvcnvn/adk-golang/pkg/config"
)

// PoolModel 实现 Model 接口，提供负载均衡能力
type PoolModel struct {
	BaseModel
	models []Model // 底层模型实例列表
	next   uint32  // 原子计数器用于轮询
}

// NewPoolModel 创建新的模型池实例
func NewPoolModel(name string, baseModel string, endpoints []config.EndpointConfig) (*PoolModel, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("至少需要一个API端点配置")
	}

	models := make([]Model, 0, len(endpoints))
	
	for i, ep := range endpoints {
		var model Model
		
		// 根据baseModel创建对应类型的模型实例
		switch baseModel {
		case "deepseek":
			// 创建DeepSeek模型，使用自定义端点和API Key
			model = &DeepSeekModel{
				BaseModel: BaseModel{name: fmt.Sprintf("%s-%d", name, i)},
				apiKey:    ep.APIKey,
				endpoint:  ep.URL,
				client:    &http.Client{},
			}
		case "gemini":
			// TODO: 补充Gemini模型实例化逻辑，目前示例代码
			return nil, fmt.Errorf("暂不支持 Gemini 模型池: %s", baseModel)
		default:
			return nil, fmt.Errorf("不支持的模型类型: %s", baseModel)
		}
		
		models = append(models, model)
		log.Printf("为池 %s 添加模型端点 %s", name, ep.URL)
	}
	
	return &PoolModel{
		BaseModel: BaseModel{name: name},
		models:    models,
		next:      0,
	}, nil
}

// Generate 实现 Model 接口，带负载均衡功能
func (m *PoolModel) Generate(ctx context.Context, messages []Message) (string, error) {
	if len(m.models) == 0 {
		return "", errors.New("模型池为空")
	}
	
	// 原子操作选择下一个模型，实现简单的轮询负载均衡
	nextIndex := atomic.AddUint32(&m.next, 1) % uint32(len(m.models))
	model := m.models[nextIndex]
	
	log.Printf("池 %s 选择端点 %d 生成响应", m.name, nextIndex)
	return model.Generate(ctx, messages)
}

// GenerateStream 实现 Model 接口的流式生成方法，带负载均衡功能
func (m *PoolModel) GenerateStream(ctx context.Context, messages []Message) (chan StreamedResponse, error) {
	if len(m.models) == 0 {
		return nil, errors.New("模型池为空")
	}
	
	// 原子操作选择下一个模型，实现简单的轮询负载均衡
	nextIndex := atomic.AddUint32(&m.next, 1) % uint32(len(m.models))
	model := m.models[nextIndex]
	
	log.Printf("池 %s 选择端点 %d 流式生成响应", m.name, nextIndex)
	return model.GenerateStream(ctx, messages)
}

// RegisterModelPools 注册配置文件中定义的所有模型池
func RegisterModelPools(cfg *config.Config) error {
	if cfg.ModelAPIPools == nil || len(cfg.ModelAPIPools) == 0 {
		log.Println("未配置模型API池")
		return nil
	}
	
	registry := GetEnhancedRegistry()
	
	for name, poolCfg := range cfg.ModelAPIPools {
		poolName := "pool:" + name
		
		// 为每个池创建模型并注册工厂函数
		pattern := fmt.Sprintf("^%s$", poolName)
		
		err := registry.RegisterPattern(pattern, func(modelName string) (Model, error) {
			return NewPoolModel(poolName, poolCfg.Base, poolCfg.Endpoints)
		})
		
		if err != nil {
			log.Printf("注册模型池 %s 失败: %v", poolName, err)
			continue
		}
		
		log.Printf("成功注册模型池 %s (基于 %s, %d 个端点)", poolName, poolCfg.Base, len(poolCfg.Endpoints))
	}
	
	return nil
}
