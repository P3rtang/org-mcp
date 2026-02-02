package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/tools"
)

func EqualString(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}

func TestHeaderGetMethod(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	type Test struct {
		name     string
		input    tools.HeaderInput
		expected []any
	}

	var testMap = []Test{}

	for _, tt := range testMap {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := json.Marshal(tt.input)
			if err != nil {
				t.Errorf("failed to marshal input: %v", err)
			}

			var inputMap map[string]any
			err = json.Unmarshal(jsonStr, &inputMap)

			res, err := tools.HeaderTool.Callback(inputMap, mcp.FuncOptions{DefaultPath: "./test.org"})
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
