package test

// 集成测试框架，用于测试 Novel 工作流与 Quad Memory 服务的整合
// 参考 pkg/flows/novel/framework.go 设计，专门用于验证完整系统协作

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/memory"
)

// TestConfig 测试配置
type TestConfig struct {
	GraphDBBaseURL string `json:"graphdb_base_url"`
	RepositoryID   string `json:"repository_id"`
	TestUserID     string `json:"test_user_id"`
	TestArchiveID  string `json:"test_archive_id"`
}

// TestResult 测试结果
type TestResult struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	MemoryQuads  int                    `json:"memory_quads,omitempty"`
	AgentResults []AgentTestResult      `json:"agent_results,omitempty"`
}

// AgentTestResult Agent测试结果
type AgentTestResult struct {
	AgentName string `json:"agent_name"`
	Success   bool   `json:"success"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

// BuildTestFramework 构建集成测试框架，结合 Novel 工作流和 Quad Memory 服务
func BuildTestFramework(config TestConfig) *agents.Agent {
	// 创建 Quad Memory 服务
	memoryConfig := memory.QuadMemoryConfig{
		BaseURL:      config.GraphDBBaseURL,
		RepositoryID: config.RepositoryID,
		MaxRetries:   3,
	}

	memoryService := memory.NewQuadMemoryService(memoryConfig)

	// 创建增强版的 Novel Agents，集成 Memory 操作
	worldviewAgent := createMemoryEnabledAgent(
		"test_worldview_agent",
		"世界观架构师（测试版）",
		"你是具有记忆能力的小说世界观架构师。输出世界观设定后，会自动保存到记忆系统。",
		memoryService,
		"worldview",
	)

	characterAgent := createMemoryEnabledAgent(
		"test_character_agent",
		"角色设计师（测试版）",
		"你是具有记忆能力的角色塑造专家。创建角色设定后，会自动保存到记忆系统。",
		memoryService,
		"character",
	)

	plotAgent := createMemoryEnabledAgent(
		"test_plot_agent",
		"剧情编剧（测试版）",
		"你是具有记忆能力的剧情编剧。设计剧情后，会自动保存到记忆系统。",
		memoryService,
		"plot",
	)

	// 创建记忆检索Agent
	memoryAgent := createMemoryRetrievalAgent("test_memory_agent", "记忆检索专家", memoryService)

	// 创建并行执行层 - 包含记忆功能的创作智能体
	executionLayer := agents.NewParallelAgent(agents.ParallelAgentConfig{
		Name:        "test_execution_layer",
		Description: "测试版执行层（并行）",
		SubAgents:   []*agents.Agent{worldviewAgent, characterAgent, plotAgent, memoryAgent},
		Workers:     3,
	})

	// 创建测试协调Agent
	coordinatorAgent := createTestCoordinatorAgent("test_coordinator", "测试协调器", config)

	// 创建顶层串行Agent
	testFramework := agents.NewSequentialAgent(agents.SequentialAgentConfig{
		Name:        "test_framework",
		Description: "Novel工作流与Quad Memory服务集成测试框架",
		SubAgents:   []*agents.Agent{coordinatorAgent, &executionLayer.Agent},
	})

	// 设置框架级别的回调处理
	testFramework.Agent.SetBeforeAgentCallback(func(ctx context.Context, input string) (string, bool) {
		log.Printf("[测试框架] 开始集成测试，输入: %s", truncateString(input, 50))

		// 验证context中的用户信息传递
		if userID, ok := ctx.Value("user_id").(string); ok {
			log.Printf("[测试框架] 检测到user_id: %s", userID)
		} else {
			log.Printf("[测试框架] 警告：未检测到user_id")
		}

		if archiveID, ok := ctx.Value("archive_id").(string); ok {
			log.Printf("[测试框架] 检测到archive_id: %s", archiveID)
		} else {
			log.Printf("[测试框架] 警告：未检测到archive_id")
		}

		return input, false // 继续处理
	})

	testFramework.Agent.SetAfterAgentCallback(func(ctx context.Context, response string) string {
		log.Printf("[测试框架] 集成测试完成，输出长度: %d 字符", len(response))

		// 生成测试报告
		report := generateTestReport(ctx, response, config)
		return report
	})

	return &testFramework.Agent
}

// createMemoryEnabledAgent 创建具有记忆能力的Agent
func createMemoryEnabledAgent(name, description, instruction string, memoryService *memory.QuadMemoryService, agentType string) *agents.Agent {
	agent := agents.NewAgent(
		agents.WithName(name),
		agents.WithModel("deepseek-chat"),
		agents.WithInstruction(instruction),
		agents.WithDescription(description),
	)

	// 设置具有记忆功能的回调
	agent.SetAfterAgentCallback(func(ctx context.Context, output string) string {
		// 从context获取用户信息
		userID := "default_user"
		archiveID := "default_archive"

		if uid, ok := ctx.Value("user_id").(string); ok && uid != "" {
			userID = uid
		}
		if aid, ok := ctx.Value("archive_id").(string); ok && aid != "" {
			archiveID = aid
		}

		// 创建层次化上下文
		hierarchicalCtx := &memory.HierarchicalContext{
			TenantID: userID,
			StoryID:  archiveID,
		}

		// 将Agent输出保存到记忆系统
		quad := memory.Quad{
			Subject:   fmt.Sprintf("agent:%s", name),
			Predicate: "produces",
			Object:    fmt.Sprintf("content:%s", truncateString(output, 200)),
			Context:   fmt.Sprintf("%s_%s", hierarchicalCtx.TenantID, hierarchicalCtx.StoryID),
		}

		_, err := memoryService.AddQuad(ctx, hierarchicalCtx, quad)
		if err != nil {
			log.Printf("[%s] 警告：保存到记忆系统失败: %v", name, err)
		} else {
			log.Printf("[%s] 成功保存到记忆系统", name)
		}

		return output
	})

	return agent
}

// createMemoryRetrievalAgent 创建记忆检索Agent
func createMemoryRetrievalAgent(name, description string, memoryService *memory.QuadMemoryService) *agents.Agent {
	agent := agents.NewAgent(
		agents.WithName(name),
		agents.WithModel("deepseek-chat"),
		agents.WithInstruction("你是记忆检索专家，负责从记忆系统中检索相关信息并整理输出。"),
		agents.WithDescription(description),
	)

	agent.SetBeforeAgentCallback(func(ctx context.Context, input string) (string, bool) {
		// 从记忆系统检索相关信息
		userID := "default_user"
		archiveID := "default_archive"

		if uid, ok := ctx.Value("user_id").(string); ok && uid != "" {
			userID = uid
		}
		if aid, ok := ctx.Value("archive_id").(string); ok && aid != "" {
			archiveID = aid
		}

		// 构造搜索查询
		hierarchicalCtx := &memory.HierarchicalContext{
			TenantID: userID,
			StoryID:  archiveID,
		}

		query := memory.QuadSearchQuery{
			Context: hierarchicalCtx,
			Scope:   "story", // 搜索整个故事范围
		}

		quads, err := memoryService.SearchQuads(ctx, query)
		if err != nil {
			log.Printf("[%s] 记忆检索失败: %v", name, err)
			return fmt.Sprintf("记忆检索失败: %v。继续处理原始输入: %s", err, input), false
		}

		// 整理检索到的记忆
		memoryContent := fmt.Sprintf("=== 检索到 %d 条记忆 ===\n", len(quads))
		for i, quad := range quads {
			memoryContent += fmt.Sprintf("%d. %s -> %s -> %s\n", i+1, quad.Subject, quad.Predicate, quad.Object)
		}

		enhancedInput := fmt.Sprintf("%s\n\n=== 记忆上下文 ===\n%s\n=== 原始输入 ===\n%s",
			memoryContent, memoryContent, input)

		log.Printf("[%s] 成功检索到 %d 条记忆", name, len(quads))
		return enhancedInput, false
	})

	return agent
}

// createTestCoordinatorAgent 创建测试协调Agent
func createTestCoordinatorAgent(name, description string, config TestConfig) *agents.Agent {
	agent := agents.NewAgent(
		agents.WithName(name),
		agents.WithModel("deepseek-chat"),
		agents.WithInstruction("你是集成测试协调器，负责分析测试需求并制定测试计划。"),
		agents.WithDescription(description),
	)

	agent.SetBeforeAgentCallback(func(ctx context.Context, input string) (string, bool) {
		testPlan := fmt.Sprintf(`
=== Novel工作流与Quad Memory服务集成测试计划 ===

测试配置:
- GraphDB URL: %s
- Repository: %s  
- 测试用户: %s
- 测试归档: %s

测试目标:
1. 验证Novel工作流Agent的协作能力
2. 验证Quad Memory服务的存储和检索功能
3. 验证Context中user_id和archive_id的传递
4. 验证完整的记忆-创作循环

原始测试输入: %s

请各Agent按计划执行测试...
`, config.GraphDBBaseURL, config.RepositoryID, config.TestUserID, config.TestArchiveID, input)

		log.Printf("[%s] 生成测试计划完成", name)
		return testPlan, true
	})

	return agent
}

// generateTestReport 生成测试报告
func generateTestReport(ctx context.Context, output string, config TestConfig) string {
	report := TestResult{
		Success: true,
		Message: "集成测试执行完成",
		Details: map[string]interface{}{
			"config":        config,
			"output_length": len(output),
		},
	}

	// 检查context传递
	if userID, ok := ctx.Value("user_id").(string); ok {
		report.Details["user_id_detected"] = userID
	}
	if archiveID, ok := ctx.Value("archive_id").(string); ok {
		report.Details["archive_id_detected"] = archiveID
	}

	// 转为JSON格式
	reportJSON, _ := json.MarshalIndent(report, "", "  ")

	return fmt.Sprintf(`
=== Novel工作流与Quad Memory服务集成测试报告 ===

%s

=== 详细输出 ===
%s
`, string(reportJSON), output)
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
