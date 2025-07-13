# Types 类型定义模块

## 概述

Types 模块提供 ADK 框架中跨包共享的核心类型定义。该模块包含流式处理模式、运行配置、转录条目、调用上下文数据和事件动作等重要类型，是整个框架的基础类型系统。

## 核心类型

### 1. StreamingMode 流式模式

定义响应的流式处理方式：

```go
type StreamingMode string

const (
    // StreamingModeNone 无流式处理
    StreamingModeNone StreamingMode = "none"
    
    // StreamingModeSSE 服务器发送事件流式处理
    StreamingModeSSE StreamingMode = "sse"
)
```

**用途:**
- 控制智能体响应的输出方式
- 支持实时流式输出和批量输出模式
- 用于配置客户端-服务器通信模式

### 2. RunConfig 运行配置

智能体调用的配置参数：

```go
type RunConfig struct {
    // StreamingMode 控制响应如何流式传输
    StreamingMode StreamingMode `json:"streamingMode,omitempty"`
    
    // MaxLlmCalls 限制 LLM 调用次数
    MaxLlmCalls int `json:"maxLlmCalls,omitempty"`
    
    // SupportCFC 表示是否支持客户端函数调用 (CFC)
    SupportCFC bool `json:"supportCfc,omitempty"`
}
```

**字段说明:**
- `StreamingMode`: 设置响应流式模式
- `MaxLlmCalls`: 限制单次调用中的 LLM 请求数量，防止无限循环
- `SupportCFC`: 启用客户端函数调用能力

### 3. TranscriptionEntry 转录条目

音频转录的条目结构：

```go
type TranscriptionEntry struct {
    // Role 转录的角色 (user 或 model)
    Role string `json:"role"`
    
    // Text 转录文本内容
    Text string `json:"text"`
}
```

**用途:**
- 存储音频转录结果
- 区分用户和模型的语音内容
- 支持语音对话的上下文管理

### 4. InvocationContextData 调用上下文数据

智能体调用过程中共享的上下文信息：

```go
type InvocationContextData struct {
    // InvocationID 调用唯一标识符
    InvocationID string `json:"invocationId"`
    
    // RunConfig 运行配置
    RunConfig *RunConfig `json:"runConfig,omitempty"`
    
    // EndInvocation 是否结束调用
    EndInvocation bool `json:"endInvocation,omitempty"`
    
    // Branch 分支标识
    Branch string `json:"branch,omitempty"`
    
    // TranscriptionCache 转录缓存 (不序列化)
    TranscriptionCache []TranscriptionEntry `json:"-"`
    
    // 私有字段
    llmCallCount int        // LLM 调用计数
    mu           sync.Mutex // 并发保护
}
```

**核心方法:**
```go
// IncrementLlmCallCount 增加并检查 LLM 调用计数
func (icd *InvocationContextData) IncrementLlmCallCount() error

// GetLlmCallCount 获取当前 LLM 调用计数
func (icd *InvocationContextData) GetLlmCallCount() int
```

### 5. EventActionsData 事件动作数据

跨包共享的事件动作信息：

```go
type EventActionsData struct {
    // TransferToAgent 转移到指定智能体
    TransferToAgent string `json:"transferToAgent,omitempty"`
}
```

**用途:**
- 控制智能体间的转移和切换
- 支持多智能体协作场景
- 实现智能体路由功能

## 使用示例

### 基本配置使用
```go
package main

import (
    "context"
    "fmt"
    
    "github.com/nvcnvn/adk-golang/pkg/types"
)

func main() {
    // 创建运行配置
    config := &types.RunConfig{
        StreamingMode: types.StreamingModeSSE,
        MaxLlmCalls:   10,
        SupportCFC:    true,
    }
    
    // 创建调用上下文数据
    contextData := &types.InvocationContextData{
        InvocationID: "inv-12345",
        RunConfig:    config,
        Branch:       "main",
    }
    
    fmt.Printf("调用ID: %s\n", contextData.InvocationID)
    fmt.Printf("流式模式: %s\n", config.StreamingMode)
    fmt.Printf("最大LLM调用次数: %d\n", config.MaxLlmCalls)
}
```

### LLM 调用计数管理
```go
func demonstrateLlmCallCounting() {
    contextData := &types.InvocationContextData{
        InvocationID: "inv-67890",
        RunConfig: &types.RunConfig{
            MaxLlmCalls: 5,
        },
    }
    
    // 模拟多次 LLM 调用
    for i := 0; i < 7; i++ {
        err := contextData.IncrementLlmCallCount()
        if err != nil {
            fmt.Printf("LLM 调用超限: %v\n", err)
            break
        }
        
        fmt.Printf("LLM 调用 %d/%d\n", 
            contextData.GetLlmCallCount(), 
            contextData.RunConfig.MaxLlmCalls)
    }
}
```

### 转录缓存管理
```go
func manageTranscriptionCache() {
    contextData := &types.InvocationContextData{
        InvocationID: "inv-audio-001",
        TranscriptionCache: []types.TranscriptionEntry{},
    }
    
    // 添加用户转录
    userEntry := types.TranscriptionEntry{
        Role: "user",
        Text: "请帮我分析这个数据",
    }
    contextData.TranscriptionCache = append(contextData.TranscriptionCache, userEntry)
    
    // 添加模型转录
    modelEntry := types.TranscriptionEntry{
        Role: "model", 
        Text: "好的，我来为您分析数据内容",
    }
    contextData.TranscriptionCache = append(contextData.TranscriptionCache, modelEntry)
    
    // 输出转录历史
    fmt.Println("转录历史:")
    for i, entry := range contextData.TranscriptionCache {
        fmt.Printf("%d. [%s]: %s\n", i+1, entry.Role, entry.Text)
    }
}
```

### 智能体转移
```go
func demonstrateAgentTransfer() {
    eventData := &types.EventActionsData{
        TransferToAgent: "specialist-agent",
    }
    
    if eventData.TransferToAgent != "" {
        fmt.Printf("转移到智能体: %s\n", eventData.TransferToAgent)
        // 执行智能体转移逻辑
        transferToAgent(eventData.TransferToAgent)
    }
}

func transferToAgent(agentName string) {
    fmt.Printf("正在转移到智能体: %s\n", agentName)
    // 实际的转移逻辑
}
```

## 高级用法

### 上下文数据包装器
```go
type ContextWrapper struct {
    *types.InvocationContextData
    startTime    time.Time
    metadata     map[string]interface{}
}

func NewContextWrapper(invocationID string, config *types.RunConfig) *ContextWrapper {
    return &ContextWrapper{
        InvocationContextData: &types.InvocationContextData{
            InvocationID: invocationID,
            RunConfig:    config,
        },
        startTime: time.Now(),
        metadata:  make(map[string]interface{}),
    }
}

func (cw *ContextWrapper) SetMetadata(key string, value interface{}) {
    cw.metadata[key] = value
}

func (cw *ContextWrapper) GetMetadata(key string) interface{} {
    return cw.metadata[key]
}

func (cw *ContextWrapper) GetDuration() time.Duration {
    return time.Since(cw.startTime)
}

func (cw *ContextWrapper) IsExpired(timeout time.Duration) bool {
    return cw.GetDuration() > timeout
}
```

### 配置验证器
```go
type ConfigValidator struct{}

func NewConfigValidator() *ConfigValidator {
    return &ConfigValidator{}
}

func (cv *ConfigValidator) ValidateRunConfig(config *types.RunConfig) error {
    if config == nil {
        return fmt.Errorf("运行配置不能为空")
    }
    
    // 验证流式模式
    if config.StreamingMode != types.StreamingModeNone && 
       config.StreamingMode != types.StreamingModeSSE {
        return fmt.Errorf("无效的流式模式: %s", config.StreamingMode)
    }
    
    // 验证 LLM 调用限制
    if config.MaxLlmCalls < 0 {
        return fmt.Errorf("最大LLM调用次数不能为负数: %d", config.MaxLlmCalls)
    }
    
    if config.MaxLlmCalls > 100 {
        return fmt.Errorf("最大LLM调用次数过大: %d (最大: 100)", config.MaxLlmCalls)
    }
    
    return nil
}

func (cv *ConfigValidator) ValidateInvocationContext(ctx *types.InvocationContextData) error {
    if ctx == nil {
        return fmt.Errorf("调用上下文不能为空")
    }
    
    if ctx.InvocationID == "" {
        return fmt.Errorf("调用ID不能为空")
    }
    
    if ctx.RunConfig != nil {
        return cv.ValidateRunConfig(ctx.RunConfig)
    }
    
    return nil
}

// 使用示例
func validateConfiguration() {
    validator := NewConfigValidator()
    
    config := &types.RunConfig{
        StreamingMode: types.StreamingModeSSE,
        MaxLlmCalls:   15,
        SupportCFC:    true,
    }
    
    if err := validator.ValidateRunConfig(config); err != nil {
        fmt.Printf("配置验证失败: %v\n", err)
        return
    }
    
    fmt.Println("配置验证成功")
}
```

### 转录管理器
```go
type TranscriptionManager struct {
    entries []types.TranscriptionEntry
    maxSize int
    mu      sync.RWMutex
}

func NewTranscriptionManager(maxSize int) *TranscriptionManager {
    return &TranscriptionManager{
        entries: make([]types.TranscriptionEntry, 0),
        maxSize: maxSize,
    }
}

func (tm *TranscriptionManager) AddEntry(role, text string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    entry := types.TranscriptionEntry{
        Role: role,
        Text: text,
    }
    
    tm.entries = append(tm.entries, entry)
    
    // 保持缓存大小限制
    if len(tm.entries) > tm.maxSize {
        tm.entries = tm.entries[1:] // 移除最旧的条目
    }
}

func (tm *TranscriptionManager) GetEntries() []types.TranscriptionEntry {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    // 返回副本以避免并发修改
    entries := make([]types.TranscriptionEntry, len(tm.entries))
    copy(entries, tm.entries)
    return entries
}

func (tm *TranscriptionManager) GetUserEntries() []types.TranscriptionEntry {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    var userEntries []types.TranscriptionEntry
    for _, entry := range tm.entries {
        if entry.Role == "user" {
            userEntries = append(userEntries, entry)
        }
    }
    return userEntries
}

func (tm *TranscriptionManager) GetModelEntries() []types.TranscriptionEntry {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    var modelEntries []types.TranscriptionEntry
    for _, entry := range tm.entries {
        if entry.Role == "model" {
            modelEntries = append(modelEntries, entry)
        }
    }
    return modelEntries
}

func (tm *TranscriptionManager) Clear() {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    tm.entries = tm.entries[:0]
}

func (tm *TranscriptionManager) Size() int {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    return len(tm.entries)
}
```

### 事件路由器
```go
type EventRouter struct {
    handlers map[string]func(*types.EventActionsData) error
    mu       sync.RWMutex
}

func NewEventRouter() *EventRouter {
    return &EventRouter{
        handlers: make(map[string]func(*types.EventActionsData) error),
    }
}

func (er *EventRouter) RegisterHandler(eventType string, handler func(*types.EventActionsData) error) {
    er.mu.Lock()
    defer er.mu.Unlock()
    
    er.handlers[eventType] = handler
}

func (er *EventRouter) ProcessEvent(eventType string, eventData *types.EventActionsData) error {
    er.mu.RLock()
    handler, exists := er.handlers[eventType]
    er.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("未找到事件类型 '%s' 的处理器", eventType)
    }
    
    return handler(eventData)
}

// 使用示例
func setupEventRouting() {
    router := NewEventRouter()
    
    // 注册智能体转移处理器
    router.RegisterHandler("agent_transfer", func(data *types.EventActionsData) error {
        if data.TransferToAgent == "" {
            return fmt.Errorf("转移目标智能体不能为空")
        }
        
        fmt.Printf("执行智能体转移到: %s\n", data.TransferToAgent)
        return nil
    })
    
    // 处理事件
    eventData := &types.EventActionsData{
        TransferToAgent: "customer-service-agent",
    }
    
    if err := router.ProcessEvent("agent_transfer", eventData); err != nil {
        fmt.Printf("事件处理失败: %v\n", err)
    }
}
```

## 并发安全

### InvocationContextData 的线程安全
```go
func demonstrateConcurrentAccess() {
    contextData := &types.InvocationContextData{
        InvocationID: "concurrent-test",
        RunConfig: &types.RunConfig{
            MaxLlmCalls: 100,
        },
    }
    
    var wg sync.WaitGroup
    
    // 启动多个 goroutine 并发调用
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < 5; j++ {
                err := contextData.IncrementLlmCallCount()
                if err != nil {
                    fmt.Printf("Goroutine %d: LLM调用失败: %v\n", id, err)
                    return
                }
                
                count := contextData.GetLlmCallCount()
                fmt.Printf("Goroutine %d: 当前调用计数: %d\n", id, count)
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Printf("最终调用计数: %d\n", contextData.GetLlmCallCount())
}
```

## 最佳实践

### 1. 配置管理
```go
// 创建默认配置
func NewDefaultRunConfig() *types.RunConfig {
    return &types.RunConfig{
        StreamingMode: types.StreamingModeNone,
        MaxLlmCalls:   20,
        SupportCFC:    false,
    }
}

// 创建流式配置
func NewStreamingRunConfig() *types.RunConfig {
    return &types.RunConfig{
        StreamingMode: types.StreamingModeSSE,
        MaxLlmCalls:   50,
        SupportCFC:    true,
    }
}
```

### 2. 调用上下文初始化
```go
func NewInvocationContext(invocationID string, config *types.RunConfig) *types.InvocationContextData {
    if config == nil {
        config = NewDefaultRunConfig()
    }
    
    return &types.InvocationContextData{
        InvocationID:       invocationID,
        RunConfig:          config,
        TranscriptionCache: make([]types.TranscriptionEntry, 0),
    }
}
```

### 3. 错误处理
```go
func handleLlmCallLimit(contextData *types.InvocationContextData) error {
    err := contextData.IncrementLlmCallCount()
    if err != nil {
        // 记录错误并采取措施
        fmt.Printf("LLM调用限制达到: %v\n", err)
        
        // 设置结束标志
        contextData.EndInvocation = true
        
        return fmt.Errorf("智能体调用终止: %w", err)
    }
    
    return nil
}
```

### 4. 资源清理
```go
func cleanupInvocationContext(contextData *types.InvocationContextData) {
    // 清理转录缓存
    contextData.TranscriptionCache = nil
    
    // 设置结束标志
    contextData.EndInvocation = true
    
    fmt.Printf("已清理调用上下文: %s\n", contextData.InvocationID)
}
```

## 依赖模块

- Go 标准库: `fmt`, `sync`
- 集成框架: 与 ADK 其他模块紧密配合使用

## 扩展开发

### 自定义流式模式
```go
const (
    // 扩展自定义流式模式
    StreamingModeWebSocket types.StreamingMode = "websocket"
    StreamingModeGRPC      types.StreamingMode = "grpc"
)

func validateCustomStreamingMode(mode types.StreamingMode) bool {
    switch mode {
    case types.StreamingModeNone, types.StreamingModeSSE,
         StreamingModeWebSocket, StreamingModeGRPC:
        return true
    default:
        return false
    }
}
```

### 扩展配置字段
```go
type ExtendedRunConfig struct {
    *types.RunConfig
    
    // 扩展字段
    Timeout       time.Duration `json:"timeout,omitempty"`
    RetryAttempts int           `json:"retryAttempts,omitempty"`
    UseCache      bool          `json:"useCache,omitempty"`
}

func NewExtendedRunConfig() *ExtendedRunConfig {
    return &ExtendedRunConfig{
        RunConfig: &types.RunConfig{
            StreamingMode: types.StreamingModeSSE,
            MaxLlmCalls:   30,
            SupportCFC:    true,
        },
        Timeout:       30 * time.Second,
        RetryAttempts: 3,
        UseCache:      true,
    }
}
```

Types 模块为 ADK-Golang 框架提供了完整的类型定义体系，确保跨模块的数据结构一致性和类型安全。通过合理使用这些类型，可以构建健壮、高效的智能体应用系统。
