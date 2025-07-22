package main

import (
	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
	test "github.com/nvcnvn/adk-golang/pkg/flows/test"
)

// pluginImpl 实现 flow.FlowPlugin，用于Novel工作流与Quad Memory服务集成测试
type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "novel_memory_integration_test" }

func (p *pluginImpl) Build() (*agents.Agent, error) {
	// 创建默认测试配置
	config := test.TestConfig{
		GraphDBBaseURL: "http://localhost:7200", // 默认GraphDB地址
		RepositoryID:   "novelai_memory_test",   // 测试用仓库ID
		TestUserID:     "test_user_123",         // 测试用户ID
		TestArchiveID:  "test_archive_456",      // 测试归档ID
	}
	
	return test.BuildTestFramework(config), nil
}

// Plugin 导出符号 (指针)，供 plugin.Open 查找。
var Plugin flow.FlowPlugin = &pluginImpl{}
