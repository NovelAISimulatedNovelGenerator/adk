package save_novel_rag_data_workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/events"
	"github.com/nvcnvn/adk-golang/pkg/memory"
	"github.com/nvcnvn/adk-golang/pkg/models"
	"github.com/nvcnvn/adk-golang/pkg/sessions"
)

// SaveNovelRagDataConfig 保存小说RAG数据的配置
type SaveNovelRagDataConfig struct {
	// MemoryService RAG内存服务实例
	MemoryService memory.MemoryService
	// LLMModel 用于内容分析的LLM模型名称
	LLMModel string
	// UserID 用户ID，对应RAG的tenant_id
	UserID string
	// ArchiveID 归档ID，对应RAG的session_id
	ArchiveID string
}

// SegmentationResult LLM分段解析结果
type SegmentationResult struct {
	Segments []Segment `json:"segments"`
}

// Segment 单个文本段落
type Segment struct {
	Content string `json:"content"`
	Type    string `json:"type,omitempty"`    // 内容类型：character, location, plot, etc.
	Summary string `json:"summary,omitempty"` // 段落摘要
}

// SaveNovelRagDataService 小说RAG数据保存服务
type SaveNovelRagDataService struct {
	config       SaveNovelRagDataConfig
	segmentAgent *agents.Agent
}

// NewSaveNovelRagDataService 创建新的服务实例
func NewSaveNovelRagDataService(config SaveNovelRagDataConfig) *SaveNovelRagDataService {
	if config.LLMModel == "" {
		config.LLMModel = "deepseek-chat"
	}

	return &SaveNovelRagDataService{
		config:       config,
		segmentAgent: createSegmentAgent(config.LLMModel),
	}
}

// createSegmentAgent 创建用于分段的智能体
func createSegmentAgent(model string) *agents.Agent {
	return agents.NewAgent(
		agents.WithName("segment_agent"),
		agents.WithModel(model),
		agents.WithInstruction(`你是小说内容分析专家。请将输入的小说内容分解成多个语义完整的段落，便于向量化存储。

输出要求：
1. 每个段落应该是语义完整的单元（2-5句话）
2. 保持原文的重要细节和信息
3. 为每个段落标注类型（character/location/plot/dialogue/setting/other）
4. 提供简短的段落摘要
5. 输出标准JSON格式：
{
  "segments": [
    {
      "content": "段落原文内容",
      "type": "段落类型",
      "summary": "段落摘要"
    }
  ]
}

注意：输出必须是有效的JSON格式，不要包含任何其他文本。`),
		agents.WithDescription("内容分段智能体"),
	)
}

// SaveNovelData 保存小说数据到RAG系统
// content: 需要保存的小说内容
// contentType: 可选，指定要重点提取的内容类型（如"location"表示重点提取地点信息）
func (s *SaveNovelRagDataService) SaveNovelData(ctx context.Context, content string, contentType ...string) (int, error) {
	if strings.TrimSpace(content) == "" {
		return 0, fmt.Errorf("输入内容不能为空")
	}

	// 构建LLM输入提示
	prompt := content
	if len(contentType) > 0 && contentType[0] != "" {
		prompt = fmt.Sprintf("请重点关注并提取【%s】相关的信息。\n\n原始内容：\n%s", contentType[0], content)
	}

	log.Printf("[SaveNovelRagData] 开始处理内容，长度: %d字符", len(content))
	if len(contentType) > 0 && contentType[0] != "" {
		log.Printf("[SaveNovelRagData] 重点提取类型: %s", contentType[0])
	}

	// 调用LLM进行分段
	response, err := s.segmentAgent.Process(ctx, prompt)
	if err != nil {
		return 0, fmt.Errorf("LLM分段处理失败: %w", err)
	}

	// 解析JSON结果
	var result SegmentationResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("[SaveNovelRagData] JSON解析失败，尝试提取JSON部分")
		// 尝试从响应中提取JSON部分
		cleaned := extractJSONFromResponse(response)
		if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
			return 0, fmt.Errorf("JSON解析失败: %w，原始响应: %s", err, response)
		}
	}

	if len(result.Segments) == 0 {
		return 0, fmt.Errorf("未解析到有效段落")
	}

	log.Printf("[SaveNovelRagData] 成功解析到 %d 个段落", len(result.Segments))

	// 逐个保存段落到RAG
	savedCount := 0
	for i, segment := range result.Segments {
		if err := s.saveSegmentToRAG(ctx, segment, i); err != nil {
			log.Printf("[SaveNovelRagData] 保存段落 %d 失败: %v", i+1, err)
			continue
		}
		savedCount++
	}

	log.Printf("[SaveNovelRagData] 成功保存 %d/%d 个段落", savedCount, len(result.Segments))
	return savedCount, nil
}

// saveSegmentToRAG 保存单个段落到RAG系统
func (s *SaveNovelRagDataService) saveSegmentToRAG(ctx context.Context, segment Segment, index int) error {
	// 构建session数据
	session := &sessions.Session{
		AppName: s.config.UserID,                                             // 使用UserID作为tenant_id
		UserID:  s.config.UserID,                                             // 设置用户ID
		ID:      fmt.Sprintf("%s_segment_%d", s.config.ArchiveID, index),    // 设置会话ID
		Events: []*events.Event{
			{
				Author: "user",
				Content: &models.Content{
					Parts: []*models.Part{
						{Text: fmt.Sprintf("类型: %s\n摘要: %s\n内容: %s", 
							segment.Type, segment.Summary, segment.Content)},
					},
				},
			},
		},
		StateMap: map[string]interface{}{
			"segment_type":    segment.Type,
			"segment_summary": segment.Summary,
			"archive_id":      s.config.ArchiveID,
			"user_id":         s.config.UserID,
		},
	}

	// 调用RAG服务保存
	return s.config.MemoryService.AddSessionToMemory(ctx, session)
}

// extractJSONFromResponse 从LLM响应中提取JSON部分
func extractJSONFromResponse(response string) string {
	response = strings.TrimSpace(response)
	
	// 查找JSON开始和结束标记
	start := strings.Index(response, "{")
	if start == -1 {
		return response
	}
	
	// 从最后一个}开始查找
	end := strings.LastIndex(response, "}")
	if end == -1 || end <= start {
		return response
	}
	
	return response[start : end+1]
}

// GetConfig 获取当前配置（用于测试等场景）
func (s *SaveNovelRagDataService) GetConfig() SaveNovelRagDataConfig {
	return s.config
}
