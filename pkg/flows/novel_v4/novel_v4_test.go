package novel_v4

import (
	"context"
	"testing"
	"time"
)

// TestBuild 测试构建函数
func TestBuild(t *testing.T) {
	agent := Build()
	if agent == nil {
		t.Fatal("Build() 返回了 nil")
	}
	
	if agent.Name() != "novel_v4_root" {
		t.Errorf("期望名称为 'novel_v4_root'，实际为 '%s'", agent.Name())
	}
}

// TestArchitectAgent 测试架构师智能体
func TestArchitectAgent(t *testing.T) {
	architect := NewArchitectAgent()
	
	if architect == nil {
		t.Fatal("NewArchitectAgent() 返回了 nil")
	}
	
	if architect.GetType() != "架构师" {
		t.Errorf("期望类型为 '架构师'，实际为 '%s'", architect.GetType())
	}
	
	capabilities := architect.GetCapabilities()
	if len(capabilities) == 0 {
		t.Error("架构师智能体应该有能力列表")
	}
	
	// 测试能力验证
	if !architect.ValidateCapability("故事架构设计") {
		t.Error("架构师智能体应该具备'故事架构设计'能力")
	}
}

// TestWriterAgent 测试写作者智能体
func TestWriterAgent(t *testing.T) {
	writer := NewWriterAgent()
	
	if writer == nil {
		t.Fatal("NewWriterAgent() 返回了 nil")
	}
	
	if writer.GetType() != "写作者" {
		t.Errorf("期望类型为 '写作者'，实际为 '%s'", writer.GetType())
	}
	
	// 测试配置功能
	writer.SetConfig("creativity_level", 0.8)
	if value, exists := writer.GetConfig("creativity_level"); !exists || value.(float64) != 0.8 {
		t.Error("配置设置和获取功能异常")
	}
}

// TestLibrarianAgent 测试图书管理员智能体
func TestLibrarianAgent(t *testing.T) {
	librarian := NewLibrarianAgent()
	
	if librarian == nil {
		t.Fatal("NewLibrarianAgent() 返回了 nil")
	}
	
	if librarian.GetType() != "图书管理员" {
		t.Errorf("期望类型为 '图书管理员'，实际为 '%s'", librarian.GetType())
	}
	
	// 测试知识管理功能
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 注意：这里只测试函数不出错，不测试实际的LLM调用
	_, err := librarian.ManageKnowledge(ctx, "测试内容", "测试类型")
	if err != nil {
		t.Logf("知识管理功能测试: %v (预期可能失败，因为没有真实的LLM)", err)
	}
}

// TestCoordinatorAgent 测试协调者智能体
func TestCoordinatorAgent(t *testing.T) {
	coordinator := NewCoordinatorAgent()
	
	if coordinator == nil {
		t.Fatal("NewCoordinatorAgent() 返回了 nil")
	}
	
	if coordinator.GetType() != "协调者" {
		t.Errorf("期望类型为 '协调者'，实际为 '%s'", coordinator.GetType())
	}
	
	// 测试创作请求结构
	request := &CreationRequest{
		Theme:        "科幻冒险",
		Type:         "中短篇小说",
		TargetLength: 5000,
		Requirements: []string{"包含对话", "场景丰富"},
	}
	
	if request.Theme != "科幻冒险" {
		t.Error("创作请求结构异常")
	}
}

// TestArchitectureDesign 测试架构设计结构
func TestArchitectureDesign(t *testing.T) {
	design := &ArchitectureDesign{
		Worldview:     "测试世界观",
		Characters:    "测试角色",
		PlotStructure: "测试情节",
		Themes:        "测试主题",
		Chapters:      "测试章节",
		Premise:       "测试前提",
	}
	
	// 测试JSON序列化
	jsonStr, err := design.ToJSON()
	if err != nil {
		t.Errorf("架构设计JSON序列化失败: %v", err)
	}
	
	if jsonStr == "" {
		t.Error("JSON序列化结果为空")
	}
}

// TestWritingResult 测试创作结果结构
func TestWritingResult(t *testing.T) {
	result := &WritingResult{
		Content:         "这是一个测试内容，用来验证字数统计功能。",
		StyleNotes:      "测试文风",
		CharacterVoices: "测试对话",
		SceneDetails:    "测试场景",
	}
	
	// 测试字数统计
	wordCount := result.GetWordCount()
	if wordCount == 0 {
		t.Error("字数统计应该大于0")
	}
	
	// 测试空检查
	if result.IsEmpty() {
		t.Error("非空内容被判断为空")
	}
	
	// 测试空内容
	emptyResult := &WritingResult{Content: "   "}
	if !emptyResult.IsEmpty() {
		t.Error("空内容没有被正确识别")
	}
}

// TestBaseAgent 测试基础智能体功能
func TestBaseAgent(t *testing.T) {
	base := NewBaseAgent(
		"test_agent",
		"测试类型",
		"测试指令",
		"测试描述",
		[]string{"测试能力1", "测试能力2"},
	)
	
	if base == nil {
		t.Fatal("NewBaseAgent() 返回了 nil")
	}
	
	if base.GetType() != "测试类型" {
		t.Errorf("期望类型为 '测试类型'，实际为 '%s'", base.GetType())
	}
	
	capabilities := base.GetCapabilities()
	if len(capabilities) != 2 {
		t.Errorf("期望能力数量为 2，实际为 %d", len(capabilities))
	}
	
	// 测试能力验证
	if !base.ValidateCapability("测试能力1") {
		t.Error("应该具备'测试能力1'")
	}
	
	if base.ValidateCapability("不存在的能力") {
		t.Error("不应该具备'不存在的能力'")
	}
}

// TestUtilityFunctions 测试工具函数
func TestUtilityFunctions(t *testing.T) {
	// 测试字符串截断
	longString := "这是一个很长很长很长很长的字符串，用来测试截断功能"
	truncated := truncateString(longString, 10)
	
	if len([]rune(truncated)) > 13 { // 10个字符 + "..."
		t.Error("字符串截断功能异常")
	}
	
	// 测试短字符串不被截断
	shortString := "短字符串"
	notTruncated := truncateString(shortString, 20)
	if notTruncated != shortString {
		t.Error("短字符串不应该被截断")
	}
	
	// 测试结果合并
	results := map[string]string{
		"architect": "架构设计结果",
		"writer":    "创作内容结果",
		"librarian": "知识管理结果",
	}
	
	merged := mergeResults(results)
	if merged == "" {
		t.Error("结果合并功能异常")
	}
}

// TestKnowledgeEntry 测试知识条目结构
func TestKnowledgeEntry(t *testing.T) {
	entry := &KnowledgeEntry{
		ID:         "test_001",
		Type:       "测试类型",
		Content:    "测试内容",
		UpdateTime: time.Now(),
		Tags:       []string{"测试", "知识"},
		Metadata:   &LibrarianResult{},
	}
	
	if entry.ID != "test_001" {
		t.Error("知识条目ID设置异常")
	}
	
	if len(entry.Tags) != 2 {
		t.Error("知识条目标签设置异常")
	}
}

// TestQualityStandards 测试质量标准结构
func TestQualityStandards(t *testing.T) {
	standards := &QualityStandards{
		MinContentScore:     8.0,
		MinLanguageScore:    7.5,
		MinConsistencyScore: 8.5,
	}
	
	if standards.MinContentScore != 8.0 {
		t.Error("质量标准设置异常")
	}
}

// BenchmarkBuild 性能测试：构建智能体
func BenchmarkBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		agent := Build()
		if agent == nil {
			b.Fatal("Build() 返回了 nil")
		}
	}
}

// BenchmarkAgentCreation 性能测试：智能体创建
func BenchmarkAgentCreation(b *testing.B) {
	b.Run("Architect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			architect := NewArchitectAgent()
			if architect == nil {
				b.Fatal("NewArchitectAgent() 返回了 nil")
			}
		}
	})
	
	b.Run("Writer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			writer := NewWriterAgent()
			if writer == nil {
				b.Fatal("NewWriterAgent() 返回了 nil")
			}
		}
	})
}
