package table_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
)

// TDD: range semantics (issue 31c3). Fixture table_manage_test has content rows
// Alice(0) Bob(1) Carol(2) under a Name,Status,Value header; empty_table is
// header-only under its own header.

func TestManageTableRowRangeSemantics(t *testing.T) {
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
		expectedResp  []string // nil covers rows
		expectedTable string
		emptyTable    bool
	}{
		{
			name: "AddRowPastEndAppendsAtEnd",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: 99, Cells: []string{"Dave", "NEXT", "40"}})},
			},
			expectedResp:  []string{"Dave,NEXT,40"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30\nDave,NEXT,40",
		},
		{
			name: "AddRowNegativeErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("55555555.table_manage_test"), Row: -1, Cells: []string{"X", "Y", "Z"}})},
			},
			expectedResp:  []string{"X,Y,Z"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30\nX,Y,Z",
		},
		{
			name: "AddRowToEmptyTableAtRowZero",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableRow{Method: "add", Uid: NewUid("77770000.empty_table"), Row: 0, Cells: []string{"first"}})},
			},
			expectedTable: "Only\nfirst",
			emptyTable:    true,
		},
		{
			name: "UpdateRowPastEndErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{Method: "update", Uid: NewUid("55555555.table_manage_test"), Row: 5, Cells: []string{"X", "Y", "Z"}})},
			},
			expectedResp:  []string{fmt.Sprintf(table.ROW_OUT_OF_RANGE, 3, 5)},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "UpdateRowNegativeErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{Method: "update", Uid: NewUid("55555555.table_manage_test"), Row: -1, Cells: []string{"X", "Y", "Z"}})},
			},
			expectedResp:  []string{"X,Y,Z"},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nX,Y,Z",
		},
		{
			name: "RemoveRowPastEndErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: 9})},
			},
			expectedResp:  []string{fmt.Sprintf(table.ROW_OUT_OF_RANGE, 3, 9)},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "RemoveRowNegativeErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableRow{Method: "remove", Uid: NewUid("55555555.table_manage_test"), Row: -2})},
			},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nCarol,TODO,30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tools.ManageTableInput{Items: tt.ops}
			runManageTableTest(t, ctx, input, false, tt.expectedResp...)
			if tt.expectedTable != "" {
				uid := "55555555.table_manage_test"
				if tt.emptyTable {
					uid = "77770000.empty_table"
				}
				if got := queryTableByUid(t, ctx, "", uid); got != tt.expectedTable {
					t.Errorf("full table mismatch:\ngot:      %s\nexpected: %s", got, tt.expectedTable)
				}
			}
		})
		os.WriteFile("./table_test_manage.org", originalContent, 0644)
	}
}

func TestManageTableColumnRangeSemantics(t *testing.T) {
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
		expectedResp  []string
		expectedTable string
	}{
		{
			name: "AddColOutOfRangeAppendsAtEnd",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: 99, Name: "New"})},
			},
			expectedResp:  []string{"New"},
			expectedTable: "Name,Status,Value,New\nAlice,TODO,10,\nBob,DONE,20,\nCarol,TODO,30,",
		},
		{
			name: "AddColNegativeErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.AddTableColumn{Method: "add_col", Uid: NewUid("55555555.table_manage_test"), Col: -1, Name: "Bad"})},
			},
			expectedResp:  []string{fmt.Sprintf(COLUMN_OUT_OF_RANGE, &table.Table{}, NewUid("55555555.table_manage_test"), 3, -1)},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "UpdateColNegativeErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.UpdateTableColumn{Method: "update_col", Uid: NewUid("55555555.table_manage_test"), Col: -1, Name: "Bad"})},
			},
			expectedResp:  []string{fmt.Sprintf(COLUMN_OUT_OF_RANGE, &table.Table{}, NewUid("55555555.table_manage_test"), 3, -1)},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
		{
			name: "RemoveColNegativeErrors",
			ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
				{Value: tools.NewManageTableInputUnion(tools.RemoveTableColumn{Method: "remove_col", Uid: NewUid("55555555.table_manage_test"), Col: -3})},
			},
			expectedResp:  []string{fmt.Sprintf(COLUMN_OUT_OF_RANGE, &table.Table{}, NewUid("55555555.table_manage_test"), 3, -3)},
			expectedTable: "Name,Status,Value\nAlice,TODO,10\nBob,DONE,20\nCarol,TODO,30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tools.ManageTableInput{Items: tt.ops}
			runManageTableTest(t, ctx, input, false, tt.expectedResp...)
			if tt.expectedTable != "" {
				if got := queryTableByUid(t, ctx, "", "55555555.table_manage_test"); got != tt.expectedTable {
					t.Errorf("full table mismatch:\ngot:      %s\nexpected: %s", got, tt.expectedTable)
				}
			}
		})
		os.WriteFile("./table_test_manage.org", originalContent, 0644)
	}
}
