package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
)

func TestViewTool(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	type Test struct {
		name     string
		input    tools.ViewInput
		expected []any
	}

	var depth = 0
	var depth2 = 2
	var Todo = orgmcp.Todo

	var testMap = []Test{
		{
			name: "GetRootNode",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "1",
						Depth: &depth,
					},
				},
			},
			expected: []any{"UID,CONTENT\\n1,* Root Header..."},
		},
		{
			name: "GetNodeByStatus",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Status: &Todo,
						Depth:  &depth,
					},
				},
			},
			expected: []any{"UID,CONTENT\\n2,* TODO Root Header with status..."},
		},
		{
			name: "GetSpecificColumns",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "3",
						Depth: &depth,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.UidCol,
					&orgmcp.ProgressCol,
					&orgmcp.ChildrenCountCol,
				},
			},
			expected: []any{"UID,PROGRESS,CHILDREN_COUNT\\n3,1/2,2"},
		},
		{
			name: "GetHeaderWithDepth",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "3",
						Depth: &depth2,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.UidCol,
				},
			},
			expected: []any{"UID\\n3\\n3.b0\\n3.b1"},
		},
	}

	for _, tt := range testMap {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := json.Marshal(tt.input)
			if err != nil {
				t.Errorf("failed to marshal input: %v", err)
			}

			var inputMap map[string]any
			err = json.Unmarshal(jsonStr, &inputMap)

			res, err := tools.ViewTool.Callback(inputMap, mcp.FuncOptions{DefaultPath: "./test.org"})
			if err != nil {
				t.Errorf("HeaderTool failed: %v", err)
			}

			for _, expectedStr := range tt.expected {
				found := false
				for _, v := range res {
					jsonStr, err := json.Marshal(v)
					if err != nil {
						t.Errorf("failed to marshal response")
					}

					// Unescape newlines and trim quotes
					str := strings.Trim(string(jsonStr), "\"\\n")

					fmt.Fprintf(os.Stderr, "%s == %s\n", strings.TrimSpace(str), strings.TrimSpace(expectedStr.(string)))

					if EqualString(str, expectedStr.(string)) {
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
}
