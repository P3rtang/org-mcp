package test

import (
	"context"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/orgmcp/code-block"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	. "github.com/p3rtang/org-mcp/tools/test/utils"
	"github.com/p3rtang/org-mcp/utils/option"
)

const CODE_BLOCK_PARSE_TEST_PATH = "./codeblock_test_parse.org"

type ParseCodeBlockTest struct {
	name         string
	uid          string
	wantExists   bool
	wantLanguage string
}

func loadParseOrgFile(t *testing.T) orgmcp.OrgFile {
	t.Helper()

	of, err := mcp.LoadOrgFile(context.TODO(), CODE_BLOCK_PARSE_TEST_PATH)
	if err != nil {
		t.Fatal(err)
	}
	mcp.WriteOrgFileToDisk(context.TODO(), of, CODE_BLOCK_PARSE_TEST_PATH)

	return of
}

func fetchCodeBlock(t *testing.T, of *orgmcp.OrgFile, uidStr string) *codeblock.CodeBlock {
	t.Helper()

	render := of.GetUid(NewUid(uidStr))
	cb, ok := option.Cast[Render, *codeblock.CodeBlock](render).Split()
	if !ok {
		return nil
	}
	return cb
}

func TestParseCodeBlock(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of := loadParseOrgFile(t)

	tests := []ParseCodeBlockTest{
		{
			name:       "Header11111111Exists",
			uid:        "11111111",
			wantExists: true,
		},
		{
			name:       "Header22222222Exists",
			uid:        "22222222",
			wantExists: true,
		},
		{
			name:       "Header33333333Exists",
			uid:        "33333333",
			wantExists: true,
		},
		{
			name:       "Header44444444Exists",
			uid:        "44444444",
			wantExists: true,
		},
		{
			name:       "Header55555555Exists",
			uid:        "55555555",
			wantExists: true,
		},
		{
			name:       "Header66666666Exists",
			uid:        "66666666",
			wantExists: true,
		},
		{
			name:         "CodeBlock22222222c0IsJavascript",
			uid:          "22222222.c0",
			wantExists:   true,
			wantLanguage: "javascript",
		},
		{
			name:       "NamedCodeBlock44444444greetExists",
			uid:        "44444444.greet",
			wantExists: true,
		},
		{
			name:         "NamedCodeBlock44444444greetIsPython",
			uid:          "44444444.greet",
			wantExists:   true,
			wantLanguage: "python",
		},
		{
			name:       "FirstCodeBlockUnder55555555Exists",
			uid:        "55555555.c1",
			wantExists: true,
		},
		{
			name:       "SecondCodeBlockUnder55555555Exists",
			uid:        "55555555.c3",
			wantExists: true,
		},
		{
			name:       "UnnamedCodeBlockUnder66666666Exists",
			uid:        "66666666.c1",
			wantExists: true,
		},
		{
			name:         "UnnamedCodeBlock66666666c0IsPython",
			uid:          "66666666.c1",
			wantExists:   true,
			wantLanguage: "python",
		},
		{
			name:         "NamedRustBlockIsRust",
			uid:          "66666666.named_rust",
			wantExists:   true,
			wantLanguage: "rust",
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

			if tt.wantLanguage != "" {
				cb := fetchCodeBlock(t, &of, tt.uid)
				if cb == nil {
					t.Errorf("expected %q to be a CodeBlock, got %T", tt.uid, render)
					return
				}

				if got := cb.Language(); got.UnwrapOrDefault() != tt.wantLanguage {
					t.Errorf("expected language %q, got %q", tt.wantLanguage, got.UnwrapOr("Nothing"))
				}
			}
		})
	}
}

func TestParseCodeBlockNoContentLeakage(t *testing.T) {
	showDebug := os.Getenv("SHOW_DEBUG")
	if showDebug == "" {
		os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)
	}

	of := loadParseOrgFile(t)

	leakedSubstrings := []string{"Comment here", "add(a, b)"}

	for _, uidStr := range []string{
		"11111111",
		"22222222",
		"33333333",
		"44444444",
		"55555555",
		"66666666",
	} {
		t.Run(uidStr, func(t *testing.T) {
			render, ok := of.GetUid(NewUid(uidStr)).Split()
			if !ok {
				t.Fatal("expected header to exist")
			}

			header, ok := render.(*orgmcp.Header)
			if !ok {
				t.Fatal("expected header type")
			}

			content := header.Preview(200)
			for _, leaked := range leakedSubstrings {
				if ContainsString(content, leaked) {
					t.Errorf("header %s has content leaking from code block: %q", uidStr, content)
				}
			}
		})
	}
}
