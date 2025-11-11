package main

import (
	. "github.com/p3rtang/org-mcp/orgmcp"
	"os"
	"strings"
	"testing"
)

// TestManageBulletAdd tests adding bullets with and without checkboxes
func TestManageBulletAdd(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Create a test org file
	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	// Parse the test content
	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Test adding a bullet with checkbox
	bullet1 := NewBullet(header, Unchecked)
	bullet1.SetContent("Test bullet with checkbox")
	bullet1.SetIndex(0)
	header.AddChildren(&bullet1)

	// Verify bullet was added
	if len(header.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(header.Children()))
	}

	// Test adding another bullet with checked checkbox
	bullet2 := NewBullet(header, Checked)
	bullet2.SetContent("Completed bullet")
	bullet2.SetIndex(1)
	header.AddChildren(&bullet2)

	if len(header.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(header.Children()))
	}

	// Test adding a bullet without checkbox
	bullet3 := NewBullet(header, NoCheck)
	bullet3.SetContent("Regular bullet without checkbox")
	bullet3.SetIndex(2)
	header.AddChildren(&bullet3)

	if len(header.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(header.Children()))
	}

	// Verify the content is rendered correctly
	builder := strings.Builder{}
	header.Render(&builder, 1)
	rendered := builder.String()

	if !strings.Contains(rendered, "[ ] Test bullet with checkbox") {
		t.Errorf("expected unchecked bullet in rendered output")
	}

	if !strings.Contains(rendered, "[x] Completed bullet") {
		t.Errorf("expected checked bullet in rendered output")
	}

	if !strings.Contains(rendered, "* Regular bullet without checkbox") {
		t.Errorf("expected bullet without checkbox in rendered output")
	}
}

// TestManageBulletRemove tests removing bullets
func TestManageBulletRemove(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Add three bullets
	bullet1 := NewBullet(header, Unchecked)
	bullet1.SetContent("Bullet 1")
	bullet1.SetIndex(0)
	header.AddChildren(&bullet1)

	bullet2 := NewBullet(header, Checked)
	bullet2.SetContent("Bullet 2")
	bullet2.SetIndex(1)
	header.AddChildren(&bullet2)

	bullet3 := NewBullet(header, NoCheck)
	bullet3.SetContent("Bullet 3")
	bullet3.SetIndex(2)
	header.AddChildren(&bullet3)

	if len(header.Children()) != 3 {
		t.Fatalf("expected 3 bullets after adding")
	}

	// Remove the second bullet
	childrenBefore := len(header.Children())
	header.RemoveChildren(header.Children()[1].Uid())

	childrenAfter := len(header.Children())

	if childrenAfter >= childrenBefore {
		t.Errorf("expected bullet to be removed, had %d, now %d", childrenBefore, childrenAfter)
	}
}

// TestManageBulletComplete tests completing/checking bullets
func TestManageBulletComplete(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Add an unchecked bullet
	bullet := NewBullet(header, Unchecked)
	bullet.SetContent("Task to complete")
	bullet.SetIndex(0)
	header.AddChildren(&bullet)

	// Get the bullet and verify it has a checkbox
	bulletFromHeader := header.Children()[0].(*Bullet)
	if !bulletFromHeader.HasCheckbox() {
		t.Errorf("expected bullet to have checkbox")
	}

	// Complete the checkbox
	bulletFromHeader.CompleteCheckbox()

	// Verify rendering shows [x]
	builder := strings.Builder{}
	bulletFromHeader.Render(&builder, -1)
	rendered := builder.String()

	if !strings.Contains(rendered, "[x]") {
		t.Errorf("expected [x] in rendered output after completing")
	}
}

// TestManageBulletToggle tests toggling bullet checkbox state
func TestManageBulletToggle(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Add an unchecked bullet
	bullet := NewBullet(header, Unchecked)
	bullet.SetContent("Toggle test")
	bullet.SetIndex(0)
	header.AddChildren(&bullet)

	bulletFromHeader := header.Children()[0].(*Bullet)

	// Toggle from unchecked to checked
	bulletFromHeader.ToggleCheckbox()

	// Verify rendering shows [x] after toggle
	builder := strings.Builder{}
	bulletFromHeader.Render(&builder, -1)
	rendered := builder.String()

	if !strings.Contains(rendered, "[x]") {
		t.Errorf("expected [x] in rendered output after first toggle")
	}

	// Toggle from checked to unchecked
	bulletFromHeader.ToggleCheckbox()

	// Verify rendering shows [ ] after toggle
	builder.Reset()
	bulletFromHeader.Render(&builder, -1)
	rendered = builder.String()

	if !strings.Contains(rendered, "[ ]") {
		t.Errorf("expected [ ] in rendered output after second toggle")
	}
}

// TestManageBulletSetContent tests updating bullet content
func TestManageBulletSetContent(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Add a bullet
	bullet := NewBullet(header, Unchecked)
	bullet.SetContent("Original content")
	bullet.SetIndex(0)
	header.AddChildren(&bullet)

	bulletFromHeader := header.Children()[0].(*Bullet)

	// Update content
	bulletFromHeader.SetContent("Updated content")

	// Verify content was updated
	builder := strings.Builder{}
	bulletFromHeader.Render(&builder, -1)
	rendered := builder.String()

	if !strings.Contains(rendered, "Updated content") {
		t.Errorf("expected updated content in rendered output")
	}

	if strings.Contains(rendered, "Original content") {
		t.Errorf("did not expect original content in rendered output")
	}
}

// TestManageBulletSequenceOperations tests a sequence of bullet operations
func TestManageBulletSequenceOperations(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Add 3 bullets
	for i := range 3 {
		bullet := NewBullet(header, Unchecked)
		bullet.SetContent("Task " + string(rune(i+'1')))
		bullet.SetIndex(i)
		header.AddChildren(&bullet)
	}

	if len(header.Children()) != 3 {
		t.Fatalf("expected 3 bullets, got %d", len(header.Children()))
	}

	// Complete first bullet
	if b, ok := header.Children()[0].(*Bullet); ok {
		b.CompleteCheckbox()
	}

	// Update content of second bullet
	if b, ok := header.Children()[1].(*Bullet); ok {
		b.SetContent("Updated Task 2")
	}

	// Render and verify output
	builder := strings.Builder{}
	header.Render(&builder, 1)
	rendered := builder.String()

	if !strings.Contains(rendered, "[x] Task 1") {
		t.Errorf("expected completed bullet 1 in output")
	}

	if !strings.Contains(rendered, "Updated Task 2") {
		t.Errorf("expected updated task 2 in output")
	}

	if !strings.Contains(rendered, "[ ] Task 3") {
		t.Errorf("expected unchecked bullet 3 in output")
	}
}

// TestManageBulletInvalidOperations tests error handling
func TestManageBulletInvalidOperations(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	testContent := `* Test Header
:PROPERTIES:
:ID: 1
:END:
`

	orgFile := OrgFileFromReader(strings.NewReader(testContent)).Unwrap()
	headers := orgFile.GetHeaderByStatus(None)

	if len(headers) == 0 {
		t.Fatalf("expected to find at least one header")
	}

	header := headers[0]

	// Add a bullet without checkbox
	bullet := NewBullet(header, NoCheck)
	bullet.SetContent("No checkbox bullet")
	bullet.SetIndex(0)
	header.AddChildren(&bullet)

	bulletFromHeader := header.Children()[0].(*Bullet)

	// Try to toggle/complete a bullet without checkbox
	bulletFromHeader.ToggleCheckbox()
	bulletFromHeader.CompleteCheckbox()

	// Verify rendering shows * (no checkbox)
	builder := strings.Builder{}
	bulletFromHeader.Render(&builder, -1)
	rendered := builder.String()

	if !strings.Contains(rendered, "* No checkbox bullet") {
		t.Errorf("expected bullet without checkbox to remain unchanged")
	}
}
