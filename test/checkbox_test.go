package main

import (
	. "github.com/p3rtang/org-mcp/orgmcp"
	"os"
	"strings"
	"testing"
)

// TestBulletToggleCheckbox tests toggling checkbox state from Unchecked to Checked
func TestBulletToggleCheckbox(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name          string
		input         string
		expectedAfter string
	}{
		{
			name:          "Toggle from Unchecked to Checked",
			input:         "* [ ] Test bullet",
			expectedAfter: "[x]",
		},
		{
			name:          "Toggle from Checked to Unchecked",
			input:         "* [x] Test bullet",
			expectedAfter: "[ ]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the bullet from string
			bulletOpt := NewBulletFromString(tt.input, nil)
			if bulletOpt.IsNone() {
				t.Fatalf("expected BulletFromString to parse successfully")
			}

			bullet := bulletOpt.Unwrap()

			// Toggle the checkbox
			bullet.ToggleCheckbox()

			// Verify by rendering
			builder := strings.Builder{}
			bullet.Render(&builder, -1)
			rendered := builder.String()

			if !strings.Contains(rendered, tt.expectedAfter) {
				t.Errorf("expected '%s' to contain '%s', got: %s", rendered, tt.expectedAfter, rendered)
			}
		})
	}
}

// TestBulletCompleteCheckbox tests completing a checkbox (marking as Checked)
func TestBulletCompleteCheckbox(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name          string
		input         string
		expectedAfter string
	}{
		{
			name:          "Complete from Unchecked to Checked",
			input:         "* [ ] Test bullet",
			expectedAfter: "[x]",
		},
		{
			name:          "Complete when already Checked",
			input:         "* [x] Test bullet",
			expectedAfter: "[x]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the bullet from string
			bulletOpt := NewBulletFromString(tt.input, nil)
			if bulletOpt.IsNone() {
				t.Fatalf("expected BulletFromString to parse successfully")
			}

			bullet := bulletOpt.Unwrap()

			// Complete the checkbox
			bullet.CompleteCheckbox()

			// Verify by rendering
			builder := strings.Builder{}
			bullet.Render(&builder, -1)
			rendered := builder.String()

			if !strings.Contains(rendered, tt.expectedAfter) {
				t.Errorf("expected '%s' to contain '%s', got: %s", rendered, tt.expectedAfter, rendered)
			}
		})
	}
}

// TestBulletHasCheckbox tests checking if a bullet has a checkbox
func TestBulletHasCheckbox(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name           string
		input          string
		expectedResult bool
	}{
		{
			name:           "Has checkbox when Unchecked",
			input:          "* [ ] Test bullet",
			expectedResult: true,
		},
		{
			name:           "Has checkbox when Checked",
			input:          "* [x] Test bullet",
			expectedResult: true,
		},
		{
			name:           "No checkbox when NoCheck",
			input:          "* Regular bullet",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the bullet from string
			bulletOpt := NewBulletFromString(tt.input, nil)
			if bulletOpt.IsNone() {
				t.Fatalf("expected BulletFromString to parse successfully")
			}

			bullet := bulletOpt.Unwrap()

			result := bullet.HasCheckbox()

			if result != tt.expectedResult {
				t.Errorf("expected HasCheckbox() to return %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

// TestHeaderToggleCheckboxByIndex tests toggling checkbox in header's children by index
func TestHeaderToggleCheckboxByIndex(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name             string
		initialBullets   []string
		indexToToggle    int
		shouldSucceed    bool
		expectedContains string
	}{
		{
			name:             "Toggle first bullet from Unchecked to Checked",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2", "Bullet 3"},
			indexToToggle:    0,
			shouldSucceed:    true,
			expectedContains: "[x]",
		},
		{
			name:             "Toggle second bullet from Checked to Unchecked",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2", "Bullet 3"},
			indexToToggle:    1,
			shouldSucceed:    true,
			expectedContains: "[ ]",
		},
		{
			name:             "Cannot toggle bullet with no checkbox",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2", "Bullet 3"},
			indexToToggle:    2,
			shouldSucceed:    false,
			expectedContains: "",
		},
		{
			name:             "Out of range index should fail",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2"},
			indexToToggle:    5,
			shouldSucceed:    false,
			expectedContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create header with bullet children
			header := NewHeaderFromString("* Test header", nil).Unwrap()

			// Add bullet children
			for _, bulletStr := range tt.initialBullets {
				bulletOpt := NewBulletFromString("* "+bulletStr, &header)
				if bulletOpt.IsSome() {
					bullet := bulletOpt.Unwrap()
					header.AddChild(&bullet)
				}
			}

			// Toggle checkbox
			bullet, err := header.ToggleCheckboxByIndex(tt.indexToToggle)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("expected ToggleCheckboxByIndex() to succeed, got error: %v", err)
				}
				if bullet != nil {
					builder := strings.Builder{}
					bullet.Render(&builder, -1)
					rendered := builder.String()

					if !strings.Contains(rendered, tt.expectedContains) {
						t.Errorf("expected '%s' to contain '%s'", rendered, tt.expectedContains)
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected ToggleCheckboxByIndex() to fail, but it succeeded")
				}
			}
		})
	}
}

// TestHeaderCompleteCheckboxByIndex tests completing checkbox in header's children by index
func TestHeaderCompleteCheckboxByIndex(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name             string
		initialBullets   []string
		indexToComplete  int
		shouldSucceed    bool
		expectedContains string
	}{
		{
			name:             "Complete first bullet from Unchecked to Checked",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2", "Bullet 3"},
			indexToComplete:  0,
			shouldSucceed:    true,
			expectedContains: "[x]",
		},
		{
			name:             "Complete second bullet (already Checked)",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2", "Bullet 3"},
			indexToComplete:  1,
			shouldSucceed:    true,
			expectedContains: "[x]",
		},
		{
			name:             "Cannot complete bullet with no checkbox",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2", "Bullet 3"},
			indexToComplete:  2,
			shouldSucceed:    false,
			expectedContains: "",
		},
		{
			name:             "Out of range index should fail",
			initialBullets:   []string{"[ ] Bullet 1", "[x] Bullet 2"},
			indexToComplete:  5,
			shouldSucceed:    false,
			expectedContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create header with bullet children
			header := NewHeaderFromString("* Test header", nil).Unwrap()

			// Add bullet children
			for _, bulletStr := range tt.initialBullets {
				bulletOpt := NewBulletFromString("* "+bulletStr, &header)
				if bulletOpt.IsSome() {
					bullet := bulletOpt.Unwrap()
					header.AddChild(&bullet)
				}
			}

			// Complete checkbox
			bullet, err := header.CompleteCheckboxByIndex(tt.indexToComplete)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("expected CompleteCheckboxByIndex() to succeed, got error: %v", err)
				}
				if bullet != nil {
					builder := strings.Builder{}
					bullet.Render(&builder, -1)
					rendered := builder.String()

					if !strings.Contains(rendered, tt.expectedContains) {
						t.Errorf("expected '%s' to contain '%s'", rendered, tt.expectedContains)
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected CompleteCheckboxByIndex() to fail, but it succeeded")
				}
			}
		})
	}
}

// TestCheckboxToggleRenderConsistency tests that toggled checkboxes render correctly
func TestCheckboxToggleRenderConsistency(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Create a bullet with Unchecked state
	bulletOpt := NewBulletFromString("* [ ] Test bullet", nil)
	if bulletOpt.IsNone() {
		t.Fatalf("expected BulletFromString to parse successfully")
	}

	bullet := bulletOpt.Unwrap()

	// Render initial state
	builder1 := strings.Builder{}
	bullet.Render(&builder1, -1)
	uncheckedRender := builder1.String()

	// Verify it contains unchecked box
	if !strings.Contains(uncheckedRender, "[ ]") {
		t.Errorf("expected unchecked bullet to render with [ ], got: %s", uncheckedRender)
	}

	// Toggle to Checked
	bullet.ToggleCheckbox()

	// Render toggled state
	builder2 := strings.Builder{}
	bullet.Render(&builder2, -1)
	checkedRender := builder2.String()

	// Verify it contains checked box
	if !strings.Contains(checkedRender, "[x]") && !strings.Contains(checkedRender, "[X]") {
		t.Errorf("expected checked bullet to render with [x] or [X], got: %s", checkedRender)
	}

	// Verify they're different
	if uncheckedRender == checkedRender {
		t.Errorf("expected toggled bullet to render differently")
	}
}

// TestCheckboxCompleteRenderConsistency tests that completed checkboxes render correctly
func TestCheckboxCompleteRenderConsistency(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Create a bullet with Unchecked state
	bulletOpt := NewBulletFromString("* [ ] Test bullet", nil)
	if bulletOpt.IsNone() {
		t.Fatalf("expected BulletFromString to parse successfully")
	}

	bullet := bulletOpt.Unwrap()

	// Render initial state
	builder1 := strings.Builder{}
	bullet.Render(&builder1, -1)
	uncheckedRender := builder1.String()

	// Complete the checkbox
	bullet.CompleteCheckbox()

	// Render completed state
	builder2 := strings.Builder{}
	bullet.Render(&builder2, -1)
	completedRender := builder2.String()

	// Verify the completed render contains checked box
	if !strings.Contains(completedRender, "[x]") && !strings.Contains(completedRender, "[X]") {
		t.Errorf("expected completed bullet to render with [x] or [X], got: %s", completedRender)
	}

	// Verify they're different
	if uncheckedRender == completedRender {
		t.Errorf("expected completed bullet to render differently")
	}
}

// TestCheckboxInBulletFile tests checkbox operations on parsed bullet file
func TestCheckboxInBulletFile(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Open the bullets.org file
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()

	// Get the title header
	children := orgFile.Children()
	if len(children) == 0 {
		t.Fatalf("expected OrgFile to have headers")
	}

	titleHeader := children[0].(*Header)
	titleChildren := titleHeader.Children()

	// Find the first sub-header with unchecked checkboxes
	var headerWithCheckboxes *Header
	for _, child := range titleChildren {
		if header, ok := child.(*Header); ok {
			// Check if this header has bullet children with unchecked checkboxes
			for _, bulletChild := range header.Children() {
				if bullet, ok := bulletChild.(*Bullet); ok && bullet.HasCheckbox() {
					// Check if it's unchecked by rendering
					builder := strings.Builder{}
					bullet.Render(&builder, -1)
					if strings.Contains(builder.String(), "[ ]") {
						headerWithCheckboxes = header
						break
					}
				}
			}
			if headerWithCheckboxes != nil {
				break
			}
		}
	}

	if headerWithCheckboxes == nil {
		t.Fatalf("expected to find a header with unchecked bullets in bullets.org")
	}

	// Find the index of an unchecked bullet
	var uncheckedIndex int = -1
	for i, child := range headerWithCheckboxes.Children() {
		if bullet, ok := child.(*Bullet); ok {
			if !bullet.HasCheckbox() {
				continue
			}
			// Check if it's unchecked by rendering
			builder := strings.Builder{}
			bullet.Render(&builder, -1)
			if strings.Contains(builder.String(), "[ ]") {
				uncheckedIndex = i
				break
			}
		}
	}

	if uncheckedIndex == -1 {
		t.Fatalf("expected to find an unchecked bullet in the header")
	}

	// Toggle the checkbox
	toggledBullet, toggleErr := headerWithCheckboxes.ToggleCheckboxByIndex(uncheckedIndex)
	if toggleErr != nil {
		t.Errorf("expected ToggleCheckboxByIndex to succeed, got error: %v", toggleErr)
	}

	// Verify the state changed
	if toggledBullet != nil {
		builder := strings.Builder{}
		toggledBullet.Render(&builder, -1)
		if !strings.Contains(builder.String(), "[x]") {
			t.Errorf("expected checkbox to be checked after toggle")
		}
	}

	// Toggle back
	toggledBullet2, toggleErr2 := headerWithCheckboxes.ToggleCheckboxByIndex(uncheckedIndex)
	if toggleErr2 != nil {
		t.Errorf("expected ToggleCheckboxByIndex to succeed on second toggle, got error: %v", toggleErr2)
	}

	// Verify it's back to Unchecked
	if toggledBullet2 != nil {
		builder := strings.Builder{}
		toggledBullet2.Render(&builder, -1)
		if !strings.Contains(builder.String(), "[ ]") {
			t.Errorf("expected checkbox to be unchecked after second toggle")
		}
	}
}

// TestCheckboxCompleteInBulletFile tests completing checkboxes on parsed bullet file
func TestCheckboxCompleteInBulletFile(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	// Open the bullets.org file
	file, err := os.Open("./files/bullets.org")
	if err != nil {
		t.Fatalf("failed to open bullets.org: %v", err)
	}
	defer file.Close()

	orgFileResult := OrgFileFromReader(file)
	if !orgFileResult.IsOk() {
		t.Fatalf("expected OrgFileFromReader to return Ok, got Err")
	}

	orgFile := orgFileResult.Unwrap()

	// Get the title header
	children := orgFile.Children()
	if len(children) == 0 {
		t.Fatalf("expected OrgFile to have headers")
	}

	titleHeader := children[0].(*Header)
	titleChildren := titleHeader.Children()

	// Find the first sub-header with unchecked checkboxes
	var headerWithCheckboxes *Header
	for _, child := range titleChildren {
		if header, ok := child.(*Header); ok {
			// Check if this header has bullet children with unchecked checkboxes
			for _, bulletChild := range header.Children() {
				if bullet, ok := bulletChild.(*Bullet); ok && bullet.HasCheckbox() {
					// Check if it's unchecked by rendering
					builder := strings.Builder{}
					bullet.Render(&builder, -1)
					if strings.Contains(builder.String(), "[ ]") {
						headerWithCheckboxes = header
						break
					}
				}
			}
			if headerWithCheckboxes != nil {
				break
			}
		}
	}

	if headerWithCheckboxes == nil {
		t.Fatalf("expected to find a header with unchecked bullets in bullets.org")
	}

	// Find the index of an unchecked bullet
	var uncheckedIndex int = -1
	for i, child := range headerWithCheckboxes.Children() {
		if bullet, ok := child.(*Bullet); ok {
			if !bullet.HasCheckbox() {
				continue
			}
			// Check if it's unchecked by rendering
			builder := strings.Builder{}
			bullet.Render(&builder, -1)
			if strings.Contains(builder.String(), "[ ]") {
				uncheckedIndex = i
				break
			}
		}
	}

	if uncheckedIndex == -1 {
		t.Fatalf("expected to find an unchecked bullet in the header")
	}

	// Complete the checkbox
	completedBullet, completeErr := headerWithCheckboxes.CompleteCheckboxByIndex(uncheckedIndex)
	if completeErr != nil {
		t.Errorf("expected CompleteCheckboxByIndex to succeed, got error: %v", completeErr)
	}

	// Verify the state changed to Checked
	if completedBullet != nil {
		builder := strings.Builder{}
		completedBullet.Render(&builder, -1)
		if !strings.Contains(builder.String(), "[x]") {
			t.Errorf("expected checkbox to be checked after complete")
		}
	}

	// Complete again (should still be Checked)
	completedBullet2, completeErr2 := headerWithCheckboxes.CompleteCheckboxByIndex(uncheckedIndex)
	if completeErr2 != nil {
		t.Errorf("expected CompleteCheckboxByIndex to succeed on second complete, got error: %v", completeErr2)
	}

	if completedBullet2 != nil {
		builder := strings.Builder{}
		completedBullet2.Render(&builder, -1)
		if !strings.Contains(builder.String(), "[x]") {
			t.Errorf("expected checkbox to remain checked after second complete")
		}
	}
}

// TestCheckboxEdgeCases tests edge cases for checkbox operations
func TestCheckboxEdgeCases(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	t.Run("Empty header children list", func(t *testing.T) {
		header := NewHeaderFromString("* Empty header", nil).Unwrap()

		// Should fail gracefully
		_, err := header.ToggleCheckboxByIndex(0)
		if err == nil {
			t.Errorf("expected ToggleCheckboxByIndex to fail on empty children")
		}

		_, err = header.CompleteCheckboxByIndex(0)
		if err == nil {
			t.Errorf("expected CompleteCheckboxByIndex to fail on empty children")
		}
	})

	t.Run("Negative index", func(t *testing.T) {
		header := NewHeaderFromString("* Test header", nil).Unwrap()

		bulletOpt := NewBulletFromString("* [ ] Test bullet", &header)
		if bulletOpt.IsSome() {
			bullet := bulletOpt.Unwrap()
			header.AddChild(&bullet)
		}

		// Should fail on negative index
		_, err := header.ToggleCheckboxByIndex(-1)
		if err == nil {
			t.Errorf("expected ToggleCheckboxByIndex to fail on negative index")
		}

		_, err = header.CompleteCheckboxByIndex(-1)
		if err == nil {
			t.Errorf("expected CompleteCheckboxByIndex to fail on negative index")
		}
	})

	t.Run("Non-bullet child in header", func(t *testing.T) {
		header := NewHeaderFromString("* Test header", nil).Unwrap()

		// Add a sub-header instead of a bullet
		subheader := NewHeaderFromString("** Subheader", nil).Unwrap()
		header.AddChild(&subheader)

		// Should fail because child is not a bullet
		_, err := header.ToggleCheckboxByIndex(0)
		if err == nil {
			t.Errorf("expected ToggleCheckboxByIndex to fail on non-bullet child")
		}

		_, err = header.CompleteCheckboxByIndex(0)
		if err == nil {
			t.Errorf("expected CompleteCheckboxByIndex to fail on non-bullet child")
		}
	})
}
