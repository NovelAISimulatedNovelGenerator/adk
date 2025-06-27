package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nvcnvn/adk-golang/pkg/config"
	"github.com/nvcnvn/adk-golang/pkg/flow"
)

func main() {
	// 解析配置文件路径
	configPath := os.Getenv("ADK_CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}
	
	// 加载配置文件
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 创建插件管理器
	manager := flow.NewManager()
	
	// 创建插件加载器，开始监控插件目录
	loader, err := flow.NewLoader(cfg.PluginDir, manager)
	if err != nil {
		log.Fatalf("创建插件加载器失败: %v", err)
	}
	loader.Start()
	// 注意：存在内存泄漏风险，如果有 Stop 方法应该调用
	// 在真实生产代码中，应该实现 Stop 方法来清理资源
	
	// 每秒输出当前已加载的插件列表
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	// 捕获退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("插件热更新监控已启动，按 Ctrl+C 退出")
	log.Printf("监控插件目录: %s", cfg.PluginDir)
	
	for {
		select {
		case <-ticker.C:
			plugins := manager.ListNames() // 正确的方法是 ListNames
			fmt.Printf("\033[2J\033[H") // 清屏
			fmt.Println("当前加载的插件列表:")
			fmt.Println("==================")
			for i, p := range plugins {
				fmt.Printf("[%d] %s\n", i+1, p)
			}
			if len(plugins) == 0 {
				fmt.Println("<无已加载插件>")
			}
			fmt.Println("==================")
			fmt.Println("等待插件更新...")
		case <-sigCh:
			log.Println("收到退出信号，停止监控")
			return
		}
	}
}
