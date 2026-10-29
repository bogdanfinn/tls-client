package tls_client

import (
	"math"
	"testing"
)

func TestGenerateH2GreaseSettingID(t *testing.T) {
	// RFC 8701 reserved 16-bit GREASE values: 0x0a0a, 0x1a1a, ..., 0xfafa
	validGrease := map[uint16]bool{
		0x0a0a: true, 0x1a1a: true, 0x2a2a: true, 0x3a3a: true,
		0x4a4a: true, 0x5a5a: true, 0x6a6a: true, 0x7a7a: true,
		0x8a8a: true, 0x9a9a: true, 0xaaaa: true, 0xbaba: true,
		0xcaca: true, 0xdada: true, 0xeaea: true, 0xfafa: true,
	}

	seen := map[uint16]bool{}
	for i := 0; i < 1000; i++ {
		id := generateH2GreaseSettingID()
		if !validGrease[id] {
			t.Errorf("generated non-RFC-8701 GREASE setting ID: 0x%04x", id)
		}
		seen[id] = true
	}

	// 1000 draws from 16 values should produce more than one distinct value,
	// confirming the ID is randomized per call (real Chrome randomizes per connection).
	if len(seen) < 2 {
		t.Errorf("expected randomized GREASE IDs, got %d distinct value(s)", len(seen))
	}
}

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
