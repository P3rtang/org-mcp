package main

import (
	. "main/orgmcp"
	"os"
	"strings"
	"testing"
)

func TestHeaderFromString(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	input := "* TODO Example [0/1] :tag1:tag2:"
	expectedLevel := 0
	expectedStatus := Todo
	expectedContent := "Example"

	header := NewHeaderFromString(input, nil).Unwrap()

	if header.Level != expectedLevel {
		t.Errorf("expected level %d, got %d", expectedLevel, header.Level)
	}

	if header.Status != expectedStatus {
		t.Errorf("expected status %v, got %v", expectedStatus, header.Status)
	}

	if header.Content != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, header.Content)
	}

	builder := strings.Builder{}
	header.Render(&builder, -1)

	if strings.TrimSpace(builder.String()) != input {
		t.Errorf("expected rendered output '%s', got '%s'", input, builder.String())
	}
}

func TestHeaderProgress(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name               string
		input              string
		expectedTotal      int
		expectedDone       int
		shouldHaveProgress bool
	}{
		{
			name:               "Progress with [2/2]",
			input:              "* Example [2/2]",
			expectedTotal:      2,
			expectedDone:       2,
			shouldHaveProgress: true,
		},
		{
			name:               "Progress with [0/3]",
			input:              "* Example [0/3]",
			expectedTotal:      3,
			expectedDone:       0,
			shouldHaveProgress: true,
		},
		{
			name:               "DONE status without progress",
			input:              "* DONE Example",
			expectedTotal:      0,
			expectedDone:       0,
			shouldHaveProgress: false,
		},
		{
			name:               "TODO status without progress",
			input:              "* TODO Example",
			expectedTotal:      0,
			expectedDone:       0,
			shouldHaveProgress: false,
		},
		{
			name:               "No progress or status",
			input:              "* Example",
			expectedTotal:      0,
			expectedDone:       0,
			shouldHaveProgress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := NewHeaderFromString(tt.input, nil).Unwrap()

			if tt.shouldHaveProgress {
				if header.Progress.IsNone() {
					t.Errorf("expected progress to be Some, but got None")
				}

				progress := header.Progress.Unwrap()
				if progress.Total != tt.expectedTotal {
					t.Errorf("expected total %d, got %d", tt.expectedTotal, progress.Total)
				}

				if progress.Complete != tt.expectedDone {
					t.Errorf("expected done %d, got %d", tt.expectedDone, progress.Complete)
				}
			} else {
				if header.Progress.IsSome() {
					t.Errorf("expected progress to be None, but got Some")
				}
			}
		})
	}
}

func TestHeaderProgressCheckProgress(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	tests := []struct {
		name                 string
		input                string
		shouldReturnProgress bool
		expectedDone         bool
	}{
		{
			name:                 "DONE status returns progress when CheckProgress called",
			input:                "* DONE Example",
			shouldReturnProgress: true,
			expectedDone:         true,
		},
		{
			name:                 "TODO status returns progress when CheckProgress called",
			input:                "* TODO Example",
			shouldReturnProgress: true,
			expectedDone:         false,
		},
		{
			name:                 "Progress with [2/2] returns progress when CheckProgress called",
			input:                "* Example [2/2]",
			shouldReturnProgress: true,
			expectedDone:         true,
		},
		{
			name:                 "No status and no progress returns None when CheckProgress called",
			input:                "* Example",
			shouldReturnProgress: false,
			expectedDone:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := NewHeaderFromString(tt.input, nil).Unwrap()
			progress := header.CheckProgress()

			if tt.shouldReturnProgress {
				if progress.IsNone() {
					t.Errorf("expected CheckProgress to return Some, but got None")
				}

				progressValue := progress.Unwrap()
				if progressValue.Done() != tt.expectedDone {
					t.Errorf("expected Done() to return %v, got %v", tt.expectedDone, progressValue.Done())
				}
			} else {
				if progress.IsSome() {
					t.Errorf("expected CheckProgress to return None, but got Some")
				}
			}
		})
	}
}
