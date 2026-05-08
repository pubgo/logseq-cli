package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

var (
	Token  string
	Host   string
	Port   string
	Output string
)

type StatusResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// ClientConnectionInfo describes the effective Logseq connection configuration.
// Token value itself is intentionally not exposed.
type ClientConnectionInfo struct {
	BaseURL         string `json:"baseURL"`
	Host            string `json:"host"`
	Port            string `json:"port"`
	HostSource      string `json:"hostSource"`
	PortSource      string `json:"portSource"`
	TokenConfigured bool   `json:"tokenConfigured"`
	TokenSource     string `json:"tokenSource"`
}

type resolvedClient struct {
	info  ClientConnectionInfo
	token string
}

func resolveClient() resolvedClient {
	host := strings.TrimSpace(Host)
	port := strings.TrimSpace(Port)
	token := strings.TrimSpace(Token)
	hostSource := "option"
	portSource := "option"
	tokenSource := "option"

	if host == "" {
		host = strings.TrimSpace(os.Getenv("LOGSEQ_HOST"))
		hostSource = "env"
	}
	if port == "" {
		port = strings.TrimSpace(os.Getenv("LOGSEQ_PORT"))
		portSource = "env"
	}
	if token == "" {
		token = strings.TrimSpace(os.Getenv("LOGSEQ_API_TOKEN"))
		tokenSource = "env"
	}

	if host == "" {
		host = "127.0.0.1"
		hostSource = "default"
	}
	if port == "" {
		port = "12315"
		portSource = "default"
	}
	if token == "" {
		tokenSource = "missing"
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, port)
	return resolvedClient{
		info: ClientConnectionInfo{
			BaseURL:         baseURL,
			Host:            host,
			Port:            port,
			HostSource:      hostSource,
			PortSource:      portSource,
			TokenConfigured: token != "",
			TokenSource:     tokenSource,
		},
		token: token,
	}
}

func CurrentClientConnectionInfo() ClientConnectionInfo {
	return resolveClient().info
}

func NewClient() *logseq.Client {
	resolved := resolveClient()
	return logseq.NewClient(
		logseq.WithBaseURL(resolved.info.BaseURL),
		logseq.WithToken(resolved.token),
	)
}
