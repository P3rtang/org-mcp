package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
	"github.com/p3rtang/org-mcp/utils/diff"
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
	server.AddTool(tools.VectorSearch, nil)

	if err := server.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
