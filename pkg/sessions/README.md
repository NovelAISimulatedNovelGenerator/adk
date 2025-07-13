# Sessions 会话管理模块

## 概述

Sessions 模块提供了完整的会话管理功能，支持智能体对话的会话创建、状态管理、事件存储和持久化。该模块是构建有状态对话系统的核心基础设施，为多轮对话和上下文保持提供强大支持。

## 核心组件

### Session (会话)
```go
type Session struct {
    AppName    string                   `json:"appName"`    // 应用名称
    UserID     string                   `json:"userId"`     // 用户ID
    ID         string                   `json:"id"`         // 会话唯一标识
    Events     []*events.Event          `json:"events"`     // 会话事件列表
    State      *State                   `json:"-"`          // 会话状态（不直接序列化）
    StateMap   map[string]interface{}   `json:"state"`      // 状态的可序列化表示
    CreateTime time.Time               `json:"createTime"` // 创建时间
    UpdateTime time.Time               `json:"updateTime"` // 更新时间
}
```

会话是对话的基本单位，包含了完整的对话历史、状态信息和元数据。

### SessionService (会话服务接口)
```go
type SessionService interface {
    // 创建新会话
    CreateSession(ctx context.Context, appName, userID string, state map[string]interface{}, sessionID string) (*Session, error)
    
    // 根据ID获取会话
    GetSession(ctx context.Context, appName, userID, sessionID string, config *GetSessionConfig) (*Session, error)
    
    // 列出用户的所有会话
    ListSessions(ctx context.Context, appName, userID string) (*ListSessionsResponse, error)
    
    // 更新会话
    UpdateSession(ctx context.Context, session *Session) error
    
    // 删除会话
    DeleteSession(ctx context.Context, appName, userID, sessionID string) error
    
    // 列出会话中的事件
    ListEvents(ctx context.Context, appName, userID, sessionID, pageToken string, maxEvents int) (*ListEventsResponse, error)
    
    // 添加事件到会话
    AddEvent(ctx context.Context, appName, userID, sessionID string, event *events.Event) error
}
```

### GetSessionConfig (获取会话配置)
```go
type GetSessionConfig struct {
    NumRecentEvents int     // 限制返回的事件数量
    AfterTimestamp  float64 // 过滤指定时间戳之后的事件
}
```

用于配置会话检索的选项。

### State (会话状态)
```go
type State struct {
    data map[string]interface{}
    mu   sync.RWMutex
}
```

线程安全的会话状态管理器。

## 实现类型

### 1. InMemorySessionService (内存会话服务)
基于内存的会话服务实现，适用于开发和测试：
- 数据存储在内存中
- 进程重启后数据丢失
- 高性能，无I/O开销
- 支持并发安全访问

### 2. DatabaseSessionService (数据库会话服务)
基于数据库的持久化会话服务：
- 数据持久化存储
- 支持多实例部署
- 事务安全保证
- 支持复杂查询和统计

### 3. VertexAISessionService (Vertex AI 会话服务)
基于 Google Vertex AI 的会话服务实现：
- 利用云端AI服务
- 智能会话分析
- 自动状态推断
- 企业级可靠性

## 使用示例

### 基本会话管理
```go
package main

import (
    "context"
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/sessions"
    "github.com/nvcnvn/adk-golang/pkg/events"
)

func main() {
    ctx := context.Background()
    
    // 创建内存会话服务
    sessionService := sessions.NewInMemorySessionService()
    
    // 创建新会话
    initialState := map[string]interface{}{
        "language": "zh-CN",
        "topic":    "Go编程",
    }
    
    session, err := sessionService.CreateSession(ctx, "my-app", "user123", initialState, "")
    if err != nil {
        fmt.Printf("创建会话失败: %v\n", err)
        return
    }
    
    fmt.Printf("创建会话成功: %s\n", session.ID)
    
    // 添加事件到会话
    userEvent := &events.Event{
        Type: "user_message",
        Data: "你好，我想学习Go语言",
    }
    
    err = sessionService.AddEvent(ctx, "my-app", "user123", session.ID, userEvent)
    if err != nil {
        fmt.Printf("添加事件失败: %v\n", err)
        return
    }
    
    // 获取会话
    config := &sessions.GetSessionConfig{
        NumRecentEvents: 10,
    }
    
    retrievedSession, err := sessionService.GetSession(ctx, "my-app", "user123", session.ID, config)
    if err != nil {
        fmt.Printf("获取会话失败: %v\n", err)
        return
    }
    
    fmt.Printf("会话包含 %d 个事件\n", len(retrievedSession.Events))
}
```

### 会话状态管理
```go
func manageSessionState() {
    // 创建会话
    session := sessions.NewSession("my-app", "user123", nil, "")
    
    // 设置状态
    session.SetState("current_step", "greeting")
    session.SetState("user_preferences", map[string]interface{}{
        "language": "zh-CN",
        "difficulty": "beginner",
    })
    
    // 获取状态
    if step, exists := session.GetState("current_step"); exists {
        fmt.Printf("当前步骤: %s\n", step)
    }
    
    // 添加事件会自动更新状态
    event := &events.Event{
        Type: "system_update",
        Data: "用户偏好已更新",
        Actions: &events.EventActions{
            StateDelta: map[string]interface{}{
                "current_step": "preferences_updated",
                "last_update": time.Now().Unix(),
            },
        },
    }
    
    session.AddEvent(event)
    
    // 状态会自动更新
    if updatedStep, exists := session.GetState("current_step"); exists {
        fmt.Printf("更新后步骤: %s\n", updatedStep)
    }
}
```

### 事件处理
```go
func handleSessionEvents(session *sessions.Session) {
    // 获取所有事件
    allEvents := session.GetAllEvents()
    fmt.Printf("总事件数: %d\n", len(allEvents))
    
    // 根据类型过滤事件
    userMessages := make([]*events.Event, 0)
    for _, event := range allEvents {
        if event.Type == "user_message" {
            userMessages = append(userMessages, event)
        }
    }
    
    // 根据ID获取特定事件
    if len(allEvents) > 0 {
        firstEvent := allEvents[0]
        foundEvent := session.GetEvent(firstEvent.ID)
        if foundEvent != nil {
            fmt.Printf("找到事件: %s\n", foundEvent.Data)
        }
    }
}
```

## 高级用法

### 分页查询事件
```go
func listEventsWithPaging(sessionService sessions.SessionService, appName, userID, sessionID string) {
    ctx := context.Background()
    var nextPageToken string
    allEvents := make([]*events.Event, 0)
    
    for {
        response, err := sessionService.ListEvents(ctx, appName, userID, sessionID, nextPageToken, 10)
        if err != nil {
            fmt.Printf("查询事件失败: %v\n", err)
            break
        }
        
        allEvents = append(allEvents, response.Events...)
        
        if response.NextPageToken == "" {
            break // 没有更多页面
        }
        nextPageToken = response.NextPageToken
    }
    
    fmt.Printf("总共获取到 %d 个事件\n", len(allEvents))
}
```

### 会话搜索和过滤
```go
func filterSessions(sessionService sessions.SessionService, appName, userID string) {
    ctx := context.Background()
    
    // 获取用户的所有会话
    response, err := sessionService.ListSessions(ctx, appName, userID)
    if err != nil {
        fmt.Printf("获取会话列表失败: %v\n", err)
        return
    }
    
    // 按创建时间排序
    sessions := response.Sessions
    sort.Slice(sessions, func(i, j int) bool {
        return sessions[i].CreateTime.After(sessions[j].CreateTime)
    })
    
    // 过滤最近一周的会话
    oneWeekAgo := time.Now().AddDate(0, 0, -7)
    recentSessions := make([]*sessions.Session, 0)
    
    for _, session := range sessions {
        if session.CreateTime.After(oneWeekAgo) {
            recentSessions = append(recentSessions, session)
        }
    }
    
    fmt.Printf("最近一周的会话数: %d\n", len(recentSessions))
}
```

## 配置和扩展

### 会话服务工厂
```go
type SessionServiceConfig struct {
    Type     string                 `yaml:"type"`     // inmemory/database/vertexai
    Options  map[string]interface{} `yaml:"options"`  // 服务特定配置
}

func CreateSessionService(config *SessionServiceConfig) sessions.SessionService {
    switch config.Type {
    case "inmemory":
        return sessions.NewInMemorySessionService()
    case "database":
        return sessions.NewDatabaseSessionService(config.Options)
    case "vertexai":
        return sessions.NewVertexAISessionService(config.Options)
    default:
        return sessions.NewInMemorySessionService()
    }
}
```

### 自定义会话服务实现
```go
type CustomSessionService struct {
    // 自定义字段
    storage map[string]*sessions.Session
    mu      sync.RWMutex
}

func (s *CustomSessionService) CreateSession(ctx context.Context, appName, userID string, state map[string]interface{}, sessionID string) (*sessions.Session, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    session := sessions.NewSession(appName, userID, state, sessionID)
    key := fmt.Sprintf("%s:%s:%s", appName, userID, session.ID)
    s.storage[key] = session
    
    return session, nil
}

// 实现其他接口方法...
```

## 数据模型设计

### 数据库表结构（参考）
```sql
-- 会话表
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    app_name VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    state_data JSON,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_user (app_name, user_id),
    INDEX idx_create_time (create_time)
);

-- 事件表
CREATE TABLE session_events (
    id VARCHAR(255) PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    event_data JSON,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    INDEX idx_session_time (session_id, create_time)
);
```

## 性能优化

### 1. 事件分页
- 大会话中的事件支持分页查询
- 避免一次性加载所有历史事件

### 2. 状态缓存
- 频繁访问的会话状态进行缓存
- 减少数据库查询次数

### 3. 异步处理
```go
func asyncUpdateSession(sessionService sessions.SessionService, session *sessions.Session) {
    go func() {
        ctx := context.Background()
        if err := sessionService.UpdateSession(ctx, session); err != nil {
            log.Printf("异步更新会话失败: %v", err)
        }
    }()
}
```

## 最佳实践

1. **会话生命周期**: 合理设置会话超时和清理策略
2. **状态管理**: 避免在会话状态中存储大量数据
3. **事件设计**: 事件应该具有明确的类型和结构
4. **并发安全**: 多线程环境下注意状态同步
5. **持久化选择**: 根据需求选择合适的存储方案
6. **监控告警**: 监控会话数量和存储使用情况

## 集成示例

### 与智能体集成
```go
type AgentWithSession struct {
    agent          agents.Agent
    sessionService sessions.SessionService
}

func (a *AgentWithSession) Process(ctx context.Context, input string) (string, error) {
    userID := ctx.Value("user_id").(string)
    sessionID := ctx.Value("session_id").(string)
    
    // 获取会话上下文
    session, err := a.sessionService.GetSession(ctx, "my-app", userID, sessionID, nil)
    if err != nil {
        return "", err
    }
    
    // 添加用户输入事件
    userEvent := &events.Event{
        Type: "user_message",
        Data: input,
    }
    session.AddEvent(userEvent)
    
    // 处理请求
    output, err := a.agent.Process(ctx, input)
    if err != nil {
        return "", err
    }
    
    // 添加助手响应事件
    assistantEvent := &events.Event{
        Type: "assistant_response",
        Data: output,
    }
    session.AddEvent(assistantEvent)
    
    // 更新会话
    err = a.sessionService.UpdateSession(ctx, session)
    return output, err
}
```

## 依赖

- `github.com/google/uuid`: UUID 生成
- `github.com/nvcnvn/adk-golang/pkg/events`: 事件类型
- Go 标准库: `context`, `time`, `sync`

Sessions 模块为 ADK-Golang 框架提供了强大的会话管理能力，是构建有状态智能对话系统的重要基础设施。
