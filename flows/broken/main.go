package main

import (
	"errors"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

// 故意设计错误的插件，测试系统容错性
type brokenPlugin struct{}

func (p *brokenPlugin) Name() string { return "broken_flow" }

func (p *brokenPlugin) Build() (*agents.Agent, error) {
	// 故意返回错误
	return nil, errors.New("这是一个故意设计的错误，用于测试系统恢复能力")
}

// 导出插件符号
var Plugin flow.FlowPlugin = &brokenPlugin{}
