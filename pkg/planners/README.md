# Planners 规划器模块

## 概述

Planners 模块为 ADK 框架提供了智能体规划能力，使智能体能够在执行任务之前生成计划来指导其行动。该模块实现了多种规划策略，包括使用模型内置思考功能的 BuiltInPlanner 和基于 Plan-ReAct 模式的结构化规划器。

## 核心接口

### Planner 接口
```go
type Planner interface {
    // BuildPlanningInstruction 构建系统指令，追加到 LLM 请求中用于规划
    BuildPlanningInstruction(
        context agents.ReadonlyContext,
        request *models.LlmRequest,
    ) string

    // ProcessPlanningResponse 处理 LLM 的规划响应
    ProcessPlanningResponse(
        context agents.CallbackContext,
        responseParts []*models.Part,
    ) []*models.Part
}
```

规划器接口定义了智能体规划的标准方法，所有规划器实现都必须遵循此接口。

## 规划器实现

### 1. BuiltInPlanner (内置规划器)

基于模型内置思考功能的规划器，利用大语言模型自身的推理能力进行规划。

```go
type BuiltInPlanner struct {
    // ThinkingConfig 包含模型思考功能的配置
    ThinkingConfig *models.ThinkingConfig
}

func NewBuiltInPlanner(thinkingConfig *models.ThinkingConfig) *BuiltInPlanner
func NewDefaultBuiltInPlanner() *BuiltInPlanner
```

**特点:**
- 依赖模型的内置思考能力
- 不需要额外的指令模板
- 适用于支持思考功能的高级模型
- 配置简单，开箱即用

### 2. PlanReActPlanner (计划-推理-行动规划器)

基于 Plan-ReAct 模式的结构化规划器，强制 LLM 在执行任何行动前先生成计划。

```go
type PlanReActPlanner struct{}

func NewPlanReActPlanner() *PlanReActPlanner
```

**特点:**
- 不依赖模型的内置思考功能
- 使用结构化的标签和指令
- 支持规划、重新规划、推理、行动和最终答案
- 适用于所有类型的语言模型

**标签系统:**
- `/*PLANNING*/`: 初始规划标签
- `/*REPLANNING*/`: 重新规划标签
- `/*REASONING*/`: 推理标签
- `/*ACTION*/`: 行动标签
- `/*FINAL_ANSWER*/`: 最终答案标签

## 使用示例

### 基础使用
```go
package main

import (
    "context"
    "fmt"
    
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/models"
    "github.com/nvcnvn/adk-golang/pkg/planners"
)

func main() {
    // 创建内置规划器
    builtInPlanner := planners.NewDefaultBuiltInPlanner()
    
    // 创建 Plan-ReAct 规划器
    planReActPlanner := planners.NewPlanReActPlanner()
    
    // 使用规划器示例
    demonstratePlanner(builtInPlanner)
    demonstratePlanner(planReActPlanner)
}

func demonstratePlanner(planner planners.Planner) {
    // 创建模拟的上下文和请求
    ctx := createMockContext()
    request := &models.LlmRequest{
        Messages: []*models.Message{
            {
                Role: "user",
                Parts: []*models.Part{
                    {Text: "请帮我制定一个学习Go语言的计划"},
                },
            },
        },
    }
    
    // 构建规划指令
    instruction := planner.BuildPlanningInstruction(ctx, request)
    if instruction != "" {
        fmt.Printf("规划指令: %s\n", instruction)
    }
    
    // 模拟 LLM 响应
    responseParts := []*models.Part{
        {Text: "/*PLANNING*/ 我需要制定一个循序渐进的Go学习计划..."},
    }
    
    // 处理规划响应
    processedParts := planner.ProcessPlanningResponse(createMockCallbackContext(), responseParts)
    
    fmt.Printf("处理后的响应部分数量: %d\n", len(processedParts))
}

func createMockContext() agents.ReadonlyContext {
    // 创建模拟的只读上下文
    return &mockReadonlyContext{}
}

func createMockCallbackContext() agents.CallbackContext {
    // 创建模拟的回调上下文
    return &mockCallbackContext{}
}
```

### 智能体集成
```go
package agents

import (
    "github.com/nvcnvn/adk-golang/pkg/planners"
    "github.com/nvcnvn/adk-golang/pkg/models"
)

type PlanningAgent struct {
    BaseAgent
    planner planners.Planner
}

func NewPlanningAgent(planner planners.Planner) *PlanningAgent {
    return &PlanningAgent{
        planner: planner,
    }
}

func (a *PlanningAgent) Process(ctx context.Context, input string) (string, error) {
    // 创建 LLM 请求
    request := &models.LlmRequest{
        Messages: []*models.Message{
            {
                Role: "user",
                Parts: []*models.Part{
                    {Text: input},
                },
            },
        },
    }
    
    // 构建规划指令
    planningInstruction := a.planner.BuildPlanningInstruction(a.context, request)
    if planningInstruction != "" {
        // 将规划指令添加到系统消息中
        systemMessage := &models.Message{
            Role: "system",
            Parts: []*models.Part{
                {Text: planningInstruction},
            },
        }
        request.Messages = append([]*models.Message{systemMessage}, request.Messages...)
    }
    
    // 发送请求到 LLM
    response, err := a.llmClient.Generate(ctx, request)
    if err != nil {
        return "", fmt.Errorf("LLM 生成失败: %w", err)
    }
    
    // 处理规划响应
    processedParts := a.planner.ProcessPlanningResponse(a.callbackContext, response.Parts)
    
    // 提取最终答案
    finalAnswer := a.extractFinalAnswer(processedParts)
    return finalAnswer, nil
}

func (a *PlanningAgent) extractFinalAnswer(parts []*models.Part) string {
    for _, part := range parts {
        if part.Text != "" && !part.IsThought {
            return part.Text
        }
    }
    return ""
}
```

### 高级配置示例
```go
// 自定义思考配置的内置规划器
func createAdvancedBuiltInPlanner() planners.Planner {
    thinkingConfig := &models.ThinkingConfig{
        // 自定义思考配置参数
        MaxThinkingTokens: 1000,
        ThinkingMode:      "detailed",
    }
    
    return planners.NewBuiltInPlanner(thinkingConfig)
}

// 自定义 Plan-ReAct 规划器（如果需要扩展）
type CustomPlanReActPlanner struct {
    *planners.PlanReActPlanner
    customTags map[string]string
}

func NewCustomPlanReActPlanner() *CustomPlanReActPlanner {
    return &CustomPlanReActPlanner{
        PlanReActPlanner: planners.NewPlanReActPlanner(),
        customTags: map[string]string{
            "ANALYSIS":    "/*ANALYSIS*/",
            "HYPOTHESIS":  "/*HYPOTHESIS*/",
            "VALIDATION":  "/*VALIDATION*/",
        },
    }
}
```

### 多规划器策略
```go
type MultiPlanner struct {
    planners []planners.Planner
    strategy string // "fallback", "ensemble", "selective"
}

func NewMultiPlanner(strategy string, planners ...planners.Planner) *MultiPlanner {
    return &MultiPlanner{
        planners: planners,
        strategy: strategy,
    }
}

func (mp *MultiPlanner) BuildPlanningInstruction(
    context agents.ReadonlyContext,
    request *models.LlmRequest,
) string {
    switch mp.strategy {
    case "fallback":
        return mp.buildFallbackInstruction(context, request)
    case "ensemble":
        return mp.buildEnsembleInstruction(context, request)
    case "selective":
        return mp.buildSelectiveInstruction(context, request)
    default:
        return mp.planners[0].BuildPlanningInstruction(context, request)
    }
}

func (mp *MultiPlanner) buildFallbackInstruction(
    context agents.ReadonlyContext,
    request *models.LlmRequest,
) string {
    // 依次尝试每个规划器，直到获得有效指令
    for _, planner := range mp.planners {
        instruction := planner.BuildPlanningInstruction(context, request)
        if instruction != "" {
            return instruction
        }
    }
    return ""
}

func (mp *MultiPlanner) buildEnsembleInstruction(
    context agents.ReadonlyContext,
    request *models.LlmRequest,
) string {
    // 合并多个规划器的指令
    var instructions []string
    for i, planner := range mp.planners {
        instruction := planner.BuildPlanningInstruction(context, request)
        if instruction != "" {
            instructions = append(instructions, 
                fmt.Sprintf("=== 规划策略 %d ===\n%s", i+1, instruction))
        }
    }
    return strings.Join(instructions, "\n\n")
}

func (mp *MultiPlanner) ProcessPlanningResponse(
    context agents.CallbackContext,
    responseParts []*models.Part,
) []*models.Part {
    switch mp.strategy {
    case "fallback":
        return mp.processFallbackResponse(context, responseParts)
    case "ensemble":
        return mp.processEnsembleResponse(context, responseParts)
    case "selective":
        return mp.processSelectiveResponse(context, responseParts)
    default:
        return mp.planners[0].ProcessPlanningResponse(context, responseParts)
    }
}
```

## Plan-ReAct 详细指令模板

PlanReActPlanner 使用的完整指令模板包含以下结构：

```text
You MUST strictly follow this multi-step reasoning process:

1. **PLANNING PHASE**: Start with /*PLANNING*/ tag
   - Analyze the question/task thoroughly
   - Break down into logical steps
   - Identify required information and tools
   - Create a step-by-step plan

2. **EXECUTION PHASE**: Use appropriate tags
   - /*REASONING*/: Think through each step
   - /*ACTION*/: Take specific actions (tool calls, analysis)
   - /*REPLANNING*/: Adjust plan if needed

3. **CONCLUSION PHASE**: End with /*FINAL_ANSWER*/
   - Provide complete, direct answer
   - Ensure all requirements are met

IMPORTANT RULES:
- Always start with /*PLANNING*/
- Use tags to structure your thinking
- Be thorough in planning before acting
- End with /*FINAL_ANSWER*/ containing the complete solution

Example format:
/*PLANNING*/
I need to analyze this request and create a plan...
1. First, I'll...
2. Then, I'll...
3. Finally, I'll...

/*REASONING*/
Based on my plan, I should start by...

/*ACTION*/
[Take specific action or tool call]

/*FINAL_ANSWER*/
Based on my analysis, the answer is...
```

## 性能优化

### 缓存规划指令
```go
type CachedPlannerWrapper struct {
    planner planners.Planner
    cache   map[string]string
}

func NewCachedPlannerWrapper(planner planners.Planner) *CachedPlannerWrapper {
    return &CachedPlannerWrapper{
        planner: planner,
        cache:   make(map[string]string),
    }
}

func (c *CachedPlannerWrapper) BuildPlanningInstruction(
    context agents.ReadonlyContext,
    request *models.LlmRequest,
) string {
    // 生成缓存键
    cacheKey := c.generateCacheKey(request)
    
    // 检查缓存
    if cached, exists := c.cache[cacheKey]; exists {
        return cached
    }
    
    // 生成指令
    instruction := c.planner.BuildPlanningInstruction(context, request)
    
    // 缓存结果
    c.cache[cacheKey] = instruction
    
    return instruction
}

func (c *CachedPlannerWrapper) generateCacheKey(request *models.LlmRequest) string {
    // 基于请求内容生成缓存键
    var parts []string
    for _, msg := range request.Messages {
        for _, part := range msg.Parts {
            if part.Text != "" {
                parts = append(parts, part.Text)
            }
        }
    }
    return strings.Join(parts, "|")
}
```

### 异步规划处理
```go
type AsyncPlanner struct {
    planner planners.Planner
    workers int
}

func NewAsyncPlanner(planner planners.Planner, workers int) *AsyncPlanner {
    return &AsyncPlanner{
        planner: planner,
        workers: workers,
    }
}

func (a *AsyncPlanner) ProcessPlanningResponseAsync(
    ctx context.Context,
    context agents.CallbackContext,
    responseParts []*models.Part,
) <-chan []*models.Part {
    resultChan := make(chan []*models.Part, 1)
    
    go func() {
        defer close(resultChan)
        
        processed := a.planner.ProcessPlanningResponse(context, responseParts)
        
        select {
        case resultChan <- processed:
        case <-ctx.Done():
            return
        }
    }()
    
    return resultChan
}
```

## 调试和监控

### 规划过程追踪
```go
type TracingPlanner struct {
    planner planners.Planner
    tracer  PlanningTracer
}

type PlanningTracer interface {
    TraceInstruction(instruction string)
    TraceResponse(responseParts []*models.Part)
    TraceProcessed(processedParts []*models.Part)
}

func NewTracingPlanner(planner planners.Planner, tracer PlanningTracer) *TracingPlanner {
    return &TracingPlanner{
        planner: planner,
        tracer:  tracer,
    }
}

func (t *TracingPlanner) BuildPlanningInstruction(
    context agents.ReadonlyContext,
    request *models.LlmRequest,
) string {
    instruction := t.planner.BuildPlanningInstruction(context, request)
    t.tracer.TraceInstruction(instruction)
    return instruction
}

func (t *TracingPlanner) ProcessPlanningResponse(
    context agents.CallbackContext,
    responseParts []*models.Part,
) []*models.Part {
    t.tracer.TraceResponse(responseParts)
    processed := t.planner.ProcessPlanningResponse(context, responseParts)
    t.tracer.TraceProcessed(processed)
    return processed
}

// 简单的控制台追踪器实现
type ConsoleTracer struct{}

func (c *ConsoleTracer) TraceInstruction(instruction string) {
    fmt.Printf("[PLANNER] 指令: %s\n", instruction)
}

func (c *ConsoleTracer) TraceResponse(responseParts []*models.Part) {
    fmt.Printf("[PLANNER] 响应部分数量: %d\n", len(responseParts))
    for i, part := range responseParts {
        fmt.Printf("[PLANNER] 部分 %d: %s\n", i, part.Text[:min(100, len(part.Text))])
    }
}

func (c *ConsoleTracer) TraceProcessed(processedParts []*models.Part) {
    fmt.Printf("[PLANNER] 处理后部分数量: %d\n", len(processedParts))
}
```

## 最佳实践

1. **规划器选择**: 
   - 高级模型使用 BuiltInPlanner
   - 通用模型使用 PlanReActPlanner
   
2. **指令设计**: 
   - 保持指令清晰简洁
   - 提供具体的格式要求
   
3. **响应处理**: 
   - 正确区分思考部分和行动部分
   - 保留重要的推理过程
   
4. **性能优化**: 
   - 对频繁使用的指令进行缓存
   - 异步处理长时间的规划任务
   
5. **错误处理**: 
   - 验证规划响应的完整性
   - 提供回退机制

## 依赖模块

- `github.com/nvcnvn/adk-golang/pkg/agents`: 智能体核心
- `github.com/nvcnvn/adk-golang/pkg/models`: 模型接口
- Go 标准库: `strings`

## 扩展开发

### 自定义规划器
```go
type CustomPlanner struct {
    template string
    tags     map[string]string
}

func NewCustomPlanner(template string, tags map[string]string) *CustomPlanner {
    return &CustomPlanner{
        template: template,
        tags:     tags,
    }
}

func (c *CustomPlanner) BuildPlanningInstruction(
    context agents.ReadonlyContext,
    request *models.LlmRequest,
) string {
    // 根据模板和标签构建自定义指令
    instruction := c.template
    for key, value := range c.tags {
        instruction = strings.ReplaceAll(instruction, fmt.Sprintf("{{%s}}", key), value)
    }
    return instruction
}

func (c *CustomPlanner) ProcessPlanningResponse(
    context agents.CallbackContext,
    responseParts []*models.Part,
) []*models.Part {
    // 实现自定义的响应处理逻辑
    var processedParts []*models.Part
    
    for _, part := range responseParts {
        processed := c.processCustomTags(part)
        processedParts = append(processedParts, processed)
    }
    
    return processedParts
}

func (c *CustomPlanner) processCustomTags(part *models.Part) *models.Part {
    // 处理自定义标签的逻辑
    text := part.Text
    
    for tag := range c.tags {
        if strings.Contains(text, tag) {
            // 标记为思考部分或保留为行动部分
            if c.isThoughtTag(tag) {
                return &models.Part{
                    Text:      text,
                    IsThought: true,
                }
            }
        }
    }
    
    return part
}

func (c *CustomPlanner) isThoughtTag(tag string) bool {
    thoughtTags := []string{"PLANNING", "REASONING", "ANALYSIS"}
    for _, t := range thoughtTags {
        if strings.Contains(tag, t) {
            return true
        }
    }
    return false
}
```

Planners 模块为 ADK-Golang 框架提供了强大的智能体规划能力，支持多种规划策略和自定义扩展，是构建高效智能体系统的重要组件。
