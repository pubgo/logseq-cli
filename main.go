package main

import (
	"fmt"
	"os"

	"github.com/pubgo/logseq-cli/cmds"
	"github.com/pubgo/redant"
	"github.com/pubgo/redant/cmds/completioncmd"
	"github.com/pubgo/redant/cmds/doccmd"
	"github.com/pubgo/redant/cmds/llmstxtcmd"
	"github.com/pubgo/redant/cmds/mcpcmd"
	"github.com/pubgo/redant/cmds/webcmd"
)

func main() {
	root := redant.Command{
		Use:   "logseq",
		Short: "Logseq CLI - command line tool for Logseq",
		Options: redant.OptionSet{
			{
				Flag:        "token",
				Shorthand:   "t",
				Description: "Logseq API token",
				Envs:        []string{"LOGSEQ_API_TOKEN"},
				Required:    true,
				Value:       redant.StringOf(&cmds.Token),
				Inherit:     true,
			},
			{
				Flag:        "host",
				Description: "Logseq API host",
				Envs:        []string{"LOGSEQ_HOST"},
				Default:     "127.0.0.1",
				Value:       redant.StringOf(&cmds.Host),
				Inherit:     true,
			},
			{
				Flag:        "port",
				Shorthand:   "p",
				Description: "Logseq API port",
				Envs:        []string{"LOGSEQ_PORT"},
				Default:     "12315",
				Value:       redant.StringOf(&cmds.Port),
				Inherit:     true,
			},
		},
		Children: []*redant.Command{
			cmds.PageCmd(),
			cmds.BlockCmd(),
			cmds.GraphCmd(),
			cmds.QueryCmd(),
			cmds.SearchCmd(),
			cmds.TagCmd(),
			cmds.WebUICmd(),
			llmstxtcmd.New(),
			doccmd.New(),
		},
	}

	webcmd.AddWebCommand(&root)
	mcpcmd.AddMCPCommand(&root)
	completioncmd.AddCompletionCommand(&root)
	if err := root.Invoke().WithOS().Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
