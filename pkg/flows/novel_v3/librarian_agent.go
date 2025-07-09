//go:build ignore
// +build ignore

package novel_v3

import (
	"fmt"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/models"
)

// LibrarianAgent 负责在每个章节后更新 StoryBible。
type LibrarianAgent struct {
	llm *models.Model
}

// NewLibrarianAgent 创建一个新的 LibrarianAgent。
func NewLibrarianAgent(llm *models.Model) *LibrarianAgent {
	return &LibrarianAgent{llm: llm}
}

// UpdateBible 读取一个章节并更新 StoryBible。
func (l *LibrarianAgent) UpdateBible(bible *StoryBible, chapter int, chapterText string) error {
	prompt := fmt.Sprintf("为以下章节文本创建一个简洁的摘要：\n\n%s", chapterText)

	summary, err := l.llm.Generate(prompt)
	if err != nil {
		return err
	}

	bible.ChapterSummaries[chapter] = summary

	// TODO: 实现更复杂的更新逻辑，比如更新角色状态。

	fmt.Printf("Librarian Agent 正在为第 %d 章生成摘要...\n", chapter)

	return nil
}

