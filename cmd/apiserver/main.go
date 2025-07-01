package main

import (
	"context"
	"flag"
	"log"
	"github.com/nvcnvn/adk-golang/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/nvcnvn/adk-golang/pkg/api"
	"github.com/nvcnvn/adk-golang/pkg/config"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

var (
	configFile = flag.String("config", "", "配置文件路径，也可通过 ADK_CONFIG 环境变量指定")
	addr = flag.String("addr", ":8080", "API 服务器监听地址")
)

func main() {
	flag.Parse()
	
	// 读取配置文件路径
	configPath := *configFile
	if configPath == "" {
		configPath = os.Getenv("ADK_CONFIG")
		if configPath == "" {
			configPath = "config.yaml"
		}
	}
	
	// 加载配置
	log.Printf("加载配置: %s", configPath)
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 创建管理器
	manager := flow.NewManager()
	
	// 创建加载器
	loader, err := flow.NewLoader(cfg.PluginDir, manager)
	if err != nil {
		log.Fatalf("创建插件加载器失败: %v", err)
	}
	
	// 启动加载器
	loader.Start()
	
	// 设置全局 Manager 实例
	flow.SetGlobalManager(manager)
	
	// 创建 HTTP 服务器
	server := api.NewHttpServer(manager, *addr)
	
	// 处理优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// 启动服务器（异步）
	go func() {
		if err := server.Start(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("API 服务器错误: %v", err)
			}
		}
	}()
	
	// 重新根据配置初始化结构化日志
	_, _ = logger.Init(cfg.LogLevel, cfg.LogDev)

	logger.S().Infow("API 服务启动完成", "workflows", manager.ListNames())
	
	// 等待退出信号
	sig := <-sigCh
	log.Printf("收到信号 %v，关闭中...", sig)
	
	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Stop(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	
	log.Println("服务器已关闭")
}
