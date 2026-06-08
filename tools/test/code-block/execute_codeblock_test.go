package test

// import (
// 	"context"
// 	"os"
// 	"strings"
// 	"testing"
//
// 	"github.com/p3rtang/org-mcp/mcp"
// 	. "github.com/p3rtang/org-mcp/orgmcp/types"
// 	"github.com/p3rtang/org-mcp/tools"
// 	. "github.com/p3rtang/org-mcp/tools/test/utils"
// )
//
// const codeBlockExecuteTestPath = "./codeblock_test_execute.org"
//
// type ExecuteCodeBlockTest struct {
// 	name              string
// 	ops               []mcp.GenericOneOf[*tools.ManageCodeBlockInputUnion, tools.CodeBlockApplicableTool]
// 	expectError       bool
// 	expectedResp      []string
// 	expectedInFile    []string
// 	notExpectedInFile []string
// }
//
// func runExecuteCodeBlockTest(t *testing.T, ctx context.Context, input tools.ManageCodeBlockInput, expectError bool, expectedResp ...string) {
// 	t.Helper()
//
// 	resp, err := tools.ManageCodeBlockTool.Callback(ctx, input, mcp.FuncOptions{DefaultPath: codeBlockExecuteTestPath})
// 	if expectError {
// 		if err == nil {
// 			t.Errorf("expected error, got nil")
// 		}
// 		return
// 	}
// 	if err != nil {
// 		t.Fatalf("ManageCodeBlockTool failed: %v", err)
// 	}
//
// 	for _, exp := range expectedResp {
// 		found := false
// 		for _, r := range resp {
// 			if str, ok := r.(string); ok && ContainsString(str, exp) {
// 				found = true
// 				break
// 			} else if errStr, ok := r.(error); ok && ContainsString(errStr.Error(), exp) {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			t.Errorf("expected response to contain %q, got %v", exp, resp)
// 		}
// 	}
// }
//
// func readExecuteCodeBlockFile(t *testing.T) string {
// 	t.Helper()
// 	content, err := os.ReadFile(codeBlockExecuteTestPath)
// 	if err != nil {
// 		t.Fatalf("failed to read %s: %v", codeBlockExecuteTestPath, err)
// 	}
// 	return string(content)
// }
//
// func newExecuteCodeBlockUnionArray(t *tools.ManageCodeBlockInputUnion) []tools.CodeBlockUnion {
// 	return []tools.CodeBlockUnion{mcp.NewGenericOneOf(t)}
// }
//
// func TestExecuteCodeBlock(t *testing.T) {
// 	showDebug := os.Getenv("SHOW_DEBUG")
// 	if showDebug == "" {
// 		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
// 	}
//
// 	originalContent, err := os.ReadFile(codeBlockExecuteTestPath)
// 	if err != nil {
// 		t.Fatalf("failed to read %s: %v", codeBlockExecuteTestPath, err)
// 	}
// 	defer os.WriteFile(codeBlockExecuteTestPath, originalContent, 0644)
//
// 	ctx := context.TODO()
//
// 	tests := []ExecuteCodeBlockTest{
// 		{
// 			name: "ExecutePythonSimple",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.hello_python"),
// 			})),
// 			expectedResp:      []string{`"hello world"`},
// 			expectedInFile:    []string{"#+RESULTS:", "hello world"},
// 			notExpectedInFile: []string{"Traceback"},
// 		},
// 		{
// 			name: "ExecutePythonWithStdout",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.print_args"),
// 			})),
// 			expectedResp:   []string{"a", "b", "c"},
// 			expectedInFile: []string{"#+RESULTS:"},
// 		},
// 		{
// 			name: "ExecutePythonWithStderr",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.warning_code"),
// 			})),
// 			expectedResp:      []string{"WARNING:"},
// 			notExpectedInFile: []string{"Traceback"},
// 		},
// 		{
// 			name: "ExecutePythonError",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.error_code"),
// 			})),
// 			expectedResp:      []string{"Traceback", "ZeroDivisionError"},
// 			notExpectedInFile: []string{"#+RESULTS:"},
// 		},
// 		{
// 			name: "ExecutePythonTimeout",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.infinite_loop"),
// 				Timeout: 5,
// 			})),
// 			expectError:   true,
// 			expectedResp:  []string{"timeout"},
// 		},
// 		{
// 			name: "ExecuteJavaScriptSimple",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.hello_js"),
// 			})),
// 			expectedResp:      []string{"hello javascript"},
// 			expectedInFile:    []string{"#+RESULTS:"},
// 		},
// 		{
// 			name: "ExecuteBashSimple",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.hello_bash"),
// 			})),
// 			expectedResp:      []string{"hello bash"},
// 			expectedInFile:    []string{"#+RESULTS:"},
// 		},
// 		{
// 			name: "ExecuteUnsupportedLanguage",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.rust_code"),
// 			})),
// 			expectError:  true,
// 			expectedResp: []string{"unsupported", "rust"},
// 		},
// 		{
// 			name: "ExecuteNonExistentBlock",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method: "execute",
// 				Uid:    NewUid("33333333.does_not_exist"),
// 			})),
// 			expectError:  true,
// 			expectedResp: []string{tools.ITEM_NOT_FOUND},
// 		},
// 		{
// 			name: "ExecutePythonNoInsertResults",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method:          "execute",
// 				Uid:             NewUid("33333333.hello_python"),
// 				InsertResults:   false,
// 			})),
// 			expectedResp:      []string{`"hello world"`},
// 			notExpectedInFile: []string{"#+RESULTS:"},
// 		},
// 		{
// 			name: "ExecutePythonCustomTimeout",
// 			ops: newExecuteCodeBlockUnionArray(tools.NewManageCodeBlockInputUnion(tools.ExecuteCodeBlock{
// 				Method:  "execute",
// 				Uid:     NewUid("33333333.hello_python"),
// 				Timeout: 30,
// 			})),
// 			expectedResp:      []string{"hello world"},
// 			expectedInFile:    []string{"#+RESULTS:"},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			input := tools.ManageCodeBlockInput{Items: tt.ops, Path: codeBlockExecuteTestPath}
// 			runExecuteCodeBlockTest(t, ctx, input, tt.expectError, tt.expectedResp...)
//
// 			if tt.expectError {
// 				return
// 			}
//
// 			content := readExecuteCodeBlockFile(t)
// 			for _, want := range tt.expectedInFile {
// 				if !strings.Contains(content, want) {
// 					t.Errorf("expected file to contain %q, got:\n%s", want, content)
// 				}
// 			}
// 			for _, notWant := range tt.notExpectedInFile {
// 				if strings.Contains(content, notWant) {
// 					t.Errorf("expected file NOT to contain %q, got:\n%s", notWant, content)
// 				}
// 			}
// 		})
// 		os.WriteFile(codeBlockExecuteTestPath, originalContent, 0644)
// 	}
// }

