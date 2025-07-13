# Runners 运行器模块

## 概述

Runners 模块为 ADK 框架提供智能体运行功能，支持单次运行和交互式会话模式。该模块包含会话管理、遥测追踪、JSON 输出格式等功能，是智能体执行的核心组件。

## 核心接口

### Runner 接口
```go
type Runner interface {
    // Run 运行智能体处理单次输入
    Run(ctx context.Context, agent *agents.Agent, input string) (string, error)
    
    // RunInteractive 运行智能体的交互式会话模式
    RunInteractive(ctx context.Context, agent *agents.Agent, in io.Reader, out io.Writer) error
    
    // SetSaveSessionEnabled 启用或禁用会话保存
    SetSaveSessionEnabled(enabled bool)
}
```

## 数据结构

### Interaction 交互记录
```go
type Interaction struct {
    Input     string    `json:"input"`     // 用户输入
    Response  string    `json:"response"`  // 智能体响应
    Timestamp time.Time `json:"timestamp"` // 时间戳
}
```

### Session 会话
```go
type Session struct {
    AgentName    string        `json:"agent_name"`    // 智能体名称
    AgentModel   string        `json:"agent_model"`   // 模型名称
    StartTime    time.Time     `json:"start_time"`    // 开始时间
    EndTime      time.Time     `json:"end_time"`      // 结束时间
    Interactions []Interaction `json:"interactions"`  // 交互记录列表
}
```

## SimpleRunner 实现

### 基本结构
```go
type SimpleRunner struct {
    saveSession bool    // 是否保存会话
    jsonOutput  bool    // 是否使用 JSON 输出
    session     Session // 当前会话数据
}
```

### 创建运行器
```go
func NewSimpleRunner() *SimpleRunner

// 示例
runner := runners.NewSimpleRunner()
runner.SetSaveSessionEnabled(true)  // 启用会话保存
runner.SetJSONOutput(false)         // 禁用 JSON 输出
```

## 使用示例

### 单次运行
```go
package main

import (
    "context"
    "fmt"
    
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/runners"
)

func main() {
    // 创建运行器
    runner := runners.NewSimpleRunner()
    
    // 创建智能体（示例）
    agent := createMyAgent()
    
    // 单次运行
    ctx := context.Background()
    response, err := runner.Run(ctx, agent, "你好，请介绍一下自己")
    if err != nil {
        fmt.Printf("运行错误: %v\n", err)
        return
    }
    
    fmt.Printf("响应: %s\n", response)
}

func createMyAgent() *agents.Agent {
    // 创建智能体的逻辑
    return &agents.Agent{} // 简化示例
}
```

### 交互式会话
```go
package main

import (
    "context"
    "os"
    
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/runners"
)

func main() {
    // 创建运行器并启用会话保存
    runner := runners.NewSimpleRunner()
    runner.SetSaveSessionEnabled(true)
    
    // 创建智能体
    agent := createMyAgent()
    
    // 启动交互式会话
    ctx := context.Background()
    err := runner.RunInteractive(ctx, agent, os.Stdin, os.Stdout)
    if err != nil {
        fmt.Printf("交互式会话错误: %v\n", err)
    }
}
```

### JSON 输出模式
```go
func runWithJSONOutput() {
    runner := runners.NewSimpleRunner()
    runner.SetJSONOutput(true)  // 启用 JSON 输出
    
    agent := createMyAgent()
    ctx := context.Background()
    
    // 在 JSON 模式下，交互式会话输出将是 JSON 格式
    err := runner.RunInteractive(ctx, agent, os.Stdin, os.Stdout)
    if err != nil {
        fmt.Printf("JSON 会话错误: %v\n", err)
    }
}
```

## 会话管理

### 自动保存会话
```go
func demonstrateSessionSaving() {
    runner := runners.NewSimpleRunner()
    runner.SetSaveSessionEnabled(true) // 启用会话保存
    
    agent := createMyAgent()
    ctx := context.Background()
    
    // 运行交互式会话
    // 会话结束时将自动保存到 ./sessions/ 目录
    runner.RunInteractive(ctx, agent, os.Stdin, os.Stdout)
    
    // 会话文件格式: {agent_name}_session_{timestamp}.json
    // 例如: MyAgent_session_20250713_143052.json
}
```

### 会话文件结构
```json
{
  "agent_name": "MyAgent",
  "agent_model": "gpt-4",
  "start_time": "2025-07-13T14:30:52Z",
  "end_time": "2025-07-13T14:35:20Z",
  "interactions": [
    {
      "input": "你好",
      "response": "你好！我是一个AI助手，很高兴为您服务。",
      "timestamp": "2025-07-13T14:31:05Z"
    },
    {
      "input": "请介绍一下Go语言",
      "response": "Go语言是Google开发的开源编程语言...",
      "timestamp": "2025-07-13T14:32:15Z"
    }
  ]
}
```

## 遥测集成

SimpleRunner 自动集成遥测功能：

```go
// 运行器自动创建追踪跨度
func (r *SimpleRunner) Run(ctx context.Context, agent *agents.Agent, input string) (string, error) {
    // 创建运行跨度
    ctx, span := telemetry.StartSpan(ctx, "SimpleRunner.Run")
    defer span.End()
    
    // 添加智能体元数据
    span.SetAttribute("agent.name", agent.Name())
    span.SetAttribute("agent.model", agent.Model())
    
    // 处理并追踪错误
    response, err := agent.Process(ctx, input)
    if err != nil {
        span.SetAttribute("error", err.Error())
        return "", err
    }
    
    return response, nil
}
```

## 高级功能

### 自定义运行器
```go
type CustomRunner struct {
    *runners.SimpleRunner
    middleware []Middleware
    validator  InputValidator
}

type Middleware func(context.Context, string) (context.Context, string, error)
type InputValidator func(string) error

func NewCustomRunner() *CustomRunner {
    return &CustomRunner{
        SimpleRunner: runners.NewSimpleRunner(),
        middleware:   []Middleware{},
    }
}

func (cr *CustomRunner) AddMiddleware(mw Middleware) {
    cr.middleware = append(cr.middleware, mw)
}

func (cr *CustomRunner) SetInputValidator(validator InputValidator) {
    cr.validator = validator
}

func (cr *CustomRunner) Run(ctx context.Context, agent *agents.Agent, input string) (string, error) {
    // 输入验证
    if cr.validator != nil {
        if err := cr.validator(input); err != nil {
            return "", fmt.Errorf("输入验证失败: %w", err)
        }
    }
    
    // 应用中间件
    processedInput := input
    for _, mw := range cr.middleware {
        var err error
        ctx, processedInput, err = mw(ctx, processedInput)
        if err != nil {
            return "", fmt.Errorf("中间件处理失败: %w", err)
        }
    }
    
    // 调用基础运行器
    return cr.SimpleRunner.Run(ctx, agent, processedInput)
}
```

### 并发运行器
```go
type ConcurrentRunner struct {
    maxConcurrency int
    semaphore      chan struct{}
}

func NewConcurrentRunner(maxConcurrency int) *ConcurrentRunner {
    return &ConcurrentRunner{
        maxConcurrency: maxConcurrency,
        semaphore:      make(chan struct{}, maxConcurrency),
    }
}

func (cr *ConcurrentRunner) RunBatch(
    ctx context.Context,
    agent *agents.Agent,
    inputs []string,
) ([]string, error) {
    results := make([]string, len(inputs))
    errors := make([]error, len(inputs))
    
    var wg sync.WaitGroup
    
    for i, input := range inputs {
        wg.Add(1)
        go func(index int, inp string) {
            defer wg.Done()
            
            // 获取信号量
            cr.semaphore <- struct{}{}
            defer func() { <-cr.semaphore }()
            
            // 运行智能体
            runner := runners.NewSimpleRunner()
            result, err := runner.Run(ctx, agent, inp)
            
            results[index] = result
            errors[index] = err
        }(i, input)
    }
    
    wg.Wait()
    
    // 检查错误
    for _, err := range errors {
        if err != nil {
            return results, err
        }
    }
    
    return results, nil
}
```

## 错误处理

### 常见错误处理
```go
func handleRunnerErrors() {
    runner := runners.NewSimpleRunner()
    agent := createMyAgent()
    ctx := context.Background()
    
    response, err := runner.Run(ctx, agent, "测试输入")
    if err != nil {
        switch {
        case errors.Is(err, context.DeadlineExceeded):
            fmt.Println("运行超时")
        case errors.Is(err, context.Canceled):
            fmt.Println("运行被取消")
        case strings.Contains(err.Error(), "agent cannot be nil"):
            fmt.Println("智能体为空")
        default:
            fmt.Printf("未知错误: %v\n", err)
        }
        return
    }
    
    fmt.Printf("成功响应: %s\n", response)
}
```

### 重试机制
```go
type RetryRunner struct {
    baseRunner runners.Runner
    maxRetries int
    backoff    time.Duration
}

func NewRetryRunner(baseRunner runners.Runner, maxRetries int, backoff time.Duration) *RetryRunner {
    return &RetryRunner{
        baseRunner: baseRunner,
        maxRetries: maxRetries,
        backoff:    backoff,
    }
}

func (rr *RetryRunner) Run(ctx context.Context, agent *agents.Agent, input string) (string, error) {
    var lastErr error
    
    for i := 0; i <= rr.maxRetries; i++ {
        response, err := rr.baseRunner.Run(ctx, agent, input)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        
        if i < rr.maxRetries {
            select {
            case <-time.After(rr.backoff * time.Duration(i+1)):
                // 退避等待
            case <-ctx.Done():
                return "", ctx.Err()
            }
        }
    }
    
    return "", fmt.Errorf("重试 %d 次后仍然失败: %w", rr.maxRetries, lastErr)
}
```

## 性能优化

### 资源池化
```go
type PooledRunner struct {
    runnerPool sync.Pool
}

func NewPooledRunner() *PooledRunner {
    return &PooledRunner{
        runnerPool: sync.Pool{
            New: func() interface{} {
                return runners.NewSimpleRunner()
            },
        },
    }
}

func (pr *PooledRunner) Run(ctx context.Context, agent *agents.Agent, input string) (string, error) {
    runner := pr.runnerPool.Get().(*runners.SimpleRunner)
    defer pr.runnerPool.Put(runner)
    
    return runner.Run(ctx, agent, input)
}
```

## 最佳实践

1. **会话管理**: 对长时间交互启用会话保存
2. **资源控制**: 使用超时和取消机制
3. **错误处理**: 实现适当的重试和回退策略
4. **遥测**: 利用内置的追踪功能进行监控
5. **并发控制**: 限制并发运行数量避免资源耗尽

## 依赖模块

- `github.com/nvcnvn/adk-golang/pkg/agents`: 智能体核心
- `github.com/nvcnvn/adk-golang/pkg/telemetry`: 遥测追踪
- Go 标准库: `context`, `encoding/json`, `io`, `os`, `path/filepath`, `time`

Runners 模块为 ADK-Golang 框架提供了强大的智能体运行能力，支持多种运行模式和高级功能，是构建智能体应用的重要基础组件。
