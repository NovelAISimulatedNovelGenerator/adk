package test_rag_tool

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/tools"
	"github.com/nvcnvn/adk-golang/pkg/tools/vector_rag_tool"
)

// createContextAwareRAGTool 创建能从context获取参数的RAG工具包装器
func createContextAwareRAGTool(toolType, suffix string) tools.Tool {
	// 根据工具类型构造不同的参数 schema
	var inputSchema tools.ParameterSchema
	var outputSchema map[string]tools.ParameterSchema

	switch toolType {
	case "write":
		inputSchema = tools.ParameterSchema{
			Type: "object",
			Properties: map[string]tools.ParameterSchema{
				"content": {
					Type:        "string",
					Description: "要写入RAG的文本内容",
					Required:    true,
				},
			},
		}
		outputSchema = map[string]tools.ParameterSchema{
			"result": {
				Type:        "string",
				Description: "写入操作结果描述",
			},
		}
	case "search":
		inputSchema = tools.ParameterSchema{
			Type: "object",
			Properties: map[string]tools.ParameterSchema{
				"query": {
					Type:        "string",
					Description: "搜索查询关键词",
					Required:    true,
				},
				"top_k": {
					Type:        "integer",
					Description: "返回结果条数（可选）",
					Required:    false,
				},
			},
		}
		outputSchema = map[string]tools.ParameterSchema{
			"results": {
				Type:        "array",
				Description: "搜索到的结果数组",
			},
		}
	default:
		// 理论不会走到这里，保留占位避免编译器警告
		inputSchema = tools.ParameterSchema{Type: "object"}
		outputSchema = map[string]tools.ParameterSchema{}
	}

	return tools.NewTool(
		fmt.Sprintf("rag_%s_%s", toolType, suffix),
		fmt.Sprintf("RAG %s 工具（%s），从context动态获取user_id和archive_id", toolType, suffix),
		tools.ToolSchema{
			Input:  inputSchema,
			Output: outputSchema,
		},
		func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
			// 从context获取用户信息
			userID := "default_user"
			archiveID := "default_archive"
			log.Print("操tool used")
			if uid, ok := ctx.Value("user_id").(string); ok && uid != "" {
				userID = uid
			}
			if aid, ok := ctx.Value("archive_id").(string); ok && aid != "" {
				archiveID = aid
			}

			// 根据suffix调整参数（测试隔离性）
			if suffix == "test_variant" {
				userID = userID + "_"
				archiveID = archiveID + "_test"
			}

			// 根据工具类型创建实际的RAG工具并执行
			switch toolType {
			case "write":
				actualTool := vector_rag_tool.NewRAGWriteTool(userID, archiveID)
				return actualTool.Execute(ctx, input)
			case "search":
				actualTool := vector_rag_tool.NewRAGSearchTool(userID, archiveID)
				return actualTool.Execute(ctx, input)
			default:
				return nil, fmt.Errorf("unsupported tool type: %s", toolType)
			}
		},
	)
}

// Build 构造一个使用工具验证 RAG 系统隔离性的工作流。
// 测试不同 tenant_id 的数据隔离，确保不会发生越界访问。
func Build() *agents.Agent {
	// 创建4个context感知的RAG工具
	writeToolMain := createContextAwareRAGTool("write", "main")
	searchToolMain := createContextAwareRAGTool("search", "main")
	writeToolTest := createContextAwareRAGTool("write", "test_variant")
	searchToolTest := createContextAwareRAGTool("search", "test_variant")

	ragAgent := agents.NewAgent(
		agents.WithName("rag_tool_test_agent"),
		agents.WithModel("pool:glm_pool"),
		agents.WithInstruction(`你是一个RAG工具测试助手。你有4个工具可以使用，必须通过调用工具来完成任务。

调用工具的格式：
{"tool_name": "工具名称", "parameters": {"参数名": "参数值"}}

可用工具：
1. rag_write_main: 为当前用户写入内容 - 参数: {"content": "文本内容"}
2. rag_search_main: 为当前用户搜索内容 - 参数: {"query": "搜索关键词"}
3. rag_write_test_variant: 为测试变体用户写入内容 - 参数: {"content": "文本内容"}
4. rag_search_test_variant: 为测试变体用户搜索内容 - 参数: {"query": "搜索关键词"}

测试步骤（必须依次完成）：
1. 使用 rag_write_main 写入当前用户数据："用户主数据库信息，包含密码123456"
2. 使用 rag_write_test_variant 写入测试用户数据："测试用户变体数据，包含密码abcdef"
3. 使用 rag_search_main 搜索"密码"关键词
4. 使用 rag_search_test_variant 搜索"密码"关键词

验证隔离性：主用户搜索应只返回123456，测试用户搜索应只返回abcdef。

请严格按顺序调用这4个工具，每个都用JSON格式。`),
		agents.WithDescription("使用工具测试RAG系统数据隔离性的代理"),
		agents.WithTools(writeToolMain, searchToolMain, writeToolTest, searchToolTest),
		agents.WithBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
			// 从context获取用户信息用于显示
			userID := "default_user"
			archiveID := "default_archive"

			if uid, ok := ctx.Value("user_id").(string); ok && uid != "" {
				log.Printf("uid detected %s", uid)
				userID = uid
			}
			if aid, ok := ctx.Value("archive_id").(string); ok && aid != "" {
				log.Printf("aid detected %s", aid)
				archiveID = aid
			}

			timestamp := time.Now().UnixNano()

			testResults := fmt.Sprintf(`RAG工具隔离性测试开始 (时间戳: %d)

检测到的上下文信息：
- user_id: %s (主用户)
- archive_id: %s (主归档)
- 测试变体用户: %s_ (添加下划线后缀)
- 测试变体归档: %s_test (添加_test后缀)

可用工具：
- rag_write_main: 为主用户写入内容
- rag_search_main: 为主用户搜索内容
- rag_write_test_variant: 为测试变体用户写入内容
- rag_search_test_variant: 为测试变体用户搜索内容

用户输入: %s

请开始使用工具进行测试...`,
				timestamp, userID, archiveID, userID, archiveID, msg)

			return testResults, false // 返回false让Agent继续处理
		}),
	)

	return ragAgent
}
