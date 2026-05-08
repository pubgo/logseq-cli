package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
			pageRefsCmd(),
			pageNamespaceCmd(),
			pagePropertiesCmd(),
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
	var content string
	return &redant.Command{
		Use:   "create <name>",
		Short: "Create a page",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "content",
				Shorthand:   "c",
				Description: "Initial block content (use '-' to read from stdin)",
				Value:       redant.StringOf(&content),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Page, error) {
			client := NewClient()
			page, err := client.CreatePage(ctx, inv.Args[0], nil, &logseq.CreatePageOptions{
				CreateFirstBlock: true,
			})
			if err != nil {
				return nil, err
			}
			if content != "" {
				body, err := readContent(inv, content)
				if err != nil {
					return nil, err
				}
				if _, err := client.AppendBlockInPage(ctx, inv.Args[0], body); err != nil {
					return nil, err
				}
			}
			return page, nil
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

func pageRefsCmd() *redant.Command {
	return &redant.Command{
		Use:   "refs <name>",
		Short: "Get backlinks (linked references) for a page",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			client := NewClient()
			result, err := client.GetPageLinkedReferences(ctx, inv.Args[0])
			if err != nil {
				return err
			}
			enc := json.NewEncoder(inv.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(json.RawMessage(result))
		},
	}
}

func pageNamespaceCmd() *redant.Command {
	var tree bool
	return &redant.Command{
		Use:   "namespace <name>",
		Short: "List pages in a namespace",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Namespace prefix"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "tree",
				Description: "Return as tree structure",
				Value:       redant.BoolOf(&tree),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			client := NewClient()
			enc := json.NewEncoder(inv.Stdout)
			enc.SetIndent("", "  ")

			if tree {
				result, err := client.GetPagesTreeFromNamespace(ctx, inv.Args[0])
				if err != nil {
					return err
				}
				return enc.Encode(json.RawMessage(result))
			}

			pages, err := client.GetPagesFromNamespace(ctx, inv.Args[0])
			if err != nil {
				return err
			}
			return enc.Encode(pages)
		},
	}
}

func pagePropertiesCmd() *redant.Command {
	return &redant.Command{
		Use:   "properties <name> [key=value ...]",
		Short: "Get or set page properties",
		Long:  "Without key=value args, prints current properties. With args, sets them.",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			client := NewClient()
			name := inv.Args[0]

			// If extra args provided, treat as key=value pairs to set
			if len(inv.Args) > 1 {
				props := make(map[string]any)
				for _, kv := range inv.Args[1:] {
					parts := strings.SplitN(kv, "=", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid property format %q, expected key=value", kv)
					}
					props[parts[0]] = parts[1]
				}
				if err := client.SetPageProperties(ctx, name, props); err != nil {
					return err
				}
				enc := json.NewEncoder(inv.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(StatusResult{OK: true, Message: "properties updated"})
			}

			// No extra args: get page and print properties
			page, err := client.GetPage(ctx, name)
			if err != nil {
				return err
			}
			if page == nil {
				return fmt.Errorf("page '%s' not found", name)
			}
			enc := json.NewEncoder(inv.Stdout)
			enc.SetIndent("", "  ")
			if page.Properties == nil {
				return enc.Encode(map[string]any{})
			}
			return enc.Encode(page.Properties)
		},
	}
}
