package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// JSONRPCMessage represents a JSON-RPC request/response message
type JSONRPCMessage struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id,omitempty"`
	Method  string `json:"method,omitempty"`
	Params  any    `json:"params,omitempty"`
}

func main() {
	// Simulate client communication with the server
	stdin := bufio.NewReader(os.Stdin)
	stdout := json.NewEncoder(os.Stdout)

	// Send an initialize request to the server
	initRequest := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  nil,
	}

	if err := stdout.Encode(initRequest); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send initialize request: %v\n", err)
		return
	}

	// Read the server's response
	response, err := stdin.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read server response: %v\n", err)
		return
	}

	fmt.Printf("Server Response: %s\n", response)
}
