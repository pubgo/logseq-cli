package cmds

import (
	"errors"
	"testing"
	"time"
)

func TestClassifyEnvelopeErrorCode(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{nil, "UPSTREAM_ERROR"},
		{errors.New("SAFETY_BLOCKED: delete blocked"), "SAFETY_BLOCKED"},
		{errors.New("query is required"), "BAD_REQUEST"},
		{errors.New("invalid limit"), "BAD_REQUEST"},
		{errors.New("missing field"), "BAD_REQUEST"},
		{errors.New("page not found"), "RESOURCE_NOT_FOUND"},
		{errors.New("methodnotexist"), "CAPABILITY_UNAVAILABLE"},
		{errors.New("API doesn't support this"), "CAPABILITY_UNAVAILABLE"},
		{errors.New("context deadline exceeded: timeout"), "TIMEOUT"},
		{errors.New("some random error"), "UPSTREAM_ERROR"},
	}

	for _, tt := range tests {
		name := "nil"
		if tt.err != nil {
			name = tt.err.Error()
		}
		t.Run(name, func(t *testing.T) {
			got := classifyEnvelopeErrorCode(tt.err)
			if got != tt.want {
				t.Errorf("classifyEnvelopeErrorCode(%v)=%q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsRetryableCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"TIMEOUT", true},
		{"UPSTREAM_ERROR", true},
		{"timeout", true},
		{"  TIMEOUT  ", true},
		{"BAD_REQUEST", false},
		{"SAFETY_BLOCKED", false},
		{"RESOURCE_NOT_FOUND", false},
		{"CAPABILITY_UNAVAILABLE", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := isRetryableCode(tt.code)
			if got != tt.want {
				t.Errorf("isRetryableCode(%q)=%v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestErrString(t *testing.T) {
	if got := errString(nil); got != "unknown error" {
		t.Errorf("errString(nil)=%q, want %q", got, "unknown error")
	}
	if got := errString(errors.New("some error")); got != "some error" {
		t.Errorf("errString(err)=%q, want %q", got, "some error")
	}
}

func TestEnvelopeSuccess(t *testing.T) {
	start := time.Now()
	env := envelopeSuccess(start, "data", withCapabilityUsed("cap1"), withHints("hint1"))
	if !env.OK {
		t.Error("expected OK=true")
	}
	if env.Data != "data" {
		t.Errorf("Data=%v, want %q", env.Data, "data")
	}
	if env.Error != nil {
		t.Error("expected Error=nil")
	}
	if len(env.Meta.CapabilityUsed) != 1 || env.Meta.CapabilityUsed[0] != "cap1" {
		t.Errorf("CapabilityUsed=%v", env.Meta.CapabilityUsed)
	}
	if len(env.Hints) != 1 || env.Hints[0] != "hint1" {
		t.Errorf("Hints=%v", env.Hints)
	}
	if env.Meta.RequestID == "" {
		t.Error("expected non-empty RequestID")
	}
}

func TestEnvelopeFailure(t *testing.T) {
	start := time.Now()
	env := envelopeFailure(start, errors.New("timeout"), "", "retry later")
	if env.OK {
		t.Error("expected OK=false")
	}
	if env.Error == nil {
		t.Fatal("expected Error != nil")
	}
	if env.Error.Code != "TIMEOUT" {
		t.Errorf("Error.Code=%q, want %q", env.Error.Code, "TIMEOUT")
	}
	if env.Error.Message != "timeout" {
		t.Errorf("Error.Message=%q, want %q", env.Error.Message, "timeout")
	}
	if !env.Error.Retryable {
		t.Error("expected Retryable=true for TIMEOUT")
	}
	if env.Error.Suggestion != "retry later" {
		t.Errorf("Suggestion=%q", env.Error.Suggestion)
	}

	// With explicit code
	env2 := envelopeFailure(start, errors.New("bad"), "BAD_REQUEST", "fix it")
	if env2.Error.Code != "BAD_REQUEST" {
		t.Errorf("expected explicit code BAD_REQUEST, got %q", env2.Error.Code)
	}
	if env2.Error.Retryable {
		t.Error("BAD_REQUEST should not be retryable")
	}
}

func TestWithPageMeta(t *testing.T) {
	start := time.Now()
	env := envelopeSuccess(start, nil, withPageMeta("42", true))
	if env.Meta.NextCursor != "42" {
		t.Errorf("NextCursor=%q, want %q", env.Meta.NextCursor, "42")
	}
	if !env.Meta.HasMore {
		t.Error("expected HasMore=true")
	}
}

func TestWithFallbackUsed(t *testing.T) {
	start := time.Now()
	env := envelopeSuccess(start, nil, withFallbackUsed("fb1", "fb2"))
	if len(env.Meta.FallbackUsed) != 2 {
		t.Errorf("FallbackUsed=%v", env.Meta.FallbackUsed)
	}
}
