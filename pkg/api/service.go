package api

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

// 定义 API 常量与错误类型
var (
	ErrWorkflowNotFound = errors.New("工作流未找到")
	ErrInvalidRequest   = errors.New("无效的请求")
	ErrInternalError    = errors.New("内部服务错误")
)

// WorkflowRequest 工作流执行请求
type WorkflowRequest struct {
	Workflow     string                 `json:"workflow"`                // 工作流名称
	Input        string                 `json:"input"`                   // 输入文本
	UserId       string                 `json:"user_id"`                 // 用户标识
	ExperimentId string                 `json:"experiment_id,omitempty"` // 实验ID（可选）
	TraceId      string                 `json:"trace_id,omitempty"`      // 追踪ID（可选）
	Parameters   map[string]interface{} `json:"parameters,omitempty"`    // 额外参数
	Timeout      int                    `json:"timeout,omitempty"`       // 超时（秒）
}

// WorkflowResponse 工作流执行结果
type WorkflowResponse struct {
	Workflow    string                 `json:"workflow"`           // 工作流名称
	Output      string                 `json:"output"`             // 输出文本
	Success     bool                   `json:"success"`            // 是否成功
	Message     string                 `json:"message,omitempty"`  // 消息（错误时有值）
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // 元数据
	ProcessTime int64                  `json:"process_time_ms"`    // 处理时间（毫秒）
	TraceId     string                 `json:"trace_id,omitempty"` // 请求追踪ID
}

// StreamCallback 流式回调函数
type StreamCallback func(data string, done bool, err error)

// WorkflowService 提供工作流执行服务
type WorkflowService struct {
	manager    *flow.Manager
	activeJobs sync.Map // 记录活跃的工作
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(manager *flow.Manager) *WorkflowService {
	return &WorkflowService{
		manager: manager,
	}
}

// Execute 执行工作流（同步）
func (s *WorkflowService) Execute(ctx context.Context, req WorkflowRequest) (*WorkflowResponse, error) {
	startTime := time.Now()
	
	// 确保有 trace_id
	if req.TraceId == "" {
		req.TraceId = flow.TraceID()
	}
	
	// 设置默认超时
	if req.Timeout <= 0 {
		req.Timeout = 30 // 默认30秒超时
	}
	
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	defer cancel()
	
	// 获取工作流
	agent, exists := s.manager.Get(req.Workflow)
	if !exists {
		log.Printf("[API] 工作流 %s 未找到", req.Workflow)
		return errorResponse(req.Workflow, "工作流未找到", req.TraceId), ErrWorkflowNotFound
	}
	
	// 执行结果通道
	type resultType struct {
		output string
		err    error
	}
	resultCh := make(chan resultType, 1)
	
	// 执行工作流（使用 goroutine 防止阻塞）
	log.Printf("[API] 开始执行工作流 %s，TraceID: %s，超时: %d秒", req.Workflow, req.TraceId, req.Timeout)
	go func() {
		log.Printf("[API] 工作流 %s 已启动异步处理，TraceID: %s", req.Workflow, req.TraceId)
		
		// 创建同步版上下文（不使用超时）
		// 尝试不同的上下文类型，因为可能是超时上下文影响了模型调用
		// syncCtx := context.Background()
		
		log.Printf("[API] 工作流 %s 调用 agent.Process，TraceID: %s", req.Workflow, req.TraceId)
		output, err := agent.Process(timeoutCtx, req.Input)
		
		if err != nil {
			log.Printf("[API] 工作流 %s 执行错误: %v，TraceID: %s", req.Workflow, err, req.TraceId)
		} else {
			log.Printf("[API] 工作流 %s 执行成功，返回结果长度: %d，TraceID: %s", req.Workflow, len(output), req.TraceId)
		}
		
		resultCh <- resultType{output, err}
	}()
	
	// 等待结果或超时
	var output string
	var err error
	select {
	case result := <-resultCh:
		output, err = result.output, result.err
		log.Printf("[API] 工作流 %s 执行完成，TraceID: %s", req.Workflow, req.TraceId)
	case <-timeoutCtx.Done():
		log.Printf("[API] 工作流 %s 执行超时，TraceID: %s", req.Workflow, req.TraceId)
		return errorResponse(req.Workflow, "工作流执行超时", req.TraceId), timeoutCtx.Err()
	}
	
	// 处理执行错误
	if err != nil {
		log.Printf("[API] 工作流 %s 执行失败: %v, TraceID: %s", req.Workflow, err, req.TraceId)
		return errorResponse(req.Workflow, err.Error(), req.TraceId), ErrInternalError
	}
	
	// 计算处理时间
	processTime := time.Since(startTime).Milliseconds()
	log.Printf("[API] 工作流 %s 执行成功，处理时间: %dms，TraceID: %s", req.Workflow, processTime, req.TraceId)
	
	// 返回结果
	return &WorkflowResponse{
		Workflow:    req.Workflow,
		Output:      output,
		Success:     true,
		ProcessTime: processTime,
		TraceId:     req.TraceId,
		Metadata: map[string]interface{}{
			"user_id":      req.UserId,
			"workflow":     req.Workflow,
			"experiment_id": req.ExperimentId,
		},
	}, nil
}

// ExecuteStream 执行工作流（异步流式）
func (s *WorkflowService) ExecuteStream(ctx context.Context, req WorkflowRequest, callback StreamCallback) error {
	// 确保有 trace_id
	if req.TraceId == "" {
		req.TraceId = flow.TraceID()
	}

	// 获取工作流
	agent, exists := s.manager.Get(req.Workflow)
	if !exists {
		callback("", false, ErrWorkflowNotFound)
		return ErrWorkflowNotFound
	}

	// 执行工作流（异步）
	log.Printf("[API] 开始流式执行工作流 %s，TraceID: %s", req.Workflow, req.TraceId)

	// 注册活跃工作
	jobID := req.TraceId
	s.activeJobs.Store(jobID, true)
	defer s.activeJobs.Delete(jobID)

	// TODO: 实现实际的流式处理
	// 这里是模拟实现，真正实现需要 Agent 提供流式接口
	go func() {
		result, err := agent.Process(ctx, req.Input)
		if err != nil {
			callback("", false, err)
			return
		}
		callback(result, true, nil)
	}()

	return nil
}

// 用于生成错误响应
func errorResponse(workflow, message, traceID string) *WorkflowResponse {
	return &WorkflowResponse{
		Workflow: workflow,
		Success:  false,
		Message:  message,
		TraceId:  traceID,
	}
}

// ListWorkflows 获取可用工作流列表
func (s *WorkflowService) ListWorkflows() []string {
	return s.manager.ListNames()
}

// GetWorkflowInfo 获取工作流详细信息
func (s *WorkflowService) GetWorkflowInfo(name string) (map[string]interface{}, error) {
	agent, exists := s.manager.Get(name)
	if !exists {
		return nil, ErrWorkflowNotFound
	}

	info := map[string]interface{}{
		"name":        name,
		"description": agent.Description(),
		"model":       agent.Model(),
		"type":        getAgentType(agent),
	}

	return info, nil
}

// 获取 Agent 类型
func getAgentType(agent *agents.Agent) string {
	// 根据 agent 特性判断类型
	// 这是一个简化实现
	return "basic" // 可能为 "sequential", "parallel", "basic" 等
}
