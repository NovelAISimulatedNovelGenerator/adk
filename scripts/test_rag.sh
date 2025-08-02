#!/bin/sh
# 自动向 /api/execute 发送请求并把返回结果写入 scripts/test_logs 目录，
# 文件名使用当前「小时分钟秒时」时间戳 (HHMMSS).

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/test_logs"
mkdir -p "$LOG_DIR"

TS="$(date +%H%M%S)"
OUTPUT_FILE="$LOG_DIR/${TS}.json"

echo "Requesting workflow, will write response to $OUTPUT_FILE"

curl -s -X POST "http://localhost:8080/api/execute" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow": "test_rag_tool_flow",
    "input": "请测试",
    "user_id": "test_user_123",
    "archive_id": "archive_456",
    "timeout": 60
  }' > "$OUTPUT_FILE"

echo "Saved response to $OUTPUT_FILE"