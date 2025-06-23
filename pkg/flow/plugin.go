package flow

// FlowPlugin 定义代码工作流插件接口。插件包需导出变量 `Plugin` 实现该接口。
// 插件编译示例：
//   go build -buildmode=plugin -o novel_flow_v1.so ./flows/novel
// 运行期由 PluginLoader 动态加载 .so，实现热更新。

import (
    "sync"

    "github.com/google/uuid"

    "github.com/nvcnvn/adk-golang/pkg/agents"
)

// FlowPlugin 插件需实现两个方法。
type FlowPlugin interface {
    Name() string                    // flow 名称 (唯一)
    Build() (*agents.Agent, error)   // 构造顶层 Agent
}

// Manager 维护所有已加载的工作流。
type Manager struct {
    mu    sync.RWMutex
    flows map[string]*agents.Agent
}

// NewManager 创建 Manager。
func NewManager() *Manager {
    return &Manager{flows: make(map[string]*agents.Agent)}
}

// Register 添加或替换工作流。
func (m *Manager) Register(name string, agent *agents.Agent) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.flows[name] = agent
}

// Unregister 删除工作流。
func (m *Manager) Unregister(name string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.flows, name)
}

// Get 查询工作流。
func (m *Manager) Get(name string) (*agents.Agent, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    a, ok := m.flows[name]
    return a, ok
}

// ListNames 返回已加载工作流名称列表。
func (m *Manager) ListNames() []string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    names := make([]string, 0, len(m.flows))
    for n := range m.flows {
        names = append(names, n)
    }
    return names
}

// TraceID 生成简单 trace id，供日志使用。
func TraceID() string {
    return uuid.NewString()
}
