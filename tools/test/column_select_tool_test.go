package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
	. "github.com/p3rtang/org-mcp/tools/test/utils"
)

type ColumnTest struct {
	name        string
	columns     []string
	expected    []any
	expectEmpty bool
}

func buildColumnSelectors(strs []string) []table.ColumnSelector {
	result := make([]table.ColumnSelector, len(strs))
	for i, s := range strs {
		result[i].UnmarshalText([]byte(s))
	}
	return result
}

func TestTableColumnSelection(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	originalContent, err := os.ReadFile("./table_test_manage.org")
	if err != nil {
		t.Fatalf("failed to read table_test_manage.org: %v", err)
	}
	defer os.WriteFile("./table_test_manage.org", originalContent, 0644)

	of, err := mcp.LoadOrgFile(context.TODO(), "./table_column_test.org")
	if err != nil {
		t.Fatal(err)
	}

	mcp.WriteOrgFileToDisk(context.TODO(), of, "./table_column_test.org")

	testMap := []TableTest{
		{
			name: "SelectSingleIndexedColumn",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Columns: buildColumnSelectors([]string{"$0"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"Name\nAlice\nBob\nCharlie"},
		},
		{
			name: "SelectMultipleIndexedColumns",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:").Unwrap(),
					Columns: buildColumnSelectors([]string{"$0", "$1"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"Name,Age\nAlice,30\nBob,25\nCharlie,35"},
		},
		{
			name: "SelectNamedColumn",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:").Unwrap(),
					Columns: buildColumnSelectors([]string{"${City}"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"City\nNYC\nLA\nChicago"},
		},
		{
			name: "SelectColumnsInNonTableOrder",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:").Unwrap(),
					Columns: buildColumnSelectors([]string{"$1", "$0"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"Age,Name\n30,Alice\n25,Bob\n35,Charlie"},
		},
		{
			name: "SelectColumnRange",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:").Unwrap(),
					Columns: buildColumnSelectors([]string{"$0:$2"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"Name,Age,City\nAlice,30,NYC\nBob,25,LA\nCharlie,35,Chicago"},
		},
		{
			name: "SelectMixedIndexedAndNamed",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:").Unwrap(),
					Columns: buildColumnSelectors([]string{"$2", "${Name}"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"City,Name\nNYC,Alice\nLA,Bob\nChicago,Charlie"},
		},
		{
			name: "SelectSingleDataRow",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:1").Unwrap(),
					Columns: buildColumnSelectors([]string{"$0", "$3"}),
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"Name,Score\nAlice,95.5"},
		},
		{
			name: "SelectAllColumnsDefault",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:     NewUid("55555555.table_56218675"),
					Method:  "simple",
					Range:   table.NewTableRange("0:").Unwrap(),
					Columns: []table.ColumnSelector{},
				}),
				Path: "./table_column_test.org",
			},
			expected: []any{"Name,Age,City,Score\nAlice,30,NYC,95.5\nBob,25,LA,82.3\nCharlie,35,Chicago,91.0"},
		},
	}

	ctx := context.TODO()

	for _, tt := range testMap {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tools.QueryTableTool.Callback(ctx, tt.input, mcp.FuncOptions{})

			if tt.expectEmpty {
				if err == nil {
					t.Errorf("expected error for test case '%s', but got nil", tt.name)
				}
				for _, entry := range tt.expected {
					if !ContainsString(err.Error(), entry.(string)) {
						t.Errorf("unexpected error message: got '%s', expected to contain '%s'", err.Error(), entry.(string))
					}
				}
				return
			}

			if err != nil {
				t.Error(err)
			}

			for _, entry := range tt.expected {
				found := false
				for _, r := range resp {
					if str, ok := r.(string); ok {
						found = found || ContainsString(str, entry.(string))
					}
				}

				if !found {
					t.Errorf("unexpected response: got %#v, expected to contain '%s'\n", resp, strings.Trim(entry.(string), "\n"))
				}
			}
		})
	}
}
