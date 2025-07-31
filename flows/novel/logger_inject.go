// flows/novel/logger_inject.go
package main

import "go.uber.org/zap"

// plgLog 默认是 Nop，防止宿主还没注入时空指针
var plgLog *zap.Logger = zap.NewNop()

// SetLogger 会被宿主的 plugin_loader 调用，注入统一日志器
func SetLogger(l *zap.Logger) { plgLog = l }

// L 返回插件内部共用的 logger，业务代码用它打印日志
func L() *zap.Logger { return plgLog }
