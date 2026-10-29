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

func TestSanitizeRawQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty query",
			input: "",
			want:  "",
		},
		{
			name:  "plain query stays",
			input: "a=1&b=2",
			want:  "a=1&b=2",
		},
		{
			name:  "space becomes %20",
			input: "scope=openid offline_access",
			want:  "scope=openid%20offline_access",
		},
		{
			name:  "multiple spaces",
			input: "scope=openid offline_access profile email",
			want:  "scope=openid%20offline_access%20profile%20email",
		},
		{
			name:  "already encoded stays",
			input: "redirect_uri=https%3A%2F%2Fx.com%2Fcb",
			want:  "redirect_uri=https%3A%2F%2Fx.com%2Fcb",
		},
		{
			name:  "mixed raw and encoded",
			input: "q=a%20b c",
			want:  "q=a%20b%20c",
		},
		{
			name:  "plus stays raw",
			input: "q=a+b",
			want:  "q=a+b",
		},
		{
			name:  "invalid ascii encoded",
			input: "a=b<>|\"c",
			want:  "a=b%3C%3E%7C%22c",
		},
		{
			name:  "non ascii encoded",
			input: "a=ä",
			want:  "a=%C3%A4",
		},
		{
			name:  "real world scope param",
			input: "response_type=code&scope=openid offline_access profile&prompt=test",
			want:  "response_type=code&scope=openid%20offline_access%20profile&prompt=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeRawQuery(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeRawQuery(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
