package main

import (
	"context"
	"github.com/p3rtang/org-mcp/utils/logging"
	"log/slog"
)

func main() {
	ctx := context.Background()

	// Setup logging to stderr so it doesn't interfere with stdout (which is used for MCP protocol)
	logger := slog.New(&logging.OrgMcpLogHandler{})
	ctx = context.WithValue(ctx, "logger", logger)
	logger.Info("Starting org-mcp server")

	rootCmd.ExecuteContext(ctx)
}
