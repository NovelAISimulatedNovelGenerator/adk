# Novel v4 Go Plugin

Novel v4 是基于 ADK 框架的改进版小说创作插件，提供了完整的分层智能体架构和插件化设计。

## 架构设计

### 核心智能体

1. **ArchitectAgent (架构师智能体)**
   - 负责小说的整体架构设计
   - 功能：世界观构建、人物关系设计、情节框架规划、主题分析
   - 输出：结构化的架构设计 (ArchitectureDesign)

2. **WriterAgent (写作者智能体)**
   - 负责具体的内容创作
   - 功能：文本创作、对话生成、场景描写、情节发展、文风调整
   - 输出：创作内容和写作分析 (WritingResult)

3. **LibrarianAgent (图书管理员智能体)**
   - 负责知识管理和内容组织
   - 功能：知识管理、内容组织、设定维护、连贯性检查、版本控制
   - 输出：知识管理报告 (LibrarianResult)

4. **CoordinatorAgent (协调者智能体)**
   - 负责整体协调和任务调度
   - 功能：任务协调、流程管理、质量控制、资源调度、进度跟踪
   - 输出：协调计划和质量报告 (CoordinationPlan)

### 架构层次

```
Novel v4 Root Agent
├── Decision Layer (决策层 - 串行)
│   └── CoordinatorAgent (协调者)
└── Execution Layer (执行层 - 并行)
    ├── ArchitectAgent (架构师)
    ├── WriterAgent (写作者)
    └── LibrarianAgent (图书管理员)
```

## 主要特性

1. **分层架构**：采用决策层和执行层分离的设计
2. **插件化**：可作为独立插件加载和使用
3. **并行处理**：执行层支持并行处理，提高效率
4. **上下文传递**：决策层实现串行上下文传递
5. **质量控制**：内置质量评估和监控机制
6. **知识管理**：完整的知识库管理系统

## 使用方法

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/flows/novel_v4"
)

func main() {
    // 构建 novel_v4 智能体
    agent := novel_v4.Build()
    
    // 创作请求
    premise := "在一个被魔法和科技共存的世界中，一个年轻的发明家发现了古老的秘密..."
    
    // 执行创作
    result, err := agent.Process(context.Background(), premise)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("创作结果:", result)
}
```

### 高级使用

```go
// 直接使用特定智能体
architect := novel_v4.NewArchitectAgent()

// 设计架构
design, err := architect.DesignArchitecture(context.Background(), premise)
if err != nil {
    panic(err)
}

// 创作内容
writer := novel_v4.NewWriterAgent()
content, err := writer.WriteContent(context.Background(), design, 1)
if err != nil {
    panic(err)
}
```

## 数据结构

### ArchitectureDesign (架构设计)
```go
type ArchitectureDesign struct {
    Worldview     string `json:"worldview"`     // 世界观设定
    Characters    string `json:"characters"`    // 人物设计
    PlotStructure string `json:"plot_structure"` // 情节结构
    Themes        string `json:"themes"`        // 主题设定
    Chapters      string `json:"chapters"`      // 章节结构
    Premise       string `json:"premise"`       // 原始前提
}
```

### WritingResult (创作结果)
```go
type WritingResult struct {
    Content         string `json:"content"`           // 创作内容
    StyleNotes      string `json:"style_notes"`       // 文风说明
    CharacterVoices string `json:"character_voices"`  // 人物对话特色
    SceneDetails    string `json:"scene_details"`     // 场景描写要点
}
```

## 配置选项

每个智能体都支持配置参数：

```go
agent := novel_v4.NewArchitectAgent()
agent.SetConfig("creativity_level", 0.8)
agent.SetConfig("detail_level", "high")
```

## 与 novel_v3 的改进

1. **更完整的插件化实现**：不再直接复用 framework.go
2. **更丰富的智能体类型**：新增协调者智能体
3. **更强的扩展性**：基于 BaseAgent 的统一抽象
4. **更好的质量控制**：内置质量评估机制
5. **更完善的知识管理**：专门的知识库管理系统

## 开发指南

### 添加新的智能体

1. 继承 BaseAgent 基类
2. 实现特定的处理逻辑
3. 在 builder.go 中注册新智能体
4. 添加相应的测试用例

### 扩展现有功能

1. 通过 SetConfig 配置参数
2. 重写 ProcessWithContext 方法
3. 实现自定义的结果解析逻辑

## 依赖关系

- `github.com/nvcnvn/adk-golang/pkg/agents`: 基础智能体框架
- `context`: 上下文管理
- `encoding/json`: JSON 序列化
- `time`: 时间处理

## 版本历史

- v4.0.0: 初始版本，完整的插件化实现
- 基于 novel_v3 的架构设计优化而来
