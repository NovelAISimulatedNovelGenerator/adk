# Logger 使用提示

## 插件与宿主共享统一日志器

自 `v0.4.3` 起，插件可通过 `SetLogger(*zap.Logger)` 与宿主进程共享同一 `*zap.Logger` 实例，确保日志格式、级别、输出目标保持一致。

### 宿主侧
- `pkg/flow/plugin_loader.go` 在加载 `.so` 后会自动 `Lookup("SetLogger")`，若存在则注入宿主的全局日志器。

### 插件侧实现示例
```go
package main

import "go.uber.org/zap"

var plgLog = zap.NewNop() // 默认空 logger，防止注入前空指针

// SetLogger 由宿主进程注入
func SetLogger(l *zap.Logger) { plgLog = l }

// L 返回统一 logger，供插件内部使用
afunc L() *zap.Logger { return plgLog }
```

> ⚠️ 若插件未导出 `SetLogger`，将退化为 `zap.NewNop()`，日志不会输出。

## 标准库 `log` 重定向
宿主初始化 logger 时已调用 `zap.RedirectStdLog(lg)`，因此插件中的 `log.Println()` 同样会写入共享 logger，无需额外配置。
