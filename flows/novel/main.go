package main

import (
	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
	novel "github.com/nvcnvn/adk-golang/pkg/flows/novel"
)

// pluginImpl 实现 flow.FlowPlugin，名称与 default_flow 对应。
type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "novel_flow_v1" }

func (p *pluginImpl) Build() (*agents.Agent, error) { return novel.Build(), nil }

// Plugin 导出符号 (指针)，供 plugin.Open 查找。
var Plugin flow.FlowPlugin = &pluginImpl{}
