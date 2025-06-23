package cli

// flow_agent.go 支持从 flow.Manager 中运行工作流插件

import (
	"context"
	"fmt"
	"os"

	//"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
	"github.com/nvcnvn/adk-golang/pkg/runners"
)

// runAgentWithPlugins 从 flow.Manager 中获取并运行命名工作流。
func runAgentWithPlugins(flowName string, saveSession bool, jsonOutput bool) error {
	fmt.Printf("尝试从插件加载工作流: %s\n", flowName)

	// 从全局 Manager 获取 Agent
	mgr := flow.GetGlobalManager()
	agent, ok := mgr.Get(flowName)
	if !ok {
		return fmt.Errorf("工作流 %s 未注册或未加载，检查配置并确保插件已编译", flowName)
	}

	fmt.Printf("已加载工作流: %s\n", flowName)
	fmt.Printf("使用模型: %s\n", agent.Model())
	fmt.Printf("Agent 描述: %s\n", agent.Description())

	// 创建 runner
	runner := runners.NewSimpleRunner()
	runner.SetJSONOutput(jsonOutput)
	ctx := context.Background()

	// 配置保存会话选项
	if saveSession {
		runner.SetSaveSessionEnabled(true)
	}

	// 交互模式运行 Agent
	return runner.RunInteractive(ctx, agent, os.Stdin, os.Stdout)
}
