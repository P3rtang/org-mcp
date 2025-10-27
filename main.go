package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"main/mcp"
	"main/orgmcp"
	"main/utils/slice"
	"os"
	"strings"
)

func main() {
	// Define command-line flags
	workspace := flag.String("workspace", "", "Path to the current workspace, when using relative file paths this is the root directory.")
	flag.Parse()

	defaultPath := ""

	if *workspace != "" {
		defaultPath += *workspace + "/"
	}

	defaultPath += ".tasks.org"

	// Setup logging to stderr so it doesn't interfere with stdout (which is used for MCP protocol)
	logger := log.New(os.Stderr, "[org-mcp] ", log.LstdFlags|log.Lshortfile)

	// Create a message sender that encodes and sends JSON to stdout
	// Using a persistent encoder for better performance and proper flushing
	encoder := json.NewEncoder(os.Stdout)
	sender := mcp.NewMessageSender(func(msg any) error {
		return encoder.Encode(msg)
	})

	// Create and run the MCP server
	server := mcp.NewServer(os.Stdin, sender, logger)

	server.AddTool(
		mcp.Tool{
			Name: "get_header_by_status",
			Description: `
		Get all headers that have the input status set.
		If a subheader is specified, only its descendant headers are considered.
		`,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "The status the headers should have.",
						"enum":        []string{"TODO", "NEXT", "PROG", "DONE"},
					},
					"subheader": map[string]any{
						"type":        "number",
						"description": "The subheader to filter by.",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "An optional file path, will default to the configured org file.",
					},
				},
			},
		},
		func(args map[string]any) (response map[string]any, err error) {

			response = map[string]any{}

			filePath, ok := args["path"].(string)

			if !ok || filePath == "" {
				filePath = defaultPath
			}

			file, err := os.Open(filePath)

			if err != nil {
				return
			}

			status := orgmcp.StatusFromString(args["status"].(string))

			if status == orgmcp.None {
				return nil, errors.New("invalid status")
			}

			orgmcp.OrgFileFromReader(file).Then(func(of orgmcp.OrgFile) {
				response["headers"] = slice.Map(of.GetHeaderByStatus(status), func(h *orgmcp.Header) string {
					builder := strings.Builder{}
					h.Render(&builder)

					return builder.String()
				})
			})

			return
		},
	)

	server.AddTool(mcp.Tool{
		Name:        "get_cwd",
		Description: "test for cwd",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(m map[string]any) (resp map[string]any, err error) {
		resp = map[string]any{}
		resp["workspace"] = "test"

		return
	})

	if err := server.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
