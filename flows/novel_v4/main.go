package main

// main.go 将 novel_v4 工作流封装成符合 flow.FlowPlugin 接口的插件，
// 便于通过 `go build -buildmode=plugin` 生成 .so 文件并由 ADK 动态加载。
// 结构参考 flows/novel_v3/main.go。

import (
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
    novel_v4 "github.com/nvcnvn/adk-golang/pkg/flows/novel_v4"
)

// pluginImpl 实现 flow.FlowPlugin 接口。
type pluginImpl struct{}

// Name 返回插件名称（必须唯一）。
func (p *pluginImpl) Name() string { return "novel_flow_v4" }

// Build 构建顶层 Agent。
func (p *pluginImpl) Build() (*agents.Agent, error) {
    return novel_v4.Build(), nil
}

// Plugin 变量供 plugin.Open 查找。
var Plugin flow.FlowPlugin = &pluginImpl{}
