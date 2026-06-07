package codeblock

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/utils/reader"
)

const RUST_CONTENT = `
fn main() {
	println!("Hello, world!");
}
`

const RUST_BLOCK = "#+BEGIN_SRC rust\n" +
	RUST_CONTENT +
	"#+END_SRC\n"

const HTML_CONTENT = `
<div class="container">
	<p>Hello, "world" &amp; <em>goodbye</em>!</p>
	<a href="https://example.com?a=1&b=2">link</a>
</div>
`

const HTML_BLOCK = "#+BEGIN_SRC html\n" + HTML_CONTENT + "#+END_SRC\n"

const C_ESCAPED_CONTENT = `
char* greeting = "He said \"hi\"";
char* path = "C:\\Users\\test";
printf("%s\n", greeting);
`

const C_ESCAPED_BLOCK = "#+BEGIN_SRC c\n" + C_ESCAPED_CONTENT + "#+END_SRC\n"

const NO_LANG_BLOCK = "#+BEGIN_SRC\n" + RUST_CONTENT + "#+END_SRC\n"

const INDENTED_END_CONTENT = `
func main() {}
`

const INDENTED_END_BLOCK = "#+BEGIN_SRC go\n" + INDENTED_END_CONTENT + "    #+END_SRC\n"

const BASH_CONTENT = `
# echo "this #+END_SRC is just text"
echo done
`

const END_SRC_IN_CONTENT_BLOCK = "#+BEGIN_SRC bash\n" + BASH_CONTENT + "#+END_SRC\n"

const UNICODE_CONTENT = `
こんにちは
世界
🎉
`

const UNICODE_BLOCK = "#+BEGIN_SRC text\n" + UNICODE_CONTENT + "#+END_SRC\n"

const EMPTY_BLOCK = "#+BEGIN_SRC\n#+END_SRC\n"

const LOWERCASE_END_BLOCK = "#+BEGIN_SRC rust\nfn main() {}\n#+end_src\n"

const CPP_BLOCK = "#+BEGIN_SRC c++\nstd::cout << \"hi\";\n#+END_SRC\n"

const TRAILING_WS_END_BLOCK = "#+BEGIN_SRC rust\nfn main() {}\n#+END_SRC   \n"

const NO_END_BLOCK = "#+BEGIN_SRC rust\ncontent\n"

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

type expectedParseSrcBlock struct {
	lang    string
	content string
}

func (e expectedParseSrcBlock) isEqual(code_block CodeBlock) bool {
	return code_block.lang.UnwrapOrDefault() == e.lang &&
		code_block.content == e.content
}

func TestParseSrcBlock(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError error
		expect      expectedParseSrcBlock
	}{
		{
			name:  "WithLanguage",
			input: RUST_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "rust",
				content: RUST_CONTENT,
			},
		},
		{
			name:  "NoLanguage",
			input: NO_LANG_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "",
				content: RUST_CONTENT,
			},
		},
		{
			name:  "WithSpecialChars",
			input: HTML_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "html",
				content: HTML_CONTENT,
			},
		},
		{
			name:  "WithEscapedChars",
			input: C_ESCAPED_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "c",
				content: C_ESCAPED_CONTENT,
			},
		},
		{
			name:  "IndentedEndSrc",
			input: INDENTED_END_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "go",
				content: INDENTED_END_CONTENT,
			},
		},
		{
			name:  "EndSrcAsSubstring",
			input: END_SRC_IN_CONTENT_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "bash",
				content: BASH_CONTENT,
			},
		},
		{
			name:  "WithUnicodeContent",
			input: UNICODE_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "text",
				content: UNICODE_CONTENT,
			},
		},
		{
			name:  "EmptyContent",
			input: EMPTY_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "",
				content: "",
			},
		},
		{
			name:  "CaseInsensitiveEndSrc",
			input: LOWERCASE_END_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "rust",
				content: "fn main() {}\n",
			},
		},
		{
			name:  "LangWithSpecialChars",
			input: CPP_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "c++",
				content: "std::cout << \"hi\";\n",
			},
		},
		{
			name:  "EndSrcWithTrailingWhitespace",
			input: TRAILING_WS_END_BLOCK,
			expect: expectedParseSrcBlock{
				lang:    "rust",
				content: "fn main() {}\n",
			},
		},
		{
			name:        "MissingEndSrc",
			input:       NO_END_BLOCK,
			expectError: fmt.Errorf(UNEXPECTED_END, "CodeBlock"),
		},
		{
			name:        "FailureInvalidPrefix",
			input:       "#+BEGIN_SR rust\nsome content\n#+END_SRC",
			expectError: fmt.Errorf(INVALID_PREFIX, "#+BEGIN_SR", "CodeBlock"),
		},
		{
			name:        "FailureUnexpectedEnd",
			input:       "",
			expectError: fmt.Errorf(UNEXPECTED_END, "CodeBlock"),
		},
		{
			name:        "FailureTruncatedBegin",
			input:       "#+BEGIN",
			expectError: fmt.Errorf(UNEXPECTED_END, "CodeBlock"),
		},
		{
			name:        "FailureMismatchedEnd",
			input:       "#+BEGIN_SRC rust\ncontent\n#+END_SC",
			expectError: fmt.Errorf(UNEXPECTED_END, "CodeBlock"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reader.NewPeekReader(bufio.NewReader(strings.NewReader(tt.input)))
			code_block, err := NewCodeBlockFromReader(r)

			if err != nil {
				if tt.expectError == nil {
					t.Errorf("Expected: `%s`, but got error: `%s`\n", tt.expect, err.Error())
				}

				if err.Error() != tt.expectError.Error() {
					t.Errorf("Expected error: `%s`, but got error: `%s`\n", tt.expectError, err.Error())
				}
			}

			if !tt.expect.isEqual(code_block) {
				t.Errorf("Expected: `%#v`, but got: %#v", tt.expect, code_block)
			}
		})
	}
}

func TestParseContentIndentationPreserved(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		beginBlockIndent int
		content          string
	}{
		{
			name: "ContentWithIndentStripped",
			input: `    #+BEGIN_SRC python
        x = 1
        y = 2
    #+END_SRC
`,
			beginBlockIndent: 4,
			content:          "    x = 1\n    y = 2\n",
		},
		{
			name: "ContentWithMixedIndentStripped",
			input: `  #+BEGIN_SRC go
  func main() {
      fmt.Println("hi")
  }
  #+END_SRC
`,
			beginBlockIndent: 2,
			content:          "func main() {\n    fmt.Println(\"hi\")\n}\n",
		},
		{
			name:             "ContentWithTabsStripped",
			input:            "\t#+BEGIN_SRC python\n\t\tif x:\n\t\t\tprint('hi')\n\t#+END_SRC\n",
			beginBlockIndent: 1,
			content:          "\t\tif x:\n\t\t\tprint('hi')\n",
		},
		{
			name: "NoContentIndent",
			input: `#+BEGIN_SRC rust
fn main() {}
#+END_SRC
`,
			beginBlockIndent: 0,
			content:          "fn main() {}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reader.NewPeekReader(bufio.NewReader(strings.NewReader(tt.input)))
			cb, err := NewCodeBlockFromReader(r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cb.content != tt.content {
				t.Errorf("content indentation not preserved:\nwant: %q\ngot: %q", tt.content, cb.content)
			}
		})
	}
}
