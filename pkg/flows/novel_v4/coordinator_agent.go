package novel_v4

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CoordinatorAgent 协调者智能体，负责整体协调和任务调度
type CoordinatorAgent struct {
	*BaseAgent
}

// NewCoordinatorAgent 创建新的协调者智能体
func NewCoordinatorAgent() *CoordinatorAgent {
	capabilities := []string{
		"任务协调",
		"流程管理",
		"质量控制",
		"资源调度",
		"进度跟踪",
		"决策支持",
	}

	instruction := `你是小说创作协调者智能体，负责整体协调和任务调度。你的核心职责包括：

1. **任务协调**：协调各个智能体之间的工作，确保协同效率
2. **流程管理**：管理创作流程，优化工作步骤和顺序
3. **质量控制**：监控创作质量，提出改进建议
4. **资源调度**：合理分配创作资源和任务优先级
5. **进度跟踪**：跟踪创作进度，识别潜在问题
6. **决策支持**：为创作决策提供数据支持和建议

协调要求：
- 分析当前创作需求和目标
- 制定合理的工作计划和时间安排
- 识别各智能体的协作点和依赖关系
- 监控执行过程并及时调整策略
- 确保最终输出的整体质量和一致性

输出格式：
{"coordinator_result": {"task_plan": "任务规划", "resource_allocation": "资源分配", "quality_metrics": "质量指标", "coordination_strategy": "协调策略", "progress_tracking": "进度跟踪"}}`

	base := NewBaseAgent(
		"coordinator_v4",
		"协调者",
		instruction,
		"Novel v4 协调者智能体 - 负责整体协调和任务调度",
		capabilities,
	)

	return &CoordinatorAgent{BaseAgent: base}
}

// CoordinateCreation 协调创作过程
func (c *CoordinatorAgent) CoordinateCreation(ctx context.Context, request *CreationRequest) (*CoordinationPlan, error) {
	c.LogActivity("开始协调创作过程")

	// 构建协调指令
	instruction := c.buildCoordinationInstruction(request)
	
	metadata := map[string]interface{}{
		"task_type": "creation_coordination",
		"request":   request,
	}

	result, err := c.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("创作协调失败: %w", err)
	}

	// 解析协调结果
	plan, err := c.parseCoordinationResult(result)
	if err != nil {
		c.LogActivity(fmt.Sprintf("协调结果解析失败: %v", err))
		return c.createFallbackPlan(request), nil
	}

	c.LogActivity("创作协调完成")
	return plan, nil
}

// buildCoordinationInstruction 构建协调指令
func (c *CoordinatorAgent) buildCoordinationInstruction(request *CreationRequest) string {
	var instruction strings.Builder
	
	instruction.WriteString("请协调以下创作请求：\n\n")
	instruction.WriteString(fmt.Sprintf("**创作主题：** %s\n", request.Theme))
	instruction.WriteString(fmt.Sprintf("**创作类型：** %s\n", request.Type))
	instruction.WriteString(fmt.Sprintf("**目标长度：** %d字\n", request.TargetLength))
	instruction.WriteString(fmt.Sprintf("**特殊要求：** %s\n", strings.Join(request.Requirements, "、")))
	
	instruction.WriteString("\n**协调任务：**\n")
	instruction.WriteString("1. 分析创作需求，制定工作计划\n")
	instruction.WriteString("2. 确定各智能体的任务分工\n")
	instruction.WriteString("3. 设计协作流程和时间安排\n")
	instruction.WriteString("4. 建立质量控制标准\n")
	instruction.WriteString("5. 制定进度跟踪机制\n")

	return instruction.String()
}

// MonitorProgress 监控创作进度
func (c *CoordinatorAgent) MonitorProgress(ctx context.Context, currentState *CreationState) (*ProgressReport, error) {
	c.LogActivity("开始监控创作进度")

	instruction := c.buildProgressInstruction(currentState)
	
	metadata := map[string]interface{}{
		"task_type":      "progress_monitoring",
		"current_state":  currentState,
	}

	result, err := c.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("进度监控失败: %w", err)
	}

	// 解析进度报告
	report, err := c.parseProgressResult(result)
	if err != nil {
		c.LogActivity(fmt.Sprintf("进度结果解析失败: %v", err))
		return c.createFallbackProgress(currentState), nil
	}

	c.LogActivity("进度监控完成")
	return report, nil
}

// buildProgressInstruction 构建进度监控指令
func (c *CoordinatorAgent) buildProgressInstruction(state *CreationState) string {
	var instruction strings.Builder
	
	instruction.WriteString("请监控当前创作进度：\n\n")
	instruction.WriteString(fmt.Sprintf("**总体进度：** %.1f%%\n", state.OverallProgress))
	instruction.WriteString(fmt.Sprintf("**当前阶段：** %s\n", state.CurrentPhase))
	instruction.WriteString(fmt.Sprintf("**已完成任务：** %d/%d\n", state.CompletedTasks, state.TotalTasks))
	
	instruction.WriteString("\n**各智能体状态：**\n")
	for agent, status := range state.AgentStatus {
		instruction.WriteString(fmt.Sprintf("- %s: %s\n", agent, status))
	}
	
	instruction.WriteString("\n**监控要求：**\n")
	instruction.WriteString("1. 评估当前进度是否符合预期\n")
	instruction.WriteString("2. 识别可能的延迟或问题\n")
	instruction.WriteString("3. 提出优化建议和调整方案\n")
	instruction.WriteString("4. 更新资源分配和任务优先级\n")

	return instruction.String()
}

// QualityControl 进行质量控制
func (c *CoordinatorAgent) QualityControl(ctx context.Context, content string, standards *QualityStandards) (*QualityReport, error) {
	c.LogActivity("开始质量控制检查")

	instruction := c.buildQualityInstruction(content, standards)
	
	metadata := map[string]interface{}{
		"task_type": "quality_control",
		"content":   content,
		"standards": standards,
	}

	result, err := c.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("质量控制失败: %w", err)
	}

	// 解析质量报告
	report, err := c.parseQualityResult(result)
	if err != nil {
		c.LogActivity(fmt.Sprintf("质量结果解析失败: %v", err))
		return c.createFallbackQuality(content, standards), nil
	}

	c.LogActivity("质量控制完成")
	return report, nil
}

// buildQualityInstruction 构建质量控制指令
func (c *CoordinatorAgent) buildQualityInstruction(content string, standards *QualityStandards) string {
	var instruction strings.Builder
	
	instruction.WriteString("请对以下内容进行质量控制：\n\n")
	instruction.WriteString("**待检查内容：**\n")
	instruction.WriteString(content)
	
	instruction.WriteString("\n\n**质量标准：**\n")
	instruction.WriteString(fmt.Sprintf("- 内容质量分数要求：>= %.1f\n", standards.MinContentScore))
	instruction.WriteString(fmt.Sprintf("- 语言流畅度要求：>= %.1f\n", standards.MinLanguageScore))
	instruction.WriteString(fmt.Sprintf("- 逻辑一致性要求：>= %.1f\n", standards.MinConsistencyScore))
	
	instruction.WriteString("\n**检查要点：**\n")
	instruction.WriteString("1. 内容质量和创意水平\n")
	instruction.WriteString("2. 语言表达和文风统一\n")
	instruction.WriteString("3. 逻辑一致性和情节合理性\n")
	instruction.WriteString("4. 人物形象和对话自然度\n")
	instruction.WriteString("5. 提出具体改进建议\n")

	return instruction.String()
}

// 解析方法实现
func (c *CoordinatorAgent) parseCoordinationResult(result string) (*CoordinationPlan, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, err
	}

	coordResult, ok := parsed["coordinator_result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的协调结果格式")
	}

	return &CoordinationPlan{
		TaskPlan:             getStringValue(coordResult, "task_plan"),
		ResourceAllocation:   getStringValue(coordResult, "resource_allocation"),
		QualityMetrics:       getStringValue(coordResult, "quality_metrics"),
		CoordinationStrategy: getStringValue(coordResult, "coordination_strategy"),
		ProgressTracking:     getStringValue(coordResult, "progress_tracking"),
	}, nil
}

func (c *CoordinatorAgent) parseProgressResult(result string) (*ProgressReport, error) {
	// 简化实现，实际应该解析JSON结果
	return &ProgressReport{
		OverallStatus:    "进行中",
		CompletionRate:   0.5,
		EstimatedTime:    "2小时",
		Recommendations:  []string{"继续按计划执行"},
		Adjustments:      []string{},
	}, nil
}

func (c *CoordinatorAgent) parseQualityResult(result string) (*QualityReport, error) {
	// 简化实现，实际应该解析JSON结果
	return &QualityReport{
		OverallScore:     8.5,
		ContentScore:     8.0,
		LanguageScore:    9.0,
		ConsistencyScore: 8.5,
		Issues:          []string{},
		Suggestions:     []string{"整体质量良好"},
	}, nil
}

// 备用方法实现
func (c *CoordinatorAgent) createFallbackPlan(request *CreationRequest) *CoordinationPlan {
	return &CoordinationPlan{
		TaskPlan:             "标准创作流程：架构设计 -> 内容创作 -> 知识管理",
		ResourceAllocation:   "平均分配各智能体资源",
		QualityMetrics:       "采用标准质量评估指标",
		CoordinationStrategy: "顺序执行，并行优化",
		ProgressTracking:     "实时监控各阶段进度",
	}
}

func (c *CoordinatorAgent) createFallbackProgress(state *CreationState) *ProgressReport {
	return &ProgressReport{
		OverallStatus:   "正常进行",
		CompletionRate:  state.OverallProgress / 100.0,
		EstimatedTime:   "预估剩余时间",
		Recommendations: []string{"继续当前工作流程"},
		Adjustments:     []string{},
	}
}

func (c *CoordinatorAgent) createFallbackQuality(content string, standards *QualityStandards) *QualityReport {
	return &QualityReport{
		OverallScore:     standards.MinContentScore,
		ContentScore:     standards.MinContentScore,
		LanguageScore:    standards.MinLanguageScore,
		ConsistencyScore: standards.MinConsistencyScore,
		Issues:          []string{},
		Suggestions:     []string{"需要更详细的质量评估"},
	}
}

// 数据结构定义
type CreationRequest struct {
	Theme        string   `json:"theme"`
	Type         string   `json:"type"`
	TargetLength int      `json:"target_length"`
	Requirements []string `json:"requirements"`
}

type CoordinationPlan struct {
	TaskPlan             string `json:"task_plan"`
	ResourceAllocation   string `json:"resource_allocation"`
	QualityMetrics       string `json:"quality_metrics"`
	CoordinationStrategy string `json:"coordination_strategy"`
	ProgressTracking     string `json:"progress_tracking"`
}

type CreationState struct {
	OverallProgress  float64           `json:"overall_progress"`
	CurrentPhase     string            `json:"current_phase"`
	CompletedTasks   int               `json:"completed_tasks"`
	TotalTasks       int               `json:"total_tasks"`
	AgentStatus      map[string]string `json:"agent_status"`
}

type ProgressReport struct {
	OverallStatus   string   `json:"overall_status"`
	CompletionRate  float64  `json:"completion_rate"`
	EstimatedTime   string   `json:"estimated_time"`
	Recommendations []string `json:"recommendations"`
	Adjustments     []string `json:"adjustments"`
}

type QualityStandards struct {
	MinContentScore     float64 `json:"min_content_score"`
	MinLanguageScore    float64 `json:"min_language_score"`
	MinConsistencyScore float64 `json:"min_consistency_score"`
}

type QualityReport struct {
	OverallScore     float64  `json:"overall_score"`
	ContentScore     float64  `json:"content_score"`
	LanguageScore    float64  `json:"language_score"`
	ConsistencyScore float64  `json:"consistency_score"`
	Issues          []string `json:"issues"`
	Suggestions     []string `json:"suggestions"`
}
