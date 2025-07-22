package scheduler

import (
    "context"
    "errors"
    "sync"
)

// Task 代表一次工作流执行任务。
// ResultChan 必须非 nil，调度器完成后会写入结果。
// Cancel 方法由调用方传入 context 控制。
type Task struct {
    Ctx       context.Context // 上下文，用于取消
    Workflow  string          // 工作流名称
    Input     string          // 原始输入
    UserID    string          // 用户标识
    ArchiveID string          // 归档标识符

    ResultChan chan Result // 返回结果
}

// Result 执行结果或错误
// 如果 Err 非 nil，Output 可能为空。
type Result struct {
    Output string
    Err    error
}

// Processor 是具体执行逻辑，将任务交给工作流/agent 并返回输出。
type Processor func(ctx context.Context, task *Task) (string, error)

// Scheduler 调度器接口
// Submit 将任务放入队列，若队列已满则返回 error。
// Stop 会阻塞直到所有 worker 退出。
type Scheduler interface {
    Submit(task *Task) error
    Start()
    Stop()
}

// ErrQueueFull 当队列已满时返回。
var ErrQueueFull = errors.New("task queue is full")

// workerPoolScheduler 简单 goroutine 池实现。

type workerPoolScheduler struct {
    tasks      chan *Task
    workers    int
    processor  Processor

    wg   sync.WaitGroup
    once sync.Once
    quit chan struct{}
}

// NewWorkerPoolScheduler 创建调度器。
// queueSize 建议 >= workers*2
func NewWorkerPoolScheduler(workers, queueSize int, p Processor) Scheduler {
    if workers <= 0 {
        workers = 4
    }
    if queueSize <= 0 {
        queueSize = workers * 2
    }
    return &workerPoolScheduler{
        tasks:     make(chan *Task, queueSize),
        workers:   workers,
        processor: p,
        quit:      make(chan struct{}),
    }
}

func (s *workerPoolScheduler) Start() {
    s.once.Do(func() {
        for i := 0; i < s.workers; i++ {
            s.wg.Add(1)
            go s.worker()
        }
    })
}

func (s *workerPoolScheduler) Stop() {
    close(s.quit)
    s.wg.Wait()
}

func (s *workerPoolScheduler) Submit(task *Task) error {
    select {
    case s.tasks <- task:
        return nil
    default:
        return ErrQueueFull
    }
}

func (s *workerPoolScheduler) worker() {
    defer s.wg.Done()
    for {
        select {
        case <-s.quit:
            return
        case task := <-s.tasks:
            if task == nil {
                continue
            }
            output, err := s.processor(task.Ctx, task)
            select {
            case task.ResultChan <- Result{Output: output, Err: err}:
            default:
                // 调用方可能没有等待结果，但仍避免阻塞
            }
        }
    }
}
