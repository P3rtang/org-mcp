package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

const MCP_INSTRUCTIONS = `
# Org-MCP Server Instructions

The org-mcp server provides surgical tools for reading and modifying org-mode files. The org file is the **single source of truth** - information not stored in it is lost between sessions.

## Core Principles

1. **Surgical over bulk** - Use targeted queries and updates. Never pass entire file contents.
2. **Tool over direct read** - Use query/manage tools to maintain consistency. Direct reading is for inspection only.
3. **Org file as memory** - Every task, decision, and piece of context belongs in the org file.
4. **CSV output** - Tool responses use CSV for token efficiency. Specify columns to avoid context bloat.

## Available Tools

| Tool | Purpose |
|------|---------|
| query_items | Search/filter headers and bullets with column selection |
| manage_header | Add, update, remove, move headers |
| manage_bullet | Add, update, remove, complete checkboxes |
| manage_text | Add, update, remove plain text blocks |
| manage_table | Add, update, remove table rows/columns |
| manage_codeblock | Add, update, remove, execute code blocks |
| status_overview | Get counts of all statuses and tags |
| vector_search | Semantic search across headers |

## Tool Input Patterns

Tools use a **oneOf union pattern** for operations:
- 'method: "add"' / '"update"' / '"remove"' / '"execute"'
- Pass only fields relevant to the operation (others are ignored)

## Columns Reference

Specify columns to control response size:

| Column | Description |
|--------|-------------|
| UID | Unique identifier (e.g., "12345678" or "12345678.b0.t1") |
| PREVIEW | First ~100 chars, clean text without org syntax |
| CONTENT | Full raw content including org syntax |
| TYPE | Go type (Header, Bullet, PlainText, CodeBlock, Table) |
| STATUS | TODO, NEXT, PROG, DONE, CHECKED, UNCHECKED |
| PROGRESS | X/Y for items with children |
| PARENT | UID of parent item |
| CHILDREN_COUNT | Number of direct children |
| LEVEL | Numeric depth (1 = top-level) |
| TAGS | Comma-separated, quoted for CSV safety |
| PATH | Hierarchical UID stack (e.g., "//ROOT/PARENT/ID") |
| SCHEDULED / DEADLINE / CLOSED | Date columns for scheduling queries |
| LANGUAGE | A field specific to code blocks, specifying the language contained in the block |

## MCP Tasks Extension

For long-running operations (async mode), the server supports the MCP Tasks extension:
- Server advertises 'io.modelcontextprotocol/tasks' capability
- Async calls return 'CreateTaskResult' with taskId
- Poll 'tasks/get' for status updates
- Task states: working → completed / failed / cancelled

## Best Practices

- **UIDs are stable within a single tool call** - Re-query after structural changes
- **Use PATH for navigation** - Token-efficient hierarchy reference
- **Bulk operations are efficient** - Multiple items in one call beats sequential calls
- **Checkboxes track sub-tasks** - Headers track parent task status
- **Tags are inherited** - Child headers inherit parent tags
- **Dates use YYYY-MM-DD format** - For filtering by date ranges
`

type FuncOptions struct {
	DefaultPath string
	Logger      *slog.Logger
}

type ToolFunc func(map[string]any, FuncOptions) ([]any, error)

// Server handles MCP protocol communication over stdio
type Server struct {
	reader    *bufio.Reader
	sender    *MessageSender
	log       *slog.Logger
	state     ServerState
	workspace string

	tools map[string]McpTool
}

// handleSetLoggingLevel handles the setting of logging levels
func (s *Server) handleSetLoggingLevel(id any, params json.RawMessage) {
	var logParams struct {
		Level string `json:"level"`
	}

	if err := json.Unmarshal(params, &logParams); err != nil {
		s.sender.SendError(id, -32602, "Invalid parameters")
		return
	}

	switch logParams.Level {
	case "debug", "info", "warn", "error":
		s.log.Debug(fmt.Sprintf("Logging level set to %s", logParams.Level))
		// Note: Actual logger level adjustment would be implemented here
		s.sender.SendResponse(id, map[string]any{"status": "success"})
	default:
		s.sender.SendError(id, -32602, "Invalid logging level")
	}
}

// ServerState tracks the server state
type ServerState struct {
	Initialized bool
}

// Handler is a function that handles a method request
type Handler func(params json.RawMessage) (any, error)

// NewServer creates a new MCP server
func NewServer(reader io.Reader, sender *MessageSender, logger *slog.Logger) *Server {
	flag.Parse()
	workspace := flag.String("workspace", "", "Path to the current workspace, when using relative file paths this is the root directory.")
	if *workspace == "" {
		*workspace += "."
	}

	server := &Server{
		reader: bufio.NewReader(reader),
		sender: sender,
		log:    logger,
		state: ServerState{
			Initialized: false,
		},
		workspace: *workspace,
		tools:     map[string]McpTool{},
	}

	NewTaskStore(server)

	return server
}

// Run starts the server and begins listening for messages
func (s *Server) Run(ctx context.Context) error {
	for {
		line, err := s.reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				s.log.Info("Connection closed")
				return nil
			}
			s.log.Error(fmt.Sprintf("Error reading message: %v\n", err))
			continue
		}

		if len(line) == 0 {
			continue
		}

		s.log.Info(strings.TrimSpace(line))

		var msg JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			s.log.Error(fmt.Sprintf("Error parsing JSON: %v\n", err))
			continue
		}

		s.handleMessage(ctx, msg)
	}
}

func (s *Server) AddTool(tool McpTool) {
	s.tools[tool.GetName()] = tool
}

// handleMessage processes a single message
func (s *Server) handleMessage(ctx context.Context, msg JSONRPCMessage) {
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg.ID, msg.Params)
	case "initialized":
		s.handleInitialized()
	case "tools/list":
		s.handleListTools(msg.ID, msg.Params)
	case "tools/call":
		go s.handleToolCall(ctx, msg.ID, msg.Params)
	case "tasks/get":
		s.handleTaskGet(ctx, msg.ID, msg.Params)
	case "tasks/result":
		s.handleTaskResult(ctx, msg.ID, msg.Params)
	case "tasks/list":
		s.handleTaskList(ctx, msg.ID)
	case "logging/setLevel":
		s.handleSetLoggingLevel(msg.ID, msg.Params)
	case "server/discover":
		err := s.handleDiscover(ctx, msg.ID, msg.Params)
		if err != nil {
			s.sender.SendError(msg.ID, -32603, err.Error())
		}
	case "ping":
		s.handlePing(msg.ID)
	default:
		if msg.ID != nil {
			s.sender.SendError(msg.ID, -32601, "Method not found")
		}
	}
}

// handleToolCall processes tool call requests
func (s *Server) handleToolCall(ctx context.Context, id any, params json.RawMessage) {
	var toolCall struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	if err := json.Unmarshal(params, &toolCall); err != nil {
		s.sender.SendError(id, -32602, fmt.Sprintf("Invalid parameters: %v", err))
		return
	}

	s.log.Debug("Tool call: %s with arguments %v", toolCall.Name, toolCall.Arguments)

	default_path := s.workspace + "/.tasks.org"

	if tool := s.tools[toolCall.Name]; tool != nil {
		startTime := time.Now()

		resp, error := tool.Execute(ctx, toolCall.Arguments, FuncOptions{DefaultPath: default_path, Logger: s.log})

		for _, r := range resp {
			if task, ok := r.(*Task); ok {
				task.result = append(task.result, resp)

				s.log.Info(fmt.Sprintf("Tool responded with a running task: %s", task.Id))
				s.sender.SendMcpTask(id, task)

				return
			}
		}

		if error != nil {
			s.log.Warn(fmt.Sprintf("Tool error: %v", error))
		}

		s.log.Info(fmt.Sprintf("Tool executed in %v, response: %v", time.Since(startTime), resp))

		if error != nil {
			s.sender.SendError(id, -32000, fmt.Sprintf("Tool error: %v", error))
		} else {
			s.sender.SendMcpContent(id, resp)
		}

		return
	}

	s.log.Warn(fmt.Sprintf("Unknown tool: %s", toolCall.Name))
	s.sender.SendError(id, -32601, fmt.Sprintf("Unknown tool: %s", toolCall.Name))
}

func (s *Server) handleDiscover(_ context.Context, id any, _ json.RawMessage) error {
	encodedTools := map[string]EncodeTool{}
	for name, tool := range s.tools {
		encodedTools[name] = tool.ToEncode()
	}

	result := DiscoverResult{
		Capabilities: ServerCapabilities{
			Tools: ToolCapabilities{ListChanged: false},
			Tasks: map[string]any{
				"list": struct{}{},
				// "cancel":   struct{}{},
				"requests": map[string]any{
					"tools": map[string]any{
						"call": struct{}{},
					},
				},
			},
			Extensions: map[string]any{
				"io.modelcontextprotocol/tasks": struct{}{},
			},
		},
		ServerInfo: ServerInfo{
			Name:    "org-mcp",
			Version: "0.2.0",
		},
		Instructions: MCP_INSTRUCTIONS,
	}

	return s.sender.SendResponse(id, result)
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(id any, _ json.RawMessage) {
	s.state.Initialized = true

	encodedTools := map[string]EncodeTool{}
	for name, tool := range s.tools {
		encodedTools[name] = tool.ToEncode()
	}

	result := InitializeResult{
		ProtocolVersion: "2025-11-25",
		Capabilities: ServerCapabilities{
			Tools: ToolCapabilities{ListChanged: false},
			Tasks: map[string]any{
				"list": struct{}{},
				"requests": map[string]any{
					"tools": map[string]any{
						"call": struct{}{},
					},
				},
			},
			Extensions: map[string]any{
				"io.modelcontextprotocol/tasks": struct{}{},
			},
		},
		ServerInfo: ServerInfo{
			Name:    "org-mcp",
			Version: "0.2.0",
		},
		Instructions: MCP_INSTRUCTIONS,
	}

	if err := s.sender.SendResponse(id, result); err != nil {
		s.log.Error(fmt.Sprintf("Error sending initialize response: %v\n", err))
	}
	s.log.Info("Initialize completed successfully")
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized() {
	s.log.Info("Client confirmed initialization")
}

// handleListTools handles the tools/list request
func (s *Server) handleListTools(id any, _ json.RawMessage) {
	encodedTools := map[string]EncodeTool{}
	for name, tool := range s.tools {
		encodedTools[name] = tool.ToEncode()
	}

	response := map[string]any{
		"tools": slices.Collect(maps.Values(encodedTools)),
	}

	if err := s.sender.SendResponse(id, response); err != nil {
		s.log.Error(fmt.Sprintf("Error sending tools list response: %v\n", err))
	}
}

func (s *Server) handlePing(id any) {
	if err := s.sender.SendResponse(id, map[string]any{}); err != nil {
		s.log.Error(fmt.Sprintf("Error sending ping response: %v\n", err))
	}
}

func (s *Server) handleTaskList(ctx context.Context, id any) error {
	fmt.Fprintf(os.Stderr, "%#v", NewTaskStore(nil).tasks)
	tasks := slices.Collect(maps.Values(NewTaskStore(nil).tasks))
	if tasks == nil {
		tasks = []*Task{}
	}

	str, _ := json.Marshal(tasks)
	s.log.Info(fmt.Sprintf("tasks/list tool call result, %s", string(str)))

	return s.sender.SendResponse(id, map[string][]*Task{
		"tasks": tasks,
	})
}

func (s *Server) handleTaskGet(ctx context.Context, id any, params json.RawMessage) error {
	type GetTaskCall struct {
		TaskId TaskId `json:"taskId"`
	}

	var taskCall GetTaskCall
	if err := json.Unmarshal(params, &taskCall); err != nil {
		return err
	}

	ts := NewTaskStore(nil)
	task := ts.Get(taskCall.TaskId)

	return s.sender.SendMcpContent(id, []any{task})
}

func (s *Server) handleTaskResult(ctx context.Context, id any, params json.RawMessage) error {
	type TaskResultCall struct {
		TaskId TaskId `json:"taskId"`
	}

	var taskCall TaskResultCall
	if err := json.Unmarshal(params, &taskCall); err != nil {
		return err
	}

	ts := NewTaskStore(nil)
	task := ts.Get(taskCall.TaskId)

	for task.Status == WORKING {
		time.Sleep(time.Millisecond * 500)
	}

	switch task.Status {
	case COMPLETED:
		return s.sender.SendMcpContent(id, task.result)
	case CANCELLED:
		return s.sender.SendMcpContent(id, []any{"CANCELLED"})
	case FAILED:
		return s.sender.SendResponse(id, map[string]any{
			"content": []any{},
			"isError": true,
		})
	default:
		panic("unreachable")
	}
}
