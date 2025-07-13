# Models 模块文档

## 模块概述

`github.com/nvcnvn/adk-golang/pkg/models` 是 ADK 系统的模型抽象层，提供了统一的 LLM（大语言模型）接口和实现体系。该模块支持多种 LLM 提供商，包括 Gemini、DeepSeek 等，并提供了模型池、注册中心等高级功能。

## 核心架构

### 模型接口体系

```
LLM (新接口)
├── BaseLlm (基础实现)
├── GeminiLLM (Gemini 实现)
├── DeepSeekLLM (DeepSeek 实现)  
└── ModelToLLMAdapter (适配器)

Model (旧接口) - 向后兼容
├── PoolModel (连接池模型)
└── 各种具体实现
```

### 工厂与注册体系

```
UnifiedModelFactory (统一工厂)
├── ModelRegistry (标准注册中心)
└── EnhancedRegistry (增强注册中心，支持正则匹配)
```

## 核心组件

### 1. LLM 接口（新一代接口）

统一的大语言模型接口，支持现代 LLM 功能：

```go
type LLM interface {
    // 返回支持的模型正则表达式列表
    SupportedModels() []string
    
    // 生成内容（单次请求）
    GenerateContent(ctx context.Context, request *LlmRequest) (*LlmResponse, error)
    
    // 生成流式内容
    GenerateContentStream(ctx context.Context, request *LlmRequest) (<-chan *LlmResponse, error)
    
    // 建立实时双向连接
    Connect(ctx context.Context, request *LlmRequest) (LlmConnection, error)
}
```

**特性：**
- 支持流式和非流式生成
- 支持实时双向连接
- 支持工具调用
- 支持复杂的请求配置

### 2. LlmRequest 结构体

LLM 请求封装，支持丰富的配置选项：

```go
type LlmRequest struct {
    // 对话内容
    Contents *Content `json:"contents,omitempty"`
    
    // 可用工具列表
    Tools []*Tool `json:"tools,omitempty"`
    
    // 工具字典（用于快速查找）
    ToolsDict map[string]*Tool `json:"-"`
    
    // 系统指令
    SystemInstructions string `json:"systemInstructions,omitempty"`
    
    // 生成参数
    Temperature     float64 `json:"temperature,omitempty"`
    TopP           float64 `json:"topP,omitempty"`
    TopK           int     `json:"topK,omitempty"`
    MaxTokens      int     `json:"maxTokens,omitempty"`
    CandidateCount int     `json:"candidateCount,omitempty"`
}
```

### 3. LlmResponse 结构体

LLM 响应封装，支持完整的响应状态：

```go
type LlmResponse struct {
    // 响应内容
    Content *Content `json:"content,omitempty"`
    
    // 是否为部分响应（流式）
    Partial bool `json:"partial,omitempty"`
    
    // 错误信息
    ErrorCode    string `json:"errorCode,omitempty"`
    ErrorMessage string `json:"errorMessage,omitempty"`
    
    // 状态标志
    Interrupted  bool `json:"interrupted,omitempty"`
    TurnComplete bool `json:"turnComplete,omitempty"`
}
```

### 4. 内容结构体系

#### Content（内容容器）
```go
type Content struct {
    Parts []*Part `json:"parts,omitempty"`
}
```

#### Part（内容部分）
支持多种内容类型：
```go
type Part struct {
    // 文本内容
    Text string `json:"text,omitempty"`
    Role string `json:"role,omitempty"`
    
    // 函数调用
    FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
    FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
    
    // 认证请求
    AuthRequest *AuthRequest `json:"authRequest,omitempty"`
    
    // 思考标记
    Thought bool `json:"thought,omitempty"`
}
```

#### FunctionCall（函数调用）
```go
type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments,omitempty"`
    ID        string `json:"id,omitempty"`
}
```

#### FunctionResponse（函数响应）
```go
type FunctionResponse struct {
    Name        string       `json:"name"`
    Content     string       `json:"content,omitempty"`
    ID          string       `json:"id,omitempty"`
    AuthRequest *AuthRequest `json:"authRequest,omitempty"`
}
```

### 5. 工具系统

#### Tool（工具定义）
```go
type Tool struct {
    Name          string                 `json:"name"`
    Description   string                 `json:"description,omitempty"`
    InputSchema   map[string]interface{} `json:"inputSchema,omitempty"`
    IsLongRunning bool                   `json:"isLongRunning,omitempty"`
}
```

### 6. 连接接口

#### LlmConnection（实时连接）
```go
type LlmConnection interface {
    // 发送消息
    Send(ctx context.Context, content Content) error
    
    // 接收响应
    Receive(ctx context.Context) (*LlmResponse, error)
    
    // 关闭连接
    Close() error
}
```

## 具体实现

### 1. BaseLlm（基础实现）

提供 LLM 接口的基础实现框架：

```go
type BaseLlm struct {
    ModelName string
}
```

**功能：**
- 提供默认的接口实现
- 子类可重写特定方法
- 统一的模型名称管理

### 2. GeminiLLM（Gemini 实现）

Google Gemini 模型的完整实现：

**特性：**
- 支持 Gemini 系列模型
- 支持流式生成
- 支持实时连接
- 支持工具调用
- 支持思考模式（ThinkingConfig）

**文件：**
- `gemini.go` - 主要实现
- `gemini_connection.go` - 连接实现

### 3. DeepSeekLLM（DeepSeek 实现）

DeepSeek 模型的实现：

**特性：**
- 支持 DeepSeek 系列模型
- API 调用优化
- 错误处理机制

**文件：**
- `deepseek.go`

### 4. PoolModel（连接池模型）

提供模型连接池功能，优化资源使用：

**特性：**
- 连接复用
- 资源管理
- 性能优化

**文件：**
- `pool_model.go`

## 工厂与注册系统

### 1. UnifiedModelFactory（统一工厂）

提供统一的模型创建接口：

```go
type UnifiedModelFactory struct {
    standardRegistry *ModelRegistry
    enhancedRegistry *EnhancedRegistry
}
```

**核心方法：**
- `GetModel(modelName)` - 获取模型实例
- `GetLLM(modelName)` - 获取 LLM 实例

**特性：**
- 双注册中心支持
- 优先级机制
- 自动适配

### 2. ModelRegistry（标准注册中心）

传统的模型注册中心：

**特性：**
- 精确匹配
- 简单高效
- 向后兼容

### 3. EnhancedRegistry（增强注册中心）

支持正则表达式匹配的注册中心：

**特性：**
- 正则表达式匹配
- 模式识别
- 灵活配置

### 4. ModelToLLMAdapter（适配器）

将旧的 Model 接口适配到新的 LLM 接口：

```go
type ModelToLLMAdapter struct {
    model Model
}
```

**功能：**
- 接口兼容
- 功能桥接
- 平滑迁移

## 配置系统

### ThinkingConfig（思考配置）

用于配置模型的思考模式：

```go
type ThinkingConfig struct {
    // 思考配置参数...
}
```

**用途：**
- 控制模型思考行为
- 优化推理质量
- 调整输出格式

## 使用示例

### 创建和使用 LLM

```go
// 通过工厂创建 LLM
factory := GetUnifiedModelFactory()
llm, err := factory.GetLLM("gemini-1.5-pro")
if err != nil {
    log.Fatal(err)
}

// 创建请求
request := &LlmRequest{
    Contents: &Content{
        Parts: []*Part{{
            Text: "你好，请介绍一下你自己",
            Role: "user",
        }},
    },
    SystemInstructions: "你是一个有用的助手",
    Temperature: 0.7,
    MaxTokens: 1000,
}

// 生成内容
response, err := llm.GenerateContent(context.Background(), request)
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Content.GetText())
```

### 流式生成

```go
// 流式生成内容
stream, err := llm.GenerateContentStream(context.Background(), request)
if err != nil {
    log.Fatal(err)
}

// 处理流式响应
for response := range stream {
    if response.ErrorMessage != "" {
        log.Printf("错误: %s", response.ErrorMessage)
        continue
    }
    
    if response.Content != nil {
        fmt.Print(response.Content.GetText())
    }
    
    if response.TurnComplete {
        break
    }
}
```

### 工具调用

```go
// 定义工具
tool := &Tool{
    Name:        "calculate",
    Description: "执行数学计算",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "expression": map[string]interface{}{
                "type":        "string",
                "description": "数学表达式",
            },
        },
        "required": []string{"expression"},
    },
}

// 添加工具到请求
request.Tools = []*Tool{tool}
request.ToolsDict = map[string]*Tool{
    tool.Name: tool,
}

// 生成内容并处理工具调用
response, err := llm.GenerateContent(context.Background(), request)
```

### 实时连接

```go
// 建立连接
conn, err := llm.Connect(context.Background(), request)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// 发送消息
err = conn.Send(context.Background(), Content{
    Parts: []*Part{{
        Text: "你好",
        Role: "user",
    }},
})

// 接收响应
response, err := conn.Receive(context.Background())
if err != nil {
    log.Fatal(err)
}
```

## 模型注册

### 注册新模型实现

```go
// 注册到标准注册中心
registry := GetRegistry()
registry.Register("custom-model", &CustomModel{})

// 注册到增强注册中心
enhancedRegistry := GetEnhancedRegistry()
enhancedRegistry.RegisterLLM(&CustomLLM{})
```

## 架构优势

### 1. 统一抽象
- 统一的接口设计
- 多提供商支持
- 一致的使用体验

### 2. 灵活扩展
- 插件化架构
- 自定义实现支持
- 适配器模式

### 3. 高性能
- 连接池支持
- 流式处理
- 异步操作

### 4. 完整功能
- 工具调用
- 实时连接
- 思考模式
- 错误处理

## 与其他模块的集成

- **agents**: 为智能体提供 LLM 支持
- **tools**: 集成工具调用能力
- **config**: 支持模型配置管理
- **telemetry**: 提供性能监控
- **auth**: 支持认证机制

## 最佳实践

### 1. 模型选择
- 根据任务类型选择合适模型
- 考虑性能和成本平衡
- 使用工厂模式创建实例

### 2. 参数调优
- 合理设置 Temperature 和 TopP
- 控制 MaxTokens 避免超限
- 根据需求调整 CandidateCount

### 3. 错误处理
- 检查响应中的错误码
- 实现重试机制
- 处理网络异常

### 4. 性能优化
- 使用连接池
- 批量处理请求
- 缓存模型实例

## 开发状态

- ✅ 核心 LLM 接口设计
- ✅ Gemini 模型实现
- ✅ DeepSeek 模型实现
- ✅ 统一工厂模式
- ✅ 注册中心体系
- ✅ 工具调用支持
- ✅ 流式生成支持
- ✅ 实时连接支持
- ✅ 适配器模式
- ✅ 连接池实现

该模块为 ADK 系统提供了强大而灵活的 LLM 抽象层，支持多种主流模型提供商，是系统的核心基础组件之一。
