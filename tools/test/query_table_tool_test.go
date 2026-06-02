package test

import (
	"context"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
	. "github.com/p3rtang/org-mcp/tools/test/utils"
)

func intoTableOneOfArray[T tools.QueryTableSimple | tools.QueryTableSql](t ...T) []mcp.GenericOneOf[*tools.QueryTableUnion, tools.TableApplicableTool] {
	entries := []mcp.GenericOneOf[*tools.QueryTableUnion, tools.TableApplicableTool]{}

	for _, input := range t {
		entries = append(entries, mcp.GenericOneOf[*tools.QueryTableUnion, tools.TableApplicableTool]{
			Value: tools.NewQueryTableUnion(input),
		})
	}

	return entries
}

func TestTableComplexScenarios(t *testing.T) {
	of, err := mcp.LoadOrgFile(context.TODO(), "./table_test_complex.org")
	if err != nil {
		t.Fatal(err)
	}

	mcp.WriteOrgFileToDisk(context.TODO(), of, "./table_test_complex.org")

	testMap := []TableTest{
		{
			name: "MultipleTablesFirstTable",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("11111111.multi_table_1"),
					Method: "simple",
					Range:  table.NewTableRange("0:").Unwrap(),
				}),
				Path: "./table_test_complex.org",
			},
			expected: []any{"A,B,C\n1,2,3\n4,5,6"},
		},
		{
			name: "MultipleTablesSecondTable",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("11111111.multi_table_2"),
					Method: "simple",
					Range:  table.NewTableRange("0:").Unwrap(),
				}),
				Path: "./table_test_complex.org",
			},
			expected: []any{"X,Y\na,b\nc,d\ne,f"},
		},
		// We expect the tool to omit empty rows completely
		{
			name: "EmptyRowsTable",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("22222222.empty_rows_table"),
					Method: "simple",
					Range:  table.NewTableRange("0:").Unwrap(),
				}),
				Path: "./table_test_complex.org",
			},
			expected: []any{"Col1,Col2\na,b\nc,d\ne,f"},
		},
		{
			name: "SingleRowTable",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("33333333.single_row_table"),
					Method: "simple",
					Range:  table.NewTableRange("0:").Unwrap(),
				}),
				Path: "./table_test_complex.org",
			},
			expected: []any{"OnlyHeader"},
		},
		// {
		// 	name: "NoTableHeader",
		// 	input: tools.QueryTableInput{
		// 		Queries: intoTableOneOfArray(tools.QueryTableSimple{
		// 			Uid:    NewUid("44444444.some_table"),
		// 			Method: "simple",
		// 			Range:  table.NewTableRange("0:1").Unwrap(),
		// 		}),
		// 		Path: "./table_test_complex.org",
		// 	},
		// 	expected:    []any{},
		// 	expectEmpty: true,
		// },
		// {
		// 	name: "RangeBeyondRowCount",
		// 	input: tools.QueryTableInput{
		// 		Queries: intoTableOneOfArray(tools.QueryTableSimple{
		// 			Uid:    NewUid("11111111.multi_table_1"),
		// 			Method: "simple",
		// 			Range:  table.NewTableRange("5:10").Unwrap(),
		// 		}),
		// 		Path: "./table_test_complex.org",
		// 	},
		// 	expected:    []any{"Start of table bounds out of range"},
		// 	expectEmpty: true,
		// },
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

func TestTableSimpleTool(t *testing.T) {
	of, err := mcp.LoadOrgFile(context.TODO(), "./table_test.org")
	if err != nil {
		t.Fatal(err)
	}

	mcp.WriteOrgFileToDisk(context.TODO(), of, "./table_test.org")

	testMap := []TableTest{
		{
			name: "RequestRow",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("42829732.table_349032"),
					Method: "simple",
					Range:  table.NewTableRange("0:1").Unwrap(),
				}),
				Path: "./table_test.org",
			},
			expected: []any{"Col 1,Col 2,Col 3\nValue 1,Value 2,Value 3"},
		},
		{
			name: "RequestRowRangeAll",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("42829732.table_349032"),
					Method: "simple",
					Range:  table.NewTableRange("0:").Unwrap(),
				}),
				Path: "./table_test.org",
			},
			expected: []any{"Col 1,Col 2,Col 3\nValue 1,Value 2,Value 3"},
		},
		{
			name: "RequestRowRangeEmpty",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("42829732.table_349032"),
					Method: "simple",
					Range:  table.NewTableRange("0:0").Unwrap(),
				}),
				Path: "./table_test.org",
			},
			expected: []any{"Col 1,Col 2,Col 3"},
		},
		{
			name: "RequestRowRangeNegative",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("42829732.table_349032"),
					Method: "simple",
					Range:  table.NewTableRange("-1:").Unwrap(),
				}),
				Path: "./table_test.org",
			},
			expected: []any{"Col 1,Col 2,Col 3\nValue 1,Value 2,Value 3"},
		},
		{
			name: "RequestRowRangeOutOfBounds",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("42829732.table_349032"),
					Method: "simple",
					Range:  table.NewTableRange("0:100").Unwrap(),
				}),
				Path: "./table_test.org",
			},
			expected: []any{"Col 1,Col 2,Col 3\nValue 1,Value 2,Value 3"},
		},
		{
			name: "RequestMultipleQueries",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(
					tools.QueryTableSimple{
						Uid:    NewUid("42829732.table_349032"),
						Method: "simple",
						Range:  table.NewTableRange("0:1").Unwrap(),
					},
					tools.QueryTableSimple{
						Uid:    NewUid("42829732.table_349032"),
						Method: "simple",
						Range:  table.NewTableRange("1:2").Unwrap(),
					},
				),
				Path: "./table_test.org",
			},
			expected: []any{"Col 1,Col 2,Col 3\nValue 1,Value 2,Value 3", "Value 1,Value 2,Value 3"},
		},
		// {
		// 	name: "RequestInvalidUid",
		// 	input: tools.QueryTableInput{
		// 		Queries: intoTableOneOfArray(tools.QueryTableSimple{
		// 			Uid:    NewUid("nonexistent.table"),
		// 			Method: "simple",
		// 			Range:  table.NewTableRange("0:1").Unwrap(),
		// 		}),
		// 		Path: "./table_test.org",
		// 	},
		// 	expected:    []any{},
		// 	expectEmpty: true,
		// },
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
