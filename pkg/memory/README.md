# Memory 记忆服务模块

## 概述

Memory 模块提供了智能体对话记忆管理功能，支持会话存储、检索和语义搜索。该模块是实现上下文感知智能体的核心组件，使智能体能够记住和利用历史对话信息。

## 核心接口

### MemoryService (记忆服务接口)
```go
type MemoryService interface {
    // 将会话添加到记忆服务
    AddSessionToMemory(ctx context.Context, session *sessions.Session) error
    
    // 搜索匹配查询的会话记忆
    SearchMemory(ctx context.Context, appName, userID, query string) (*SearchMemoryResponse, error)
}
```

记忆服务的核心接口，定义了会话存储和记忆检索的基本功能。

### MemoryResult (记忆结果)
```go
type MemoryResult struct {
    SessionID string         `json:"sessionId"` // 关联的会话ID
    Events    []*events.Event `json:"events"`   // 会话中的事件列表
}
```

单个记忆检索结果，包含会话ID和相关事件。

### SearchMemoryResponse (搜索响应)
```go
type SearchMemoryResponse struct {
    Memories []*MemoryResult `json:"memories"` // 匹配搜索查询的记忆结果列表
}
```

记忆搜索操作的响应结构体。

## 实现类型

### 1. InMemoryMemoryService (内存记忆服务)
基于内存的记忆服务实现，主要用于原型开发和测试：

```go
type InMemoryMemoryService struct {
    sessionEvents map[string][]*events.Event // 会话事件映射
    mu            sync.RWMutex               // 读写锁保护并发安全
}
```

**特性:**
- 使用关键词匹配而非语义搜索
- 线程安全的并发访问
- 适用于原型开发和测试环境
- 数据存储在内存中，重启后丢失

**创建实例:**
```go
func NewInMemoryMemoryService() *InMemoryMemoryService
```

### 2. CustomRagMemoryService (自定义RAG记忆服务)
基于RAG（检索增强生成）的记忆服务实现，支持向量化存储和语义搜索。

### 3. VertexAI 记忆服务
基于 Google Vertex AI 的记忆服务实现，利用云端AI服务进行智能记忆管理。

## 使用示例

### 基本使用流程
```go
package main

import (
    "context"
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/memory"
    "github.com/nvcnvn/adk-golang/pkg/sessions"
    "github.com/nvcnvn/adk-golang/pkg/events"
)

func main() {
    ctx := context.Background()
    
    // 创建内存记忆服务
    memoryService := memory.NewInMemoryMemoryService()
    
    // 创建会话和事件
    session := &sessions.Session{
        ID:     "session-123",
        AppName: "my-app",
        UserID:  "user-456",
        Events: []*events.Event{
            {
                Type: "user_message",
                Data: "你好，我想了解Go语言",
            },
            {
                Type: "assistant_response", 
                Data: "Go是Google开发的编程语言，具有高性能和简洁语法",
            },
        },
    }
    
    // 添加会话到记忆
    err := memoryService.AddSessionToMemory(ctx, session)
    if err != nil {
        fmt.Printf("添加会话到记忆失败: %v\n", err)
        return
    }
    
    // 搜索相关记忆
    response, err := memoryService.SearchMemory(ctx, "my-app", "user-456", "Go语言")
    if err != nil {
        fmt.Printf("搜索记忆失败: %v\n", err)
        return
    }
    
    // 处理搜索结果
    fmt.Printf("找到 %d 条相关记忆\n", len(response.Memories))
    for i, memory := range response.Memories {
        fmt.Printf("记忆 %d (会话ID: %s):\n", i+1, memory.SessionID)
        for j, event := range memory.Events {
            fmt.Printf("  事件 %d: %s\n", j+1, event.Data)
        }
    }
}
```

### 在智能体中集成记忆服务
```go
// 智能体处理前回调 - 检索历史记忆
func (agent *MyAgent) BeforeProcess(ctx context.Context, input string) (string, error) {
    userID := ctx.Value("user_id").(string)
    
    // 搜索相关历史记忆
    response, err := agent.memoryService.SearchMemory(ctx, "my-app", userID, input)
    if err != nil {
        return input, err // 记忆检索失败不影响主流程
    }
    
    // 将历史记忆拼接到当前输入
    var contextualInput strings.Builder
    contextualInput.WriteString("历史对话上下文:\n")
    
    for _, memory := range response.Memories {
        for _, event := range memory.Events {
            contextualInput.WriteString(fmt.Sprintf("- %s\n", event.Data))
        }
    }
    
    contextualInput.WriteString("\n当前问题: ")
    contextualInput.WriteString(input)
    
    return contextualInput.String(), nil
}

// 智能体处理后回调 - 保存当前会话
func (agent *MyAgent) AfterProcess(ctx context.Context, input, output string) error {
    userID := ctx.Value("user_id").(string)
    sessionID := generateSessionID() // 生成会话ID
    
    session := &sessions.Session{
        ID:      sessionID,
        AppName: "my-app", 
        UserID:  userID,
        Events: []*events.Event{
            {Type: "user_message", Data: input},
            {Type: "assistant_response", Data: output},
        },
    }
    
    // 保存当前会话到记忆
    return agent.memoryService.AddSessionToMemory(ctx, session)
}
```

## 数据结构

### 会话键格式
内存实现使用以下格式存储会话：
```
格式: {app_name}/{user_id}/{session_id}
示例: "my-app/user123/session-456"
```

### 搜索机制
- **关键词匹配**: InMemoryMemoryService 使用简单的关键词匹配
- **语义搜索**: 高级实现支持基于向量的语义搜索
- **上下文感知**: 根据用户和应用维度进行搜索

## API 集成支持

该模块与 API 层深度集成，支持以下工作流：

1. **HTTP请求**: 包含 `user_id` 字段
2. **任务调度**: WorkflowService 透传用户ID到 `scheduler.Task.UserID`
3. **上下文注入**: Worker 回调中通过 `context.WithValue` 注入用户ID
4. **记忆操作**: Agent 在 before/after 回调中执行记忆存储和检索

## 配置和扩展

### 自定义记忆服务实现
```go
type MyCustomMemoryService struct {
    // 自定义字段
}

func (s *MyCustomMemoryService) AddSessionToMemory(ctx context.Context, session *sessions.Session) error {
    // 自定义实现
    return nil
}

func (s *MyCustomMemoryService) SearchMemory(ctx context.Context, appName, userID, query string) (*SearchMemoryResponse, error) {
    // 自定义搜索逻辑
    return &SearchMemoryResponse{}, nil
}
```

### 记忆服务选择策略
```go
func CreateMemoryService(config *config.Config) memory.MemoryService {
    switch config.MemoryType {
    case "inmemory":
        return memory.NewInMemoryMemoryService()
    case "rag":
        return memory.NewCustomRagMemoryService(config.RAGConfig)
    case "vertexai":
        return memory.NewVertexAIMemoryService(config.VertexAIConfig)
    default:
        return memory.NewInMemoryMemoryService() // 默认实现
    }
}
```

## 最佳实践

1. **会话管理**: 合理设置会话边界，避免单个会话过长
2. **存储优化**: 定期清理过期或不重要的记忆数据
3. **搜索优化**: 使用合适的查询关键词提高搜索准确性
4. **并发安全**: 多线程环境下注意记忆服务的线程安全性
5. **错误处理**: 记忆操作失败不应影响主业务流程
6. **隐私保护**: 合理管理用户数据，遵循隐私保护原则

## 性能考虑

- **内存使用**: InMemoryMemoryService 需要考虑内存占用
- **检索速度**: 大量数据时考虑使用索引优化检索性能
- **存储持久化**: 生产环境建议使用持久化存储方案
- **缓存策略**: 频繁访问的记忆可以考虑缓存机制

## 依赖

- `github.com/nvcnvn/adk-golang/pkg/events`: 事件类型定义
- `github.com/nvcnvn/adk-golang/pkg/sessions`: 会话管理
- Go 标准库: `context`, `strings`, `sync`

## 扩展接口

模块设计支持灵活扩展，可以轻松添加新的记忆服务实现：
- 数据库持久化记忆服务
- 向量数据库记忆服务  
- 分布式记忆服务
- 云端记忆服务

该模块是构建智能对话系统的重要基础设施，为智能体提供了强大的上下文感知能力。
