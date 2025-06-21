pkg/events 目录总结

本目录提供 ADK-Golang 中的“事件(Event)”基础实现，负责在多智能体系统内统一描述、发布与处理事件。整体划分为三部分：

1. events.go —— 定义 Event 事件结构体及辅助方法，是事件模型的核心。
2. event_actions.go —— 为事件附加动作(EventActions)的结构体与逻辑，描述事件触发后可执行的额外行为。
3. event_bus.go —— 轻量级发布订阅(EventBus)实现，用于跨组件分发事件。

------------------------------------------------------------

events.go

主要成员/方法及作用:
- type Content = models.Content : 使用类型别名减少对 models 包的耦合。
- type Event struct : 描述一次事件的全部字段，包括 ID、Content、Actions 等。
- func NewEvent() : 生成带随机 UUID 的事件，并初始化 Actions。
- func (e *Event) IsFinalResponse() : 判断事件是否为一次最终响应（出现错误、完成内容输出或转交给下一 Agent）。
- func (e *Event) GetFunctionCalls() : 从 Content 中提取所有函数调用，便于后续工具层面处理。
- func GenerateID() : 生成 8 位可读随机字符串，作为轻量级标识。
- func init() : 在包加载时初始化随机种子。


------------------------------------------------------------

event_actions.go

主要成员/方法及作用:
- type EventActions struct : 表示事件附带的一组动作，例如 StateDelta、TransferToAgent、Escalate。
- func NewEventActions() : 创建并初始化空的 EventActions，避免 nil map。
- func (a *EventActions) Update(other *EventActions) : 将另一份动作合并进当前实例（OR 语义与 Map Merge）。

------------------------------------------------------------

event_bus.go

主要成员/方法及作用:
- type EventType string : 事件总线内部使用的事件类型枚举。
- 预定义常量 : ToolCalled、ToolError、ToolResultReceived —— 针对工具调用场景的三种事件。
- type EventHandler : 事件处理函数签名。
- func Subscribe(...) : 注册处理函数。线程安全，内部加锁写。
- func Unsubscribe(...) : 取消注册。通过比较函数指针地址实现，可能不够精确。
- func Publish(...) : 发布事件，依次调用所有订阅者。读锁保护。


------------------------------------------------------------

