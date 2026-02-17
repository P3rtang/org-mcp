package main

import (
	"context"
	"github.com/p3rtang/org-mcp/utils/logging"
	"log/slog"
)

func main() {
	logger := slog.New(&logging.OrgMcpLogHandler{})
	logger.Info("Starting org-mcp server")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "logger", logger)

	rootCmd.ExecuteContext(ctx)
}
