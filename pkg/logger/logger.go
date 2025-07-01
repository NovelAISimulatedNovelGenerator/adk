package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	global *zap.Logger
	once   sync.Once
)

// Init 初始化全局 logger。
// level 字符串形如 "debug"|"info"|"warn"|"error"。
// dev 为 true 时使用开发者友好配置（彩色日志、caller）。
func Init(level string, dev bool) (*zap.Logger, error) {
	var lv zapcore.Level
	if err := lv.UnmarshalText([]byte(level)); err != nil {
		lv = zap.InfoLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(lv),
		Development:      dev,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
			LineEnding:    zapcore.DefaultLineEnding,
		},
	}

	lg, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	// 保证全局只初始化一次
	once.Do(func() {
		global = lg
		// 将标准库 log 重定向到 zap
		_ = zap.RedirectStdLog(lg)
	})

	return lg, nil
}

// L 返回全局 *zap.Logger。如未初始化则使用默认 info 级别。
func L() *zap.Logger {
	if global == nil {
		_, _ = Init("info", false)
	}
	return global
}

// S 返回全局 SugaredLogger
func S() *zap.SugaredLogger {
	return L().Sugar()
}

// Sync 刷新日志缓冲区，通常在程序退出时调用
func Sync() {
	_ = L().Sync()
}
