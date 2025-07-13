# Code Executors 代码执行器模块

## 概述

Code Executors 模块提供了安全的代码执行功能，支持多种编程语言的代码片段执行。该模块是智能体代码生成和执行能力的核心组件，提供了从本地执行到容器化执行的完整解决方案。

## 核心组件

### 基础接口

#### CodeExecutor 接口
```go
type CodeExecutor interface {
    // Execute 执行给定的代码并返回结果
    Execute(ctx context.Context, code string, files []File) (*ExecutionResult, error)
}
```

核心代码执行接口，定义了统一的代码执行标准。

#### File 结构
```go
type File struct {
    Name    string  // 文件名
    Content []byte  // 文件内容
}
```

表示执行过程中的输入或输出文件。

#### ExecutionResult 结构
```go
type ExecutionResult struct {
    Stdout      string // 标准输出
    Stderr      string // 标准错误输出
    OutputFiles []File // 输出文件列表
}
```

代码执行的结果封装，包含完整的执行信息。

## 执行器实现类型

### 1. PythonExecutor (Python执行器)
专门用于执行Python代码的执行器：
- 支持完整的Python语法
- 自动处理依赖包安装
- 支持文件输入输出
- 收集执行过程中生成的文件

### 2. JavaScriptExecutor (JavaScript执行器)
基于Node.js的JavaScript代码执行器：
- 支持现代JavaScript语法
- 支持npm包导入
- 支持异步代码执行
- 自动收集输出文件

### 3. BaseExecutor (基础执行器)
提供通用的执行功能：
- 临时目录管理
- 文件保存和清理
- 公共执行逻辑

### 4. UnsafeLocalCodeExecutor (不安全本地执行器)
直接在本地环境执行代码，适用于开发和测试：
- ⚠️ 无安全隔离，仅限开发环境使用
- 高性能执行
- 直接访问本地文件系统

### 5. ContainerCodeExecutor (容器执行器)
基于容器的安全代码执行：
- Docker容器隔离
- 安全的执行环境
- 资源限制控制
- 适合生产环境

### 6. VertexAICodeExecutor (Vertex AI执行器)
基于Google Vertex AI的云端代码执行：
- 云端计算资源
- 企业级安全性
- 自动扩展能力
- 集成AI分析功能

## 使用示例

### 基本Python代码执行
```go
package main

import (
    "context"
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/code_executors"
)

func main() {
    ctx := context.Background()
    
    // 创建Python执行器
    executor, err := code_executors.NewPythonExecutor()
    if err != nil {
        fmt.Printf("创建Python执行器失败: %v\n", err)
        return
    }
    defer executor.Cleanup()
    
    // 执行Python代码
    code := `
import math

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# 计算前10个斐波那契数
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")

# 计算圆的面积
radius = 5
area = math.pi * radius ** 2
print(f"半径为{radius}的圆的面积: {area:.2f}")
`
    
    result, err := executor.Execute(ctx, code, nil)
    if err != nil {
        fmt.Printf("代码执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("执行结果:\n%s\n", result.Stdout)
    if result.Stderr != "" {
        fmt.Printf("错误输出:\n%s\n", result.Stderr)
    }
}
```

### JavaScript代码执行
```go
func executeJavaScript() {
    ctx := context.Background()
    
    // 创建JavaScript执行器
    executor, err := code_executors.NewJavaScriptExecutor()
    if err != nil {
        fmt.Printf("创建JavaScript执行器失败: %v\n", err)
        return
    }
    defer executor.Cleanup()
    
    // 执行JavaScript代码
    code := `
// 异步函数示例
async function fetchData() {
    return new Promise(resolve => {
        setTimeout(() => {
            resolve("数据加载完成");
        }, 1000);
    });
}

// 数组操作
const numbers = [1, 2, 3, 4, 5];
const doubled = numbers.map(n => n * 2);
console.log("原数组:", numbers);
console.log("翻倍后:", doubled);

// 使用异步函数
(async () => {
    const data = await fetchData();
    console.log(data);
    
    // JSON处理
    const obj = { name: "测试", value: 42 };
    console.log("JSON字符串:", JSON.stringify(obj));
})();
`
    
    result, err := executor.Execute(ctx, code, nil)
    if err != nil {
        fmt.Printf("代码执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("执行结果:\n%s\n", result.Stdout)
}
```

### 带文件输入的代码执行
```go
func executeWithFiles() {
    ctx := context.Background()
    
    executor, err := code_executors.NewPythonExecutor()
    if err != nil {
        fmt.Printf("创建执行器失败: %v\n", err)
        return
    }
    defer executor.Cleanup()
    
    // 准备输入文件
    inputFiles := []code_executors.File{
        {
            Name:    "data.txt",
            Content: []byte("1,2,3,4,5\n6,7,8,9,10\n11,12,13,14,15"),
        },
        {
            Name:    "config.json",
            Content: []byte(`{"delimiter": ",", "header": false}`),
        },
    }
    
    // 处理文件的Python代码
    code := `
import json
import csv

# 读取配置
with open('config.json', 'r') as f:
    config = json.load(f)

# 读取数据文件
data = []
with open('data.txt', 'r') as f:
    reader = csv.reader(f, delimiter=config['delimiter'])
    for row in reader:
        data.append([int(x) for x in row])

print("读取的数据:")
for i, row in enumerate(data):
    print(f"行 {i+1}: {row}")

# 计算统计信息
all_numbers = [num for row in data for num in row]
print(f"总数: {len(all_numbers)}")
print(f"最小值: {min(all_numbers)}")
print(f"最大值: {max(all_numbers)}")
print(f"平均值: {sum(all_numbers) / len(all_numbers):.2f}")

# 保存处理结果
result = {
    "total_count": len(all_numbers),
    "min_value": min(all_numbers),
    "max_value": max(all_numbers),
    "average": sum(all_numbers) / len(all_numbers)
}

with open('result.json', 'w') as f:
    json.dump(result, f, indent=2)

print("结果已保存到 result.json")
`
    
    result, err := executor.Execute(ctx, code, inputFiles)
    if err != nil {
        fmt.Printf("代码执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("执行结果:\n%s\n", result.Stdout)
    
    // 处理输出文件
    for _, file := range result.OutputFiles {
        fmt.Printf("输出文件 %s:\n%s\n", file.Name, string(file.Content))
    }
}
```

## 高级功能

### 执行器工厂
```go
func createExecutorByLanguage(language string) (code_executors.CodeExecutor, error) {
    executor, err := code_executors.NewCodeExecutor(language)
    if err != nil {
        return nil, fmt.Errorf("创建%s执行器失败: %w", language, err)
    }
    return executor, nil
}

// 使用示例
func executeCodeByLanguage(language, code string) {
    ctx := context.Background()
    
    executor, err := createExecutorByLanguage(language)
    if err != nil {
        fmt.Printf("创建执行器失败: %v\n", err)
        return
    }
    
    // 如果执行器有清理方法，调用它
    if cleaner, ok := executor.(interface{ Cleanup() error }); ok {
        defer cleaner.Cleanup()
    }
    
    result, err := executor.Execute(ctx, code, nil)
    if err != nil {
        fmt.Printf("代码执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("执行结果:\n%s\n", result.Stdout)
}
```

### 容器化执行
```go
func executeInContainer() {
    ctx := context.Background()
    
    // 创建容器执行器
    executor, err := code_executors.NewContainerCodeExecutor(&code_executors.ContainerConfig{
        Image: "python:3.9-slim",
        WorkDir: "/workspace",
        Memory: "512m",
        CPUs: "0.5",
        Timeout: 30 * time.Second,
    })
    if err != nil {
        fmt.Printf("创建容器执行器失败: %v\n", err)
        return
    }
    defer executor.Cleanup()
    
    code := `
import sys
import os
print(f"Python版本: {sys.version}")
print(f"工作目录: {os.getcwd()}")
print(f"环境变量: {dict(os.environ)}")

# 一些计算密集型任务
import time
start = time.time()
result = sum(i**2 for i in range(100000))
end = time.time()
print(f"计算结果: {result}")
print(f"耗时: {end - start:.4f}秒")
`
    
    result, err := executor.Execute(ctx, code, nil)
    if err != nil {
        fmt.Printf("容器执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("容器执行结果:\n%s\n", result.Stdout)
}
```

### 批量代码执行
```go
type CodeTask struct {
    Language string
    Code     string
    Files    []code_executors.File
}

func executeBatchCodes(tasks []CodeTask) {
    ctx := context.Background()
    
    for i, task := range tasks {
        fmt.Printf("执行任务 %d/%d (语言: %s)\n", i+1, len(tasks), task.Language)
        
        executor, err := code_executors.NewCodeExecutor(task.Language)
        if err != nil {
            fmt.Printf("任务 %d 创建执行器失败: %v\n", i+1, err)
            continue
        }
        
        result, err := executor.Execute(ctx, task.Code, task.Files)
        if err != nil {
            fmt.Printf("任务 %d 执行失败: %v\n", i+1, err)
        } else {
            fmt.Printf("任务 %d 执行成功:\n%s\n", i+1, result.Stdout)
        }
        
        // 清理资源
        if cleaner, ok := executor.(interface{ Cleanup() error }); ok {
            cleaner.Cleanup()
        }
        
        fmt.Println("---")
    }
}

// 使用示例
func batchExecutionExample() {
    tasks := []CodeTask{
        {
            Language: "python",
            Code:     "print('Hello from Python!')\nprint(2 + 3)",
        },
        {
            Language: "javascript",
            Code:     "console.log('Hello from JavaScript!'); console.log(2 + 3);",
        },
    }
    
    executeBatchCodes(tasks)
}
```

## 安全性和最佳实践

### 安全执行配置
```go
type SafeExecutionConfig struct {
    Timeout      time.Duration
    MemoryLimit  string
    CPULimit     string
    NetworkAccess bool
    FileSystemAccess bool
    AllowedPackages []string
}

func createSafeExecutor(config *SafeExecutionConfig) code_executors.CodeExecutor {
    // 生产环境建议使用容器执行器
    containerConfig := &code_executors.ContainerConfig{
        Image:   "python:3.9-slim",
        Memory:  config.MemoryLimit,
        CPUs:    config.CPULimit,
        Timeout: config.Timeout,
        NetworkAccess: config.NetworkAccess,
        ReadOnlyRootFS: !config.FileSystemAccess,
    }
    
    executor, err := code_executors.NewContainerCodeExecutor(containerConfig)
    if err != nil {
        // 降级到安全的本地执行器
        return code_executors.NewSafeLocalExecutor(config)
    }
    
    return executor
}
```

### 错误处理和重试
```go
func executeWithRetry(executor code_executors.CodeExecutor, code string, maxRetries int) (*code_executors.ExecutionResult, error) {
    ctx := context.Background()
    
    var lastErr error
    for attempt := 1; attempt <= maxRetries; attempt++ {
        fmt.Printf("执行尝试 %d/%d\n", attempt, maxRetries)
        
        result, err := executor.Execute(ctx, code, nil)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        fmt.Printf("尝试 %d 失败: %v\n", attempt, err)
        
        if attempt < maxRetries {
            // 指数退避
            waitTime := time.Duration(attempt*attempt) * time.Second
            fmt.Printf("等待 %v 后重试...\n", waitTime)
            time.Sleep(waitTime)
        }
    }
    
    return nil, fmt.Errorf("经过 %d 次尝试后仍然失败，最后错误: %w", maxRetries, lastErr)
}
```

### 执行监控
```go
type ExecutionMonitor struct {
    mu          sync.RWMutex
    executions  map[string]*ExecutionStats
}

type ExecutionStats struct {
    TotalRuns    int64         `json:"total_runs"`
    SuccessRuns  int64         `json:"success_runs"`
    FailureRuns  int64         `json:"failure_runs"`
    AvgDuration  time.Duration `json:"avg_duration"`
    LastExecution time.Time    `json:"last_execution"`
}

func NewExecutionMonitor() *ExecutionMonitor {
    return &ExecutionMonitor{
        executions: make(map[string]*ExecutionStats),
    }
}

func (m *ExecutionMonitor) RecordExecution(language string, duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if _, exists := m.executions[language]; !exists {
        m.executions[language] = &ExecutionStats{}
    }
    
    stats := m.executions[language]
    stats.TotalRuns++
    stats.LastExecution = time.Now()
    
    if success {
        stats.SuccessRuns++
    } else {
        stats.FailureRuns++
    }
    
    // 更新平均执行时间
    if stats.TotalRuns == 1 {
        stats.AvgDuration = duration
    } else {
        stats.AvgDuration = time.Duration(
            (int64(stats.AvgDuration)*(stats.TotalRuns-1) + int64(duration)) / stats.TotalRuns,
        )
    }
}

func (m *ExecutionMonitor) GetStats(language string) *ExecutionStats {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if stats, exists := m.executions[language]; exists {
        return stats
    }
    return nil
}
```

## 配置管理

### 执行器配置
```go
type ExecutorConfig struct {
    Type         string                 `yaml:"type"`         // python/javascript/container
    Timeout      time.Duration          `yaml:"timeout"`
    WorkDir      string                 `yaml:"work_dir"`
    Environment  map[string]string      `yaml:"environment"`
    Container    *ContainerConfig       `yaml:"container,omitempty"`
    Security     *SecurityConfig        `yaml:"security,omitempty"`
}

type ContainerConfig struct {
    Image      string `yaml:"image"`
    Memory     string `yaml:"memory"`
    CPUs       string `yaml:"cpus"`
    NetworkMode string `yaml:"network_mode"`
}

type SecurityConfig struct {
    AllowNetworkAccess bool     `yaml:"allow_network_access"`
    AllowFileAccess    bool     `yaml:"allow_file_access"`
    RestrictedPackages []string `yaml:"restricted_packages"`
}

func LoadExecutorConfig(configPath string) (*ExecutorConfig, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config ExecutorConfig
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### 配置文件示例
```yaml
# executor_config.yaml
type: "container"
timeout: "30s"
work_dir: "/workspace"
environment:
  PYTHONPATH: "/workspace"
  NODE_PATH: "/workspace/node_modules"

container:
  image: "python:3.9-slim"
  memory: "512m"
  cpus: "0.5"
  network_mode: "none"

security:
  allow_network_access: false
  allow_file_access: true
  restricted_packages:
    - "os"
    - "subprocess"
    - "sys"
```

## 最佳实践

1. **安全性**: 生产环境必须使用容器执行器或其他隔离机制
2. **资源管理**: 设置合理的超时时间和资源限制
3. **错误处理**: 实现完善的错误处理和重试机制
4. **监控日志**: 记录执行统计信息，便于性能分析
5. **文件管理**: 及时清理临时文件和执行环境
6. **并发控制**: 控制同时执行的代码数量，避免资源耗尽

## 依赖模块

- `github.com/nvcnvn/adk-golang/pkg/events`: 事件系统
- `github.com/nvcnvn/adk-golang/pkg/telemetry`: 遥测监控
- Go 标准库: `context`, `os/exec`, `io`, `os`

## 扩展开发

### 自定义执行器
```go
type CustomExecutor struct {
    *code_executors.BaseExecutor
    runtime string
}

func NewCustomExecutor(runtime string) (*CustomExecutor, error) {
    base, err := code_executors.NewBaseExecutor()
    if err != nil {
        return nil, err
    }
    
    return &CustomExecutor{
        BaseExecutor: base,
        runtime:      runtime,
    }, nil
}

func (e *CustomExecutor) Execute(ctx context.Context, code string, files []code_executors.File) (*code_executors.ExecutionResult, error) {
    // 自定义执行逻辑
    return &code_executors.ExecutionResult{
        Stdout: "自定义执行器输出",
        Stderr: "",
        OutputFiles: nil,
    }, nil
}
```

Code Executors 模块为 ADK-Golang 框架提供了强大而安全的代码执行能力，是构建智能代码生成和执行系统的核心基础设施。
