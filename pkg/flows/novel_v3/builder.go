package novel_v3

import (
    novel "github.com/nvcnvn/adk-golang/pkg/flows/novel"
    "github.com/nvcnvn/adk-golang/pkg/agents"
)

// Build 返回版本 v3 的顶层 Agent。目前直接复用经过充分测试的
// pkg/flows/novel.Build 结果，确保插件能够成功加载并运行。
func Build() *agents.Agent {
    return novel.Build()
}
