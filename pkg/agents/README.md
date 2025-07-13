# Agents 模块文档

## 模块概述

`github.com/nvcnvn/adk-golang/pkg/agents` 是 ADK 系统的核心智能体模块，提供了完整的智能体抽象和实现体系。该模块实现了灵活的智能体架构，支持串行、并行、循环和远程执行等多种智能体类型。

## 核心架构

### 智能体接口体系

```
BaseAgent (接口)
├── LlmAgent (LLM 智能体)
├── Agent (通用智能体)
├── SequentialAgent (串行智能体)
├── ParallelAgent (并行智能体)
├── LoopAgent (循环智能体)
└── RemoteAgent (远程智能体)
```

### 执行上下文体系

- **InvocationContext**: 调用上下文，封装智能体执行时的环境信息
- **CallbackContext**: 回调上下文，用于模型调用前后的钩子函数
- **RunConfig**: 运行配置，定义智能体的执行参数

## 核心组件

### 1. BaseAgent 接口

所有智能体的基础接口，定义了智能体的核心行为：

```go
type BaseAgent interface {
    Name() string
    Run(ctx context.Context, invocationContext *InvocationContext) (<-chan *events.Event, error)
    RunLive(ctx context.Context, invocationContext *InvocationContext) (<-chan *events.Event, error)
    RootAgent() BaseAgent
    FindAgent(name string) BaseAgent
}
```

**核心方法说明：**
- `Name()`: 返回智能体名称
- `Run()`: 执行智能体（批处理模式）
- `RunLive()`: 执行智能体（实时模式）
- `RootAgent()`: 获取根智能体
- `FindAgent()`: 按名称查找智能体

### 2. LlmAgent 结构体

基于 LLM 模型的专用智能体：

```go
type LlmAgent struct {
    name                 string
    SystemInstructions   string
    CanonicalModel      models.BaseLlm
    CanonicalTools      []tools.Tool
    BeforeModelCallback func(*CallbackContext, *models.LlmRequest) *models.LlmResponse
    AfterModelCallback  func(*CallbackContext, *models.LlmResponse) *models.LlmResponse
    parentAgent         BaseAgent
}
```

**特性：**
- 直接集成 LLM 模型
- 支持系统指令定制
- 工具调用能力
- 模型调用前后回调支持

### 3. Agent 结构体

通用智能体实现，支持层次化组织：

```go
type Agent struct {
    name                string
    model              string
    instruction        string
    description        string
    tools              []tools.Tool
    subAgents          []*Agent
    parentAgent        *Agent
    beforeAgentCallback BeforeAgentCallback
    afterAgentCallback  AfterAgentCallback
    registry           *agentRegistry
}
```

**核心特性：**
- 层次化智能体组织
- 子智能体支持
- 工具集成
- 处理前后回调机制
- 智能体注册管理

### 4. 专用智能体类型

#### SequentialAgent (串行智能体)
- **用途**: 智能体按顺序执行，下游智能体接收上游输出
- **特性**: 上下文传递链，适用于决策层
- **文件**: `sequential_agent.go`

#### ParallelAgent (并行智能体)
- **用途**: 智能体并行执行，各智能体独立处理
- **特性**: 
  - 可配置并发数 (`Workers` 字段)
  - 错误聚合机制 (`MultiError`)
  - 上下文取消支持
- **适用场景**: 执行层，如世界观、角色、剧情并行生成
- **文件**: `parallel_agent.go`

#### LoopAgent (循环智能体)
- **用途**: 循环执行智能体直到满足条件
- **特性**: 条件控制循环
- **文件**: `loop_agent.go`

#### RemoteAgent (远程智能体)
- **用途**: 通过网络调用远程智能体服务
- **特性**: 跨网络智能体调用
- **文件**: `remote_agent.go`

## 执行上下文系统

### InvocationContext

智能体调用上下文，封装执行环境：

```go
type InvocationContext struct {
    Message     string
    Config      *RunConfig
    // 其他上下文信息...
}
```

### RunConfig

智能体运行配置：

```go
type RunConfig struct {
    // 配置字段...
}
```

## 回调机制

### 智能体级回调

```go
type BeforeAgentCallback func(ctx context.Context, message string) (string, bool)
type AfterAgentCallback func(ctx context.Context, response string) string
```

- **BeforeAgentCallback**: 智能体处理前调用
- **AfterAgentCallback**: 智能体处理后调用

### 模型级回调

```go
BeforeModelCallback func(*CallbackContext, *models.LlmRequest) *models.LlmResponse
AfterModelCallback  func(*CallbackContext, *models.LlmResponse) *models.LlmResponse
```

- 用于 LLM 模型调用的前后处理
- 支持请求修改和响应后处理

## 智能体注册系统

### AgentRegistry

全局智能体注册中心，支持智能体的注册和检索：

```go
type agentRegistry struct {
    agents map[string]*Agent
    mu     sync.RWMutex
}
```

**主要方法：**
- `Register(name, agent)`: 注册智能体
- `Get(name)`: 获取智能体
- `Export(agent)`: 导出智能体供 CLI 使用

## 配置构建器

### Config 结构体

智能体创建配置：

```go
type Config struct {
    Name                string
    Model              string
    Instruction        string
    Description        string
    Tools              []tools.Tool
    SubAgents          []*Agent
    BeforeAgentCallback BeforeAgentCallback
    AfterAgentCallback  AfterAgentCallback
}
```

### Option 模式

支持通过 Option 函数灵活配置智能体：

```go
// 创建智能体示例
agent := NewAgent(
    WithName("示例智能体"),
    WithModel("gpt-4"),
    WithInstruction("你是一个有用的助手"),
    WithTools(tool1, tool2),
    WithSubAgents(childAgent1, childAgent2),
)
```

**可用配置项：**
- `WithName()`: 设置名称
- `WithModel()`: 设置模型
- `WithInstruction()`: 设置指令
- `WithDescription()`: 设置描述
- `WithTools()`: 设置工具
- `WithSubAgents()`: 设置子智能体
- `WithBeforeAgentCallback()`: 设置前置回调
- `WithAfterAgentCallback()`: 设置后置回调

## 智能体验证

### AgentValidator

智能体配置验证器，确保智能体配置的有效性：

- 验证必需字段
- 检查配置一致性
- 验证依赖关系

**文件**: `agent_validator.go`

## 使用示例

### 创建基础智能体

```go
agent := NewAgent(
    WithName("助手"),
    WithModel("gpt-4"),
    WithInstruction("你是一个专业的编程助手"),
    WithDescription("帮助用户解决编程问题"),
)

response, err := agent.Process(ctx, "如何使用 Go 语言？")
```

### 创建层次化智能体

```go
// 创建子智能体
codeAgent := NewAgent(
    WithName("代码生成器"),
    WithInstruction("专门生成高质量代码"),
)

testAgent := NewAgent(
    WithName("测试生成器"),
    WithInstruction("为代码生成测试用例"),
)

// 创建父智能体
mainAgent := NewAgent(
    WithName("开发助手"),
    WithSubAgents(codeAgent, testAgent),
)
```

### 创建 LLM 智能体

```go
llmAgent := NewLlmAgent("专家", model)
llmAgent.SystemInstructions = "你是领域专家"
llmAgent.CanonicalTools = []tools.Tool{tool1, tool2}
```

## 架构优势

### 1. 灵活的智能体组织
- 支持层次化结构
- 支持多种执行模式（串行、并行、循环）
- 支持远程智能体调用

### 2. 强大的扩展性
- 插件化工具系统
- 回调机制支持定制化
- 配置驱动的智能体创建

### 3. 完整的生命周期管理
- 智能体注册和发现
- 执行上下文管理
- 错误处理和恢复

### 4. 高性能执行
- 并行执行支持
- 异步事件流
- 资源管理优化

## 与其他模块的集成

- **models**: 提供 LLM 模型支持
- **tools**: 提供工具调用能力
- **events**: 提供事件流机制
- **telemetry**: 提供监控和追踪
- **memory**: 支持对话记忆功能

## 最佳实践

### 1. 智能体设计原则
- 单一职责原则：每个智能体专注特定任务
- 层次化组织：合理划分父子智能体关系
- 配置外化：通过配置而非硬编码定制行为

### 2. 性能优化
- 合理使用并行智能体提高效率
- 避免过深的智能体层次
- 适当使用缓存和批处理

### 3. 错误处理
- 实现适当的重试机制
- 使用结构化错误信息
- 提供降级处理策略

## 开发状态

- ✅ 核心智能体接口和实现
- ✅ 串行和并行执行支持
- ✅ LLM 智能体集成
- ✅ 智能体注册和发现
- ✅ 回调机制
- ✅ 配置验证
- ✅ 错误处理和恢复

该模块是 ADK 系统的核心，为上层应用提供了强大而灵活的智能体编程框架。
