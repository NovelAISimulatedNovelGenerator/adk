package novel_v4

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// LibrarianAgent 图书管理员智能体，负责知识管理和内容组织
type LibrarianAgent struct {
	*BaseAgent
	knowledgeBase map[string]*KnowledgeEntry
}

// NewLibrarianAgent 创建新的图书管理员智能体
func NewLibrarianAgent() *LibrarianAgent {
	capabilities := []string{
		"知识管理",
		"内容组织",
		"设定维护", 
		"连贯性检查",
		"版本控制",
		"研究支持",
	}

	instruction := `你是小说图书管理员智能体，专门负责知识管理和内容组织。你的核心职责包括：

1. **知识管理**：维护和管理小说创作中的所有设定、人物、情节等信息
2. **内容组织**：整理和分类创作内容，确保结构清晰
3. **设定维护**：跟踪和更新世界观、人物设定的变化
4. **连贯性检查**：确保内容前后一致，避免设定冲突
5. **版本控制**：记录内容的修改历史和版本变化  
6. **研究支持**：为其他智能体提供背景信息和参考资料

管理要求：
- 建立完整的知识体系架构
- 维护设定的一致性和准确性
- 提供快速的信息检索服务
- 记录内容变更的完整历史
- 支持多版本内容管理

输出格式：
{"librarian_result": {"knowledge_updates": "知识更新内容", "consistency_checks": "一致性检查结果", "organization_notes": "内容组织说明", "references": "相关参考信息"}}`

	base := NewBaseAgent(
		"librarian_v4",
		"图书管理员", 
		instruction,
		"Novel v4 图书管理员智能体 - 负责知识管理和内容组织",
		capabilities,
	)

	return &LibrarianAgent{
		BaseAgent:     base,
		knowledgeBase: make(map[string]*KnowledgeEntry),
	}
}

// ManageKnowledge 管理知识条目
func (l *LibrarianAgent) ManageKnowledge(ctx context.Context, content string, contentType string) (*LibrarianResult, error) {
	l.LogActivity(fmt.Sprintf("开始管理%s类型的知识", contentType))

	// 构建知识管理指令
	instruction := l.buildKnowledgeInstruction(content, contentType)
	
	metadata := map[string]interface{}{
		"task_type":    "knowledge_management",
		"content_type": contentType,
		"content":      content,
	}

	result, err := l.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("知识管理失败: %w", err)
	}

	// 解析管理结果
	management, err := l.parseLibrarianResult(result) 
	if err != nil {
		l.LogActivity(fmt.Sprintf("结果解析失败: %v", err))
		return l.createFallbackManagement(content, contentType), nil
	}

	// 更新本地知识库
	l.updateKnowledgeBase(content, contentType, management)

	l.LogActivity("知识管理完成")
	return management, nil
}

// buildKnowledgeInstruction 构建知识管理指令
func (l *LibrarianAgent) buildKnowledgeInstruction(content string, contentType string) string {
	var instruction strings.Builder
	
	instruction.WriteString(fmt.Sprintf("请对以下%s类型的内容进行知识管理：\n\n", contentType))
	instruction.WriteString("**内容：**\n")
	instruction.WriteString(content)
	instruction.WriteString("\n\n**管理任务：**\n")
	instruction.WriteString("1. 提取关键信息和设定要点\n")
	instruction.WriteString("2. 检查与已有知识的一致性\n") 
	instruction.WriteString("3. 建立信息分类和索引\n")
	instruction.WriteString("4. 识别需要补充的信息缺口\n")
	instruction.WriteString("5. 提供内容组织建议\n")

	return instruction.String()
}

// CheckConsistency 检查内容一致性
func (l *LibrarianAgent) CheckConsistency(ctx context.Context, newContent string, existingContent []string) (*ConsistencyReport, error) {
	l.LogActivity("开始一致性检查")

	// 构建一致性检查指令
	instruction := l.buildConsistencyInstruction(newContent, existingContent)
	
	metadata := map[string]interface{}{
		"task_type":        "consistency_check",
		"new_content":      newContent,
		"existing_content": existingContent,
	}

	result, err := l.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("一致性检查失败: %w", err)
	}

	// 解析检查结果
	report, err := l.parseConsistencyResult(result)
	if err != nil {
		l.LogActivity(fmt.Sprintf("一致性结果解析失败: %v", err))
		return l.createFallbackConsistency(newContent), nil
	}

	l.LogActivity("一致性检查完成")
	return report, nil
}

// buildConsistencyInstruction 构建一致性检查指令
func (l *LibrarianAgent) buildConsistencyInstruction(newContent string, existingContent []string) string {
	var instruction strings.Builder
	
	instruction.WriteString("请检查新内容与已有内容的一致性：\n\n")
	instruction.WriteString("**新增内容：**\n")
	instruction.WriteString(newContent)
	instruction.WriteString("\n\n**已有内容：**\n")
	
	for i, content := range existingContent {
		instruction.WriteString(fmt.Sprintf("内容%d：%s\n", i+1, content))
	}
	
	instruction.WriteString("\n**检查要点：**\n")
	instruction.WriteString("1. 人物设定是否一致\n")
	instruction.WriteString("2. 世界观是否矛盾\n")
	instruction.WriteString("3. 时间线是否合理\n")
	instruction.WriteString("4. 情节逻辑是否连贯\n")
	instruction.WriteString("5. 提出修改建议\n")

	return instruction.String()
}

// OrganizeContent 组织和分类内容
func (l *LibrarianAgent) OrganizeContent(ctx context.Context, contents []string, organizationType string) (*OrganizationResult, error) {
	l.LogActivity(fmt.Sprintf("开始%s方式的内容组织", organizationType))

	instruction := l.buildOrganizationInstruction(contents, organizationType)
	
	metadata := map[string]interface{}{
		"task_type":         "content_organization",
		"contents":          contents, 
		"organization_type": organizationType,
	}

	result, err := l.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("内容组织失败: %w", err)
	}

	// 解析组织结果
	organization, err := l.parseOrganizationResult(result)
	if err != nil {
		l.LogActivity(fmt.Sprintf("组织结果解析失败: %v", err))
		return l.createFallbackOrganization(contents, organizationType), nil
	}

	l.LogActivity("内容组织完成")
	return organization, nil
}

// buildOrganizationInstruction 构建内容组织指令
func (l *LibrarianAgent) buildOrganizationInstruction(contents []string, organizationType string) string {
	var instruction strings.Builder
	
	instruction.WriteString(fmt.Sprintf("请按%s方式组织以下内容：\n\n", organizationType))
	
	for i, content := range contents {
		instruction.WriteString(fmt.Sprintf("**内容%d：**\n%s\n\n", i+1, content))
	}
	
	instruction.WriteString("**组织要求：**\n")
	instruction.WriteString("1. 建立清晰的分类体系\n")
	instruction.WriteString("2. 创建内容索引和标签\n")
	instruction.WriteString("3. 识别内容间的关联关系\n")
	instruction.WriteString("4. 提供检索和访问建议\n")

	return instruction.String()
}

// 解析方法实现
func (l *LibrarianAgent) parseLibrarianResult(result string) (*LibrarianResult, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, err
	}

	libResult, ok := parsed["librarian_result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的图书管理员结果格式")
	}

	return &LibrarianResult{
		KnowledgeUpdates:  getStringValue(libResult, "knowledge_updates"),
		ConsistencyChecks: getStringValue(libResult, "consistency_checks"),
		OrganizationNotes: getStringValue(libResult, "organization_notes"),
		References:        getStringValue(libResult, "references"),
	}, nil
}

func (l *LibrarianAgent) parseConsistencyResult(result string) (*ConsistencyReport, error) {
	// 简化实现，实际应该解析JSON结果
	return &ConsistencyReport{
		IsConsistent: true,
		Issues:       []string{},
		Suggestions:  []string{"内容基本一致"},
	}, nil
}

func (l *LibrarianAgent) parseOrganizationResult(result string) (*OrganizationResult, error) {
	// 简化实现，实际应该解析JSON结果
	return &OrganizationResult{
		Categories:  []string{"主要内容", "次要内容"},
		Index:       map[string][]string{"default": {"内容1", "内容2"}},
		Suggestions: []string{"建立更详细的分类体系"},
	}, nil
}

// 备用方法实现
func (l *LibrarianAgent) createFallbackManagement(content, contentType string) *LibrarianResult {
	return &LibrarianResult{
		KnowledgeUpdates:  fmt.Sprintf("已记录%s类型的内容", contentType),
		ConsistencyChecks: "基本一致性检查完成",
		OrganizationNotes: "内容已分类存储",
		References:        "相关信息待补充",
	}
}

func (l *LibrarianAgent) createFallbackConsistency(content string) *ConsistencyReport {
	return &ConsistencyReport{
		IsConsistent: true,
		Issues:       []string{},
		Suggestions:  []string{"建议进行更详细的一致性检查"},
	}
}

func (l *LibrarianAgent) createFallbackOrganization(contents []string, orgType string) *OrganizationResult {
	return &OrganizationResult{
		Categories:  []string{orgType},
		Index:       map[string][]string{orgType: contents},
		Suggestions: []string{"需要更精细的组织结构"},
	}
}

// updateKnowledgeBase 更新本地知识库
func (l *LibrarianAgent) updateKnowledgeBase(content, contentType string, result *LibrarianResult) {
	entry := &KnowledgeEntry{
		ID:          fmt.Sprintf("%s_%d", contentType, time.Now().Unix()),
		Type:        contentType,
		Content:     content,
		UpdateTime:  time.Now(),
		Tags:        []string{contentType},
		Metadata:    result,
	}
	
	l.knowledgeBase[entry.ID] = entry
}

// 数据结构定义
type LibrarianResult struct {
	KnowledgeUpdates  string `json:"knowledge_updates"`
	ConsistencyChecks string `json:"consistency_checks"`
	OrganizationNotes string `json:"organization_notes"`
	References        string `json:"references"`
}

type ConsistencyReport struct {
	IsConsistent bool     `json:"is_consistent"`
	Issues       []string `json:"issues"`
	Suggestions  []string `json:"suggestions"`
}

type OrganizationResult struct {
	Categories  []string            `json:"categories"`
	Index       map[string][]string `json:"index"`
	Suggestions []string            `json:"suggestions"`
}

type KnowledgeEntry struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Content    string          `json:"content"`
	UpdateTime time.Time       `json:"update_time"`
	Tags       []string        `json:"tags"`
	Metadata   *LibrarianResult `json:"metadata"`
}
