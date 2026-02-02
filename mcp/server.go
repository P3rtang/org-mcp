package mcp

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"slices"
)

type FuncOptions struct {
	DefaultPath string
}

type ToolFunc func(map[string]any, FuncOptions) ([]any, error)

// Server handles MCP protocol communication over stdio
type Server struct {
	reader    *bufio.Reader
	sender    *MessageSender
	log       *log.Logger
	state     ServerState
	workspace string

	tools  map[string]Tool
	toolCb map[string]ToolFunc
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
		s.log.Printf("Logging level set to %s", logParams.Level)
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
func NewServer(reader io.Reader, sender *MessageSender, logger *log.Logger) *Server {
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
		tools:     map[string]Tool{},
		toolCb:    map[string]ToolFunc{},
	}
}

// Run starts the server and begins listening for messages
func (s *Server) Run() error {
	for {
		line, err := s.reader.ReadString('\n')

		// write the method to a log file
		file, _ := os.OpenFile("org-mcp.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		fmt.Fprintf(file, "%s", line)

		if err != nil {
			if err == io.EOF {
				s.log.Println("Connection closed")
				return nil
			}
			s.log.Printf("Error reading message: %v\n", err)
			continue
		}

		if len(line) == 0 {
			continue
		}

		var msg JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			s.log.Printf("Error parsing JSON: %v\n", err)
			continue
		}

		s.handleMessage(msg)
	}
}

func (s *Server) AddTool(tool Tool) {
	s.tools[tool.Name] = tool
}

// handleMessage processes a single message
func (s *Server) handleMessage(msg JSONRPCMessage) {
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg.ID, msg.Params)
	case "initialized":
		s.handleInitialized()
	case "tools/list":
		s.handleListTools(msg.ID, msg.Params)
	case "tools/call":
		s.handleToolCall(msg.ID, msg.Params)
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
func (s *Server) handleToolCall(id any, params json.RawMessage) {
	var toolCall struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	if err := json.Unmarshal(params, &toolCall); err != nil {
		s.sender.SendError(id, -32602, fmt.Sprintf("Invalid parameters: %v", err))
		return
	}

	default_path := s.workspace + "/.tasks.org"

	if tool := s.toolCb[toolCall.Name]; tool != nil {
		resp, err := tool(toolCall.Arguments, FuncOptions{DefaultPath: default_path})

		if err != nil {
			s.sender.SendError(id, -32000, fmt.Sprintf("Tool error: %v", err))
		} else {
			s.sender.SendMcpContent(id, resp)
		}

		return
	}

	if tool := s.tools[toolCall.Name]; tool.Callback != nil {
		resp, error := tool.Callback(toolCall.Arguments, FuncOptions{DefaultPath: default_path})

		fmt.Fprintln(os.Stderr, resp)

		if error != nil {
			s.sender.SendError(id, -32000, fmt.Sprintf("Tool error: %v", error))
		} else {
			s.sender.SendMcpContent(id, resp)
		}

		return
	}

	s.sender.SendError(id, -32601, fmt.Sprintf("Unknown tool: %s", toolCall.Name))
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(id any, _ json.RawMessage) {
	s.state.Initialized = true

	result := InitializeResult{
		ProtocolVersion: "2025-11-25",
		Capabilities: ServerCapabilities{
			Tools: s.tools,
		},
		ServerInfo: ServerInfo{
			Name:    "org-mcp",
			Version: "0.2.0",
		},
		Instructions: `
		When working with org-mcp, you can use the tool to parse / change org-mode files.
		This should be used to increase your understanding of the project, tasks and notes.

		Tooling to traverse headers, extract metadata, and update statuses are available.
		Using these tools has the benifit of keeping the org-mode file consistent and reducing the amount of data in the context window.
		Since targeted updates and queries are possible, there is no need to pass the entire file contents back and forth.

		The org-mode file is the single source of truth for both the programmer and the LLM.
		Every task, bullet point, or step must be added to the document.
		Implementation begins with extracting the relevant metadata and statuses from the document.

		Direct modification of the org file is discouraged. Use the tooling provided by the org-mcp server to ensure consistency and maintain integrity.
		Direct reading should be avoided as it may lead to blowing up the context window and losing important information.

		An org file serves as a long-term memory and organizational tool for the project. Always refer to it as the main reference point.
		It also functions as long term memory between session, this means that any information not stored in the org file will be lost between sessions.
		Use this together with the programmer to ensure that all important information is captured in the org file.
		`,
	}

	if err := s.sender.SendResponse(id, result); err != nil {
		s.log.Printf("Error sending initialize response: %v\n", err)
	}
	s.log.Println("Initialize completed successfully")
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized() {
	s.log.Println("Client confirmed initialization")
}

// handleListTools handles the tools/list request
func (s *Server) handleListTools(id any, _ json.RawMessage) {
	response := map[string]any{
		"tools": slices.Collect(maps.Values(s.tools)),
	}

	if err := s.sender.SendResponse(id, response); err != nil {
		s.log.Printf("Error sending tools list response: %v\n", err)
	}
}

func (s *Server) handlePing(id any) {
	if err := s.sender.SendResponse(id, map[string]any{}); err != nil {
		s.log.Printf("Error sending ping response: %v\n", err)
	}
}
