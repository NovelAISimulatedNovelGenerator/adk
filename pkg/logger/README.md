# Logger 日志模块

## 概述

Logger 模块为 ADK 框架提供统一的日志记录功能，基于 Uber 的 zap 日志库构建。该模块提供了高性能、结构化的日志记录能力，支持多种日志级别、格式化输出和全局日志管理。

## 核心功能

### 全局日志器
- 单例模式的全局日志器管理
- 线程安全的初始化机制
- 支持开发和生产环境配置

### 日志级别
- **Debug**: 详细的调试信息
- **Info**: 一般信息记录
- **Warn**: 警告信息
- **Error**: 错误信息

### 输出格式
- JSON 格式结构化日志
- 支持时间戳、日志级别、调用者信息
- 可配置的编码器选项

## API 接口

### Init 函数
```go
func Init(level string, dev bool) (*zap.Logger, error)
```

初始化全局日志器。

**参数:**
- `level`: 日志级别字符串 ("debug", "info", "warn", "error")
- `dev`: 开发模式标志，为 true 时启用开发者友好的配置

**返回:**
- `*zap.Logger`: 配置好的日志器实例
- `error`: 初始化错误

### L 函数
```go
func L() *zap.Logger
```

获取全局结构化日志器。如果未初始化，将使用默认的 Info 级别配置。

### S 函数
```go
func S() *zap.SugaredLogger
```

获取全局的语法糖日志器，提供更简便的日志记录接口。

### Sync 函数
```go
func Sync()
```

刷新日志缓冲区，确保所有日志都已写入。通常在程序退出时调用。

## 使用示例

### 基础使用
```go
package main

import (
    "log"
    
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

func main() {
    // 初始化日志器
    _, err := logger.Init("info", false)
    if err != nil {
        log.Fatalf("日志器初始化失败: %v", err)
    }
    
    // 确保程序退出时刷新日志缓冲区
    defer logger.Sync()
    
    // 使用结构化日志器
    logger.L().Info("应用程序启动",
        zap.String("version", "1.0.0"),
        zap.Int("port", 8080),
    )
    
    logger.L().Error("发生错误",
        zap.String("component", "database"),
        zap.Error(err),
    )
    
    // 使用语法糖日志器
    logger.S().Infof("用户 %s 登录成功", "alice")
    logger.S().Warnw("内存使用率较高",
        "usage", 85.5,
        "threshold", 80.0,
    )
}
```

### 开发模式配置
```go
func setupDevelopmentLogger() error {
    // 开发模式：启用彩色日志、调用者信息等
    _, err := logger.Init("debug", true)
    if err != nil {
        return fmt.Errorf("开发日志器初始化失败: %w", err)
    }
    
    logger.L().Debug("开发模式日志器已启用")
    return nil
}
```

### 生产模式配置
```go
func setupProductionLogger() error {
    // 生产模式：JSON格式，Info级别
    _, err := logger.Init("info", false)
    if err != nil {
        return fmt.Errorf("生产日志器初始化失败: %w", err)
    }
    
    logger.L().Info("生产模式日志器已启用")
    return nil
}
```

### 智能体日志记录
```go
package agents

import (
    "context"
    "time"
    
    "go.uber.org/zap"
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

type ChatAgent struct {
    ID   string
    Name string
}

func (a *ChatAgent) Process(ctx context.Context, input string) (string, error) {
    start := time.Now()
    
    // 记录处理开始
    logger.L().Info("智能体开始处理请求",
        zap.String("agent_id", a.ID),
        zap.String("agent_name", a.Name),
        zap.String("input_preview", input[:min(100, len(input))]),
        zap.Int("input_length", len(input)),
    )
    
    // 模拟处理逻辑
    response, err := a.processInternal(ctx, input)
    
    duration := time.Since(start)
    
    if err != nil {
        // 记录错误
        logger.L().Error("智能体处理失败",
            zap.String("agent_id", a.ID),
            zap.String("agent_name", a.Name),
            zap.Error(err),
            zap.Duration("duration", duration),
        )
        return "", err
    }
    
    // 记录成功处理
    logger.L().Info("智能体处理完成",
        zap.String("agent_id", a.ID),
        zap.String("agent_name", a.Name),
        zap.Int("response_length", len(response)),
        zap.Duration("duration", duration),
    )
    
    return response, nil
}

func (a *ChatAgent) processInternal(ctx context.Context, input string) (string, error) {
    // 实际处理逻辑
    return "处理结果", nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### HTTP 服务器日志中间件
```go
package middleware

import (
    "net/http"
    "time"
    
    "go.uber.org/zap"
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // 创建响应记录器
        recorder := &responseRecorder{
            ResponseWriter: w,
            statusCode:     http.StatusOK,
        }
        
        // 记录请求开始
        logger.L().Info("HTTP请求开始",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("remote_addr", r.RemoteAddr),
            zap.String("user_agent", r.UserAgent()),
        )
        
        // 处理请求
        next.ServeHTTP(recorder, r)
        
        duration := time.Since(start)
        
        // 记录请求完成
        logger.L().Info("HTTP请求完成",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.Int("status_code", recorder.statusCode),
            zap.Duration("duration", duration),
            zap.Int64("response_size", recorder.size),
        )
    })
}

type responseRecorder struct {
    http.ResponseWriter
    statusCode int
    size       int64
}

func (r *responseRecorder) WriteHeader(statusCode int) {
    r.statusCode = statusCode
    r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
    size, err := r.ResponseWriter.Write(data)
    r.size += int64(size)
    return size, err
}
```

### 数据库操作日志
```go
package database

import (
    "context"
    "time"
    
    "go.uber.org/zap"
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

type DatabaseLogger struct{}

func (l *DatabaseLogger) LogQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) {
    if err != nil {
        logger.L().Error("数据库查询失败",
            zap.String("query", query),
            zap.Any("args", args),
            zap.Duration("duration", duration),
            zap.Error(err),
        )
    } else {
        logger.L().Debug("数据库查询成功",
            zap.String("query", query),
            zap.Any("args", args),
            zap.Duration("duration", duration),
        )
    }
}

func (l *DatabaseLogger) LogTransaction(ctx context.Context, operation string, duration time.Duration, err error) {
    fields := []zap.Field{
        zap.String("operation", operation),
        zap.Duration("duration", duration),
    }
    
    if err != nil {
        logger.L().Error("数据库事务失败", append(fields, zap.Error(err))...)
    } else {
        logger.L().Info("数据库事务完成", fields...)
    }
}
```

### 异步任务日志
```go
package tasks

import (
    "context"
    "fmt"
    "time"
    
    "go.uber.org/zap"
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

type TaskRunner struct {
    ID   string
    Name string
}

func (r *TaskRunner) RunTask(ctx context.Context, taskID string, payload interface{}) error {
    // 创建任务专用的日志字段
    taskFields := []zap.Field{
        zap.String("task_runner_id", r.ID),
        zap.String("task_runner_name", r.Name),
        zap.String("task_id", taskID),
    }
    
    logger.L().Info("任务开始执行", taskFields...)
    
    start := time.Now()
    
    defer func() {
        if r := recover(); r != nil {
            logger.L().Error("任务执行发生panic",
                append(taskFields,
                    zap.Any("panic", r),
                    zap.Duration("duration", time.Since(start)),
                )...,
            )
            panic(r) // 重新抛出panic
        }
    }()
    
    // 执行任务
    err := r.executeTask(ctx, payload)
    
    duration := time.Since(start)
    
    if err != nil {
        logger.L().Error("任务执行失败",
            append(taskFields,
                zap.Error(err),
                zap.Duration("duration", duration),
            )...,
        )
        return fmt.Errorf("任务 %s 执行失败: %w", taskID, err)
    }
    
    logger.L().Info("任务执行成功",
        append(taskFields,
            zap.Duration("duration", duration),
        )...,
    )
    
    return nil
}

func (r *TaskRunner) executeTask(ctx context.Context, payload interface{}) error {
    // 实际任务执行逻辑
    time.Sleep(100 * time.Millisecond) // 模拟工作
    return nil
}
```

## 高级配置

### 自定义日志器配置
```go
package logger

import (
    "os"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func InitWithCustomConfig(config *CustomLoggerConfig) (*zap.Logger, error) {
    var cores []zapcore.Core
    
    // 控制台输出核心
    if config.EnableConsole {
        consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
        consoleCore := zapcore.NewCore(
            consoleEncoder,
            zapcore.AddSync(os.Stdout),
            config.ConsoleLevel,
        )
        cores = append(cores, consoleCore)
    }
    
    // 文件输出核心
    if config.EnableFile && config.LogFile != "" {
        file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return nil, err
        }
        
        fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
        fileCore := zapcore.NewCore(
            fileEncoder,
            zapcore.AddSync(file),
            config.FileLevel,
        )
        cores = append(cores, fileCore)
    }
    
    // 组合多个核心
    core := zapcore.NewTee(cores...)
    
    // 创建日志器
    logger := zap.New(core)
    
    if config.AddCaller {
        logger = logger.WithOptions(zap.AddCaller())
    }
    
    if config.AddStacktrace {
        logger = logger.WithOptions(zap.AddStacktrace(zapcore.ErrorLevel))
    }
    
    return logger, nil
}

type CustomLoggerConfig struct {
    EnableConsole   bool
    ConsoleLevel    zapcore.Level
    EnableFile      bool
    LogFile         string
    FileLevel       zapcore.Level
    AddCaller       bool
    AddStacktrace   bool
}
```

### 日志轮转配置
```go
package logger

import (
    "gopkg.in/natefinch/lumberjack.v2"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func InitWithRotation(config *RotationConfig) (*zap.Logger, error) {
    // 配置日志轮转
    writer := &lumberjack.Logger{
        Filename:   config.Filename,
        MaxSize:    config.MaxSize,    // MB
        MaxBackups: config.MaxBackups,
        MaxAge:     config.MaxAge,     // 天
        Compress:   config.Compress,
    }
    
    encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
    core := zapcore.NewCore(
        encoder,
        zapcore.AddSync(writer),
        config.Level,
    )
    
    logger := zap.New(core, zap.AddCaller())
    return logger, nil
}

type RotationConfig struct {
    Filename   string        // 日志文件路径
    MaxSize    int          // 单个日志文件最大大小(MB)
    MaxBackups int          // 保留的旧日志文件数量
    MaxAge     int          // 保留日志文件的最大天数
    Compress   bool         // 是否压缩旧日志文件
    Level      zapcore.Level // 日志级别
}
```

### 上下文日志记录
```go
package logger

import (
    "context"
    
    "go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "logger"

// WithLogger 将日志器添加到上下文中
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

// FromContext 从上下文中获取日志器
func FromContext(ctx context.Context) *zap.Logger {
    if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
        return logger
    }
    return L() // 返回全局日志器
}

// WithFields 为上下文日志器添加字段
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
    logger := FromContext(ctx).With(fields...)
    return WithLogger(ctx, logger)
}

// 使用示例
func handleRequest(ctx context.Context, userID, requestID string) error {
    // 为请求添加上下文字段
    ctx = WithFields(ctx,
        zap.String("user_id", userID),
        zap.String("request_id", requestID),
    )
    
    // 使用上下文日志器
    FromContext(ctx).Info("开始处理请求")
    
    if err := processRequest(ctx); err != nil {
        FromContext(ctx).Error("处理请求失败", zap.Error(err))
        return err
    }
    
    FromContext(ctx).Info("请求处理完成")
    return nil
}

func processRequest(ctx context.Context) error {
    // 子函数也可以使用相同的上下文日志器
    FromContext(ctx).Debug("执行具体处理逻辑")
    return nil
}
```

## 性能优化

### 字段重用
```go
// 预定义常用字段，避免重复创建
var (
    componentField = zap.String("component", "agent_service")
    versionField   = zap.String("version", "1.0.0")
)

func logWithReusedFields() {
    logger.L().Info("服务启动", componentField, versionField)
}
```

### 条件日志记录
```go
func expensiveOperation() {
    // 检查日志级别，避免不必要的计算
    if logger.L().Core().Enabled(zap.DebugLevel) {
        expensiveData := computeExpensiveDebugData()
        logger.L().Debug("调试信息", zap.Any("data", expensiveData))
    }
}

func computeExpensiveDebugData() interface{} {
    // 耗时的数据计算
    return map[string]interface{}{
        "detailed_info": "expensive computation result",
    }
}
```

## 监控和告警

### 错误日志统计
```go
package monitoring

import (
    "sync"
    "time"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "github.com/nvcnvn/adk-golang/pkg/logger"
)

type ErrorCounter struct {
    mu     sync.RWMutex
    errors map[string]int
}

func NewErrorCounter() *ErrorCounter {
    return &ErrorCounter{
        errors: make(map[string]int),
    }
}

func (c *ErrorCounter) IncrementError(component string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.errors[component]++
}

func (c *ErrorCounter) GetErrorCounts() map[string]int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    result := make(map[string]int)
    for k, v := range c.errors {
        result[k] = v
    }
    return result
}

// 自定义钩子，统计错误日志
type ErrorCountingHook struct {
    counter *ErrorCounter
}

func (h *ErrorCountingHook) Write(entry zapcore.Entry, fields []zapcore.Field) error {
    if entry.Level >= zapcore.ErrorLevel {
        component := "unknown"
        for _, field := range fields {
            if field.Key == "component" {
                component = field.String
                break
            }
        }
        h.counter.IncrementError(component)
    }
    return nil
}
```

## 最佳实践

1. **初始化**: 在应用启动时初始化日志器，并在退出时调用 Sync()
2. **级别选择**: 开发环境使用 Debug 级别，生产环境使用 Info 级别
3. **结构化**: 使用结构化字段而非格式化字符串
4. **性能**: 避免在热路径中记录过多日志
5. **上下文**: 利用上下文传递请求相关的日志字段
6. **错误处理**: 确保日志记录不会影响业务逻辑的错误处理

## 依赖模块

- `go.uber.org/zap`: 高性能日志库
- `go.uber.org/zap/zapcore`: zap 核心组件
- Go 标准库: `sync`

Logger 模块为 ADK-Golang 框架提供了强大的日志记录能力，是系统监控、问题诊断和性能分析的重要工具。
