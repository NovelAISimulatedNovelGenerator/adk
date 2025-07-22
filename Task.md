# Task.md

> 本文件用于与上级 LLM Agent 沟通当前项目的整体任务、阶段进度与后续计划。完成一个总任务后，请务必更新并查看此文件。

## 当前阶段 / Current Stage
- 正在开发四元组memory系统的接口。
- 已完成GraphDB API调研，确定采用层次化逻辑分区方案。

## 任务列表 / Task List
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
    - [ ] **API层**: 在 WorkflowRequest 结构体中添加 `ArchiveId string \`json:"archive_id"\`` 字段（必填，不带omitempty）
    - [ ] **Scheduler层**: 在 `pkg/scheduler/scheduler.go` 中的 `scheduler.Task` 结构体添加 `ArchiveID string` 字段
    - [ ] **数据传递**: 在 `pkg/api/service.go` 中创建 Task 时传递 `ArchiveId: req.ArchiveId`
    - [ ] **Context注入**: 在 worker 回调中添加 `context.WithValue(ctx, "archive_id", task.ArchiveID)` 注入
    - [ ] **插件层访问**: 验证插件层可通过 `ctx.Value("archive_id")` 获取 archive_id
    - [ ] **文档更新**: 更新 `pkg/api/README.md` 文档，说明新字段用途与完整传递链路
    - [ ] **字段说明**: 确保 WorkflowRequest 的字段说明包含 archive_id 的描述
    - [ ] **测试覆盖**: 更新相关测试用例，覆盖 archive_id 字段和插件可访问性
    - [ ] **全链路验证**: 验证 API 调用时 archive_id 字段的序列化/反序列化及插件访问正常
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

请在后续开发过程中维护此文件，确保信息同步与可追溯性。