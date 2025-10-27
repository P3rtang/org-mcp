package option

import (
	"testing"
)

// Helper function for panic testing
func expectPanic(t *testing.T, fn func()) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic but none occurred")
		}
	}()
	fn()
}

// ============================================================================
// CONSTRUCTOR TESTS
// ============================================================================

// TestNone tests the None constructor
func TestNone(t *testing.T) {
	option := None[int]()
	if option.IsSome() {
		t.Error("None() should create an empty option")
	}
	if !option.IsNone() {
		t.Error("None() should be marked as None")
	}
}

// TestSome tests the Some constructor
func TestSome(t *testing.T) {
	value := 42
	option := Some(value)
	if option.IsNone() {
		t.Error("Some() should create an option with a value")
	}
	if !option.IsSome() {
		t.Error("Some() should be marked as Some")
	}
	// Verify the value is stored correctly using Unwrap() instead of direct field access
	if option.Unwrap() != value {
		t.Errorf("Some() should store the value correctly, got %d, want %d", option.Unwrap(), value)
	}
}

// ============================================================================
// STATE CHECK TESTS
// ============================================================================

// TestIsNone tests the IsNone method
func TestIsNone(t *testing.T) {
	tests := []struct {
		name     string
		option   Option[string]
		wantNone bool
	}{
		{"None returns true", None[string](), true},
		{"Some returns false", Some("hello"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.IsNone(); got != tt.wantNone {
				t.Errorf("IsNone() = %v, want %v", got, tt.wantNone)
			}
		})
	}
}

// TestIsSome tests the IsSome method
func TestIsSome(t *testing.T) {
	tests := []struct {
		name     string
		option   Option[int]
		wantSome bool
	}{
		{"None returns false", None[int](), false},
		{"Some returns true", Some(123), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.IsSome(); got != tt.wantSome {
				t.Errorf("IsSome() = %v, want %v", got, tt.wantSome)
			}
		})
	}
}

// ============================================================================
// VALUE EXTRACTION TESTS
// ============================================================================

// TestUnwrap tests the Unwrap method
func TestUnwrap(t *testing.T) {
	tests := []struct {
		name        string
		option      Option[int]
		wantValue   int
		shouldPanic bool
	}{
		{"Unwrap Some returns value", Some(99), 99, false},
		{"Unwrap None panics", None[int](), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				expectPanic(t, func() {
					_ = tt.option.Unwrap()
				})
			} else {
				got := tt.option.Unwrap()
				if got != tt.wantValue {
					t.Errorf("Unwrap() = %v, want %v", got, tt.wantValue)
				}
			}
		})
	}
}

// TestUnwrapPtr tests the UnwrapPtr method
func TestUnwrapPtr(t *testing.T) {
	tests := []struct {
		name        string
		option      Option[int]
		shouldPanic bool
		wantValue   int
	}{
		{
			name:        "UnwrapPtr Some returns pointer",
			option:      Some(42),
			shouldPanic: false,
			wantValue:   42,
		},
		{
			name:        "UnwrapPtr None panics",
			option:      None[int](),
			shouldPanic: true,
			wantValue:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				expectPanic(t, func() {
					_ = tt.option.UnwrapPtr()
				})
			} else {
				ptr := tt.option.UnwrapPtr()
				if ptr == nil {
					t.Error("UnwrapPtr() returned nil for Some")
				}
				if *ptr != tt.wantValue {
					t.Errorf("UnwrapPtr() = %v, want %v", *ptr, tt.wantValue)
				}
			}
		})
	}
}

// TestUnwrapOrElse tests the UnwrapOrElse method
func TestUnwrapOrElse(t *testing.T) {
	tests := []struct {
		name           string
		option         Option[int]
		wantValue      int
		shouldCallFunc bool
	}{
		{"UnwrapOrElse Some returns value without calling func", Some(50), 50, false},
		{"UnwrapOrElse None calls function", None[int](), 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcCalled := false
			got := tt.option.UnwrapOrElse(func() int {
				funcCalled = true
				return 100
			})

			if got != tt.wantValue {
				t.Errorf("UnwrapOrElse() = %v, want %v", got, tt.wantValue)
			}

			if funcCalled != tt.shouldCallFunc {
				t.Errorf("UnwrapOrElse() function called = %v, want %v", funcCalled, tt.shouldCallFunc)
			}
		})
	}
}

// ============================================================================
// TRANSFORMATION TESTS
// ============================================================================

// TestThen tests the Then method
func TestThen(t *testing.T) {
	tests := []struct {
		name          string
		option        Option[int]
		expectedCalls int
	}{
		{"Then Some calls function", Some(5), 1},
		{"Then None doesn't call function", None[int](), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			tt.option.Then(func(val int) {
				calls++
				// Verify the value passed is correct (only for Some case)
				if tt.expectedCalls > 0 && val != 5 {
					t.Errorf("Then() passed value = %v, want 5", val)
				}
			})

			if calls != tt.expectedCalls {
				t.Errorf("Then() called %d times, want %d", calls, tt.expectedCalls)
			}
		})
	}
}

// TestAndThen tests the AndThen method
func TestAndThen(t *testing.T) {
	tests := []struct {
		name     string
		option   Option[int]
		wantBool bool
	}{
		{"AndThen Some with true condition", Some(10), true},
		{"AndThen Some with false condition", Some(2), false},
		{"AndThen None returns false", None[int](), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.option.AndThen(func(val int) bool {
				return val > 5
			})
			if got != tt.wantBool {
				t.Errorf("AndThen() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

// TestMap tests the Map function
func TestMap(t *testing.T) {
	tests := []struct {
		name      string
		option    Option[int]
		wantValue int
		wantSome  bool
	}{
		{"Map Some applies function", Some(5), 10, true},
		{"Map None returns None", None[int](), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.option, func(val int) int {
				return val * 2
			})

			if result.IsSome() != tt.wantSome {
				t.Errorf("Map() IsSome() = %v, want %v", result.IsSome(), tt.wantSome)
			}

			if tt.wantSome && result.Unwrap() != tt.wantValue {
				t.Errorf("Map() value = %v, want %v", result.Unwrap(), tt.wantValue)
			}
		})
	}
}

// TestMapChaining tests chaining multiple Map calls
func TestMapChaining(t *testing.T) {
	option := Some(5)

	// Chain multiple maps
	result := Map(option, func(val int) int {
		return val * 2
	})

	result = Map(result, func(val int) int {
		return val + 3
	})

	if !result.IsSome() {
		t.Error("Chained Map() should return Some")
	}

	got := result.Unwrap()
	if got != 13 { // (5 * 2) + 3 = 13
		t.Errorf("Chained Map() = %v, want 13", got)
	}
}

// TestMapDifferentTypes tests Map with different input and output types
func TestMapDifferentTypes(t *testing.T) {
	option := Some("hello")

	result := Map(option, func(val string) int {
		return len(val)
	})

	if !result.IsSome() {
		t.Error("Map() with type conversion should return Some")
	}

	got := result.Unwrap()
	if got != 5 {
		t.Errorf("Map() with type conversion = %v, want 5", got)
	}
}

// ============================================================================
// GENERIC TYPE TESTS (ensuring generics work with various types)
// ============================================================================

// TestWithPointerType tests Option with pointer types
func TestWithPointerType(t *testing.T) {
	value := 42
	ptr := &value
	option := Some(ptr)

	if !option.IsSome() {
		t.Error("Some() with pointer should be Some")
	}

	retrieved := option.Unwrap()
	if retrieved != ptr {
		t.Error("Pointer should be preserved through Some/Unwrap")
	}
	if *retrieved != value {
		t.Errorf("Pointer dereference = %v, want %v", *retrieved, value)
	}
}

// TestWithSliceType tests Option with slice types
func TestWithSliceType(t *testing.T) {
	slice := []int{1, 2, 3}
	option := Some(slice)

	if !option.IsSome() {
		t.Error("Some() with slice should be Some")
	}

	retrieved := option.Unwrap()
	if len(retrieved) != 3 {
		t.Errorf("Slice length = %d, want 3", len(retrieved))
	}

	// Test Map with slice type
	result := Map(option, func(s []int) int {
		return len(s)
	})
	if result.Unwrap() != 3 {
		t.Error("Map on slice type failed")
	}
}

// TestWithStructType tests Option with custom struct types
func TestWithStructType(t *testing.T) {
	type Person struct {
		name string
		age  int
	}

	person := Person{"Alice", 30}
	option := Some(person)

	if !option.IsSome() {
		t.Error("Some() with struct should be Some")
	}

	retrieved := option.Unwrap()
	if retrieved.name != "Alice" || retrieved.age != 30 {
		t.Error("Struct fields not preserved")
	}

	// Test Map with struct type
	result := Map(option, func(p Person) string {
		return p.name
	})
	if result.Unwrap() != "Alice" {
		t.Error("Map on struct type failed")
	}
}
