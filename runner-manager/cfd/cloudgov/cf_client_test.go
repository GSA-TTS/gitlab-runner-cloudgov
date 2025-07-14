package cloudgov

import "testing"

func Test_parsePortRange(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		prange  string
		want    int
		want2   int
		wantErr bool
	}{
		{name: "parses a range", prange: "80-85", want: 80, want2: 85},
		{name: "parses redundant range", prange: "81-81", want: 81, want2: 81},
		{name: "parses single port", prange: "80", want: 80, want2: 80},
		{name: "parses single number with separator", prange: "8-", want: 8, want2: 8},
		{name: "parses zero", prange: "0", want: 0, want2: 0},
		{name: "fails to parse range with non-int", prange: "60-cat", wantErr: true},
		{name: "fails to parse single with non-int", prange: "cat", wantErr: true},
		{name: "fails with empty string", prange: "", wantErr: true},
		{name: "fails with only separator", prange: "-", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2, gotErr := parsePortRange(tt.prange)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parsePortRange() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parsePortRange() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("parsePortRange() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("parsePortRange() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
