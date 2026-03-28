package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
)

func TestTextTool(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	// Save original test.org content to restore at end
	originalContent, err := os.ReadFile("./test.org")
	if err != nil {
		t.Fatalf("failed to read test.org: %v", err)
	}
	defer func() {
		os.WriteFile("./test.org", originalContent, 0644)
	}()

	// Use stable UIDs from test.org
	addHeaderUid := "99998888"
	updateHeaderUid := "99998889"
	plainTextUid := "99998889.t0"

	// UID for a bullet in test.org (header 3 has bullets 3.b0 and 3.b1)
	bulletWithChildrenUid := "3.b0"

	tests := []ManageTextTest{
		{
			name: "AddTextToHeader",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "Added text content",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
			},
			expected: []any{"Added text content"},
		},
		{
			name: "UpdateTextContent",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     plainTextUid,
						Method:  "update",
						Content: "Updated text content",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
			},
			expected: []any{"Updated text content"},
		},
		{
			name: "AddTextWithMultipleColumns",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "Additional text",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColContentValue,
				},
			},
			expected: []any{"Additional text"},
		},
		{
			name: "ShowDiff",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     plainTextUid,
						Method:  "update",
						Content: "Content with diff",
					},
				},
				ShowDiff: true,
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"Content with diff"},
		},
		{
			name: "InvalidUid",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:    "invalid-uid-12345",
						Method: "add",
					},
				},
			},
			expected: []any{"Uid invalid-uid-12345 not found"},
		},
		{
			name: "UpdateNonPlainTextFails",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     updateHeaderUid,
						Method:  "update",
						Content: "This should fail",
					},
				},
			},
			expected: []any{"is not a plain text element, cannot update content"},
		},
		{
			name: "NewlinesReplacedWithSpaces",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "Text with\nnewlines",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColContentValue,
				},
			},
			expected: []any{"newlines will be replaced with spaces"},
		},
		{
			name: "ShowAffectedFalse",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "Test content",
					},
				},
				ShowAffected: func() *bool { b := false; return &b }(),
			},
			expectEmpty: true,
		},
		{
			name: "RemoveTextContent",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:    plainTextUid,
						Method: "remove",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"99998889.t0"},
		},
		{
			name: "AddMultipleTextsToSameHeader",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "First addition",
					},
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "Second addition",
					},
					{
						Uid:     addHeaderUid,
						Method:  "add",
						Content: "Third addition",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
			},
			expected: []any{"First addition", "Second addition", "Third addition"},
		},
		{
			name: "RemoveNonExistentTextFails",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:    "99998889.t999",
						Method: "remove",
					},
				},
			},
			expected: []any{"Uid 99998889.t999 not found in ./test.org"},
		},
		{
			// Plain text CAN be added as a child of a bullet
			name: "AddPlainTextToBullet",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     bulletWithChildrenUid,
						Method:  "add",
						Content: "Text under a bullet",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
				ShowDiff: true,
			},
			expected: []any{"Text under a bullet"},
		},
		{
			// Update plain text that exists under a bullet in test.org
			name: "UpdatePlainTextUnderBullet",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     "99998891.b0.t0",
						Method:  "update",
						Content: "Updated plain text under bullet",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
			},
			expected: []any{"Updated plain text under bullet"},
		},

		{
			name: "AddTextToHeaderWithBullet",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     "3",
						Method:  "add",
						Content: "Text under header with bullets",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
			},
			expected: []any{"Text under header with bullets"},
		},
		{
			name: "HeaderCorruptionRepro",
			input: tools.TextInputSchema{
				Texts: []tools.TextInputValue{
					{
						Uid:     "99998892.t0",
						Method:  "update",
						Content: "Updated text content.",
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColPreviewValue,
				},
				ShowDiff: true,
			},
			expected: []any{"Updated text content."},
		},
	}

	ctx := context.TODO()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tools.TextTool.Callback(ctx, tt.input, mcp.FuncOptions{DefaultPath: "./test.org"})
			if err != nil {
				t.Errorf("TextTool failed: %v", err)
			}

			if tt.expectEmpty {
				if len(res) != 0 {
					t.Errorf("expected empty response, got %#v", res)
				}
				return
			}

			fmt.Fprintf(os.Stderr, "Response for test '%s': %#v\n", tt.name, res)

			for _, expectedStr := range tt.expected {
				found := false
				for _, v := range res {
					var str string
					switch val := v.(type) {
					case string:
						str = val
					default:
						jsonBytes, jsonErr := json.Marshal(v)
						if jsonErr != nil {
							t.Errorf("failed to marshal response: %v", jsonErr)
							continue
						}
						str = strings.Trim(string(jsonBytes), "\"\\n")
					}

					fmt.Fprintf(os.Stderr, "%s == %s\n", strings.TrimSpace(str), strings.TrimSpace(expectedStr.(string)))

					if ContainsString(str, expectedStr.(string)) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("unexpected response: got %#v, expected to contain '%s'\n", res, strings.Trim(expectedStr.(string), "\n"))
				}
			}
		})
	}

	// Restore original state
	os.WriteFile("./test.org", originalContent, 0644)
}
