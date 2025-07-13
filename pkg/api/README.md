# API 模块文档

## 模块概述

`github.com/nvcnvn/adk-golang/pkg/api` 是 ADK 系统的 HTTP API 服务层，提供对工作流管理与执行的 RESTful API 接口。该模块基于 Go 标准库的 `net/http`，实现了完整的工作流执行 Web 服务，支持同步执行、流式执行和实时监控等功能。

## 核心架构

### 分层设计

```
HTTP 层
├── HttpServer (HTTP 服务器)
│   ├── 路由管理
│   ├── 请求解析
│   ├── 响应编码
│   └── 错误处理
│
服务层
├── WorkflowService (工作流服务)
│   ├── 工作流执行
│   ├── 任务调度
│   ├── 状态管理
│   └── 流式处理
│
依赖层
├── flow.Manager (工作流管理器)
├── scheduler.Scheduler (任务调度器)
└── agents.Agent (智能体)
```

### 请求流程

```
客户端请求 → HTTP路由 → 参数解析 → WorkflowService → 
Scheduler → Agent执行 → 结果返回 → HTTP响应
```

## 核心组件

### 1. HttpServer（HTTP 服务器）

HTTP API 服务器的主要实现，负责处理所有 Web 层面的细节：

```go
type HttpServer struct {
    service *WorkflowService
    sched   scheduler.Scheduler
    addr    string
    server  *http.Server
}
```

**核心功能：**
- HTTP 路由管理
- 请求参数解析和验证
- 响应格式化和错误处理
- 服务生命周期管理
- 健康检查支持

**主要方法：**
- `NewHttpServer(manager, addr)` - 创建服务器实例
- `Start()` - 启动 HTTP 服务
- `Stop(ctx)` - 优雅停止服务

### 2. WorkflowService（工作流服务）

业务逻辑层的核心服务，处理工作流相关的所有业务操作：

```go
type WorkflowService struct {
    manager    *flow.Manager
    sched      scheduler.Scheduler
    activeJobs sync.Map
}
```

**核心功能：**
- 工作流执行管理
- 任务调度协调
- 活跃任务追踪
- 同步和异步执行支持
- 流式数据处理

**主要方法：**
- `Execute(ctx, req)` - 同步执行工作流
- `ExecuteStream(ctx, req, callback)` - 流式执行工作流
- `ListWorkflows()` - 列出可用工作流
- `GetWorkflowInfo(name)` - 获取工作流详情

## 数据结构

### 1. WorkflowRequest（工作流请求）

工作流执行请求的数据结构：

```go
type WorkflowRequest struct {
    Workflow     string                 `json:"workflow"`                // 工作流名称
    Input        string                 `json:"input"`                   // 输入文本
    UserId       string                 `json:"user_id"`                 // 用户标识
    ExperimentId string                 `json:"experiment_id,omitempty"` // 实验ID（可选）
    TraceId      string                 `json:"trace_id,omitempty"`      // 追踪ID（可选）
    Parameters   map[string]interface{} `json:"parameters,omitempty"`    // 额外参数
    Timeout      int                    `json:"timeout,omitempty"`       // 超时（秒）
}
```

**字段说明：**
- `Workflow`: 要执行的工作流名称，必须已注册到 flow.Manager
- `Input`: 工作流的输入文本或数据
- `UserId`: 用户标识，用于追踪和记忆系统
- `ExperimentId`: 实验标识，用于A/B测试等场景
- `TraceId`: 请求追踪标识，用于日志关联
- `Parameters`: 额外的参数字典，传递给工作流
- `Timeout`: 执行超时时间（秒），0表示无限制

### 2. WorkflowResponse（工作流响应）

工作流执行结果的数据结构：

```go
type WorkflowResponse struct {
    Workflow    string                 `json:"workflow"`           // 工作流名称
    Output      string                 `json:"output"`             // 输出文本
    Success     bool                   `json:"success"`            // 是否成功
    Message     string                 `json:"message,omitempty"`  // 消息（错误时有值）
    Metadata    map[string]interface{} `json:"metadata,omitempty"` // 元数据
    ProcessTime int64                  `json:"process_time_ms"`    // 处理时间（毫秒）
    TraceId     string                 `json:"trace_id,omitempty"` // 请求追踪ID
}
```

**字段说明：**
- `Workflow`: 对应的工作流名称
- `Output`: 工作流执行的输出结果
- `Success`: 执行是否成功的布尔标志
- `Message`: 错误信息或状态消息
- `Metadata`: 额外的元数据信息
- `ProcessTime`: 处理耗时（毫秒）
- `TraceId`: 与请求对应的追踪ID

### 3. StreamCallback（流式回调）

流式执行的回调函数类型：

```go
type StreamCallback func(data string, done bool, err error)
```

**参数说明：**
- `data`: 流式数据片段
- `done`: 是否完成标志
- `err`: 错误信息（如有）

## API 路由

### 路由总览

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 服务状态检查 |
| GET | `/api/workflows` | 列出工作流 | 获取所有可用工作流 |
| GET | `/api/workflows/{name}` | 工作流详情 | 获取特定工作流信息 |
| POST | `/api/execute` | 同步执行 | 阻塞式工作流执行 |
| POST | `/api/stream` | 流式执行 | Server-Sent Events 流式执行 |

### 1. 健康检查

**路由：** `GET /health`

**响应示例：**
```json
{
  "status": "ok",
  "timestamp": "2025-07-13T20:41:10Z",
  "version": "1.0.0"
}
```

### 2. 列出工作流

**路由：** `GET /api/workflows`

**响应示例：**
```json
{
  "workflows": [
    "novel_v4",
    "text_summarizer",
    "code_generator"
  ]
}
```

### 3. 获取工作流详情

**路由：** `GET /api/workflows/{name}`

**响应示例：**
```json
{
  "name": "novel_v4",
  "description": "小说生成工作流",
  "agent_type": "SequentialAgent",
  "sub_agents": ["architect", "writer", "librarian"],
  "tools": ["search", "memory"],
  "created_at": "2025-07-13T20:41:10Z"
}
```

### 4. 同步执行工作流

**路由：** `POST /api/execute`

**请求示例：**
```json
{
  "workflow": "novel_v4",
  "input": "写一个关于未来世界的科幻小说",
  "user_id": "user123",
  "parameters": {
    "genre": "science_fiction",
    "length": "short"
  },
  "timeout": 300
}
```

**响应示例：**
```json
{
  "workflow": "novel_v4",
  "output": "在2145年的地球上...",
  "success": true,
  "metadata": {
    "word_count": 1500,
    "genre": "science_fiction"
  },
  "process_time_ms": 15000,
  "trace_id": "req_789"
}
```

### 5. 流式执行工作流

**路由：** `POST /api/stream`

**请求格式：** 与同步执行相同

**响应格式：** Server-Sent Events (SSE)

**响应示例：**
```
data: {"type": "start", "message": "开始执行工作流"}

data: {"type": "progress", "data": "第一章：未来的黎明\n在遥远的2145年..."}

data: {"type": "progress", "data": "人类已经掌握了星际航行技术..."}

data: {"type": "complete", "result": {"success": true, "process_time_ms": 15000}}
```

## 错误处理

### 错误常量

```go
var (
    ErrWorkflowNotFound = errors.New("工作流未找到") // 工作流未找到错误
    ErrInvalidRequest   = errors.New("无效的请求")   // 无效请求错误  
    ErrInternalError    = errors.New("内部服务错误") // 服务器内部错误
)
```

### HTTP 状态码

| 状态码 | 场景 | 说明 |
|--------|------|------|
| 200 | 成功 | 请求成功处理 |
| 400 | 请求错误 | 参数格式错误或缺失 |
| 404 | 未找到 | 工作流不存在 |
| 500 | 服务错误 | 内部服务器错误 |
| 503 | 服务不可用 | 调度器繁忙或服务停机 |

### 错误响应格式

```json
{
  "workflow": "requested_workflow",
  "output": "",
  "success": false,
  "message": "工作流未找到: unknown_workflow",
  "process_time_ms": 0,
  "trace_id": "req_123"
}
```

## 高级特性

### 1. 并发控制

**WorkerPool 调度器配置：**
```go
// 默认配置：8 workers, 队列深度 32
sched := scheduler.NewWorkerPoolScheduler(8, 32, processFunc)
```

**并发测试支持：**
- 内置并发测试用例 (`api_concurrency_test.go`)
- 支持高并发请求处理
- 队列溢出保护机制

### 2. 活跃任务追踪

```go
type WorkflowService struct {
    activeJobs sync.Map  // 追踪正在执行的任务
}
```

**功能：**
- 实时任务状态跟踪
- 任务取消支持
- 资源泄漏防护

### 3. 流式处理

**Server-Sent Events 支持：**
- 实时数据推送
- 长连接管理
- 断线重连支持

**流式事件类型：**
- `start` - 开始执行
- `progress` - 进度数据
- `error` - 错误信息
- `complete` - 执行完成

### 4. 请求追踪

**TraceId 支持：**
- 全链路请求追踪
- 日志关联
- 调试支持

### 5. 记忆系统集成

根据记忆中的信息，API 模块支持对话记忆功能：

**UserId 透传：**
- HTTP 请求体中的 `user_id` 字段
- WorkflowService 透传到 scheduler.Task.UserID
- Worker 回调中写入 context: `ctx = context.WithValue(ctx, "user_id", task.UserID)`

**记忆服务调用：**
- Agent 可在 before/after callbacks 中调用 memory.MemoryService
- SearchMemory 拼接历史对话
- AddSessionToMemory 写入本轮对话
- 实现持久化上下文支持

## 使用示例

### 基本服务启动

```go
package main

import (
    "log"
    
    "github.com/nvcnvn/adk-golang/pkg/api"
    "github.com/nvcnvn/adk-golang/pkg/flow"
)

func main() {
    // 创建工作流管理器
    manager := flow.NewManager()
    
    // 注册工作流
    // manager.Register("novel_v4", novel.NewBuilder())
    
    // 创建并启动 HTTP 服务器
    server := api.NewHttpServer(manager, ":8080")
    
    log.Println("启动 API 服务器在 :8080")
    if err := server.Start(); err != nil {
        log.Fatal("服务器启动失败:", err)
    }
}
```

### 客户端调用示例

```go
// 同步执行示例
func executeWorkflow(client *http.Client) {
    reqBody := api.WorkflowRequest{
        Workflow: "novel_v4",
        Input:    "写一个科幻小说",
        UserId:   "user123",
        Timeout:  300,
    }
    
    jsonData, _ := json.Marshal(reqBody)
    resp, err := client.Post("http://localhost:8080/api/execute", 
        "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    
    var result api.WorkflowResponse
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.Success {
        fmt.Println("执行成功:", result.Output)
    } else {
        fmt.Println("执行失败:", result.Message)
    }
}
```

### 流式执行示例

```javascript
// 前端 JavaScript 示例
const eventSource = new EventSource('/api/stream');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    if (data.type === 'progress') {
        console.log('进度:', data.data);
    } else if (data.type === 'complete') {
        console.log('完成:', data.result);
        eventSource.close();
    } else if (data.type === 'error') {
        console.error('错误:', data.message);
        eventSource.close();
    }
};
```

### 记忆系统集成示例

```go
// 在 Agent 的回调中使用记忆系统
agent := agents.NewAgent(
    agents.WithBeforeAgentCallback(func(ctx context.Context, message string) (string, bool) {
        userID := ctx.Value("user_id").(string)
        
        // 搜索历史记忆
        memoryService := memory.GetMemoryService()
        history, _ := memoryService.SearchMemory(ctx, userID, message)
        
        // 拼接历史上下文
        enhancedMessage := history + "\n\n" + message
        return enhancedMessage, true
    }),
    agents.WithAfterAgentCallback(func(ctx context.Context, response string) string {
        userID := ctx.Value("user_id").(string)
        
        // 保存本轮对话
        memoryService := memory.GetMemoryService()
        memoryService.AddSessionToMemory(ctx, userID, response)
        
        return response
    }),
)
```

## 配置与部署

### 环境变量

```bash
# 服务端口
HTTP_PORT=8080

# 调度器配置
SCHEDULER_WORKERS=8
SCHEDULER_QUEUE_SIZE=32

# 超时配置
DEFAULT_TIMEOUT=300

# 日志级别
LOG_LEVEL=INFO
```

### Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o adk-api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/adk-api .
EXPOSE 8080
CMD ["./adk-api"]
```

### 负载均衡配置

```nginx
upstream adk_api {
    server api1:8080;
    server api2:8080;
    server api3:8080;
}

server {
    listen 80;
    location /api/ {
        proxy_pass http://adk_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## 监控与调试

### 日志格式

```
[API] 2025-07-13T20:41:10Z INFO 工作流执行开始 workflow=novel_v4 user_id=user123 trace_id=req_789
[HTTP] 2025-07-13T20:41:10Z INFO POST /api/execute 200 15000ms
```

### 指标监控

推荐监控指标：
- 请求总数和成功率
- 响应时间分布
- 并发连接数
- 活跃任务数量
- 错误率按工作流分类

### 健康检查

```bash
# 基本健康检查
curl http://localhost:8080/health

# 检查特定工作流
curl http://localhost:8080/api/workflows/novel_v4
```

## 与其他模块的集成

- **flow**: 工作流管理器，提供工作流注册和查找
- **scheduler**: 任务调度器，提供并发执行能力
- **agents**: 智能体系统，执行具体的工作流逻辑
- **memory**: 记忆系统，提供上下文持久化
- **events**: 事件系统，用于状态通知
- **telemetry**: 遥测系统，用于性能监控

## 最佳实践

### 1. 性能优化
- 合理配置 Worker Pool 大小
- 使用连接池管理 HTTP 连接
- 实现请求缓存机制
- 避免长时间阻塞操作

### 2. 错误处理
- 实现优雅的错误降级
- 提供详细的错误信息
- 设置合理的超时时间
- 记录完整的错误上下文

### 3. 安全考虑
- 实现请求限流
- 验证输入参数
- 添加认证和授权
- 防止资源耗尽攻击

### 4. 可观测性
- 添加详细的日志记录
- 实现指标收集
- 提供调试接口
- 支持分布式追踪

## 开发状态

- ✅ HTTP 服务器基础框架
- ✅ 工作流服务核心逻辑
- ✅ RESTful API 路由实现
- ✅ Server-Sent Events 流式支持
- ✅ 并发调度和任务管理
- ✅ 错误处理和状态码
- ✅ 健康检查和监控
- ✅ 记忆系统集成支持
- ✅ 请求追踪和日志
- ✅ 并发测试覆盖

该模块为 ADK 系统提供了完整的 HTTP API 服务能力，是系统与外部世界交互的重要入口，支持各种客户端应用的集成需求。
