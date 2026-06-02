package codeblock

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/utils/reader"
)

func TestParseBeginSrc(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError error
		expect      any
	}{
		{
			name:   "WithLanguage",
			input:  "#+BEGIN_SRC rust\n",
			expect: "rust",
		},
		{
			name:   "WithoutLanguage",
			input:  "#+BEGIN_SRC\n",
			expect: "",
		},
		{
			name:   "WithSpaces",
			input:  "    #+BEGIN_SRC go    \n",
			expect: "go",
		},
		{
			name:   "WithSpaces",
			input:  "    #+BEGIN_SRC    go    \n",
			expect: "go",
		},
		{
			name:        "FailureUnexpectedEnd",
			input:       "#+BEG",
			expectError: fmt.Errorf(UNEXPECTED_END, "CodeBlock"),
		},
		{
			name:        "FailureInvalidPrefix",
			input:       "#+BEGIN_SCR\n",
			expectError: fmt.Errorf(INVALID_PREFIX, "#+BEGIN_SCR", "CodeBlock"),
		},
		{
			name:        "WithDanglingChar",
			input:       "#+BEGIN_SRCd go\n",
			expectError: fmt.Errorf(INVALID_PREFIX, "#+BEGIN_SRCd", "CodeBlock"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reader.NewPeekReader(bufio.NewReader(strings.NewReader(tt.input)))
			begin, err := parseBeginSrc(r)

			if err != nil {
				if tt.expectError == nil {
					t.Errorf("Expected: `%s`, but got error: `%s`\n", tt.expect, err.Error())
				}

				if err.Error() != tt.expectError.Error() {
					t.Errorf("Expected error: `%s`, but got error: `%s`\n", tt.expectError, err.Error())
				}
			}

			if err == nil && tt.expectError != nil {
				t.Errorf("Expected an error but got: `%s`\n", begin.lang)
			}

			if begin != nil && begin.lang != tt.expect {
				t.Errorf("Expected: `%s`, but got: `%s`\n", tt.expect, tt.input)
			}

			if begin == nil && tt.expect != nil {
				t.Errorf("Expected: `%s`, but got nothing\n", tt.expect)
			}
		})
	}
}
