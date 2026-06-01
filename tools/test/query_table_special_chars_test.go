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

const specialCharsTablePath = "./table_test_special_chars.org"

func TestTableSpecialCharsQuery(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of, err := mcp.LoadOrgFile(context.TODO(), specialCharsTablePath)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, specialCharsTablePath)

	tests := []TableTest{
		{
			name: "EscapedPipeBetweenLetters",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("66666666.special_chars_table"),
					Method: "simple",
					Range:  table.NewTableRange("0:1").Unwrap(),
				}),
				Path: specialCharsTablePath,
			},
			expected: []any{"hello,a|b,\"x,y\",\"q\""},
		},
		{
			name: "EscapedPipeInsideWord",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("66666666.special_chars_table"),
					Method: "simple",
					Range:  table.NewTableRange("1:2").Unwrap(),
				}),
				Path: specialCharsTablePath,
			},
			expected: []any{"middle,mid|end,\"a,,b\",'apos'"},
		},
		{
			name: "TrailingEscapedPipe",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("66666666.special_chars_table"),
					Method: "simple",
					Range:  table.NewTableRange("2:3").Unwrap(),
				}),
				Path: specialCharsTablePath,
			},
			expected: []any{`end,trailing|,",,","x""y"`},
		},
		{
			name: "FullTableWithSpecialChars",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("66666666.special_chars_table"),
					Method: "simple",
					Range:  table.NewTableRange("0:").Unwrap(),
				}),
				Path: specialCharsTablePath,
			},
			expected: []any{
				`hello,a|b,"x,y","q"`,
				`middle,mid|end,"a,,b",'apos'`,
				`end,trailing|,",,","x""y"`,
			},
		},
		{
			name: "HeaderRowUnaffected",
			input: tools.QueryTableInput{
				Queries: intoTableOneOfArray(tools.QueryTableSimple{
					Uid:    NewUid("66666666.special_chars_table"),
					Method: "simple",
					Range:  table.NewTableRange("0:0").Unwrap(),
				}),
				Path: specialCharsTablePath,
			},
			expected: []any{"Plain,EscapedPipe,Comma,Quote"},
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
				for _, entry := range tt.expected {
					if !ContainsString(err.Error(), entry.(string)) {
						t.Errorf("unexpected error message: got '%s', expected to contain '%s'", err.Error(), entry.(string))
					}
				}
				return
			}

			if err != nil {
				t.Error(err)
			}

			for _, entry := range tt.expected {
				found := false
				for _, r := range resp {
					if str, ok := r.(string); ok {
						found = found || ContainsString(str, entry.(string))
					}
				}

				if !found {
					t.Errorf("unexpected response: got %#v, expected to contain '%s'\n", resp, entry.(string))
				}
			}
		})
	}
}

func querySpecialCharsCell(t *testing.T, ctx context.Context, row int) string {
	t.Helper()

	resp, err := tools.QueryTableTool.Callback(ctx, tools.QueryTableInput{
		Queries: intoTableOneOfArray(tools.QueryTableSimple{
			Uid:    NewUid("66666666.special_chars_table"),
			Method: "simple",
			Range:  table.NewTableRangeFull(row, row+1),
		}),
		Path: specialCharsTablePath,
	}, mcp.FuncOptions{DefaultPath: specialCharsTablePath})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	for _, r := range resp {
		if str, ok := r.(string); ok {
			return strings.TrimSpace(str)
		}
	}

	t.Fatal("no string response from query")
	return ""
}

func reloadOrgFile(t *testing.T, ctx context.Context) {
	t.Helper()

	of, err := mcp.LoadOrgFile(ctx, specialCharsTablePath)
	if err != nil {
		t.Fatalf("failed to reload org file: %v", err)
	}
	if _, err := mcp.WriteOrgFileToDisk(ctx, of, specialCharsTablePath); err != nil {
		t.Fatalf("failed to rewrite org file: %v", err)
	}
}

func TestTablePipeEscapingOnUpdate(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	originalContent, err := os.ReadFile(specialCharsTablePath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", specialCharsTablePath, err)
	}
	defer os.WriteFile(specialCharsTablePath, originalContent, 0644)

	ctx := context.TODO()

	updateCases := []struct {
		name            string
		cells           []string
		expectedEscaped string
		expectedRawCell string
	}{
		{
			name:            "PipeBetweenLetters",
			cells:           []string{"plain", "a|b", "x,y", "\"q\""},
			expectedEscaped: `a\vert{}b`,
			expectedRawCell: "a|b",
		},
		{
			name:            "PipeAtStartOfCell",
			cells:           []string{"plain", "|start", "x,y", "\"q\""},
			expectedEscaped: `\vert{}start`,
			expectedRawCell: "|start",
		},
		{
			name:            "PipeAtEndOfCell",
			cells:           []string{"plain", "end|", "x,y", "\"q\""},
			expectedEscaped: `end\vert`,
			expectedRawCell: "end|",
		},
		{
			name:            "MultiplePipesInsideWord",
			cells:           []string{"plain", "a|b|c", "x,y", "\"q\""},
			expectedEscaped: `a\vert{}b\vert{}c`,
			expectedRawCell: "a|b|c",
		},
	}

	for _, tc := range updateCases {
		t.Run(tc.name, func(t *testing.T) {
			os.WriteFile(specialCharsTablePath, originalContent, 0644)
			reloadOrgFile(t, ctx)

			input := tools.ManageTableInput{
				Items: []mcp.GenericOneOf[*tools.ManageTableInputUnion, tools.TableApplicableTool]{
					{Value: tools.NewManageTableInputUnion(tools.UpdateTableRow{
						Method: "update",
						Uid:    NewUid("66666666.special_chars_table"),
						Row:    1,
						Cells:  tc.cells,
					})},
				},
				Path: specialCharsTablePath,
			}

			if _, err := tools.ManageTableTool.Callback(ctx, input, mcp.FuncOptions{DefaultPath: specialCharsTablePath}); err != nil {
				t.Fatalf("update failed: %v", err)
			}

			content, err := os.ReadFile(specialCharsTablePath)
			if err != nil {
				t.Fatal(err)
			}
			fileStr := string(content)

			if !strings.Contains(fileStr, tc.expectedEscaped) {
				t.Errorf("expected file to contain escaped form %q, got:\n%s", tc.expectedEscaped, fileStr)
			}

			if strings.Contains(fileStr, tc.expectedRawCell) && !strings.Contains(fileStr, tc.expectedEscaped) {
				t.Errorf("file contains raw cell value %q without the escaped form %q, got:\n%s", tc.expectedRawCell, tc.expectedEscaped, fileStr)
			}

			reloadOrgFile(t, ctx)

			got := querySpecialCharsCell(t, ctx, 1)
			if !ContainsString(got, tc.expectedRawCell) {
				t.Errorf("round-trip query mismatch: expected cell to contain %q, got %q", tc.expectedRawCell, got)
			}
		})

		os.WriteFile(specialCharsTablePath, originalContent, 0644)
	}
}
