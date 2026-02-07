![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)

# Claude-Code in GO 
An autonomous AI coding assistant built in Go that can read files, write code, and execute shell commands through LLM-powered tool calling and agent loops.

## Overview

This project implements the core architecture of AI coding assistants like Claude Code and GitHub Copilot Workspace. It demonstrates how Large Language Models can autonomously complete programming tasks by:

- Making decisions about which tools to use
- Executing file operations and shell commands
- Iterating until tasks are complete
- Maintaining conversation context across multiple steps

## Features

- 🔄 **Autonomous Agent Loop** - Continuously thinks, acts, and observes until task completion
- 📁 **File Operations** - Read and write files with full content manipulation
- 💻 **Command Execution** - Run shell commands and capture output
- 🧠 **LLM Tool Calling** - JSON schema-based tool definitions for structured AI interactions
- 📝 **Conversation History** - Maintains context across multiple turns
- ⚡ **Built with Go** - Leverages Go's simplicity and performance

## How It Works
```
User Prompt → LLM → Tool Call Decision → Execute Tool → Return Result → LLM → ...
                ↑                                                              ↓
                └──────────────── Agent Loop (repeats until done) ────────────┘
```

1. **User provides a prompt** (e.g., "Refactor main.go to use error handling")
2. **LLM analyzes** the request and decides which tools to call
3. **Tools execute** (Read files, Write files, Run commands)
4. **Results feed back** into the LLM for next decision
5. **Loop continues** until task is complete

## Available Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `Read` | Read file contents | `file_path` |
| `Write` | Write content to file | `file_path`, `content` |
| `Bash` | Execute shell commands | `command` |

## Installation

### Prerequisites
- Go 1.21 or higher
- OpenRouter API key (or any OpenAI-compatible API)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/ai-coding-agent-go.git
cd ai-coding-agent-go
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export OPENROUTER_API_KEY="your-api-key-here"
export OPENROUTER_BASE_URL="https://openrouter.ai/api/v1"  # Optional
```

## Usage

### Basic Example
```bash
go run main.go -p "Create a hello.txt file with 'Hello, World!'"
```

### Real-World Examples

**Refactor code:**
```bash
go run main.go -p "Read main.go and refactor it to use better error handling"
```

**Analyze and fix:**
```bash
go run main.go -p "Find all .go files, check for unused imports, and remove them"
```

**Multi-step task:**
```bash
go run main.go -p "Create a new Go module called 'utils', add a StringReverse function, write tests, and run them"
```

## Architecture

### Agent Loop
The core of the agent is an infinite loop that:
1. Sends messages to the LLM
2. Receives tool calls from the LLM
3. Executes tools locally
4. Appends results to conversation history
5. Repeats until no more tool calls

### Tool Calling Flow
```go
// 1. Define tools as JSON schemas
tools := []openai.ChatCompletionToolUnionParam{...}

// 2. LLM returns tool calls
message := resp.Choices[0].Message

// 3. Execute tools
for _, toolCall := range message.ToolCalls {
    result := executeTool(toolCall)
    // 4. Feed results back to LLM
    messages = append(messages, result)
}
```

### Conversation Management
Messages are stored in a growing slice that maintains full context:
- User messages
- Assistant responses
- Tool calls
- Tool results

## Technical Deep Dive

### Why Go?
- **Concurrency primitives** for potential parallel tool execution
- **Strong typing** for safe JSON schema handling
- **Single binary** deployment
- **Excellent stdlib** for file I/O and command execution

### Key Design Decisions

**Tool Execution Safety:**
- Commands are split using `strings.Fields()` (basic protection)
- File operations use standard permissions (0644)
- All errors are captured and returned to the LLM

**Context Window Management:**
- Full conversation history maintained
- No truncation (future improvement needed for long conversations)

**Error Handling:**
- Tool errors are returned as text to the LLM
- LLM can see errors and adjust strategy

## Limitations & Future Improvements

- [ ] Add conversation history truncation for long tasks
- [ ] Implement sandboxing for command execution
- [ ] Add support for Model Context Protocol (MCP)
- [ ] Integrate LSP for code understanding
- [ ] Build interactive TUI
- [ ] Add streaming responses
- [ ] Support for multiple LLM providers
- [ ] Tool execution parallelization with goroutines

## What I Learned

- How tool calling works under the hood (JSON schemas → function execution)
- Implementing autonomous agent loops that don't get stuck
- Managing LLM context windows and conversation history
- Safe command execution and file manipulation in Go
- Designing extensible tool systems

## Built With

- [openai-go](https://github.com/openai/openai-go) - OpenAI SDK for Go
- [OpenRouter](https://openrouter.ai/) - LLM API gateway
- Claude Haiku 4.5 - Default model (configurable)

## Inspired By

This project was built as part of the [CodeCrafters](https://codecrafters.io/) AI Agent challenge, which teaches how tools like Claude Code work internally.

## Contributing

Contributions are welcome! Areas for improvement:
- Additional tools (Git operations, HTTP requests, etc.)
- Better error handling and recovery
- Interactive mode
- Configuration file support

## License

MIT License - see LICENSE file for details

## Acknowledgments

- CodeCrafters for the excellent challenge structure
- Anthropic for Claude and inspiration from Claude Code
- The Go community for amazing libraries

---

**⭐ If you found this helpful, consider giving it a star!**

**💬 Questions?** Open an issue or reach out on [LinkedIn](your-linkedin-url)
```

---

## **Additional Files to Include:**

### **.gitignore**
```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
ai-coding-agent-go

# Test binary
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# Environment variables
.env

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
```

### **LICENSE** (MIT License example)
```
MIT License

Copyright (c) 2025 [Your Name]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction...
