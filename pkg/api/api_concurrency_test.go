package api

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync"
    "sync/atomic"
    "testing"
    "time"

    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
)

// TestExecuteEndpointConcurrency 启动内存级 HTTP 服务器并并发调用 /api/execute。
func TestExecuteEndpointConcurrency(t *testing.T) {
    taskCount := 500         // 请求并发量
    workers := 8             // 与默认 HttpServer 保持一致
    queueSize := 64          // 足够大
    _ = workers               // 保持一致性（目前不需要调整构造函数）
    _ = queueSize

    // 创建 bench_agent
    benchAgent := agents.NewAgent(
        agents.WithName("bench_agent"),
        agents.WithInstruction("并发测试 Echo Agent"),
        agents.WithDescription("仅用于并发测试，快速返回 OK"),
        agents.WithBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
            return "OK", true
        }),
    )

    // Manager 并注册
    mgr := flow.NewManager()
    mgr.Register("bench_flow", benchAgent)

    // 创建 HTTP Server 实例（内部会启动 scheduler）
    httpSrv := NewHttpServer(mgr, ":0")

    // 仅构建 handler，不调用 Start()，使用 httptest Server
    mux := http.NewServeMux()
    mux.HandleFunc("/api/execute", httpSrv.handleExecute)
    testServer := httptest.NewServer(mux)
    defer testServer.Close()
    // 关闭 scheduler
    defer httpSrv.sched.Stop()

    client := &http.Client{Timeout: 5 * time.Second}

    var wg sync.WaitGroup
    wg.Add(taskCount)
    var errCnt atomic.Int64

    for i := 0; i < taskCount; i++ {
        go func() {
            defer wg.Done()
            body, _ := json.Marshal(map[string]interface{}{
                "workflow":   "bench_flow",
                "input":      "hello",
                "user_id":    "test_user_123",
                "archive_id": "test_archive_456",
            })
            resp, err := client.Post(testServer.URL+"/api/execute", "application/json", bytes.NewReader(body))
            if err != nil {
                errCnt.Add(1)
                return
            }
            defer resp.Body.Close()
            if resp.StatusCode != http.StatusOK {
                errCnt.Add(1)
                return
            }
            var apiResp struct {
                Success bool   `json:"success"`
                Output  string `json:"output"`
            }
            if json.NewDecoder(resp.Body).Decode(&apiResp) != nil {
                errCnt.Add(1)
                return
            }
            if !apiResp.Success || apiResp.Output != "OK" {
                errCnt.Add(1)
            }
        }()
    }

    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // ok
    case <-time.After(15 * time.Second):
        t.Fatalf("timeout waiting for API responses")
    }

    if errCnt.Load() != 0 {
        t.Fatalf("%d requests failed", errCnt.Load())
    }
}
