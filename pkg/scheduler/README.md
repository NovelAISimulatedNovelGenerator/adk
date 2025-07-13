# Scheduler 调度器模块

## 概述

Scheduler 模块提供了任务调度和执行的核心功能，实现了基于 goroutine 池的异步任务处理机制。该模块主要用于管理工作流任务的执行队列，支持并发处理和上下文取消。

## 核心组件

### Task (任务)
```go
type Task struct {
    Ctx      context.Context // 上下文，用于取消
    Workflow string          // 工作流名称
    Input    string          // 原始输入
    UserID   string          // 用户标识
    ResultChan chan Result   // 返回结果
}
```

任务结构体包含执行所需的所有信息：
- **Ctx**: 用于控制任务取消和超时
- **Workflow**: 指定要执行的工作流名称
- **Input**: 传递给工作流的原始输入数据
- **UserID**: 用户标识，支持多用户场景
- **ResultChan**: 结果通道，用于异步接收执行结果

### Result (结果)
```go
type Result struct {
    Output string
    Err    error
}
```

封装任务执行结果，包含输出内容和可能的错误信息。

### Processor (处理器)
```go
type Processor func(ctx context.Context, task *Task) (string, error)
```

处理器函数类型，定义了具体的任务执行逻辑。将任务交给相应的工作流或智能体进行处理。

### Scheduler (调度器接口)
```go
type Scheduler interface {
    Submit(task *Task) error  // 提交任务到队列
    Start()                   // 启动调度器
    Stop()                    // 停止调度器
}
```

调度器核心接口，定义了任务提交、启动和停止的方法。

## 实现

### WorkerPoolScheduler
基于 goroutine 池的调度器实现：

```go
func NewWorkerPoolScheduler(workers, queueSize int, p Processor) Scheduler
```

**参数说明:**
- `workers`: 工作线程数量
- `queueSize`: 任务队列大小（建议 >= workers*2）
- `p`: 任务处理器函数

**特性:**
- 固定数量的 worker goroutines
- 带缓冲的任务队列
- 优雅关闭机制
- 队列满时返回 `ErrQueueFull` 错误

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/scheduler"
)

func main() {
    // 创建处理器
    processor := func(ctx context.Context, task *scheduler.Task) (string, error) {
        // 实际的工作流执行逻辑
        return fmt.Sprintf("处理工作流 %s，输入: %s", task.Workflow, task.Input), nil
    }

    // 创建调度器
    s := scheduler.NewWorkerPoolScheduler(4, 10, processor)
    
    // 启动调度器
    s.Start()
    defer s.Stop()

    // 创建任务
    resultChan := make(chan scheduler.Result, 1)
    task := &scheduler.Task{
        Ctx:        context.Background(),
        Workflow:   "example-workflow",
        Input:      "测试输入",
        UserID:     "user123",
        ResultChan: resultChan,
    }

    // 提交任务
    if err := s.Submit(task); err != nil {
        fmt.Printf("提交任务失败: %v\n", err)
        return
    }

    // 等待结果
    result := <-resultChan
    if result.Err != nil {
        fmt.Printf("任务执行失败: %v\n", result.Err)
    } else {
        fmt.Printf("任务执行成功: %s\n", result.Output)
    }
}
```

## 错误处理

- **ErrQueueFull**: 当任务队列已满时返回，调用方可以选择重试或丢弃任务
- **Context 取消**: 支持通过 context 取消正在执行的任务
- **优雅关闭**: Stop() 方法会等待所有正在执行的任务完成

## 最佳实践

1. **队列大小设置**: 建议队列大小至少为 worker 数量的 2 倍
2. **错误处理**: 始终检查 Submit 返回的错误
3. **资源清理**: 使用 defer 确保调用 Stop() 方法
4. **上下文管理**: 合理设置 context 超时时间
5. **结果通道**: 确保 ResultChan 有足够的缓冲区

## 依赖

- Go 标准库: `context`, `sync`, `errors`
- 无外部依赖

## 测试

运行测试：
```bash
go test ./pkg/scheduler
```

该模块包含完整的单元测试，覆盖正常流程、错误场景和并发安全性。
