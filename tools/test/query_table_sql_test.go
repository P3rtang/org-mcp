package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
)

const sqlTestTablePath = "./table_test_sql.org"

func TestTableSqlQueries(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of, err := mcp.LoadOrgFile(context.TODO(), sqlTestTablePath)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, sqlTestTablePath)

	tests := []TableTest{
		{
			name: "SelectAll",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT * FROM "77777777.sql_test_table"`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"Name,Age,Department,Salary",
			},
		},
		{
			name: "WhereEquals",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name FROM "77777777.sql_test_table" WHERE Department = 'Eng'`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"Name",
				"Alice",
				"Carol",
			},
		},
		{
			name: "WhereGreaterThan",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name, Salary FROM "77777777.sql_test_table" WHERE Salary > 80000`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"Name,Salary",
				"Alice,90000",
				"Carol,100000",
				"Eve,85000",
			},
		},
		{
			name: "OrderBy",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name, Salary FROM "77777777.sql_test_table" ORDER BY Salary DESC`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"Name,Salary",
				"Carol,100000",
				"Alice,90000",
				"Eve,85000",
				"Dave,70000",
				"Bob,60000",
			},
		},
		{
			name: "Count",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT COUNT(*) FROM "77777777.sql_test_table"`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"COUNT(*)",
				"5",
			},
		},
		{
			name: "SumAndAverage",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT SUM(Salary), AVG(Salary) FROM "77777777.sql_test_table"`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"SUM(Salary),AVG(Salary)",
			},
		},
		{
			name: "GroupByDepartment",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Department, COUNT(*) FROM "77777777.sql_test_table" GROUP BY Department`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"Department,COUNT(*)",
				"Eng,2",
				"Sales,2",
				"Marketing,1",
			},
		},
		{
			name: "Limit",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name FROM "77777777.sql_test_table" ORDER BY Age LIMIT 2`,
				}),
				Path: sqlTestTablePath,
			},
			expected: []any{
				"Name",
				"Bob",
				"Dave",
			},
		},
	}

	ctx := context.TODO()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tools.QueryTableTool.Callback(ctx, tt.input, mcp.FuncOptions{})

			if tt.expectEmpty {
				if err == nil {
					t.Errorf("expected error for test case '%s', but got nil", tt.name)
				}
				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			combined := strings.Builder{}
			for _, r := range resp {
				if str, ok := r.(string); ok {
					combined.WriteString(str)
				}
			}
			combinedStr := combined.String()

			for _, entry := range tt.expected {
				expected := entry.(string)
				if !strings.Contains(combinedStr, expected) {
					t.Errorf("expected response to contain %q, got:\n%s", expected, combinedStr)
				}
			}
		})
	}
}
