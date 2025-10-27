package itertools

import (
	"iter"
	"testing"
)

// TestFlattenBasic tests the basic functionality of Flatten with integer slices
func TestFlattenBasic(t *testing.T) {
	slices := [][]int{
		{1, 2, 3},
		{4, 5},
		{6},
	}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	for val := range flattened {
		result = append(result, val)
	}

	expected := []int{1, 2, 3, 4, 5, 6}
	if !intSlicesEqual(result, expected) {
		t.Errorf("Flatten() = %v, want %v", result, expected)
	}
}

// TestFlattenEmpty tests Flatten with empty input
func TestFlattenEmpty(t *testing.T) {
	slices := [][]int{}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	for val := range flattened {
		result = append(result, val)
	}

	expected := []int{}
	if !intSlicesEqual(result, expected) {
		t.Errorf("Flatten() with empty input = %v, want %v", result, expected)
	}
}

// TestFlattenWithEmptySlices tests Flatten when some slices are empty
func TestFlattenWithEmptySlices(t *testing.T) {
	slices := [][]int{
		{1, 2},
		{},
		{3},
		{},
		{4, 5},
	}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	for val := range flattened {
		result = append(result, val)
	}

	expected := []int{1, 2, 3, 4, 5}
	if !intSlicesEqual(result, expected) {
		t.Errorf("Flatten() with empty slices = %v, want %v", result, expected)
	}
}

// TestFlattenSingleSlice tests Flatten with a single slice
func TestFlattenSingleSlice(t *testing.T) {
	slices := [][]int{{1, 2, 3, 4, 5}}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	for val := range flattened {
		result = append(result, val)
	}

	expected := []int{1, 2, 3, 4, 5}
	if !intSlicesEqual(result, expected) {
		t.Errorf("Flatten() with single slice = %v, want %v", result, expected)
	}
}

// TestFlattenWithStrings tests Flatten with string slices
func TestFlattenWithStrings(t *testing.T) {
	slices := [][]string{
		{"hello", "world"},
		{"foo"},
		{"bar", "baz"},
	}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []string{}
	for val := range flattened {
		result = append(result, val)
	}

	expected := []string{"hello", "world", "foo", "bar", "baz"}
	if !stringSlicesEqual(result, expected) {
		t.Errorf("Flatten() with strings = %v, want %v", result, expected)
	}
}

// TestFlattenEarlyBreak tests that Flatten respects early termination
func TestFlattenEarlyBreak(t *testing.T) {
	slices := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	count := 0
	for val := range flattened {
		result = append(result, val)
		count++
		if count == 5 {
			break
		}
	}

	expected := []int{1, 2, 3, 4, 5}
	if !intSlicesEqual(result, expected) {
		t.Errorf("Flatten() with early break = %v, want %v", result, expected)
	}
}

// TestFlattenAllEmpty tests Flatten when all slices are empty
func TestFlattenAllEmpty(t *testing.T) {
	slices := [][]int{
		{},
		{},
		{},
	}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	for val := range flattened {
		result = append(result, val)
	}

	expected := []int{}
	if !intSlicesEqual(result, expected) {
		t.Errorf("Flatten() with all empty slices = %v, want %v", result, expected)
	}
}

// TestFlattenLargeData tests Flatten with larger datasets
func TestFlattenLargeData(t *testing.T) {
	slices := make([][]int, 100)
	expectedCount := 0
	for i := range 100 {
		slices[i] = make([]int, i+1)
		for j := 0; j < i+1; j++ {
			slices[i][j] = i*100 + j
		}
		expectedCount += i + 1
	}

	seq := sliceSeq(slices)
	flattened := Flatten(seq)

	result := []int{}
	for val := range flattened {
		result = append(result, val)
	}

	if len(result) != expectedCount {
		t.Errorf("Flatten() with large data length = %d, want %d", len(result), expectedCount)
	}

	// Verify some values are in correct order
	if result[0] != 0 {
		t.Errorf("First element = %v, want 0", result[0])
	}
	if result[len(result)-1] != 9999 {
		t.Errorf("Last element = %v, want 9999", result[len(result)-1])
	}
}

// Helper functions

// sliceSeq converts a slice of slices into an iter.Seq
func sliceSeq[T any](slices [][]T) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		for _, slice := range slices {
			if !yield(slice) {
				return
			}
		}
	}
}

// intSlicesEqual compares two integer slices for equality
func intSlicesEqual(a, b []int) bool {
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

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
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
