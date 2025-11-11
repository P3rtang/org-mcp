package main

import (
	. "github.com/p3rtang/org-mcp/orgmcp"
	"os"
	"testing"
)

// TestHeaderCheckProgressWithStatus tests CheckProgress with headers that have status but no children
func TestHeaderCheckProgressWithStatus(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	cases := []struct {
		name           string
		orgContent     string
		expectedStatus HeaderStatus
		expectedDone   bool
	}{
		{
			name:           "TODO header with no progress returns progress with done=false",
			orgContent:     "* TODO Test header",
			expectedStatus: Todo,
			expectedDone:   false,
		},
		{
			name:           "DONE header with no progress returns progress with done=true",
			orgContent:     "* DONE Test header",
			expectedStatus: Done,
			expectedDone:   true,
		},
		{
			name:           "NEXT header with no progress returns progress with done=false",
			orgContent:     "* NEXT Test header",
			expectedStatus: Next,
			expectedDone:   false,
		},
		{
			name:           "PROG header with no progress returns progress with done=false",
			orgContent:     "* PROG Test header",
			expectedStatus: Prog,
			expectedDone:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHeaderFromString(tc.orgContent, nil).Unwrap()
			progress := h.CheckProgress()

			if progress.IsNone() {
				t.Fatalf("expected CheckProgress to return Some, but got None")
			}

			p := progress.Unwrap()
			if p.Done() != tc.expectedDone {
				t.Errorf("expected Done()=%v, got %v", tc.expectedDone, p.Done())
			}

			if h.Status != tc.expectedStatus {
				t.Errorf("expected status %v, got %v", tc.expectedStatus, h.Status)
			}
		})
	}
}

// TestHeaderCheckProgressWithNoStatus tests headers without status
func TestHeaderCheckProgressWithNoStatus(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	cases := []struct {
		name       string
		orgContent string
	}{
		{
			name:       "Header with no status and no progress returns None",
			orgContent: "* Test header",
		},
		{
			name:       "Header with tags but no status returns None",
			orgContent: "* Test header :tag1:tag2:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHeaderFromString(tc.orgContent, nil).Unwrap()
			progress := h.CheckProgress()

			if progress.IsSome() {
				t.Errorf("expected CheckProgress to return None for header without status, but got Some")
			}
		})
	}
}

// TestHeaderCheckProgressDoneStatus tests that DONE status is preserved
func TestHeaderCheckProgressDoneStatus(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	cases := []struct {
		name       string
		orgContent string
	}{
		{
			name:       "DONE header stays DONE",
			orgContent: "* DONE Test header",
		},
		{
			name:       "DONE header with tags stays DONE",
			orgContent: "* DONE Test header :tag1:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHeaderFromString(tc.orgContent, nil).Unwrap()
			originalStatus := h.Status
			h.CheckProgress()

			if h.Status != Done {
				t.Errorf("expected DONE header to stay DONE, but got %v", h.Status)
			}

			if h.Status != originalStatus {
				t.Errorf("expected status to remain unchanged, got %v (was %v)", h.Status, originalStatus)
			}
		})
	}
}

// TestHeaderCheckProgressBasic tests basic behavior of CheckProgress
func TestHeaderCheckProgressBasic(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	h := NewHeaderFromString("* TODO Test", nil).Unwrap()

	// Initial state: TODO
	if h.Status != Todo {
		t.Errorf("expected initial status TODO, got %v", h.Status)
	}

	// Call CheckProgress
	progress := h.CheckProgress()

	// Should return a progress
	if progress.IsNone() {
		t.Errorf("expected CheckProgress to return Some for header with status, got None")
	}

	// Status should remain TODO since there are no children
	if h.Status != Todo {
		t.Errorf("expected TODO status to remain after CheckProgress, got %v", h.Status)
	}
}

// TestHeaderCheckProgressReturnValues tests that CheckProgress returns the correct progress values
func TestHeaderCheckProgressReturnValues(t *testing.T) {
	os.Stderr, _ = os.OpenFile("/dev/null", os.O_WRONLY, 0644)

	cases := []struct {
		name       string
		orgContent string
		wantTotal  int
		wantDone   int
	}{
		{
			name:       "TODO header returns 0/0 progress (no children)",
			orgContent: "* TODO Test header",
			wantTotal:  0,
			wantDone:   0,
		},
		{
			name:       "DONE header returns 0/0 progress (no children)",
			orgContent: "* DONE Test header",
			wantTotal:  0,
			wantDone:   0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHeaderFromString(tc.orgContent, nil).Unwrap()
			progress := h.CheckProgress()

			if progress.IsNone() {
				t.Fatalf("expected progress to be Some, got None")
			}

			p := progress.Unwrap()
			if p.Total != tc.wantTotal {
				t.Errorf("expected total %d, got %d", tc.wantTotal, p.Total)
			}
			if p.Complete != tc.wantDone {
				t.Errorf("expected done %d, got %d", tc.wantDone, p.Complete)
			}
		})
	}
}
