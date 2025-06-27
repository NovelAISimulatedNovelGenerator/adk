package main

import (
	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
	novel "github.com/nvcnvn/adk-golang/pkg/flows/novel"
)

// pluginImpl 实现 flow.FlowPlugin，名称与 default_flow 对应。
type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "novel_flow_v2" }

func (p *pluginImpl) Build() (*agents.Agent, error) {
	// v2 版本仍然复用基础构建函数，但可以后期扩展
	agent := novel.Build()
	// 可以对 agent 进行额外配置
	return agent, nil
}

// Plugin 导出符号，供 plugin.Open 查找。
var Plugin flow.FlowPlugin = &pluginImpl{}
