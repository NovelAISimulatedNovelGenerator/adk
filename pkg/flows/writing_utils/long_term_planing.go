package writingutils

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/memory"
)

const (
	agentName        string = "long_term_planning_agent"
	agentDescription string = "为小说提供长期规划的agent"
	agentInstruction string = `
	你是一个专门负责为小说背景提供长期规划的专家。
	你的任务：生成一份包含世界观、角色、主线、关键情节点的长期写作规划。
	完成标准：当规划覆盖三幕结构、主要角色弧线、关键冲突与高潮时，设置 "finish": true。
	工作流：每轮只对尚未完善的部分做增量补全，不要重复已完成内容。
	思考→计划→输出三步走，但 只输出 JSON。
	使用简体中文；各字段内容尽量简洁；每个场景描述≤80字。
	输出标准格式:
	{
	"finish": false,          // true=规划已完成，false=继续调用
	"plan": [                 // 长期规划的分段 / 里程碑
		{ "id": 1, "title": "第一幕：世界观奠基", "status": "done" },
		{ "id": 2, "title": "第二幕：冲突升级",    "status": "in_progress" },
		{ "id": 3, "title": "第三幕：高潮与结局",  "status": "todo" }
	],
	"next_request": "请继续完善第二幕的冲突细节",  // 交回给上层的提示
	"reason": "需要更多角色动机细节才能进入第三幕" // 为什么未 finish
	}
	如需调用工具，仅返回 function-call JSON；  
	如需正常回复/输出规划，严格使用预定义的规划 JSON；  
	绝不同时混合两种格式。
	`
)

// 配置长期规划小说的结构体
type LongTermPlanningConfig struct {
	// MemoryService RAG内存服务实例
	MemoryService memory.MemoryService
	// LLMModel 用于内容分析的LLM模型名称
	LLMModel string
	// UserID 用户ID，对应RAG的tenant_id
	UserID string
	// ArchiveID 归档ID，对应RAG的session_id
	ArchiveID string
}

type LongTermPlanningService struct {
    config         LongTermPlanningConfig
    planningAgents *agents.Agent
}

const defaultMaxIterations = 10

// NewLongTermPlanningService 创建新的服务实例
func NewLongTermPlanningService(config LongTermPlanningConfig) *LongTermPlanningService {

	return &LongTermPlanningService{
		config:         config,
		planningAgents: createPlanningAgent(config.LLMModel),
	}
}

func createPlanningAgent(model string) *agents.Agent {
    // TODO: 在此处注入工具，例如检索或保存规划的工具
    return agents.NewAgent(
        agents.WithName(agentName),
        agents.WithModel(model),
        agents.WithInstruction(agentInstruction),
        agents.WithDescription(agentDescription),
        // agents.WithTools(tool1, tool2), // 预留
    )
}

// GeneratePlan 根据给定前提/提示生成长期规划。
// 会循环调用 planningAgents，直到 "finish" 字段为 true，或达到最大迭代次数。
func (s *LongTermPlanningService) GeneratePlan(ctx context.Context, premise string) (string, error) {
    if s.planningAgents == nil {
        return "", fmt.Errorf("planning agent not initialized")
    }

    currentRequest := premise
    for i := 0; i < defaultMaxIterations; i++ {
        // 1. 调用规划 Agent
        resp, err := s.planningAgents.Process(ctx, currentRequest)
        if err != nil {
            return "", err
        }

        // 2. 解析 JSON 响应
        var payload map[string]interface{}
        if err := json.Unmarshal([]byte(resp), &payload); err != nil {
            log.Printf("[LongTermPlanning] 无效的 JSON 响应: %v, raw=%s", err, resp)
            return "", fmt.Errorf("invalid JSON response: %w", err)
        }

        // 3. 检查是否完成
        if finish, ok := payload["finish"].(bool); ok && finish {
            // TODO: 将最终规划保存到 MemoryService 或其他持久化工具
            return resp, nil
        }

        // 4. 获取下一步请求
        nextReq, ok := payload["next_request"].(string)
        if !ok || nextReq == "" {
            return resp, nil // 找不到 next_request，则视为终止
        }

        // TODO: 根据需要调用工具（检索 / 保存等），此处预留

        currentRequest = nextReq
    }

    return "", fmt.Errorf("max iterations(%d) reached without finish", defaultMaxIterations)
}
