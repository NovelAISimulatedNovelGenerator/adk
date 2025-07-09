package main

import (
	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
	novel_v3 "github.com/nvcnvn/adk-golang/pkg/flows/novel_v3"
)

// pluginImpl 实现 flow.FlowPlugin。
type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "novel_flow_v3" }

func (p *pluginImpl) Build() (*agents.Agent, error) { return novel_v3.Build(), nil }

// Plugin 导出符号，供 plugin.Open 查找。
var Plugin flow.FlowPlugin = &pluginImpl{}
