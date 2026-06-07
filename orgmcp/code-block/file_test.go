package codeblock_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/orgmcp/code-block"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

type roundTripCodeBlockTest struct {
	caseName         string
	orgContent       string
	headerUid        string
	codeBlockUid     string
	expectedRendered string
}

type parseFileCodeBlockTest struct {
	caseName       string
	orgContent     string
	headerUid      string
	expectedBlocks []expectedFileCodeBlock
}

type expectedFileCodeBlock struct {
	uid        string
	wantLang   string
	wantName   string
	contentHas string
}

func TestParseFileWithCodeBlock(t *testing.T) {
	tests := []parseFileCodeBlockTest{
		{
			caseName: "SingleCodeBlock",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 11111111
  :END:
  #+BEGIN_SRC rust
  fn main() {}
  #+END_SRC
`,
			headerUid: "11111111",
			expectedBlocks: []expectedFileCodeBlock{
				{
					uid:        "11111111.c0",
					wantLang:   "rust",
					contentHas: "fn main()",
				},
			},
		},
		{
			caseName: "NamedCodeBlock",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 22222222
  :END:
  #+NAME: greet
  #+BEGIN_SRC python
  def greet(name):
      print(name)
  #+END_SRC
`,
			headerUid: "22222222",
			expectedBlocks: []expectedFileCodeBlock{
				{
					uid:        "22222222.greet",
					wantLang:   "python",
					wantName:   "greet",
					contentHas: "def greet",
				},
			},
		},
		{
			caseName: "MultipleCodeBlocks",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 33333333
  :END:
  #+BEGIN_SRC go
  package main
  #+END_SRC
  #+BEGIN_SRC bash
  echo hello
  #+END_SRC
`,
			headerUid: "33333333",
			expectedBlocks: []expectedFileCodeBlock{
				{
					uid:        "33333333.c0",
					wantLang:   "go",
					contentHas: "package main",
				},
				{
					uid:        "33333333.c1",
					wantLang:   "bash",
					contentHas: "echo hello",
				},
			},
		},
		{
			caseName: "CodeBlockWithoutLanguage",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 44444444
  :END:
  #+BEGIN_SRC
  some content
  #+END_SRC
`,
			headerUid: "44444444",
			expectedBlocks: []expectedFileCodeBlock{
				{
					uid:        "44444444.c0",
					wantLang:   "",
					contentHas: "some content",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			showDebug := os.Getenv("SHOW_DEBUG")
			if showDebug == "" {
				os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
			}

			of, err := orgmcp.OrgFileFromReader(context.TODO(), strings.NewReader(tt.orgContent)).Split()
			if err != nil {
				t.Fatalf("failed to parse org file: %v", err)
			}

			headerRender, ok := of.GetUid(NewUid(tt.headerUid)).Split()
			if !ok {
				t.Fatalf("expected header %s to exist", tt.headerUid)
			}

			header, ok := headerRender.(*orgmcp.Header)
			if !ok {
				t.Fatalf("expected %s to be a Header, got %T", tt.headerUid, headerRender)
			}

			for _, expected := range tt.expectedBlocks {
				cbRender, ok := of.GetUid(NewUid(expected.uid)).Split()
				if !ok {
					t.Errorf("expected code block %s to exist", expected.uid)
					continue
				}

				cb, ok := cbRender.(*codeblock.CodeBlock)
				if !ok {
					t.Errorf("expected %s to be a CodeBlock, got %T", expected.uid, cbRender)
					continue
				}

				found := false
				for _, child := range header.Children() {
					if child.Uid() == cb.Uid() {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected code block %s to be a child of header %s", expected.uid, tt.headerUid)
				}

				if lang := cb.Language().UnwrapOrDefault(); lang != expected.wantLang {
					t.Errorf("expected language %q, got %q", expected.wantLang, lang)
				}

				if name := cb.Name().UnwrapOrDefault(); name != expected.wantName {
					t.Errorf("expected name %q, got %q", expected.wantName, name)
				}

				if expected.contentHas != "" {
					allLines := strings.Join(cb.Lines(), "\n")
					if !strings.Contains(allLines, expected.contentHas) {
						t.Errorf("expected content to contain %q, got %q", expected.contentHas, allLines)
					}
				}
			}
		})
	}
}

func TestRoundTripCodeBlock(t *testing.T) {
	tests := []roundTripCodeBlockTest{
		{
			caseName: "BasicWithLanguage",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 11111111
  :END:
  #+BEGIN_SRC rust
  fn main() {}
  #+END_SRC
`,
			headerUid:    "11111111",
			codeBlockUid: "11111111.c0",
			expectedRendered: `  #+BEGIN_SRC rust
  fn main() {}
  #+END_SRC
`,
		},
		{
			caseName: "WithName",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 22222222
  :END:
  #+NAME: greet
  #+BEGIN_SRC python
  def greet(name):
    print(name)
  #+END_SRC
`,
			headerUid:    "22222222",
			codeBlockUid: "22222222.greet",
			expectedRendered: `  #+NAME: greet
  #+BEGIN_SRC python
  def greet(name):
    print(name)
  #+END_SRC
`,
		},
		{
			caseName: "WithoutLanguage",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 33333333
  :END:
  #+BEGIN_SRC
  some content
  #+END_SRC
`,
			headerUid:    "33333333",
			codeBlockUid: "33333333.c0",
			expectedRendered: `  #+BEGIN_SRC
  some content
  #+END_SRC
`,
		},
		{
			caseName: "EmptyContent",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 44444444
  :END:
  #+BEGIN_SRC rust
  #+END_SRC
`,
			headerUid:    "44444444",
			codeBlockUid: "44444444.c0",
			expectedRendered: `  #+BEGIN_SRC rust
  #+END_SRC
`,
		},
		{
			caseName: "MultiLineContent",
			orgContent: `* Header
  :PROPERTIES:
  :ID: 55555555
  :END:
  #+BEGIN_SRC go
  package main
  
  import "fmt"
  
  func main() {
  	  fmt.Println("hello")
  }
  #+END_SRC
`,
			headerUid:    "55555555",
			codeBlockUid: "55555555.c0",
			expectedRendered: `  #+BEGIN_SRC go
  package main
  
  import "fmt"
  
  func main() {
  	  fmt.Println("hello")
  }
  #+END_SRC
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			showDebug := os.Getenv("SHOW_DEBUG")
			if showDebug == "" {
				os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
			}

			of, err := orgmcp.OrgFileFromReader(context.TODO(), strings.NewReader(tt.orgContent)).Split()
			if err != nil {
				t.Fatalf("failed to parse org file: %v", err)
			}

			cbRender, ok := of.GetUid(NewUid(tt.codeBlockUid)).Split()
			if !ok {
				t.Fatalf("expected code block %s to exist", tt.codeBlockUid)
			}

			cb, ok := cbRender.(*codeblock.CodeBlock)
			if !ok {
				t.Fatalf("expected %s to be a CodeBlock, got %T", tt.codeBlockUid, cbRender)
			}

			builder := &strings.Builder{}
			cb.Render(builder, 0)

			if got := builder.String(); got != tt.expectedRendered {
				t.Errorf("round-trip mismatch:\nwant:\n%s\ngot:\n%s", tt.expectedRendered, got)
			}
		})
	}
}
