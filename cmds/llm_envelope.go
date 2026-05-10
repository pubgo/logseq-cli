package cmds

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type llmEnvelopeError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type llmEnvelopeMeta struct {
	RequestID      string   `json:"request_id"`
	DurationMs     int64    `json:"duration_ms"`
	CapabilityUsed []string `json:"capability_used,omitempty"`
	FallbackUsed   []string `json:"fallback_used,omitempty"`
	NextCursor     string   `json:"next_cursor,omitempty"`
	HasMore        bool     `json:"has_more,omitempty"`
}

type llmEnvelope struct {
	OK    bool              `json:"ok"`
	Data  any               `json:"data,omitempty"`
	Error *llmEnvelopeError `json:"error,omitempty"`
	Meta  llmEnvelopeMeta   `json:"meta"`
	Hints []string          `json:"hints,omitempty"`
}

type llmEnvelopeOption func(*llmEnvelope)

var envelopeCounter uint64

func newLLMRequestID() string {
	n := atomic.AddUint64(&envelopeCounter, 1)
	return fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), n)
}

func withCapabilityUsed(v ...string) llmEnvelopeOption {
	return func(e *llmEnvelope) {
		e.Meta.CapabilityUsed = append(e.Meta.CapabilityUsed, v...)
	}
}

func withFallbackUsed(v ...string) llmEnvelopeOption {
	return func(e *llmEnvelope) {
		e.Meta.FallbackUsed = append(e.Meta.FallbackUsed, v...)
	}
}

func withHints(v ...string) llmEnvelopeOption {
	return func(e *llmEnvelope) {
		e.Hints = append(e.Hints, v...)
	}
}

func withPageMeta(nextCursor string, hasMore bool) llmEnvelopeOption {
	return func(e *llmEnvelope) {
		e.Meta.NextCursor = strings.TrimSpace(nextCursor)
		e.Meta.HasMore = hasMore
	}
}

func envelopeSuccess(start time.Time, data any, opts ...llmEnvelopeOption) *llmEnvelope {
	env := &llmEnvelope{
		OK:   true,
		Data: data,
		Meta: llmEnvelopeMeta{
			RequestID:  newLLMRequestID(),
			DurationMs: time.Since(start).Milliseconds(),
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(env)
		}
	}
	return env
}

func envelopeFailure(start time.Time, err error, code string, suggestion string, opts ...llmEnvelopeOption) *llmEnvelope {
	if strings.TrimSpace(code) == "" {
		code = classifyEnvelopeErrorCode(err)
	}

	env := &llmEnvelope{
		OK: false,
		Error: &llmEnvelopeError{
			Code:       code,
			Message:    strings.TrimSpace(errString(err)),
			Retryable:  isRetryableCode(code),
			Suggestion: strings.TrimSpace(suggestion),
		},
		Meta: llmEnvelopeMeta{
			RequestID:  newLLMRequestID(),
			DurationMs: time.Since(start).Milliseconds(),
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(env)
		}
	}
	return env
}

func errString(err error) string {
	if err == nil {
		return "unknown error"
	}
	return err.Error()
}

func classifyEnvelopeErrorCode(err error) string {
	if err == nil {
		return "UPSTREAM_ERROR"
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(msg, "safety_blocked"):
		return "SAFETY_BLOCKED"
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"), strings.Contains(msg, "missing"):
		return "BAD_REQUEST"
	case strings.Contains(msg, "not found"):
		return "RESOURCE_NOT_FOUND"
	case strings.Contains(msg, "methodnotexist"), strings.Contains(msg, "doesn't support"):
		return "CAPABILITY_UNAVAILABLE"
	case strings.Contains(msg, "timeout"):
		return "TIMEOUT"
	default:
		return "UPSTREAM_ERROR"
	}
}

func isRetryableCode(code string) bool {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "TIMEOUT", "UPSTREAM_ERROR":
		return true
	default:
		return false
	}
}
