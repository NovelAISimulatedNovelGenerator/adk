# Task.md

> 本文件用于与上级 LLM Agent 沟通当前项目的整体任务、阶段进度与后续计划。完成一个总任务后，请务必更新并查看此文件。

## 当前阶段 / Current Stage


## 任务列表 / Task List
- [ ] **向量RAG写入读取Tool模块完整开发**
  - **Description:** 基于现有架构设计完整的向量RAG写入读取Tool模块，复用pkg/memory/custom_rag.go和pkg/flows/save_novel_rag_data_workflow的成熟功能，为智能体系统提供标准化的向量数据库操作工具。
  - **Technical Context:** 基于pkg/tools/tool.go接口规范，整合现有的CustomRagMemoryService和内容处理逻辑，实现可插拔的RAG工具模块。
  - **Acceptance Criteria:**
    - [ ] **目录结构**: 创建pkg/tools/rag_tool目录，包含完整的RAG工具实现
    - [ ] **核心工具实现**: 实现RAGWriteTool和RAGReadTool，遵循Tool接口规范
    - [ ] **配置管理**: 实现RagToolConfig结构体，支持RAG服务配置
    - [ ] **内容处理**: 集成SaveNovelRagDataService的内容分段和处理逻辑
    - [ ] **错误处理**: 完整的错误处理、重试机制和中文日志记录
    - [ ] **健康检查**: 实现RAG服务健康检查功能
    - [ ] **Schema定义**: 定义完整的JSON Schema用于工具参数验证
    - [ ] **单元测试**: 编写完整的单元测试用例，覆盖核心功能
    - [ ] **性能基准**: 实现基准测试，验证工具性能
    - [ ] **使用示例**: 提供完整的使用示例和文档
    - [ ] **集成测试**: 验证与现有智能体系统的集成能力
    - [ ] **文档完善**: 编写详细的README.md文档
  - **详细实现规划:**
    - **Task 1: 目录结构和基础文件创建**
      - 创建pkg/tools/rag_tool目录
      - 创建rag_tool.go（核心实现）
      - 创建config.go（配置管理）
      - 创建schema.go（Schema定义）
      - 创建rag_tool_test.go（测试文件）
      - 创建examples_test.go（示例文件）
      - 创建README.md（文档）
    - **Task 2: 核心数据结构定义**
      - 定义RagToolConfig结构体
      - 定义RAGWriteInput/RAGWriteOutput结构体
      - 定义RAGReadInput/RAGReadOutput结构体
      - 定义ContentSegment结构体
      - 定义HealthCheckResult结构体
    - **Task 3: RAGWriteTool实现**
      - 实现RAGWriteTool结构体
      - 实现Name()、Description()方法
      - 实现Schema()方法，定义输入输出JSON Schema
      - 实现Execute()方法，集成内容处理和RAG写入逻辑
      - 实现内容分段处理
      - 实现批量写入优化
    - **Task 4: RAGReadTool实现**
      - 实现RAGReadTool结构体
      - 实现Name()、Description()方法
      - 实现Schema()方法，定义查询参数和结果Schema
      - 实现Execute()方法，集成RAG搜索和结果处理
      - 实现查询结果格式化
      - 实现相似度筛选和排序
    - **Task 5: 配置管理和工厂方法**
      - 实现NewRagToolConfig()构造函数
      - 实现NewRAGWriteTool()工厂方法
      - 实现NewRAGReadTool()工厂方法
      - 实现配置验证逻辑
      - 实现默认配置支持
    - **Task 6: 健康检查和监控**
      - 实现RAG服务健康检查
      - 实现连接状态监控
      - 实现性能指标收集
      - 实现错误统计和报告
    - **Task 7: 错误处理和重试机制**
      - 实现完整的错误处理链
      - 实现指数退避重试机制
      - 实现上下文取消支持
      - 实现中文错误日志记录
    - **Task 8: JSON Schema定义**
      - 定义RAGWriteTool的输入输出Schema
      - 定义RAGReadTool的输入输出Schema
      - 实现Schema验证逻辑
      - 实现参数类型检查
    - **Task 9: 单元测试和集成测试**
      - 编写RAGWriteTool单元测试
      - 编写RAGReadTool单元测试
      - 编写配置管理测试
      - 编写Schema验证测试
      - 编写错误处理测试
      - 编写性能基准测试
      - 编写端到端集成测试
    - **Task 10: 文档和示例**
      - 编写详细的README.md
      - 编写使用示例和最佳实践
      - 编写API文档
      - 编写配置说明
      - 编写故障排除指南
- [ ] 根据项目需求补充阶段性任务
- [ ] **Novel工作流与Quad Memory服务集成测试及交付**
  - **Description:** 参考 `pkg/flows/novel/framework.go` 和 `pkg/flows/test`，创建完整的 `flows/test` 集成测试框架，实际调用 `pkg/memory/quad_memory_service.go` 进行记忆存储操作，完善 novel plugin 的构建流程，最终完成 Docker 化交付。
  - **Technical Context:** 需要验证 novel 工作流与 quad_memory_service 的整合，确保在实际场景中能够正常协作并进行记忆存储操作。
  - **Acceptance Criteria:**
    - [ ] **集成测试框架**: 创建 `flows/test` 目录，包含完整的集成测试用例
    - [ ] **Memory服务集成**: 在测试中实际调用 `pkg/memory/quad_memory_service.go` 的 AddQuad 和 SearchQuads 方法
    - [ ] **Novel工作流测试**: 验证 `pkg/flows/novel/framework.go` 中的各个Agent能正常协作
    - [ ] **Context传递验证**: 验证 user_id 和 archive_id 在 novel 工作流中的可访问性
    - [x] **Plugin构建**: 参考 `flows/novel/main.go`，为 `flows/test` 创建相似的 plugin 入口 ✅ **已完成**
    - [x] **构建流程**: 确保 `go build` 能正常编译 test plugin ✅ **已完成**
    - [x] **Docker化**: 完成 `docker build` 和相关配置，实现一键启动 ✅ **已完成**
      - **解决问题**: 修复了Docker环境下插件加载失败问题，移除了docker-compose.yml中覆盖插件的volume挂载
    - [ ] **文档完善**: 更新相关 README 和使用说明文档
- [x] **为API服务器WorkflowRequest添加archive_id字段及全链路透传支持** ✅ **已完成**
  - **Description:** 为 pkg/api 模块的 WorkflowRequest 结构体添加必填的 archive_id 字段，并确保 archive_id 能通过 scheduler.Task、worker回调、context 机制完整传递到 plugin 层（如 framework.go），实现归档标识全链路可用。
  - **Technical Context:** 基于Plugin系统Context传递机制分析，需要实现与 user_id 相似的完整数据传递链路。
  - **Acceptance Criteria:**
    - [x] **API层**: 在 WorkflowRequest 结构体中添加 `ArchiveId string \`json:"archive_id"\`` 字段（必填，不带omitempty）
    - [x] **Scheduler层**: 在 `pkg/scheduler/scheduler.go` 中的 `scheduler.Task` 结构体添加 `ArchiveID string` 字段
    - [x] **数据传递**: 在 `pkg/api/service.go` 中创建 Task 时传递 `ArchiveId: req.ArchiveId`
    - [x] **Context注入**: 在 worker 回调中添加 `context.WithValue(ctx, "archive_id", task.ArchiveID)` 注入
    - [x] **插件层访问**: 验证插件层可通过 `ctx.Value("archive_id")` 获取 archive_id
    - [ ] **文档更新**: 更新 `pkg/api/README.md` 文档，说明新字段用途与完整传递链路
    - [x] **字段说明**: 确保 WorkflowRequest 的字段说明包含 archive_id 的描述
    - [x] **测试覆盖**: 更新相关测试用例，覆盖 archive_id 字段和插件可访问性
    - [x] **全链路验证**: 验证 API 调用时 archive_id 字段的序列化/反序列化及插件访问正常
- [x] **设计与实现四元组memory系统的接口** ✅ **已完成**
  - [x] **Task 1: File and Type Scaffolding** ✅
    - **Description:** Create the necessary files and define all the core data structures required for the service.
    - **Acceptance Criteria:**
      - ✅ Create a new file: `pkg/memory/quad_memory_service.go`.
      - ✅ In the new file, define the following structs: `QuadMemoryConfig`, `QuadMemoryService`, `Quad`, `AddQuadRequest`, `QuadSearchQuery`, `SearchQuadsRequest`.
      - ✅ Create a new test file: `pkg/memory/quad_memory_service_test.go`.
  - [x] **Task 2: Implement Constructor and Health Check** ✅
    - **Description:** Implement the client's constructor and the `HealthCheck` method.
    - **Acceptance Criteria:**
      - ✅ Implement the `NewQuadMemoryService(config QuadMemoryConfig)` function.
      - ✅ Implement the `HealthCheck(ctx context.Context)` method.
      - ✅ Write unit tests in `quad_memory_service_test.go` for the `HealthCheck` method, covering success, failure (4xx/5xx), and context cancellation scenarios.
  - [x] **Task 3: Implement `AddQuad` Method** ✅
    - **Description:** Implement the logic for adding a new quad to the memory service.
    - **Acceptance Criteria:**
      - ✅ Implement the `AddQuad(ctx context.Context, hierarchicalCtx *HierarchicalContext, quad Quad) (*Quad, error)` method with hierarchical context support.
      - ✅ The method correctly uses SPARQL INSERT DATA to add quads to named graphs based on hierarchical context.
      - ✅ The method handles GraphDB REST API responses and implements proper error handling.
      - ✅ The method implements retry logic for transient errors as specified in the design.
      - ✅ Write unit tests for `AddQuad`, including success, failure, and retry scenarios.
  - [x] **Task 4: Implement `SearchQuads` Method** ✅
    - **Description:** Implement the logic for searching for quads in the memory service.
    - **Acceptance Criteria:**
      - ✅ Implement the `SearchQuads(ctx context.Context, query QuadSearchQuery) ([]*Quad, error)` method with hierarchical context and scope support.
      - ✅ The method correctly uses SPARQL SELECT queries with appropriate named graph filtering based on scope (exact, story, tenant).
      - ✅ The method handles SPARQL JSON response format and deserializes results into Quad objects.
      - ✅ The method implements the same retry logic as `AddQuad`.
      - ✅ Write unit tests for `SearchQuads`, covering success, failure (no results found), and retry scenarios.
  - [x] **Task 5: Finalize and Document** ✅
    - **Description:** Add final touches, including code comments and updating the main project task file.
    - **Acceptance Criteria:**
      - ✅ Add package and function-level comments to `quad_memory_service.go` (全面中文化完成).
      - ✅ Review and ensure logging is adequate for debugging and monitoring.
      - ✅ Run all tests to confirm the feature is complete and correct (所有测试通过).
      - ✅ Update the root `Task.md` to mark the `quad-memory-system` task as complete.
- [ ] 每完成一个阶段后更新进度

## 技术设计方案 / Technical Design

### GraphDB 层次化逻辑分区方案

#### 背景与约束
- **GraphDB Free版本限制**: 最多支持5个repositories
- **业务需求**: 支持多租户、多故事、多章节等层次化数据组织
- **解决方案**: 采用单repository + 命名图(Named Graph)的逻辑分区方案

#### 层次化命名图URI设计
```
urn:tenant:{tenantID}                           // 租户级别
urn:tenant:{tenantID}:story:{storyID}           // 故事级别  
urn:tenant:{tenantID}:story:{storyID}:chapter:{chapterID}  // 章节级别
urn:tenant:{tenantID}:story:{storyID}:character:{characterID} // 角色级别
```

#### API扩展设计
```go
// 层次化上下文结构
type HierarchicalContext struct {
    TenantID    string `json:"tenant_id"`
    StoryID     string `json:"story_id,omitempty"`
    ChapterID   string `json:"chapter_id,omitempty"`
    CharacterID string `json:"character_id,omitempty"`
}

// 扩展的搜索查询
type QuadSearchQuery struct {
    Subject     string                `json:"subject,omitempty"`
    Predicate   string                `json:"predicate,omitempty"`
    Object      string                `json:"object,omitempty"`
    Context     *HierarchicalContext  `json:"context,omitempty"`
    Scope       string                `json:"scope,omitempty"` // "exact", "story", "tenant"
}
```

#### 核心实现方法
1. **AddQuadWithContext**: 支持层次化上下文的数据插入
2. **SearchQuadsWithScope**: 支持不同范围的数据查询
   - `exact`: 精确匹配指定层次
   - `story`: 查询整个故事的所有相关数据
   - `tenant`: 查询整个租户的所有数据

#### SPARQL实现示例
```sparql
-- 插入数据到特定层次
INSERT DATA { 
    GRAPH <urn:tenant:user123:story:story1:chapter:ch1> { 
        <chapter1> <hasContent> "It was a dark and stormy night..." 
    } 
}

-- 查询整个故事的数据
SELECT ?s ?p ?o ?g WHERE { 
    GRAPH ?g { 
        ?s ?p ?o 
    }
    FILTER(STRSTARTS(STR(?g), "urn:tenant:user123:story:story1"))
}
```

#### 技术优势
- **无限扩展**: 支持任意层次的数据组织
- **灵活查询**: 支持精确查询和范围查询
- **性能优化**: GraphDB基于URI前缀高效过滤
- **数据隔离**: 每个层次完全隔离，确保数据安全
- **成本友好**: 完全在GraphDB Free版本限制内

### Plugin系统Context传递机制

#### 背景与重要性
- **业务需求**: API层的user_id、archive_id等上下文信息需要传递到插件工作流中
- **技术挑战**: 确保编译型插件系统能够访问HTTP请求中的元数据
- **关键影响**: 影响个性化处理、记忆服务、归档管理等核心功能

#### 完整数据传递链路分析

**user_id传递机制** ✅ **已完全支持**
```
HTTP请求 → WorkflowRequest.UserId → scheduler.Task.UserID → context.WithValue → 插件ctx.Value("user_id")
```

**关键实现点**:
1. **API层**: `pkg/api/service.go:97` - 将`req.UserId`传递给`scheduler.Task.UserID`
2. **Scheduler层**: `pkg/scheduler/scheduler.go:105` - Worker调用`s.processor(task.Ctx, task)`
3. **Context注入**: Worker回调中通过`context.WithValue(ctx, "user_id", task.UserID)`注入
4. **插件访问**: 在`framework.go`回调函数中可直接使用`ctx.Value("user_id")`

**插件使用示例**:
```go
// 在 pkg/flows/novel/framework.go 中
executionLayer.Agent.SetBeforeAgentCallback(func(ctx context.Context, msg string) (string, bool) {
    // ✅ 可以直接获取user_id
    if userID, ok := ctx.Value("user_id").(string); ok {
        log.Printf("[执行层] 处理用户 %s 的请求", userID)
        // 可用于个性化处理、记忆服务调用等
    }
    return "处理完成", true
})
```

**archive_id传递机制** ❌ **尚未支持，需要补全**

**当前缺失环节**:
1. ❗ `scheduler.Task`结构体缺少`ArchiveID`字段
2. ❗ Worker回调缺少`context.WithValue(ctx, "archive_id", task.ArchiveID)`注入
3. ❗ 插件层无法通过`ctx.Value("archive_id")`获取

**完整解决方案** (需要实施):
1. **扩展scheduler.Task结构体**:
   ```go
   // pkg/scheduler/scheduler.go
   type Task struct {
       Ctx       context.Context
       Workflow  string
       Input     string
       UserID    string
       ArchiveID string  // ❗ 新增字段
       ResultChan chan Result
   }
   ```

2. **修改API层数据传递**:
   ```go
   // pkg/api/service.go
   task := &scheduler.Task{
       Ctx:        timeoutCtx,
       Workflow:   req.Workflow,
       Input:      req.Input,
       UserID:     req.UserId,
       ArchiveID:  req.ArchiveId,  // ❗ 新增传递
       ResultChan: resultCh,
   }
   ```

3. **在Worker回调中注入context**:
   ```go
   // 在processor实现中
   ctx = context.WithValue(ctx, "archive_id", task.ArchiveID)
   ```

4. **插件层访问**:
   ```go
   // pkg/flows/novel/framework.go
   if archiveID, ok := ctx.Value("archive_id").(string); ok {
       log.Printf("[执行层] 归档ID: %s", archiveID)
       // 可用于数据归档、版本管理等
   }
   ```

#### 技术影响总结

| 字段 | 当前状态 | 插件可访问性 | 需要的工作 |
|------|---------|-------------|----------|
| `user_id` | ✅ 完全支持 | ✅ 可直接使用 `ctx.Value("user_id")` | 无 |
| `archive_id` | ❌ 未支持 | ❌ 无法访问 | 需要补全整条传递链路 |

**结论**: Plugin系统**可以**读取user_id，但**无法**读取archive_id。要实现archive_id的完整支持，需要同时修改API层、Scheduler层和Context注入机制。

## 进度记录 / Progress Log
| 日期 | 负责人 | 阶段 | 备注 |
| ---- | ------ | ---- | ---- |
| 2025-07-22 | Cascade | 初始化 | 创建 Task.md |
| 2025-07-22 | Gemini | 开发中 | 开始设计与实现四元组memory系统的接口，已完成文件和类型脚手架。 |

---

## 新增任务：统一项目配置管理方案

### 任务背景
当前项目中配置参数分散在环境变量和config.yaml文件中，需要统一配置管理方案，确保所有业务配置集中管理，同时保持必要的系统级环境变量支持。

### 当前配置分布统计
- **环境变量使用：** `ADK_CONFIG`, `DEEPSEEK_API_KEY`, `GOOGLE_API_KEY`, `RAG_SERVICE_URL` 等
- **配置文件使用：** `plugin_dir`, `default_flow`, `db.dsn`, `queue.*`, `model_api_pools` 等
- **冲突配置：** 模型API密钥和端点同时存在环境变量和配置文件结构

### 统一方案：配置文件为主，环境变量覆盖
**技术栈：**
- 配置管理：viper (支持YAML、JSON、TOML、env)
- 配置验证：go-playground/validator
- 开发环境：godotenv (.env文件支持)

### 实施计划
- [ ] **配置结构重构**：扩展config.yaml结构，包含所有当前环境变量配置
- [ ] **代码迁移**：将pkg/models/deepseek.go、pkg/models/gemini.go中的环境变量读取改为配置读取
- [ ] **viper集成**：在pkg/config中集成viper，支持环境变量覆盖
- [ ] **.env支持**：添加.env文件支持，用于开发环境默认值
- [ ] **配置验证**：实现配置结构和必填项验证
- [ ] **向后兼容**：确保现有环境变量在过渡期内仍能工作
- [ ] **文档更新**：更新README.md，说明新的配置使用方法
- [ ] **Docker更新**：更新docker-compose.yml，移除不必要的volume挂载
- [ ] **示例配置**：更新config.example.yaml，包含所有配置项示例

### 技术实现要点
```yaml
# 扩展后的config.yaml结构示例
plugin_dir: ./plugins
default_flow: novel_flow_v1
log_level: info

# 统一模型配置
models:
  deepseek:
    api_key: "your-deepseek-key"
    endpoint: "https://api.deepseek.com/v1"
  gemini:
    api_key: "your-gemini-key"
    endpoint: "https://generativelanguage.googleapis.com/v1beta"

# 外部服务配置
services:
  rag_service:
    url: "http://localhost:8080"

# 原有配置保持不变
db:
  dsn: "postgres://user:pass@localhost:5432/adk?sslmode=disable"

queue:
  impl: redis
  addr: "redis://localhost:6379"
  stream: adk_tasks
```

---

请在后续开发过程中维护此文件，确保信息同步与可追溯性。

## 当前系统启动流程

```
配置加载 → 创建管理器和加载器 → 
Quad Memory服务验证 → Custom RAG Memory服务验证 → 
启动插件加载器 → 启动API服务器 → 优雅关闭处理
```

### 详细启动步骤

1. **配置解析**: 读取config.yaml或环境变量配置
2. **Flow管理器创建**: 初始化workflow管理器
3. **插件加载器创建**: 准备插件系统
4. **内存服务验证**:
   - **Quad Memory**: 验证GraphDB连接 (localhost:7200)
   - **Custom RAG Memory**: 验证FastAPI+Milvus连接 (localhost:18000)
   - 验证失败时发出警告但不阻断启动
5. **插件加载**: 从./plugins目录加载.so插件文件
6. **HTTP服务器启动**: 在指定端口启动API服务
7. **优雅关闭**: 监听SIGINT/SIGTERM信号，支持优雅关闭

### 健康检查策略

- **非阻塞验证**: 任何外部服务不可用都不会阻止系统启动
- **超时控制**: 健康检查使用10秒超时
- **详细日志**: 提供服务可用性的明确反馈
- **容错设计**: 系统在部分服务不可用时仍可正常运行

### 依赖服务

| 服务 | 端口 | 用途 | 必需性 |
|------|------|------|--------|
| GraphDB | 7200 | 四元组知识图谱存储 | 可选 |
| RAG API | 18000 | 向量搜索和RAG功能 | 可选 |
| Redis | 6379 | 任务队列 | 必需 |
| PostgreSQL | 5432 | 数据持久化 | 必需 |