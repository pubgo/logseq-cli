package cmds

import (
	"encoding/json"
	"testing"
)

func TestAnyToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    int
		wantErr bool
	}{
		{"float64", float64(42), 42, false},
		{"int", int(7), 7, false},
		{"int64", int64(100), 100, false},
		{"json.Number", json.Number("55"), 55, false},
		{"string", "hello", 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := anyToInt(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("anyToInt(%v)=%d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestShortErr(t *testing.T) {
	if got := shortErr(nil); got != "" {
		t.Errorf("shortErr(nil)=%q, want empty", got)
	}

	// Short message stays as-is
	short := "some error"
	if got := shortErr(errFromString(short)); got != short {
		t.Errorf("shortErr=%q, want %q", got, short)
	}

	// Long message gets truncated
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}
	got := shortErr(errFromString(string(long)))
	if len(got) > 184 { // 180 + "..."
		t.Errorf("shortErr len=%d, expected <= 184", len(got))
	}
}

func TestEnvBoolOrDefault(t *testing.T) {
	key := "TEST_ENVBOOL_DF34K"

	tests := []struct {
		env  string
		def  bool
		want bool
	}{
		{"", true, true},
		{"", false, false},
		{"true", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"on", false, true},
		{"false", true, false},
		{"0", true, false},
		{"no", true, false},
		{"off", true, false},
		{"garbage", true, true},
		{"garbage", false, false},
	}

	for _, tt := range tests {
		t.Run("env="+tt.env, func(t *testing.T) {
			t.Setenv(key, tt.env)
			got := envBoolOrDefault(key, tt.def)
			if got != tt.want {
				t.Errorf("envBoolOrDefault(%q, %v)=%v, want %v", tt.env, tt.def, got, tt.want)
			}
		})
	}
}

func TestEnvIntOrDefault(t *testing.T) {
	key := "TEST_ENVINT_XK92P"

	tests := []struct {
		env  string
		def  int
		want int
	}{
		{"", 42, 42},
		{"100", 42, 100},
		{"0", 42, 0},
		{"-5", 42, -5},
		{"abc", 42, 42},
	}

	for _, tt := range tests {
		t.Run("env="+tt.env, func(t *testing.T) {
			t.Setenv(key, tt.env)
			got := envIntOrDefault(key, tt.def)
			if got != tt.want {
				t.Errorf("envIntOrDefault(%q, %d)=%d, want %d", tt.env, tt.def, got, tt.want)
			}
		})
	}
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }

func errFromString(s string) error {
	return simpleErr(s)
}
