package novel_v4

import (
	"context"
	"fmt"

	"github.com/nvcnvn/adk-golang/pkg/agents"
)

// BaseAgent 为 novel_v4 插件提供基础智能体功能
// 封装了通用的智能体行为和配置
type BaseAgent struct {
	agents.Agent
	agentType    string
	capabilities []string
	config       map[string]interface{}
}

// NewBaseAgent 创建基础智能体实例
func NewBaseAgent(name, agentType, instruction, description string, capabilities []string) *BaseAgent {
	base := &BaseAgent{
		agentType:    agentType,
		capabilities: capabilities,
		config:       make(map[string]interface{}),
	}

	// 配置底层 Agent
	base.Agent = *agents.NewAgent(
		agents.WithName(name),
		agents.WithModel("deepseek-chat"),
		agents.WithInstruction(instruction),
		agents.WithDescription(description),
	)

	return base
}

// GetType 返回智能体类型
func (b *BaseAgent) GetType() string {
	return b.agentType
}

// GetCapabilities 返回智能体能力列表
func (b *BaseAgent) GetCapabilities() []string {
	return b.capabilities
}

// SetConfig 设置配置参数
func (b *BaseAgent) SetConfig(key string, value interface{}) {
	b.config[key] = value
}

// GetConfig 获取配置参数
func (b *BaseAgent) GetConfig(key string) (interface{}, bool) {
	value, exists := b.config[key]
	return value, exists
}

// ProcessWithContext 处理带上下文的请求，提供额外的元数据支持
func (b *BaseAgent) ProcessWithContext(ctx context.Context, msg string, metadata map[string]interface{}) (string, error) {
	// 构建增强的处理消息
	enhancedMsg := b.buildEnhancedMessage(msg, metadata)
	
	// 调用底层处理逻辑
	return b.Agent.Process(ctx, enhancedMsg)
}

// buildEnhancedMessage 构建增强的消息内容
func (b *BaseAgent) buildEnhancedMessage(msg string, metadata map[string]interface{}) string {
	enhanced := fmt.Sprintf("=== %s 智能体处理 ===\n", b.agentType)
	enhanced += fmt.Sprintf("能力范围: %v\n", b.capabilities)
	
	if metadata != nil && len(metadata) > 0 {
		enhanced += "上下文信息:\n"
		for key, value := range metadata {
			enhanced += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}
	
	enhanced += fmt.Sprintf("\n原始请求:\n%s", msg)
	return enhanced
}

// ValidateCapability 验证智能体是否具备指定能力
func (b *BaseAgent) ValidateCapability(capability string) bool {
	for _, cap := range b.capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// LogActivity 记录智能体活动
func (b *BaseAgent) LogActivity(activity string) {
	fmt.Printf("[%s] %s: %s\n", b.agentType, b.Agent.Name(), activity)
}
