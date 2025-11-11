package main

import (
	"bufio"
	. "main/orgmcp"
	"main/utils/reader"
	"os"
	"strings"
	"testing"
)

// TestPlainTextParsing tests basic plain text parsing from a reader
func TestPlainTextParsing(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple plain text",
			input: "This is plain text\n",
			want:  "This is plain text\n",
		},
		{
			name:  "Plain text with leading spaces",
			input: "   Indented plain text\n",
			want:  "   Indented plain text\n",
		},
		{
			name:  "Plain text with trailing spaces",
			input: "Plain text with trailing spaces   \n",
			want:  "Plain text with trailing spaces\n",
		},
		{
			name:  "Plain text with multiple words",
			input: "This is a longer piece of plain text\n",
			want:  "This is a longer piece of plain text\n",
		},
		{
			name:  "Empty line",
			input: "\n",
			want:  "\n",
		},
		{
			name:  "Plain text with special characters",
			input: "Plain text with @#$%^&*() special characters\n",
			want:  "Plain text with @#$%^&*() special characters\n",
		},
		{
			name:  "Plain text with numbers",
			input: "This contains 12345 numbers\n",
			want:  "This contains 12345 numbers\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.input))
			peekReader := reader.NewPeekReader(r)
			plainText := NewPlainTextFromReader(peekReader)

			if plainText.IsNone() {
				t.Errorf("expected Some, got None")
				return
			}

			pt := plainText.Unwrap()
			// Access the content through Render to verify it was parsed
			builder := strings.Builder{}
			pt.Render(&builder, -1)
			got := builder.String()

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPlainTextFromString tests parsing plain text from a string
func TestPlainTextFromString(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testCases := []string{
		"This is a test",
		"Another line of text",
		"Text with numbers 123",
		"Text with symbols !@#$%^&*()",
	}

	for _, tc := range testCases {
		input := tc + "\n"
		r := bufio.NewReader(strings.NewReader(input))
		peekReader := reader.NewPeekReader(r)
		plainText := NewPlainTextFromReader(peekReader)

		if plainText.IsNone() {
			t.Errorf("failed to parse plain text: %s", tc)
		}
	}
}

// TestPlainTextCheckProgress tests that PlainText has no progress
func TestPlainTextCheckProgress(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	progress := plainText.CheckProgress()
	if progress.IsSome() {
		t.Errorf("expected plain text to have no progress, but got Some")
	}
}

// TestPlainTextIndentLevel tests the indent level of plain text
func TestPlainTextIndentLevel(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	indent := plainText.IndentLevel()
	if indent != 0 {
		t.Errorf("expected indent level 0, got %d", indent)
	}
}

// TestPlainTextChildren tests that PlainText has no children
func TestPlainTextChildren(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	children := plainText.Children()
	if len(children) != 0 {
		t.Errorf("expected no children, got %d", len(children))
	}

	childrenRec := plainText.ChildrenRec()
	if len(childrenRec) != 0 {
		t.Errorf("expected no recursive children, got %d", len(childrenRec))
	}
}

// TestPlainTextAddChildren tests that PlainText cannot have children added
func TestPlainTextAddChildren(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	// Try to add children
	err := plainText.AddChildren()
	if err == nil {
		t.Errorf("expected error when adding children to plain text, got nil")
	}
}

// TestPlainTextRemoveChildren tests that PlainText cannot have children removed
func TestPlainTextRemoveChildren(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	// Try to remove children
	err := plainText.RemoveChildren()
	if err == nil {
		t.Errorf("expected error when removing children from plain text, got nil")
	}
}

// TestPlainTextWithWhitespace tests plain text with various whitespace patterns
func TestPlainTextWithWhitespace(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Text with tabs",
			input: "Text\twith\ttabs\n",
			want:  "Text\twith\ttabs\n",
		},
		{
			name:  "Text with multiple spaces",
			input: "Text   with   multiple   spaces\n",
			want:  "Text   with   multiple   spaces\n",
		},
		{
			name:  "Leading and trailing whitespace",
			input: "   text with whitespace   \n",
			want:  "   text with whitespace\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.input))
			peekReader := reader.NewPeekReader(r)
			plainText := NewPlainTextFromReader(peekReader)

			if plainText.IsNone() {
				t.Errorf("expected Some, got None")
				return
			}

			pt := plainText.Unwrap()
			builder := strings.Builder{}
			pt.Render(&builder, -1)
			got := builder.String()

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPlainTextMultipleLines tests parsing multiple plain text lines
func TestPlainTextMultipleLines(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "Line 1\nLine 2\nLine 3\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)

	// Parse first line
	pt1 := NewPlainTextFromReader(peekReader)
	if pt1.IsNone() {
		t.Errorf("failed to parse first line")
		return
	}

	builder1 := strings.Builder{}
	pt1.Unwrap().Render(&builder1, -1)
	if builder1.String() != "Line 1\n" {
		t.Errorf("got %q, want %q", builder1.String(), "Line 1")
	}

	// Parse second line
	pt2 := NewPlainTextFromReader(peekReader)
	if pt2.IsNone() {
		t.Errorf("failed to parse second line")
		return
	}

	builder2 := strings.Builder{}
	pt2.Unwrap().Render(&builder2, -1)
	if builder2.String() != "Line 2\n" {
		t.Errorf("got %q, want %q", builder2.String(), "Line 2")
	}

	// Parse third line
	pt3 := NewPlainTextFromReader(peekReader)
	if pt3.IsNone() {
		t.Errorf("failed to parse third line")
		return
	}

	builder3 := strings.Builder{}
	pt3.Unwrap().Render(&builder3, -1)
	if builder3.String() != "Line 3\n" {
		t.Errorf("got %q, want %q", builder3.String(), "Line 3")
	}
}

// TestPlainTextLongContent tests parsing long plain text content
func TestPlainTextLongContent(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	longText := "This is a very long piece of plain text that contains multiple words and should still be parsed correctly as plain text without any special formatting or structure\n"
	input := longText + "\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader)

	if plainText.IsNone() {
		t.Errorf("expected Some, got None")
		return
	}

	builder := strings.Builder{}
	plainText.Unwrap().Render(&builder, -1)
	got := builder.String()

	if got != longText {
		t.Errorf("got %q, want %q", got, longText)
	}
}

// TestPlainTextUid tests that plain text has a proper UID
func TestPlainTextUid(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	uid := plainText.Uid()
	// PlainText without a parent should return -1
	if uid != -1 {
		t.Errorf("expected UID -1 for plain text without parent, got %d", uid)
	}
}

// TestPlainTextParentUid tests that plain text can have a parent UID
func TestPlainTextParentUid(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "This is plain text\n"
	r := bufio.NewReader(strings.NewReader(input))
	peekReader := reader.NewPeekReader(r)
	plainText := NewPlainTextFromReader(peekReader).Unwrap()

	parentUid := plainText.ParentUid()
	// PlainText without a parent should return 0
	if parentUid != 0 {
		t.Errorf("expected parent UID 0 for plain text without parent, got %d", parentUid)
	}
}
