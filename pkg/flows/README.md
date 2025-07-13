# Flows 工作流模块

## 概述

Flows 模块提供了智能体工作流的实现框架，支持构建复杂的多步骤处理流程。该模块是智能体执行复杂任务的核心引擎，提供了从简单的LLM调用到复杂的代码执行和多智能体协作的完整工作流解决方案。

## 核心组件

### 工作流类型

#### 1. BasicFlow (基础工作流)
标准的LLM处理工作流，支持请求预处理和响应后处理。

#### 2. IdentityFlow (身份工作流)
最小化的透传工作流，不进行任何处理，主要用于测试和调试。

#### 3. CodeExecutionFlow (代码执行工作流)
支持代码生成和执行的工作流，集成了代码执行器功能。

### 工作流工厂方法

```go
// 创建基础工作流
func CreateBasicFlow() *llm_flows.BasicFlow

// 创建身份工作流
func CreateIdentityFlow() *llm_flows.IdentityFlow

// 创建代码执行工作流
func CreateCodeExecutionFlow(codeExecutorName string) (*llm_flows.BasicFlow, error)

// 创建代码执行器
func CreateCodeExecutor(name string) (code_executors.CodeExecutor, error)
```

## 子模块结构

### 1. llm_flows (LLM工作流)
核心的语言模型处理工作流实现，提供：
- 请求处理器链
- 响应处理器链
- 内容处理和指令处理
- 流式处理支持

### 2. novel (小说生成工作流)
专门用于创意写作和小说生成的工作流实现。

### 3. novel_v3 & novel_v4 (小说生成工作流 v3/v4)
升级版的小说生成工作流，支持更复杂的创作流程和质量控制。

### 4. test_rag_flow (RAG测试工作流)
用于检索增强生成(RAG)的测试和验证工作流。

## 使用示例

### 基础工作流使用
```go
package main

import (
    "context"
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/flows"
)

func main() {
    ctx := context.Background()
    
    // 创建基础工作流
    flow := flows.CreateBasicFlow()
    
    // 配置工作流参数
    request := &FlowRequest{
        Query: "请帮我写一个Go语言的Hello World程序",
        Context: map[string]interface{}{
            "language": "go",
            "style": "simple",
        },
    }
    
    // 执行工作流
    response, err := flow.Execute(ctx, request)
    if err != nil {
        fmt.Printf("工作流执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("工作流响应: %s\n", response.Content)
}
```

### 代码执行工作流
```go
func codeExecutionExample() {
    ctx := context.Background()
    
    // 创建代码执行工作流
    flow, err := flows.CreateCodeExecutionFlow("local")
    if err != nil {
        fmt.Printf("创建代码执行工作流失败: %v\n", err)
        return
    }
    
    // 请求生成并执行代码
    request := &FlowRequest{
        Query: "写一个计算斐波那契数列的Python函数并执行计算前10项",
        Context: map[string]interface{}{
            "language": "python",
            "execute": true,
        },
    }
    
    response, err := flow.Execute(ctx, request)
    if err != nil {
        fmt.Printf("代码执行工作流失败: %v\n", err)
        return
    }
    
    fmt.Printf("代码执行结果: %s\n", response.Content)
    
    // 获取执行详情
    if executionResult, ok := response.Metadata["execution_result"]; ok {
        fmt.Printf("执行详情: %v\n", executionResult)
    }
}
```

### 身份工作流（调试用）
```go
func identityFlowExample() {
    ctx := context.Background()
    
    // 创建身份工作流
    flow := flows.CreateIdentityFlow()
    
    // 透传请求
    request := &FlowRequest{
        Query: "这个请求会被直接返回",
        Context: map[string]interface{}{
            "debug": true,
        },
    }
    
    response, err := flow.Execute(ctx, request)
    if err != nil {
        fmt.Printf("身份工作流失败: %v\n", err)
        return
    }
    
    // 身份工作流会直接返回输入
    fmt.Printf("透传结果: %s\n", response.Content)
}
```

## 高级工作流配置

### 自定义工作流构建
```go
type CustomFlowBuilder struct {
    processors []RequestProcessor
    config     *FlowConfig
}

func NewCustomFlowBuilder() *CustomFlowBuilder {
    return &CustomFlowBuilder{
        processors: make([]RequestProcessor, 0),
        config:     &FlowConfig{},
    }
}

func (b *CustomFlowBuilder) AddProcessor(processor RequestProcessor) *CustomFlowBuilder {
    b.processors = append(b.processors, processor)
    return b
}

func (b *CustomFlowBuilder) SetConfig(config *FlowConfig) *CustomFlowBuilder {
    b.config = config
    return b
}

func (b *CustomFlowBuilder) Build() *llm_flows.BasicFlow {
    flow := llm_flows.NewBasicFlow()
    flow.RequestProcessors = b.processors
    flow.Config = b.config
    return flow
}

// 使用示例
func buildCustomFlow() *llm_flows.BasicFlow {
    return NewCustomFlowBuilder().
        AddProcessor(llm_flows.NewInstructionsProcessor()).
        AddProcessor(llm_flows.NewContentsProcessor()).
        AddProcessor(NewCustomValidationProcessor()).
        SetConfig(&FlowConfig{
            MaxRetries: 3,
            Timeout:    30 * time.Second,
        }).
        Build()
}
```

### 工作流链式处理
```go
type FlowChain struct {
    flows []Flow
    config *ChainConfig
}

func NewFlowChain() *FlowChain {
    return &FlowChain{
        flows: make([]Flow, 0),
        config: &ChainConfig{},
    }
}

func (c *FlowChain) AddFlow(flow Flow) *FlowChain {
    c.flows = append(c.flows, flow)
    return c
}

func (c *FlowChain) Execute(ctx context.Context, request *FlowRequest) (*FlowResponse, error) {
    var response *FlowResponse
    var err error
    
    currentRequest := request
    
    for i, flow := range c.flows {
        fmt.Printf("执行工作流 %d/%d\n", i+1, len(c.flows))
        
        response, err = flow.Execute(ctx, currentRequest)
        if err != nil {
            return nil, fmt.Errorf("工作流 %d 执行失败: %w", i+1, err)
        }
        
        // 将前一个工作流的输出作为下一个工作流的输入
        if i < len(c.flows)-1 {
            currentRequest = &FlowRequest{
                Query:   response.Content,
                Context: mergeContexts(currentRequest.Context, response.Metadata),
            }
        }
    }
    
    return response, nil
}

func mergeContexts(ctx1, ctx2 map[string]interface{}) map[string]interface{} {
    merged := make(map[string]interface{})
    
    for k, v := range ctx1 {
        merged[k] = v
    }
    for k, v := range ctx2 {
        merged[k] = v
    }
    
    return merged
}
```

## 专业领域工作流

### 小说生成工作流
```go
func novelGenerationExample() {
    ctx := context.Background()
    
    // 假设存在小说生成工作流
    novelFlow := flows.NewNovelFlow(&NovelConfig{
        Genre:     "科幻",
        Length:    "短篇",
        Style:     "现代",
        Character: 3,
    })
    
    request := &FlowRequest{
        Query: "写一个关于时间旅行的科幻短篇小说",
        Context: map[string]interface{}{
            "theme":      "时间悖论",
            "setting":    "2050年的未来世界",
            "protagonist": "年轻的物理学家",
        },
    }
    
    response, err := novelFlow.Execute(ctx, request)
    if err != nil {
        fmt.Printf("小说生成失败: %v\n", err)
        return
    }
    
    fmt.Printf("生成的小说:\n%s\n", response.Content)
    
    // 获取创作元数据
    if metadata, ok := response.Metadata["novel_metadata"]; ok {
        fmt.Printf("创作信息: %v\n", metadata)
    }
}
```

### RAG工作流
```go
func ragFlowExample() {
    ctx := context.Background()
    
    // 创建RAG工作流
    ragFlow := flows.NewRAGFlow(&RAGConfig{
        VectorStore:   "chroma",
        EmbeddingModel: "text-embedding-ada-002",
        TopK:          5,
        Threshold:     0.7,
    })
    
    request := &FlowRequest{
        Query: "Go语言中如何实现并发编程？",
        Context: map[string]interface{}{
            "knowledge_base": "golang_docs",
            "max_tokens":     1000,
        },
    }
    
    response, err := ragFlow.Execute(ctx, request)
    if err != nil {
        fmt.Printf("RAG工作流失败: %v\n", err)
        return
    }
    
    fmt.Printf("RAG响应: %s\n", response.Content)
    
    // 获取检索到的相关文档
    if sources, ok := response.Metadata["sources"]; ok {
        fmt.Printf("参考来源: %v\n", sources)
    }
}
```

## 工作流监控和调试

### 工作流执行监控
```go
type FlowMonitor struct {
    metrics map[string]*FlowMetrics
    mu      sync.RWMutex
}

type FlowMetrics struct {
    ExecutionCount int64         `json:"execution_count"`
    SuccessCount   int64         `json:"success_count"`
    FailureCount   int64         `json:"failure_count"`
    AvgDuration    time.Duration `json:"avg_duration"`
    LastExecution  time.Time     `json:"last_execution"`
}

func NewFlowMonitor() *FlowMonitor {
    return &FlowMonitor{
        metrics: make(map[string]*FlowMetrics),
    }
}

func (m *FlowMonitor) RecordExecution(flowName string, duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if _, exists := m.metrics[flowName]; !exists {
        m.metrics[flowName] = &FlowMetrics{}
    }
    
    metrics := m.metrics[flowName]
    metrics.ExecutionCount++
    metrics.LastExecution = time.Now()
    
    if success {
        metrics.SuccessCount++
    } else {
        metrics.FailureCount++
    }
    
    // 更新平均执行时间
    if metrics.ExecutionCount == 1 {
        metrics.AvgDuration = duration
    } else {
        metrics.AvgDuration = time.Duration(
            (int64(metrics.AvgDuration)*(metrics.ExecutionCount-1) + int64(duration)) / metrics.ExecutionCount,
        )
    }
}

func (m *FlowMonitor) GetMetrics(flowName string) *FlowMetrics {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if metrics, exists := m.metrics[flowName]; exists {
        return metrics
    }
    return nil
}

func (m *FlowMonitor) GetAllMetrics() map[string]*FlowMetrics {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    result := make(map[string]*FlowMetrics)
    for name, metrics := range m.metrics {
        result[name] = metrics
    }
    return result
}
```

### 工作流调试工具
```go
type FlowDebugger struct {
    enabled bool
    steps   []DebugStep
    mu      sync.Mutex
}

type DebugStep struct {
    StepName  string                 `json:"step_name"`
    Input     interface{}            `json:"input"`
    Output    interface{}            `json:"output"`
    Duration  time.Duration          `json:"duration"`
    Error     string                 `json:"error,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

func NewFlowDebugger(enabled bool) *FlowDebugger {
    return &FlowDebugger{
        enabled: enabled,
        steps:   make([]DebugStep, 0),
    }
}

func (d *FlowDebugger) RecordStep(stepName string, input, output interface{}, duration time.Duration, err error) {
    if !d.enabled {
        return
    }
    
    d.mu.Lock()
    defer d.mu.Unlock()
    
    step := DebugStep{
        StepName:  stepName,
        Input:     input,
        Output:    output,
        Duration:  duration,
        Timestamp: time.Now(),
    }
    
    if err != nil {
        step.Error = err.Error()
    }
    
    d.steps = append(d.steps, step)
}

func (d *FlowDebugger) GetDebugTrace() []DebugStep {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    result := make([]DebugStep, len(d.steps))
    copy(result, d.steps)
    return result
}

func (d *FlowDebugger) Clear() {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    d.steps = d.steps[:0] // 重置但保留容量
}
```

## 配置管理

### 工作流配置
```go
type FlowConfig struct {
    Name           string                 `yaml:"name"`
    Type           string                 `yaml:"type"`           // basic/identity/code_execution
    MaxRetries     int                    `yaml:"max_retries"`
    Timeout        time.Duration          `yaml:"timeout"`
    EnableDebug    bool                   `yaml:"enable_debug"`
    EnableMonitor  bool                   `yaml:"enable_monitor"`
    Processors     []ProcessorConfig      `yaml:"processors"`
    CodeExecutor   *CodeExecutorConfig    `yaml:"code_executor,omitempty"`
    RAGConfig      *RAGConfig            `yaml:"rag_config,omitempty"`
    CustomConfig   map[string]interface{} `yaml:"custom_config,omitempty"`
}

type ProcessorConfig struct {
    Name   string                 `yaml:"name"`
    Type   string                 `yaml:"type"`
    Config map[string]interface{} `yaml:"config"`
}

type CodeExecutorConfig struct {
    Type    string `yaml:"type"`    // local/docker/cloud
    Runtime string `yaml:"runtime"` // python/go/nodejs
    Sandbox bool   `yaml:"sandbox"`
}

func LoadFlowConfig(configPath string) (*FlowConfig, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config FlowConfig
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### 配置文件示例
```yaml
# flow_config.yaml
name: "my_custom_flow"
type: "basic"
max_retries: 3
timeout: "30s"
enable_debug: true
enable_monitor: true

processors:
  - name: "instructions_processor"
    type: "instructions"
    config:
      template: "请按照以下指令执行: {query}"
  
  - name: "contents_processor"
    type: "contents"
    config:
      max_length: 4000

code_executor:
  type: "local"
  runtime: "python"
  sandbox: true

custom_config:
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 2000
```

## 最佳实践

1. **工作流选择**: 根据任务复杂度选择合适的工作流类型
2. **处理器顺序**: 合理安排请求处理器的执行顺序
3. **错误处理**: 在每个处理步骤中实现适当的错误处理
4. **监控调试**: 在生产环境中启用监控，开发环境中启用调试
5. **配置管理**: 使用配置文件管理复杂的工作流参数
6. **性能优化**: 监控工作流执行时间，优化性能瓶颈

## 依赖模块

- `github.com/nvcnvn/adk-golang/pkg/code_executors`: 代码执行器
- `github.com/nvcnvn/adk-golang/pkg/flows/llm_flows`: LLM工作流实现

## 扩展开发

### 自定义处理器
```go
type CustomProcessor struct {
    config map[string]interface{}
}

func NewCustomProcessor(config map[string]interface{}) *CustomProcessor {
    return &CustomProcessor{config: config}
}

func (p *CustomProcessor) Process(ctx context.Context, request *FlowRequest) (*FlowRequest, error) {
    // 实现自定义处理逻辑
    processedRequest := &FlowRequest{
        Query:   p.transformQuery(request.Query),
        Context: p.enhanceContext(request.Context),
    }
    
    return processedRequest, nil
}

func (p *CustomProcessor) transformQuery(query string) string {
    // 查询转换逻辑
    return query
}

func (p *CustomProcessor) enhanceContext(ctx map[string]interface{}) map[string]interface{} {
    // 上下文增强逻辑
    return ctx
}
```

Flows 模块为 ADK-Golang 框架提供了强大而灵活的工作流执行能力，是构建复杂智能体系统的核心基础设施。
