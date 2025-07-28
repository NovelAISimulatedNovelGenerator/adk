package save_novel_rag_data_workflow

import (
	"context"

	"github.com/nvcnvn/adk-golang/pkg/memory"
)

// SaveNovelRagDataOptions 构建选项
type SaveNovelRagDataOptions struct {
	MemoryService memory.MemoryService
	LLMModel      string
	UserID        string
	ArchiveID     string
	RAGBaseURL    string
	RAGTopK       int
}

// NewSaveNovelRagDataServiceWithDefaults 使用默认RAG服务创建实例
// 这是最常用的构造函数，自动创建CustomRagMemoryService
func NewSaveNovelRagDataServiceWithDefaults(userID, archiveID string) *SaveNovelRagDataService {
	ragService := memory.NewCustomRagMemoryServiceWithDefaults()
	
	config := SaveNovelRagDataConfig{
		MemoryService: ragService,
		LLMModel:      "deepseek-chat",
		UserID:        userID,
		ArchiveID:     archiveID,
	}
	
	return NewSaveNovelRagDataService(config)
}

// NewSaveNovelRagDataServiceWithOptions 使用选项创建实例
func NewSaveNovelRagDataServiceWithOptions(options SaveNovelRagDataOptions) *SaveNovelRagDataService {
	// 如果没有提供MemoryService，创建默认的
	if options.MemoryService == nil {
		if options.RAGBaseURL != "" {
			if options.RAGTopK > 0 {
				options.MemoryService = memory.NewCustomRagMemoryService(options.RAGBaseURL, options.RAGTopK)
			} else {
				options.MemoryService = memory.NewCustomRagMemoryServiceWithURL(options.RAGBaseURL)
			}
		} else if options.RAGTopK > 0 {
			options.MemoryService = memory.NewCustomRagMemoryServiceWithTopK(options.RAGTopK)
		} else {
			options.MemoryService = memory.NewCustomRagMemoryServiceWithDefaults()
		}
	}
	
	if options.LLMModel == "" {
		options.LLMModel = "deepseek-chat"
	}
	
	config := SaveNovelRagDataConfig{
		MemoryService: options.MemoryService,
		LLMModel:      options.LLMModel,
		UserID:        options.UserID,
		ArchiveID:     options.ArchiveID,
	}
	
	return NewSaveNovelRagDataService(config)
}

// QuickSave 快速保存小说内容的便利函数
// 这是一个无状态的便利函数，用于快速保存内容
func QuickSave(ctx context.Context, userID, archiveID, content string, contentType ...string) (int, error) {
	service := NewSaveNovelRagDataServiceWithDefaults(userID, archiveID)
	return service.SaveNovelData(ctx, content, contentType...)
}

// QuickSaveWithRAGURL 使用指定RAG服务URL快速保存内容
func QuickSaveWithRAGURL(ctx context.Context, ragBaseURL, userID, archiveID, content string, contentType ...string) (int, error) {
	options := SaveNovelRagDataOptions{
		RAGBaseURL: ragBaseURL,
		UserID:     userID,
		ArchiveID:  archiveID,
	}
	service := NewSaveNovelRagDataServiceWithOptions(options)
	return service.SaveNovelData(ctx, content, contentType...)
}
