#!/usr/bin/env bash
# scaffold_flow.sh
# 根据 pkg/flows/<name> 目录，生成对应的可编译插件目录 flows/<name>
# 用法： ./scripts/scaffold_flow.sh <flow_name>
# 例如： ./scripts/scaffold_flow.sh novel
# -------------------------------------------
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "用法: $0 <flow_name>" >&2
  exit 1
fi

FLOW_NAME="$1"
PKG_DIR="$(dirname "$0")/../pkg/flows/${FLOW_NAME}"
TARGET_DIR="$(dirname "$0")/../flows/${FLOW_NAME}"

if [ ! -d "$PKG_DIR" ]; then
  echo "错误: ${PKG_DIR} 不存在" >&2
  exit 1
fi

# 验证 pkg/flows/<name> 是否包含 Build() 函数
if ! grep -R "func \\(.*\\)\?Build(" "$PKG_DIR"/*.go >/dev/null 2>&1; then
  echo "错误: ${PKG_DIR} 下未找到 Build() 函数，无法生成插件" >&2
  exit 1
fi

mkdir -p "$TARGET_DIR"

# 生成 main.go
cat >"${TARGET_DIR}/main.go" <<EOF
package main

import (
    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/flow"
    ${FLOW_NAME} "github.com/nvcnvn/adk-golang/pkg/flows/${FLOW_NAME}"
)

type pluginImpl struct{}

func (p *pluginImpl) Name() string { return "${FLOW_NAME}_flow" }

func (p *pluginImpl) Build() (*agents.Agent, error) { return ${FLOW_NAME}.Build(), nil }

var Plugin flow.FlowPlugin = &pluginImpl{}
EOF

# 生成 logger_inject.go（若不存在）
cat >"${TARGET_DIR}/logger_inject.go" <<'EOF'
package main

import "go.uber.org/zap"

var plgLog *zap.Logger = zap.NewNop()

func SetLogger(l *zap.Logger) { plgLog = l }

func L() *zap.Logger { return plgLog }
EOF

echo "已生成 ${TARGET_DIR}"
