package mcp

import (
	"encoding/json"
	"sync"
)

// JSONRPCMessage represents a JSON-RPC 2.0 message
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   any             `json:"error,omitempty"`
}

type McpTool interface {
	GetName() string
	GetDescription() string
	GetSchema() map[string]any
	ToEncode() EncodeTool
	Execute(map[string]any, FuncOptions) ([]any, error)
}

type EncodeTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Tool defines an MCP tool with its description and input schema
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`

	Callback func(map[string]any, FuncOptions) ([]any, error) `json:"-"`
}

func (t *Tool) GetName() string {
	return t.Name
}

func (t *Tool) GetDescription() string {
	return t.Description
}

func (t *Tool) GetSchema() map[string]any {
	return t.InputSchema
}

func (t *Tool) ToEncode() EncodeTool {
	return EncodeTool{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: t.InputSchema,
	}
}

func (t *Tool) Execute(input map[string]any, options FuncOptions) ([]any, error) {
	return t.Callback(input, options)
}

type GenericTool[Schema any] struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Callback func(Schema, FuncOptions) ([]any, error) `json:"-"`
}

func (g *GenericTool[Schema]) GetName() string {
	return g.Name
}

func (g *GenericTool[Schema]) GetDescription() string {
	return g.Description
}

func (g *GenericTool[Schema]) GetSchema() map[string]any {
	var schema Schema
	return GenerateSchema(schema)
}

func (g *GenericTool[Schema]) ToEncode() EncodeTool {
	return EncodeTool{
		Name:        g.Name,
		Description: g.Description,
		InputSchema: g.GetSchema(),
	}
}

func (g *GenericTool[Schema]) Execute(input map[string]any, options FuncOptions) ([]any, error) {
	var schema Schema

	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &schema)
	if err != nil {
		return nil, err
	}

	return g.Callback(schema, options)
}

// ServerCapabilities defines what the MCP server can do
type ServerCapabilities struct {
	Logging   map[string]any        `json:"logging,omitempty"`
	Tools     map[string]EncodeTool `json:"tools,omitempty"`
	Resources map[string]any        `json:"resources,omitempty"`
	Prompts   map[string]any        `json:"prompts,omitempty"`
}

// ServerInfo contains information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the response to initialize request
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions"`
}

// MessageSender handles thread-safe message sending
type MessageSender struct {
	mu sync.Mutex
	fn func(any) error
}

// NewMessageSender creates a new message sender
func NewMessageSender(sendFn func(any) error) *MessageSender {
	return &MessageSender{
		fn: sendFn,
	}
}

// Send sends a message thread-safely
func (ms *MessageSender) Send(msg any) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.fn(msg)
}

// SendResponse sends a JSON-RPC response
func (ms *MessageSender) SendResponse(id any, result any) error {
	return ms.Send(JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

// SendError sends a JSON-RPC error response
func (ms *MessageSender) SendError(id any, code int, message string) error {
	return ms.Send(JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

// SendNotification sends a JSON-RPC notification (no ID)
func (ms *MessageSender) SendNotification(method string, params any) error {
	return ms.Send(JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  toRawMessage(params),
	})
}

// toRawMessage converts a value to json.RawMessage
func toRawMessage(v any) json.RawMessage {
	if raw, ok := v.(json.RawMessage); ok {
		return raw
	}
	b, _ := json.Marshal(v)
	return b
}

func (ms *MessageSender) SendMcpContent(id any, content []any) error {
	contentList := []any{}

	for _, c := range content {
		var text []byte
		var err error

		if t, ok := c.(string); ok {
			text = []byte(t)
		} else {
			text, err = json.Marshal(c)
		}

		if err != nil {
			return err
		}

		contentList = append(contentList, map[string]string{
			"type": "text",
			"text": string(text),
		})
	}

	return ms.SendResponse(id, map[string]any{
		"content": contentList,
	})
}
