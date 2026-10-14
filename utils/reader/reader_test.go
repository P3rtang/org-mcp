package reader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// newTestReader wraps a string in a bufio.Reader and a PeekReader.
// Returns the PeekReader so tests can drive it directly.
func newTestReader(input string) *PeekReader {
	return NewPeekReader(bufio.NewReader(strings.NewReader(input)))
}

// bytesEqual compares two byte slices for exact equality.
// Used instead of reflect.DeepEqual so test failures show
// the actual mismatched bytes.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// PeekBytes
// ---------------------------------------------------------------------------

// TestPeekBytesIncludesDelimiter covers the contract that PeekBytes returns
// the bytes UP TO AND INCLUDING the delimiter. This matches bufio.Reader.ReadBytes.
//
// The current implementation has an inconsistency: the first call (when the
// buffer is empty) returns the delimiter, but subsequent calls return only
// the content before the delimiter. After the fix, both branches should
// behave the same way: include the delimiter.
func TestPeekBytesIncludesDelimiter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{
			name:  "FirstCall",
			input: "first\n",
			want:  []byte("first\n"),
		},
		{
			name:  "FirstCallLongerLine",
			input: "hello world\n",
			want:  []byte("hello world\n"),
		},
		{
			name:  "FirstCallEmptyContent",
			input: "\n",
			want:  []byte("\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestReader(tt.input)
			got, err := r.PeekBytes('\n')
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytesEqual(got, tt.want) {
				t.Errorf("PeekBytes returned %q (len=%d), want %q (len=%d)",
					got, len(got), tt.want, len(tt.want))
			}
		})
	}
}

// TestPeekBytesConsistentAcrossCalls verifies that the second and subsequent
// PeekBytes calls return the same thing as the first. Today the first call
// returns the delimiter but the second does not — this is the inconsistency
// flagged in the bug report.
func TestPeekBytesConsistentAcrossCalls(t *testing.T) {
	input := "first\nsecond\n"
	r := newTestReader(input)

	first, err := r.PeekBytes('\n')
	if err != nil {
		t.Fatalf("first PeekBytes error: %v", err)
	}

	second, err := r.PeekBytes('\n')
	if err != nil {
		t.Fatalf("second PeekBytes error: %v", err)
	}

	if !bytesEqual(first, second) {
		t.Errorf("PeekBytes is inconsistent: first=%q, second=%q", first, second)
	}
	if !bytesEqual(first, []byte("first\n")) {
		t.Errorf("PeekBytes returned %q, want %q", first, "first\n")
	}
}

// TestPeekBytesReturnsSliceWithoutTrailingGarbage covers the case where the
// caller relies on the length of the returned slice matching the content.
// With the current bug, when the buffer is non-empty and the delimiter is
// present, the returned slice length matches len(peekBuffer) — which is fine
// if the peekBuffer is exactly the matched portion. After the fix this
// invariant should hold in all cases.
func TestPeekBytesReturnsExactLength(t *testing.T) {
	input := "abc\n"
	r := newTestReader(input)

	got, err := r.PeekBytes('\n')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 4 {
		t.Errorf("PeekBytes returned slice of length %d, want 4 (content %q)", len(got), got)
	}
	if string(got) != "abc\n" {
		t.Errorf("PeekBytes returned %q, want %q", got, "abc\n")
	}
}

// TestPeekBytesCachesForSubsequentCalls verifies that PeekBytes populates an
// internal buffer so multiple calls return the same data without re-reading
// from the underlying source.
func TestPeekBytesCachesForSubsequentCalls(t *testing.T) {
	input := "abc\ndef\nghi\n"
	r := newTestReader(input)

	first, _ := r.PeekBytes('\n')
	second, _ := r.PeekBytes('\n')
	third, _ := r.PeekBytes('\n')

	want := []byte("abc\n")
	if !bytesEqual(first, want) {
		t.Errorf("first peek: got %q, want %q", first, want)
	}
	if !bytesEqual(second, want) {
		t.Errorf("second peek: got %q, want %q", second, want)
	}
	if !bytesEqual(third, want) {
		t.Errorf("third peek: got %q, want %q", third, want)
	}
}

// TestPeekBytesAdvancesUnderlyingReader verifies that PeekBytes causes the
// underlying reader's read position to advance, so subsequent ReadBytes
// (after Discard) returns the next line, not the same one.
func TestPeekBytesAdvancesUnderlyingReader(t *testing.T) {
	input := "first\nsecond\n"
	r := newTestReader(input)

	_, _ = r.PeekBytes('\n')
	r.Discard()

	got, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(got) != "second\n" {
		t.Errorf("after Discard, ReadBytes returned %q, want %q", got, "second\n")
	}
}

// TestPeekBytesEOFWithoutDelimiter covers the case where the input has no
// trailing delimiter. The underlying reader returns io.EOF along with the
// partial bytes — PeekBytes should propagate that.
func TestPeekBytesEOFWithoutDelimiter(t *testing.T) {
	input := "no newline at end"
	r := newTestReader(input)

	got, err := r.PeekBytes('\n')
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got %v", err)
	}
	if string(got) != "no newline at end" {
		t.Errorf("PeekBytes at EOF returned %q, want %q", got, "no newline at end")
	}
}

// TestPeekBytesEmptyInput covers the case where the input is completely empty.
// PeekBytes should return io.EOF immediately.
func TestPeekBytesEmptyInput(t *testing.T) {
	r := newTestReader("")

	got, err := r.PeekBytes('\n')
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF for empty input, got err=%v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty bytes for empty input, got %q", got)
	}
}

// TestPeekBytesMultiLineSequence covers a realistic usage: peek-discard-peek
// across multiple lines, verifying the buffer correctly carries state and
// the underlying reader advances past peeked data after Discard.
func TestPeekBytesMultiLineSequence(t *testing.T) {
	input := "alpha\nbeta\ngamma\n"
	r := newTestReader(input)

	// Peek at first line, then discard it (move past it).
	peek1, _ := r.PeekBytes('\n')
	if string(peek1) != "alpha\n" {
		t.Errorf("peek 1: got %q, want %q", peek1, "alpha\n")
	}
	r.Discard()

	// Peek at second line, then discard it.
	peek2, _ := r.PeekBytes('\n')
	if string(peek2) != "beta\n" {
		t.Errorf("peek 2: got %q, want %q", peek2, "beta\n")
	}
	r.Discard()

	// Peek at third line.
	peek3, _ := r.PeekBytes('\n')
	if string(peek3) != "gamma\n" {
		t.Errorf("peek 3: got %q, want %q", peek3, "gamma\n")
	}
}

// ---------------------------------------------------------------------------
// ReadBytes
// ---------------------------------------------------------------------------

// TestReadBytesEmptyBuffer covers the simplest case: buffer is empty,
// ReadBytes reads from the underlying reader. Should return the line
// including the delimiter.
func TestReadBytesEmptyBuffer(t *testing.T) {
	input := "hello\n"
	r := newTestReader(input)

	got, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "hello\n" {
		t.Errorf("ReadBytes returned %q, want %q", got, "hello\n")
	}
}

// TestReadBytesEmptyInput covers the case where there's nothing to read.
// Should return io.EOF with empty bytes.
func TestReadBytesEmptyInput(t *testing.T) {
	r := newTestReader("")

	got, err := r.ReadBytes('\n')
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty bytes, got %q", got)
	}
}

// TestReadBytesWithoutTrailingNewline covers the case where the last line
// has no trailing delimiter. bufio returns the partial bytes with io.EOF.
func TestReadBytesWithoutTrailingNewline(t *testing.T) {
	input := "no newline"
	r := newTestReader(input)

	got, err := r.ReadBytes('\n')
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got %v", err)
	}
	if string(got) != "no newline" {
		t.Errorf("ReadBytes returned %q, want %q", got, "no newline")
	}
}

// TestReadBytesPopulatesFromPeek covers the pattern: PeekBytes, then ReadBytes
// should return the same data that was peeked.
func TestReadBytesPopulatesFromPeek(t *testing.T) {
	input := "alpha\nbeta\n"
	r := newTestReader(input)

	peeked, _ := r.PeekBytes('\n')
	if string(peeked) != "alpha\n" {
		t.Fatalf("setup: PeekBytes returned %q, want %q", peeked, "alpha\n")
	}

	read, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(read) != "alpha\n" {
		t.Errorf("ReadBytes returned %q, want %q", read, "alpha\n")
	}
}

// TestReadBytesNoTrailingGarbage covers the garbage-tail bug. If the internal
// buffer has bytes AFTER the matched delimiter, ReadBytes currently returns
// a slice with the same length as the buffer — the extra bytes are
// uninitialized (zero) values. This is a real bug because callers using
// len(result) or copying into a fixed-size buffer get wrong results.
//
// We construct this scenario by mixing PeekBytes with a different delimiter,
// then calling ReadBytes with the original delimiter.
func TestReadBytesNoTrailingGarbage(t *testing.T) {
	// "abc\ndef\nghi" — three lines, all terminated except the last.
	// If PeekBytes is called with a delimiter NOT in the buffer, the
	// implementation reads from the underlying and stores everything up to
	// that delimiter. Then a subsequent ReadBytes with the original
	// delimiter would return more bytes than expected.
	input := "abc\ndef\nghi"
	r := newTestReader(input)

	// Peek to populate the buffer with "abc\n" (4 bytes, ends with \n).
	_, _ = r.PeekBytes('\n')

	// Simulate a buffer that's longer than the matched portion by calling
	// PeekBytes with a different delimiter that's not in "abc\n". This
	// forces the second branch of PeekBytes, which appends to peekBuffer
	// and returns the full buffer length. In current code this means the
	// next call sees peekBuffer = "abc\ndef\nghi" (11 bytes).
	_, _ = r.PeekBytes('z') // not found in "abc\n", so reads rest into buffer

	// Now ReadBytes('\n') should return "abc\n" (4 bytes), NOT
	// "abc\n" + 7 zero bytes.
	got, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("ReadBytes returned slice of length %d, want 4 (content: %v)", len(got), got)
	}
	if string(got) != "abc\n" {
		t.Errorf("ReadBytes returned %q, want %q", got, "abc\n")
	}
}

// ---------------------------------------------------------------------------
// Peek(n)
// ---------------------------------------------------------------------------

// TestPeekEmptyBuffer covers the most critical Peek(n) bug. When the peek
// buffer is empty, the function currently returns n zero bytes silently
// without ever reading from the underlying reader. After the fix, it should
// populate the buffer from the underlying reader and return real bytes.
func TestPeekEmptyBuffer(t *testing.T) {
	input := "hello world"
	r := newTestReader(input)

	got, err := r.Peek(5)
	if err != nil {
		t.Fatalf("Peek(5) error: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("Peek(5) on empty buffer returned %q, want %q", got, "hello")
	}
	if len(got) != 5 {
		t.Errorf("Peek(5) returned slice of length %d, want 5", len(got))
	}
}

// TestPeekFullInput covers Peek(n) when n is larger than the available input.
// Should return whatever bytes are available, with io.EOF.
func TestPeekFullInput(t *testing.T) {
	input := "hi"
	r := newTestReader(input)

	got, err := r.Peek(10)
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got %v", err)
	}
	if string(got) != "hi" {
		t.Errorf("Peek(10) on 'hi' returned %q, want %q", got, "hi")
	}
}

// TestPeekExactLength covers Peek(n) when exactly n bytes are available.
func TestPeekExactLength(t *testing.T) {
	input := "hello"
	r := newTestReader(input)

	got, err := r.Peek(5)
	if err != nil {
		t.Fatalf("Peek(5) error: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("Peek(5) on 'hello' returned %q, want %q", got, "hello")
	}
}

// TestPeekThenPeekAgain covers the case where the buffer has been populated
// by a previous Peek(n) call. The second Peek should return the same data
// (because peek is non-destructive — data goes into the internal buffer).
func TestPeekThenPeekAgain(t *testing.T) {
	input := "hello world"
	r := newTestReader(input)

	first, _ := r.Peek(5)
	second, _ := r.Peek(5)

	if !bytesEqual(first, second) {
		t.Errorf("Peek not cached: first=%q, second=%q", first, second)
	}
	if string(first) != "hello" {
		t.Errorf("Peek(5) returned %q, want %q", first, "hello")
	}
}

// TestPeekPopulatesBufferForSubsequentRead covers a follow-up: after Peek(n),
// the data should be in the PeekReader's buffer so a subsequent
// PeekBytes/ReadBytes can see it.
func TestPeekPopulatesBufferForSubsequentRead(t *testing.T) {
	input := "hello\nworld\n"
	r := newTestReader(input)

	// Peek 8 bytes — should populate buffer with "hello\nwo" (8 bytes).
	peeked, err := r.Peek(8)
	if err != nil {
		t.Fatalf("Peek error: %v", err)
	}
	if string(peeked) != "hello\nwo" {
		t.Fatalf("Peek(8) returned %q, want %q", peeked, "hello\nwo")
	}

	// PeekBytes should now find the '\n' at index 5 in the buffer.
	got, err := r.PeekBytes('\n')
	if err != nil {
		t.Fatalf("PeekBytes error: %v", err)
	}
	if string(got) != "hello\n" {
		t.Errorf("PeekBytes after Peek returned %q, want %q", got, "hello\n")
	}
}

// TestPeekDoesNotAdvanceUnderlyingReader covers the contract that Peek is
// non-destructive — after Peek(n), the underlying reader should still have
// the data available for subsequent ReadBytes calls.
func TestPeekDoesNotAdvanceUnderlyingReader(t *testing.T) {
	input := "first line\nsecond line\n"
	r := newTestReader(input)

	_, _ = r.Peek(5)

	// After Peek(5), the first 5 bytes are in the PeekReader's buffer.
	// A subsequent ReadBytes should still return "first line\n" — meaning
	// the underlying reader's position should NOT have been advanced past
	// those 5 bytes.
	got, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(got) != "first line\n" {
		t.Errorf("after Peek(5), ReadBytes returned %q, want %q", got, "first line\n")
	}
}

// ---------------------------------------------------------------------------
// Discard (renamed from Continue)
// ---------------------------------------------------------------------------

// TestDiscardClearsBuffer verifies that Discard empties the internal peek
// buffer so the next read pulls fresh data from the underlying reader.
func TestDiscardClearsBuffer(t *testing.T) {
	input := "first\nsecond\n"
	r := newTestReader(input)

	_, _ = r.PeekBytes('\n')
	r.Discard()

	// After Discard, the next read should NOT re-return "first\n".
	// It should return "second\n" because Discard does not rewind the
	// underlying reader's position.
	got, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(got) != "second\n" {
		t.Errorf("after Discard, ReadBytes returned %q, want %q", got, "second\n")
	}
}

// TestDiscardOnEmptyBuffer verifies that calling Discard when no data has
// been peeked is safe.
func TestDiscardOnEmptyBuffer(t *testing.T) {
	r := newTestReader("anything\n")

	// No-op on empty buffer.
	r.Discard()
	r.Discard() // calling twice is also safe

	got, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(got) != "anything\n" {
		t.Errorf("ReadBytes returned %q, want %q", got, "anything\n")
	}
}

// TestDiscardThenPeekReadsNextLine covers the documented usage pattern:
// PeekBytes to look ahead, Discard to commit, PeekBytes again to see the
// next line.
func TestDiscardThenPeekReadsNextLine(t *testing.T) {
	input := "line1\nline2\nline3\n"
	r := newTestReader(input)

	peek1, _ := r.PeekBytes('\n')
	if string(peek1) != "line1\n" {
		t.Errorf("peek 1: got %q, want %q", peek1, "line1\n")
	}
	r.Discard()

	peek2, _ := r.PeekBytes('\n')
	if string(peek2) != "line2\n" {
		t.Errorf("peek 2: got %q, want %q", peek2, "line2\n")
	}
	r.Discard()

	peek3, _ := r.PeekBytes('\n')
	if string(peek3) != "line3\n" {
		t.Errorf("peek 3: got %q, want %q", peek3, "line3\n")
	}
}

// ---------------------------------------------------------------------------
// Integration: realistic line-by-line parsing
// ---------------------------------------------------------------------------

// TestRealisticParsingLoop simulates the pattern used in
// orgmcp/file.go:OrgFileFromReader: peek, check, discard, repeat.
func TestRealisticParsingLoop(t *testing.T) {
	input := "* Header 1\nbody line\n* Header 2\nanother body\n"
	r := newTestReader(input)

	var lines []string

	for {
		val, err := r.PeekBytes('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("PeekBytes error: %v", err)
		}

		lines = append(lines, string(val))
		r.Discard()
	}

	want := []string{
		"* Header 1\n",
		"body line\n",
		"* Header 2\n",
		"another body\n",
	}

	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d (lines: %v)", len(lines), len(want), lines)
	}
	for i, line := range lines {
		if line != want[i] {
			t.Errorf("line %d: got %q, want %q", i, line, want[i])
		}
	}
}

// TestPeekThenReadSequence covers the pattern where you peek to make a
// decision, then read with a different method.
func TestPeekThenReadSequence(t *testing.T) {
	input := "abc\ndef\n"
	r := newTestReader(input)

	// Peek to check what kind of line this is.
	peeked, _ := r.PeekBytes('\n')
	if !bytes.HasPrefix(peeked, []byte("abc")) {
		t.Fatalf("setup: unexpected peek result %q", peeked)
	}

	// Now consume via ReadBytes.
	read, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(read) != "abc\n" {
		t.Errorf("ReadBytes returned %q, want %q", read, "abc\n")
	}

	// And the next line should be "def\n".
	read2, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes 2 error: %v", err)
	}
	if string(read2) != "def\n" {
		t.Errorf("ReadBytes 2 returned %q, want %q", read2, "def\n")
	}
}
