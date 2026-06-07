package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/tools"
	. "github.com/p3rtang/org-mcp/tools/test/utils"
)

const codeBlockManageTestPath = "./codeblock_test_manage.org"

type ManageCodeBlockTest struct {
	name              string
	ops               []mcp.GenericOneOf[*tools.ManageCodeBlockInputUnion, tools.CodeBlockApplicableTool]
	expectError       bool
	expectedResp      []string
	expectedInFile    []string
	notExpectedInFile []string
}

func runManageCodeBlockTest(t *testing.T, ctx context.Context, input tools.ManageCodeBlockInput, expectError bool, expectedResp ...string) {
	t.Helper()

	resp, err := tools.ManageCodeBlockTool.Callback(ctx, input, mcp.FuncOptions{DefaultPath: codeBlockManageTestPath})
	if expectError {
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf("ManageCodeBlockTool failed: %v", err)
	}

	for _, exp := range expectedResp {
		found := false
		for _, r := range resp {
			if str, ok := r.(string); ok && ContainsString(str, exp) {
				found = true
				break
			} else if errStr, ok := r.(error); ok && ContainsString(errStr.Error(), exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected response to contain %q, got %v", exp, resp)
		}
	}
}

func readCodeBlockFile(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(codeBlockManageTestPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", codeBlockManageTestPath, err)
	}
	return string(content)
}

func newCodeBlockUnionArray(t *tools.ManageCodeBlockInputUnion) []tools.CodeBlockUnion {
	return []tools.CodeBlockUnion{mcp.NewGenericOneOf(t)}
}

func TestManageCodeBlock(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	originalContent, err := os.ReadFile(codeBlockManageTestPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", codeBlockManageTestPath, err)
	}
	defer os.WriteFile(codeBlockManageTestPath, originalContent, 0644)

	ctx := context.TODO()

	tests := []ManageCodeBlockTest{
		{
			name: "AddNamedCodeBlock",
			ops: newCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.AddCodeBlock{
				Method:   "add",
				Parent:   NewUid("11111111"),
				Language: "go",
				Name:     "new_block",
				Content:  "package main\n\nfunc main() {}\n",
			})),
			expectedResp:   []string{"new_block"},
			expectedInFile: []string{"#+BEGIN_SRC go", "#+NAME: new_block"},
		},
		{
			name: "AddUnnamedCodeBlock",
			ops: []mcp.GenericOneOf[*tools.ManageCodeBlockInputUnion, tools.CodeBlockApplicableTool]{
				{Value: tools.NewManageCodeBlockInputUnion(tools.AddCodeBlock{
					Method:   "add",
					Parent:   NewUid("22222222"),
					Language: "rust",
					Content:  "fn main() {}\n",
				})},
			},
			expectedInFile: []string{"#+BEGIN_SRC rust", "fn main()"},
		},
		{
			name: "UpdateCodeBlockContent",
			ops: newCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.UpdateCodeBlock{
				Method:  "update",
				Uid:     NewUid("11111111.c1"),
				Content: "print(\"updated\")\n",
			})),
			expectedResp:   []string{"updated"},
			expectedInFile: []string{"updated"},
		},
		{
			name: "RemoveCodeBlock",
			ops: newCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.RemoveCodeBlock{
				Method: "remove",
				Uid:    NewUid("11111111.c1"),
			})),
			notExpectedInFile: []string{"#+BEGIN_SRC python"},
		},
		{
			name: "UpdateNonExistentCodeBlock",
			ops: newCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.UpdateCodeBlock{
				Method:  "update",
				Uid:     NewUid("11111111.c99"),
				Content: "x",
			})),
			expectedResp: []string{fmt.Sprintf(tools.ITEM_NOT_FOUND, "11111111.c99")},
		},
		{
			name: "RemoveNonExistentCodeBlock",
			ops: newCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.RemoveCodeBlock{
				Method: "remove",
				Uid:    NewUid("11111111.c99"),
			})),
			expectedResp: []string{fmt.Sprintf(tools.ITEM_NOT_FOUND, "11111111.c99")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tools.ManageCodeBlockInput{Items: tt.ops, Path: codeBlockManageTestPath}
			runManageCodeBlockTest(t, ctx, input, tt.expectError, tt.expectedResp...)

			if tt.expectError {
				return
			}

			content := readCodeBlockFile(t)
			for _, want := range tt.expectedInFile {
				if !strings.Contains(content, want) {
					t.Errorf("expected file to contain %q, got:\n%s", want, content)
				}
			}
			for _, notWant := range tt.notExpectedInFile {
				if strings.Contains(content, notWant) {
					t.Errorf("expected file NOT to contain %q, got:\n%s", notWant, content)
				}
			}
		})
		os.WriteFile(codeBlockManageTestPath, originalContent, 0644)
	}
}
