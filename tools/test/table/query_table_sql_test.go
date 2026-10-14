package table_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
	. "github.com/p3rtang/org-mcp/tools/test/types"
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
			Name: "SelectAll",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT * FROM "77777777.sql_test_table"`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"Name,Age,Department,Salary",
			},
		},
		{
			Name: "WhereEquals",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name FROM "77777777.sql_test_table" WHERE Department = 'Eng'`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"Name",
				"Alice",
				"Carol",
			},
		},
		{
			Name: "WhereGreaterThan",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name, Salary FROM "77777777.sql_test_table" WHERE Salary > 80000`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"Name,Salary",
				"Alice,90000",
				"Carol,100000",
				"Eve,85000",
			},
		},
		{
			Name: "OrderBy",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name, Salary FROM "77777777.sql_test_table" ORDER BY Salary DESC`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"Name,Salary",
				"Carol,100000",
				"Alice,90000",
				"Eve,85000",
				"Dave,70000",
				"Bob,60000",
			},
		},
		{
			Name: "Count",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT COUNT(*) FROM "77777777.sql_test_table"`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"COUNT(*)",
				"5",
			},
		},
		{
			Name: "SumAndAverage",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT SUM(Salary), AVG(Salary) FROM "77777777.sql_test_table"`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"SUM(Salary),AVG(Salary)",
			},
		},
		{
			Name: "GroupByDepartment",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Department, COUNT(*) FROM "77777777.sql_test_table" GROUP BY Department`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"Department,COUNT(*)",
				"Eng,2",
				"Sales,2",
				"Marketing,1",
			},
		},
		{
			Name: "Limit",
			Input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSql{
					Uid:    NewUid("77777777.sql_test_table"),
					Method: "sql",
					Query:  `SELECT Name FROM "77777777.sql_test_table" ORDER BY Age LIMIT 2`,
				}),
				Path: sqlTestTablePath,
			},
			Expected: []any{
				"Name",
				"Bob",
				"Dave",
			},
		},
	}

	ctx := context.TODO()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			resp, err := tools.QueryTableTool.Callback(ctx, tt.Input, mcp.FuncOptions{})

			if tt.ExpectEmpty {
				if err == nil {
					t.Errorf("expected error for test case '%s', but got nil", tt.Name)
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

			for _, entry := range tt.Expected {
				expected := entry.(string)
				if !strings.Contains(combinedStr, expected) {
					t.Errorf("expected response to contain %q, got:\n%s", expected, combinedStr)
				}
			}
		})
	}
}
