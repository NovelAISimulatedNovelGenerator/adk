package scheduler_test

import (
    "context"
    "sync"
    "testing"
    "time"

    "github.com/nvcnvn/adk-golang/pkg/scheduler"
)

// TestWorkerPoolSchedulerConcurrency 验证 worker pool 调度器在高并发场景下能正确执行所有任务且无死锁/超时。
func TestWorkerPoolSchedulerConcurrency(t *testing.T) {
    workers := 8          // 与 HttpServer 默认配置保持一致
    taskCount := 1000     // 并发任务数量
    queueSize := taskCount // 队列大小设置为任务数，避免 Submit 时队列溢出

    // Processor 模拟轻量级耗时操作
    proc := func(ctx context.Context, task *scheduler.Task) (string, error) {
        // 模拟处理耗时 1ms
        time.Sleep(1 * time.Millisecond)
        return task.Input + "_done", nil
    }

    sched := scheduler.NewWorkerPoolScheduler(workers, queueSize, proc)
    sched.Start()
    defer sched.Stop()

    var wg sync.WaitGroup
    wg.Add(taskCount)

    // 提交 taskCount 个任务，队列容量已足够大，确保不会返回 ErrQueueFull
    for i := 0; i < taskCount; i++ {
        tsk := &scheduler.Task{
            Ctx:        context.Background(),
            Workflow:   "bench_flow",
            Input:      "hello",
            ResultChan: make(chan scheduler.Result, 1),
        }
        if err := sched.Submit(tsk); err != nil {
            t.Fatalf("Submit error: %v", err)
        }

        // 异步等待结果
        go func(task *scheduler.Task) {
            defer wg.Done()
            res := <-task.ResultChan
            if res.Err != nil {
                t.Errorf("Task error: %v", res.Err)
            }
            if res.Output != "hello_done" {
                t.Errorf("Unexpected output: %s", res.Output)
            }
        }(tsk)
    }

    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    // 设置整体超时，防止死锁
    select {
    case <-done:
        // ok
    case <-time.After(10 * time.Second):
        t.Fatalf("timeout: not all tasks finished in time")
    }
}

// BenchmarkWorkerPoolScheduler 基准测试，用于测量调度器吞吐。
func BenchmarkWorkerPoolScheduler(b *testing.B) {
    workers := 8
    queueSize := 32

    proc := func(ctx context.Context, task *scheduler.Task) (string, error) {
        return "ok", nil
    }

    sched := scheduler.NewWorkerPoolScheduler(workers, queueSize, proc)
    sched.Start()
    defer sched.Stop()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tsk := &scheduler.Task{
            Ctx:        context.Background(),
            Workflow:   "bench_flow",
            Input:      "",
            ResultChan: make(chan scheduler.Result, 1),
        }
        if err := sched.Submit(tsk); err != nil {
            b.Fatalf("submit error: %v", err)
        }
        <-tsk.ResultChan
    }
}
