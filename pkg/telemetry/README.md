# Telemetry 遥测模块

## 概述

Telemetry 模块为 ADK 框架提供遥测功能，包括日志记录和分布式追踪能力。该模块支持不同级别的日志输出、跨度（Span）追踪、事件记录和属性设置，是系统监控和调试的核心组件。

## 核心功能

### 1. 日志记录
支持四个级别的日志记录：DEBUG、INFO、WARNING、ERROR

### 2. 分布式追踪
提供 Span 和 Tracer 接口，支持操作追踪和性能监控

### 3. 事件和属性
支持为 Span 添加事件和属性，便于调试和分析

## 日志功能

### 日志级别
```go
type LogLevel int

const (
    LevelDebug   LogLevel = iota // 调试级别
    LevelInfo                    // 信息级别  
    LevelWarning                 // 警告级别
    LevelError                   // 错误级别
)
```

### 日志函数
```go
// 设置日志级别
func SetLogLevel(level LogLevel)

// 不同级别的日志记录
func Debug(format string, v ...interface{})
func Info(format string, v ...interface{})
func Warning(format string, v ...interface{})
func Error(format string, v ...interface{})
```

### 使用示例
```go
package main

import "github.com/nvcnvn/adk-golang/pkg/telemetry"

func main() {
    // 设置日志级别
    telemetry.SetLogLevel(telemetry.LevelDebug)
    
    // 记录不同级别的日志
    telemetry.Debug("这是调试信息: %s", "debug message")
    telemetry.Info("这是信息日志: %d", 42)
    telemetry.Warning("这是警告信息: %v", map[string]string{"key": "value"})
    telemetry.Error("这是错误信息: %s", "error occurred")
}
```

## 追踪功能

### 核心接口

#### Span 接口
```go
type Span interface {
    // End 完成跨度
    End()
    
    // AddEvent 为跨度添加事件
    AddEvent(name string, attributes map[string]string)
    
    // SetAttribute 为跨度设置属性
    SetAttribute(key, value string)
}
```

#### Tracer 接口
```go
type Tracer interface {
    // Start 启动新的跨度
    Start(ctx context.Context, name string) (context.Context, Span)
}
```

### SimpleSpan 实现
```go
type SimpleSpan struct {
    Name       string                 // 跨度名称
    StartTime  time.Time             // 开始时间
    EndTime    time.Time             // 结束时间  
    Events     []SpanEvent           // 事件列表
    Attributes map[string]string     // 属性映射
    mu         sync.Mutex            // 并发保护
}

type SpanEvent struct {
    Name       string                // 事件名称
    Time       time.Time            // 事件时间
    Attributes map[string]string    // 事件属性
}
```

### 使用示例

#### 基本追踪
```go
package main

import (
    "context"
    
    "github.com/nvcnvn/adk-golang/pkg/telemetry"
)

func main() {
    // 创建简单追踪器
    tracer := telemetry.NewSimpleTracer()
    telemetry.SetDefaultTracer(tracer)
    
    // 创建跨度
    ctx := context.Background()
    ctx, span := telemetry.StartSpan(ctx, "main_operation")
    defer span.End()
    
    // 设置属性
    span.SetAttribute("user_id", "12345")
    span.SetAttribute("operation_type", "data_processing")
    
    // 添加事件
    span.AddEvent("processing_started", map[string]string{
        "input_size": "1024",
    })
    
    // 执行业务逻辑
    processData(ctx)
    
    // 添加完成事件
    span.AddEvent("processing_completed", map[string]string{
        "status": "success",
    })
}

func processData(ctx context.Context) {
    // 创建子跨度
    ctx, span := telemetry.StartSpan(ctx, "data_processing")
    defer span.End()
    
    span.SetAttribute("method", "batch_process")
    
    // 模拟处理逻辑
    span.AddEvent("validation_step", nil)
    
    // 处理逻辑...
    
    span.AddEvent("transformation_step", map[string]string{
        "records_processed": "500",
    })
}
```

#### 智能体追踪集成
```go
package agents

import (
    "context"
    
    "github.com/nvcnvn/adk-golang/pkg/telemetry"
)

type TrackedAgent struct {
    name  string
    model string
}

func (a *TrackedAgent) Process(ctx context.Context, input string) (string, error) {
    // 创建智能体处理跨度
    ctx, span := telemetry.StartSpan(ctx, "agent.process")
    defer span.End()
    
    // 设置智能体信息
    span.SetAttribute("agent.name", a.name)
    span.SetAttribute("agent.model", a.model)
    span.SetAttribute("input.length", fmt.Sprintf("%d", len(input)))
    
    // 添加开始处理事件
    span.AddEvent("processing_started", map[string]string{
        "input_preview": input[:min(50, len(input))],
    })
    
    // 执行 LLM 调用
    response, err := a.callLLM(ctx, input)
    if err != nil {
        span.SetAttribute("error", err.Error())
        span.AddEvent("processing_failed", map[string]string{
            "error": err.Error(),
        })
        return "", err
    }
    
    // 设置响应信息
    span.SetAttribute("response.length", fmt.Sprintf("%d", len(response)))
    span.AddEvent("processing_completed", map[string]string{
        "response_preview": response[:min(50, len(response))],
    })
    
    return response, nil
}

func (a *TrackedAgent) callLLM(ctx context.Context, input string) (string, error) {
    // 创建 LLM 调用跨度
    ctx, span := telemetry.StartSpan(ctx, "llm.call")
    defer span.End()
    
    span.SetAttribute("llm.model", a.model)
    span.SetAttribute("llm.provider", "openai")
    
    // 模拟 LLM 调用
    span.AddEvent("request_sent", nil)
    
    // 实际的 LLM 调用逻辑...
    response := "模拟 LLM 响应"
    
    span.AddEvent("response_received", map[string]string{
        "token_count": "150",
    })
    
    return response, nil
}
```

### 高级功能

#### 自定义追踪器
```go
type CustomTracer struct {
    spans    []*SimpleSpan
    exporter SpanExporter
    mu       sync.Mutex
}

type SpanExporter interface {
    Export(spans []*SimpleSpan) error
}

func NewCustomTracer(exporter SpanExporter) *CustomTracer {
    return &CustomTracer{
        spans:    make([]*SimpleSpan, 0),
        exporter: exporter,
    }
}

func (ct *CustomTracer) Start(ctx context.Context, name string) (context.Context, Span) {
    span := telemetry.NewSimpleSpan(name)
    
    ct.mu.Lock()
    ct.spans = append(ct.spans, span)
    ct.mu.Unlock()
    
    return ctx, span
}

func (ct *CustomTracer) Export() error {
    ct.mu.Lock()
    spans := make([]*SimpleSpan, len(ct.spans))
    copy(spans, ct.spans)
    ct.spans = ct.spans[:0] // 清空已导出的跨度
    ct.mu.Unlock()
    
    if ct.exporter != nil {
        return ct.exporter.Export(spans)
    }
    
    return nil
}

// 控制台导出器示例
type ConsoleExporter struct{}

func (ce *ConsoleExporter) Export(spans []*SimpleSpan) error {
    for _, span := range spans {
        fmt.Printf("跨度: %s, 持续时间: %v\n", 
            span.Name, 
            span.EndTime.Sub(span.StartTime))
            
        for key, value := range span.Attributes {
            fmt.Printf("  属性: %s = %s\n", key, value)
        }
        
        for _, event := range span.Events {
            fmt.Printf("  事件: %s 在 %v\n", event.Name, event.Time)
        }
    }
    return nil
}
```

#### 性能分析
```go
type PerformanceTracker struct {
    tracer   *telemetry.SimpleTracer
    metrics  map[string][]time.Duration
    mu       sync.Mutex
}

func NewPerformanceTracker() *PerformanceTracker {
    return &PerformanceTracker{
        tracer:  telemetry.NewSimpleTracer(),
        metrics: make(map[string][]time.Duration),
    }
}

func (pt *PerformanceTracker) TrackOperation(name string, operation func()) {
    ctx, span := pt.tracer.Start(context.Background(), name)
    defer func() {
        span.End()
        
        // 记录性能指标
        if simpleSpan, ok := span.(*telemetry.SimpleSpan); ok {
            duration := simpleSpan.EndTime.Sub(simpleSpan.StartTime)
            pt.recordMetric(name, duration)
        }
    }()
    
    operation()
}

func (pt *PerformanceTracker) recordMetric(name string, duration time.Duration) {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    
    pt.metrics[name] = append(pt.metrics[name], duration)
}

func (pt *PerformanceTracker) GetStatistics(name string) (avg, min, max time.Duration) {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    
    durations := pt.metrics[name]
    if len(durations) == 0 {
        return 0, 0, 0
    }
    
    var total time.Duration
    min = durations[0]
    max = durations[0]
    
    for _, d := range durations {
        total += d
        if d < min {
            min = d
        }
        if d > max {
            max = d
        }
    }
    
    avg = total / time.Duration(len(durations))
    return avg, min, max
}

// 使用示例
func demonstratePerformanceTracking() {
    tracker := NewPerformanceTracker()
    
    // 跟踪数据库操作
    tracker.TrackOperation("database.query", func() {
        time.Sleep(50 * time.Millisecond) // 模拟数据库查询
    })
    
    // 跟踪 API 调用
    tracker.TrackOperation("api.call", func() {
        time.Sleep(100 * time.Millisecond) // 模拟 API 调用
    })
    
    // 获取统计信息
    avg, min, max := tracker.GetStatistics("database.query")
    fmt.Printf("数据库查询 - 平均: %v, 最小: %v, 最大: %v\n", avg, min, max)
}
```

### 集成监控系统

#### OpenTelemetry 集成
```go
package telemetry

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

type OpenTelemetryTracer struct {
    tracer trace.Tracer
}

func NewOpenTelemetryTracer(tracerName string) *OpenTelemetryTracer {
    return &OpenTelemetryTracer{
        tracer: otel.Tracer(tracerName),
    }
}

func (ott *OpenTelemetryTracer) Start(ctx context.Context, name string) (context.Context, Span) {
    ctx, otelSpan := ott.tracer.Start(ctx, name)
    
    return ctx, &OpenTelemetrySpan{
        span: otelSpan,
    }
}

type OpenTelemetrySpan struct {
    span trace.Span
}

func (ots *OpenTelemetrySpan) End() {
    ots.span.End()
}

func (ots *OpenTelemetrySpan) AddEvent(name string, attributes map[string]string) {
    // 转换属性格式
    otelAttrs := make([]trace.EventOption, 0, len(attributes))
    for k, v := range attributes {
        otelAttrs = append(otelAttrs, trace.WithAttributes(
            attribute.String(k, v),
        ))
    }
    
    ots.span.AddEvent(name, otelAttrs...)
}

func (ots *OpenTelemetrySpan) SetAttribute(key, value string) {
    ots.span.SetAttributes(attribute.String(key, value))
}
```

#### 云端追踪集成
```go
type CloudTraceExporter struct {
    projectID string
    client    *cloudtrace.Client
}

func NewCloudTraceExporter(projectID string) (*CloudTraceExporter, error) {
    client, err := cloudtrace.NewClient(context.Background())
    if err != nil {
        return nil, err
    }
    
    return &CloudTraceExporter{
        projectID: projectID,
        client:    client,
    }, nil
}

func (cte *CloudTraceExporter) Export(spans []*SimpleSpan) error {
    traces := cte.convertToCloudTraces(spans)
    
    for _, trace := range traces {
        req := &cloudtracepb.PatchTracesRequest{
            ProjectId: cte.projectID,
            Traces:    trace,
        }
        
        if err := cte.client.PatchTraces(context.Background(), req); err != nil {
            return err
        }
    }
    
    return nil
}

func (cte *CloudTraceExporter) convertToCloudTraces(spans []*SimpleSpan) []*cloudtracepb.Traces {
    // 转换逻辑实现...
    return nil
}
```

## 默认追踪器管理

### 全局追踪器设置
```go
// 设置默认追踪器
func SetDefaultTracer(tracer Tracer)

// 获取默认追踪器
func GetDefaultTracer() Tracer

// 使用默认追踪器创建跨度
func StartSpan(ctx context.Context, name string) (context.Context, Span)
```

### 使用示例
```go
func initializeTracing() {
    // 创建并设置自定义追踪器
    tracer := telemetry.NewSimpleTracer()
    telemetry.SetDefaultTracer(tracer)
    
    // 现在可以在整个应用中使用默认追踪器
    ctx, span := telemetry.StartSpan(context.Background(), "application_startup")
    defer span.End()
    
    span.SetAttribute("version", "1.0.0")
    span.AddEvent("initialization_completed", nil)
}
```

## 最佳实践

### 1. 日志级别管理
```go
func configureLogging() {
    // 生产环境使用 INFO 级别
    if os.Getenv("ENV") == "production" {
        telemetry.SetLogLevel(telemetry.LevelInfo)
    } else {
        // 开发环境使用 DEBUG 级别
        telemetry.SetLogLevel(telemetry.LevelDebug)
    }
}
```

### 2. 跨度命名规范
```go
// 好的跨度命名
ctx, span := telemetry.StartSpan(ctx, "user.authenticate")
ctx, span := telemetry.StartSpan(ctx, "database.query.users")
ctx, span := telemetry.StartSpan(ctx, "llm.generate.response")

// 避免的命名
ctx, span := telemetry.StartSpan(ctx, "function") // 太泛化
ctx, span := telemetry.StartSpan(ctx, "step1")   // 不具描述性
```

### 3. 属性设置
```go
func setMeaningfulAttributes(span telemetry.Span, userID string, requestSize int) {
    // 设置有意义的属性
    span.SetAttribute("user.id", userID)
    span.SetAttribute("request.size", fmt.Sprintf("%d", requestSize))
    span.SetAttribute("service.version", "v1.2.3")
    
    // 避免敏感信息
    // span.SetAttribute("password", password) // 不要这样做
    // span.SetAttribute("api.key", apiKey)    // 不要这样做
}
```

### 4. 错误处理
```go
func processWithErrorHandling(ctx context.Context) error {
    ctx, span := telemetry.StartSpan(ctx, "data.process")
    defer span.End()
    
    err := doSomeWork()
    if err != nil {
        // 记录错误信息
        span.SetAttribute("error", "true")
        span.SetAttribute("error.message", err.Error())
        span.AddEvent("error_occurred", map[string]string{
            "error_type": fmt.Sprintf("%T", err),
        })
        
        telemetry.Error("处理失败: %v", err)
        return err
    }
    
    span.SetAttribute("status", "success")
    return nil
}
```

### 5. 性能监控
```go
func monitorPerformance(ctx context.Context) {
    ctx, span := telemetry.StartSpan(ctx, "performance.critical.operation")
    defer span.End()
    
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        span.SetAttribute("duration_ms", fmt.Sprintf("%.2f", float64(duration.Nanoseconds())/1e6))
        
        // 性能警告
        if duration > 1*time.Second {
            telemetry.Warning("操作耗时过长: %v", duration)
            span.AddEvent("performance_warning", map[string]string{
                "threshold": "1s",
                "actual":    duration.String(),
            })
        }
    }()
    
    // 执行关键操作
    performCriticalOperation()
}
```

## 依赖模块

- Go 标准库: `context`, `log`, `os`, `sync`, `time`
- 可选集成: OpenTelemetry, Google Cloud Trace

## 扩展开发

### 自定义 Span 实现
```go
type CustomSpan struct {
    *telemetry.SimpleSpan
    customData map[string]interface{}
}

func NewCustomSpan(name string) *CustomSpan {
    return &CustomSpan{
        SimpleSpan: telemetry.NewSimpleSpan(name),
        customData: make(map[string]interface{}),
    }
}

func (cs *CustomSpan) SetCustomData(key string, value interface{}) {
    cs.customData[key] = value
}

func (cs *CustomSpan) GetCustomData(key string) interface{} {
    return cs.customData[key]
}
```

Telemetry 模块为 ADK-Golang 框架提供了全面的遥测能力，通过日志记录和分布式追踪，帮助开发者监控系统性能、调试问题和优化应用。该模块设计简洁、功能强大，支持多种集成方案，是构建可观测性系统的重要基础。
