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
	"strings"
	"sync"
)

// ParallelAgent runs its sub-agents in parallel and aggregates their responses.
type ParallelAgent struct {
	Agent
	subAgents []*Agent
}

// ParallelAgentConfig holds configuration for creating a ParallelAgent.
type ParallelAgentConfig struct {
	Name        string
	Description string
	SubAgents   []*Agent
}

// NewParallelAgent creates a new agent that processes sub-agents in parallel.
func NewParallelAgent(config ParallelAgentConfig) *ParallelAgent {
	return &ParallelAgent{
		Agent: Agent{
			name:        config.Name,
			description: config.Description,
		},
		subAgents: config.SubAgents,
	}
}

// Process handles a message by processing it through all sub-agents in parallel.
func (a *ParallelAgent) Process(ctx context.Context, message string) (string, error) {
	var wg sync.WaitGroup
	responses := make(chan string, len(a.subAgents))

	for _, subAgent := range a.subAgents {
		wg.Add(1)
		go func(sa *Agent) {
			defer wg.Done()
			response, err := sa.Process(ctx, message)
			if err != nil {
				// TODO: Better error handling
				return
			}
			select {
			case responses <- response:
			case <-ctx.Done():
				return
			}
		}(subAgent)
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	var allResponses []string
	for response := range responses {
		allResponses = append(allResponses, response)
	}

	return strings.Join(allResponses, "\n"), nil
}

// SubAgents returns the sub-agents of this parallel agent.
func (a *ParallelAgent) SubAgents() []*Agent {
	return a.subAgents
}
