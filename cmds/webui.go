package cmds

import (
	"context"
	"fmt"

	"github.com/pubgo/logseq-cli/internal/webui"
	"github.com/pubgo/redant"
)

func WebUICmd() *redant.Command {
	var addr string
	var open bool

	return &redant.Command{
		Use:   "webui",
		Short: "Start a simple web UI for Logseq data operations",
		Options: redant.OptionSet{
			{
				Flag:        "addr",
				Description: "HTTP listen address",
				Default:     "127.0.0.1:18090",
				Value:       redant.StringOf(&addr),
			},
			{
				Flag:        "open",
				Description: "Open browser automatically",
				Default:     "true",
				Value:       redant.BoolOf(&open),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			connInfo := CurrentClientConnectionInfo()
			fmt.Fprintf(inv.Stdout, "[webui] listening on http://%s\n", addr)
			fmt.Fprintf(inv.Stdout, "[webui] logseq api: %s (host=%s, port=%s)\n", connInfo.BaseURL, connInfo.HostSource, connInfo.PortSource)
			fmt.Fprintln(inv.Stdout, "[webui] press Ctrl+C to stop")
			return webui.Run(ctx, addr, open, NewClient(), connInfo)
		},
	}
}
