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

// Package main provides an example search assistant agent implementation.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nvcnvn/adk-golang/pkg/agents"
	"github.com/nvcnvn/adk-golang/pkg/tools"
	"github.com/nvcnvn/adk-golang/pkg/models"
)

func main() {
	// Define the search assistant agent
	searchAgent := agents.NewAgent(
		agents.WithName("search_assistant"),
		agents.WithModel("deepseek-chat"), // 使用 DeepSeek 模型
		agents.WithInstruction("You are a helpful assistant. Answer user questions using Google Search when needed."),
		agents.WithDescription("An assistant that can search the web."),
		agents.WithTools(tools.GoogleSearch),
	)

	// Register DeepSeek model in the standard registry so agent.Process can find it
	if _, ok := models.GetRegistry().Get("deepseek-chat"); !ok {
		if m, err := models.NewDeepSeekModel("deepseek-chat"); err == nil {
			models.GetRegistry().Register(m)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to init DeepSeek model: %v\n", err)
		}
	}

	// Export for potential external CLI usage (optional)
	agents.Export(searchAgent)

	// Interactive demo
	if len(os.Args) > 1 && os.Args[1] == "run" {
		fmt.Println("Starting interactive session with search assistant (DeepSeek)")
		fmt.Println("Type your query or 'exit' to quit")

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "exit" {
				break
			}

			resp, err := searchAgent.Process(context.Background(), input)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Println("Agent:", resp)
		}
	}
}
