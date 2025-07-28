package save_novel_rag_data_workflow

import (
	"context"
	"testing"

	"github.com/nvcnvn/adk-golang/pkg/memory"
	"github.com/nvcnvn/adk-golang/pkg/sessions"
)

// MockMemoryService 模拟内存服务，用于测试
type MockMemoryService struct {
	addSessionCalls   []*sessions.Session
	searchMemoryCalls []SearchMemoryCall
}

type SearchMemoryCall struct {
	AppName string
	UserID  string
	Query   string
}

func (m *MockMemoryService) AddSessionToMemory(ctx context.Context, session *sessions.Session) error {
	m.addSessionCalls = append(m.addSessionCalls, session)
	return nil
}

func (m *MockMemoryService) SearchMemory(ctx context.Context, appName, userID, query string) (*memory.SearchMemoryResponse, error) {
	m.searchMemoryCalls = append(m.searchMemoryCalls, SearchMemoryCall{
		AppName: appName,
		UserID:  userID,
		Query:   query,
	})
	return &memory.SearchMemoryResponse{Memories: []*memory.MemoryResult{}}, nil
}

func TestNewSaveNovelRagDataService(t *testing.T) {
	mockMemory := &MockMemoryService{}
	config := SaveNovelRagDataConfig{
		MemoryService: mockMemory,
		LLMModel:      "test-model",
		UserID:        "test-user",
		ArchiveID:     "test-archive",
	}

	service := NewSaveNovelRagDataService(config)
	
	if service == nil {
		t.Fatal("服务创建失败")
	}
	
	if service.config.LLMModel != "test-model" {
		t.Errorf("期望LLM模型为 'test-model'，实际为 '%s'", service.config.LLMModel)
	}
	
	if service.config.UserID != "test-user" {
		t.Errorf("期望用户ID为 'test-user'，实际为 '%s'", service.config.UserID)
	}
}

func TestNewSaveNovelRagDataServiceWithDefaults(t *testing.T) {
	service := NewSaveNovelRagDataServiceWithDefaults("user123", "archive456")
	
	if service == nil {
		t.Fatal("服务创建失败")
	}
	
	config := service.GetConfig()
	if config.UserID != "user123" {
		t.Errorf("期望用户ID为 'user123'，实际为 '%s'", config.UserID)
	}
	
	if config.ArchiveID != "archive456" {
		t.Errorf("期望归档ID为 'archive456'，实际为 '%s'", config.ArchiveID)
	}
	
	if config.LLMModel != "deepseek-chat" {
		t.Errorf("期望默认LLM模型为 'deepseek-chat'，实际为 '%s'", config.LLMModel)
	}
}

func TestExtractJSONFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "纯JSON",
			input:    `{"segments": [{"content": "测试"}]}`,
			expected: `{"segments": [{"content": "测试"}]}`,
		},
		{
			name:     "带前后文本的JSON",
			input:    `这是一段说明文字 {"segments": [{"content": "测试"}]} 还有后续文字`,
			expected: `{"segments": [{"content": "测试"}]}`,
		},
		{
			name:     "无JSON的文本",
			input:    `这里没有JSON内容`,
			expected: `这里没有JSON内容`,
		},
		{
			name:     "嵌套JSON",
			input:    `{"outer": {"inner": {"segments": [{"content": "测试"}]}}}`,
			expected: `{"outer": {"inner": {"segments": [{"content": "测试"}]}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromResponse(tt.input)
			if result != tt.expected {
				t.Errorf("期望结果: %s，实际结果: %s", tt.expected, result)
			}
		})
	}
}

func TestSaveSegmentToRAG(t *testing.T) {
	mockMemory := &MockMemoryService{}
	config := SaveNovelRagDataConfig{
		MemoryService: mockMemory,
		LLMModel:      "test-model",
		UserID:        "test-user",
		ArchiveID:     "test-archive",
	}

	service := NewSaveNovelRagDataService(config)
	ctx := context.Background()

	segment := Segment{
		Content: "主角在神秘的森林中发现了古老的遗迹。",
		Type:    "location",
		Summary: "发现古遗迹",
	}

	err := service.saveSegmentToRAG(ctx, segment, 0)
	if err != nil {
		t.Fatalf("保存段落失败: %v", err)
	}

	if len(mockMemory.addSessionCalls) != 1 {
		t.Fatalf("期望调用AddSessionToMemory 1次，实际调用 %d次", len(mockMemory.addSessionCalls))
	}

	session := mockMemory.addSessionCalls[0]
	if session.AppName != "test-user" {
		t.Errorf("期望AppName为 'test-user'，实际为 '%s'", session.AppName)
	}

	if session.ID != "test-archive_segment_0" {
		t.Errorf("期望ID为 'test-archive_segment_0'，实际为 '%s'", session.ID)
	}

	if len(session.Events) != 1 {
		t.Fatalf("期望1个事件，实际有 %d个", len(session.Events))
	}

	// 检查状态数据
	if session.StateMap["segment_type"] != "location" {
		t.Errorf("期望segment_type为 'location'，实际为 '%v'", session.StateMap["segment_type"])
	}

	if session.StateMap["user_id"] != "test-user" {
		t.Errorf("期望user_id为 'test-user'，实际为 '%v'", session.StateMap["user_id"])
	}

	if session.StateMap["archive_id"] != "test-archive" {
		t.Errorf("期望archive_id为 'test-archive'，实际为 '%v'", session.StateMap["archive_id"])
	}
}

func TestNewSaveNovelRagDataServiceWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  SaveNovelRagDataOptions
		expected SaveNovelRagDataConfig
	}{
		{
			name: "完整选项",
			options: SaveNovelRagDataOptions{
				MemoryService: &MockMemoryService{},
				LLMModel:      "custom-model",
				UserID:        "user123",
				ArchiveID:     "archive123",
			},
			expected: SaveNovelRagDataConfig{
				LLMModel:  "custom-model",
				UserID:    "user123",
				ArchiveID: "archive123",
			},
		},
		{
			name: "默认LLM模型",
			options: SaveNovelRagDataOptions{
				MemoryService: &MockMemoryService{},
				UserID:        "user456",
				ArchiveID:     "archive456",
			},
			expected: SaveNovelRagDataConfig{
				LLMModel:  "deepseek-chat",
				UserID:    "user456",
				ArchiveID: "archive456",
			},
		},
		{
			name: "指定RAG URL和TopK",
			options: SaveNovelRagDataOptions{
				RAGBaseURL: "http://localhost:8080",
				RAGTopK:    20,
				UserID:     "user789",
				ArchiveID:  "archive789",
			},
			expected: SaveNovelRagDataConfig{
				LLMModel:  "deepseek-chat",
				UserID:    "user789",
				ArchiveID: "archive789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSaveNovelRagDataServiceWithOptions(tt.options)
			if service == nil {
				t.Fatal("服务创建失败")
			}

			config := service.GetConfig()
			if config.LLMModel != tt.expected.LLMModel {
				t.Errorf("期望LLM模型为 '%s'，实际为 '%s'", tt.expected.LLMModel, config.LLMModel)
			}
			if config.UserID != tt.expected.UserID {
				t.Errorf("期望用户ID为 '%s'，实际为 '%s'", tt.expected.UserID, config.UserID)
			}
			if config.ArchiveID != tt.expected.ArchiveID {
				t.Errorf("期望归档ID为 '%s'，实际为 '%s'", tt.expected.ArchiveID, config.ArchiveID)
			}
			if config.MemoryService == nil {
				t.Error("MemoryService不应为nil")
			}
		})
	}
}

// BenchmarkExtractJSONFromResponse 性能基准测试
func BenchmarkExtractJSONFromResponse(b *testing.B) {
	input := `这是一段很长的说明文字，包含很多细节。{"segments": [{"content": "测试内容", "type": "test", "summary": "测试摘要"}]} 还有更多的后续文字和说明。`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractJSONFromResponse(input)
	}
}
