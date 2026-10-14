package codeblock

import (
	"strings"
	"testing"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
)

type renderCodeBlockTest struct {
	caseName  string
	blockName option.Option[string]
	content   string
	lang      option.Option[string]
	indent    int
	depth     int
	expected  string
}

type previewCodeBlockTest struct {
	caseName string
	content  string
	lang     option.Option[string]
	expected string
}

type mockParent struct {
	uid         Uid
	indentLevel int
	childIndent int
}

func newMockParent(indent int) *mockParent {
	return &mockParent{
		uid:         NewUid("mock"),
		indentLevel: 0,
		childIndent: indent,
	}
}

func (m *mockParent) Uid() Uid                               { return m.uid }
func (m *mockParent) ParentUid() Uid                         { return NewUid(0) }
func (m *mockParent) Level() int                             { return 0 }
func (m *mockParent) IndentLevel() int                       { return m.indentLevel }
func (m *mockParent) ChildIndentLevel() int                  { return m.childIndent }
func (m *mockParent) Location(map[Uid]int) int               { return 0 }
func (m *mockParent) Path() string                           { return "" }
func (m *mockParent) Status() RenderStatus                   { return "" }
func (m *mockParent) CheckProgress() option.Option[Progress] { return option.None[Progress]() }
func (m *mockParent) TagList() TagList                       { return nil }
func (m *mockParent) Render(*strings.Builder, int)           {}
func (m *mockParent) RenderMarkdown(*strings.Builder, int)   {}
func (m *mockParent) Preview(int) string                     { return "" }
func (m *mockParent) AddChildren(...Render) error            { return nil }
func (m *mockParent) RemoveChildren(...Uid) error            { return nil }
func (m *mockParent) Children() []Render                     { return nil }
func (m *mockParent) ChildrenRec(int) []Render               { return nil }
func (m *mockParent) Insert(int, Render) error               { return nil }
func (m *mockParent) Move(MoveOperation) error               { return nil }
func (m *mockParent) SetParent(Render) error                 { return nil }

func TestRenderCodeBlock(t *testing.T) {
	tests := []renderCodeBlockTest{
		{
			caseName: "WithLanguage",
			content:  "fn main() {}\n",
			lang:     option.Some("rust"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC rust\nfn main() {}\n#+END_SRC\n",
		},
		{
			caseName: "WithoutLanguage",
			content:  "some content\n",
			lang:     option.None[string](),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC\nsome content\n#+END_SRC\n",
		},
		{
			caseName: "Indented",
			content:  "let x = 1;\n",
			lang:     option.Some("rust"),
			indent:   2,
			depth:    1,
			expected: "  #+BEGIN_SRC rust\n  let x = 1;\n  #+END_SRC\n",
		},
		{
			caseName: "EmptyContent",
			content:  "",
			lang:     option.Some("rust"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC rust\n#+END_SRC\n",
		},
		{
			caseName: "MultiLineContent",
			content:  "line 1\nline 2\nline 3\n",
			lang:     option.Some("python"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC python\nline 1\nline 2\nline 3\n#+END_SRC\n",
		},
		{
			caseName: "SpecialChars",
			content:  "<p>\"hi\" & <em>bye</em></p>\n",
			lang:     option.Some("html"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC html\n<p>\"hi\" & <em>bye</em></p>\n#+END_SRC\n",
		},
		{
			caseName: "EscapedChars",
			content:  "char* s = \"He said \\\"hi\\\"\";\n",
			lang:     option.Some("c"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC c\nchar* s = \"He said \\\"hi\\\"\";\n#+END_SRC\n",
		},
		{
			caseName:  "Named",
			blockName: option.Some("greet"),
			content:   "fn main() {}\n",
			lang:      option.Some("rust"),
			indent:    0,
			depth:     1,
			expected:  "#+NAME: greet\n#+BEGIN_SRC rust\nfn main() {}\n#+END_SRC\n",
		},
		{
			caseName: "LangWithSpecialChars",
			content:  "std::cout << \"hi\";\n",
			lang:     option.Some("c++"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC c++\nstd::cout << \"hi\";\n#+END_SRC\n",
		},
		{
			caseName: "UnicodeContent",
			content:  "こんにちは\n世界\n",
			lang:     option.Some("text"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC text\nこんにちは\n世界\n#+END_SRC\n",
		},
		{
			caseName: "TabsAndSpecialChars",
			content:  "if x:\n\tprint('hi', \"world\")\n",
			lang:     option.Some("python"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC python\nif x:\n\tprint('hi', \"world\")\n#+END_SRC\n",
		},
		{
			caseName: "EndSrcAsSubstring",
			content:  "# echo \"this #+END_SRC is just text\"\necho done\n",
			lang:     option.Some("bash"),
			indent:   0,
			depth:    1,
			expected: "#+BEGIN_SRC bash\n# echo \"this #+END_SRC is just text\"\necho done\n#+END_SRC\n",
		},
		{
			caseName: "DepthZeroRendersFully",
			content:  "fn main() {}\n",
			lang:     option.Some("rust"),
			indent:   0,
			depth:    0,
			expected: "#+BEGIN_SRC rust\nfn main() {}\n#+END_SRC\n",
		},
		{
			caseName: "HighIndent",
			content:  "x\n",
			lang:     option.Some("rust"),
			indent:   32,
			depth:    1,
			expected: "                                #+BEGIN_SRC rust\n                                x\n                                #+END_SRC\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			cb := NewCodeBlock(tt.content, tt.blockName, tt.lang)
			cb.SetParent(newMockParent(tt.indent))

			builder := &strings.Builder{}
			cb.Render(builder, tt.depth)

			if got := builder.String(); got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestRenderMarkdownCodeBlock(t *testing.T) {
	tests := []renderCodeBlockTest{
		{
			caseName: "WithLanguage",
			content:  "fn main() {}\n",
			lang:     option.Some("rust"),
			indent:   0,
			depth:    1,
			expected: "```rust\nfn main() {}\n```\n",
		},
		{
			caseName: "WithoutLanguage",
			content:  "some content\n",
			lang:     option.None[string](),
			indent:   0,
			depth:    1,
			expected: "```\nsome content\n```\n",
		},
		{
			caseName: "Indented",
			content:  "let x = 1;\n",
			lang:     option.Some("rust"),
			indent:   2,
			depth:    1,
			expected: "  ```rust\n  let x = 1;\n  ```\n",
		},
		{
			caseName: "EmptyContent",
			content:  "",
			lang:     option.Some("rust"),
			indent:   0,
			depth:    1,
			expected: "```rust\n```\n",
		},
		{
			caseName: "MultiLineContent",
			content:  "line 1\nline 2\nline 3\n",
			lang:     option.Some("python"),
			indent:   0,
			depth:    1,
			expected: "```python\nline 1\nline 2\nline 3\n```\n",
		},
		{
			caseName: "SpecialChars",
			content:  "<p>\"hi\" & <em>bye</em></p>\n",
			lang:     option.Some("html"),
			indent:   0,
			depth:    1,
			expected: "```html\n<p>\"hi\" & <em>bye</em></p>\n```\n",
		},
		{
			caseName: "LangWithSpecialChars",
			content:  "std::cout << \"hi\";\n",
			lang:     option.Some("c++"),
			indent:   0,
			depth:    1,
			expected: "```c++\nstd::cout << \"hi\";\n```\n",
		},
		{
			caseName: "UnicodeContent",
			content:  "こんにちは\n世界\n",
			lang:     option.Some("text"),
			indent:   0,
			depth:    1,
			expected: "```text\nこんにちは\n世界\n```\n",
		},
		{
			caseName: "TabsAndSpecialChars",
			content:  "if x:\n\tprint('hi', \"world\")\n",
			lang:     option.Some("python"),
			indent:   0,
			depth:    1,
			expected: "```python\nif x:\n\tprint('hi', \"world\")\n```\n",
		},
		{
			caseName: "EndSrcAsSubstring",
			content:  "# echo \"this #+END_SRC is just text\"\necho done\n",
			lang:     option.Some("bash"),
			indent:   0,
			depth:    1,
			expected: "```bash\n# echo \"this #+END_SRC is just text\"\necho done\n```\n",
		},
		{
			caseName: "DepthZeroRendersFully",
			content:  "fn main() {}\n",
			lang:     option.Some("rust"),
			indent:   0,
			depth:    0,
			expected: "```rust\nfn main() {}\n```\n",
		},
		{
			caseName: "HighIndent",
			content:  "x\n",
			lang:     option.Some("rust"),
			indent:   32,
			depth:    1,
			expected: "                                ```rust\n                                x\n                                ```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			cb := NewCodeBlock(tt.content, tt.blockName, tt.lang)
			cb.SetParent(newMockParent(tt.indent))

			builder := &strings.Builder{}
			cb.RenderMarkdown(builder, tt.depth)

			if got := builder.String(); got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestPreviewCodeBlock(t *testing.T) {
	tests := []previewCodeBlockTest{
		{
			caseName: "EmptyWithLanguage",
			content:  "",
			lang:     option.Some("rust"),
			expected: "EMPTY CODE BLOCK ; LANG=`rust`",
		},
		{
			caseName: "EmptyWithoutLanguage",
			content:  "",
			lang:     option.None[string](),
			expected: "EMPTY CODE BLOCK ; LANG=`UNKNOWN`",
		},
		{
			caseName: "WithLanguage",
			content:  "fn main() {}\nlet x = 1;\n",
			lang:     option.Some("rust"),
			expected: "CODE BLOCK ; LANG=`rust`; FIRST LINE: fn main() {}",
		},
		{
			caseName: "WithoutLanguage",
			content:  "some content\nmore content\n",
			lang:     option.None[string](),
			expected: "CODE BLOCK ; LANG=`UNKNOWN`; FIRST LINE: some content",
		},
		{
			caseName: "WithSpecialCharsInFirstLine",
			content:  "<p>\"hi\"</p>\nrest\n",
			lang:     option.Some("html"),
			expected: "CODE BLOCK ; LANG=`html`; FIRST LINE: <p>\"hi\"</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			cb := NewCodeBlock(tt.content, option.None[string](), tt.lang)

			if got := cb.Preview(80); got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestRenderContentIndentationNotModified(t *testing.T) {
	tests := []struct {
		caseName    string
		content     string
		lang        option.Option[string]
		indent      int
		contentLine string
	}{
		{
			caseName:    "ContentWithLeadingSpaces",
			content:     "    x = 1\n    y = 2\n",
			lang:        option.Some("python"),
			indent:      2,
			contentLine: "      x = 1",
		},
		{
			caseName:    "ContentWithTabs",
			content:     "\tif x:\n\t\tprint('hi')\n",
			lang:        option.Some("python"),
			indent:      2,
			contentLine: "  \tif x:",
		},
		{
			caseName:    "ContentNoIndent",
			content:     "fn main() {}\n",
			lang:        option.Some("rust"),
			indent:      0,
			contentLine: "fn main() {}",
		},
		{
			caseName:    "HighIndentDoesNotAffectContent",
			content:     "    nested = True\n",
			lang:        option.Some("python"),
			indent:      8,
			contentLine: "            nested = True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			cb := NewCodeBlock(tt.content, option.None[string](), tt.lang)
			cb.SetParent(newMockParent(tt.indent))

			builder := &strings.Builder{}
			cb.Render(builder, 1)

			rendered := builder.String()

			if !strings.Contains(rendered, tt.contentLine) {
				t.Errorf("content indentation modified:\ncontent line %q not found in:\n%s", tt.contentLine, rendered)
			}

			if tt.indent > 0 && strings.Contains(rendered, strings.Repeat(" ", tt.indent)+tt.contentLine) {
				t.Errorf("content incorrectly indented by parent indent:\n%s, expected: %s\n", rendered, strings.Repeat(" ", tt.indent)+tt.contentLine)
			}
		})
	}
}
