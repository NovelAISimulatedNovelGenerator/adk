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

// Package vector_rag_tool implements standardized tools for writing to and
// reading from a vector RAG backend. The implementation re-uses
// CustomRagMemoryService so that agents can directly interact with the
// FastAPI + Milvus service via structured tool calls.
//
// Two tools are exported:
//   - RAGWrite  – ingests text content as a session in the RAG service
//   - RAGSearch – retrieves similar memories from the RAG service
//
// Both tools follow the generic Tool interface defined in pkg/tools/tool.go so
// that they can be easily registered to any agent.
package vector_rag_tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/nvcnvn/adk-golang/pkg/events"
	"github.com/nvcnvn/adk-golang/pkg/logger"
	"github.com/nvcnvn/adk-golang/pkg/memory"
	"github.com/nvcnvn/adk-golang/pkg/models"
	"github.com/nvcnvn/adk-golang/pkg/sessions"
	"github.com/nvcnvn/adk-golang/pkg/tools"
)

//-------------------------------------------------------------------------
// Shared memory service instance
//-------------------------------------------------------------------------

// ragMemory is a package-level CustomRagMemoryService. The default constructor
// uses environment variables RAG_BASE_URL and RAG_TOP_K if present (handled in
// NewCustomRagMemoryServiceWithDefaults).
var ragMemory = memory.NewCustomRagMemoryServiceWithDefaults()

// WithMemory allows overriding the default memory service (e.g. in tests).
func WithMemory(mem memory.MemoryService) {
	if c, ok := mem.(*memory.CustomRagMemoryService); ok && c != nil {
		ragMemory = c
	}
}

//-------------------------------------------------------------------------
// RAGWrite tool factories – write / ingest content with pre-configured parameters
//-------------------------------------------------------------------------

// NewRAGWriteTool creates a RAG write tool with pre-configured user_id and archive_id.
// This eliminates the need for AI to provide these parameters at runtime.
func NewRAGWriteTool(userID, archiveID string) tools.Tool {
	return tools.NewTool(
		fmt.Sprintf("rag_write_%s_%s", userID, archiveID),
		fmt.Sprintf("Write plain text into the vector RAG memory service for user '%s' and archive '%s'. "+
			"Only requires 'content' parameter.", userID, archiveID),
		tools.ToolSchema{
			Input: tools.ParameterSchema{
				Type: "object",
				Properties: map[string]tools.ParameterSchema{
					"content": {
						Type:        "string",
						Description: "Text content to write to memory (required)",
						Required:    true,
					},
				},
			},
			Output: map[string]tools.ParameterSchema{
				"result": {
					Type:        "string",
					Description: "Status message (success / error)",
				},
			},
		},
		func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
			// Only validate content parameter - userID and archiveID are pre-configured
			content, ok := input["content"].(string)
			if !ok || strings.TrimSpace(content) == "" {
				return nil, fmt.Errorf("content is required and must be a string")
			}
			author := "user"
			sessionID := archiveID

			// Build a minimal Session structure using pre-configured parameters
			session := &sessions.Session{
				AppName: userID,
				UserID:  userID,
				ID:      sessionID,
				Events: []*events.Event{
					{
						Author: author,
						Content: &models.Content{
							Parts: []*models.Part{{Text: content}},
						},
					},
				},
			}

			logger.S().Infow("RAG写入操作", "user_id", userID, "archive_id", archiveID, "content_length", len(content))
			if err := ragMemory.AddSessionToMemory(ctx, session); err != nil {
				return nil, fmt.Errorf("failed to add session to memory: %w", err)
			}

			return map[string]interface{}{
				"result": fmt.Sprintf("Content ingested into RAG for user=%s, archive=%s", userID, archiveID),
			}, nil
		},
	)
}

// NewRAGSearchTool creates a RAG search tool with pre-configured user_id and archive_id.
func NewRAGSearchTool(userID, archiveID string) tools.Tool {
	return tools.NewTool(
		fmt.Sprintf("rag_search_%s_%s", userID, archiveID),
		fmt.Sprintf("Search the vector RAG memory service for user '%s' and archive '%s'. "+
			"Only requires 'query' parameter.", userID, archiveID),
		tools.ToolSchema{
			Input: tools.ParameterSchema{
				Type: "object",
				Properties: map[string]tools.ParameterSchema{
					"query": {
						Type:        "string",
						Description: "Search query to find relevant content (required)",
						Required:    true,
					},
					"top_k": {
						Type:        "integer",
						Description: "Number of results to return (optional, defaults to backend setting)",
						Required:    false,
					},
				},
			},
			Output: map[string]tools.ParameterSchema{
				"results": {
					Type:        "array",
					Description: "Array of relevant text content sorted by relevance",
				},
			},
		},
		func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
			// Only validate query parameter - userID and archiveID are pre-configured
			query, ok := input["query"].(string)
			if !ok || strings.TrimSpace(query) == "" {
				return nil, fmt.Errorf("query is required and must be a string")
			}

			logger.S().Infow("RAG搜索操作", "user_id", userID, "archive_id", archiveID, "query", query)
			results, err := ragMemory.SearchMemory(ctx, "app_name_placeholder", userID, archiveID, query)
			if err != nil {
				return nil, fmt.Errorf("failed to search memory: %w", err)
			}

			// 如果用户提供 top_k，则在客户端侧做一次截断，避免返回过多内容
			if tkVal, ok := input["top_k"].(float64); ok && int(tkVal) > 0 && int(tkVal) < len(results.Memories) {
				results.Memories = results.Memories[:int(tkVal)]
			}

			return map[string]interface{}{
				"results": results,
			}, nil
		},
	)
}

