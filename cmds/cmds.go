package cmds

import (
	"fmt"

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

func NewClient() *logseq.Client {
	baseURL := fmt.Sprintf("http://%s:%s", Host, Port)
	return logseq.NewClient(
		logseq.WithBaseURL(baseURL),
		logseq.WithToken(Token),
	)
}
