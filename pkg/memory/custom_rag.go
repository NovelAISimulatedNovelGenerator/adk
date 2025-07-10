// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package memory provides interfaces and implementations for memory services.
package memory

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/nvcnvn/adk-golang/pkg/events"
    "github.com/nvcnvn/adk-golang/pkg/sessions"
    "github.com/nvcnvn/adk-golang/pkg/models"
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

// CustomRagMemoryService connects to a user-provided RAG service (FastAPI + Milvus).
// The service exposes two endpoints:
//   POST /add_session      → ingest a session
//   POST /search_memory    → retrieve related memories
//
// This implementation translates the MemoryService interface to those HTTP calls.
type CustomRagMemoryService struct {
    // BaseURL is the base URL of the FastAPI service, e.g. http://localhost:8000
    BaseURL string

    // SimilarityTopK is the default top-k for SearchMemory if not overridden
    SimilarityTopK int

    // HTTP client reused across requests
    httpClient *http.Client

    mu sync.RWMutex
}

// NewCustomRagMemoryService creates a new service instance.
func NewCustomRagMemoryService(baseURL string, similarityTopK int) *CustomRagMemoryService {
    if similarityTopK <= 0 {
        similarityTopK = 10
    }
    return &CustomRagMemoryService{
        BaseURL:        strings.TrimRight(baseURL, "/"),
        SimilarityTopK: similarityTopK,
        httpClient: &http.Client{
            Timeout: 20 * time.Second,
        },
    }
}

// -----------------------------------------------------------------------------
// request / response payloads – mirror FastAPI definitions
// -----------------------------------------------------------------------------

type ragMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type addSessionRequest struct {
    TenantID  string                 `json:"tenant_id"`
    SessionID string                 `json:"session_id"`
    Messages  []ragMessage           `json:"messages"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type searchMemoryRequest struct {
    TenantID string                 `json:"tenant_id"`
    Query    string                 `json:"query"`
    TopK     int                    `json:"top_k"`
    Filter   map[string]interface{} `json:"filter,omitempty"`
}

type searchResult struct {
    Content string  `json:"content"`
    Score   float64 `json:"score"`
}

type searchMemoryResponse struct {
    Results   []searchResult `json:"results"`
    ElapsedMS float64        `json:"elapsed_ms"`
}

// -----------------------------------------------------------------------------
// MemoryService implementation
// -----------------------------------------------------------------------------

// AddSessionToMemory uploads the session to the external RAG service.
// tenant_id == session.AppName (可根据需要调整)
func (c *CustomRagMemoryService) AddSessionToMemory(ctx context.Context, session *sessions.Session) error {
    c.mu.RLock()
    baseURL := c.BaseURL
    client := c.httpClient
    c.mu.RUnlock()

    reqPayload := addSessionRequest{
        TenantID:  session.AppName,
        SessionID: session.ID,
    }

    for _, ev := range session.Events {
        if ev.Content == nil || len(ev.Content.Parts) == 0 {
            continue
        }
        var sb strings.Builder
        for _, part := range ev.Content.Parts {
            if part.Text != "" {
                if sb.Len() > 0 {
                    sb.WriteString("\n")
                }
                sb.WriteString(part.Text)
            }
        }
        if sb.Len() == 0 {
            continue
        }
        reqPayload.Messages = append(reqPayload.Messages, ragMessage{
            Role:    ev.Author,
            Content: sb.String(),
        })
    }

    if len(reqPayload.Messages) == 0 {
        logger.S().Infow("CustomRAG: no events with content to add", "session", session.ID)
        return nil
    }

    body, _ := json.Marshal(reqPayload)

    // ------------------------------------------------------------------
    // FastAPI + Milvus 在首次写入新 tenant 时，如果对应集合尚未创建，
    // 会先抛出 500（Milvus CollectionNotExists），随后自动去创建集合。
    // 为了让客户端“写一次就成功”，这里增加一次簡單的重试逻輯：
    //   1. 仅当网络错误或 HTTP >= 500 时才重试；
    //   2. 最多 3 次，线性退避 0.5s、1s；
    //   3. 若遇到 4xx（参数错误、鉴权失败等）立即返回，不做重试。
    // 这样即可避免第一次 500 导致工作流失败，同时不会给正常错误造成无限重试。
    // ------------------------------------------------------------------
    const maxRetries = 3
    var lastErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        // 每次循环都重新创建 *http.Request，避免 body 在前一次已被读取。
        httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/add_session", bytes.NewReader(body))
        if err != nil {
            return fmt.Errorf("create request failed: %w", err)
        }
        httpReq.Header.Set("Content-Type", "application/json")

        resp, err := client.Do(httpReq)
        if err != nil {
            lastErr = fmt.Errorf("http request error: %w", err)
        } else {
            defer resp.Body.Close()
            if resp.StatusCode < 300 {
                return nil // 成功写入
            }

            respBody, _ := io.ReadAll(resp.Body)
            lastErr = fmt.Errorf("rag service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))

            // 遇到 4xx 直接返回，不做重试
            if resp.StatusCode < 500 {
                break
            }
        }

        // 若未达到最大次数，则等待后重试
        if attempt < maxRetries {
            wait := time.Duration(500*attempt) * time.Millisecond
            logger.S().Warnw("CustomRAG add_session failed, will retry", "attempt", attempt, "max", maxRetries, "err", lastErr, "wait", wait)
            select {
            case <-time.After(wait):
            case <-ctx.Done():
                return ctx.Err()
            }
            continue
        }
    }

    return lastErr
}

// SearchMemory queries the external RAG service for related contents.
func (c *CustomRagMemoryService) SearchMemory(ctx context.Context, appName, userID, query string) (*SearchMemoryResponse, error) {
    c.mu.RLock()
    baseURL := c.BaseURL
    client := c.httpClient
    topK := c.SimilarityTopK
    c.mu.RUnlock()

    // 使用 appName 作为 tenant_id
    reqPayload := searchMemoryRequest{
        TenantID: appName,
        Query:    query,
        TopK:     topK,
    }

    body, _ := json.Marshal(reqPayload)
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/search_memory", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request failed: %w", err)
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("http request error: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        respBody, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("rag service returned %d: %s", resp.StatusCode, string(respBody))
    }

    var ragResp searchMemoryResponse
    if err := json.NewDecoder(resp.Body).Decode(&ragResp); err != nil {
        return nil, fmt.Errorf("decode response failed: %w", err)
    }

    // Convert to MemoryResult list
    memResp := &SearchMemoryResponse{Memories: []*MemoryResult{}}
    sessionID := fmt.Sprintf("rag_%d", time.Now().UnixNano()) // synthetic session id

    for _, r := range ragResp.Results {
        ev := &events.Event{
            Author: "memory",
            Content: &models.Content{
                Parts: []*models.Part{{Text: r.Content}},
            },
        }
        memResp.Memories = append(memResp.Memories, &MemoryResult{
            SessionID: sessionID,
            Events:    []*events.Event{ev},
        })
    }
    return memResp, nil
}
