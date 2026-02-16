package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func TODO[T any]() T {
	_, file, line, _ := runtime.Caller(1)
	println(fmt.Sprintf("%s:%d: TODO: not implemented yet", file, line))
	panic("")
}

type OrgMcpLogHandler struct {
	path  string
	level slog.Level
}

func (o *OrgMcpLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if o.path == "" {
		o.path = "org-mcp.log"
	}

	file, err := os.OpenFile(o.path, os.O_RDWR|os.O_APPEND, os.ModeAppend)
	if err != nil {
		return err
	}

	logTime := time.Now().Format(time.StampMilli)

	_, f, line, ok := runtime.Caller(3)
	var location string
	if ok {
		var relativePath string
		cwd, err := os.Getwd()
		relativePath, err = filepath.Rel(cwd, f)

		if err != nil {
			relativePath = filepath.Base(f)
		}
		location = fmt.Sprintf("%s:%d", relativePath, line)
	}

	fmt.Fprintf(file, "%s [%s] %s: %s\n", logTime, record.Level.String(), location, strings.ReplaceAll(record.Message, "\n", "\\n"))

	return nil
}

func (o *OrgMcpLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= o.level
}

func (o *OrgMcpLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return o
}

func (o *OrgMcpLogHandler) WithGroup(group string) slog.Handler {
	return o
}
