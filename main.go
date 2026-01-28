package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
	"github.com/p3rtang/org-mcp/utils/diff"
	"github.com/p3rtang/org-mcp/utils/slice"
)

// GetDiffOnly renders the OrgFile and returns a diff against the current disk content
// without modifying the file on disk.
func GetDiffOnly(of orgmcp.OrgFile, filePath string) (diffStr string, err error) {
	oldContent, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	builder := strings.Builder{}
	of.Render(&builder, -1)
	newContent := builder.String()

	diffStr = fmt.Sprintf("%v", diff.GetDiff(filePath, string(oldContent), newContent))
	return
}

func main() {
	defaultPath := ".tasks.org"

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

	server.AddTool(tools.HeaderTool, nil)
	server.AddTool(tools.BulletTool, nil)
	server.AddTool(tools.StatusTool, nil)

	server.AddTool(mcp.Tool{
		Name: "vector_search",
		Description: "Perform a vector search on all headers in the org file based on the provided query string. " +
			"Returns the top N most relevant headers.\n" +
			"It is optimal to include as much information as possible in the query, overflowing the text limit is hard. " +
			"So include context like timeframes, people involved, locations, etc. to get the best results.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search query string.",
				},
				"top_n": map[string]any{
					"type":        "number",
					"description": "The number of top relevant headers to return.",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "An optional file path, will default to the configured org file.",
				},
			},
		},
	}, func(args map[string]any, _ mcp.FuncOptions) (resp []any, err error) {
		filePath, ok := args["path"].(string)
		if !ok || strings.TrimSpace(filePath) == "" {
			filePath = defaultPath
		}

		query, ok := args["query"].(string)
		if !ok || strings.TrimSpace(query) == "" {
			return nil, errors.New("invalid or missing query parameter")
		}

		top_n, ok := args["top_n"].(float64)
		if !ok || top_n <= 0 {
			top_n = 3.0
		}

		of, err := mcp.LoadOrgFile(filePath)
		if err != nil {
			return
		}

		headers, err := of.VectorSearch(args["query"].(string), int(top_n))

		resp = []any{map[string]any{
			"results": slice.Map(headers, func(h *orgmcp.Header) map[string]any {
				builder := strings.Builder{}
				h.Render(&builder, 1)

				return map[string]any{
					"uid":        h.Uid(),
					"content":    builder.String(),
					"parent_uid": h.ParentUid(),
				}
			}),
		}}

		return
	})

	if err := server.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
