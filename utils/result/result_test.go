package result

import (
	"errors"
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

// TestOk tests the Ok constructor
func TestOk(t *testing.T) {
	value := 42
	result := Ok(value)
	if result.IsErr() {
		t.Error("Ok() should create a result with a value")
	}
	if !result.IsOk() {
		t.Error("Ok() should be marked as Ok")
	}
	// Verify the value is stored correctly using Unwrap() instead of direct field access
	if result.Unwrap() != value {
		t.Errorf("Ok() should store the value correctly, got %d, want %d", result.Unwrap(), value)
	}
}

// TestErr tests the Err constructor
func TestErr(t *testing.T) {
	testErr := errors.New("test error")
	result := Err[int](testErr)
	if result.IsOk() {
		t.Error("Err() should create a result with an error")
	}
	if !result.IsErr() {
		t.Error("Err() should be marked as Err")
	}
}

// TestTryFunc tests the TryFunc wrapper function
func TestTryFunc(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() (int, error)
		shouldErr bool
		wantValue int
	}{
		{
			name: "TryFunc with successful function",
			fn: func() (int, error) {
				return 100, nil
			},
			shouldErr: false,
			wantValue: 100,
		},
		{
			name: "TryFunc with error",
			fn: func() (int, error) {
				return 0, errors.New("function error")
			},
			shouldErr: true,
			wantValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryFunc(tt.fn)
			if result.IsErr() != tt.shouldErr {
				t.Errorf("TryFunc() IsErr() = %v, want %v", result.IsErr(), tt.shouldErr)
			}
			if !tt.shouldErr && result.Unwrap() != tt.wantValue {
				t.Errorf("TryFunc() value = %v, want %v", result.Unwrap(), tt.wantValue)
			}
		})
	}
}

// ============================================================================
// STATE CHECK TESTS
// ============================================================================

// TestIsOk tests the IsOk method
func TestIsOk(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[string]
		wantOk   bool
	}{
		{"Err returns false", Err[string](errors.New("error")), false},
		{"Ok returns true", Ok("value"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsOk(); got != tt.wantOk {
				t.Errorf("IsOk() = %v, want %v", got, tt.wantOk)
			}
		})
	}
}

// TestIsErr tests the IsErr method
func TestIsErr(t *testing.T) {
	tests := []struct {
		name      string
		result    Result[int]
		wantErr   bool
	}{
		{"Err returns true", Err[int](errors.New("error")), true},
		{"Ok returns false", Ok(123), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsErr(); got != tt.wantErr {
				t.Errorf("IsErr() = %v, want %v", got, tt.wantErr)
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
		result      Result[int]
		wantValue   int
		shouldPanic bool
	}{
		{"Unwrap Ok returns value", Ok(99), 99, false},
		{"Unwrap Err panics", Err[int](errors.New("error")), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				expectPanic(t, func() {
					_ = tt.result.Unwrap()
				})
			} else {
				got := tt.result.Unwrap()
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
		result      Result[int]
		shouldPanic bool
		wantValue   int
	}{
		{
			name:        "UnwrapPtr Ok returns pointer",
			result:      Ok(42),
			shouldPanic: false,
			wantValue:   42,
		},
		{
			name:        "UnwrapPtr Err panics",
			result:      Err[int](errors.New("error")),
			shouldPanic: true,
			wantValue:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				expectPanic(t, func() {
					_ = tt.result.UnwrapPtr()
				})
			} else {
				ptr := tt.result.UnwrapPtr()
				if ptr == nil {
					t.Error("UnwrapPtr() returned nil for Ok")
				}
				if *ptr != tt.wantValue {
					t.Errorf("UnwrapPtr() = %v, want %v", *ptr, tt.wantValue)
				}
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
		result        Result[int]
		expectedCalls int
	}{
		{"Then Ok calls function", Ok(5), 1},
		{"Then Err doesn't call function", Err[int](errors.New("error")), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			tt.result.Then(func(val int) {
				calls++
				// Verify the value passed is correct (only for Ok case)
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
		result   Result[int]
		wantBool bool
	}{
		{"AndThen Ok with true condition", Ok(10), true},
		{"AndThen Ok with false condition", Ok(2), false},
		{"AndThen Err returns false", Err[int](errors.New("error")), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.AndThen(func(val int) bool {
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
		result    Result[int]
		wantValue int
		wantOk    bool
	}{
		{"Map Ok applies function", Ok(5), 10, true},
		{"Map Err returns Err", Err[int](errors.New("error")), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mappedResult := Map(tt.result, func(val int) int {
				return val * 2
			})

			if mappedResult.IsOk() != tt.wantOk {
				t.Errorf("Map() IsOk() = %v, want %v", mappedResult.IsOk(), tt.wantOk)
			}

			if tt.wantOk && mappedResult.Unwrap() != tt.wantValue {
				t.Errorf("Map() value = %v, want %v", mappedResult.Unwrap(), tt.wantValue)
			}
		})
	}
}

// TestMapChaining tests chaining multiple Map calls
func TestMapChaining(t *testing.T) {
	result := Ok(5)

	// Chain multiple maps
	mappedResult := Map(result, func(val int) int {
		return val * 2
	})

	mappedResult = Map(mappedResult, func(val int) int {
		return val + 3
	})

	if !mappedResult.IsOk() {
		t.Error("Chained Map() should return Ok")
	}

	got := mappedResult.Unwrap()
	if got != 13 { // (5 * 2) + 3 = 13
		t.Errorf("Chained Map() = %v, want 13", got)
	}
}

// TestMapDifferentTypes tests Map with different input and output types
func TestMapDifferentTypes(t *testing.T) {
	result := Ok("hello")

	mappedResult := Map(result, func(val string) int {
		return len(val)
	})

	if !mappedResult.IsOk() {
		t.Error("Map() with type conversion should return Ok")
	}

	got := mappedResult.Unwrap()
	if got != 5 {
		t.Errorf("Map() with type conversion = %v, want 5", got)
	}
}

// TestTry tests the Try function (flatMap equivalent)
func TestTry(t *testing.T) {
	tests := []struct {
		name      string
		result    Result[int]
		fn        func(int) Result[int]
		wantValue int
		wantOk    bool
	}{
		{
			name:   "Try Ok with Ok result",
			result: Ok(5),
			fn: func(val int) Result[int] {
				return Ok(val * 2)
			},
			wantValue: 10,
			wantOk:    true,
		},
		{
			name:   "Try Ok with Err result",
			result: Ok(5),
			fn: func(val int) Result[int] {
				return Err[int](errors.New("inner error"))
			},
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:   "Try Err skips function",
			result: Err[int](errors.New("outer error")),
			fn: func(val int) Result[int] {
				t.Error("Try() should not call function for Err")
				return Ok(val)
			},
			wantValue: 0,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tryResult := Try(tt.result, tt.fn)

			if tryResult.IsOk() != tt.wantOk {
				t.Errorf("Try() IsOk() = %v, want %v", tryResult.IsOk(), tt.wantOk)
			}

			if tt.wantOk && tryResult.Unwrap() != tt.wantValue {
				t.Errorf("Try() value = %v, want %v", tryResult.Unwrap(), tt.wantValue)
			}
		})
	}
}

// TestTryChaining tests chaining multiple Try calls
func TestTryChaining(t *testing.T) {
	result := Ok(5)

	// Chain multiple tries
	tryResult := Try(result, func(val int) Result[int] {
		return Ok(val * 2)
	})

	tryResult = Try(tryResult, func(val int) Result[int] {
		return Ok(val + 3)
	})

	if !tryResult.IsOk() {
		t.Error("Chained Try() should return Ok")
	}

	got := tryResult.Unwrap()
	if got != 13 { // (5 * 2) + 3 = 13
		t.Errorf("Chained Try() = %v, want 13", got)
	}
}

// TestTryWithError tests Try propagating errors correctly
func TestTryWithError(t *testing.T) {
	result := Ok(5)
	testErr := errors.New("test error")

	// Chain tries with an error in the middle
	tryResult := Try(result, func(val int) Result[int] {
		return Ok(val * 2)
	})

	tryResult = Try(tryResult, func(val int) Result[int] {
		// Return an error
		return Err[int](testErr)
	})

	tryResult = Try(tryResult, func(val int) Result[int] {
		t.Error("Try() should not call function after error")
		return Ok(val)
	})

	if !tryResult.IsErr() {
		t.Error("Try() should propagate error")
	}
}

// ============================================================================
// GENERIC TYPE TESTS (ensuring generics work with various types)
// ============================================================================

// TestWithPointerType tests Result with pointer types
func TestWithPointerType(t *testing.T) {
	value := 42
	ptr := &value
	result := Ok(ptr)

	if !result.IsOk() {
		t.Error("Ok() with pointer should be Ok")
	}

	retrieved := result.Unwrap()
	if retrieved != ptr {
		t.Error("Pointer should be preserved through Ok/Unwrap")
	}
	if *retrieved != value {
		t.Errorf("Pointer dereference = %v, want %v", *retrieved, value)
	}
}

// TestWithSliceType tests Result with slice types
func TestWithSliceType(t *testing.T) {
	slice := []int{1, 2, 3}
	result := Ok(slice)

	if !result.IsOk() {
		t.Error("Ok() with slice should be Ok")
	}

	retrieved := result.Unwrap()
	if len(retrieved) != 3 {
		t.Errorf("Slice length = %d, want 3", len(retrieved))
	}

	// Test Map with slice type
	mappedResult := Map(result, func(s []int) int {
		return len(s)
	})
	if mappedResult.Unwrap() != 3 {
		t.Error("Map on slice type failed")
	}
}

// TestWithStructType tests Result with custom struct types
func TestWithStructType(t *testing.T) {
	type Person struct {
		name string
		age  int
	}

	person := Person{"Alice", 30}
	result := Ok(person)

	if !result.IsOk() {
		t.Error("Ok() with struct should be Ok")
	}

	retrieved := result.Unwrap()
	if retrieved.name != "Alice" || retrieved.age != 30 {
		t.Error("Struct fields not preserved")
	}

	// Test Map with struct type
	mappedResult := Map(result, func(p Person) string {
		return p.name
	})
	if mappedResult.Unwrap() != "Alice" {
		t.Error("Map on struct type failed")
	}
}

// TestErrorPreservation tests that errors are preserved through transformations
func TestErrorPreservation(t *testing.T) {
	testErr := errors.New("specific error")
	result := Err[int](testErr)

	// Map should preserve the error
	mappedResult := Map(result, func(val int) int {
		return val * 2
	})

	if !mappedResult.IsErr() {
		t.Error("Map() should preserve error for Err result")
	}

	// Try should preserve the error
	tryResult := Try(result, func(val int) Result[int] {
		return Ok(val)
	})

	if !tryResult.IsErr() {
		t.Error("Try() should preserve error for Err result")
	}
}

