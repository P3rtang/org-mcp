package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
	. "github.com/p3rtang/org-mcp/tools/test/utils"
)

func runManageTableTest(t *testing.T, ctx context.Context, input tools.ManageTableInput, expectError bool, expectedResp ...string) {
	t.Helper()

	resp, err := tools.ManageTableTool.Callback(ctx, input, mcp.FuncOptions{DefaultPath: "./table_test_manage.org"})
	if expectError {
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		return
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
			} else if str, ok := r.(error); ok && ContainsString(str.Error(), exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected response to contain %s, got %v", exp, resp)
		}
	}
}

func queryFullTable(t *testing.T, ctx context.Context) string {
	t.Helper()
	resp, err := tools.QueryTableTool.Callback(ctx, tools.QueryTableInput{
		Queries: []mcp.GenericOneOf[*tools.QueryTableUnion, tools.TableApplicableTool]{
			{Value: tools.NewQueryTableUnion(tools.QueryTableSimple{
				Uid:    NewUid("55555555.table_manage_test"),
				Method: "simple",
				Range:  table.NewTableRange("0:").Unwrap(),
			})},
		},
		Path: "./table_test_manage.org",
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

func resetTableFile(t *testing.T, ctx context.Context) {
	t.Helper()
	of, err := mcp.LoadOrgFile(ctx, "./table_test_manage.org")
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(ctx, of, "./table_test_manage.org")
}

func TestManageTableOperations(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	originalContent, err := os.ReadFile("./table_test_manage.org")
	if err != nil {
		t.Fatalf("failed to read table_test_manage.org: %v", err)
	}
	defer os.WriteFile("./table_test_manage.org", originalContent, 0644)

	ctx := t.Context()

	tests := []struct {
		name          string
		ops           []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]
		expectError   bool
		expectedResp  []string
		expectedTable string
	}{
		{
			name: "AddRowAtEnd",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 3, Cells: []string{"Dave", "NEXT", "40"}})},
			},
			expectedResp:  []string{"Carol,TODO,30\nDave,NEXT,40"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30\nDave,NEXT,40",
		},
		{
			name: "AddRowAtBeginning",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 0, Cells: []string{"Zara", "DONE", "99"}})},
			},
			expectedResp:  []string{"Zara,DONE,99\nAlice,TODO,10"},
			expectedTable: "Name,Status,Value\nZara,DONE,99\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "AddRowInMiddle",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 1, Cells: []string{"Eve", "PROG", "50"}})},
			},
			expectedResp:  []string{"Alice,TODO,10\nEve,PROG,50\nBob,DONE,20"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nEve,PROG,50\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "AddRowWithTooFewCells",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 0, Cells: []string{"Frank"}})},
			},
			expectedResp:  []string{"Frank,,\nAlice,TODO,10"},
			expectedTable: "Name,Status,Value\nFrank,,\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "AddRowWithTooManyCells",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 0, Cells: []string{"Grace", "DONE", "60", "extra"}})},
			},
			expectedResp:  []string{"Grace,DONE,60\nAlice,TODO,10"},
			expectedTable: "Name,Status,Value\nGrace,DONE,60\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "UpdateFirstRow",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{Method: "update", Uid: NewUid("55555555.table_manage_test"), Row: 0, Cells: []string{"Alice", "DONE", "100"}})},
			},
			expectedResp:  []string{"Alice,DONE,100\nBob,DONE,20"},
			expectedTable: "Name,Status,Value\nAlice,DONE,100\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "UpdateLastRow",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{Method: "update", Uid: NewUid("55555555.table_manage_test"), Row: 2, Cells: []string{"Carol", "DONE", "300"}})},
			},
			expectedResp:  []string{"Bob,DONE,20\nCarol,DONE,300"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,DONE,300",
		},
		{
			name: "UpdateRowOutOfBounds",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{Method: "update", Uid: NewUid("55555555.table_manage_test"), Row: 99, Cells: []string{"X", "Y", "Z"}})},
			},
			expectedResp: []string{fmt.Sprintf(table.ROW_OUT_OF_RANGE, 3, 99)},
		},
		{
			name: "RemoveFirstRow",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: 0})},
			},
			expectedResp:  []string{"Bob,DONE,20"},
			expectedTable: "Name,Status,Value\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "RemoveLastRow",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: 2})},
			},
			expectedResp:  []string{"Bob,DONE,20"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20",
		},
		{
			name: "RemoveMiddleRow",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: 1})},
			},
			expectedResp:  []string{"Alice,TODO,10\nCarol,TODO,30"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nCarol,TODO,30",
		},
		{
			name: "RemoveRowOutOfBounds",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: 99})},
			},
			expectedResp: []string{fmt.Sprintf(table.ROW_OUT_OF_RANGE, 3, 99)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tools.ManageTableInput{Items: tt.ops}
			runManageTableTest(t, ctx, input, tt.expectError, tt.expectedResp...)
			if !tt.expectError && tt.expectedTable != "" {
				if got := queryFullTable(t, ctx); got != tt.expectedTable {
					t.Errorf("full table mismatch:\ngot: %s\nexpected: %s", got, tt.expectedTable)
				}
			}
		})
		os.WriteFile("./table_test_manage.org", originalContent, 0644)
	}
}

func TestManageTableMultipleOperations(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	originalContent, err := os.ReadFile("./table_test_manage.org")
	if err != nil {
		t.Fatalf("failed to read table_test_manage.org: %v", err)
	}
	defer os.WriteFile("./table_test_manage.org", originalContent, 0644)

	ctx := context.TODO()
	resetTableFile(t, ctx)

	input := tools.ManageTableInput{
		Items: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: 1})},
			{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 1, Cells: []string{"Eve", "PROG", "50"}})},
			{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{Method: "update", Uid: NewUid("55555555.table_manage_test"), Row: 0, Cells: []string{"Alice", "DONE", "10"}})},
		},
	}

	runManageTableTest(t, ctx, input, false, "Alice,TODO,10\nEve,PROG,50\nCarol,TODO,30")

	if got := queryFullTable(t, ctx); got != "Name,Status,Value\nAlice,DONE,10\nEve,PROG,50\nCarol,TODO,30" {
		t.Errorf("full table mismatch: got %s", got)
	}
}
