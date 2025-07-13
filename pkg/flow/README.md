# Flow 工作流核心模块

## 概述

Flow 模块提供了 ADK 框架的工作流核心功能，支持插件化的工作流管理和动态加载。该模块实现了分层智能体的配置管理、插件系统和工作流路由，是构建复杂智能体工作流的基础架构组件。

## 核心组件

### 1. FlowPlugin 接口
```go
type FlowPlugin interface {
    Name() string                   // 工作流名称 (唯一标识)
    Build() (*agents.Agent, error)  // 构造顶层智能体
}
```

定义了工作流插件的标准接口，所有工作流插件都必须实现此接口。

### 2. Manager 工作流管理器
```go
type Manager struct {
    mu    sync.RWMutex
    flows map[string]*agents.Agent
}
```

提供线程安全的工作流管理功能：
- 工作流注册和注销
- 工作流查询和列表
- 并发访问控制

### 3. 配置系统

#### AgentConfig (智能体配置)
```go
type AgentConfig struct {
    ID           string                 `json:"id" validate:"required"`
    Type         string                 `json:"type" validate:"required,oneof=sequential parallel leaf"`
    Model        string                 `json:"model,omitempty"`
    Instruction  string                 `json:"instruction,omitempty"`
    Description  string                 `json:"description,omitempty"`
    Workers      int                    `json:"workers,omitempty"`
    StreamOutput bool                   `json:"stream_output,omitempty"`
    Params       map[string]interface{} `json:"params,omitempty"`
    SubAgents    []AgentConfig          `json:"sub_agents,omitempty"`
}
```

支持递归的分层智能体配置，可构建复杂的智能体层次结构。

#### FlowConfig (工作流配置)
```go
type FlowConfig struct {
    Name        string            `json:"name" validate:"required"`
    Description string            `json:"description,omitempty"`
    Version     string            `json:"version,omitempty"`
    Queue       *QueueConfig      `json:"queue,omitempty"`
    PreGenerate *PreGenerateConfig `json:"pre_generate,omitempty"`
    Agents      []AgentConfig     `json:"agents" validate:"required,dive"`
    Routes      []RouteConfig     `json:"routes,omitempty"`
    Storage     *StorageConfig    `json:"storage,omitempty"`
}
```

完整的工作流配置结构，支持队列、存储、路由等高级功能。

## 使用示例

### 基础工作流插件
```go
package main

import (
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
)

// 实现 FlowPlugin 接口
type ChatFlowPlugin struct{}

func (p *ChatFlowPlugin) Name() string {
    return "chat_flow_v1"
}

func (p *ChatFlowPlugin) Build() (*agents.Agent, error) {
    // 创建聊天智能体配置
    config := &agents.AgentConfig{
        ID:          "chat_agent",
        Type:        "leaf",
        Model:       "gpt-4",
        Instruction: "你是一个友好的聊天助手，请用中文回答用户的问题。",
        Description: "智能聊天助手",
        StreamOutput: true,
    }
    
    // 构建智能体
    agent, err := agents.NewAgent(config)
    if err != nil {
        return nil, err
    }
    
    return agent, nil
}

// 导出插件变量（必需）
var Plugin = &ChatFlowPlugin{}

func main() {
    // 演示用法
    manager := flow.NewManager()
    
    // 注册工作流
    plugin := &ChatFlowPlugin{}
    agent, err := plugin.Build()
    if err != nil {
        panic(err)
    }
    
    manager.Register(plugin.Name(), agent)
    
    // 使用工作流
    chatFlow, exists := manager.Get("chat_flow_v1")
    if exists {
        // 执行工作流
        ctx := context.Background()
        result, err := chatFlow.Process(ctx, "你好，请介绍一下你自己")
        if err != nil {
            panic(err)
        }
        fmt.Printf("智能体回复: %s\n", result)
    }
}
```

### 分层智能体工作流
```go
type NovelWritingFlowPlugin struct{}

func (p *NovelWritingFlowPlugin) Name() string {
    return "novel_writing_v4"
}

func (p *NovelWritingFlowPlugin) Build() (*agents.Agent, error) {
    // 分层智能体配置
    config := &agents.AgentConfig{
        ID:          "novel_root",
        Type:        "sequential",
        Description: "小说创作工作流",
        SubAgents: []agents.AgentConfig{
            // 决策层
            {
                ID:          "decision_layer",
                Type:        "sequential", 
                Description: "决策层：规划和评估",
                SubAgents: []agents.AgentConfig{
                    {
                        ID:          "architect",
                        Type:        "leaf",
                        Model:       "gpt-4",
                        Instruction: "你是小说架构师，负责创建故事大纲和结构。",
                        Description: "故事架构师",
                    },
                    {
                        ID:          "planner", 
                        Type:        "leaf",
                        Model:       "gpt-4",
                        Instruction: "你是情节规划师，负责详细的章节规划。",
                        Description: "情节规划师",
                    },
                },
            },
            // 执行层
            {
                ID:          "execution_layer",
                Type:        "parallel",
                Description: "执行层：并行内容生成",
                Workers:     3,
                SubAgents: []agents.AgentConfig{
                    {
                        ID:          "character_writer",
                        Type:        "leaf",
                        Model:       "gpt-4",
                        Instruction: "你是角色创作专家，负责角色设定和对话。",
                        Description: "角色创作者",
                    },
                    {
                        ID:          "scene_writer",
                        Type:        "leaf", 
                        Model:       "gpt-4",
                        Instruction: "你是场景描写专家，负责环境和场景描述。",
                        Description: "场景描写者",
                    },
                    {
                        ID:          "plot_writer",
                        Type:        "leaf",
                        Model:       "gpt-4", 
                        Instruction: "你是情节编写专家，负责故事情节发展。",
                        Description: "情节编写者",
                    },
                },
            },
        },
    }
    
    return agents.NewAgent(config)
}

var Plugin = &NovelWritingFlowPlugin{}
```

### 工作流配置文件
```json
{
  "name": "advanced_novel_flow",
  "description": "高级小说创作工作流",
  "version": "2.0.0",
  "queue": {
    "impl": "redis",
    "stream": "novel_tasks",
    "max_len": 1000
  },
  "pre_generate": {
    "enabled": true,
    "agent": "outline_generator",
    "timeout_ms": 30000
  },
  "agents": [
    {
      "id": "novel_coordinator",
      "type": "sequential",
      "description": "小说创作协调器",
      "sub_agents": [
        {
          "id": "planning_stage",
          "type": "sequential",
          "description": "规划阶段",
          "sub_agents": [
            {
              "id": "theme_analyzer",
              "type": "leaf",
              "model": "gpt-4",
              "instruction": "分析用户输入的主题和要求，提取关键元素",
              "description": "主题分析器"
            },
            {
              "id": "outline_creator",
              "type": "leaf", 
              "model": "gpt-4",
              "instruction": "根据主题分析结果创建详细的故事大纲",
              "description": "大纲创建器"
            }
          ]
        },
        {
          "id": "writing_stage",
          "type": "parallel",
          "description": "写作阶段",
          "workers": 4,
          "sub_agents": [
            {
              "id": "character_developer",
              "type": "leaf",
              "model": "gpt-4",
              "instruction": "深度开发人物角色，包括背景、性格、动机",
              "description": "角色开发者",
              "stream_output": true
            },
            {
              "id": "world_builder",
              "type": "leaf",
              "model": "gpt-4", 
              "instruction": "构建故事世界观，包括设定、规则、历史",
              "description": "世界构建者"
            },
            {
              "id": "dialogue_writer",
              "type": "leaf",
              "model": "gpt-4",
              "instruction": "编写生动的对话和角色互动",
              "description": "对话编写者"
            },
            {
              "id": "narrative_writer",
              "type": "leaf",
              "model": "gpt-4",
              "instruction": "编写叙述性文本和情节描述",
              "description": "叙述编写者"
            }
          ]
        }
      ]
    }
  ],
  "routes": [
    {
      "path": "/api/novel/create",
      "agent": "novel_coordinator"
    },
    {
      "path": "/api/novel/character",
      "agent": "character_developer"
    },
    {
      "path": "/api/novel/world",
      "agent": "world_builder"
    }
  ],
  "storage": {
    "dsn": "postgres://user:pass@localhost/novel_db?sslmode=disable"
  }
}
```

### 插件动态加载
```go
package main

import (
    "fmt"
    "plugin"
    
    "github.com/nvcnvn/adk-golang/pkg/flow"
)

func loadFlowPlugin(pluginPath string) error {
    // 加载插件动态库
    p, err := plugin.Open(pluginPath)
    if err != nil {
        return fmt.Errorf("加载插件失败: %w", err)
    }
    
    // 查找插件变量
    sym, err := p.Lookup("Plugin")
    if err != nil {
        return fmt.Errorf("插件未导出Plugin变量: %w", err)
    }
    
    // 类型断言
    flowPlugin, ok := sym.(flow.FlowPlugin)
    if !ok {
        return fmt.Errorf("Plugin不是有效的FlowPlugin类型")
    }
    
    // 构建智能体
    agent, err := flowPlugin.Build()
    if err != nil {
        return fmt.Errorf("构建智能体失败: %w", err)
    }
    
    // 注册到管理器
    GlobalManager.Register(flowPlugin.Name(), agent)
    
    fmt.Printf("成功加载工作流插件: %s\n", flowPlugin.Name())
    return nil
}

// 批量加载插件
func loadPluginsFromDirectory(pluginDir string) error {
    files, err := filepath.Glob(filepath.Join(pluginDir, "*.so"))
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if err := loadFlowPlugin(file); err != nil {
            fmt.Printf("警告: 加载插件 %s 失败: %v\n", file, err)
        }
    }
    
    return nil
}

// 插件热重载
func reloadPlugin(pluginName, pluginPath string) error {
    // 注销旧插件
    GlobalManager.Unregister(pluginName)
    
    // 加载新插件
    return loadFlowPlugin(pluginPath)
}
```

### 工作流执行管理
```go
type FlowExecutor struct {
    manager    *flow.Manager
    tracer     *telemetry.Tracer
    queue      *Queue
    storage    *Storage
}

func NewFlowExecutor(manager *flow.Manager) *FlowExecutor {
    return &FlowExecutor{
        manager: manager,
        tracer:  telemetry.NewTracer(),
    }
}

func (e *FlowExecutor) ExecuteFlow(ctx context.Context, flowName, input string) (*ExecutionResult, error) {
    // 生成追踪ID
    traceID := flow.TraceID()
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    // 开始追踪
    span := e.tracer.StartSpan(ctx, "flow_execution")
    defer span.End()
    
    // 获取工作流
    agent, exists := e.manager.Get(flowName)
    if !exists {
        return nil, fmt.Errorf("工作流 %s 不存在", flowName)
    }
    
    // 记录开始时间
    startTime := time.Now()
    
    // 执行工作流
    result, err := agent.Process(ctx, input)
    if err != nil {
        span.RecordError(err)
        return nil, fmt.Errorf("工作流执行失败: %w", err)
    }
    
    // 计算执行时间
    duration := time.Since(startTime)
    
    // 记录指标
    span.SetAttributes(map[string]interface{}{
        "flow_name":     flowName,
        "input_length":  len(input),
        "output_length": len(result),
        "duration_ms":   duration.Milliseconds(),
        "success":       true,
    })
    
    return &ExecutionResult{
        FlowName:     flowName,
        TraceID:      traceID,
        Input:        input,
        Output:       result,
        Duration:     duration,
        Success:      true,
    }, nil
}

type ExecutionResult struct {
    FlowName     string        `json:"flow_name"`
    TraceID      string        `json:"trace_id"`
    Input        string        `json:"input"`
    Output       string        `json:"output"`
    Duration     time.Duration `json:"duration"`
    Success      bool          `json:"success"`
    Error        string        `json:"error,omitempty"`
}
```

## 高级功能

### 队列集成
```go
type QueueHandler struct {
    manager  *flow.Manager
    executor *FlowExecutor
}

func (h *QueueHandler) ProcessMessage(ctx context.Context, msg *QueueMessage) error {
    // 解析消息
    var request FlowRequest
    if err := json.Unmarshal(msg.Data, &request); err != nil {
        return fmt.Errorf("解析请求失败: %w", err)
    }
    
    // 执行工作流
    result, err := h.executor.ExecuteFlow(ctx, request.FlowName, request.Input)
    if err != nil {
        return fmt.Errorf("执行工作流失败: %w", err)
    }
    
    // 发送结果
    if request.CallbackURL != "" {
        return h.sendCallback(request.CallbackURL, result)
    }
    
    return nil
}

type FlowRequest struct {
    FlowName    string `json:"flow_name"`
    Input       string `json:"input"`
    CallbackURL string `json:"callback_url,omitempty"`
    UserID      string `json:"user_id,omitempty"`
    SessionID   string `json:"session_id,omitempty"`
}
```

### 存储集成
```go
type FlowStorage struct {
    db *gorm.DB
}

type FlowExecution struct {
    ID        uint      `gorm:"primarykey"`
    FlowName  string    `gorm:"index"`
    TraceID   string    `gorm:"uniqueIndex"`
    Input     string    `gorm:"type:text"`
    Output    string    `gorm:"type:text"`
    UserID    string    `gorm:"index"`
    SessionID string    `gorm:"index"`
    Duration  int64     // 毫秒
    Success   bool
    Error     string    `gorm:"type:text"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (s *FlowStorage) SaveExecution(result *ExecutionResult) error {
    execution := &FlowExecution{
        FlowName:  result.FlowName,
        TraceID:   result.TraceID,
        Input:     result.Input,
        Output:    result.Output,
        Duration:  result.Duration.Milliseconds(),
        Success:   result.Success,
        Error:     result.Error,
    }
    
    return s.db.Create(execution).Error
}

func (s *FlowStorage) GetExecutionHistory(userID, sessionID string, limit int) ([]*FlowExecution, error) {
    var executions []*FlowExecution
    
    query := s.db.Where("user_id = ? AND session_id = ?", userID, sessionID).
        Order("created_at DESC").
        Limit(limit)
    
    err := query.Find(&executions).Error
    return executions, err
}
```

### 监控和指标
```go
type FlowMetrics struct {
    TotalExecutions   int64         `json:"total_executions"`
    SuccessfulRuns    int64         `json:"successful_runs"`
    FailedRuns        int64         `json:"failed_runs"`
    AvgDuration       time.Duration `json:"avg_duration"`
    TopFlows          []FlowStat    `json:"top_flows"`
    RecentErrors      []ErrorStat   `json:"recent_errors"`
}

type FlowStat struct {
    Name       string        `json:"name"`
    Count      int64         `json:"count"`
    AvgDuration time.Duration `json:"avg_duration"`
    SuccessRate float64       `json:"success_rate"`
}

type ErrorStat struct {
    FlowName  string    `json:"flow_name"`
    Error     string    `json:"error"`
    Count     int64     `json:"count"`
    LastSeen  time.Time `json:"last_seen"`
}

func (e *FlowExecutor) GetMetrics() *FlowMetrics {
    // 实现指标收集逻辑
    return &FlowMetrics{
        TotalExecutions: e.totalExecutions,
        SuccessfulRuns:  e.successfulRuns,
        FailedRuns:      e.failedRuns,
        AvgDuration:     e.avgDuration,
        TopFlows:        e.getTopFlows(),
        RecentErrors:    e.getRecentErrors(),
    }
}
```

### 配置验证
```go
import "github.com/go-playground/validator/v10"

func (cfg *FlowConfig) Validate() error {
    validate := validator.New()
    
    // 注册自定义验证器
    validate.RegisterValidation("agent_type", validateAgentType)
    validate.RegisterValidation("model_available", validateModelAvailable)
    
    if err := validate.Struct(cfg); err != nil {
        return fmt.Errorf("配置验证失败: %w", err)
    }
    
    // 自定义业务逻辑验证
    return cfg.validateBusinessRules()
}

func validateAgentType(fl validator.FieldLevel) bool {
    agentType := fl.Field().String()
    validTypes := []string{"sequential", "parallel", "leaf"}
    
    for _, vt := range validTypes {
        if agentType == vt {
            return true
        }
    }
    return false
}

func (cfg *FlowConfig) validateBusinessRules() error {
    // 检查智能体引用的有效性
    agentMap := make(map[string]bool)
    if err := cfg.buildAgentMap(cfg.Agents, agentMap); err != nil {
        return err
    }
    
    // 检查路由配置
    for _, route := range cfg.Routes {
        if !agentMap[route.Agent] {
            return fmt.Errorf("路由 %s 引用了不存在的智能体: %s", route.Path, route.Agent)
        }
    }
    
    return nil
}
```

## 最佳实践

1. **插件开发**: 遵循插件接口标准，实现清晰的错误处理
2. **配置管理**: 使用JSON配置文件，支持验证和版本控制
3. **资源管理**: 合理设置并发数，避免资源耗尽
4. **监控日志**: 启用追踪和指标收集，便于问题诊断
5. **热更新**: 利用插件系统实现工作流的热更新
6. **存储优化**: 合理使用数据库索引，优化查询性能

## 依赖模块

- `github.com/nvcnvn/adk-golang/pkg/agents`: 智能体核心
- `github.com/google/uuid`: UUID生成
- `github.com/go-playground/validator/v10`: 配置验证
- `gorm.io/gorm`: 数据库ORM
- Go 标准库: `plugin`, `sync`, `context`

## 扩展开发

### 自定义插件加载器
```go
type CustomPluginLoader struct {
    pluginDir string
    watcher   *fsnotify.Watcher
    manager   *flow.Manager
}

func (l *CustomPluginLoader) WatchAndReload() {
    for {
        select {
        case event := <-l.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                if strings.HasSuffix(event.Name, ".so") {
                    l.reloadPlugin(event.Name)
                }
            }
        case err := <-l.watcher.Errors:
            log.Printf("文件监控错误: %v", err)
        }
    }
}
```

Flow 模块为 ADK-Golang 框架提供了强大的工作流管理和插件系统，是构建复杂智能体应用的核心基础设施。
