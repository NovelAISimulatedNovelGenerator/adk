package novel_v3

// StoryBible 是我们故事的单一事实来源。
// 它包含了所有关于故事的元数据。
type StoryBible struct {
	Title         string            `json:"title"`
	Premise       string            `json:"premise"`
	Setting       string            `json:"setting"`
	Characters    []*CharacterSheet `json:"characters"`
	Outline       []*PlotPoint      `json:"outline"`
	ChapterSummaries map[int]string    `json:"chapter_summaries"`
}

// CharacterSheet 定义了故事中的一个角色。
type CharacterSheet struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Motivation  string `json:"motivation"`
	History     string `json:"history"`
}

// PlotPoint 是故事大纲中的一个情节节点。
type PlotPoint struct {
	Chapter      int    `json:"chapter"`
	Description  string `json:"description"`
	KeyEvents    []string `json:"key_events"`
}
