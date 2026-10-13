package table_test

import (
	"context"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
)

// Regression for issue 31 (31b2): two tables created under one header must
// survive a reload as distinct tables. Header.Render must emit a blank line
// between adjacent table blocks, because the parser merges contiguous | rows.
func TestManageTableCreateTablesStayDistinct(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	originalContent, err := os.ReadFile(createFixture)
	if err != nil {
		t.Fatalf("failed to read %s: %v", createFixture, err)
	}
	defer os.WriteFile(createFixture, originalContent, 0644)

	ctx := context.TODO()

	input := tools.ManageTableInput{
		Items: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.CreateTable{
				Method: "create",
				Uid:    NewUid("55555555"),
				Name:   "another_table",
				Header: []string{"A", "B"},
			})},
		},
		Path: createFixture,
	}

	runManageTableTest(t, ctx, input, false)

	// after reload both siblings exist with their own content
	if got := queryTableByUid(t, ctx, createFixture, "55555555.existing_table"); got != "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20" {
		t.Errorf("existing_table mismatch:\n%s", got)
	}
	if got := queryTableByUid(t, ctx, createFixture, "55555555.another_table"); got != "A,B" {
		t.Errorf("another_table mismatch:\n%s", got)
	}
}
