package table_test

import (
	"context"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
)

const createFixture = "./table_test_create.org"

// TDD (docs/tables/issue-31.org, 31b2): CreateTable does not have an
// implementation yet -- CreateTable.Apply panics by design. These tests
// define the spec; they go green when the real implementation lands.
type CreateTableTest struct {
	name         string
	ops          []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]
	expectError  bool
	expectedResp []string
	expectedCSV  string
	expectedUID  string // table uid to query after the create
}

var createTests = []CreateTableTest{
	{
		name: "UnderHeaderWithoutTable",
		ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.CreateTable{
				Method: "create",
				Uid:    NewUid("88880000"),
				Name:   "created_table",
				Header: []string{"Name", "Role"},
			})},
		},
		expectedCSV: "Name,Role",
		expectedUID: "88880000.created_table",
	},
	{
		name: "SecondTableUnderSameHeaderAllowed",
		// org-mode supports multiple tables per header; uids disambiguate via #+NAME
		ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.CreateTable{
				Method: "create",
				Uid:    NewUid("55555555"),
				Name:   "another_table",
				Header: []string{"A", "B"},
			})},
		},
		expectedCSV: "A,B",
		expectedUID: "55555555.another_table",
	},
	{
		name: "DuplicateNameUnderSameHeaderErrors",
		ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.CreateTable{
				Method: "create",
				Uid:    NewUid("55555555"),
				Name:   "existing_table",
				Header: []string{"A", "B"},
			})},
		},
		expectError: true,
		expectedResp: []string{
			"already contains a child with UID `55555555.existing_table`",
		},
	},
	{
		name: "UidPointsAtExistingTableNotHeader",
		ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.CreateTable{
				Method: "create",
				Uid:    NewUid("55555555.existing_table"),
				Name:   "nope",
				Header: []string{"A", "B"},
			})},
		},
		expectError: true,
		expectedResp: []string{
			"cannot have children",
		},
	},
	{
		name: "WithHeaderRowCells",
		ops: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
			{Value: tools.NewManageTableInputUnion(tools.CreateTable{
				Method: "create",
				Uid:    NewUid("88880000"),
				Name:   "styled_table",
				Header: []string{"Task", "Status", "Due"},
			})},
		},
		expectedCSV: "Task,Status,Due",
		expectedUID: "88880000.styled_table",
	},
}

func TestManageTableCreate(t *testing.T) {
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

	for _, tt := range createTests {
		t.Run(tt.name, func(t *testing.T) {
			os.WriteFile(createFixture, originalContent, 0644) // reset per subtest

			input := tools.ManageTableInput{
				Items: tt.ops,
				Path:  createFixture,
			}

			runManageTableTest(t, ctx, input, tt.expectError, tt.expectedResp...)

			if !tt.expectError && tt.expectedCSV != "" {
				if got := queryTableByUid(t, ctx, createFixture, tt.expectedUID); got != tt.expectedCSV {
					t.Errorf("full table mismatch:\ngot: %s\nexpected: %s", got, tt.expectedCSV)
				}
			}
		})
	}
}
