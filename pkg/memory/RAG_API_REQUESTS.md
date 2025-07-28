# API 请求文档

本文档说明了 `Multi-Tenant RAG Retrieval Service` 提供的所有 HTTP 接口、调用方式、请求 / 响应参数及示例，方便集成与调试。

> 基础地址示例：`http://<HOST>:<PORT>`
> 
> - Docker Compose 默认映射为 `http://localhost:18000`。
> - 以下示例均基于该地址，可根据部署环境自行替换。

---

## 1. POST /add_session

将一段对话（多条消息）写入指定租户的向量集合中。

| 属性           | 类型               | 必填 | 说明                                                         |
|----------------|--------------------|------|--------------------------------------------------------------|
| tenant_id      | string             | ✅   | **租户 ID**。用于区分不同客户的数据隔离。                   |
| session_id     | string             | ✅   | **会话 ID**。同一会话的多条消息可多次调用写入。             |
| messages       | Message[]          | ✅   | 消息列表，至少 1 条。每条消息需包含 `role` 与 `content`。    |
| metadata       | object (dict)      | ❌   | 业务自定义元数据，可选。                                      |

### Message

| 属性  | 类型   | 必填 | 说明                                     |
|-------|--------|------|------------------------------------------|
| role  | string | ✅   | 消息角色，如 `user` / `assistant` / …     |
| content | string | ✅ | 消息内容文本                              |

### 请求示例（cURL）
```bash
curl -X POST http://localhost:18000/add_session \
  -H "Content-Type: application/json" \
  -d '{
        "tenant_id": "tenant1",
        "session_id": "session1",
        "messages": [
          {"role": "user", "content": "你好"},
          {"role": "assistant", "content": "你好，有什么可以帮助你的吗？"}
        ]
      }'
```

### 响应
```json
{
  "success": true
}
```

---

## 2. POST /search_memory

在指定租户的数据集中进行语义搜索，返回最相似的消息片段。

| 属性      | 类型   | 必填 | 默认值 | 说明                                                             |
|-----------|--------|------|--------|------------------------------------------------------------------|
| tenant_id | string | ✅   | —      | **租户 ID**，与写入时保持一致。                                  |
| query     | string | ✅   | —      | 检索文本。                                                       |
| top_k     | int    | ❌   | `10`   | 返回前 `k` 条最相似结果。                                        |
| filter    | object | ❌   | `null` | 预留字段，按需支持 Milvus 的 server-side 过滤表达式。             |

### 请求示例（cURL）
```bash
curl -X POST http://localhost:18000/search_memory \
  -H "Content-Type: application/json" \
  -d '{
        "tenant_id": "tenant1",
        "query": "你好",
        "top_k": 5
      }'
```

### 响应
```json
{
  "results": [
    {
      "content": "你好",
      "score": 638.2063
    },
    {
      "content": "你好，有什么可以帮助你的吗？",
      "score": 553.54114
    }
  ],
  "elapsed_ms": 10.91
}
```
- `score` 为相似度（Inner Product），值越大越相似。
- `elapsed_ms` 为本次搜索耗时（毫秒）。

---

## 3. GET /health

Milvus 连接健康检查。

### 请求
```
GET http://localhost:18000/health
```

### 响应
```json
{
  "status": "healthy"
}
```
- 若无法连接 Milvus，将返回 `503 Service Unavailable`，`detail` 字段包含具体错误描述。

---

## 4. 错误响应格式

统一使用 FastAPI `HTTPException`，返回示例如下：

```json
{
  "detail": "错误描述"
}
```

常见 HTTP 状态码：
- `400 Bad Request`  参数校验失败
- `404 Not Found`    资源不存在
- `500 Internal Server Error`  服务器内部错误
- `503 Service Unavailable`    依赖服务不可用（如 Milvus）

---

## 5. 认证与安全

当前示例项目未内置身份认证逻辑，如需在生产环境使用，请务必：
1. 在反向代理（如 Nginx、API Gateway）或应用层加入鉴权拦截。
2. 考虑为敏感接口（写入 / 搜索）添加 **Token / OAuth2 / JWT** 等认证机制。

---

## 6. 附录

### 环境变量
| 变量名              | 说明                                          | 默认值                              |
|---------------------|-----------------------------------------------|-------------------------------------|
| `MILVUS_HOST`       | Milvus 服务地址                               | `milvus`                            |
| `MILVUS_PORT`       | Milvus 服务端口                               | `19530`                             |
| `EMBED_SERVICE_URLS`| Embedding 服务地址，支持逗号分隔多个URL        | `http://embed-service:8001/embed`    |

> 完整系统部署说明请参考根目录 `README.md`。
