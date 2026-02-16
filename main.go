package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
	"github.com/p3rtang/org-mcp/utils/diff"
	"github.com/p3rtang/org-mcp/utils/logging"
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
	ctx := context.Background()

	// Setup logging to stderr so it doesn't interfere with stdout (which is used for MCP protocol)
	logger := slog.New(&logging.OrgMcpLogHandler{})
	ctx = context.WithValue(ctx, "logger", logger)
	logger.Info("Starting org-mcp server")

	// Check if the export flag is set and print the export format if so
	if len(os.Args) > 1 && os.Args[1] == "export" {
		f := ".tasks.org"
		o := "out.md"

		if len(os.Args) > 2 {
			f = os.Args[2]
		}

		if len(os.Args) > 3 {
			o = os.Args[3]
		}

		file, err := os.Open(f)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to open %s: %v", f, err))
			os.Exit(1)
		}

		orgFile, err := orgmcp.OrgFileFromReader(ctx, file).Split()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse %s: %v", f, err))
			os.Exit(1)
		}

		file.Close()

		builder := strings.Builder{}
		orgFile.RenderMarkdown(&builder, -1)

		out, err := os.Create(o)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create %s: %v", o, err))
			os.Exit(1)
		}

		_, err = out.WriteString(builder.String())
		return
	}

	// Create a message sender that encodes and sends JSON to stdout
	// Using a persistent encoder for better performance and proper flushing
	encoder := json.NewEncoder(os.Stdout)
	sender := mcp.NewMessageSender(func(msg any) error {
		return encoder.Encode(msg)
	})

	// Create and run the MCP server
	server := mcp.NewServer(os.Stdin, sender, logger)

	server.AddTool(&tools.ViewTool)
	server.AddTool(&tools.HeaderTool)
	server.AddTool(&tools.BulletTool)
	server.AddTool(&tools.StatusTool)
	server.AddTool(&tools.VectorSearch)
	server.AddTool(&tools.TextTool)

	if err := server.Run(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server error: %v", err))
		os.Exit(1)
	}
}
