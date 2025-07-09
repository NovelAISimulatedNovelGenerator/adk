package novel_v4

import (
	"context"
	"encoding/json"
	"fmt"
)

// ArchitectAgent 架构师智能体，负责小说的整体架构设计
type ArchitectAgent struct {
	*BaseAgent
}

// NewArchitectAgent 创建新的架构师智能体
func NewArchitectAgent() *ArchitectAgent {
	capabilities := []string{
		"故事架构设计",
		"世界观构建", 
		"人物关系设计",
		"情节框架规划",
		"主题分析",
	}

	instruction := `你是小说架构师智能体，专门负责小说的整体架构设计和规划。你的核心职责包括：

1. **世界观构建**：创建完整的故事世界观，包括时空背景、规则体系、社会结构等
2. **人物架构**：设计主要人物及其关系网络，确保人物设定的逻辑一致性
3. **情节框架**：规划故事的主线和支线结构，设计关键转折点和冲突
4. **主题设计**：明确故事要传达的核心主题和价值观念
5. **结构分析**：设计章节结构和叙事节奏

输出要求：
- 使用结构化的 JSON 格式输出架构设计
- 确保各元素间的逻辑关联性和一致性
- 提供清晰的设计理由和实现建议
- 格式：{"architect_result": {"worldview": "...", "characters": "...", "plot_structure": "...", "themes": "...", "chapters": "..."}}`

	base := NewBaseAgent(
		"architect_v4", 
		"架构师",
		instruction,
		"Novel v4 架构师智能体 - 负责小说整体架构设计",
		capabilities,
	)

	return &ArchitectAgent{BaseAgent: base}
}

// DesignArchitecture 设计小说架构
func (a *ArchitectAgent) DesignArchitecture(ctx context.Context, premise string) (*ArchitectureDesign, error) {
	a.LogActivity("开始设计小说架构")
	
	metadata := map[string]interface{}{
		"task_type": "architecture_design",
		"premise":   premise,
	}

	result, err := a.ProcessWithContext(ctx, premise, metadata)
	if err != nil {
		return nil, fmt.Errorf("架构设计失败: %w", err)
	}

	// 解析结果
	design, err := a.parseArchitectureResult(result)
	if err != nil {
		a.LogActivity(fmt.Sprintf("架构解析失败: %v", err))
		// 如果解析失败，返回基础设计
		return a.createFallbackDesign(premise), nil
	}

	a.LogActivity("架构设计完成")
	return design, nil
}

// parseArchitectureResult 解析架构设计结果
func (a *ArchitectAgent) parseArchitectureResult(result string) (*ArchitectureDesign, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, err
	}

	archResult, ok := parsed["architect_result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的架构结果格式")
	}

	design := &ArchitectureDesign{
		Worldview:     getStringValue(archResult, "worldview"),
		Characters:    getStringValue(archResult, "characters"),
		PlotStructure: getStringValue(archResult, "plot_structure"),
		Themes:        getStringValue(archResult, "themes"),
		Chapters:      getStringValue(archResult, "chapters"),
	}

	return design, nil
}

// createFallbackDesign 创建备用架构设计
func (a *ArchitectAgent) createFallbackDesign(premise string) *ArchitectureDesign {
	return &ArchitectureDesign{
		Worldview:     "基于前提构建的基础世界观",
		Characters:    "待完善的人物设定",
		PlotStructure: "基础三幕式结构",
		Themes:        "从前提中提取的核心主题",
		Chapters:      "标准章节划分",
		Premise:       premise,
	}
}

// ValidateDesign 验证架构设计的完整性
func (a *ArchitectAgent) ValidateDesign(design *ArchitectureDesign) []string {
	var issues []string

	if design.Worldview == "" {
		issues = append(issues, "缺少世界观设定")
	}
	if design.Characters == "" {
		issues = append(issues, "缺少人物设计")
	}
	if design.PlotStructure == "" {
		issues = append(issues, "缺少情节结构")
	}
	if design.Themes == "" {
		issues = append(issues, "缺少主题设定")
	}

	return issues
}

// ArchitectureDesign 架构设计结果结构
type ArchitectureDesign struct {
	Worldview     string `json:"worldview"`     // 世界观设定
	Characters    string `json:"characters"`    // 人物设计
	PlotStructure string `json:"plot_structure"` // 情节结构
	Themes        string `json:"themes"`        // 主题设定
	Chapters      string `json:"chapters"`      // 章节结构
	Premise       string `json:"premise"`       // 原始前提
}

// ToJSON 将架构设计转换为JSON格式
func (d *ArchitectureDesign) ToJSON() (string, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// getStringValue 从map中安全获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}
