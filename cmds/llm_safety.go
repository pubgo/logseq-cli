package cmds

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	llmWriteModeReadOnly = "read-only"
	llmWriteModeConfirm  = "confirm"
	llmWriteModeDirect   = "direct"
)

type appendSafeRecord struct {
	BlockUUID      string    `json:"block_uuid"`
	Page           string    `json:"page"`
	ContentPreview string    `json:"content_preview"`
	CreatedAt      time.Time `json:"created_at"`
}

var (
	appendSafeStoreMu sync.Mutex
	appendSafeStore   = map[string]appendSafeRecord{}
)

func currentLLMWriteMode() string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("LOGSEQ_LLM_WRITE_MODE")))
	switch mode {
	case llmWriteModeReadOnly, llmWriteModeConfirm, llmWriteModeDirect:
		return mode
	default:
		return llmWriteModeReadOnly
	}
}

func llmDeleteRequiresConfirm() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE")))
	if raw == "" {
		return true
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func ensureWriteAllowed(action string, dryRun bool, confirm bool, dangerous bool) error {
	if dryRun {
		return nil
	}

	mode := currentLLMWriteMode()
	if mode == llmWriteModeReadOnly {
		return fmt.Errorf("SAFETY_BLOCKED: %s blocked by LOGSEQ_LLM_WRITE_MODE=read-only", action)
	}

	if mode == llmWriteModeConfirm && !confirm {
		return fmt.Errorf("SAFETY_BLOCKED: %s requires --confirm in LOGSEQ_LLM_WRITE_MODE=confirm", action)
	}

	if dangerous && llmDeleteRequiresConfirm() && !confirm {
		return fmt.Errorf("SAFETY_BLOCKED: %s requires --confirm (LOGSEQ_LLM_REQUIRE_CONFIRM_FOR_DELETE=true)", action)
	}

	return nil
}

func getAppendSafeRecord(key string) (appendSafeRecord, bool) {
	k := strings.TrimSpace(key)
	if k == "" {
		return appendSafeRecord{}, false
	}
	appendSafeStoreMu.Lock()
	defer appendSafeStoreMu.Unlock()
	v, ok := appendSafeStore[k]
	return v, ok
}

func setAppendSafeRecord(key string, rec appendSafeRecord) {
	k := strings.TrimSpace(key)
	if k == "" {
		return
	}
	appendSafeStoreMu.Lock()
	defer appendSafeStoreMu.Unlock()
	appendSafeStore[k] = rec
}
