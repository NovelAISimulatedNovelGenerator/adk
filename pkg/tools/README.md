# Tools 模块文档

## 模块概述

`github.com/nvcnvn/adk-golang/pkg/tools` 是 ADK 系统的工具抽象层，提供了统一的工具接口和丰富的工具实现。该模块使智能体能够调用各种外部服务、API 和功能，极大扩展了智能体的能力边界。

## 核心架构

### 工具接口体系

```
Tool (核心接口)
├── BaseTool (基础实现)
├── FunctionTool (函数包装工具)
├── LlmToolAdaptor (LLM工具适配器)
├── AgentTool (智能体工具)
├── LongRunningTool (长时间运行工具)
└── 各种专用工具实现
```

### 专用工具类别

```
内置工具
├── BuiltInCodeExecution (代码执行)
├── ExitLoop (退出循环)
├── TransferToAgent (智能体转移)
└── GoogleSearch (Google搜索)

API集成工具
├── OpenAPI工具
├── Google API工具
├── APIHub工具
└── Application Integration工具

高级工具
├── MCP工具 (Model Context Protocol)
├── Retrieval工具 (检索)
└── Vertex AI搜索
```

## 核心接口

### 1. Tool 接口

所有工具的基础接口：

```go
type Tool interface {
    // 工具名称
    Name() string
    
    // 工具描述
    Description() string
    
    // 执行工具
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
    
    // 获取工具Schema
    Schema() ToolSchema
}
```

**核心方法说明：**
- `Name()`: 返回工具的唯一标识名称
- `Description()`: 返回工具功能的人类可读描述
- `Execute()`: 执行工具逻辑，处理输入并返回结果
- `Schema()`: 返回工具的输入输出Schema定义

### 2. ToolSchema 结构体

定义工具的输入输出规范：

```go
type ToolSchema struct {
    Input  ParameterSchema            `json:"input"`
    Output map[string]ParameterSchema `json:"output"`
}

type ParameterSchema struct {
    Type        string                     `json:"type"`
    Description string                     `json:"description"`
    Required    bool                       `json:"required,omitempty"`
    Properties  map[string]ParameterSchema `json:"properties,omitempty"`
}
```

**特性：**
- JSON Schema 兼容
- 支持嵌套对象
- 必需字段标记
- 类型验证支持

## 核心实现

### 1. BaseTool（基础工具）

提供 Tool 接口的基础实现：

```go
type BaseTool struct {
    name        string
    description string
    schema      ToolSchema
    executeFn   func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}
```

**使用示例：**
```go
tool := NewTool(
    "calculator",
    "执行数学计算",
    ToolSchema{
        Input: ParameterSchema{
            Type: "object",
            Properties: map[string]ParameterSchema{
                "expression": {
                    Type:        "string",
                    Description: "要计算的数学表达式",
                    Required:    true,
                },
            },
        },
    },
    func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        expr := input["expression"].(string)
        // 计算逻辑...
        return map[string]interface{}{"result": result}, nil
    },
)
```

### 2. FunctionTool（函数工具）

将 Go 函数包装成工具的高级实现：

```go
type FunctionTool struct {
    *BaseTool
    function     interface{}
    takesCtx     bool
    takesToolCtx bool
}
```

**配置结构：**
```go
type FunctionToolConfig struct {
    Name          string
    Description   string
    InputSchema   ParameterSchema
    OutputSchema  map[string]ParameterSchema
    IsLongRunning bool
}
```

**使用示例：**
```go
// 定义函数
func Add(a, b int) int {
    return a + b
}

// 包装为工具
tool, err := NewFunctionTool(Add, &FunctionToolConfig{
    Name:        "add",
    Description: "计算两个数的和",
    InputSchema: ParameterSchema{
        Type: "object",
        Properties: map[string]ParameterSchema{
            "a": {Type: "integer", Description: "第一个数"},
            "b": {Type: "integer", Description: "第二个数"},
        },
    },
})
```

**支持的函数签名：**
- `func(...) result`
- `func(...) (result, error)`
- `func(context.Context, ...) ...`
- `func(*ToolContext, ...) ...`

### 3. LlmToolAdaptor（LLM工具适配器）

为工具添加 LLM 特定功能的适配器：

```go
type LlmToolAdaptor struct {
    tool                      Tool
    isLongRunning            bool
    processLlmRequestFunc    func(ctx context.Context, toolContext *ToolContext, llmRequest *models.LlmRequest) error
}
```

**核心功能：**
- 包装现有工具
- 添加 LLM 请求预处理
- 支持函数调用执行
- 长时间运行标记

### 4. ToolContext（工具上下文）

提供工具执行时的上下文信息：

```go
type ToolContext struct {
    InvocationContext InvocationContext
    EventActions      *events.EventActions
}
```

**上下文信息包括：**
- 调用ID和智能体名称
- 事件动作管理
- 转录缓存
- 调用结束标志

## 内置工具实现

### 1. BuiltInCodeExecution（代码执行工具）

支持在安全环境中执行代码：

**特性：**
- 多语言支持
- 安全沙箱执行
- 结果捕获和返回
- 错误处理

**文件：** `built_in_code_execution.go`

### 2. GoogleSearch（Google搜索工具）

集成 Google 搜索功能：

**特性：**
- 网页搜索
- 结果格式化
- API 密钥管理
- 搜索结果过滤

**文件：** `google_search.go`

### 3. ExitLoop（退出循环工具）

用于循环智能体的退出控制：

**特性：**
- 循环条件检查
- 优雅退出机制
- 状态传递

**文件：** `exit_loop.go`

### 4. TransferToAgent（智能体转移工具）

在智能体之间转移执行控制：

**特性：**
- 智能体查找
- 上下文传递
- 执行转移

**文件：** `transfer_to_agent.go`

### 5. AgentTool（智能体工具）

将其他智能体包装成工具：

**特性：**
- 智能体封装
- 工具化接口
- 结果处理

**文件：** `agent_tool.go`

## 高级工具集成

### 1. OpenAPI 工具

支持通过 OpenAPI 规范动态创建工具：

**特性：**
- OpenAPI 3.0 支持
- 自动Schema生成
- HTTP 客户端集成
- 认证支持

**目录：** `openapi_tool/`

### 2. Google API 工具

集成 Google Cloud 服务：

**特性：**
- Google Cloud API 集成
- 认证管理
- 服务发现
- 批量操作

**目录：** `google_api_tool/`

### 3. MCP 工具

Model Context Protocol 工具集成：

**特性：**
- MCP 协议支持
- 上下文管理
- 远程工具调用
- 协议适配

**目录：** `mcp_tool/`

### 4. Retrieval 工具

文档检索和知识库工具：

**特性：**
- 向量搜索
- 文档索引
- 相似度匹配
- 结果排序

**目录：** `retrieval/`

### 5. Application Integration 工具

应用程序集成工具：

**特性：**
- 第三方应用集成
- Webhook 支持
- 数据转换
- 工作流集成

**目录：** `application_integration_tool/`

### 6. APIHub 工具

API 中心集成：

**特性：**
- API 目录
- 动态调用
- 版本管理
- 监控统计

**目录：** `apihub_tool/`

### 7. Vertex AI 搜索

Google Vertex AI 搜索集成：

**特性：**
- 企业搜索
- AI 增强搜索
- 多模态搜索
- 结果优化

**文件：** `vertex_ai_search.go`

## 长时间运行工具

### LongRunningTool

支持长时间执行的工具抽象：

```go
type LongRunningTool interface {
    Tool
    IsLongRunning() bool
    Cancel(ctx context.Context) error
    Status(ctx context.Context) (string, error)
}
```

**特性：**
- 异步执行支持
- 状态查询
- 取消操作
- 进度跟踪

**文件：** `long_running_tool.go`

## 使用示例

### 创建简单工具

```go
// 使用 BaseTool
tool := NewTool(
    "weather",
    "获取天气信息",
    ToolSchema{
        Input: ParameterSchema{
            Type: "object",
            Properties: map[string]ParameterSchema{
                "city": {
                    Type:        "string",
                    Description: "城市名称",
                    Required:    true,
                },
            },
        },
    },
    func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        city := input["city"].(string)
        // 获取天气信息的逻辑...
        return map[string]interface{}{
            "temperature": 25,
            "condition":   "晴天",
        }, nil
    },
)
```

### 使用 FunctionTool

```go
// 定义函数
func GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
    // 获取用户信息逻辑...
    return &UserInfo{Name: "张三", Age: 30}, nil
}

// 包装为工具
tool, err := NewFunctionTool(GetUserInfo, &FunctionToolConfig{
    Name:        "get_user_info",
    Description: "获取用户信息",
    InputSchema: ParameterSchema{
        Type: "object",
        Properties: map[string]ParameterSchema{
            "userID": {
                Type:        "string",
                Description: "用户ID",
                Required:    true,
            },
        },
    },
})
```

### 在智能体中使用工具

```go
// 创建智能体并添加工具
agent := agents.NewAgent(
    agents.WithName("助手"),
    agents.WithTools(weatherTool, userInfoTool),
)

// 处理包含工具调用的请求
response, err := agent.Process(ctx, "请告诉我北京的天气")
```

### 使用 LlmToolAdaptor

```go
// 创建适配器
adaptor := NewLlmToolAdaptor(baseTool, false)

// 设置 LLM 请求预处理
adaptor.SetProcessLlmRequestFunc(func(ctx context.Context, toolContext *ToolContext, llmRequest *models.LlmRequest) error {
    // 自定义预处理逻辑
    return nil
})

// 执行函数调用
result, err := adaptor.ExecuteFunctionCall(ctx, toolContext, functionCall)
```

## 工具注册与发现

### 工具注册中心

虽然当前实现中没有显式的工具注册中心，但可以通过以下方式管理工具：

```go
// 工具集合管理
type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(name string, tool Tool) {
    r.tools[name] = tool
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
    tool, exists := r.tools[name]
    return tool, exists
}
```

## 架构优势

### 1. 统一抽象
- 一致的工具接口
- 标准化的输入输出
- 统一的错误处理

### 2. 灵活扩展
- 插件化架构
- 函数包装能力
- 适配器模式支持

### 3. 丰富的集成
- 多种 API 集成
- 云服务支持
- 第三方应用连接

### 4. 企业级特性
- 长时间运行支持
- 错误恢复机制
- 监控和日志

## 与其他模块的集成

- **agents**: 为智能体提供工具调用能力
- **models**: 支持 LLM 的工具调用功能
- **events**: 工具执行事件处理
- **auth**: 工具认证和授权
- **config**: 工具配置管理

## 最佳实践

### 1. 工具设计原则
- 单一职责：每个工具专注特定功能
- 幂等性：重复调用产生相同结果
- 错误处理：提供清晰的错误信息

### 2. Schema 设计
- 详细的参数描述
- 合理的类型定义
- 必需字段标记

### 3. 性能优化
- 避免阻塞操作
- 合理使用缓存
- 异步处理长时间任务

### 4. 安全考虑
- 输入验证和清理
- 权限检查
- 敏感信息保护

## 开发状态

- ✅ 核心工具接口设计
- ✅ BaseTool 基础实现
- ✅ FunctionTool 函数包装
- ✅ LlmToolAdaptor 适配器
- ✅ 内置工具实现
- ✅ Google API 集成
- ✅ OpenAPI 工具支持
- ✅ MCP 协议支持
- ✅ 检索工具集成
- ✅ 长时间运行工具
- ✅ 应用集成工具

该模块为 ADK 系统提供了强大的工具生态系统，使智能体能够与外部世界进行丰富的交互，是实现智能体实际应用价值的关键组件。
