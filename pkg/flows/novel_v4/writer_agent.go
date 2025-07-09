package novel_v4

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// WriterAgent 写作者智能体，负责具体的内容创作
type WriterAgent struct {
	*BaseAgent
}

// NewWriterAgent 创建新的写作者智能体
func NewWriterAgent() *WriterAgent {
	capabilities := []string{
		"文本创作",
		"对话生成",
		"场景描写",
		"情节发展",
		"文风调整",
		"内容润色",
	}

	instruction := `你是小说写作者智能体，专门负责具体的文本内容创作。你的核心职责包括：

1. **内容创作**：根据架构设计创作具体的小说内容
2. **对话生成**：创作符合人物性格的自然对话
3. **场景描写**：生动细致地描绘场景和氛围
4. **情节推进**：确保情节发展的连贯性和逻辑性
5. **文风统一**：保持整体文风的一致性
6. **内容润色**：优化文本表达和语言质量

创作要求：
- 根据提供的架构设计进行创作
- 保持人物性格的一致性
- 情节发展要合理自然
- 语言表达要生动有趣
- 注意节奏控制和情绪渲染

输出格式：
{"writer_result": {"content": "创作的具体内容", "style_notes": "文风说明", "character_voices": "人物对话特色", "scene_details": "场景描写要点"}}`

	base := NewBaseAgent(
		"writer_v4",
		"写作者",
		instruction,
		"Novel v4 写作者智能体 - 负责具体内容创作",
		capabilities,
	)

	return &WriterAgent{BaseAgent: base}
}

// WriteContent 根据架构设计创作内容
func (w *WriterAgent) WriteContent(ctx context.Context, design *ArchitectureDesign, chapter int) (*WritingResult, error) {
	w.LogActivity(fmt.Sprintf("开始创作第%d章内容", chapter))

	// 构建创作指令
	instruction := w.buildWritingInstruction(design, chapter)
	
	metadata := map[string]interface{}{
		"task_type": "content_writing",
		"chapter":   chapter,
		"design":    design,
	}

	result, err := w.ProcessWithContext(ctx, instruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("内容创作失败: %w", err)
	}

	// 解析创作结果
	writing, err := w.parseWritingResult(result)
	if err != nil {
		w.LogActivity(fmt.Sprintf("结果解析失败: %v", err))
		// 返回基础创作结果
		return w.createFallbackWriting(design, chapter), nil
	}

	w.LogActivity(fmt.Sprintf("第%d章创作完成", chapter))
	return writing, nil
}

// buildWritingInstruction 构建创作指令
func (w *WriterAgent) buildWritingInstruction(design *ArchitectureDesign, chapter int) string {
	var instruction strings.Builder
	
	instruction.WriteString(fmt.Sprintf("请根据以下架构设计创作第%d章内容：\n\n", chapter))
	instruction.WriteString(fmt.Sprintf("**世界观设定：**\n%s\n\n", design.Worldview))
	instruction.WriteString(fmt.Sprintf("**人物设计：**\n%s\n\n", design.Characters))
	instruction.WriteString(fmt.Sprintf("**情节结构：**\n%s\n\n", design.PlotStructure))
	instruction.WriteString(fmt.Sprintf("**主题设定：**\n%s\n\n", design.Themes))
	instruction.WriteString(fmt.Sprintf("**章节结构：**\n%s\n\n", design.Chapters))
	
	instruction.WriteString("创作要求：\n")
	instruction.WriteString("1. 内容要符合整体架构设计\n")
	instruction.WriteString("2. 保持人物性格的一致性\n")
	instruction.WriteString("3. 情节发展要自然合理\n")
	instruction.WriteString("4. 语言表达要生动有趣\n")
	instruction.WriteString("5. 字数控制在1000-2000字\n")

	return instruction.String()
}

// parseWritingResult 解析创作结果
func (w *WriterAgent) parseWritingResult(result string) (*WritingResult, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, err
	}

	writerResult, ok := parsed["writer_result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的创作结果格式")
	}

	writing := &WritingResult{
		Content:         getStringValue(writerResult, "content"),
		StyleNotes:      getStringValue(writerResult, "style_notes"),
		CharacterVoices: getStringValue(writerResult, "character_voices"),
		SceneDetails:    getStringValue(writerResult, "scene_details"),
	}

	return writing, nil
}

// createFallbackWriting 创建备用创作结果
func (w *WriterAgent) createFallbackWriting(design *ArchitectureDesign, chapter int) *WritingResult {
	return &WritingResult{
		Content:         fmt.Sprintf("第%d章内容创作中，基于架构设计进行展开...", chapter),
		StyleNotes:      "采用现代小说创作风格",
		CharacterVoices: "根据人物设定调整对话风格",
		SceneDetails:    "注重场景氛围的营造",
	}
}

// RefineContent 内容润色和优化
func (w *WriterAgent) RefineContent(ctx context.Context, content string, requirements []string) (*WritingResult, error) {
	w.LogActivity("开始内容润色")

	refinementInstruction := w.buildRefinementInstruction(content, requirements)
	
	metadata := map[string]interface{}{
		"task_type":     "content_refinement",
		"original":      content,
		"requirements":  requirements,
	}

	result, err := w.ProcessWithContext(ctx, refinementInstruction, metadata)
	if err != nil {
		return nil, fmt.Errorf("内容润色失败: %w", err)
	}

	// 解析润色结果
	refined, err := w.parseWritingResult(result)
	if err != nil {
		w.LogActivity(fmt.Sprintf("润色结果解析失败: %v", err))
		return &WritingResult{Content: content}, nil
	}

	w.LogActivity("内容润色完成")
	return refined, nil
}

// buildRefinementInstruction 构建润色指令
func (w *WriterAgent) buildRefinementInstruction(content string, requirements []string) string {
	var instruction strings.Builder
	
	instruction.WriteString("请对以下内容进行润色优化：\n\n")
	instruction.WriteString("**原始内容：**\n")
	instruction.WriteString(content)
	instruction.WriteString("\n\n**优化要求：**\n")
	
	for i, req := range requirements {
		instruction.WriteString(fmt.Sprintf("%d. %s\n", i+1, req))
	}
	
	instruction.WriteString("\n请保持原有故事情节，专注于语言表达和文风优化。")
	
	return instruction.String()
}

// GenerateDialogue 生成对话内容
func (w *WriterAgent) GenerateDialogue(ctx context.Context, characters []string, situation string) (string, error) {
	w.LogActivity("开始生成对话")

	dialogueInstruction := fmt.Sprintf(`请为以下情况生成对话：

**参与角色：** %s
**情境：** %s

要求：
1. 对话要符合各角色的性格特点
2. 推进情节发展
3. 自然流畅，避免生硬
4. 每句对话不超过50字
5. 用"角色名："格式标注发言人

`, strings.Join(characters, "、"), situation)

	metadata := map[string]interface{}{
		"task_type":   "dialogue_generation",
		"characters":  characters,
		"situation":   situation,
	}

	dialogue, err := w.ProcessWithContext(ctx, dialogueInstruction, metadata)
	if err != nil {
		return "", fmt.Errorf("对话生成失败: %w", err)
	}

	w.LogActivity("对话生成完成")
	return dialogue, nil
}

// WritingResult 创作结果结构
type WritingResult struct {
	Content         string `json:"content"`           // 创作内容
	StyleNotes      string `json:"style_notes"`       // 文风说明
	CharacterVoices string `json:"character_voices"`  // 人物对话特色
	SceneDetails    string `json:"scene_details"`     // 场景描写要点
}

// ToJSON 将创作结果转换为JSON格式
func (r *WritingResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetWordCount 获取内容字数
func (r *WritingResult) GetWordCount() int {
	return len([]rune(r.Content))
}

// IsEmpty 检查创作结果是否为空
func (r *WritingResult) IsEmpty() bool {
	return strings.TrimSpace(r.Content) == ""
}
