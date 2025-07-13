# Evaluation 评估模块

## 概述

Evaluation 模块提供了完整的智能体性能评估框架，支持多维度的智能体行为评估，包括响应质量评估、工具使用轨迹评估和对话表现评估。该模块是构建高质量智能体系统的重要质量保证工具。

## 核心组件

### 数据结构

#### EvaluationEntry (评估条目)
```go
type EvaluationEntry map[string]interface{}
```

表示评估数据集中的单个条目，包含查询、响应、参考答案和工具使用等信息。

#### EvaluationConversation (评估对话)
```go
type EvaluationConversation []EvaluationEntry
```

表示包含多个评估条目的完整对话。

#### EvaluationDataset (评估数据集)
```go
type EvaluationDataset []EvaluationConversation
```

评估对话的集合，构成完整的评估数据集。

#### ToolUse (工具使用)
```go
type ToolUse struct {
    ToolName   string                 `json:"tool_name"`    // 工具名称
    ToolInput  map[string]interface{} `json:"tool_input"`   // 工具输入参数
    ToolOutput interface{}            `json:"mock_tool_output,omitempty"` // 模拟工具输出
}
```

表示智能体的工具调用实例。

## 评估维度和指标

### 评估常量
```go
const (
    // 数据字段常量
    Query           = "query"            // 用户查询
    ExpectedToolUse = "expected_tool_use" // 期望的工具使用
    Response        = "response"         // 智能体响应
    Reference       = "reference"        // 参考答案
    ActualToolUse   = "actual_tool_use"  // 实际工具使用
    
    // 评估指标常量
    ToolTrajectoryScoreKey     = "tool_trajectory_avg_score"  // 工具轨迹平均分
    ResponseEvaluationScoreKey = "response_evaluation_score"  // 响应评估分数
    ResponseMatchScoreKey      = "response_match_score"       // 响应匹配分数
)
```

### 默认评估标准
```go
const (
    DefaultToolTrajectoryScore = 1.0 // 1分制，1.0为完美
    DefaultResponseMatchScore  = 0.8 // Rouge-1文本匹配，0.8为默认值
)

var DefaultCriteria = map[string]float64{
    ToolTrajectoryScoreKey: DefaultToolTrajectoryScore,
    ResponseMatchScoreKey:  DefaultResponseMatchScore,
}
```

## 评估器类型

### 1. AgentEvaluator (智能体评估器)
综合评估智能体的整体性能，包括响应质量和工具使用能力。

### 2. ResponseEvaluator (响应评估器)
专门评估智能体响应的质量，包括准确性、相关性和完整性。

### 3. TrajectoryEvaluator (轨迹评估器)
评估智能体的行为轨迹，特别是工具调用的合理性和执行顺序。

### 4. EvaluationGenerator (评估生成器)
生成评估数据集和测试用例。

## 使用示例

### 基本评估流程
```go
package main

import (
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/evaluation"
)

func main() {
    // 创建评估条目
    entry := evaluation.EvaluationEntry{
        evaluation.Query:     "今天北京的天气如何？",
        evaluation.Reference: "今天北京晴天，气温25°C，微风。",
        evaluation.ExpectedToolUse: []evaluation.ToolUse{
            {
                ToolName: "weather_api",
                ToolInput: map[string]interface{}{
                    "city": "北京",
                    "date": "today",
                },
            },
        },
    }
    
    // 获取查询和参考答案
    query := entry.GetQuery()
    reference := entry.GetReference()
    
    fmt.Printf("查询: %s\n", query)
    fmt.Printf("参考答案: %s\n", reference)
    
    // 获取期望的工具使用
    expectedTools, err := entry.GetExpectedToolUse()
    if err != nil {
        fmt.Printf("获取期望工具使用失败: %v\n", err)
        return
    }
    
    fmt.Printf("期望使用工具: %s\n", expectedTools[0].ToolName)
}
```

### 设置智能体响应和实际工具使用
```go
func evaluateAgentResponse() {
    entry := evaluation.EvaluationEntry{}
    
    // 设置智能体响应
    agentResponse := "根据天气API查询，今天北京天气晴朗，温度为25度，有轻微的风。"
    entry.SetResponse(agentResponse)
    
    // 设置实际工具使用
    actualTools := []evaluation.ToolUse{
        {
            ToolName: "weather_api",
            ToolInput: map[string]interface{}{
                "city": "北京",
                "date": "2024-01-15",
            },
            ToolOutput: map[string]interface{}{
                "temperature": 25,
                "condition":   "sunny",
                "wind":        "light",
            },
        },
    }
    entry.SetActualToolUse(actualTools)
    
    // 获取并比较实际工具使用
    retrievedTools, err := entry.GetActualToolUse()
    if err != nil {
        fmt.Printf("获取实际工具使用失败: %v\n", err)
        return
    }
    
    fmt.Printf("实际使用的工具数量: %d\n", len(retrievedTools))
    for i, tool := range retrievedTools {
        fmt.Printf("工具 %d: %s\n", i+1, tool.ToolName)
    }
}
```

### 创建和使用评估数据集
```go
func createEvaluationDataset() evaluation.EvaluationDataset {
    // 创建多个评估对话
    dataset := evaluation.EvaluationDataset{
        // 第一个对话 - 天气查询
        evaluation.EvaluationConversation{
            evaluation.EvaluationEntry{
                evaluation.Query:     "今天上海的天气怎么样？",
                evaluation.Reference: "今天上海多云，气温22°C。",
                evaluation.ExpectedToolUse: []evaluation.ToolUse{
                    {
                        ToolName: "weather_api",
                        ToolInput: map[string]interface{}{"city": "上海"},
                    },
                },
            },
        },
        // 第二个对话 - 计算任务
        evaluation.EvaluationConversation{
            evaluation.EvaluationEntry{
                evaluation.Query:     "计算 15 * 23 等于多少？",
                evaluation.Reference: "15 * 23 = 345",
                evaluation.ExpectedToolUse: []evaluation.ToolUse{
                    {
                        ToolName: "calculator",
                        ToolInput: map[string]interface{}{
                            "operation": "multiply",
                            "a":         15,
                            "b":         23,
                        },
                    },
                },
            },
        },
    }
    
    return dataset
}

func runEvaluationOnDataset(dataset evaluation.EvaluationDataset) {
    fmt.Printf("评估数据集包含 %d 个对话\n", len(dataset))
    
    for i, conversation := range dataset {
        fmt.Printf("对话 %d 包含 %d 个条目\n", i+1, len(conversation))
        
        for j, entry := range conversation {
            query := entry.GetQuery()
            reference := entry.GetReference()
            
            fmt.Printf("  条目 %d:\n", j+1)
            fmt.Printf("    查询: %s\n", query)
            fmt.Printf("    参考: %s\n", reference)
        }
    }
}
```

## 高级评估功能

### 工具轨迹评估
```go
func evaluateToolTrajectory(expected, actual []evaluation.ToolUse) float64 {
    if len(expected) == 0 && len(actual) == 0 {
        return 1.0 // 完美匹配
    }
    
    if len(expected) != len(actual) {
        return 0.0 // 工具数量不匹配
    }
    
    score := 0.0
    for i, expectedTool := range expected {
        if i < len(actual) {
            actualTool := actual[i]
            
            // 工具名称匹配
            if expectedTool.ToolName == actualTool.ToolName {
                score += 0.5
                
                // 参数匹配评估
                if evaluateToolParameters(expectedTool.ToolInput, actualTool.ToolInput) {
                    score += 0.5
                }
            }
        }
    }
    
    return score / float64(len(expected))
}

func evaluateToolParameters(expected, actual map[string]interface{}) bool {
    if len(expected) != len(actual) {
        return false
    }
    
    for key, expectedValue := range expected {
        if actualValue, exists := actual[key]; !exists || actualValue != expectedValue {
            return false
        }
    }
    
    return true
}
```

### 响应质量评估
```go
func evaluateResponseQuality(response, reference string) float64 {
    // 简单的文本相似度评估（实际实现可能使用更复杂的NLP方法）
    response = strings.ToLower(strings.TrimSpace(response))
    reference = strings.ToLower(strings.TrimSpace(reference))
    
    if response == reference {
        return 1.0 // 完全匹配
    }
    
    // 计算Rouge-1分数或其他文本相似度指标
    return calculateTextSimilarity(response, reference)
}

func calculateTextSimilarity(text1, text2 string) float64 {
    // 简化的实现 - 实际应用中可能使用更复杂的算法
    words1 := strings.Fields(text1)
    words2 := strings.Fields(text2)
    
    commonWords := 0
    wordSet2 := make(map[string]bool)
    for _, word := range words2 {
        wordSet2[word] = true
    }
    
    for _, word := range words1 {
        if wordSet2[word] {
            commonWords++
        }
    }
    
    if len(words1) == 0 {
        return 0.0
    }
    
    return float64(commonWords) / float64(len(words1))
}
```

### 综合评估报告
```go
type EvaluationReport struct {
    TotalEntries           int                    `json:"total_entries"`
    ToolTrajectoryScore    float64               `json:"tool_trajectory_score"`
    ResponseEvaluationScore float64              `json:"response_evaluation_score"`
    ResponseMatchScore     float64               `json:"response_match_score"`
    OverallScore          float64               `json:"overall_score"`
    DetailedResults       []EntryEvaluationResult `json:"detailed_results"`
}

type EntryEvaluationResult struct {
    Query               string  `json:"query"`
    ToolScore          float64 `json:"tool_score"`
    ResponseScore      float64 `json:"response_score"`
    OverallEntryScore  float64 `json:"overall_entry_score"`
}

func generateEvaluationReport(dataset evaluation.EvaluationDataset) *EvaluationReport {
    report := &EvaluationReport{
        DetailedResults: make([]EntryEvaluationResult, 0),
    }
    
    totalToolScore := 0.0
    totalResponseScore := 0.0
    entryCount := 0
    
    for _, conversation := range dataset {
        for _, entry := range conversation {
            // 评估工具使用
            expectedTools, _ := entry.GetExpectedToolUse()
            actualTools, _ := entry.GetActualToolUse()
            toolScore := evaluateToolTrajectory(expectedTools, actualTools)
            
            // 评估响应质量
            response := entry.GetResponse()
            reference := entry.GetReference()
            responseScore := evaluateResponseQuality(response, reference)
            
            // 记录详细结果
            entryResult := EntryEvaluationResult{
                Query:             entry.GetQuery(),
                ToolScore:         toolScore,
                ResponseScore:     responseScore,
                OverallEntryScore: (toolScore + responseScore) / 2.0,
            }
            report.DetailedResults = append(report.DetailedResults, entryResult)
            
            totalToolScore += toolScore
            totalResponseScore += responseScore
            entryCount++
        }
    }
    
    if entryCount > 0 {
        report.TotalEntries = entryCount
        report.ToolTrajectoryScore = totalToolScore / float64(entryCount)
        report.ResponseEvaluationScore = totalResponseScore / float64(entryCount)
        report.ResponseMatchScore = report.ResponseEvaluationScore // 简化处理
        report.OverallScore = (report.ToolTrajectoryScore + report.ResponseEvaluationScore) / 2.0
    }
    
    return report
}
```

## 配置和扩展

### 评估配置
```go
type EvaluationConfig struct {
    Criteria       map[string]float64 `yaml:"criteria"`       // 评估标准权重
    EnabledMetrics []string          `yaml:"enabled_metrics"` // 启用的评估指标
    OutputFormat   string            `yaml:"output_format"`   // 输出格式 (json/yaml/csv)
}

func LoadEvaluationConfig(configPath string) (*EvaluationConfig, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config EvaluationConfig
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    // 设置默认值
    if config.Criteria == nil {
        config.Criteria = evaluation.DefaultCriteria
    }
    
    return &config, nil
}
```

### 自定义评估器
```go
type CustomEvaluator struct {
    config *EvaluationConfig
}

func NewCustomEvaluator(config *EvaluationConfig) *CustomEvaluator {
    return &CustomEvaluator{config: config}
}

func (e *CustomEvaluator) EvaluateEntry(entry evaluation.EvaluationEntry) map[string]float64 {
    scores := make(map[string]float64)
    
    // 自定义评估逻辑
    for metric := range e.config.Criteria {
        switch metric {
        case evaluation.ToolTrajectoryScoreKey:
            scores[metric] = e.evaluateToolTrajectory(entry)
        case evaluation.ResponseEvaluationScoreKey:
            scores[metric] = e.evaluateResponse(entry)
        case evaluation.ResponseMatchScoreKey:
            scores[metric] = e.evaluateResponseMatch(entry)
        }
    }
    
    return scores
}

func (e *CustomEvaluator) evaluateToolTrajectory(entry evaluation.EvaluationEntry) float64 {
    // 自定义工具轨迹评估逻辑
    return 1.0
}

func (e *CustomEvaluator) evaluateResponse(entry evaluation.EvaluationEntry) float64 {
    // 自定义响应评估逻辑
    return 1.0
}

func (e *CustomEvaluator) evaluateResponseMatch(entry evaluation.EvaluationEntry) float64 {
    // 自定义响应匹配评估逻辑
    return 1.0
}
```

## 批量评估和性能监控

### 批量评估
```go
func RunBatchEvaluation(dataset evaluation.EvaluationDataset, agent agents.Agent) *EvaluationReport {
    ctx := context.Background()
    
    for conversationIdx, conversation := range dataset {
        for entryIdx, entry := range conversation {
            query := entry.GetQuery()
            
            // 运行智能体
            response, err := agent.Process(ctx, query)
            if err != nil {
                fmt.Printf("智能体处理失败 [%d:%d]: %v\n", conversationIdx, entryIdx, err)
                continue
            }
            
            // 记录响应
            entry.SetResponse(response)
            
            // 如果智能体有工具使用记录，也记录下来
            if toolUser, ok := agent.(ToolUser); ok {
                actualTools := toolUser.GetUsedTools()
                entry.SetActualToolUse(actualTools)
            }
        }
    }
    
    return generateEvaluationReport(dataset)
}
```

### 实时评估监控
```go
type EvaluationMonitor struct {
    recentScores []float64
    threshold    float64
    mu           sync.RWMutex
}

func NewEvaluationMonitor(threshold float64) *EvaluationMonitor {
    return &EvaluationMonitor{
        recentScores: make([]float64, 0),
        threshold:    threshold,
    }
}

func (m *EvaluationMonitor) RecordScore(score float64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.recentScores = append(m.recentScores, score)
    
    // 保持最近100个评分
    if len(m.recentScores) > 100 {
        m.recentScores = m.recentScores[1:]
    }
    
    // 检查是否低于阈值
    if score < m.threshold {
        fmt.Printf("警告: 评估分数 %.2f 低于阈值 %.2f\n", score, m.threshold)
    }
}

func (m *EvaluationMonitor) GetAverageScore() float64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if len(m.recentScores) == 0 {
        return 0.0
    }
    
    sum := 0.0
    for _, score := range m.recentScores {
        sum += score
    }
    
    return sum / float64(len(m.recentScores))
}
```

## 最佳实践

1. **评估数据质量**: 确保评估数据集具有代表性和多样性
2. **多维度评估**: 综合考虑工具使用、响应质量和用户体验
3. **基准建立**: 建立稳定的评估基准，便于比较不同版本的性能
4. **持续监控**: 在生产环境中持续监控智能体性能
5. **A/B测试**: 使用评估框架进行不同智能体版本的对比测试
6. **反馈循环**: 将评估结果用于智能体的持续改进

## 依赖

- `encoding/json`: JSON 序列化
- Go 标准库: `fmt`, `strings`, `sync`

## 应用场景

- 智能体性能基准测试
- 模型版本对比评估
- 生产环境质量监控
- 智能体能力验证
- 用户体验优化

Evaluation 模块为 ADK-Golang 框架提供了企业级的智能体评估能力，确保智能体系统的质量和可靠性。
