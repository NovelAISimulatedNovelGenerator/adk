# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Commands

### Build & Development
```bash
# Build the main ADK CLI
go build -o adk ./cmd/adk

# Run tests
go test ./...                     # Run all tests
go test ./pkg/flows/novel_v4      # Run specific package tests
go test -bench=. ./...           # Run benchmarks

# Build plugin (for flows)
go build -buildmode=plugin -o plugins/novel_flow_v4.so ./flows/novel_v4

# Build individual components
go build -o apiserver ./cmd/apiserver
go build -o monitor ./cmd/monitor
```

### CLI Usage
```bash
# Run an agent interactively
./adk run ./flows/novel_v4        # Run novel v4 flow
./adk run ./examples/search_assistant

# Start web interface
./adk web --port 8080

# Start API server only
./adk api_server --port 8000

# List available commands
./adk --help
```

### Development Workflow
```bash
# Standard Go development
go mod tidy                    # Clean dependencies
go fmt ./...                   # Format code
go vet ./...                   # Static analysis
go test -race ./...           # Race condition testing
```

## Architecture Overview

### Project Structure
- **cmd/**: Main entry points for different applications
  - `adk/`: Main CLI tool
  - `apiserver/`: Standalone API server
  - `monitor/`: Monitoring service
- **pkg/**: Core library packages
  - **agents/**: Agent composition framework (Sequential, Parallel, Loop, Remote)
  - **flows/**: High-level workflows for specific tasks
  - **models/**: Model abstractions (DeepSeek, Gemini, VertexAI)
  - **tools/**: Built-in tools (search, code execution, retrieval, etc.)
  - **memory/**: Vector memory systems
  - **evaluation/**: Agent evaluation framework
- **examples/**: Sample implementations
- **flows/**: Executable agent flows (as plugins)

### Key Design Patterns
- **Agent Architecture**: Hierarchical agents with clear separation between decision-layer (strategy/planning) and execution-layer (domain-specific agents)
- **Plugin System**: Flows compiled as Go plugins loaded at runtime from `./plugins/`
- **Model Abstraction**: Unified interface for different LLM providers
- **Event-Driven**: Streaming events throughout agent execution
- **Configurable**: YAML-based configuration with support for model API pools

### Core Components Flow
```
CLI (---> Agents (Composition) ---> Models ---> Tools
  ↓
Web/API ---> Runners ---> Events/Evaluation/Memory
```

### Configuration
- Primary: `config.yaml` (copy from `config.example.yaml`)
- Environment: `ADK_CONFIG` points to config file
- Key sections: database, queue, model_api_pools, log_level

### Agent Development
1. Create agent package in `pkg/flows/`
2. Implement `Build() *agents.Agent` function
3. Register model via `models.GetRegistry().Register()`
4. Export agent via `agents.Export(agent)`
5. Build as plugin: `go build -buildmode=plugin`

### Testing Patterns
```bash
# Unit tests for agents
go test ./pkg/flows/novel_v4/...

# API endpoint tests
go test ./pkg/api/...

# Performance benchmarks
go test -bench=. ./pkg/flows/novel_v4/...

# Specific test cases
TestBuild, TestArchitectAgent, TestWriterAgent, TestCoordinatorAgent
```

### Deployment
```bash
# Cloud Run deployment (WIP)
./adk deploy cloud_run ./flows/novel_v4 --project my-project --region us-central1

# Manual deployment
go build -o adk ./cmd/adk
./adk web --port $PORT --log_level info
```

### Model Pool Configuration
Enable load balancing across multiple API endpoints:

```yaml
model_api_pools:
  deepseek_pool:
    base: deepseek
    endpoints:
      - url: "https://api1.example.com/v1"
        apikey: "sk-xxxx1"
      - url: "https://api2.example.com/v1"
        apikey: "sk-xxxx2"
```

Use in flow: `"pool:deepseek_pool"