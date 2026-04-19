package tls_client

import (
	"math"
	"testing"
)

func TestInt64ToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   int64
		want    int
		wantErr bool
	}{
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "positive value within range",
			input:   12345,
			want:    12345,
			wantErr: false,
		},
		{
			name:    "negative value within range",
			input:   -12345,
			want:    -12345,
			wantErr: false,
		},
		{
			name:    "max int value",
			input:   int64(math.MaxInt),
			want:    math.MaxInt,
			wantErr: false,
		},
		{
			name:    "min int value",
			input:   int64(math.MinInt),
			want:    math.MinInt,
			wantErr: false,
		},
	}

	// Add overflow tests only on 32-bit systems
	// On 64-bit systems, int and int64 have the same range
	// Use hardcoded int32 limits to test overflow behavior on 32-bit systems
	if math.MaxInt < math.MaxInt64 {
		// We're on a 32-bit system, add overflow tests
		tests = append(tests, []struct {
			name    string
			input   int64
			want    int
			wantErr bool
		}{
			{
				name:    "value exceeds max int32",
				input:   math.MaxInt32 + 1,
				want:    0,
				wantErr: true,
			},
			{
				name:    "value below min int32",
				input:   math.MinInt32 - 1,
				want:    0,
				wantErr: true,
			},
			{
				name:    "max int64 value on 32-bit",
				input:   math.MaxInt64,
				want:    0,
				wantErr: true,
			},
			{
				name:    "min int64 value on 32-bit",
				input:   math.MinInt64,
				want:    0,
				wantErr: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int64ToInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int64ToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int64ToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
