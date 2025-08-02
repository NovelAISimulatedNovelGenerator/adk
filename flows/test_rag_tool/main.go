package main

import (
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
    test_rag_tool "github.com/nvcnvn/adk-golang/pkg/flows/test_rag_tool"
)

type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "test_rag_tool_flow" }

func (p *pluginImpl) Build() (*agents.Agent, error) { return test_rag_tool.Build(), nil }

var Plugin flow.FlowPlugin = &pluginImpl{}
