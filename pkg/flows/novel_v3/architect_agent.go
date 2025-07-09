//go:build ignore
// +build ignore

package novel_v3

import (
	"fmt"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/models"
)

// ArchitectAgent 负责创建故事的初始结构。
type ArchitectAgent struct {
	llm *models.Model
}

// NewArchitectAgent 创建一个新的 ArchitectAgent。
func NewArchitectAgent(llm *models.Model) *ArchitectAgent {
	return &ArchitectAgent{llm: llm}
}

// Plan 创建故事的 StoryBible。
func (a *ArchitectAgent) Plan(premise string) (*StoryBible, error) {
	prompt := fmt.Sprintf("基于这个前提：'%s'，为一部小说创建一个详细的 Story Bible。包括标题、设定、至少3个角色的角色卡，以及一个至少5个章节的情节大纲。", premise)

	response, err := a.llm.Generate(prompt)
	if err != nil {
		return nil, err
	}

	// 在一个真实的实现中，我们会在这里解析 LLM 的响应
	// 并填充 StoryBible 结构。现在，我们只返回一个占位符。
	// TODO: 实现从 LLM 响应到 StoryBible 的解析逻辑。
	bible := &StoryBible{
		Title:   "一个由AI生成的故事",
		Premise: premise,
		Setting: "一个由AI想象的世界",
		Characters: []*CharacterSheet{
			{Name: "主角", Description: "勇敢的英雄"},
		},
		Outline: []*PlotPoint{
			{Chapter: 1, Description: "故事的开始"},
		},
		ChapterSummaries: make(map[int]string),
	}

	fmt.Println("Architect Agent 生成的计划：")
	fmt.Printf("LLM 响应: %s\n", response)

	return bible, nil
}
