package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/logger"
	"go.uber.org/zap"
)

// QuadMemoryConfig 四元组内存服务的配置结构
type QuadMemoryConfig struct {
	BaseURL      string        // GraphDB 基础URL，例如 http://localhost:7200
	RepositoryID string        // GraphDB 仓库ID，例如 "main"
	Username     string        // GraphDB 认证用户名
	Password     string        // GraphDB 认证密码
	MaxRetries   int          // 最大重试次数
	RetryBackoff time.Duration // 重试间隔时间
}

// QuadMemoryService 与远程四元组内存服务通信的客户端
type QuadMemoryService struct {
	config     QuadMemoryConfig
	httpClient *http.Client
	logger     *zap.SugaredLogger
}

// NewQuadMemoryService 从配置创建新的客户端实例
func NewQuadMemoryService(config QuadMemoryConfig) *QuadMemoryService {
	// 设置默认值
	if config.MaxRetries <= 0 {
		config.MaxRetries = 1 // 默认不重试，只尝试一次
	}
	if config.RetryBackoff <= 0 {
		config.RetryBackoff = 500 * time.Millisecond
	}
	if config.RepositoryID == "" {
		config.RepositoryID = "main" // 默认仓库ID
	}

	return &QuadMemoryService{
		config: config,
		httpClient: &http.Client{
			Timeout: 20 * time.Second, // 默认超时时间
		},
		logger: logger.S(),
	}
}

// HealthCheck 验证GraphDB服务和仓库是否可用
// 执行简单查询来检查仓库可访问性
func (s *QuadMemoryService) HealthCheck(ctx context.Context) error {
	// 使用简单的SPARQL查询测试仓库可访问性
	repoURL := s.buildRepositoryURL("")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoURL, nil)
	if err != nil {
		return fmt.Errorf("创建健康检查请求失败: %w", err)
	}

	// 添加SPARQL结果的Accept头部
	req.Header.Set("Accept", "application/sparql-results+json")

	_, err = s.doRequest(req, nil)
	if err != nil {
		return fmt.Errorf("GraphDB仓库 '%s' 不可访问: %w", s.config.RepositoryID, err)
	}
	return nil
}

// doRequest 处理HTTP请求的完整生命周期，包括认证、重试和响应处理
// 这是一个私有辅助方法
func (s *QuadMemoryService) doRequest(req *http.Request, result any) (*http.Response, error) {
	var lastErr error

	// 如果配置了认证信息则添加
	if s.config.Username != "" || s.config.Password != "" {
		req.SetBasicAuth(s.config.Username, s.config.Password)
	}

	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		// 为重试总是克隆请求，因为请求体只能读取一次
		clonedReq := req.Clone(req.Context())
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("读取请求体失败: %w", err)
			}
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			clonedReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err := s.httpClient.Do(clonedReq)
		if err != nil {
			lastErr = fmt.Errorf("HTTP请求失败: %w", err)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if result != nil {
					if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
						return resp, fmt.Errorf("解析成功响应失败: %w", err)
					}
				}
				return resp, nil // 成功
			}

			// 读取响应体获取错误消息
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API返回状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))

			// 4xx客户端错误不重试
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				break
			}
		}

		// 如果不是最后一次尝试，等待并重试
		if attempt < s.config.MaxRetries {
			wait := s.config.RetryBackoff * time.Duration(attempt) // 线性退避
			s.logger.Warnw("请求失败，正在重试...", "attempt", attempt, "max_retries", s.config.MaxRetries, "wait", wait, "error", lastErr)
			select {
			case <-time.After(wait):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}
	}

	return nil, lastErr
}


// HierarchicalContext 表示用于逻辑分区的层次化上下文
// 支持多层次数据组织，如 租户 -> 故事 -> 章节/角色
type HierarchicalContext struct {
	TenantID    string `json:"tenant_id"`              // 必需：租户标识符
	StoryID     string `json:"story_id,omitempty"`     // 可选：故事标识符
	ChapterID   string `json:"chapter_id,omitempty"`   // 可选：章节标识符
	CharacterID string `json:"character_id,omitempty"` // 可选：角色标识符
	// 可根据需要添加更多层次
}

// Quad 表示一个主语-谓语-宾语-上下文的四元组语句
type Quad struct {
	ID        string `json:"id,omitempty"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Context   string `json:"context,omitempty"` // 将从 HierarchicalContext 自动生成
}

// AddQuadRequest 添加四元组的请求载荷
type AddQuadRequest struct {
	TenantID string `json:"tenant_id"`
	Quad     Quad   `json:"quad"`
}

// QuadSearchQuery 定义支持层次化上下文的搜索条件
type QuadSearchQuery struct {
	Subject   string                `json:"subject,omitempty"`   // 可选：按主语过滤
	Predicate string                `json:"predicate,omitempty"` // 可选：按谓语过滤
	Object    string                `json:"object,omitempty"`    // 可选：按宾语过滤
	Context   *HierarchicalContext  `json:"context,omitempty"`   // 可选：用于过滤的层次化上下文
	Scope     string                `json:"scope,omitempty"`     // 可选：查询范围 ("exact", "story", "tenant")
}

// SearchQuadsRequest 搜索四元组的请求载荷
type SearchQuadsRequest struct {
	TenantID string          `json:"tenant_id"`
	Query    QuadSearchQuery `json:"query"`
}

// buildGraphURI 从给定的上下文构建层次化命名图 URI
// 示例:
//   - urn:tenant:user123
//   - urn:tenant:user123:story:story1
//   - urn:tenant:user123:story:story1:chapter:ch1
//   - urn:tenant:user123:story:story1:character:alice
func (s *QuadMemoryService) buildGraphURI(ctx *HierarchicalContext) string {
	if ctx == nil || ctx.TenantID == "" {
		return ""
	}

	uri := fmt.Sprintf("urn:tenant:%s", ctx.TenantID)

	if ctx.StoryID != "" {
		uri += fmt.Sprintf(":story:%s", ctx.StoryID)

		if ctx.ChapterID != "" {
			uri += fmt.Sprintf(":chapter:%s", ctx.ChapterID)
		} else if ctx.CharacterID != "" {
			uri += fmt.Sprintf(":character:%s", ctx.CharacterID)
		}
	}

	return uri
}

// buildRepositoryURL 为给定的端点构建 GraphDB 仓库 URL
// 示例:
//   - http://localhost:7200/repositories/main (用于查询)
//   - http://localhost:7200/repositories/main/statements (用于更新)
func (s *QuadMemoryService) buildRepositoryURL(endpoint string) string {
	baseURL := strings.TrimSuffix(s.config.BaseURL, "/")
	repoURL := fmt.Sprintf("%s/repositories/%s", baseURL, s.config.RepositoryID)
	if endpoint != "" {
		repoURL += "/" + endpoint
	}
	return repoURL
}

// validateHierarchicalContext 验证层次化上下文
func (s *QuadMemoryService) validateHierarchicalContext(ctx *HierarchicalContext) error {
	if ctx == nil {
		return fmt.Errorf("层次化上下文不能为空")
	}
	if ctx.TenantID == "" {
		return fmt.Errorf("层次化上下文中需要 tenant_id")
	}
	return nil
}

// AddQuad 向内存服务添加新的四元组，支持层次化上下文
// 此方法使用 SPARQL INSERT DATA 将四元组添加到适当的命名图中
func (s *QuadMemoryService) AddQuad(ctx context.Context, hierarchicalCtx *HierarchicalContext, quad Quad) (*Quad, error) {
	// 验证输入
	if err := s.validateHierarchicalContext(hierarchicalCtx); err != nil {
		return nil, fmt.Errorf("无效的层次化上下文: %w", err)
	}
	if quad.Subject == "" || quad.Predicate == "" || quad.Object == "" {
		return nil, fmt.Errorf("主语、谓语和宾语都是必需的")
	}

	// 构建命名图 URI
	graphURI := s.buildGraphURI(hierarchicalCtx)
	if graphURI == "" {
		return nil, fmt.Errorf("从上下文构建图 URI 失败")
	}

	// 如果未提供 ID 则生成
	resultQuad := quad
	if resultQuad.ID == "" {
		resultQuad.ID = fmt.Sprintf("%s-%s-%s-%d", 
			hierarchicalCtx.TenantID, 
			strings.ReplaceAll(quad.Subject, ":", "-"),
			strings.ReplaceAll(quad.Predicate, ":", "-"),
			time.Now().UnixNano())
	}
	resultQuad.Context = graphURI

	// 构建 SPARQL INSERT DATA 语句
	sparqlUpdate := fmt.Sprintf(`INSERT DATA { 
		GRAPH <%s> { 
			<%s> <%s> <%s> 
		} 
	}`, graphURI, quad.Subject, quad.Predicate, quad.Object)

	// 执行 SPARQL 更新
	if err := s.executeSPARQLUpdate(ctx, sparqlUpdate); err != nil {
		return nil, fmt.Errorf("添加四元组失败: %w", err)
	}

	s.logger.Infow("成功添加四元组", 
		"graph_uri", graphURI,
		"subject", quad.Subject,
		"predicate", quad.Predicate,
		"object", quad.Object)

	return &resultQuad, nil
}

// executeSPARQLUpdate 对 GraphDB 仓库执行 SPARQL UPDATE 语句
func (s *QuadMemoryService) executeSPARQLUpdate(ctx context.Context, sparqlUpdate string) error {
	// 构建更新端点 URL
	updateURL := s.buildRepositoryURL("statements")

	// 为 SPARQL 更新准备表单数据
	formData := fmt.Sprintf("update=%s", sparqlUpdate)
	reqBody := strings.NewReader(formData)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, updateURL, reqBody)
	if err != nil {
		return fmt.Errorf("创建 SPARQL 更新请求失败: %w", err)
	}

	// 设置 SPARQL 更新所需的头部
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/plain")

	// 执行请求
	resp, err := s.doRequest(req, nil)
	if err != nil {
		return fmt.Errorf("SPARQL 更新请求失败: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Debugw("SPARQL 更新执行成功", 
		"status_code", resp.StatusCode,
		"update_url", updateURL)

	return nil
}

// SearchQuads 搜索支持层次化上下文和范围的四元组
// 此方法使用 SPARQL SELECT 从适当的命名图中查询四元组
func (s *QuadMemoryService) SearchQuads(ctx context.Context, query QuadSearchQuery) ([]*Quad, error) {
	// 验证输入
	if query.Context == nil {
		return nil, fmt.Errorf("需要层次化上下文")
	}
	if err := s.validateHierarchicalContext(query.Context); err != nil {
		return nil, fmt.Errorf("无效的层次化上下文: %w", err)
	}
	if query.Scope == "" {
		query.Scope = "exact" // 默认为精确范围
	}

	// 根据范围构建 SPARQL 查询
	sparqlQuery, err := s.buildSPARQLQuery(query)
	if err != nil {
		return nil, fmt.Errorf("构建 SPARQL 查询失败: %w", err)
	}

	// 执行 SPARQL 查询
	results, err := s.executeSPARQLQuery(ctx, sparqlQuery)
	if err != nil {
		return nil, fmt.Errorf("执行搜索查询失败: %w", err)
	}

	s.logger.Infow("成功执行搜索查询", 
		"scope", query.Scope,
		"results_count", len(results))

	return results, nil
}

// buildSPARQLQuery 根据搜索条件和范围构建 SPARQL SELECT 查询
func (s *QuadMemoryService) buildSPARQLQuery(query QuadSearchQuery) (string, error) {
	var graphPattern string
	var whereClause string

	// 根据范围构建图模式
	switch query.Scope {
	case "exact":
		// 指定层次级别的精确匹配
		graphURI := s.buildGraphURI(query.Context)
		if graphURI == "" {
			return "", fmt.Errorf("为精确范围构建图 URI 失败")
		}
		graphPattern = fmt.Sprintf("GRAPH <%s>", graphURI)

	case "story":
		// 在整个故事内搜索（包括章节、角色等）
		if query.Context.StoryID == "" {
			return "", fmt.Errorf("故事范围需要 story_id")
		}
		storyPrefix := fmt.Sprintf("urn:tenant:%s:story:%s", query.Context.TenantID, query.Context.StoryID)
		graphPattern = fmt.Sprintf("GRAPH ?g FILTER(STRSTARTS(STR(?g), \"%s\"))", storyPrefix)

	case "tenant":
		// 在整个租户内搜索
		tenantPrefix := fmt.Sprintf("urn:tenant:%s", query.Context.TenantID)
		graphPattern = fmt.Sprintf("GRAPH ?g FILTER(STRSTARTS(STR(?g), \"%s\"))", tenantPrefix)

	default:
		return "", fmt.Errorf("无效范围: %s（必须是 'exact'、'story' 或 'tenant'）", query.Scope)
	}

	// 构建带有可选过滤器的 WHERE 子句
	subject := "?s"
	if query.Subject != "" {
		subject = fmt.Sprintf("<%s>", query.Subject)
	}

	predicate := "?p"
	if query.Predicate != "" {
		predicate = fmt.Sprintf("<%s>", query.Predicate)
	}

	object := "?o"
	if query.Object != "" {
		object = fmt.Sprintf("<%s>", query.Object)
	}

	whereClause = fmt.Sprintf("%s %s %s", subject, predicate, object)

	// 构建完整的 SPARQL 查询
	sparqlQuery := fmt.Sprintf(`
		SELECT ?s ?p ?o ?g WHERE { 
			%s { 
				%s . 
			} 
		}
	`, graphPattern, whereClause)

	return sparqlQuery, nil
}

// SPARQLResult 表示 SPARQL 查询的单个结果
type SPARQLResult struct {
	S struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"s"`
	P struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"p"`
	O struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"o"`
	G struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"g"`
}

// SPARQLResponse 表示 SPARQL SELECT 查询的响应
type SPARQLResponse struct {
	Head struct {
		Vars []string `json:"vars"`
	} `json:"head"`
	Results struct {
		Bindings []SPARQLResult `json:"bindings"`
	} `json:"results"`
}

// executeSPARQLQuery 对 GraphDB 仓库执行 SPARQL SELECT 查询
func (s *QuadMemoryService) executeSPARQLQuery(ctx context.Context, sparqlQuery string) ([]*Quad, error) {
	// 构建查询端点 URL
	queryURL := s.buildRepositoryURL("")

	// 为 SPARQL 查询准备表单数据
	formData := fmt.Sprintf("query=%s", sparqlQuery)
	reqBody := strings.NewReader(formData)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建 SPARQL 查询请求失败: %w", err)
	}

	// 设置 SPARQL 查询所需的头部
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/sparql-results+json")

	// 执行请求
	var sparqlResponse SPARQLResponse
	resp, err := s.doRequest(req, &sparqlResponse)
	if err != nil {
		return nil, fmt.Errorf("SPARQL 查询请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 将 SPARQL 结果转换为 Quad 对象
	quads := make([]*Quad, 0, len(sparqlResponse.Results.Bindings))
	for i, binding := range sparqlResponse.Results.Bindings {
		quad := &Quad{
			ID:        fmt.Sprintf("result-%d-%d", time.Now().UnixNano(), i),
			Subject:   binding.S.Value,
			Predicate: binding.P.Value,
			Object:    binding.O.Value,
			Context:   binding.G.Value,
		}
		quads = append(quads, quad)
	}

	s.logger.Debugw("SPARQL 查询执行成功", 
		"status_code", resp.StatusCode,
		"query_url", queryURL,
		"results_count", len(quads))

	return quads, nil
}
