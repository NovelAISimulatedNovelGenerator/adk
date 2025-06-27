package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

// ComplexPlugin 实现一个更复杂的插件，支持动态配置
type ComplexPlugin struct {
	sync.RWMutex
	name    string
	config  map[string]interface{}
	builder func() (*agents.Agent, error)
}

// Name 实现 FlowPlugin 接口
func (p *ComplexPlugin) Name() string {
	p.RLock()
	defer p.RUnlock()
	return p.name
}

// Build 实现 FlowPlugin 接口
func (p *ComplexPlugin) Build() (*agents.Agent, error) {
	p.RLock()
	defer p.RUnlock()
	
	// 记录构建信息
	log.Printf("[高级插件] 构建工作流 %s，配置项 %d 个", p.name, len(p.config))
	
	if p.builder != nil {
		return p.builder()
	}
	
	// 默认创建一个简单的 Agent
	agent := agents.NewAgent(
		agents.WithName(fmt.Sprintf("dynamic_%s", p.name)),
		agents.WithModel("deepseek-chat"),
		agents.WithInstruction("这是一个动态创建的高级插件，支持运行时配置"),
	)
	
	return agent, nil
}

// SetConfig 设置运行时配置（扩展接口）
func (p *ComplexPlugin) SetConfig(key string, value interface{}) {
	p.Lock()
	defer p.Unlock()
	p.config[key] = value
}

// 创建插件实例并设置一些默认值
var pluginInstance = &ComplexPlugin{
	name:   "advanced_flow",
	config: make(map[string]interface{}),
	builder: func() (*agents.Agent, error) {
		// 构造一个简单的代理链
		mainAgent := agents.NewAgent(
			agents.WithName("main_agent"),
			agents.WithModel("deepseek-chat"),
			agents.WithInstruction("这是高级插件中的主要代理。"),
		)
		
		helpAgent := agents.NewAgent(
			agents.WithName("helper_agent"),
			agents.WithModel("deepseek-chat"),
			agents.WithInstruction("这是一个辅助代理。"),
		)
		
		seqAgent := agents.NewSequentialAgent(agents.SequentialAgentConfig{
			Name:        "advanced_flow",
			Description: "高级测试工作流",
			SubAgents:   []*agents.Agent{mainAgent, helpAgent},
		})
		// 将 SequentialAgent 转换为 Agent
		return &seqAgent.Agent, nil
	},
}

// 导出插件符号
var Plugin flow.FlowPlugin = pluginInstance
