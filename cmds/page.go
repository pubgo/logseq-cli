package cmds

import (
	"context"
	"encoding/json"
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
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]logseq.Page, error) {
			client := NewClient()
			return client.GetAllPages(ctx)
		}),
	}
}

func pageGetCmd() *redant.Command {
	var withBlocks bool
	return &redant.Command{
		Use:   "get <name>",
		Short: "Get page info",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "blocks",
				Shorthand:   "b",
				Description: "Include page block tree",
				Value:       redant.BoolOf(&withBlocks),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			client := NewClient()
			name := inv.Args[0]

			enc := json.NewEncoder(inv.Stdout)
			enc.SetIndent("", "  ")

			if withBlocks {
				blocks, err := client.GetPageBlocksTree(ctx, name)
				if err != nil {
					return err
				}
				return enc.Encode(blocks)
			}

			page, err := client.GetPage(ctx, name)
			if err != nil {
				return err
			}
			if page == nil {
				return fmt.Errorf("page '%s' not found", name)
			}
			return enc.Encode(page)
		},
	}
}

func pageCreateCmd() *redant.Command {
	return &redant.Command{
		Use:   "create <name>",
		Short: "Create a page",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Page, error) {
			client := NewClient()
			return client.CreatePage(ctx, inv.Args[0], nil, &logseq.CreatePageOptions{
				CreateFirstBlock: true,
			})
		}),
	}
}

func pageDeleteCmd() *redant.Command {
	return &redant.Command{
		Use:   "delete <name>",
		Short: "Delete a page",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			if err := client.DeletePage(ctx, inv.Args[0]); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "deleted page: " + inv.Args[0]}, nil
		}),
	}
}

func pageRenameCmd() *redant.Command {
	return &redant.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a page",
		Args: redant.ArgSet{
			{Name: "old-name", Required: true, Value: redant.StringOf(new(string)), Description: "Current page name"},
			{Name: "new-name", Required: true, Value: redant.StringOf(new(string)), Description: "New page name"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			if err := client.RenamePage(ctx, inv.Args[0], inv.Args[1]); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "renamed: " + inv.Args[0] + " -> " + inv.Args[1]}, nil
		}),
	}
}
