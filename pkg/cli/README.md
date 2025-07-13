# CLI 命令行界面模块

## 概述

CLI 模块提供了 ADK (Agent Development Kit) 的完整命令行界面，支持智能体的开发、运行、测试、部署和管理。该模块是开发者与 ADK 框架交互的主要入口点，提供了丰富的命令和工具来简化智能体开发流程。

## 核心命令

### 主命令结构
```bash
adk - Agent Development Kit

可用命令:
  run        运行智能体的交互式模式
  serve      启动 Web 服务器，提供 API 和 UI 界面
  eval       评估智能体性能
  deploy     部署智能体到云平台
  version    显示版本信息
  help       显示帮助信息
```

## 命令详解

### 1. run - 运行智能体
```bash
adk run [agent_module]

参数:
  agent_module    智能体模块路径

选项:
  --save-session  保存会话到数据库
  --json-output   以JSON格式输出结果
  --help         显示帮助信息

示例:
  adk run ./agents/chat_agent
  adk run ./flows/novel_v4 --save-session
  adk run ./agents/code_assistant --json-output
```

智能体运行时将进入交互式模式，支持：
- 实时对话交互
- 会话历史管理
- 错误处理和重试
- 性能监控和日志记录

### 2. serve - 启动Web服务
```bash
adk serve [agents_directory]

参数:
  agents_directory    智能体目录路径 (默认: ./agents)

选项:
  --port PORT              服务端口 (默认: 8080)
  --ui                     启用Web UI界面
  --session-db URL         会话数据库URL (默认: sqlite://sessions.db)  
  --log-to-tmp            将日志写入临时目录
  --trace-to-cloud        启用云端追踪
  --log-level LEVEL       日志级别 (debug/info/warn/error)
  --allow-origins ORIGINS 允许的跨域来源，逗号分隔

示例:
  adk serve                                    # 启动基础API服务
  adk serve --ui --port 8080                   # 启动带UI的Web服务
  adk serve ./my-agents --session-db postgres://... # 使用自定义数据库
  adk serve --log-level debug --trace-to-cloud      # 启用调试和云端追踪
```

Web服务提供：
- RESTful API 接口
- 交互式 Web UI
- 会话管理和持久化
- 实时监控和日志
- 跨域资源共享支持

### 3. eval - 智能体评估
```bash
adk eval [agent_path] [evaluation_sets...]

参数:
  agent_path          智能体路径
  evaluation_sets     评估数据集路径

选项:
  --config PATH       评估配置文件路径
  --detailed         显示详细的评估结果
  --output FORMAT    输出格式 (json/table/csv)

示例:
  adk eval ./agents/qa_agent ./eval/qa_dataset.json
  adk eval ./flows/novel_v4 ./eval/creative_writing.json --detailed
  adk eval ./agents/code_gen ./eval/coding_tasks.json --config eval_config.yaml
```

评估功能包括：
- 多维度性能评估
- 批量测试执行
- 结果统计分析
- 自定义评估指标
- 评估报告生成

### 4. deploy - 云端部署
```bash
adk deploy [agent_folder]

参数:
  agent_folder        要部署的智能体文件夹

选项:
  --project PROJECT     GCP项目ID
  --region REGION       部署区域 (默认: us-central1)
  --service SERVICE     Cloud Run服务名称
  --app-name NAME       应用名称
  --temp-folder PATH    临时构建目录
  --port PORT          服务端口 (默认: 8080)
  --with-cloud-trace   启用Cloud Trace
  --with-ui           包含Web UI

示例:
  adk deploy ./agents/production_agent --project my-gcp-project
  adk deploy ./flows/novel_v4 --project my-project --service novel-api --with-ui
  adk deploy ./agents/chat_bot --region europe-west1 --with-cloud-trace
```

部署功能特性：
- 自动容器化构建
- Google Cloud Run 集成
- 环境变量配置
- 健康检查和监控
- 自动扩缩容

## 使用示例

### 开发工作流示例
```go
package main

import (
    "fmt"
    "github.com/nvcnvn/adk-golang/pkg/cli"
)

func main() {
    // CLI的主要入口点
    if err := cli.Execute(); err != nil {
        fmt.Printf("CLI执行失败: %v\n", err)
        os.Exit(1)
    }
}
```

### 智能体交互式运行
```bash
# 1. 启动智能体交互模式
$ adk run ./agents/my_chat_agent

ADK Agent CLI v1.0.0
正在加载智能体: ./agents/my_chat_agent
智能体已启动，输入 'quit' 退出

> 你好，请介绍一下你的能力
智能体: 您好！我是一个多功能的对话智能体，我可以：
1. 回答各种问题
2. 协助编程和技术讨论  
3. 进行创意写作
4. 数据分析和处理
请问我可以如何帮助您？

> 帮我写一段Python代码，实现斐波那契数列
智能体: 当然可以！这里是一个Python实现的斐波那契数列：

```python
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# 生成前10个斐波那契数
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
```

> quit
会话已结束，再见！
```

### Web服务启动
```bash
# 启动带UI的Web服务
$ adk serve ./agents --ui --port 8080 --log-level info

2025/01/15 10:00:00 INFO 正在启动ADK Web服务器...
2025/01/15 10:00:00 INFO 智能体目录: ./agents
2025/01/15 10:00:00 INFO 会话数据库: sqlite://sessions.db
2025/01/15 10:00:00 INFO 启用Web UI界面
2025/01/15 10:00:00 INFO 服务器启动在端口: 8080
2025/01/15 10:00:00 INFO Web UI可访问: http://localhost:8080
2025/01/15 10:00:00 INFO API端点可访问: http://localhost:8080/api
```

### 批量评估示例
```bash
# 运行智能体评估
$ adk eval ./agents/qa_agent ./eval/qa_dataset.json --detailed

ADK 智能体评估工具 v1.0.0
正在加载智能体: ./agents/qa_agent
正在加载评估数据集: ./eval/qa_dataset.json

评估进行中...
处理问题 1/100: 什么是人工智能？
处理问题 2/100: 机器学习的主要类型有哪些？
...
处理问题 100/100: 深度学习与传统机器学习的区别？

评估完成！

=== 评估结果摘要 ===
总问题数: 100
正确回答: 92
准确率: 92.0%
平均响应时间: 1.2秒
平均置信度: 0.89

=== 详细分析 ===
按类别分析:
- 基础概念: 95% (19/20)
- 技术细节: 90% (18/20) 
- 应用场景: 88% (17.6/20)
- 最新发展: 87% (17.4/20)
- 编程相关: 94% (18.8/20)

问题类型分析:
- 事实性问题: 96%
- 分析性问题: 89%
- 创造性问题: 85%

建议改进:
1. 加强最新技术发展的知识更新
2. 提高创造性问题的回答质量
3. 优化响应时间到1秒以内
```

### 云端部署示例
```bash
# 部署到Google Cloud Run
$ adk deploy ./agents/production_agent \
    --project my-gcp-project \
    --service chat-agent-api \
    --region us-central1 \
    --with-ui \
    --with-cloud-trace

ADK 云端部署工具 v1.0.0
正在部署智能体: ./agents/production_agent

步骤 1/6: 准备部署环境...
步骤 2/6: 构建Docker镜像...
正在构建镜像: gcr.io/my-gcp-project/chat-agent-api:latest

步骤 3/6: 推送镜像到Container Registry...
推送完成: gcr.io/my-gcp-project/chat-agent-api:latest

步骤 4/6: 部署到Cloud Run...
正在创建服务: chat-agent-api
区域: us-central1

步骤 5/6: 配置服务设置...
- 启用Cloud Trace追踪
- 启用Web UI界面
- 设置环境变量
- 配置健康检查

步骤 6/6: 验证部署...
部署成功！

=== 部署信息 ===
服务名称: chat-agent-api
项目: my-gcp-project
区域: us-central1
服务URL: https://chat-agent-api-xxx-uc.a.run.app
Web UI: https://chat-agent-api-xxx-uc.a.run.app/ui
API文档: https://chat-agent-api-xxx-uc.a.run.app/docs

请保存以上信息以便后续管理和访问。
```

## 高级功能

### 配置文件管理
CLI 支持多种配置文件格式：

```yaml
# adk-config.yaml
server:
  port: 8080
  host: "0.0.0.0"
  enable_ui: true
  
database:
  url: "postgres://user:pass@localhost/adk_sessions"
  max_connections: 10
  
logging:
  level: "info"
  output: "/var/log/adk.log"
  
tracing:
  enabled: true
  cloud_trace: true
  
agents:
  directory: "./agents"
  auto_reload: true
  
cors:
  allowed_origins:
    - "http://localhost:3000"
    - "https://my-app.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
```

### 环境变量配置
```bash
# 设置环境变量
export ADK_AGENTS_DIR="./agents"
export ADK_SESSION_DB="postgres://localhost/adk"
export ADK_LOG_LEVEL="debug"
export ADK_PORT="8080"
export ADK_ENABLE_UI="true"
export ADK_CLOUD_TRACE="true"

# 使用环境变量启动
adk serve
```

### 自定义命令扩展
```go
package main

import (
    "github.com/spf13/cobra"
    "github.com/nvcnvn/adk-golang/pkg/cli"
)

func init() {
    // 添加自定义命令
    customCmd := &cobra.Command{
        Use:   "benchmark",
        Short: "运行智能体性能基准测试",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runBenchmark(args)
        },
    }
    
    // 将自定义命令添加到根命令
    cli.RootCmd.AddCommand(customCmd)
}

func runBenchmark(args []string) error {
    // 自定义基准测试逻辑
    return nil
}
```

### 插件系统
```go
// 注册CLI插件
type CLIPlugin interface {
    Name() string
    Commands() []*cobra.Command
    Initialize() error
}

func RegisterPlugin(plugin CLIPlugin) {
    if err := plugin.Initialize(); err != nil {
        log.Printf("插件初始化失败: %v", err)
        return
    }
    
    for _, cmd := range plugin.Commands() {
        cli.RootCmd.AddCommand(cmd)
    }
}

// 使用示例
type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-custom-plugin"
}

func (p *MyPlugin) Commands() []*cobra.Command {
    return []*cobra.Command{
        {
            Use:   "mycmd",
            Short: "我的自定义命令",
            RunE:  p.runMyCommand,
        },
    }
}

func (p *MyPlugin) Initialize() error {
    // 插件初始化逻辑
    return nil
}

func (p *MyPlugin) runMyCommand(cmd *cobra.Command, args []string) error {
    // 命令执行逻辑
    return nil
}
```

## 监控和调试

### 日志管理
```bash
# 启用详细日志
adk serve --log-level debug --log-to-tmp

# 查看日志
tail -f /tmp/adk-*.log

# 结构化日志输出
2025/01/15 10:00:00 DEBUG [CLI] 正在解析命令参数
2025/01/15 10:00:01 INFO  [Server] 启动HTTP服务器，端口: 8080
2025/01/15 10:00:01 INFO  [Agent] 加载智能体模块: ./agents/chat_agent
2025/01/15 10:00:02 DEBUG [Database] 连接到会话数据库: sqlite://sessions.db
2025/01/15 10:00:02 INFO  [Server] 服务器就绪，接受请求
```

### 性能监控
```bash
# 启用云端追踪
adk serve --trace-to-cloud

# 查看性能指标
curl http://localhost:8080/metrics

# 健康检查
curl http://localhost:8080/health
```

### 错误处理
CLI 提供完善的错误处理机制：

```go
type CLIError struct {
    Code    int
    Message string
    Cause   error
}

func (e *CLIError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

// 错误类型
var (
    ErrAgentNotFound     = &CLIError{Code: 404, Message: "智能体未找到"}
    ErrInvalidConfig     = &CLIError{Code: 400, Message: "配置文件无效"}
    ErrServerStartFailed = &CLIError{Code: 500, Message: "服务器启动失败"}
    ErrDeploymentFailed  = &CLIError{Code: 500, Message: "部署失败"}
)
```

## 集成开发

### IDE集成
支持与主流IDE集成：

```json
// VSCode 任务配置 (.vscode/tasks.json)
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "ADK: 运行智能体",
            "type": "shell",
            "command": "adk",
            "args": ["run", "${workspaceFolder}/agents/my_agent"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "ADK: 启动开发服务器",
            "type": "shell", 
            "command": "adk",
            "args": ["serve", "--ui", "--log-level", "debug"],
            "group": "build",
            "isBackground": true
        }
    ]
}
```

### CI/CD 集成
```yaml
# GitHub Actions 工作流
name: ADK 智能体测试和部署
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: 构建ADK CLI
        run: go build -o adk ./cmd/adk
      
      - name: 运行智能体评估
        run: ./adk eval ./agents/test_agent ./eval/test_dataset.json
      
      - name: 运行性能基准测试
        run: ./adk benchmark ./agents/test_agent

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v2
      - uses: google-github-actions/setup-gcloud@v0
        with:
          service_account_key: ${{ secrets.GCP_SA_KEY }}
          project_id: ${{ secrets.GCP_PROJECT }}
      
      - name: 部署到Cloud Run
        run: |
          ./adk deploy ./agents/production_agent \
            --project ${{ secrets.GCP_PROJECT }} \
            --service production-agent \
            --with-cloud-trace
```

## 最佳实践

1. **项目结构**: 遵循标准的ADK项目结构
2. **配置管理**: 使用配置文件而非硬编码参数
3. **环境分离**: 为开发、测试、生产环境使用不同配置
4. **日志记录**: 启用适当级别的日志记录
5. **监控告警**: 在生产环境中启用监控和告警
6. **版本控制**: 对配置文件和智能体代码进行版本控制

## 依赖模块

- `github.com/spf13/cobra`: 命令行框架
- `github.com/fatih/color`: 彩色输出
- `github.com/nvcnvn/adk-golang/pkg/agents`: 智能体核心
- `github.com/nvcnvn/adk-golang/pkg/runners`: 运行器
- `github.com/nvcnvn/adk-golang/pkg/telemetry`: 遥测
- `github.com/nvcnvn/adk-golang/pkg/version`: 版本信息

## 故障排除

### 常见问题

1. **智能体启动失败**
   ```bash
   错误: 智能体模块未找到
   解决: 检查路径是否正确，确保智能体模块存在
   ```

2. **端口已被占用**
   ```bash
   错误: bind: address already in use
   解决: 使用 --port 参数指定其他端口
   ```

3. **数据库连接失败**
   ```bash
   错误: 无法连接到数据库
   解决: 检查数据库URL和连接配置
   ```

4. **Cloud Run部署失败**
   ```bash
   错误: 权限不足
   解决: 确保GCP服务账号具有必要权限
   ```

CLI 模块为 ADK-Golang 框架提供了完整的命令行工具集，是智能体开发、测试、部署和管理的核心入口点。
