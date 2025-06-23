package flow

// 配置结构体定义，直接映射 <flow_name>.json。
// 使用 go-playground/validator 进行字段校验，但此处仅给出结构体与 Validate() 骨架，
// 具体实现后续补充。

import (
    "time"
)

// QueueConfig 定义消息队列配置（Redis / NATS 等）。
type QueueConfig struct {
    Impl   string `json:"impl" validate:"required"`  // redis / nats
    Stream string `json:"stream,omitempty"`          // Redis Stream 名称 / NATS subject
    MaxLen int64  `json:"max_len,omitempty"`         // 流最大长度，超过后自动裁剪
}

// PreGenerateConfig 定义预生成阶段。
type PreGenerateConfig struct {
    Enabled   bool   `json:"enabled"`
    Agent     string `json:"agent,omitempty"`
    TimeoutMs int    `json:"timeout_ms,omitempty"`
}

// AgentConfig 定义单个 Agent（既可以是叶子，也可以是容器）的运行参数，
// 通过 SubAgents 递归描述层级结构，以支持 cmd/adk/main.go 中的分层智能体配置。
type AgentConfig struct {
    ID           string                 `json:"id" validate:"required"`                                           // agent 唯一标识
    Type         string                 `json:"type" validate:"required,oneof=sequential parallel leaf"`         // leaf 表示无子节点
    Model        string                 `json:"model,omitempty"`                                                  // 叶子 agent 指定模型
    Instruction  string                 `json:"instruction,omitempty"`                                            // Prompt 指令
    Description  string                 `json:"description,omitempty"`
    Workers      int                    `json:"workers,omitempty"`                                                // parallel agent 专用
    StreamOutput bool                   `json:"stream_output,omitempty"`
    Params       map[string]interface{} `json:"params,omitempty"`
    SubAgents    []AgentConfig          `json:"sub_agents,omitempty" validate:"omitempty,dive"`                   // 子 Agent 列表
}

// RouteConfig 将 HTTP/gRPC 路径映射到 Agent。
type RouteConfig struct {
    Path  string `json:"path" validate:"required"`
    Agent string `json:"agent" validate:"required"`
}

// StorageConfig 定义持久化数据库连接信息（Gorm 方言）。
type StorageConfig struct {
    DSN string `json:"dsn" validate:"required"` // gorm DSN，可指向 MySQL / PostgreSQL
}

// FlowConfig 为完整工作流配置。
type FlowConfig struct {
    Name        string             `json:"name" validate:"required"`
    Queue       QueueConfig        `json:"queue" validate:"required,dive"`
    PreGenerate PreGenerateConfig  `json:"pre_generate,omitempty"`
    Agents      []AgentConfig      `json:"agents" validate:"required,min=1,dive"`
    Routes      []RouteConfig      `json:"routes" validate:"required,min=1,dive"`
    Storage     StorageConfig      `json:"storage" validate:"required,dive"`
    Version     string             `json:"version,omitempty"`

    UpdatedAt time.Time `json:"-"` // 热更新时间戳，程序内使用
}

// Validate 对 FlowConfig 进行字段校验；后续补充 validator 逻辑。
func (fc *FlowConfig) Validate() error {
    // TODO: 引入 validator 并实现校验逻辑
    return nil
}
