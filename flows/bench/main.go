package main

import (
    "context"

    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
)

// bench_flow 插件用于并发测试，不调用任何外部模型 API，
// 仅快速返回固定字符串，便于压测调度器 & API 性能。

type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "bench_flow" }

func (p *pluginImpl) Build() (*agents.Agent, error) {
    echoAgent := agents.NewAgent(
        agents.WithName("bench_agent"),
        agents.WithInstruction("并发测试 Echo Agent"),
        agents.WithDescription("仅用于并发测试，快速返回 OK"),
        // 使用 BeforeAgentCallback 跳过模型调用，直接返回固定响应。
        agents.WithBeforeAgentCallback(func(ctx context.Context, message string) (string, bool) {
            return "OK", true
        }),
    )

    return echoAgent, nil
}

// Plugin 为导出符号，供 plugin.Open 查找
var Plugin flow.FlowPlugin = &pluginImpl{}
