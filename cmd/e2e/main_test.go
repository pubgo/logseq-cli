package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestE2ESmoke(t *testing.T) {
	if os.Getenv("LOGSEQ_E2E") != "1" {
		t.Skip("skip e2e smoke test: set LOGSEQ_E2E=1 to enable")
	}

	token := strings.TrimSpace(os.Getenv("LOGSEQ_API_TOKEN"))
	if token == "" {
		t.Skip("skip e2e smoke test: LOGSEQ_API_TOKEN is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	args := []string{"run", ".", "--token", token}
	if cliBin := strings.TrimSpace(os.Getenv("LOGSEQ_E2E_CLI_BIN")); cliBin != "" {
		args = append(args, "--cli-bin", cliBin)
	}
	if pagePrefix := strings.TrimSpace(os.Getenv("LOGSEQ_E2E_PAGE_PREFIX")); pagePrefix != "" {
		args = append(args, "--page-prefix", pagePrefix)
	}
	if timeout := strings.TrimSpace(os.Getenv("LOGSEQ_E2E_TIMEOUT")); timeout != "" {
		args = append(args, "--timeout", timeout)
	}
	if envBool("LOGSEQ_E2E_KEEP_PAGE") {
		args = append(args, "--keep-page")
	}
	if host := strings.TrimSpace(os.Getenv("LOGSEQ_HOST")); host != "" {
		args = append(args, "--host", host)
	}
	if port := strings.TrimSpace(os.Getenv("LOGSEQ_PORT")); port != "" {
		args = append(args, "--port", port)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = "."
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("e2e smoke failed: %v\noutput:\n%s", err, string(out))
	}
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
