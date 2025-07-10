package test_rag_flow

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/events"
	"github.com/nvcnvn/adk-golang/pkg/memory"
	"github.com/nvcnvn/adk-golang/pkg/models"
	"github.com/nvcnvn/adk-golang/pkg/sessions"
)

// Build 构造一个用于验证 CustomRagMemoryService 的简单工作流。
// 输入任何文本，将查询自定义 RAG 服务并返回检索结果 JSON。
func Build() *agents.Agent {
	// 读取 RAG 服务地址，可通过环境变量覆盖。
	baseURL := os.Getenv("RAG_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:18000"
	}

	ragMem := memory.NewCustomRagMemoryService(baseURL, 5)

	ragAgent := agents.NewAgent(
		agents.WithName("rag_test_agent"),
		agents.WithModel("deepseek-chat"), // 保留字段以与其他代理保持一致
		agents.WithInstruction("输入任意查询，先将其存入 RAG，再搜索并返回结果。"),
		agents.WithDescription("RAG 测试代理"),
		agents.WithBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
			// 1. 生成一段测试内容并写入 RAG
			testContent := fmt.Sprintf("自动生成测试内容 %d", time.Now().UnixNano())
			session := sessions.NewSession("rag_test", "", nil, "")
			ev := events.NewEvent()
			ev.Author = "ai"
			ev.Content = &models.Content{Parts: []*models.Part{{Text: testContent}}}
			session.AddEvent(ev)
			if err := ragMem.AddSessionToMemory(ctx, session); err != nil {
				return fmt.Sprintf("{\"error\":%q}", err.Error()), true
			}

			// 2. 构造会话并写入 RAG
			session = sessions.NewSession("rag_test", "", nil, "")
			ev = events.NewEvent()
			ev.Author = "user"
			ev.Content = &models.Content{Parts: []*models.Part{{Text: msg}}}
			session.AddEvent(ev)
			if err := ragMem.AddSessionToMemory(ctx, session); err != nil {
				return fmt.Sprintf("{\"error\":%q}", err.Error()), true
			}

			// 2. 搜索相似内容
			resp, err := ragMem.SearchMemory(ctx, "rag_test", "", testContent)
			if err != nil {
				return fmt.Sprintf("{\"error\":%q}", err.Error()), true
			}

			// 3. 格式化输出
			var sb strings.Builder
			sb.WriteString("{\"results\":[")
			for i, m := range resp.Memories {
				if len(m.Events) == 0 || m.Events[0].Content == nil || len(m.Events[0].Content.Parts) == 0 {
					continue
				}
				if i > 0 {
					sb.WriteString(",")
				}
				text := m.Events[0].Content.Parts[0].Text
				sb.WriteString(fmt.Sprintf("%q", text))
			}
			sb.WriteString("]}")
			return sb.String(), true
		}),
	)

	return ragAgent
}
