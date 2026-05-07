package cmds

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

var (
	Token  string
	Host   string
	Port   string
	Output string
)

func NewClient() *logseq.Client {
	baseURL := fmt.Sprintf("http://%s:%s", Host, Port)
	return logseq.NewClient(
		logseq.WithBaseURL(baseURL),
		logseq.WithToken(Token),
	)
}

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
