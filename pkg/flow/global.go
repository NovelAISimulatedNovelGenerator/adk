package flow

// global.go 提供全局 Manager 访问点

import (
	"sync"
)

var (
	globalManager     *Manager
	globalManagerOnce sync.Once
)

// GetGlobalManager 返回全局单例 Manager 实例。
func GetGlobalManager() *Manager {
	globalManagerOnce.Do(func() {
		if globalManager == nil {
			globalManager = NewManager()
		}
	})
	return globalManager
}

// SetGlobalManager 设置全局 Manager 实例，通常在 main.go 中初始化时调用。
func SetGlobalManager(m *Manager) {
	globalManager = m
}
