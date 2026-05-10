package cmds

import (
	"testing"
)

func TestCurrentLLMWriteMode(t *testing.T) {
	tests := []struct {
		env  string
		want string
	}{
		{"", llmWriteModeReadOnly},
		{"read-only", llmWriteModeReadOnly},
		{"confirm", llmWriteModeConfirm},
		{"direct", llmWriteModeDirect},
		{"CONFIRM", llmWriteModeConfirm},
		{"Direct", llmWriteModeDirect},
		{"  read-only  ", llmWriteModeReadOnly},
		{"garbage", llmWriteModeReadOnly},
	}

	for _, tt := range tests {
		t.Run("env="+tt.env, func(t *testing.T) {
			t.Setenv("LOGSEQ_LLM_WRITE_MODE", tt.env)
			got := currentLLMWriteMode()
			if got != tt.want {
				t.Errorf("currentLLMWriteMode()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestLLMDeleteRequiresConfirm(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"", true},
		{"1", true},
		{"true", true},
		{"yes", true},
		{"on", true},
		{"0", false},
		{"false", false},
		{"no", false},
		{"off", false},
		{"garbage", true},
	}

	for _, tt := range tests {
		t.Run("env="+tt.env, func(t *testing.T) {
			t.Setenv("LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE", tt.env)
			got := llmDeleteRequiresConfirm()
			if got != tt.want {
				t.Errorf("llmDeleteRequiresConfirm()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureWriteAllowed(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		deleteEnv string
		action    string
		dryRun    bool
		confirm   bool
		dangerous bool
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "dryRun always passes",
			mode:    "read-only",
			action:  "delete",
			dryRun:  true,
			wantErr: false,
		},
		{
			name:      "read-only blocks writes",
			mode:      "read-only",
			action:    "update",
			wantErr:   true,
			errSubstr: "SAFETY_BLOCKED",
		},
		{
			name:      "confirm mode without confirm flag",
			mode:      "confirm",
			action:    "update",
			confirm:   false,
			wantErr:   true,
			errSubstr: "requires --confirm",
		},
		{
			name:    "confirm mode with confirm flag",
			mode:    "confirm",
			action:  "update",
			confirm: true,
			wantErr: false,
		},
		{
			name:    "direct mode allows writes",
			mode:    "direct",
			action:  "update",
			wantErr: false,
		},
		{
			name:      "dangerous action requires confirm even in direct",
			mode:      "direct",
			deleteEnv: "true",
			action:    "delete",
			dangerous: true,
			confirm:   false,
			wantErr:   true,
			errSubstr: "SAFETY_BLOCKED",
		},
		{
			name:      "dangerous action with confirm passes",
			mode:      "direct",
			deleteEnv: "true",
			action:    "delete",
			dangerous: true,
			confirm:   true,
			wantErr:   false,
		},
		{
			name:      "dangerous action with delete confirm disabled",
			mode:      "direct",
			deleteEnv: "false",
			action:    "delete",
			dangerous: true,
			confirm:   false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOGSEQ_LLM_WRITE_MODE", tt.mode)
			if tt.deleteEnv != "" {
				t.Setenv("LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE", tt.deleteEnv)
			}
			err := ensureWriteAllowed(tt.action, tt.dryRun, tt.confirm, tt.dangerous)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAppendSafeStore(t *testing.T) {
	// Clean state
	appendSafeStoreMu.Lock()
	appendSafeStore = map[string]appendSafeRecord{}
	appendSafeStoreMu.Unlock()

	// Empty key returns not found
	_, ok := getAppendSafeRecord("")
	if ok {
		t.Error("empty key should return false")
	}

	// Set and get
	setAppendSafeRecord("page1", appendSafeRecord{BlockUUID: "uuid-1", Page: "page1"})
	rec, ok := getAppendSafeRecord("page1")
	if !ok {
		t.Fatal("expected to find record")
	}
	if rec.BlockUUID != "uuid-1" {
		t.Errorf("BlockUUID=%q, want %q", rec.BlockUUID, "uuid-1")
	}

	// Empty key set is a no-op
	setAppendSafeRecord("", appendSafeRecord{BlockUUID: "nope"})
	_, ok = getAppendSafeRecord("")
	if ok {
		t.Error("empty key should still return false after set")
	}

	// Overwrite
	setAppendSafeRecord("page1", appendSafeRecord{BlockUUID: "uuid-2", Page: "page1"})
	rec, _ = getAppendSafeRecord("page1")
	if rec.BlockUUID != "uuid-2" {
		t.Errorf("BlockUUID=%q, want %q after overwrite", rec.BlockUUID, "uuid-2")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
