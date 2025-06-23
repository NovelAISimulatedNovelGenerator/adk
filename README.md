# ADK-Golang 项目总览

> 本文档旨在帮助新成员在 **完全不了解历史背景** 的前提下，通过阅读 README 与记忆系统即可迅速掌握本仓库的核心设计与局限。若发现文档与代码不符，请 **优先修正文档** 并同步更新记忆系统。

## 顶层目标

ADK-Golang (Agent Development Kit) 试图提供一套 **可扩展、可组合、可部署** 的多智能体框架，使开发者能够在 Go 生态中快速构建、评估并上线基于 LLM 的智能体应用。

## `pkg` 目录分层说明

| 子包            | 作用概述 | 关键设计 & 足够暴露的接口 | 当前不足 / 改进建议 |
|-----------------|---------|--------------------------|--------------------|
| `agents`        | 定义 **Agent** 抽象与组合模型：`SequentialAgent`、`ParallelAgent`、`LoopAgent` 及 `RemoteAgent`。支持子树结构与事件流式返回。 | ‑ 统一的 `Process`/`Run` 接口便于 Runner 解耦<br>- 并行实现基于 `goroutine + WaitGroup`，简单直观 | ‑ 缺少超时 / 取消机制，长链路任务易泄露 goroutine<br>- 缺少对依赖关系图 (DAG) 的更灵活编排 |
| `artifacts`     | 提供运行过程产物存取服务，支持本地内存与 GCS。 | ‑ `ArtifactService` 接口抽象良好 | ‑ GCS 客户端无重试 / 限流策略<br>- 未提供本地文件系统实现，影响离线场景 |
| `auth`          | 通用鉴权封装，简化与外部 API (OpenAI 等) 的 token 管理。 | ‑ 将密钥注入逻辑集中，方便替换 | ‑ 功能表浅，缺少 OAuth2 / IAM 等高级方案 |
| `cli`           | 基于 `cobra` 的命令行工具，支持 **run / web / api_server / eval / deploy** 等子命令。 | ‑ 统一入口，覆盖开发->部署全流程 | ‑ 交互式体验受限：无命令补全、history 保存<br>- 大量逻辑堆叠在单文件，测试空白 |
| `code_executors`| 封装沙箱执行 Python / JS 等代码片段，供工具链调用。 | ‑ 统一 `Executor` 接口，方便扩展多语言 | ‑ 资源隔离不足，未使用 `seccomp`/cgroups<br>- 缺乏超时与内存限制配置 |
| `evaluation`    | 提供评测框架：读取 eval set、调度 Agent、产出指标。 | ‑ 易于自定义指标 | ‑ 缺乏可视化报告；指标体系偏简单 |
| `events`        | 定义交互事件模型，贯穿 agent-runner-UI。 | ‑ 基于通道的流式推送，降低内存压力 | ‑ 事件种类有限；序列化格式固定为 JSON，不够灵活 |
| `flows`         | 针对常见 LLM 工作流（如 RAG、Chain-of-Thought）提供模板封装。 | ‑ `llm_flows` 子目录拆分不同模式，易复用 | ‑ 与 `agents` 耦合度高；缺乏状态持久化支持 |
| `memory`        | 统一向量记忆接口，隐藏具体存储 (Supabase、Pinecone 等)。 | ‑ 抽象合理，可热插拔 | ‑ 未支持多段落插入 / 批量查询优化 |
| `models`        | 对上层隐藏 OpenAI / Ollama / Vertex AI 等模型差异。 | ‑ 简单工厂返回 `LLMClient` 实例 | ‑ 无自动重试 / 速率限制；缺乏 streaming API 支持完整性验证 |
| `planners`      | 实现任务分解与步骤规划 (如 GPT-4 planner)。 | ‑ 与 `agents` 解耦良好 | ‑ 算法仍偏黑盒，缺少可插拔 cost 函数 |
| `runners`       | 驱动 Agent 与终端 / WebSocket / HTTP 的桥接器。 | ‑ `SimpleRunner` 足够轻量 | ‑ 日志格式与 `telemetry` 重叠；并发 session 管理缺失 |
| `sessions`      | 会话记录与恢复。支持 JSON 与数据库驱动。 | ‑ API 简单易懂 | ‑ 加载策略单一；不支持分页检索 |
| `telemetry`     | 追踪与日志聚合，可对接 Cloud Trace。 | ‑ 抽象 `Tracer` 方便切换实现 | ‑ 默认实现功能有限；度量指标 (metrics) 缺席 |
| `tools`         | 内置工具库（WebSearch、FileOps、Browser 等），供 Agent 调用。 | ‑ 工具分层清晰，新增工具成本低 | ‑ 权限控制不足，潜在安全隐患；未提供调用配额管理 |
| `types`         | 基础类型定义，降低循环依赖。 | ‑ 保持纯粹，仅存放通用结构体 | ‑ 文档注释不足 |
| `version`       | 版本号常量，供 CLI 引用。 | — | — |

## 全局观察

1. **优点**
   * 模块划分贴合 Agent 生命周期，易于增量替换。
   * 充分利用 Go 并发特性实现并行 Agent。
   * CLI-Web-API 三位一体，覆盖主要使用场景。

2. **主要不足**
   * 测试覆盖率低，缺少端到端与并发场景测试。
   * 缺乏自动化 CI/CD 与静态分析，易引入回归。
   * 安全与资源隔离关注不足，高风险工具无沙箱。
   * 文档体系不完善：need ADR、序列图、性能基准。

## 下一步建议

| 优先级 | 任务 | 预期收益 |
|--------|-----|---------|
| 高 | 引入 **context.CancelFunc** 贯穿 `agents` 并行链路 | 避免 goroutine 泄露，支持客户端取消 |
| 高 | 构建 E2E 测试 (agent -> runner -> model mock) | 提升回归发现率 |
| 中 | 实现本地文件系统 `ArtifactService` | 离线与本机开发友好 |
| 中 | `code_executors` 集成 `docker` or `wasm` 沙箱 | 提升安全性 |
| 低 | 完善 README 与记忆系统双向更新自动化脚本 | 减少文档漂移 |

---

> _最后更新: 2025-06-22_  
> 如有疑问或建议，请在 Discussion 中反馈。

我希望配置一个代码工作流，使用adk-golang实现，专注于处理来自内网的api请求，因为本身是高并发服务，所以需要能够多线程处理，queue等系统。支持json文件配置多种工作流 例如在cmd/adk/main中的整个工作流作为一个json文件，然后通过json文件名/其他方式选择使用。
可以是通过Hertz-adk_api双端通信 -> 初始化(通过json文件配置)服务 -> 使用服务 -> 断开通信 -> 销毁服务
你需要仔细思考系统架构，保持架构的鲁棒性，你需要持有批判性思维，你需要提问以完善架构