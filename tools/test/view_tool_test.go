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
	"github.com/p3rtang/org-mcp/utils/slice"
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

	var (
		depth        = 0
		depth2       = 2
		Todo         = orgmcp.RenderStatus(orgmcp.Todo)
		twoDaysRange = 2
		startDate    = "2025-12-31"
	)

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
			expected: []any{"UID,PREVIEW\\n1,Root Header"},
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
			expected: []any{"UID,PREVIEW\\n2,Root Header with status"},
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
					&orgmcp.ColUidValue,
					&orgmcp.ColProgressValue,
					&orgmcp.ColChildrenCountValue,
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
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"UID\\n3\\n3.b0\\n3.b1"},
		},
		{
			name: "GetByRegex",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Content: "with status",
						Depth:   &depth,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"UID\\n2"},
		},
		{
			name: "GetByOverdue",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Date: &tools.DateFilter{
							Match: orgmcp.DeadlineValue,
						},
						Depth: &depth,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"UID\\n95718900"},
		},
		{
			name: "GetByDateShowClosed",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Date: &tools.DateFilter{
							Match:      orgmcp.DeadlineValue,
							ShowClosed: true,
						},
						Depth: &depth,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"UID\\n95718900\\n95718910"},
		},
		{
			name: "GetByDateRange",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Date: &tools.DateFilter{
							Match: orgmcp.DeadlineValue,
							Date:  &startDate,
							Range: &twoDaysRange,
						},
						Depth: &depth,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
			},
			expected: []any{"UID\\n95718900"},
		},
		{
			name: "GetAllColumns",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "95718920",
						Depth: &depth,
					},
				},
				Columns: slice.Ref(orgmcp.AllColumns),
			},
			expected: []any{"TYPE,UID,PREVIEW,CONTENT,STATUS,PROGRESS,PARENT,CHILDREN_COUNT,TAGS,LEVEL,PATH,SCHEDULED,DEADLINE,CLOSED\\n*orgmcp.Header,95718920,All columns,* DONE All columns [1/3] :tag:...,DONE,1/3,0,3,tag,1,/95718920,2026-02-02,2026-02-03,2026-02-02"},
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
