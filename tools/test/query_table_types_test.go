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
)

const typesTestTablePath = "./table_test_types.org"
const typesTestTableUid = "88888888.table_typed"

func runTableTypesQuery(t *testing.T, ctx context.Context, hideTypes bool) string {
	t.Helper()

	input := tools.QueryTableInput{
		Queries: intoTableOneOfArray(tools.QueryTableSimple{
			Uid:    NewUid(typesTestTableUid),
			Method: "simple",
			Range:  table.NewTableRange("0:").Unwrap(),
		}),
		HideTypes: hideTypes,
		Path:      typesTestTablePath,
	}

	resp, err := tools.QueryTableTool.Callback(ctx, input, mcp.FuncOptions{DefaultPath: typesTestTablePath})
	if err != nil {
		t.Fatalf("QueryTableTool failed: %v", err)
	}

	combined := strings.Builder{}
	for _, r := range resp {
		if str, ok := r.(string); ok {
			combined.WriteString(str)
		}

		if err, ok := r.(error); ok {
			combined.WriteString(err.Error())
		}
	}
	return combined.String()
}

func TestTableTypesShownByDefault(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of, err := mcp.LoadOrgFile(context.TODO(), typesTestTablePath)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, typesTestTablePath)

	ctx := context.TODO()
	output := runTableTypesQuery(t, ctx, false)

	if !strings.HasPrefix(output, "#+TYPE:") {
		t.Errorf("expected output to start with #+TYPE: line, got:\n%s", output)
	}

	wantTypeLine := "#+TYPE: Name=text,Age=int,Department=text,Salary=int"
	if !strings.Contains(output, wantTypeLine) {
		t.Errorf("expected output to contain %q, got:\n%s", wantTypeLine, output)
	}

	wantHeader := "Name,Age,Department,Salary"
	if !strings.Contains(output, wantHeader) {
		t.Errorf("expected output to contain CSV header %q, got:\n%s", wantHeader, output)
	}

	wantRow := "Alice,30,Eng,90000"
	if !strings.Contains(output, wantRow) {
		t.Errorf("expected output to contain data row %q, got:\n%s", wantRow, output)
	}
}

func TestTableTypesHiddenWithFlag(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of, err := mcp.LoadOrgFile(context.TODO(), typesTestTablePath)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, typesTestTablePath)

	ctx := context.TODO()
	output := runTableTypesQuery(t, ctx, true)

	if strings.Contains(output, "#+TYPE:") {
		t.Errorf("expected output NOT to contain #+TYPE: when hide_types=true, got:\n%s", output)
	}

	if !strings.HasPrefix(output, "Name,") {
		t.Errorf("expected output to start with CSV header when hide_types=true, got:\n%s", output)
	}

	wantHeader := "Name,Age,Department,Salary"
	if !strings.Contains(output, wantHeader) {
		t.Errorf("expected output to contain CSV header %q, got:\n%s", wantHeader, output)
	}
}

func TestTableNoTypeMetadata(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of, err := mcp.LoadOrgFile(context.TODO(), "./table_test.org")
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, "./table_test.org")

	input := tools.QueryTableInput{
		Queries: intoTableOneOfArray(tools.QueryTableSimple{
			Uid:    NewUid("42829732.table_349032"),
			Method: "simple",
			Range:  table.NewTableRange("0:").Unwrap(),
		}),
		Path: "./table_test.org",
	}

	resp, err := tools.QueryTableTool.Callback(context.TODO(), input, mcp.FuncOptions{DefaultPath: "./table_test.org"})
	if err != nil {
		t.Fatalf("QueryTableTool failed: %v", err)
	}

	combined := strings.Builder{}
	for _, r := range resp {
		if str, ok := r.(string); ok {
			combined.WriteString(str)
		}
	}
	output := combined.String()

	if strings.Contains(output, "#+TYPE:") {
		t.Errorf("expected output NOT to contain #+TYPE: for table without type metadata, got:\n%s", output)
	}
}
