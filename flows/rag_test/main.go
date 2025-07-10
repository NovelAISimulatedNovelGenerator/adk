package main

import (
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
    testrag "github.com/nvcnvn/adk-golang/pkg/flows/test_rag_flow"
)

// ragTestPlugin implements the flow.FlowPlugin interface referencing the framework in pkg/flows.
type ragTestPlugin struct{}

// Name returns the unique workflow name.
func (p *ragTestPlugin) Name() string { return "rag_test_flow" }

// Build constructs the root Agent by delegating to the shared framework implementation.
func (p *ragTestPlugin) Build() (*agents.Agent, error) {
    agent := testrag.Build()
    return agent, nil
}

// Plugin exported symbol for plugin.Open lookup.
var Plugin flow.FlowPlugin = &ragTestPlugin{}
