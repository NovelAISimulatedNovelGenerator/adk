package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

// TestArchiveIdFullChainAccess 测试archive_id从API层到插件层的完整传递链路
func TestArchiveIdFullChainAccess(t *testing.T) {
	// 记录在插件层捕获到的上下文信息
	var capturedUserID string
	var capturedArchiveID string
	var capturedInput string

	// 创建测试Agent，用于验证插件层能否访问到context中的user_id和archive_id
	testAgent := agents.NewAgent(
		agents.WithName("archive_test_agent"),
		agents.WithInstruction("测试archive_id传递的Agent"),
		agents.WithDescription("验证context中user_id和archive_id的传递"),
		agents.WithBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
			// 从context中获取user_id
			if userID, ok := ctx.Value("user_id").(string); ok {
				capturedUserID = userID
			}
			
			// 从context中获取archive_id
			if archiveID, ok := ctx.Value("archive_id").(string); ok {
				capturedArchiveID = archiveID
			}
			
			capturedInput = msg
			
			// 返回包含捕获信息的响应
			response := fmt.Sprintf("捕获成功: user_id=%s, archive_id=%s, input=%s", 
				capturedUserID, capturedArchiveID, capturedInput)
			return response, true
		}),
	)

	// 创建Manager并注册测试工作流
	mgr := flow.NewManager()
	mgr.Register("archive_test_flow", testAgent)

	// 创建HTTP Server
	server := NewHttpServer(mgr, ":0")
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/execute" {
			server.handleExecute(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	// 准备测试数据
	testUserID := "test_user_12345"
	testArchiveID := "test_archive_67890"
	testInput := "测试archive_id传递链路"

	// 创建HTTP请求体，包含archive_id字段
	requestBody := map[string]interface{}{
		"workflow":   "archive_test_flow",
		"input":      testInput,
		"user_id":    testUserID,
		"archive_id": testArchiveID,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("序列化请求体失败: %v", err)
	}

	// 发送HTTP请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(testServer.URL+"/api/execute", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 验证HTTP响应状态
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际得到 %d", resp.StatusCode)
	}

	// 解析响应
	var apiResp WorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应成功
	if !apiResp.Success {
		t.Fatalf("工作流执行失败: %s", apiResp.Message)
	}

	// **关键验证：检查插件层是否成功捕获到user_id和archive_id**
	if capturedUserID != testUserID {
		t.Errorf("插件层捕获的user_id不匹配: 期望 %s, 实际 %s", testUserID, capturedUserID)
	}

	if capturedArchiveID != testArchiveID {
		t.Errorf("插件层捕获的archive_id不匹配: 期望 %s, 实际 %s", testArchiveID, capturedArchiveID)
	}

	if capturedInput != testInput {
		t.Errorf("插件层捕获的input不匹配: 期望 %s, 实际 %s", testInput, capturedInput)
	}

	// 验证响应输出包含正确信息
	expectedOutputContent := fmt.Sprintf("user_id=%s", testUserID)
	if !containsString(apiResp.Output, expectedOutputContent) {
		t.Errorf("响应输出不包含期望的user_id信息: %s", apiResp.Output)
	}

	expectedArchiveContent := fmt.Sprintf("archive_id=%s", testArchiveID)
	if !containsString(apiResp.Output, expectedArchiveContent) {
		t.Errorf("响应输出不包含期望的archive_id信息: %s", apiResp.Output)
	}

	t.Logf("✅ archive_id全链路传递测试成功!")
	t.Logf("   - API层接收: user_id=%s, archive_id=%s", testUserID, testArchiveID)
	t.Logf("   - 插件层捕获: user_id=%s, archive_id=%s", capturedUserID, capturedArchiveID)
	t.Logf("   - 响应输出: %s", apiResp.Output)
}

// containsString 检查字符串s是否包含子字符串substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
