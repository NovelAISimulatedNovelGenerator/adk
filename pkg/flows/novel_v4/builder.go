package novel_v4

import (
	"context"
	"fmt"
	"log"

	"github.com/nvcnvn/adk-golang/pkg/agents"
)

// Build 构建 novel_v4 版本的分层智能体架构
// 相比 novel_v3，本版本提供了更完整的插件化实现
func Build() *agents.Agent {
	// 创建核心智能体组件
	architect := NewArchitectAgent()
	writer := NewWriterAgent()
	librarian := NewLibrarianAgent()
	coordinator := NewCoordinatorAgent()

	// 构建执行层 - 并行处理核心任务
	executionLayer := agents.NewParallelAgent(agents.ParallelAgentConfig{
		Name:        "novel_v4_execution",
		Description: "Novel v4 执行层 - 并行创作处理",
		SubAgents:   []*agents.Agent{&architect.Agent, &writer.Agent, &librarian.Agent},
		Workers:     3, // 设置并发工作者数量
	})

	// 构建决策层 - 串行规划与协调
	decisionLayer := agents.NewSequentialAgent(agents.SequentialAgentConfig{
		Name:        "novel_v4_decision",
		Description: "Novel v4 决策层 - 串行规划协调",
		SubAgents:   []*agents.Agent{&coordinator.Agent},
	})

	// 构建根智能体 - 统一调度
	root := agents.NewSequentialAgent(agents.SequentialAgentConfig{
		Name:        "novel_v4_root",
		Description: "Novel v4 根智能体 - 统一创作调度系统",
		SubAgents:   []*agents.Agent{&decisionLayer.Agent, &executionLayer.Agent},
	})

	// 设置处理逻辑
	setupProcessingLogic(root, decisionLayer, executionLayer, coordinator.BaseAgent, architect.BaseAgent, writer.BaseAgent, librarian.BaseAgent)

	log.Println("[Novel v4] 插件构建完成")
	return &root.Agent
}

// setupProcessingLogic 设置各层级的处理逻辑
func setupProcessingLogic(root *agents.SequentialAgent, decision *agents.SequentialAgent, execution *agents.ParallelAgent,
	coordinator, architect, writer, librarian *BaseAgent) {

	// 决策层处理逻辑
	decision.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
		log.Printf("[Novel v4 决策层] 开始协调处理: %s", truncateString(msg, 50))
		
		// 通过协调器进行决策规划
		result, err := coordinator.Process(ctx, msg)
		if err != nil {
			return fmt.Sprintf("[协调器错误]: %v", err), true
		}
		
		log.Println("[Novel v4 决策层] 决策协调完成")
		return result, true
	})

	// 执行层处理逻辑
	execution.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
		log.Printf("[Novel v4 执行层] 开始并行创作: %s", truncateString(msg, 50))
		
		// 并行执行各专业智能体
		results := make(map[string]string)
		
		// 架构师处理
		if archResult, err := architect.Process(ctx, msg); err == nil {
			results["architect"] = archResult
		} else {
			log.Printf("[架构师错误]: %v", err)
		}
		
		// 写作者处理
		if writerResult, err := writer.Process(ctx, msg); err == nil {
			results["writer"] = writerResult
		} else {
			log.Printf("[写作者错误]: %v", err)
		}
		
		// 图书管理员处理
		if libResult, err := librarian.Process(ctx, msg); err == nil {
			results["librarian"] = libResult
		} else {
			log.Printf("[图书管理员错误]: %v", err)
		}
		
		// 合并结果
		finalResult := mergeResults(results)
		log.Println("[Novel v4 执行层] 并行创作完成")
		return finalResult, true
	})

	// 根智能体处理逻辑
	root.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
		log.Printf("[Novel v4 根智能体] 统一调度开始: %s", truncateString(msg, 50))
		
		// 先进行决策规划
		decisionResult, err := decision.Agent.Process(ctx, msg)
		if err != nil {
			return fmt.Sprintf("[决策层错误]: %v", err), true
		}
		
		// 再进行执行创作
		executionResult, err := execution.Agent.Process(ctx, decisionResult)
		if err != nil {
			return fmt.Sprintf("[执行层错误]: %v", err), true
		}
		
		log.Println("[Novel v4 根智能体] 统一调度完成")
		return executionResult, true
	})
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// mergeResults 合并多个智能体的结果
func mergeResults(results map[string]string) string {
	merged := "=== Novel v4 创作结果 ===\n"
	
	if arch, ok := results["architect"]; ok {
		merged += fmt.Sprintf("【架构设计】\n%s\n\n", arch)
	}
	
	if writer, ok := results["writer"]; ok {
		merged += fmt.Sprintf("【内容创作】\n%s\n\n", writer)
	}
	
	if lib, ok := results["librarian"]; ok {
		merged += fmt.Sprintf("【知识管理】\n%s\n\n", lib)
	}
	
	return merged
}
