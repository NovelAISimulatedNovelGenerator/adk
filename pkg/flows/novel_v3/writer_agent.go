//go:build ignore
// +build ignore

package novel_v3

import (
	"fmt"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/models"
)

// WriterAgent 负责撰写小说的章节。
type WriterAgent struct {
	llm *models.Model
}

// NewWriterAgent 创建一个新的 WriterAgent。
func NewWriterAgent(llm *models.Model) *WriterAgent {
	return &WriterAgent{llm: llm}
}

// WriteChapter 根据 StoryBible 和章节大纲撰写一个章节。
func (w *WriterAgent) WriteChapter(bible *StoryBible, chapter int) (string, error) {
	// TODO: 构建一个更复杂的提示，包括角色、设定和之前的摘要。
	plotPoint := bible.Outline[chapter-1]
	prompt := fmt.Sprintf("根据这个情节要点写第%d章：'%s'", chapter, plotPoint.Description)

	chapterText, err := w.llm.Generate(prompt)
	if err != nil {
		return "", err
	}

	fmt.Printf("Writer Agent 正在撰写第 %d 章...\n", chapter)

	return chapterText, nil
}

