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
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "children",
				Shorthand:   "c",
				Description: "Include child blocks",
				Value:       redant.BoolOf(&includeChildren),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			block, err := client.GetBlock(ctx, inv.Args[0], includeChildren)
			if err != nil {
				return nil, err
			}
			if block == nil {
				return nil, fmt.Errorf("block '%s' not found", inv.Args[0])
			}
			return block, nil
		}),
	}
}

func blockInsertCmd() *redant.Command {
	var sibling bool
	return &redant.Command{
		Use:   "insert <target-uuid> <content>",
		Short: "Insert a block",
		Args: redant.ArgSet{
			{Name: "target-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Target block UUID"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "sibling",
				Shorthand:   "s",
				Description: "Insert as sibling (default: child)",
				Value:       redant.BoolOf(&sibling),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			return client.InsertBlock(ctx, inv.Args[0], inv.Args[1], &logseq.InsertBlockOptions{
				Sibling: sibling,
			})
		}),
	}
}

func blockUpdateCmd() *redant.Command {
	return &redant.Command{
		Use:   "update <uuid> <content>",
		Short: "Update block content",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "New block content"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			if err := client.UpdateBlock(ctx, inv.Args[0], inv.Args[1]); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "updated block: " + inv.Args[0]}, nil
		}),
	}
}

func blockRemoveCmd() *redant.Command {
	return &redant.Command{
		Use:   "remove <uuid>",
		Short: "Remove a block",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			if err := client.RemoveBlock(ctx, inv.Args[0]); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "removed block: " + inv.Args[0]}, nil
		}),
	}
}

func blockMoveCmd() *redant.Command {
	return &redant.Command{
		Use:   "move <src-uuid> <target-uuid>",
		Short: "Move a block",
		Args: redant.ArgSet{
			{Name: "src-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Source block UUID"},
			{Name: "target-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Target block UUID"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			if err := client.MoveBlock(ctx, inv.Args[0], inv.Args[1], nil); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "moved block: " + inv.Args[0] + " -> " + inv.Args[1]}, nil
		}),
	}
}

func blockPrependCmd() *redant.Command {
	return &redant.Command{
		Use:   "prepend <page> <content>",
		Short: "Prepend block to page",
		Args: redant.ArgSet{
			{Name: "page", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			return client.PrependBlockInPage(ctx, inv.Args[0], inv.Args[1])
		}),
	}
}

func blockAppendCmd() *redant.Command {
	return &redant.Command{
		Use:   "append <page> <content>",
		Short: "Append block to page",
		Args: redant.ArgSet{
			{Name: "page", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			return client.AppendBlockInPage(ctx, inv.Args[0], inv.Args[1])
		}),
	}
}
