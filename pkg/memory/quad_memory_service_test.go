package memory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	// Test cases
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectError    bool
		cancelContext  bool
	}{
		{
			name: "Successful Health Check",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError: false,
		},
		{
			name: "Failed Health Check (500)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectError: true,
		},
		{
			name: "Context Cancelled",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			expectError:   true,
			cancelContext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			config := QuadMemoryConfig{
				BaseURL: server.URL,
			}
			service := NewQuadMemoryService(config)

			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.cancelContext {
				ctx, cancel = context.WithCancel(ctx)
				cancel() // Cancel immediately
			}

			err := service.HealthCheck(ctx)

			if (err != nil) != tt.expectError {
				t.Errorf("HealthCheck() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestAddQuad(t *testing.T) {
	t.Parallel()

	// 测试用例
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		quad           Quad
		hierarchicalCtx *HierarchicalContext
		expectError    bool
		expectedQuad   *Quad
	}{
		{
			name: "成功添加四元组",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("期望 POST 请求，得到 %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
					t.Errorf("期望 Content-Type: application/x-www-form-urlencoded")
				}
				w.WriteHeader(http.StatusOK)
			},
			quad: Quad{
				Subject:   "http://example.com/subject",
				Predicate: "http://example.com/predicate",
				Object:    "http://example.com/object",
			},
			hierarchicalCtx: &HierarchicalContext{
				TenantID: "test-tenant",
				StoryID:  "test-story",
			},
			expectError: false,
		},
		{
			name: "无效的层次化上下文",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			quad: Quad{
				Subject:   "http://example.com/subject",
				Predicate: "http://example.com/predicate",
				Object:    "http://example.com/object",
			},
			hierarchicalCtx: &HierarchicalContext{}, // 缺少 TenantID
			expectError:     true,
		},
		{
			name: "缺少必需字段",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			quad: Quad{
				Subject:   "http://example.com/subject",
				// 缺少 Predicate 和 Object
			},
			hierarchicalCtx: &HierarchicalContext{
				TenantID: "test-tenant",
			},
			expectError: true,
		},
		{
			name: "服务器错误 (500)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			quad: Quad{
				Subject:   "http://example.com/subject",
				Predicate: "http://example.com/predicate",
				Object:    "http://example.com/object",
			},
			hierarchicalCtx: &HierarchicalContext{
				TenantID: "test-tenant",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			config := QuadMemoryConfig{
				BaseURL:      server.URL,
				RepositoryID: "test-repo",
				Username:     "test-user",
				Password:     "test-pass",
				MaxRetries:   1,
				RetryBackoff: 10 * time.Millisecond,
			}
			service := NewQuadMemoryService(config)

			ctx := context.Background()
			result, err := service.AddQuad(ctx, tt.hierarchicalCtx, tt.quad)

			if (err != nil) != tt.expectError {
				t.Errorf("AddQuad() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if result == nil {
					t.Error("期望返回非空的四元组结果")
					return
				}
				if result.Subject != tt.quad.Subject {
					t.Errorf("期望主语 %s，得到 %s", tt.quad.Subject, result.Subject)
				}
				if result.Predicate != tt.quad.Predicate {
					t.Errorf("期望谓语 %s，得到 %s", tt.quad.Predicate, result.Predicate)
				}
				if result.Object != tt.quad.Object {
					t.Errorf("期望宾语 %s，得到 %s", tt.quad.Object, result.Object)
				}
				if result.ID == "" {
					t.Error("期望生成非空的四元组ID")
				}
				if result.Context == "" {
					t.Error("期望生成非空的上下文")
				}
			}
		})
	}
}
func TestSearchQuads(t *testing.T) {
	t.Parallel()

	// 测试用例
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		query        QuadSearchQuery
		expectError  bool
		expectedLen  int
	}{
		{
			name: "成功搜索四元组",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("期望 POST 请求，得到 %s", r.Method)
				}
				if r.Header.Get("Accept") != "application/sparql-results+json" {
					t.Errorf("期望 Accept: application/sparql-results+json")
				}
				// 模拟 SPARQL JSON 响应
				response := `{
					"head": {
						"vars": ["s", "p", "o", "g"]
					},
					"results": {
						"bindings": [
							{
								"s": {"type": "uri", "value": "http://example.com/subject1"},
								"p": {"type": "uri", "value": "http://example.com/predicate1"},
								"o": {"type": "uri", "value": "http://example.com/object1"},
								"g": {"type": "uri", "value": "urn:tenant:test-tenant:story:test-story"}
							},
							{
								"s": {"type": "uri", "value": "http://example.com/subject2"},
								"p": {"type": "uri", "value": "http://example.com/predicate2"},
								"o": {"type": "uri", "value": "http://example.com/object2"},
								"g": {"type": "uri", "value": "urn:tenant:test-tenant:story:test-story"}
							}
						]
					}
				}`
				w.Header().Set("Content-Type", "application/sparql-results+json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			},
			query: QuadSearchQuery{
				Context: &HierarchicalContext{
					TenantID: "test-tenant",
					StoryID:  "test-story",
				},
				Scope: "exact",
			},
			expectError: false,
			expectedLen: 2,
		},
		{
			name: "无结果搜索",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// 模拟空结果 SPARQL JSON 响应
				response := `{
					"head": {
						"vars": ["s", "p", "o", "g"]
					},
					"results": {
						"bindings": []
					}
				}`
				w.Header().Set("Content-Type", "application/sparql-results+json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			},
			query: QuadSearchQuery{
				Context: &HierarchicalContext{
					TenantID: "test-tenant",
					StoryID:  "nonexistent-story",
				},
				Scope: "exact",
			},
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "无效的层次化上下文",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			query: QuadSearchQuery{
				Context: nil, // 缺少上下文
				Scope:   "exact",
			},
			expectError: true,
			expectedLen: 0,
		},
		{
			name: "服务器错误 (500)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			query: QuadSearchQuery{
				Context: &HierarchicalContext{
					TenantID: "test-tenant",
				},
				Scope: "tenant",
			},
			expectError: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			config := QuadMemoryConfig{
				BaseURL:      server.URL,
				RepositoryID: "test-repo",
				Username:     "test-user",
				Password:     "test-pass",
				MaxRetries:   1,
				RetryBackoff: 10 * time.Millisecond,
			}
			service := NewQuadMemoryService(config)

			ctx := context.Background()
			results, err := service.SearchQuads(ctx, tt.query)

			if (err != nil) != tt.expectError {
				t.Errorf("SearchQuads() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if len(results) != tt.expectedLen {
					t.Errorf("期望结果数量 %d，得到 %d", tt.expectedLen, len(results))
				}

				// 验证结果的基本结构
				for i, result := range results {
					if result == nil {
						t.Errorf("结果[%d] 不应为空", i)
						continue
					}
					if result.Subject == "" {
						t.Errorf("结果[%d] 主语不应为空", i)
					}
					if result.Predicate == "" {
						t.Errorf("结果[%d] 谓语不应为空", i)
					}
					if result.Object == "" {
						t.Errorf("结果[%d] 宾语不应为空", i)
					}
					if result.Context == "" {
						t.Errorf("结果[%d] 上下文不应为空", i)
					}
				}
			}
		})
	}
}
