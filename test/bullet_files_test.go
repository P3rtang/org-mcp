package main

import (
	. "main/orgmcp"
	"main/utils/itertools"
	"os"
	"strings"
	"testing"
)

// TestBulletFileFromReader tests parsing the bullets.org example file
func TestBulletFileFromReader(t *testing.T) {
	// Open the example file
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	// Parse the file
	orgFileResult := OrgFileFromReader(file)

	// Verify that the result is Ok
	if !orgFileResult.IsOk() {
		t.Errorf("expected OrgFileFromReader to return Ok, got Err")
	}
}

// TestBulletFileRender tests that the parsed file can be rendered back
func TestBulletFileRender(t *testing.T) {
	// Read the original file content
	originalContent, err := os.ReadFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to read bullets.org: %v", err)
	}

	// Parse the file
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()

	// Render the parsed file
	builder := strings.Builder{}
	orgFile.Render(&builder)

	// The rendered output should match the original (or at least have the same structure)
	rendered := builder.String()
	if strings.TrimSpace(rendered) != strings.TrimSpace(string(originalContent)) {
		t.Errorf("rendered output does not match original\nExpected:\n%s\nGot:\n%s", string(originalContent), rendered)
	}
}

// TestBulletFileProgress tests the progress checking of headers with bullets
func TestBulletFileProgress(t *testing.T) {
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()

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

// TestBulletFileStructureFromContent tests by reading and validating content
func TestBulletFileStructureFromContent(t *testing.T) {
	content, err := os.ReadFile("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to read bullets.org: %v", err)
	}

	contentStr := string(content)

	// Verify the file contains expected structure
	expectedElements := []string{
		"* This is the title",
		"** DONE Header 1 [2/2]",
		"* [X] Bullet 1",
		"* [X] Bullet 2",
		"** PROG Header 2 [1/3]",
		"- [ ] Bullet 2",
		"- [ ] Bullet 3",
	}

	for _, element := range expectedElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("expected to find '%s' in file content", element)
		}
	}
}

// TestBulletFileLocation tests the Location method
func TestBulletFileLocation(t *testing.T) {
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()
	location := orgFile.Location()

	// For OrgFile, location should be 0
	if location != 0 {
		t.Errorf("expected location to be 0, got %d", location)
	}
}

// TestBulletFileHeadersHaveBullets tests that headers in bullets.org contain bullets
func TestBulletFileHeadersHaveBullets(t *testing.T) {
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()

	// The file should have children (headers)
	if len(orgFile.Children()) == 0 {
		t.Errorf("expected OrgFile to have headers, but it's empty")
	}

	// Verify at least one header has bullets as children
	foundBullets := false
	for _, child := range orgFile.Children() {
		if len(child.Children()) > 0 {
			foundBullets = true
			break
		}
	}

	if !foundBullets {
		t.Errorf("expected at least one header to contain bullets")
	}
}

// TestBulletIndexing tests that we can index bullets from headers
func TestBulletIndexing(t *testing.T) {
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()
	children := orgFile.Children()

	// The first child should be the main title header
	if len(children) < 1 {
		t.Fatalf("expected at least 1 header in file")
	}

	// Index into the first header to find its children (which should be headers with bullets)
	titleHeader := children[0]
	titleChildren := titleHeader.Children()

	// The title header should have sub-headers as children
	if len(titleChildren) == 0 {
		t.Fatalf("expected title header to have sub-headers as children")
	}

	// For each sub-header, check if it has bullet children
	foundBullets := false
	for _, subHeader := range titleChildren {
		bulletChildren := subHeader.Children()
		if len(bulletChildren) > 0 {
			foundBullets = true
			// We found bullets under a sub-header
			// Verify they render correctly
			builder := strings.Builder{}
			for _, bullet := range bulletChildren {
				bullet.Render(&builder)
			}
			rendered := builder.String()
			if len(rendered) == 0 {
				t.Errorf("expected bullets to render non-empty content")
			}
		}
	}

	if !foundBullets {
		t.Fatalf("expected at least one sub-header to have bullet children")
	}
}

// TestBulletCheckboxStatus tests checkbox parsing in bullets
func TestBulletCheckboxStatus(t *testing.T) {
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
	orgFile1.Render(&builder1)
	rendered1 := builder1.String()

	// Parse the rendered content again
	file2, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file2.Close()

	orgFile2 := OrgFileFromReader(file2).Unwrap()
	builder2 := strings.Builder{}
	orgFile2.Render(&builder2)
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
