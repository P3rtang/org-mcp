package table_test

import (
	"context"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

// Org semantics: contiguous | rows are ONE table. A blank line (or any
// non-| content) is what separates two tables under the same header.
// https://github.com/p3rtang/org-mcp/issues/31 header 31b2
func TestParseTableBoundaries(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}
	ctx := context.TODO()

	tests := []struct {
		name        string
		content     string
		expectCount int
		expectUIDs  []string
		expectRows  []int
	}{
		{
			name:        "AdjacentTableLinesMergeIntoOneTable",
			content:     "* Header with existing table\n  :PROPERTIES:\n  :ID: 55555555\n  :END:\n  | Name  | Status | Value |\n  |-------+--------+-------|\n  | Alice | TODO   | 10    |\n  | Bob   | DONE   | 20    |\n  #+NAME: existing_table\n  | A | B |\n  #+NAME: another_table\n",
			expectCount: 1,
			expectUIDs:  []string{"55555555.another_table"}, // trailing #+NAME wins
		},
		{
			name:        "BlankLineSeparatesTwoTables",
			content:     "* Header with existing table\n  :PROPERTIES:\n  :ID: 55555555\n  :END:\n  | Name  | Status | Value |\n  |-------+--------+-------|\n  | Alice | TODO   | 10    |\n  | Bob   | DONE   | 20    |\n  #+NAME: existing_table\n\n  | A | B |\n  #+NAME: another_table\n",
			expectCount: 2,
			expectUIDs:  []string{"55555555.existing_table", "55555555.another_table"},
			expectRows:  []int{2, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "./table_boundary_" + tt.name + ".org"
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			defer os.Remove(path)

			of, err := mcp.LoadOrgFile(ctx, path)
			if err != nil {
				t.Fatal(err)
			}

			tables := orgmcp.GetByType[*table.Table](&of)
			if len(tables) != tt.expectCount {
				t.Errorf("expected %d tables, got %d: %v", tt.expectCount, len(tables), tables)
				return
			}
			for _, expected := range tt.expectUIDs {
				if _, ok := tables[NewUid(expected)]; !ok {
					t.Errorf("expected table %s to exist, got uids %v", expected, tables)
				}
			}
			for i, expected := range tt.expectUIDs {
				tbl, ok := tables[NewUid(expected)]
				if !ok {
					continue
				}
				if tt.expectRows != nil && len(tbl.ContentRows(false)) != tt.expectRows[i] {
					t.Errorf("table %s: expected %d rows, got %d", expected, tt.expectRows[i], len(tbl.ContentRows(false)))
				}
			}
		})
	}
}
