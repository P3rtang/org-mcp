package main

import (
	. "main/orgmcp"
	"os"
	"strings"
	"testing"
)

// TestSimpleFileFromReader tests parsing the simple.org example file
func TestSimpleFileFromReader(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Open the example file
	file, err := os.Open("./files/simple.org")
	if err != nil {
		t.Fatalf("failed to open simple.org: %v", err)
	}
	defer file.Close()

	// Parse the file
	orgFileResult := OrgFileFromReader(file)

	// Verify that the result is Ok
	if !orgFileResult.IsOk() {
		t.Errorf("expected OrgFileFromReader to return Ok, got Err")
	}
}

// TestSimpleFileRender tests that the parsed file can be rendered back
func TestSimpleFileRender(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Read the original file content
	originalContent, err := os.ReadFile("./files/simple.org")
	if err != nil {
		t.Fatalf("failed to read simple.org: %v", err)
	}

	// Parse the file
	file, err := os.Open("./files/simple.org")
	if err != nil {
		t.Fatalf("failed to open simple.org: %v", err)
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
	orgFile.Render(&builder, -1)

	// The rendered output should match the original (or at least have the same structure)
	rendered := builder.String()
	if strings.TrimSpace(rendered) != strings.TrimSpace(string(originalContent)) {
		t.Errorf("rendered output does not match original\nExpected:\n%s\nGot:\n%s", string(originalContent), rendered)
	}
}

// TestSimpleFileProgress tests the progress checking of the parsed file
func TestSimpleFileProgress(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	file, err := os.Open("./files/simple.org")
	if err != nil {
		t.Fatalf("failed to open simple.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)

	// Check if parsing was successful
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()

	// Check progress of the file
	progress := orgFile.CheckProgress()

	// For a file without specific progress tracking, it should return None
	if progress.IsSome() {
		t.Errorf("expected progress to be None for OrgFile, got Some")
	}
}

// TestSimpleFileStructureFromContent tests by reading and validating content
func TestSimpleFileStructureFromContent(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	content, err := os.ReadFile("./files/simple.org")
	if err != nil {
		t.Fatalf("failed to read simple.org: %v", err)
	}

	contentStr := string(content)

	// Verify the file contains expected structure
	expectedElements := []string{
		"* This is a test Title :title:test:",
		"** Header 1 [1/3] :header:",
		"*** TODO test this",
		"*** DONE completed",
		"*** TODO test 2",
	}

	for _, element := range expectedElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("expected to find '%s' in file content", element)
		}
	}
}

// TestSimpleFileLocation tests the Location method
func TestSimpleFileLocation(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	file, err := os.Open("./files/simple.org")
	if err != nil {
		t.Fatalf("failed to open simple.org: %v", err)
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

// TestSimpleFileGetTag tests the GetTag method for retrieving headers by tag
func TestSimpleFileGetTag(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	file, _ := os.Open("./files/simple.org")
	defer file.Close()

	orgFile := OrgFileFromReader(file).Unwrap()

	// Test getting a header by tag "title"
	titleHeader := orgFile.GetTag("title")
	if titleHeader.IsNone() {
		t.Errorf("expected to find header with tag 'title', got None")
	}

	// Test getting a header by tag "header"
	headerTag := orgFile.GetTag("header")
	if headerTag.IsNone() {
		t.Errorf("expected to find header with tag 'header', got None")
	}

	// Test getting a header by non-existent tag
	nonExistentTag := orgFile.GetTag("nonexistent")
	if nonExistentTag.IsSome() {
		t.Errorf("expected to find no header with tag 'nonexistent', got Some")
	}
}

// TODO: rewrite with the GetUid() method
// // TestSimpleFileGetLine tests the GetLine method for retrieving content by line number
// func TestSimpleFileGetLine(t *testing.T) {
// 	file, _ := os.Open("./files/simple.org")
// 	defer file.Close()
//
// 	orgFile := OrgFileFromReader(file).Unwrap()
//
// 	// Test getting line 0 (should be the OrgFile itself)
// 	line0 := orgFile.GetLine(0)
// 	if line0.IsNone() {
// 		t.Errorf("expected to get content for line 0, got nil")
// 	}
//
// 	// Test getting line 1 (should be the first header)
// 	line1 := orgFile.GetLine(1)
// 	if line1.IsNone() {
// 		t.Errorf("expected to get content for line 1, got nil")
// 	}
//
// 	// Test getting line 2 (should be a child header or content)
// 	line2 := orgFile.GetLine(2)
// 	if line2.IsNone() {
// 		t.Errorf("expected to get content for line 2, got nil")
// 	}
//
// 	// Test getting a line that may not exist
// 	line100 := orgFile.GetLine(100)
// 	if line100.IsSome() {
// 		t.Errorf("expected to get nil for line 100 (non-existent), got non-nil")
// 	}
// }
