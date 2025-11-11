package main

import (
	"github.com/p3rtang/org-mcp/orgmcp"
	"os"
	"strings"
	"testing"
)

func TestFileReproduction(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	content, err := os.ReadFile("./files/schedule.org")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	of := orgmcp.OrgFileFromReader(strings.NewReader(string(content)))

	if of.IsErr() {
		t.Fatalf("failed to parse org file: %v", of.UnwrapErr())
	}

	builder := strings.Builder{}
	of.UnwrapPtr().Render(&builder, -1)

	got := builder.String()
	want := string(content)
	if builder.String() != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
