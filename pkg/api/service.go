package api

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/flow"
	"github.com/nvcnvn/adk-golang/pkg/scheduler"
)

// 错误常量定义。
//
// ErrWorkflowNotFound 表示指定的工作流不存在。
// ErrInvalidRequest 表示请求参数不合法。
// ErrInternalError 表示服务器内部错误。
var (
	ErrWorkflowNotFound = errors.New("工作流未找到") // 工作流未找到错误
	ErrInvalidRequest   = errors.New("无效的请求")   // 无效请求错误
	ErrInternalError    = errors.New("内部服务错误") // 服务器内部错误
)

// WorkflowRequest 工作流执行请求
type WorkflowRequest struct {
	Workflow     string                 `json:"workflow"`                // 工作流名称
	Input        string                 `json:"input"`                   // 输入文本
	UserId       string                 `json:"user_id"`                 // 用户标识
	ArchiveId    string                 `json:"archive_id"`              // 归档标识符
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
	manager *flow.Manager
	sched   scheduler.Scheduler
	activeJobs sync.Map // 记录活跃的工作
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(manager *flow.Manager, sched scheduler.Scheduler) *WorkflowService {
	return &WorkflowService{
		manager: manager,
		sched:   sched,
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
	// 检查工作流是否存在
	if _, exists := s.manager.Get(req.Workflow); !exists {
		log.Printf("[API] 工作流 %s 未找到", req.Workflow)
		return errorResponse(req.Workflow, "工作流未找到", req.TraceId), ErrWorkflowNotFound
	}
	
    // 通过调度器提交任务
    resultCh := make(chan scheduler.Result, 1)
    task := &scheduler.Task{
        Ctx:        timeoutCtx,
        Workflow:   req.Workflow,
        Input:      req.Input,
        UserID:     req.UserId,
        ArchiveID:  req.ArchiveId,
        ResultChan: resultCh,
    }

    if err := s.sched.Submit(task); err != nil {
        if err == scheduler.ErrQueueFull {
            return errorResponse(req.Workflow, "系统繁忙，请稍后再试", req.TraceId), err
        }
        return errorResponse(req.Workflow, "提交任务失败", req.TraceId), err
    }

    // 等待结果或超时
    var output string
    var err error
    select {
    case res := <-resultCh:
        output, err = res.Output, res.Err
    case <-timeoutCtx.Done():
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
