package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type statusResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type pageResp struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type blockResp struct {
	UUID    string `json:"uuid"`
	Content string `json:"content"`
}

type config struct {
	cliBin     string
	token      string
	host       string
	port       string
	pagePrefix string
	keepPage   bool
	timeout    time.Duration
}

type runner struct {
	globalArgs []string
	bin        string
	env        []string
	timeout    time.Duration
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ e2e failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ e2e finished successfully")
}

func run() error {
	loadEnvFromDotEnv()

	cfg := parseFlags()
	if strings.TrimSpace(cfg.token) == "" {
		return errors.New("LOGSEQ_API_TOKEN is required (or pass --token)")
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	binPath := cfg.cliBin
	cleanupBin := func() {}
	if strings.TrimSpace(binPath) == "" {
		var cleanup func()
		binPath, cleanup, err = buildCLIBinary(workDir)
		if err != nil {
			return err
		}
		cleanupBin = cleanup
	}
	defer cleanupBin()

	r := &runner{
		globalArgs: []string{"--token", cfg.token, "--host", cfg.host, "--port", cfg.port},
		bin:        binPath,
		env: mergeEnv(map[string]string{
			"LOGSEQ_API_TOKEN": cfg.token,
			"LOGSEQ_HOST":      cfg.host,
			"LOGSEQ_PORT":      cfg.port,
		}),
		timeout: cfg.timeout,
	}

	pageName := fmt.Sprintf("%s_%d", cfg.pagePrefix, time.Now().UnixNano())
	shouldCleanup := !cfg.keepPage

	if shouldCleanup {
		defer func() {
			_, _, _ = r.run("page", "delete", pageName)
		}()
	}

	if err := step("graph info", func() error {
		var graph map[string]any
		if err := r.runJSON(&graph, "graph", "info"); err != nil {
			return err
		}
		if len(graph) == 0 {
			return errors.New("graph info is empty")
		}
		return nil
	}); err != nil {
		return err
	}

	if err := step("page create", func() error {
		var page pageResp
		if err := r.runJSON(&page, "page", "create", pageName); err != nil {
			return err
		}
		if page.UUID == "" {
			return fmt.Errorf("empty page uuid, response=%+v", page)
		}
		return nil
	}); err != nil {
		return err
	}

	const v1 = "copilot e2e block v1"
	const v2 = "copilot e2e block v2"

	var blockUUID string

	if err := step("block append", func() error {
		var block blockResp
		if err := r.runJSON(&block, "block", "append", pageName, v1); err != nil {
			return err
		}
		if block.UUID == "" {
			return fmt.Errorf("empty block uuid, response=%+v", block)
		}
		blockUUID = block.UUID
		return nil
	}); err != nil {
		return err
	}

	if err := step("block get(v1)", func() error {
		var block blockResp
		if err := r.runJSON(&block, "block", "get", blockUUID); err != nil {
			return err
		}
		if block.Content != v1 {
			return fmt.Errorf("unexpected content, want=%q got=%q", v1, block.Content)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := step("block update", func() error {
		var st statusResp
		if err := r.runJSON(&st, "block", "update", blockUUID, v2); err != nil {
			return err
		}
		if !st.OK {
			return fmt.Errorf("update failed: %+v", st)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := step("block get(v2)", func() error {
		var block blockResp
		if err := r.runJSON(&block, "block", "get", blockUUID); err != nil {
			return err
		}
		if block.Content != v2 {
			return fmt.Errorf("unexpected content, want=%q got=%q", v2, block.Content)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := step("block remove", func() error {
		var st statusResp
		if err := r.runJSON(&st, "block", "remove", blockUUID); err != nil {
			return err
		}
		if !st.OK {
			return fmt.Errorf("remove failed: %+v", st)
		}
		return nil
	}); err != nil {
		return err
	}

	if shouldCleanup {
		if err := step("page delete", func() error {
			var st statusResp
			if err := r.runJSON(&st, "page", "delete", pageName); err != nil {
				return err
			}
			if !st.OK {
				return fmt.Errorf("delete failed: %+v", st)
			}
			return nil
		}); err != nil {
			return err
		}

		if err := step("query verify page removed", func() error {
			query := fmt.Sprintf(`[:find ?e :where [?e :block/name %q]]`, pageName)
			var result []any
			if err := r.runJSON(&result, "query", "datalog", query); err != nil {
				return err
			}
			if len(result) != 0 {
				return fmt.Errorf("expected empty query result, got=%v", result)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if cfg.keepPage {
		fmt.Printf("ℹ️ keep page enabled, left test page: %s\n", pageName)
	}

	return nil
}

func parseFlags() config {
	var cfg config

	flag.StringVar(&cfg.cliBin, "cli-bin", "", "Path to prebuilt logseq CLI binary (empty means auto-build current repo)")
	flag.StringVar(&cfg.token, "token", envOrDefault("LOGSEQ_API_TOKEN", ""), "Logseq API token")
	flag.StringVar(&cfg.host, "host", envOrDefault("LOGSEQ_HOST", "127.0.0.1"), "Logseq API host")
	flag.StringVar(&cfg.port, "port", envOrDefault("LOGSEQ_PORT", "12315"), "Logseq API port")
	flag.StringVar(&cfg.pagePrefix, "page-prefix", "__copilot_e2e", "Temporary page name prefix")
	flag.BoolVar(&cfg.keepPage, "keep-page", false, "Keep created page for debugging")
	flag.DurationVar(&cfg.timeout, "timeout", 60*time.Second, "Per command timeout")

	flag.Parse()
	return cfg
}

func envOrDefault(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func mergeEnv(overrides map[string]string) []string {
	env := os.Environ()
	used := make(map[string]struct{}, len(overrides))

	filtered := make([]string, 0, len(env)+len(overrides))
	for _, item := range env {
		idx := strings.Index(item, "=")
		if idx <= 0 {
			filtered = append(filtered, item)
			continue
		}
		k := item[:idx]
		if _, ok := overrides[k]; ok {
			continue
		}
		filtered = append(filtered, item)
	}

	for k, v := range overrides {
		if _, ok := used[k]; ok {
			continue
		}
		filtered = append(filtered, k+"="+v)
		used[k] = struct{}{}
	}

	return filtered
}

func buildCLIBinary(workDir string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "logseq-e2e-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	binPath := filepath.Join(tmpDir, "logseq-cli")

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = workDir
	cmd.Env = os.Environ()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("build cli: %w\n%s", err, strings.TrimSpace(stderr.String()))
	}

	return binPath, cleanup, nil
}

func (r *runner) runJSON(dst any, args ...string) error {
	out, errOut, err := r.run(args...)
	if err != nil {
		return fmt.Errorf("cmd failed (%s): %w, stdout=%s, stderr=%s", strings.Join(args, " "), err, out, errOut)
	}
	if err := json.Unmarshal([]byte(out), dst); err != nil {
		return fmt.Errorf("decode json failed (%s): %w, stdout=%s, stderr=%s", strings.Join(args, " "), err, out, errOut)
	}
	return nil
}

func (r *runner) run(args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	fullArgs := append(append([]string{}, r.globalArgs...), args...)
	cmd := exec.CommandContext(ctx, r.bin, fullArgs...)
	cmd.Env = r.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func step(name string, fn func() error) error {
	start := time.Now()
	fmt.Printf("▶ %s\n", name)
	if err := fn(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	fmt.Printf("✔ %s (%s)\n", name, time.Since(start).Round(time.Millisecond))
	return nil
}

func loadEnvFromDotEnv() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	data, err := os.ReadFile(filepath.Join(wd, ".env"))
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.Trim(strings.TrimSpace(line[idx+1:]), `"'`)
		if key == "" || value == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
