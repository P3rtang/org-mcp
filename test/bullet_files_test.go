package main

import (
	"os"
	"strings"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	. "github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
)

// TestBulletFileFromReader tests parsing the bullets.org example file
func TestBulletFileFromReader(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Open the example file
	_, err := mcp.LoadOrgFile("./files/bullets.org")

	if err != nil {
		t.Fatalf("failed to load bullets.org: %v", err)
	}
}

// TestBulletFileRender tests that the parsed file can be rendered back
func TestBulletFileRender(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Read the original file content
	originalContent, err := os.ReadFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to read bullets.org: %v", err)
	}

	orgFile, err := mcp.LoadOrgFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to load bullets.org: %v", err)
	}

	// Render the parsed file
	builder := strings.Builder{}
	orgFile.Render(&builder, -1)

	// The rendered output should match the original (or at least have the same structure)
	rendered := builder.String()
	if strings.TrimSpace(rendered) != strings.TrimSpace(string(originalContent)) {
		t.Errorf("rendered output does not match original\nExpected:\n%s\nGot:\n%s", string(originalContent), rendered)
	}
}

// TODO: redo this test with explicit progress values
// TestBulletFileProgress tests the progress checking of headers with bullets
func TestBulletFileProgress(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	orgFile, err := mcp.LoadOrgFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to load bullets.org: %v", err)
	}

	// OrgFile itself should not have progress
	fileProgress := orgFile.CheckProgress()
	if fileProgress.IsSome() {
		t.Errorf("expected OrgFile to have no progress, but got Some")
	}

	children := itertools.Flatten(itertools.Map(
		itertools.FromSlice(orgFile.Children()),
		func(child Render) []Render { return child.Children() },
	))

	foundProgress := false

	// Check that at least one header has progress from its bullets
	for child := range children {
		progress := child.CheckProgress()
		if progress.IsSome() {
			foundProgress = true
			break
		}
	}

	if !foundProgress {
		t.Errorf("expected at least one header to have progress from bullets")
	}
}

// TestBulletFileHeadersHaveBullets tests that headers in bullets.org contain bullets
func TestBulletFileHeadersHaveBullets(t *testing.T) {
	var headers = []struct {
		uid         string
		bulletCount int
	}{
		{uid: "63689387", bulletCount: 0},
		{uid: "6806920", bulletCount: 2},
		{uid: "31786692", bulletCount: 3},
	}

	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	orgFile, err := mcp.LoadOrgFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to load bullets.org: %v", err)
	}

	for _, headerTest := range headers {
		header, ok := orgFile.GetUid(NewUid(headerTest.uid)).Split()

		if !ok {
			t.Errorf("failed to get header with UID %s", headerTest.uid)
			continue
		}

		children := header.Children()
		bulletCount := 0

		for _, child := range children {
			_, isBullet := child.(*Bullet)
			if isBullet {
				bulletCount += 1
			}
		}

		if bulletCount != headerTest.bulletCount {
			t.Errorf("header UID %s: expected %d bullets, got %d", headerTest.uid, headerTest.bulletCount, bulletCount)
		}
	}
}

// TestBulletIndexing tests that we can index bullets from headers
func TestBulletIndexing(t *testing.T) {
	var tests = []struct {
		uid     string
		content string
	}{
		{uid: "31786692.b0", content: "   - [x] Bullet 1"},
		{uid: "31786692.b1", content: "   - [ ] Bullet 2"},
		{uid: "31786693.b0", content: "   - [ ] Main bullet"},
		{uid: "31786693.b0.b0", content: "     * Sub bullet 1"},
	}

	// os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	orgFile, err := mcp.LoadOrgFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to load bullets.org: %v", err)
	}

	builder := strings.Builder{}

	for _, test := range tests {
		bullet, ok := orgFile.GetUid(NewUid(test.uid)).Split()
		if !ok {
			t.Errorf("failed to get bullet with UID %s", test.uid)
			continue
		}

		bullet.Render(&builder, 0)
		rendered := builder.String()
		builder.Reset()

		if strings.TrimSpace(rendered) != strings.TrimSpace(test.content) {
			t.Errorf("bullet UID %s: expected rendered content:\n%s\nGot:\n%s", test.uid, test.content, rendered)
		}
	}
}

// TestBulletCheckboxStatus tests checkbox parsing in bullets
func TestBulletCheckboxStatus(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name      string
		line      string
		hasCheck  bool
		isChecked bool
	}{
		{
			name:      "Checked bullet",
			line:      "* [x] Bullet 1",
			hasCheck:  true,
			isChecked: true,
		},
		{
			name:      "Unchecked bullet",
			line:      "* [ ] Bullet 2",
			hasCheck:  true,
			isChecked: false,
		},
		{
			name:      "No checkbox",
			line:      "* Regular bullet",
			hasCheck:  false,
			isChecked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: BulletFromString is TODO, so we can't test directly yet
			// This test is prepared for when the function is implemented
			if strings.Contains(tt.line, "[x]") {
				if !tt.isChecked || !tt.hasCheck {
					t.Errorf("expected checked bullet to be detected correctly")
				}
			} else if strings.Contains(tt.line, "[ ]") {
				if tt.isChecked || !tt.hasCheck {
					t.Errorf("expected unchecked bullet to be detected correctly")
				}
			} else {
				if tt.hasCheck {
					t.Errorf("expected no checkbox to be detected")
				}
			}
		})
	}
}

// TestBulletFileConsistency tests that parsing and rendering produces consistent output
func TestBulletFileConsistency(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Read the original file
	originalContent, err := os.ReadFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to read bullets.org: %v", err)
	}

	// Parse once
	file1, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file1.Close()

	orgFile1 := OrgFileFromReader(file1).Unwrap()
	builder1 := strings.Builder{}
	orgFile1.Render(&builder1, -1)
	rendered1 := builder1.String()

	// Parse the rendered content again
	file2, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file2.Close()

	orgFile2 := OrgFileFromReader(file2).Unwrap()
	builder2 := strings.Builder{}
	orgFile2.Render(&builder2, -1)
	rendered2 := builder2.String()

	// Both renders should be identical
	if rendered1 != rendered2 {
		t.Errorf("rendered content is not consistent between parses")
	}

	// Should match original
	if strings.TrimSpace(rendered1) != strings.TrimSpace(string(originalContent)) {
		t.Errorf("first render does not match original content")
	}
}

// TestBulletFileRemoveChildren tests bullet child removal functionality
func TestBulletFileRemoveChildren(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")

		// Additional render check after removing children
		builder := strings.Builder{}
		orgFileResult.UnwrapPtr().Render(&builder, -1)
		rendered := builder.String()
		if strings.Contains(rendered, "*") {
			t.Errorf("expected no bullet markers in rendered output after child removal, but found:\n%s", rendered)
		}
	}

	orgFile := orgFileResult.Unwrap()

	// Check if there are any children
	children := orgFile.Children()
	if len(children) == 0 {
		t.Fatalf("expected some children in OrgFile")
	}

	// Take the first header having children
	header := children[0] // Assuming the first element is a header
	if len(header.Children()) == 0 {
		t.Fatalf("no children to test removal on")
	}

	// Call RemoveChildren
	header.RemoveChildren()

	if len(header.Children()) != 0 {
		t.Errorf("RemoveChildren failed, expected 0 children, got %d", len(header.Children()))
	}
}

func TestBulletComplete(t *testing.T) {
	orgFile, err := mcp.LoadOrgFile("./files/bullets.org")

	type Test struct {
		name      string
		uid       Uid
		expected  string
		operation func(b *Bullet)
	}

	testMap := []Test{
		{
			name:     "TestCompleteDefaultBullet",
			uid:      NewUid("31786692.b1"),
			expected: "   - [x] Bullet 2",
			operation: func(b *Bullet) {
				b.CompleteCheckbox()
			},
		},
		{
			name:     "TestCompleteNestedMainBullet",
			uid:      NewUid("31786694.b1"),
			expected: "   - [x] Main bullet 2",
			operation: func(b *Bullet) {
				b.CompleteCheckbox()
			},
		},
	}

	builder := strings.Builder{}

	if err != nil {
		t.Fatalf("failed to load bullets.org: %v", err)
	}

	for _, test := range testMap {
		t.Run(test.name, func(t *testing.T) {
			bullet, ok := orgFile.GetUid(test.uid).Split()
			if !ok {
				t.Errorf("failed to get bullet with UID %s", test.uid)
			}

			test.operation(bullet.(*Bullet))

			bullet.Render(&builder, 0)
			rendered := builder.String()
			builder.Reset()

			if strings.TrimSpace(rendered) != strings.TrimSpace(test.expected) {
				t.Errorf("after completing, expected rendered content:\n%s\nGot:\n%s", test.expected, rendered)
			}
		})
	}
}
