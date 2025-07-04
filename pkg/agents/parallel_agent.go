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

// Package agents provides the core agent types and functionality.
package agents

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ParallelAgent runs its sub-agents in parallel and aggregates their responses.
type ParallelAgent struct {
	Agent
	subAgents []*Agent
	workers   int
}

// ParallelAgentConfig holds configuration for creating a ParallelAgent.
type ParallelAgentConfig struct {
	Name        string
	Description string
	SubAgents   []*Agent
	Workers     int // 最大并发工作协程数，<=0 为 len(SubAgents)
}

// MultiError 聚合并行执行过程中发生的多个错误。
// 当 len(errs)>0 时返回，用于让调用方能看到全部错误而非首个错误。
// 如果仅有一个错误，也会保留在切片中，保持一致性。
// 注意: 如果所有 err 均为 nil，可直接传空 slice 即无需使用 MultiError。
// implements: error
// 用 fmt.Sprintf 生成整合后的错误信息，从而保证 fmt 导入被有效使用。

type MultiError []error

// Error 返回格式化后的错误信息。例如: "encountered 2 error(s): err1; err2"
func (m MultiError) Error() string {
	if len(m) == 0 {
		return ""
	}
	var parts []string
	for _, err := range m {
		if err != nil {
			parts = append(parts, err.Error())
		}
	}
	return fmt.Sprintf("encountered %d error(s): %s", len(parts), strings.Join(parts, "; "))
}

// NewParallelAgent creates a new agent that processes sub-agents in parallel.
func NewParallelAgent(config ParallelAgentConfig) *ParallelAgent {
    workers := config.Workers
    if workers <= 0 {
        workers = len(config.SubAgents)
    }
    return &ParallelAgent{
        Agent: Agent{
            name:        config.Name,
            description: config.Description,
        },
        subAgents: config.SubAgents,
        workers:   workers,
    }
}

// Process 处理输入消息，按配置的 worker 数并发执行所有子 Agent，收敛错误并支持 ctx 取消。
func (a *ParallelAgent) Process(ctx context.Context, message string) (string, error) {
    if len(a.subAgents) == 0 {
        return "", nil
    }

    if a.workers <= 0 {
        a.workers = len(a.subAgents)
    }

    var (
        wg        sync.WaitGroup
        sem       = make(chan struct{}, a.workers)
        mu        sync.Mutex
        responses []string
        errs      []error
    )

    for _, subAgent := range a.subAgents {
        // 若上层已经取消，则提前退出
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
        }

        wg.Add(1)
        go func(sa *Agent) {
            defer wg.Done()

            // 限流：获取 token
            sem <- struct{}{}
            defer func() { <-sem }()

            resp, err := sa.Process(ctx, message)

            mu.Lock()
            defer mu.Unlock()
            if err != nil {
                errs = append(errs, err)
                return
            }
            responses = append(responses, resp)
        }(subAgent)
    }

    wg.Wait()

    combined := strings.Join(responses, "\n")
    if len(errs) > 0 {
        return combined, MultiError(errs)
    }
    return combined, nil
}

// SubAgents returns the sub-agents of this parallel agent.
func (a *ParallelAgent) SubAgents() []*Agent {
	return a.subAgents
}
