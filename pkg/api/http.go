package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/flow"
)

// HttpServer 提供 HTTP API 服务
type HttpServer struct {
	service *WorkflowService
	addr    string
	server  *http.Server
}

// NewHttpServer 创建 HTTP API 服务器
func NewHttpServer(manager *flow.Manager, addr string) *HttpServer {
	service := NewWorkflowService(manager)
	return &HttpServer{
		service: service,
		addr:    addr,
	}
}

// Start 启动 HTTP 服务
func (s *HttpServer) Start() error {
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/workflows", s.handleListWorkflows)
	mux.HandleFunc("/api/workflows/", s.handleWorkflowInfo)
	mux.HandleFunc("/api/execute", s.handleExecute)
	mux.HandleFunc("/api/stream", s.handleExecuteStream)
	mux.HandleFunc("/health", s.handleHealth)

	// 创建 HTTP 服务器
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	// 启动服务器
	log.Printf("[HTTP] API 服务启动于 %s", s.addr)
	return s.server.ListenAndServe()
}

// Stop 停止 HTTP 服务
func (s *HttpServer) Stop(ctx context.Context) error {
	log.Println("[HTTP] 关闭 API 服务")
	return s.server.Shutdown(ctx)
}

// handleHealth 健康检查
func (s *HttpServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	workflows := s.service.ListWorkflows()
	resp := map[string]interface{}{
		"status":     "ok",
		"version":    "1.0.0",
		"time":       time.Now().Format(time.RFC3339),
		"workflows":  len(workflows),
		"workflow_names": workflows,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleListWorkflows 列出工作流
func (s *HttpServer) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	workflows := s.service.ListWorkflows()
	resp := map[string]interface{}{
		"workflows": workflows,
		"count":     len(workflows),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleWorkflowInfo 获取工作流信息
func (s *HttpServer) handleWorkflowInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	// 从路径提取工作流名称
	name := r.URL.Path[len("/api/workflows/"):]
	if name == "" {
		http.Error(w, "缺少工作流名称", http.StatusBadRequest)
		return
	}

	info, err := s.service.GetWorkflowInfo(name)
	if err != nil {
		if err == ErrWorkflowNotFound {
			http.Error(w, "工作流未找到", http.StatusNotFound)
		} else {
			http.Error(w, "获取工作流信息失败", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleExecute 执行工作流
func (s *HttpServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var req WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}

	// 执行工作流
	ctx := r.Context()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	resp, err := s.service.Execute(ctx, req)
	if err != nil {
		switch err {
		case ErrWorkflowNotFound:
			http.Error(w, "工作流未找到", http.StatusNotFound)
		case ErrInvalidRequest:
			http.Error(w, "无效的请求", http.StatusBadRequest)
		default:
			http.Error(w, "执行工作流失败", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleExecuteStream 流式执行工作流
func (s *HttpServer) handleExecuteStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	// 设置流式响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "流式传输不支持", http.StatusInternalServerError)
		return
	}

	var req WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorEvent(w, "请求格式错误")
		flusher.Flush()
		return
	}

	// 执行流式工作流
	ctx := r.Context()
	err := s.service.ExecuteStream(ctx, req, func(data string, done bool, err error) {
		if err != nil {
			sendErrorEvent(w, err.Error())
			flusher.Flush()
			return
		}

		if done {
			sendEvent(w, "done", data)
		} else {
			sendEvent(w, "data", data)
		}
		flusher.Flush()
	})

	if err != nil {
		sendErrorEvent(w, err.Error())
		flusher.Flush()
		return
	}
}

// 发送 SSE 事件
func sendEvent(w io.Writer, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}

// 发送错误事件
func sendErrorEvent(w io.Writer, message string) {
	errorData, _ := json.Marshal(map[string]string{"error": message})
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
}
