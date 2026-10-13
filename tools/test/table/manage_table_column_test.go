package table_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
	. "github.com/p3rtang/org-mcp/tools/test/utils"
)

func TestManageTableColumnOperations(t *testing.T) {
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
			name: "AddColumnAtEnd",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 3, Name: "Notes"})},
			},
			expectedResp:  []string{"Notes"},
			expectedTable: "Name,Status,Value,Notes\nAlice,TODO,10,\nBob,DONE,20,\nCarol,TODO,30,",
		},
		{
			name: "AddColumnAtBeginning",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 0, Name: "ID"})},
			},
			expectedResp:  []string{"ID"},
			expectedTable: "ID,Name,Status,Value\n,Alice,TODO,10\n,Bob,DONE,20\n,Carol,TODO,30",
		},
		{
			name: "AddColumnInMiddle",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 2, Name: "Priority"})},
			},
			expectedResp:  []string{"Priority"},
			expectedTable: "Name,Status,Priority,Value\nAlice,TODO,,10\nBob,DONE,,20\nCarol,TODO,,30",
		},
		{
			name: "AddColumnOutOfBounds",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 99, Name: "Extra"})},
			},
			expectedResp:  []string{"Extra"},
			expectedTable: "Name,Status,Value,Extra\nAlice,TODO,10,\nBob,DONE,20,\nCarol,TODO,30,",
		},
		{
			name: "UpdateColumnHeader",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableColumn{Method: "update_col", Uid: NewUid("55555555.table_manage_test"), Col: 1, Name: "State"})},
			},
			expectedResp:  []string{"State"},
			expectedTable: "Name,State,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "UpdateColumnHeaderOutOfBounds",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableColumn{Method: "update_col", Uid: NewUid("55555555.table_manage_test"), Col: 99, Name: "Bad"})},
			},
			expectedResp: []string{fmt.Sprintf(COLUMN_OUT_OF_RANGE, &table.Table{}, NewUid("55555555.table_manage_test"), 3, 99)},
		},
		{
			name: "RemoveColumn",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableColumn{Method: "remove_col", Uid: NewUid("55555555.table_manage_test"), Col: 1})},
			},
			expectedResp:  []string{"Name,Value"},
			expectedTable: "Name,Value\nAlice,10\nBob,20\nCarol,30",
		},
		{
			name: "RemoveColumnOutOfBounds",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableColumn{Method: "remove_col", Uid: NewUid("55555555.table_manage_test"), Col: 99})},
			},
			expectedResp: []string{fmt.Sprintf(COLUMN_OUT_OF_RANGE, &table.Table{}, NewUid("55555555.table_manage_test"), 3, 99)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tools.ManageTableInput{Items: tt.ops}
			runManageTableTest(t, ctx, input, tt.expectError, tt.expectedResp...)
			if !tt.expectError && tt.expectedTable != "" {
				if got := queryTableByUid(t, ctx, "", "55555555.table_manage_test"); got != tt.expectedTable {
					t.Errorf("full table mismatch:\ngot: %s\nexpected: %s", got, tt.expectedTable)
				}
			}
		})
		os.WriteFile("./table_test_manage.org", originalContent, 0644)
	}
}

func TestManageTableMultipleColumnOperations(t *testing.T) {
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
			{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 0, Name: "ID"})},
			{Value: tools.NewManageTableInputUnion(tools.UpdateTableColumn{Method: "update_col", Uid: NewUid("55555555.table_manage_test"), Col: 2, Name: "State"})},
			{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 4, Name: "Notes"})},
		},
	}

	runManageTableTest(t, ctx, input, false, "ID", "State", "Notes")

	if got := queryTableByUid(t, ctx, "", "55555555.table_manage_test"); !ContainsString(got, "ID,Name,State,Value,Notes\n,Alice,TODO,10,\n,Bob,DONE,20,\n,Carol,TODO,30,") {
		t.Errorf("full table mismatch: got %s, expected: %s", got, "ID,Name,State,Value,Notes\n,Alice,TODO,10,\n,Bob,DONE,20,\n,Carol,TODO,30,")
	}
}
