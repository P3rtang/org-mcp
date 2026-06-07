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
	. "github.com/p3rtang/org-mcp/tools/test/utils"
)

const CODE_BLOCK_QUERY_TEST_PATH = "./codeblock_test_query.org"

var DEPTH0 = 0
var DEPTH1 = 1

type CodeBlockViewTest struct {
	name     string
	input    tools.ViewInput
	expected []any
}

func runCodeBlockViewTest(t *testing.T, tt CodeBlockViewTest) {
	t.Helper()

	res, err := tools.ViewTool.Callback(context.TODO(), tt.input, mcp.FuncOptions{DefaultPath: CODE_BLOCK_QUERY_TEST_PATH})
	if err != nil {
		t.Errorf("ViewTool failed: %v", err)
		return
	}

	for _, expectedStr := range tt.expected {
		found := false
		for _, v := range res {
			jsonStr, err := json.Marshal(v)
			if err != nil {
				t.Errorf("failed to marshal response")
				continue
			}

			str := strings.Trim(string(jsonStr), "\"")

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
}

func TestQueryItemsCodeBlocks(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of, err := mcp.LoadOrgFile(context.TODO(), CODE_BLOCK_QUERY_TEST_PATH)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, CODE_BLOCK_QUERY_TEST_PATH)

	tests := []CodeBlockViewTest{
		{
			name: "GetCodeBlockByNamedUid",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "33333333.http_handler",
						Depth: &DEPTH0,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColTypeValue,
				},
				Path: CODE_BLOCK_QUERY_TEST_PATH,
			},
			expected: []any{
				`UID,TYPE\n33333333.http_handler,*codeblock.CodeBlock`,
			},
		},
		{
			name: "GetCodeBlockLanguageColumn",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "33333333.http_handler",
						Depth: &DEPTH0,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColLanguageValue,
				},
				Path: CODE_BLOCK_QUERY_TEST_PATH,
			},
			expected: []any{
				`UID,LANGUAGE\n33333333.http_handler,go`,
			},
		},
		{
			name: "GetUnnamedCodeBlockLanguage",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "44444444.c0",
						Depth: &DEPTH0,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
					&orgmcp.ColLanguageValue,
				},
				Path: CODE_BLOCK_QUERY_TEST_PATH,
			},
			expected: []any{
				`UID,LANGUAGE\n44444444.c0,rust`,
			},
		},
		{
			name: "FilterCodeBlocksByContentRegex",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Content: "def greet",
						Depth:   &DEPTH0,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
				Path: CODE_BLOCK_QUERY_TEST_PATH,
			},
			expected: []any{
				`UID\n22222222.parse_json`,
			},
		},
		{
			name: "GetAllCodeBlocksUnderHeader",
			input: tools.ViewInput{
				Items: []tools.ViewItem{
					{
						Uid:   "33333333",
						Depth: &DEPTH1,
					},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue,
				},
				Path: CODE_BLOCK_QUERY_TEST_PATH,
			},
			expected: []any{
				`UID\n33333333\n33333333.http_handler\n33333333.main_func`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCodeBlockViewTest(t, tt)
		})
	}
}
