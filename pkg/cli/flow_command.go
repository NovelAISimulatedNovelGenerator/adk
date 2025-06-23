package cli

// flow_command.go 提供 flow 子命令，用于运行工作流插件

import (
	"github.com/spf13/cobra"
)

var (
	flowCmd = &cobra.Command{
		Use:   "flow [workflow_name]",
		Short: "运行已注册的工作流插件",
		Long:  "从已加载的工作流插件中运行指定名称的工作流，使用 flow.Manager 直接获取。",
		Args:  cobra.ExactArgs(1), // 要求提供工作流名称
		RunE: func(cmd *cobra.Command, args []string) error {
			// 获取工作流名称（第一个也是唯一一个参数）
			workflowName := args[0]

			// 获取选项
			saveSession, _ := cmd.Flags().GetBool("save_session")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// 调用 runAgentWithPlugins 函数运行插件
			return runAgentWithPlugins(workflowName, saveSession, jsonOutput)
		},
	}
)

func init() {
	// 添加到根命令
	rootCmd.AddCommand(flowCmd)

	// 添加与 run 相同的标志
	flowCmd.Flags().BoolP("save_session", "", false, "保存会话到 JSON 文件")
	flowCmd.Flags().BoolP("json", "j", false, "以 JSON 格式输出交互内容")
}
