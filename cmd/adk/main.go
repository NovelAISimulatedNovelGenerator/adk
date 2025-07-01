// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main implements the ADK CLI.
/*
# NovelAI 分层智能体系统设计

## 1. 整体架构

NovelAI 分层智能体系统采用"决策-执行"双层架构，通过清晰的职责分离实现小说内容的高质量生成。

```
┌───────────────────────────────────────────────────────────────┐
│                     决策层 (Decision Layer)                    │
│                                                               │
│  ┌─────────────────┐    ┌─────────────────┐    ┌────────────┐ │
│  │  策略Agent      │    │  规划Agent      │    │ 评估Agent   │ │
│  │ (Strategy)      │◄───►│ (Planner)      │◄───►│(Evaluator) │ │
│  └───────┬─────────┘    └────────┬────────┘    └─────┬──────┘ │
└──────────┼──────────────────────┼────────────────────┼────────┘
           │                      │                    │
           ▼                      ▼                    ▼
┌──────────────────────────────────────────────────────────────┐
│                    执行层 (Execution Layer)                   │
│                                                              │
│  ┌────────────────┐   ┌────────────────┐   ┌───────────────┐ │
│  │  世界观Agent   │   │   角色Agent     │   │  剧情Agent    │ │
│  │  (Worldview)   │   │  (Character)   │   │   (Plot)      │ │
│  └────────────────┘   └────────────────┘   └───────────────┘ │
│                                                              │
│  ┌────────────────┐   ┌────────────────┐   ┌───────────────┐ │
│  │   对话Agent    │   │  背景Agent     │   │ JSON格式化Agent│ │
│  │  (Dialogue)    │   │ (Background)   │   │  (Formatter)  │ │
│  └────────────────┘   └────────────────┘   └───────────────┘ │
└──────────────────────────────────────────────────────────────┘

*/
package main

import (
    "fmt"
    "os"
    "time"
    "github.com/nvcnvn/adk-golang/pkg/logger"

    "github.com/nvcnvn/adk-golang/pkg/agents"
    "github.com/nvcnvn/adk-golang/pkg/cli"
    "github.com/nvcnvn/adk-golang/pkg/config"
    "github.com/nvcnvn/adk-golang/pkg/flow"
    novel "github.com/nvcnvn/adk-golang/pkg/flows/novel"
    "github.com/nvcnvn/adk-golang/pkg/models"
)

// --- NovelAI DeepSeek Agent Framework (plugin) ---
func buildNovelAIFramework() *agents.Agent {
	return novel.Build()
}

func init() {
    // 加载配置文件
    cfgPath := os.Getenv("ADK_CONFIG")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
        os.Exit(1)
    }

    // 初始化结构化日志
    _, _ = logger.Init(cfg.LogLevel, cfg.LogDev)

    // 初始化插件 Manager 与 Loader
    mgr := flow.NewManager()
    flow.SetGlobalManager(mgr) // 设置全局访问点
    loader, err := flow.NewLoader(cfg.PluginDir, mgr)
    if err != nil {
        fmt.Fprintf(os.Stderr, "初始化插件 Loader 失败: %v\n", err)
        os.Exit(1)
    }
    loader.Start()

    // 等待初始插件加载完成一个时间窗，简单处理
    time.Sleep(300 * time.Millisecond)

    // 如果默认 flow 已加载，则导出，供 CLI 使用
    if agent, ok := mgr.Get(cfg.DefaultFlow); ok {
        agents.Export(agent)
    } else {
        // 退化为内置 novel.Build()
        agents.Export(buildNovelAIFramework())
    }

	// 注册 DeepSeek 模型，确保可用
	if _, ok := models.GetRegistry().Get("deepseek-chat"); !ok {
		if m, err := models.NewDeepSeekModel("deepseek-chat"); err == nil {
			models.GetRegistry().Register(m)
		}
	}

	// 导出智能体，供 CLI 或外部调用
	agents.Export(buildNovelAIFramework())
}

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
