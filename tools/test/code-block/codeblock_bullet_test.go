package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	codeblock "github.com/p3rtang/org-mcp/orgmcp/code-block"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
)

const CODE_BLOCK_BULLET_TEST_PATH = "./codeblock_bullet_test.org"

type parseBulletCodeBlockTest struct {
	name               string
	uid                string
	wantExists         bool
	wantType           string
	wantLanguage       string
	wantContentHas     string
	wantParentIsBullet bool
}

type renderBulletCodeBlockTest struct {
	name           string
	bulletUid      string
	codeBlockUid   string
	expectedRender string
}

type roundTripBulletCodeBlockTest struct {
	caseName         string
	orgContent       string
	bulletUid        string
	codeBlockUid     string
	expectedRendered string
}

func loadBulletCodeBlockOrgFile(t *testing.T) orgmcp.OrgFile {
	t.Helper()

	of, err := mcp.LoadOrgFile(context.TODO(), CODE_BLOCK_BULLET_TEST_PATH)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, CODE_BLOCK_BULLET_TEST_PATH)

	return of
}

func fetchCodeBlockFromOrg(t *testing.T, of *orgmcp.OrgFile, uidStr string) *codeblock.CodeBlock {
	t.Helper()

	render := of.GetUid(NewUid(uidStr))
	cb, ok := option.Cast[Render, *codeblock.CodeBlock](render).Split()
	if !ok {
		return nil
	}
	return cb
}

func fetchBulletFromOrg(t *testing.T, of *orgmcp.OrgFile, uidStr string) *orgmcp.Bullet {
	t.Helper()

	render, ok := of.GetUid(NewUid(uidStr)).Split()
	if !ok {
		return nil
	}
	b, ok := render.(*orgmcp.Bullet)
	if !ok {
		return nil
	}
	return b
}

func TestParseCodeBlockUnderBullet(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of := loadBulletCodeBlockOrgFile(t)

	tests := []parseBulletCodeBlockTest{
		{
			name:       "Header11111111Exists",
			uid:        "11111111",
			wantExists: true,
			wantType:   "Header",
		},
		{
			name:               "Bullet11111111b0Exists",
			uid:                "11111111.b0",
			wantExists:         true,
			wantType:           "Bullet",
			wantParentIsBullet: false,
		},
		{
			name:           "CodeBlock11111111b0c0IsPython",
			uid:            "11111111.b0.c0",
			wantExists:     true,
			wantType:       "CodeBlock",
			wantLanguage:   "python",
			wantContentHas: "print",
		},
		{
			name:       "Header22222222Exists",
			uid:        "22222222",
			wantExists: true,
			wantType:   "Header",
		},
		{
			name:       "Bullet22222222b0Exists",
			uid:        "22222222.b0",
			wantExists: true,
			wantType:   "Bullet",
		},
		{
			name:           "NamedCodeBlock22222222b0greetExists",
			uid:            "22222222.b0.greet",
			wantExists:     true,
			wantType:       "CodeBlock",
			wantLanguage:   "python",
			wantContentHas: "greet",
		},
		{
			name:       "Header33333333Exists",
			uid:        "33333333",
			wantExists: true,
			wantType:   "Header",
		},
		{
			name:       "Bullet33333333b0Exists",
			uid:        "33333333.b0",
			wantExists: true,
			wantType:   "Bullet",
		},
		{
			name:           "FirstCodeBlockUnder33333333b0",
			uid:            "33333333.b0.c0",
			wantExists:     true,
			wantType:       "CodeBlock",
			wantLanguage:   "go",
			wantContentHas: "package main",
		},
		{
			name:           "SecondCodeBlockUnder33333333b0",
			uid:            "33333333.b0.c1",
			wantExists:     true,
			wantType:       "CodeBlock",
			wantLanguage:   "bash",
			wantContentHas: "echo",
		},
		{
			name:       "Header44444444Exists",
			uid:        "44444444",
			wantExists: true,
			wantType:   "Header",
		},
		{
			name:       "Bullet44444444b0Exists",
			uid:        "44444444.b0",
			wantExists: true,
			wantType:   "Bullet",
		},
		{
			name:           "CodeBlock44444444b0c0IsRust",
			uid:            "44444444.b0.c0",
			wantExists:     true,
			wantType:       "CodeBlock",
			wantLanguage:   "rust",
			wantContentHas: "fn main",
		},
		{
			name:       "SubBullet44444444b0b1Exists",
			uid:        "44444444.b0.b1",
			wantExists: true,
			wantType:   "Bullet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			render, ok := of.GetUid(NewUid(tt.uid)).Split()
			if ok != tt.wantExists {
				t.Errorf("expected exists=%v for uid %s, got exists=%v", tt.wantExists, tt.uid, ok)
				return
			}
			if !ok {
				return
			}

			switch tt.wantType {
			case "Header":
				if _, isHeader := render.(*orgmcp.Header); !isHeader {
					t.Errorf("expected %s to be a Header, got %T", tt.uid, render)
				}
			case "Bullet":
				if _, isBullet := render.(*orgmcp.Bullet); !isBullet {
					t.Errorf("expected %s to be a Bullet, got %T", tt.uid, render)
				}
			case "CodeBlock":
				cb, isCodeBlock := render.(*codeblock.CodeBlock)
				if !isCodeBlock {
					t.Errorf("expected %s to be a CodeBlock, got %T", tt.uid, render)
					return
				}
				if tt.wantLanguage != "" {
					if got := cb.Language().UnwrapOrDefault(); got != tt.wantLanguage {
						t.Errorf("expected language %q, got %q", tt.wantLanguage, got)
					}
				}
				if tt.wantContentHas != "" {
					allLines := strings.Join(cb.Lines(), "\n")
					if !strings.Contains(allLines, tt.wantContentHas) {
						t.Errorf("expected content to contain %q, got %q", tt.wantContentHas, allLines)
					}
				}
			}

			if tt.wantParentIsBullet {
				if bullet, ok := render.(*orgmcp.Bullet); ok {
					parentUid := bullet.ParentUid()
					if parentUid.String() == "" || parentUid == NewUid(0) {
						t.Errorf("expected %s to have a bullet parent", tt.uid)
					}
				}
			}
		})
	}
}

func TestParseCodeBlockUnderBulletContentLeakage(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of := loadBulletCodeBlockOrgFile(t)

	for _, uidStr := range []string{
		"11111111.b0",
		"22222222.b0",
		"33333333.b0",
		"44444444.b0",
	} {
		t.Run(uidStr, func(t *testing.T) {
			render, ok := of.GetUid(NewUid(uidStr)).Split()
			if !ok {
				t.Fatal("expected bullet to exist")
			}

			bullet, ok := render.(*orgmcp.Bullet)
			if !ok {
				t.Fatal("expected bullet type")
			}

			content := bullet.Preview(200)
			leakedSubstrings := []string{"print", "greet", "package main", "fn main"}
			for _, leaked := range leakedSubstrings {
				if strings.Contains(content, leaked) {
					t.Errorf("bullet %s has content leaking from code block: %q", uidStr, content)
				}
			}
		})
	}
}

func TestRenderCodeBlockUnderBullet(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of := loadBulletCodeBlockOrgFile(t)

	tests := []renderBulletCodeBlockTest{
		{
			name:           "Bullet11111111b0WithPythonCodeBlock",
			bulletUid:      "11111111.b0",
			codeBlockUid:   "11111111.b0.c0",
			expectedRender: "    #+BEGIN_SRC python\n    print(\"hello\")\n    #+END_SRC\n",
		},
		{
			name:           "Bullet22222222b0WithNamedCodeBlock",
			bulletUid:      "22222222.b0",
			codeBlockUid:   "22222222.b0.greet",
			expectedRender: "    #+NAME: greet\n    #+BEGIN_SRC python\n    def greet(name):\n        return f\"Hello, {name}!\"\n    #+END_SRC\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := fetchCodeBlockFromOrg(t, &of, tt.codeBlockUid)
			if cb == nil {
				t.Fatalf("expected code block %s to exist", tt.codeBlockUid)
			}

			builder := &strings.Builder{}
			cb.Render(builder, 0)

			if got := builder.String(); got != tt.expectedRender {
				t.Errorf("render mismatch for %s:\nwant:\n%s\ngot:\n%s", tt.codeBlockUid, tt.expectedRender, got)
			}
		})
	}
}

func TestRoundTripCodeBlockUnderBullet(t *testing.T) {
	tests := []roundTripBulletCodeBlockTest{
		{
			caseName: "BasicBulletWithCodeBlock",
			orgContent: `* Header
			:PROPERTIES:
			:ID: 11111111
			:END:
			- [ ] Task
			#+BEGIN_SRC rust
			fn main() {}
			#+END_SRC
			`,
			bulletUid:    "11111111.b0",
			codeBlockUid: "11111111.b0.c0",
			expectedRendered: `    #+BEGIN_SRC rust
			fn main() {}
			#+END_SRC
			`,
		},
		{
			caseName: "BulletWithNamedCodeBlock",
			orgContent: `* Header
			:PROPERTIES:
			:ID: 22222222
			:END:
			- [x] Done
			#+NAME: greet
			#+BEGIN_SRC python
			def greet():
			print("hi")
			#+END_SRC
			`,
			bulletUid:    "22222222.b0",
			codeBlockUid: "22222222.b0.greet",
			expectedRendered: `    #+NAME: greet
			#+BEGIN_SRC python
			def greet():
			print("hi")
			#+END_SRC
			`,
		},
		{
			caseName: "BulletWithMultipleCodeBlocks",
			orgContent: `* Header
			:PROPERTIES:
			:ID: 33333333
			:END:
			- [ ] Task
			#+BEGIN_SRC go
			package main
			#+END_SRC
			#+BEGIN_SRC bash
			echo hello
			#+END_SRC
			`,
			bulletUid:    "33333333.b0",
			codeBlockUid: "33333333.b0.c0",
			expectedRendered: `    #+BEGIN_SRC go
			package main
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

func TestRoundTripCodeBlockUnderBulletFullParse(t *testing.T) {
	tests := []struct {
		caseName             string
		orgContent           string
		wantBulletUid        string
		wantCodeBlockUid     string
		wantCodeBlockLang    string
		wantCodeBlockContent string
	}{
		{
			caseName: "ParseAndRenderBulletWithCodeBlock",
			orgContent: `* Header
			:PROPERTIES:
			:ID: 11111111
			:END:
			- [ ] Task
			#+BEGIN_SRC rust
			fn main() {}
			#+END_SRC
			`,
			wantBulletUid:        "11111111.b0",
			wantCodeBlockUid:     "11111111.b0.c0",
			wantCodeBlockLang:    "rust",
			wantCodeBlockContent: "fn main()",
		},
		{
			caseName: "ParseAndRenderBulletWithNamedCodeBlock",
			orgContent: `* Header
			:PROPERTIES:
			:ID: 22222222
			:END:
			- [x] Done
			#+NAME: my_func
			#+BEGIN_SRC python
			def my_func():
			pass
			#+END_SRC
			`,
			wantBulletUid:        "22222222.b0",
			wantCodeBlockUid:     "22222222.b0.my_func",
			wantCodeBlockLang:    "python",
			wantCodeBlockContent: "def my_func",
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

			bulletRender, ok := of.GetUid(NewUid(tt.wantBulletUid)).Split()
			if !ok {
				t.Fatalf("expected bullet %s to exist", tt.wantBulletUid)
			}

			bullet, ok := bulletRender.(*orgmcp.Bullet)
			if !ok {
				t.Fatalf("expected %s to be a Bullet, got %T", tt.wantBulletUid, bulletRender)
			}

			children := bullet.Children()
			if len(children) == 0 {
				t.Fatalf("expected bullet %s to have children", tt.wantBulletUid)
			}

			foundCodeBlock := false
			for _, child := range children {
				if cb, ok := child.(*codeblock.CodeBlock); ok {
					foundCodeBlock = true
					if lang := cb.Language().UnwrapOrDefault(); lang != tt.wantCodeBlockLang {
						t.Errorf("expected language %q, got %q", tt.wantCodeBlockLang, lang)
					}
					allLines := strings.Join(cb.Lines(), "\n")
					if !strings.Contains(allLines, tt.wantCodeBlockContent) {
						t.Errorf("expected content to contain %q, got %q", tt.wantCodeBlockContent, allLines)
					}
				}
			}

			if !foundCodeBlock {
				t.Errorf("expected to find code block child under bullet %s", tt.wantBulletUid)
			}

			builder := &strings.Builder{}
			of.Render(builder, -1)
			rendered := builder.String()

			of2, err := orgmcp.OrgFileFromReader(context.TODO(), strings.NewReader(rendered)).Split()
			if err != nil {
				t.Fatalf("failed to re-parse rendered org file: %v", err)
			}

			cbRender, ok := of2.GetUid(NewUid(tt.wantCodeBlockUid)).Split()
			if !ok {
				t.Fatalf("expected code block %s to exist after re-parsing", tt.wantCodeBlockUid)
			}

			cb, ok := cbRender.(*codeblock.CodeBlock)
			if !ok {
				t.Fatalf("expected %s to be a CodeBlock after re-parsing, got %T", tt.wantCodeBlockUid, cbRender)
			}

			if lang := cb.Language().UnwrapOrDefault(); lang != tt.wantCodeBlockLang {
				t.Errorf("after re-parse, expected language %q, got %q", tt.wantCodeBlockLang, lang)
			}
		})
	}
}

