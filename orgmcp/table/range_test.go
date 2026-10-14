package table

import "testing"

func Test_parsePythonRange(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		str     string
		want    TableRange
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parsePythonRange(tt.str)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parsePythonRange() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parsePythonRange() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("parsePythonRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
