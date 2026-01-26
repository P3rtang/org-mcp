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

// Tool defines an MCP tool with its description and input schema
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`

	Callback func(map[string]any, FuncOptions) ([]any, error) `json:"-"`
}

// ServerCapabilities defines what the MCP server can do
type ServerCapabilities struct {
	Logging   map[string]any  `json:"logging,omitempty"`
	Tools     map[string]Tool `json:"tools,omitempty"`
	Resources map[string]any  `json:"resources,omitempty"`
	Prompts   map[string]any  `json:"prompts,omitempty"`
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
		text, err := json.Marshal(c)

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
