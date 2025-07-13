// Package api 提供对 ADK 工作流管理与执行的 HTTP API。
//
// 本包围绕两层核心抽象：`HttpServer` 与 `WorkflowService`。
//
//   - HttpServer: 基于 net/http 封装所有 REST/SSE 路由，负责协议解析、状态码与错误处理，
//     以及健康检查等 Web 层细节。
//   - WorkflowService: 面向业务逻辑，负责与 flow.Manager 及 scheduler.Scheduler 协作，
//     实现工作流生命周期管理、同步与流式执行、活跃任务追踪等功能。
//
// # 快速开始
//
// 以下示例展示了如何在应用中启动一个 API 服务器：
//
//	// 创建并注册工作流
//	manager := flow.NewManager()
//	// manager.Register("novel_v4", novel.NewBuilder())
//
//	// 启动 HTTP 服务
//	server := api.NewHttpServer(manager, ":8080")
//	if err := server.Start(); err != nil {
//	    log.Fatal(err)
//	}
//
// # 路由说明
//
//  1. GET  /api/workflows           列出可用工作流
//  2. GET  /api/workflows/{name}    获取指定工作流信息
//  3. POST /api/execute             同步执行工作流，返回 JSON 结果
//  4. POST /api/stream              流式执行工作流，返回 Server-Sent Events
//  5. GET  /health                  服务健康检查
//
// 请求/响应体均采用 JSON 编码。字段含义请参考各结构体的 GoDoc 注释。
//
// 所有日志均以 "[API]" 或 "[HTTP]" 前缀输出，方便定位相关信息。
package api
