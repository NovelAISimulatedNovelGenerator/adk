# SaveNovelRagDataWorkflow - 小说内容RAG存储工具

## 概述

`save_novel_rag_data_workflow` 是一个辅助工具包，旨在将小说内容智能分段并存储到RAG（检索增强生成）向量数据库中。该工具基于LLM进行内容分析和分段，然后将处理后的内容存储到向量数据库，便于后续的语义检索和内容生成。

## 核心功能

- **智能内容分段**：使用LLM将长文本分解为语义完整的段落
- **内容类型识别**：自动识别并标注内容类型（角色、地点、剧情、对话等）
- **特定内容提取**：支持重点提取特定类型的内容（如地点信息、角色信息等）
- **RAG系统集成**：直接调用CustomRagMemoryService存储到向量数据库
- **ID全链路透传**：支持user_id和archive_id的完整传递

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/nvcnvn/adk-golang/pkg/flows/save_novel_rag_data_workflow"
)

func main() {
    ctx := context.Background()
    userID := "user123"
    archiveID := "novel_001"
    
    content := `
    林小雨踏入了神秘的翠竹林深处，古老的石碑上刻着模糊的文字。
    她小心翼翼地走向前方，突然听到了奇怪的低语声。
    "这里曾经是修仙者的练功之地。" 一个苍老的声音从石碑后传来。
    `
    
    // 快速保存内容
    savedCount, err := save_novel_rag_data_workflow.QuickSave(ctx, userID, archiveID, content)
    if err != nil {
        log.Printf("保存失败: %v", err)
        return
    }
    
    log.Printf("成功保存 %d 个段落到RAG系统", savedCount)
}
```

### 特定内容类型提取

```go
// 重点提取地点信息
savedCount, err := save_novel_rag_data_workflow.QuickSave(ctx, userID, archiveID, content, "location")

// 重点提取角色信息  
savedCount, err := save_novel_rag_data_workflow.QuickSave(ctx, userID, archiveID, content, "character")

// 重点提取对话内容
savedCount, err := save_novel_rag_data_workflow.QuickSave(ctx, userID, archiveID, content, "dialogue")
```

### 高级配置

```go
// 使用自定义配置
options := save_novel_rag_data_workflow.SaveNovelRagDataOptions{
    RAGBaseURL: "http://localhost:18000", // 自定义RAG服务地址
    RAGTopK:    15,                       // 自定义TopK值
    LLMModel:   "gpt-4",                  // 使用不同的LLM模型
    UserID:     "premium_user",
    ArchiveID:  "premium_archive",
}

service := save_novel_rag_data_workflow.NewSaveNovelRagDataServiceWithOptions(options)
savedCount, err := service.SaveNovelData(ctx, content, "setting")
```

## API 文档

### 核心类型

#### SaveNovelRagDataConfig

```go
type SaveNovelRagDataConfig struct {
    MemoryService memory.MemoryService // RAG内存服务实例
    LLMModel      string               // 用于内容分析的LLM模型名称
    UserID        string               // 用户ID，对应RAG的tenant_id
    ArchiveID     string               // 归档ID，对应RAG的session_id
}
```

#### Segment

```go
type Segment struct {
    Content string `json:"content"`          // 段落内容
    Type    string `json:"type,omitempty"`    // 内容类型：character, location, plot, dialogue, setting, other
    Summary string `json:"summary,omitempty"` // 段落摘要
}
```

### 主要函数

#### QuickSave

```go
func QuickSave(ctx context.Context, userID, archiveID, content string, contentType ...string) (int, error)
```

快速保存小说内容的便利函数。

**参数：**
- `ctx`: 上下文
- `userID`: 用户ID
- `archiveID`: 归档ID
- `content`: 要保存的小说内容
- `contentType`: 可选，指定要重点提取的内容类型

**返回值：**
- `int`: 成功保存的段落数量
- `error`: 错误信息

#### SaveNovelData

```go
func (s *SaveNovelRagDataService) SaveNovelData(ctx context.Context, content string, contentType ...string) (int, error)
```

保存小说数据到RAG系统。

**参数：**
- `ctx`: 上下文
- `content`: 需要保存的小说内容
- `contentType`: 可选，指定要重点提取的内容类型

**返回值：**
- `int`: 成功保存的段落数量
- `error`: 错误信息

## 支持的内容类型

- `character`: 角色相关信息
- `location`: 地点、场景信息
- `plot`: 剧情、情节信息
- `dialogue`: 对话内容
- `setting`: 世界观、设定信息
- `other`: 其他类型内容

## 技术架构

### 数据流程

1. **输入处理**：接收小说内容字符串和可选的内容类型参数
2. **LLM分析**：使用指定的LLM模型对内容进行分段和类型识别
3. **JSON解析**：解析LLM返回的JSON格式结果
4. **数据存储**：将每个段落转换为Session格式，调用CustomRagMemoryService存储
5. **结果返回**：返回成功存储的段落数量

### 依赖组件

- **agents框架**：用于LLM调用和内容分析
- **memory.CustomRagMemoryService**：RAG向量数据库接口
- **sessions包**：会话数据结构定义

### 错误处理

- 输入验证：检查内容是否为空
- JSON解析：自动提取和清理LLM响应中的JSON部分
- 存储错误：单个段落存储失败不影响其他段落
- 详细日志：提供完整的处理过程日志

## 配置说明

### 默认配置

- **RAG服务地址**：`http://localhost:18000`
- **LLM模型**：`deepseek-chat`
- **相似度TopK**：`10`

### 自定义配置

可以通过以下方式自定义配置：

1. **使用选项构造函数**：`NewSaveNovelRagDataServiceWithOptions`
2. **指定RAG URL**：`QuickSaveWithRAGURL`
3. **提供自定义MemoryService**：直接传入实现了`memory.MemoryService`接口的实例

## 性能考虑

- **并发安全**：支持多协程并发调用
- **内存优化**：流式处理，不缓存大量数据
- **错误恢复**：单个段落失败不影响整体流程
- **超时控制**：支持context超时控制

## 测试

运行测试：

```bash
go test ./pkg/flows/save_novel_rag_data_workflow/
```

运行基准测试：

```bash
go test -bench=. ./pkg/flows/save_novel_rag_data_workflow/
```

## 使用示例

完整的使用示例请参考 `example_test.go` 文件，包括：

- 基本用法示例
- 特定内容类型提取示例
- 服务实例批量处理示例
- 自定义配置示例
- 错误处理示例

## 注意事项

1. **RAG服务依赖**：需要确保RAG服务（FastAPI + Milvus）正常运行
2. **LLM模型可用性**：确保指定的LLM模型可访问
3. **网络连接**：需要稳定的网络连接访问RAG服务
4. **内容长度**：过长的内容可能需要预先分段处理
5. **JSON格式**：LLM返回的内容必须包含有效的JSON格式

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基本的内容分段和存储功能
- 支持特定内容类型提取
- 提供便利函数和自定义配置选项
- 完整的错误处理和测试覆盖
