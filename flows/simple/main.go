package main

import (
	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

// 简单工作流插件，用于测试API与CLI一致性
// 特别关注上下文取消与超时机制

// pluginImpl 实现 flow.FlowPlugin 接口
type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "simple_flow" }

func (p *pluginImpl) Build() (*agents.Agent, error) {
	// 创建一个简单的代理，使用deepseek-chat模型
	simpleAgent := agents.NewAgent(
		agents.WithName("simple_agent"),
		agents.WithModel("deepseek-chat"),
		agents.WithInstruction("你是一个简单的测试代理。收到用户输入后，简要回应并确认收到。"),
		agents.WithDescription("简单测试代理"),
	)

	return simpleAgent, nil
}

// Plugin 导出符号，供 plugin.Open 查找
var Plugin flow.FlowPlugin = &pluginImpl{}
