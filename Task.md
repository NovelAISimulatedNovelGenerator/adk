# Task.md

> 本文件用于与上级 LLM Agent 沟通当前项目的整体任务、阶段进度与后续计划。完成一个总任务后，请务必更新并查看此文件。

## 当前阶段 / Current Stage
- 正在开发四元组memory系统的接口。
- 已完成GraphDB API调研，确定采用层次化逻辑分区方案。

## 任务列表 / Task List
- [ ] 根据项目需求补充阶段性任务
- [ ] **为API服务器WorkflowRequest添加archive_id字段**
  - **Description:** 为 pkg/api 模块的 WorkflowRequest 结构体添加必填的 archive_id 字段，用于标识归档/存档标识符。
  - **Acceptance Criteria:**
    - [ ] 在 WorkflowRequest 结构体中添加 `ArchiveId string \`json:"archive_id"\`` 字段（必填，不带omitempty）
    - [ ] 更新相关的 README.md 文档，说明新字段的用途
    - [ ] 确保 WorkflowRequest 的字段说明包含 archive_id 的描述
    - [ ] 更新相关的测试用例以包含 archive_id 字段
    - [ ] 验证 API 调用时 archive_id 字段的序列化/反序列化正常工作
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

## 进度记录 / Progress Log
| 日期 | 负责人 | 阶段 | 备注 |
| ---- | ------ | ---- | ---- |
| 2025-07-22 | Cascade | 初始化 | 创建 Task.md |
| 2025-07-22 | Gemini | 开发中 | 开始设计与实现四元组memory系统的接口，已完成文件和类型脚手架。 |

---

请在后续开发过程中维护此文件，确保信息同步与可追溯性。