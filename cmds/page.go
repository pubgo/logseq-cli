package cmds

import (
	"context"
	"fmt"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

func PageCmd() *redant.Command {
	return &redant.Command{
		Use:   "page",
		Short: "Page management",
		Children: []*redant.Command{
			pageListCmd(),
			pageGetCmd(),
			pageCreateCmd(),
			pageDeleteCmd(),
			pageRenameCmd(),
		},
	}
}

func pageListCmd() *redant.Command {
	return &redant.Command{
		Use:   "list",
		Short: "List all pages",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			client := NewClient()
			pages, err := client.GetAllPages(ctx)
			if err != nil {
				return err
			}
			return PrintJSON(pages)
		},
	}
}

func pageGetCmd() *redant.Command {
	var withBlocks bool
	return &redant.Command{
		Use:   "get <name>",
		Short: "Get page info",
		Options: redant.OptionSet{
			{
				Flag:        "blocks",
				Shorthand:   "b",
				Description: "Include page block tree",
				Value:       redant.BoolOf(&withBlocks),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) == 0 {
				return fmt.Errorf("page name required")
			}
			client := NewClient()
			name := inv.Args[0]

			if withBlocks {
				blocks, err := client.GetPageBlocksTree(ctx, name)
				if err != nil {
					return err
				}
				return PrintJSON(blocks)
			}

			page, err := client.GetPage(ctx, name)
			if err != nil {
				return err
			}
			if page == nil {
				return fmt.Errorf("page '%s' not found", name)
			}
			return PrintJSON(page)
		},
	}
}

func pageCreateCmd() *redant.Command {
	return &redant.Command{
		Use:   "create <name>",
		Short: "Create a page",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) == 0 {
				return fmt.Errorf("page name required")
			}
			client := NewClient()
			page, err := client.CreatePage(ctx, inv.Args[0], nil, &logseq.CreatePageOptions{
				CreateFirstBlock: true,
			})
			if err != nil {
				return err
			}
			return PrintJSON(page)
		},
	}
}

func pageDeleteCmd() *redant.Command {
	return &redant.Command{
		Use:   "delete <name>",
		Short: "Delete a page",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) == 0 {
				return fmt.Errorf("page name required")
			}
			client := NewClient()
			if err := client.DeletePage(ctx, inv.Args[0]); err != nil {
				return err
			}
			fmt.Fprintf(inv.Stdout, "deleted page: %s\n", inv.Args[0])
			return nil
		},
	}
}

func pageRenameCmd() *redant.Command {
	return &redant.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a page",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) < 2 {
				return fmt.Errorf("old name and new name required")
			}
			client := NewClient()
			if err := client.RenamePage(ctx, inv.Args[0], inv.Args[1]); err != nil {
				return err
			}
			fmt.Fprintf(inv.Stdout, "renamed: %s -> %s\n", inv.Args[0], inv.Args[1])
			return nil
		},
	}
}
