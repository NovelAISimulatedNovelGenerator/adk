# Config 配置管理模块

## 概述

Config 模块提供了应用程序的配置管理功能，支持从 YAML 文件加载配置信息。该模块定义了完整的配置结构，包括插件目录、默认工作流、日志配置、数据库连接、消息队列和模型 API 池等核心配置项。

## 核心组件

### Config (主配置结构)
```go
type Config struct {
    PluginDir   string `yaml:"plugin_dir"`   // 插件目录路径
    DefaultFlow string `yaml:"default_flow"` // 默认工作流名称
    LogLevel    string `yaml:"log_level"`    // 日志级别
    LogDev      bool   `yaml:"log_dev"`      // 开发模式日志

    DB struct {
        DSN string `yaml:"dsn"` // 数据库连接字符串
    } `yaml:"db"`

    Queue struct {
        Impl   string `yaml:"impl"`   // 队列实现类型
        Addr   string `yaml:"addr"`   // 队列地址
        Stream string `yaml:"stream"` // 队列流名称
    } `yaml:"queue"`
    
    // 模型 API 池配置，按模型类型分组
    ModelAPIPools map[string]ModelPoolConfig `yaml:"model_api_pools"`
}
```

### EndpointConfig (API端点配置)
```go
type EndpointConfig struct {
    URL    string `yaml:"url"`    // API端点URL
    APIKey string `yaml:"apikey"` // API密钥
}
```

定义单个 API 端点的配置信息，包含访问地址和认证密钥。

### ModelPoolConfig (模型池配置)
```go
type ModelPoolConfig struct {
    Base      string           `yaml:"base"`      // 基础模型类型：deepseek/gemini等
    Endpoints []EndpointConfig `yaml:"endpoints"` // 端点列表  
}
```

定义模型 API 池的配置，支持为不同类型的模型配置多个端点，实现负载均衡和故障转移。

## 配置文件示例

### config.yaml
```yaml
# 基础配置
plugin_dir: "./plugins"
default_flow: "novel_generation"
log_level: "info"
log_dev: false

# 数据库配置
db:
  dsn: "user:password@tcp(localhost:3306)/adk?charset=utf8mb4&parseTime=True&loc=Local"

# 消息队列配置
queue:
  impl: "redis"
  addr: "localhost:6379"
  stream: "adk_tasks"

# 模型 API 池配置
model_api_pools:
  deepseek:
    base: "deepseek"
    endpoints:
      - url: "https://api.deepseek.com/v1"
        apikey: "your-deepseek-api-key"
      - url: "https://api.deepseek.com/v2"
        apikey: "your-backup-api-key"
  
  gemini:
    base: "gemini"
    endpoints:
      - url: "https://generativelanguage.googleapis.com/v1"
        apikey: "your-gemini-api-key"
        
  openai:
    base: "openai"
    endpoints:
      - url: "https://api.openai.com/v1"
        apikey: "your-openai-api-key"
```

## 使用方法

### 加载配置
```go
package main

import (
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/config"
)

func main() {
    // 使用默认路径 ./config.yaml
    cfg, err := config.Load("")
    if err != nil {
        panic(fmt.Sprintf("加载配置失败: %v", err))
    }

    // 或指定配置文件路径
    cfg2, err := config.Load("/path/to/config.yaml")
    if err != nil {
        panic(fmt.Sprintf("加载配置失败: %v", err))
    }

    // 使用配置
    fmt.Printf("插件目录: %s\n", cfg.PluginDir)
    fmt.Printf("默认工作流: %s\n", cfg.DefaultFlow)
    fmt.Printf("日志级别: %s\n", cfg.LogLevel)
    fmt.Printf("数据库DSN: %s\n", cfg.DB.DSN)
    
    // 访问模型API池配置
    if pool, exists := cfg.ModelAPIPools["deepseek"]; exists {
        fmt.Printf("DeepSeek 模型池有 %d 个端点\n", len(pool.Endpoints))
        for i, endpoint := range pool.Endpoints {
            fmt.Printf("  端点 %d: %s\n", i+1, endpoint.URL)
        }
    }
}
```

### 配置访问模式
```go
// 获取特定模型的 API 配置
func getModelConfig(cfg *config.Config, modelType string) (*config.ModelPoolConfig, bool) {
    pool, exists := cfg.ModelAPIPools[modelType]
    return &pool, exists
}

// 获取数据库连接配置
func getDatabaseConfig(cfg *config.Config) string {
    return cfg.DB.DSN
}

// 获取队列配置
func getQueueConfig(cfg *config.Config) (impl, addr, stream string) {
    return cfg.Queue.Impl, cfg.Queue.Addr, cfg.Queue.Stream
}
```

## 配置项说明

### 基础配置
- **plugin_dir**: 插件目录路径，用于加载动态插件
- **default_flow**: 默认使用的工作流名称
- **log_level**: 日志级别 (debug/info/warn/error)
- **log_dev**: 是否启用开发模式日志格式

### 数据库配置
- **db.dsn**: 数据库连接字符串，支持 MySQL/PostgreSQL 等

### 消息队列配置
- **queue.impl**: 队列实现类型 (redis/rabbitmq/memory)
- **queue.addr**: 队列服务器地址
- **queue.stream**: 消息流名称

### 模型 API 池配置
- **model_api_pools**: 按模型类型分组的 API 端点配置
- 支持多端点负载均衡和故障转移
- 每个端点包含 URL 和 API 密钥

## 最佳实践

1. **环境变量**: 敏感信息如 API 密钥建议通过环境变量注入
2. **配置验证**: 加载配置后应验证必要字段的完整性
3. **热重载**: 在生产环境可考虑实现配置热重载机制
4. **安全存储**: API 密钥等敏感信息应安全存储，避免明文配置
5. **默认值**: 为关键配置项设置合理的默认值

## 扩展配置

如需添加新的配置项，请：
1. 在相应的结构体中添加字段
2. 添加 `yaml` 标签
3. 更新配置文件示例
4. 添加相应的访问方法

## 依赖

- **gopkg.in/yaml.v3**: YAML 文件解析
- Go 标准库: `os`

## 注意事项

- 配置文件默认路径为 `./config.yaml`
- 所有配置字段保持首字母大写以支持 YAML 解码
- 模型 API 池支持动态配置多个端点
- 配置加载失败会返回详细错误信息
