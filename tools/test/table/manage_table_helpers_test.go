package table_test

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

const tableTestFixture = "./table_test_manage.org"

// runManageTableTest is the canonical runner for every manage_table case.
// Error convention (docs/errors.org, 41497269): top-level err is reserved for
// server failures (500-class); tool-level ("400-class") errors surface as
// response items -- strings or error objects -- matched against expectedResp.
//
// With expectError:
//   - top-level err non-nil  -> return (500-class detected)
//   - top-level err nil      -> expectedResp must be present in response items
//
// The path comes from input.Path when set, otherwise tableTestFixture.
func runManageTableTest(t *testing.T, ctx context.Context, input tools.ManageTableInput, expectError bool, expectedResp ...string) {
	t.Helper()

	path := input.Path
	if path == "" {
		path = tableTestFixture
	}

	resp, err := tools.ManageTableTool.Callback(ctx, input, mcp.FuncOptions{DefaultPath: path})
	if expectError && err != nil {
		return // 500-class failure detected at top level
	}
	if err != nil {
		t.Fatalf("ManageTableTool failed: %v", err)
	}

	for _, exp := range expectedResp {
		found := false
		for _, r := range resp {
			if str, ok := r.(string); ok && ContainsString(str, exp) {
				found = true
				break
			} else if e, ok := r.(error); ok && ContainsString(e.Error(), exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected response to contain %s, got %v", exp, resp)
		}
	}

	if expectError {
		return // expected resp item matched, nothing further to verify
	}
}

// queryTableByUid queries the full table (simple range "0:") at uid.
// An empty path falls back to tableTestFixture.
func queryTableByUid(t *testing.T, ctx context.Context, path string, uid string) string {
	t.Helper()

	if path == "" {
		path = tableTestFixture
	}

	resp, err := tools.QueryTableTool.Callback(ctx, tools.QueryTableInput{
		Queries: []mcp.GenericOneOf[*tools.QueryTableUnion, tools.TableApplicableTool]{
			{Value: tools.NewQueryTableUnion(tools.QueryTableSimple{
				Uid:    NewUid(uid),
				Method: "simple",
				Range:  table.NewTableRange("0:").Unwrap(),
			})},
		},
		Path: path,
	}, mcp.FuncOptions{})
	if err != nil {
		t.Fatalf("QueryTableTool failed: %v", err)
	}
	for _, r := range resp {
		if str, ok := r.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	t.Fatal("no string response from QueryTableTool")
	return ""
}

// resetTableFile re-loads and re-writes the fixture, normalizing its layout.
func resetTableFile(t *testing.T, ctx context.Context) {
	t.Helper()

	of, err := mcp.LoadOrgFile(ctx, tableTestFixture)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(ctx, of, tableTestFixture)
}
