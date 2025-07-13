# Artifacts 制品管理模块

## 概述

Artifacts 模块提供了完整的制品存储和管理功能，用于在智能体系统中存储、检索和管理各种类型的数字制品。该模块支持文本和二进制内容的版本化管理，是智能体持久化存储能力的核心组件。

## 核心概念

### Part (制品部分)
```go
type Part struct {
    Text     string // 文本内容
    Data     []byte // 二进制数据
    MimeType string // MIME类型
}
```

表示一个制品的基本单元，可以包含文本或二进制数据，与 Google GenAI 的 Part 概念保持一致。

### ArtifactService (制品服务接口)
```go
type ArtifactService interface {
    // SaveArtifact 保存制品到存储系统
    SaveArtifact(ctx context.Context, appName, userID, sessionID, filename string, artifact Part) (int, error)
    
    // LoadArtifact 从存储系统加载制品
    LoadArtifact(ctx context.Context, appName, userID, sessionID, filename string, version *int) (*Part, error)
    
    // ListArtifactKeys 列出会话中的所有制品文件名
    ListArtifactKeys(ctx context.Context, appName, userID, sessionID string) ([]string, error)
    
    // DeleteArtifact 删除制品
    DeleteArtifact(ctx context.Context, appName, userID, sessionID, filename string) error
    
    // ListVersions 列出制品的所有版本
    ListVersions(ctx context.Context, appName, userID, sessionID, filename string) ([]int, error)
}
```

定义了制品管理的核心操作接口，支持多层级的制品组织结构。

## 实现类型

### 1. InMemoryArtifactService (内存制品服务)
基于内存的制品存储实现：
- 高性能访问
- 适合开发和测试环境
- 进程重启后数据丢失
- 支持完整的版本管理

### 2. GCSArtifactService (Google Cloud Storage制品服务)
基于Google Cloud Storage的制品存储实现：
- 云端持久化存储
- 高可用性和扩展性
- 企业级安全性
- 自动版本管理和备份

## 使用示例

### 基础制品操作
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/nvcnvn/adk-golang/pkg/artifacts"
)

func main() {
    ctx := context.Background()
    
    // 创建内存制品服务
    service := artifacts.NewInMemoryArtifactService()
    
    // 创建文本制品
    textPart := artifacts.FromText("Hello, 这是一个测试文档", "text/plain")
    
    // 保存制品
    version, err := service.SaveArtifact(ctx, "myapp", "user123", "session456", "test.txt", textPart)
    if err != nil {
        log.Fatalf("保存制品失败: %v", err)
    }
    fmt.Printf("制品已保存，版本号: %d\n", version)
    
    // 加载制品（最新版本）
    loadedPart, err := service.LoadArtifact(ctx, "myapp", "user123", "session456", "test.txt", nil)
    if err != nil {
        log.Fatalf("加载制品失败: %v", err)
    }
    fmt.Printf("加载的制品内容: %s\n", loadedPart.Text)
    
    // 列出会话中的所有制品
    keys, err := service.ListArtifactKeys(ctx, "myapp", "user123", "session456")
    if err != nil {
        log.Fatalf("列出制品失败: %v", err)
    }
    fmt.Printf("会话中的制品: %v\n", keys)
}
```

### 二进制制品处理
```go
func handleBinaryArtifacts() {
    ctx := context.Background()
    service := artifacts.NewInMemoryArtifactService()
    
    // 创建二进制制品（例如图片数据）
    imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG头部
    imagePart := artifacts.FromBytes(imageData, "image/png")
    
    // 保存图片制品
    version, err := service.SaveArtifact(ctx, "imageapp", "photographer1", "shoot20250115", "photo1.png", imagePart)
    if err != nil {
        log.Fatalf("保存图片制品失败: %v", err)
    }
    fmt.Printf("图片制品已保存，版本: %d\n", version)
    
    // 加载图片制品
    loadedImage, err := service.LoadArtifact(ctx, "imageapp", "photographer1", "shoot20250115", "photo1.png", nil)
    if err != nil {
        log.Fatalf("加载图片制品失败: %v", err)
    }
    
    fmt.Printf("加载的图片大小: %d 字节\n", len(loadedImage.Data))
    fmt.Printf("图片MIME类型: %s\n", loadedImage.MimeType)
    
    // 可以将数据写入文件
    err = os.WriteFile("downloaded_photo.png", loadedImage.Data, 0644)
    if err != nil {
        log.Printf("保存图片文件失败: %v", err)
    } else {
        fmt.Println("图片已保存到 downloaded_photo.png")
    }
}
```

### 版本管理
```go
func demonstrateVersioning() {
    ctx := context.Background()
    service := artifacts.NewInMemoryArtifactService()
    
    appName := "docapp"
    userID := "writer1"
    sessionID := "draft001"
    filename := "document.md"
    
    // 保存多个版本
    versions := []string{
        "# 文档标题\n\n这是第一版内容。",
        "# 文档标题\n\n这是第一版内容。\n\n## 新增章节\n增加了一些内容。",
        "# 改进的文档标题\n\n这是修订后的第一版内容。\n\n## 新增章节\n增加了更多详细内容。\n\n## 结论\n文档完成。",
    }
    
    var savedVersions []int
    for i, content := range versions {
        part := artifacts.FromText(content, "text/markdown")
        version, err := service.SaveArtifact(ctx, appName, userID, sessionID, filename, part)
        if err != nil {
            log.Fatalf("保存版本 %d 失败: %v", i+1, err)
        }
        savedVersions = append(savedVersions, version)
        fmt.Printf("保存了版本 %d\n", version)
    }
    
    // 列出所有版本
    allVersions, err := service.ListVersions(ctx, appName, userID, sessionID, filename)
    if err != nil {
        log.Fatalf("列出版本失败: %v", err)
    }
    fmt.Printf("所有版本: %v\n", allVersions)
    
    // 加载特定版本
    firstVersion := allVersions[0]
    firstPart, err := service.LoadArtifact(ctx, appName, userID, sessionID, filename, &firstVersion)
    if err != nil {
        log.Fatalf("加载第一版失败: %v", err)
    }
    fmt.Printf("第一版内容:\n%s\n", firstPart.Text)
    
    // 加载最新版本
    latestPart, err := service.LoadArtifact(ctx, appName, userID, sessionID, filename, nil)
    if err != nil {
        log.Fatalf("加载最新版失败: %v", err)
    }
    fmt.Printf("最新版内容:\n%s\n", latestPart.Text)
}
```

### Google Cloud Storage集成
```go
func useGCSArtifactService() {
    ctx := context.Background()
    
    // 创建GCS制品服务
    config := &artifacts.GCSConfig{
        BucketName: "my-artifacts-bucket",
        ProjectID:  "my-gcp-project",
        KeyPath:    "/path/to/service-account-key.json",
    }
    
    service, err := artifacts.NewGCSArtifactService(config)
    if err != nil {
        log.Fatalf("创建GCS制品服务失败: %v", err)
    }
    
    // 使用方式与内存服务完全相同
    part := artifacts.FromText("云端存储的文档内容", "text/plain")
    version, err := service.SaveArtifact(ctx, "cloudapp", "user1", "session1", "cloud-doc.txt", part)
    if err != nil {
        log.Fatalf("保存到云端失败: %v", err)
    }
    
    fmt.Printf("文档已保存到云端，版本: %d\n", version)
    
    // 从云端加载
    cloudPart, err := service.LoadArtifact(ctx, "cloudapp", "user1", "session1", "cloud-doc.txt", nil)
    if err != nil {
        log.Fatalf("从云端加载失败: %v", err)
    }
    
    fmt.Printf("从云端加载的内容: %s\n", cloudPart.Text)
}
```

## 高级功能

### 制品服务工厂
```go
type ArtifactServiceConfig struct {
    Type   string `yaml:"type"`   // "memory" 或 "gcs"
    GCS    *GCSConfig `yaml:"gcs,omitempty"`
}

type GCSConfig struct {
    BucketName string `yaml:"bucket_name"`
    ProjectID  string `yaml:"project_id"`
    KeyPath    string `yaml:"key_path"`
}

func CreateArtifactService(config *ArtifactServiceConfig) (artifacts.ArtifactService, error) {
    switch config.Type {
    case "memory":
        return artifacts.NewInMemoryArtifactService(), nil
    case "gcs":
        if config.GCS == nil {
            return nil, fmt.Errorf("GCS配置不能为空")
        }
        return artifacts.NewGCSArtifactService(config.GCS)
    default:
        return nil, fmt.Errorf("不支持的制品服务类型: %s", config.Type)
    }
}

// 配置文件示例 (artifacts.yaml)
/*
type: "gcs"
gcs:
  bucket_name: "my-company-artifacts"
  project_id: "my-gcp-project"
  key_path: "/etc/gcp/service-account.json"
*/
```

### 制品批量操作
```go
type ArtifactBatch struct {
    service artifacts.ArtifactService
    appName string
    userID  string
    sessionID string
}

func NewArtifactBatch(service artifacts.ArtifactService, appName, userID, sessionID string) *ArtifactBatch {
    return &ArtifactBatch{
        service:   service,
        appName:   appName,
        userID:    userID,
        sessionID: sessionID,
    }
}

func (b *ArtifactBatch) SaveMultiple(ctx context.Context, artifacts map[string]artifacts.Part) (map[string]int, error) {
    results := make(map[string]int)
    
    for filename, part := range artifacts {
        version, err := b.service.SaveArtifact(ctx, b.appName, b.userID, b.sessionID, filename, part)
        if err != nil {
            return nil, fmt.Errorf("保存制品 %s 失败: %w", filename, err)
        }
        results[filename] = version
    }
    
    return results, nil
}

func (b *ArtifactBatch) LoadAll(ctx context.Context) (map[string]*artifacts.Part, error) {
    // 获取所有制品键
    keys, err := b.service.ListArtifactKeys(ctx, b.appName, b.userID, b.sessionID)
    if err != nil {
        return nil, fmt.Errorf("列出制品键失败: %w", err)
    }
    
    results := make(map[string]*artifacts.Part)
    for _, key := range keys {
        part, err := b.service.LoadArtifact(ctx, b.appName, b.userID, b.sessionID, key, nil)
        if err != nil {
            return nil, fmt.Errorf("加载制品 %s 失败: %w", key, err)
        }
        results[key] = part
    }
    
    return results, nil
}

// 使用示例
func batchOperationExample() {
    ctx := context.Background()
    service := artifacts.NewInMemoryArtifactService()
    batch := NewArtifactBatch(service, "batchapp", "user1", "batch-session")
    
    // 批量保存
    artifactsToSave := map[string]artifacts.Part{
        "config.json": artifacts.FromText(`{"version": "1.0"}`, "application/json"),
        "readme.md":   artifacts.FromText("# 项目说明\n\n这是一个测试项目。", "text/markdown"),
        "data.csv":    artifacts.FromText("name,age\nAlice,30\nBob,25", "text/csv"),
    }
    
    versions, err := batch.SaveMultiple(ctx, artifactsToSave)
    if err != nil {
        log.Fatalf("批量保存失败: %v", err)
    }
    
    fmt.Printf("批量保存结果: %v\n", versions)
    
    // 批量加载
    allArtifacts, err := batch.LoadAll(ctx)
    if err != nil {
        log.Fatalf("批量加载失败: %v", err)
    }
    
    for filename, part := range allArtifacts {
        fmt.Printf("制品 %s: %s\n", filename, part.Text)
    }
}
```

### 制品搜索和过滤
```go
type ArtifactFilter struct {
    MimeType   string
    MinVersion int
    MaxVersion int
    Pattern    string // 文件名模式匹配
}

func (b *ArtifactBatch) SearchArtifacts(ctx context.Context, filter *ArtifactFilter) (map[string]*artifacts.Part, error) {
    keys, err := b.service.ListArtifactKeys(ctx, b.appName, b.userID, b.sessionID)
    if err != nil {
        return nil, err
    }
    
    results := make(map[string]*artifacts.Part)
    
    for _, key := range keys {
        // 文件名模式匹配
        if filter.Pattern != "" {
            matched, _ := filepath.Match(filter.Pattern, key)
            if !matched {
                continue
            }
        }
        
        part, err := b.service.LoadArtifact(ctx, b.appName, b.userID, b.sessionID, key, nil)
        if err != nil {
            continue
        }
        
        // MIME类型过滤
        if filter.MimeType != "" && part.MimeType != filter.MimeType {
            continue
        }
        
        results[key] = part
    }
    
    return results, nil
}

// 搜索示例
func searchExample() {
    ctx := context.Background()
    service := artifacts.NewInMemoryArtifactService()
    batch := NewArtifactBatch(service, "searchapp", "user1", "search-session")
    
    // 搜索所有Markdown文件
    filter := &ArtifactFilter{
        MimeType: "text/markdown",
        Pattern:  "*.md",
    }
    
    markdownFiles, err := batch.SearchArtifacts(ctx, filter)
    if err != nil {
        log.Fatalf("搜索失败: %v", err)
    }
    
    fmt.Printf("找到 %d 个Markdown文件:\n", len(markdownFiles))
    for filename, part := range markdownFiles {
        fmt.Printf("- %s (%d 字符)\n", filename, len(part.Text))
    }
}
```

## 性能优化和监控

### 缓存机制
```go
type CachedArtifactService struct {
    underlying artifacts.ArtifactService
    cache      map[string]*artifacts.Part
    mu         sync.RWMutex
    maxSize    int
}

func NewCachedArtifactService(underlying artifacts.ArtifactService, maxSize int) *CachedArtifactService {
    return &CachedArtifactService{
        underlying: underlying,
        cache:      make(map[string]*artifacts.Part),
        maxSize:    maxSize,
    }
}

func (c *CachedArtifactService) LoadArtifact(ctx context.Context, appName, userID, sessionID, filename string, version *int) (*artifacts.Part, error) {
    // 构建缓存键
    key := fmt.Sprintf("%s:%s:%s:%s", appName, userID, sessionID, filename)
    if version != nil {
        key += fmt.Sprintf(":%d", *version)
    }
    
    // 检查缓存
    c.mu.RLock()
    if cached, exists := c.cache[key]; exists {
        c.mu.RUnlock()
        return cached, nil
    }
    c.mu.RUnlock()
    
    // 从底层服务加载
    part, err := c.underlying.LoadArtifact(ctx, appName, userID, sessionID, filename, version)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    c.mu.Lock()
    if len(c.cache) >= c.maxSize {
        // 简单的LRU：删除第一个元素
        for k := range c.cache {
            delete(c.cache, k)
            break
        }
    }
    c.cache[key] = part
    c.mu.Unlock()
    
    return part, nil
}

// 实现其他接口方法...
```

### 监控和指标
```go
type ArtifactMetrics struct {
    TotalSaves    int64         `json:"total_saves"`
    TotalLoads    int64         `json:"total_loads"`
    TotalDeletes  int64         `json:"total_deletes"`
    AvgSaveTime   time.Duration `json:"avg_save_time"`
    AvgLoadTime   time.Duration `json:"avg_load_time"`
    CacheHitRate  float64       `json:"cache_hit_rate"`
    StorageSize   int64         `json:"storage_size_bytes"`
}

type MonitoredArtifactService struct {
    underlying artifacts.ArtifactService
    metrics    *ArtifactMetrics
    mu         sync.RWMutex
}

func (m *MonitoredArtifactService) SaveArtifact(ctx context.Context, appName, userID, sessionID, filename string, artifact artifacts.Part) (int, error) {
    start := time.Now()
    
    version, err := m.underlying.SaveArtifact(ctx, appName, userID, sessionID, filename, artifact)
    
    duration := time.Since(start)
    m.updateSaveMetrics(duration, err == nil)
    
    return version, err
}

func (m *MonitoredArtifactService) updateSaveMetrics(duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if success {
        m.metrics.TotalSaves++
        // 更新平均保存时间
        if m.metrics.TotalSaves == 1 {
            m.metrics.AvgSaveTime = duration
        } else {
            m.metrics.AvgSaveTime = time.Duration(
                (int64(m.metrics.AvgSaveTime)*(m.metrics.TotalSaves-1) + int64(duration)) / m.metrics.TotalSaves,
            )
        }
    }
}

func (m *MonitoredArtifactService) GetMetrics() *ArtifactMetrics {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // 返回副本以避免竞态条件
    return &ArtifactMetrics{
        TotalSaves:   m.metrics.TotalSaves,
        TotalLoads:   m.metrics.TotalLoads,
        TotalDeletes: m.metrics.TotalDeletes,
        AvgSaveTime:  m.metrics.AvgSaveTime,
        AvgLoadTime:  m.metrics.AvgLoadTime,
        CacheHitRate: m.metrics.CacheHitRate,
        StorageSize:  m.metrics.StorageSize,
    }
}
```

## 最佳实践

1. **制品命名**: 使用有意义的文件名和适当的扩展名
2. **版本管理**: 合理使用版本功能，避免过度版本化
3. **MIME类型**: 正确设置MIME类型，便于后续处理
4. **内存管理**: 大型制品应考虑使用流式处理
5. **错误处理**: 实现完善的错误处理和重试机制
6. **安全性**: 对敏感制品进行适当的访问控制

## 依赖关系

- Go 标准库: `context`, `fmt`, `sync`
- Google Cloud Storage SDK (用于GCS实现)
- 项目内部无其他依赖

## 扩展开发

### 自定义制品服务
```go
type CustomArtifactService struct {
    // 自定义字段
}

func (c *CustomArtifactService) SaveArtifact(ctx context.Context, appName, userID, sessionID, filename string, artifact artifacts.Part) (int, error) {
    // 自定义实现
    return 1, nil
}

// 实现其他接口方法...
```

Artifacts 模块为 ADK-Golang 框架提供了强大的制品管理能力，是构建具有持久化存储需求的智能体系统的重要基础设施。
