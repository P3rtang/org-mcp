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
	"slices"
	"strings"
	"time"
)

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

	return &Server{
		reader: bufio.NewReader(reader),
		sender: sender,
		log:    logger,
		state: ServerState{
			Initialized: false,
		},
		workspace: *workspace,
		tools:     map[string]McpTool{},
	}
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
		s.handleToolCall(ctx, msg.ID, msg.Params)
	case "logging/setLevel":
		s.handleSetLoggingLevel(msg.ID, msg.Params)
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

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(id any, params json.RawMessage) {
	s.state.Initialized = true

	var req struct {
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}
	// a client that sends no clientInfo gets only the base instructions
	json.Unmarshal(params, &req)

	encodedTools := map[string]EncodeTool{}
	for name, tool := range s.tools {
		encodedTools[name] = tool.ToEncode()
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: encodedTools,
		},
		ServerInfo: ServerInfo{
			Name:    "org-mcp",
			Version: "0.2.0",
		},
		Instructions: buildInstructions(baseInstructions, req.ClientInfo.Name),
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
