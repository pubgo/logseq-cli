package cmds

import (
	"context"
	"fmt"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

func BlockCmd() *redant.Command {
	return &redant.Command{
		Use:   "block",
		Short: "Block management",
		Children: []*redant.Command{
			blockGetCmd(),
			blockInsertCmd(),
			blockUpdateCmd(),
			blockRemoveCmd(),
			blockMoveCmd(),
			blockPrependCmd(),
			blockAppendCmd(),
		},
	}
}

func blockGetCmd() *redant.Command {
	var includeChildren bool
	return &redant.Command{
		Use:   "get <uuid>",
		Short: "Get a block by UUID",
		Options: redant.OptionSet{
			{
				Flag:        "children",
				Shorthand:   "c",
				Description: "Include child blocks",
				Value:       redant.BoolOf(&includeChildren),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) == 0 {
				return fmt.Errorf("block UUID required")
			}
			client := NewClient()
			block, err := client.GetBlock(ctx, inv.Args[0], includeChildren)
			if err != nil {
				return err
			}
			if block == nil {
				return fmt.Errorf("block '%s' not found", inv.Args[0])
			}
			return PrintJSON(block)
		},
	}
}

func blockInsertCmd() *redant.Command {
	var sibling bool
	return &redant.Command{
		Use:   "insert <target-uuid> <content>",
		Short: "Insert a block",
		Options: redant.OptionSet{
			{
				Flag:        "sibling",
				Shorthand:   "s",
				Description: "Insert as sibling (default: child)",
				Value:       redant.BoolOf(&sibling),
			},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) < 2 {
				return fmt.Errorf("target UUID and content required")
			}
			client := NewClient()
			block, err := client.InsertBlock(ctx, inv.Args[0], inv.Args[1], &logseq.InsertBlockOptions{
				Sibling: sibling,
			})
			if err != nil {
				return err
			}
			return PrintJSON(block)
		},
	}
}

func blockUpdateCmd() *redant.Command {
	return &redant.Command{
		Use:   "update <uuid> <content>",
		Short: "Update block content",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) < 2 {
				return fmt.Errorf("block UUID and content required")
			}
			client := NewClient()
			if err := client.UpdateBlock(ctx, inv.Args[0], inv.Args[1]); err != nil {
				return err
			}
			fmt.Fprintf(inv.Stdout, "updated block: %s\n", inv.Args[0])
			return nil
		},
	}
}

func blockRemoveCmd() *redant.Command {
	return &redant.Command{
		Use:   "remove <uuid>",
		Short: "Remove a block",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) == 0 {
				return fmt.Errorf("block UUID required")
			}
			client := NewClient()
			if err := client.RemoveBlock(ctx, inv.Args[0]); err != nil {
				return err
			}
			fmt.Fprintf(inv.Stdout, "removed block: %s\n", inv.Args[0])
			return nil
		},
	}
}

func blockMoveCmd() *redant.Command {
	return &redant.Command{
		Use:   "move <src-uuid> <target-uuid>",
		Short: "Move a block",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) < 2 {
				return fmt.Errorf("source UUID and target UUID required")
			}
			client := NewClient()
			if err := client.MoveBlock(ctx, inv.Args[0], inv.Args[1], nil); err != nil {
				return err
			}
			fmt.Fprintf(inv.Stdout, "moved block: %s -> %s\n", inv.Args[0], inv.Args[1])
			return nil
		},
	}
}

func blockPrependCmd() *redant.Command {
	return &redant.Command{
		Use:   "prepend <page> <content>",
		Short: "Prepend block to page",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) < 2 {
				return fmt.Errorf("page name and content required")
			}
			client := NewClient()
			block, err := client.PrependBlockInPage(ctx, inv.Args[0], inv.Args[1])
			if err != nil {
				return err
			}
			return PrintJSON(block)
		},
	}
}

func blockAppendCmd() *redant.Command {
	return &redant.Command{
		Use:   "append <page> <content>",
		Short: "Append block to page",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) < 2 {
				return fmt.Errorf("page name and content required")
			}
			client := NewClient()
			block, err := client.AppendBlockInPage(ctx, inv.Args[0], inv.Args[1])
			if err != nil {
				return err
			}
			return PrintJSON(block)
		},
	}
}
