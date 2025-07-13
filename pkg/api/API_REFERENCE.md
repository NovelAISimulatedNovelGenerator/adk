# ADK 工作流 HTTP API 参考

该文档介绍 `pkg/api` 包所暴露的 HTTP 接口，供外部系统集成调用。

> 默认示例以 `localhost:8080` 为基准，请根据部署环境调整。

---

## 基本信息

| 属性 | 说明 |
| ---- | ---- |
| Base URL | `http://<host>:<port>` |
| 编码 | 请求与响应均使用 `application/json; charset=utf-8`（除流式接口） |
| 鉴权 | 当前版本未内置鉴权，建议通过上层网关实现 |
| 版本 | `v1`（随接口稳定度变化） |

---

## 健康检查

`GET /health`

```bash
curl http://localhost:8080/health
```

成功响应示例：

```json
{
    "status": "ok",
    "version": "1.0.0",
    "time": "2025-07-12T07:10:00Z",
    "workflows": 3,
    "workflow_names": ["novel_v4", "chat", "summary"]
}
```

---

## 列出工作流

`GET /api/workflows`

```bash
curl http://localhost:8080/api/workflows
```

成功响应：

```json
{
    "workflows": ["novel_v4", "chat"],
    "count": 2
}
```

---

## 获取指定工作流信息

`GET /api/workflows/{name}`

```bash
curl http://localhost:8080/api/workflows/novel_v4
```

成功响应：

```json
{
    "name": "novel_v4",
    "description": "Novel generation agent v4",
    "model": "gpt-4o",
    "type": "sequential"
}
```

`404 Not Found`：工作流不存在。

---

## 同步执行工作流

`POST /api/execute`

请求头：`Content-Type: application/json`

### 请求体字段

| 字段 | 类型 | 必填 | 说明 |
| ---- | ---- | ---- | ---- |
| `workflow` | string | ✔ | 目标工作流名称 |
| `input` | string | ✔ | 输入文本 |
| `user_id` | string | ✖ | 调用方用户标识 |
| `experiment_id` | string | ✖ | 实验/灰度标识 |
| `trace_id` | string | ✖ | 自定义链路 ID（若为空服务端自动生成） |
| `parameters` | object | ✖ | 任务额外参数（由具体工作流自行解析） |
| `timeout` | int | ✖ | 超时（秒），默认 30s |

### 请求示例

```bash
curl -X POST http://localhost:8080/api/execute \
  -H "Content-Type: application/json" \
  -d '{
        "workflow": "novel_v4",
        "input": "写一篇科幻短篇",
        "user_id": "u123",
        "timeout": 60
      }'
```

### 成功响应

```json
{
    "workflow": "novel_v4",
    "output": "……",
    "success": true,
    "process_time_ms": 5432,
    "trace_id": "adk-64ae…",
    "metadata": {
        "user_id": "u123",
        "workflow": "novel_v4",
        "experiment_id": ""
    }
}
```

### 失败响应常见格式

```json
{
    "workflow": "novel_v4",
    "success": false,
    "message": "工作流未找到",
    "trace_id": "adk-64ae…"
}
```

| HTTP 状态码 | 含义 |
| ----------- | ---- |
| `200` | 请求成功，字段 `success` 决定业务状态 |
| `400` | 参数错误 (`ErrInvalidRequest`) |
| `404` | 工作流不存在 (`ErrWorkflowNotFound`) |
| `429` | 队列已满/系统繁忙 (`scheduler.ErrQueueFull`) |
| `500` | 内部错误 (`ErrInternalError`) |

---

## 流式执行工作流 (SSE)

`POST /api/stream`

- 服务端采用 **Server-Sent Events** 协议推送数据。
- 请求体字段同 `/api/execute`。
- 客户端需在请求头中将 `Accept` 设为 `text/event-stream`（浏览器 `EventSource` 会自动添加）。

### 事件格式

| 事件名 | 说明 |
| ------ | ---- |
| `data` | 工作流部分输出（可能多次触发） |
| `done` | 工作流完成时触发，`data` 字段包含最终结果 |
| `error` | 发生错误，`data` 字段为错误 JSON |

### 示例 (cURL)

```bash
curl -N -X POST http://localhost:8080/api/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"workflow":"novel_v4","input":"写一首诗"}'
```

接收示例：

```text
event: data
data: 火星在夜空中燃烧……

event: data
data: ……

event: done
data: {"result":"全文完成"}
```

---

## 错误码一览

| 业务错误码 | HTTP 状态 | 说明 |
| ---------- | -------- | ---- |
| `ErrWorkflowNotFound` | 404 | 工作流名称无效 |
| `ErrInvalidRequest` | 400 | 请求格式或参数错误 |
| `ErrInternalError` | 500 | 服务内部错误 |

---

## 最佳实践

1. **超时控制**：合理设置 `timeout`，并在客户端也做超时兜底。
2. **幂等性**：可使用自定义 `trace_id` 关联一次业务调用，便于重试与排障。
3. **并发限制**：若大量高并发调用，建议在应用侧加入排队或限流，以防调度器队列耗尽导致 `429`。
4. **版本兼容**：接口升级将遵循 SemVer 原则，破坏性变更会在主版本升级时发布并在文档中标注。

---

如有疑问或新需求，请与 ADK 团队联系。
