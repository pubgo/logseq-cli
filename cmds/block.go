package cmds

import (
	"context"
	"encoding/json"
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
			blockPropertyCmd(),
			blockCollapseCmd(),
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
		Short: "Insert a block (use '-' as content to read from stdin)",
		Args: redant.ArgSet{
			{Name: "target-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Target block UUID"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content (use '-' for stdin)"},
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
			content, err := readContent(inv, inv.Args[1])
			if err != nil {
				return nil, err
			}
			client := NewClient()
			return client.InsertBlock(ctx, inv.Args[0], content, &logseq.InsertBlockOptions{
				Sibling: sibling,
			})
		}),
	}
}

func blockUpdateCmd() *redant.Command {
	return &redant.Command{
		Use:   "update <uuid> <content>",
		Short: "Update block content (use '-' as content to read from stdin)",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "New block content (use '-' for stdin)"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			content, err := readContent(inv, inv.Args[1])
			if err != nil {
				return StatusResult{}, err
			}
			client := NewClient()
			if err := client.UpdateBlock(ctx, inv.Args[0], content); err != nil {
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
	var before bool
	return &redant.Command{
		Use:   "move <src-uuid> <target-uuid>",
		Short: "Move a block",
		Args: redant.ArgSet{
			{Name: "src-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Source block UUID"},
			{Name: "target-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Target block UUID"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "before",
				Description: "Move before the target (default: after)",
				Value:       redant.BoolOf(&before),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			var opts map[string]any
			if before {
				opts = map[string]any{"before": true}
			}
			client := NewClient()
			if err := client.MoveBlock(ctx, inv.Args[0], inv.Args[1], opts); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "moved block: " + inv.Args[0] + " -> " + inv.Args[1]}, nil
		}),
	}
}

func blockPrependCmd() *redant.Command {
	return &redant.Command{
		Use:   "prepend <page> <content>",
		Short: "Prepend block to page (use '-' as content to read from stdin)",
		Args: redant.ArgSet{
			{Name: "page", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content (use '-' for stdin)"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			content, err := readContent(inv, inv.Args[1])
			if err != nil {
				return nil, err
			}
			client := NewClient()
			return client.PrependBlockInPage(ctx, inv.Args[0], content)
		}),
	}
}

func blockAppendCmd() *redant.Command {
	return &redant.Command{
		Use:   "append <page> <content>",
		Short: "Append block to page (use '-' as content to read from stdin)",
		Args: redant.ArgSet{
			{Name: "page", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content (use '-' for stdin)"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			content, err := readContent(inv, inv.Args[1])
			if err != nil {
				return nil, err
			}
			client := NewClient()
			return client.AppendBlockInPage(ctx, inv.Args[0], content)
		}),
	}
}

func blockPropertyCmd() *redant.Command {
	return &redant.Command{
		Use:   "property",
		Short: "Block property operations",
		Children: []*redant.Command{
			{
				Use:   "get <uuid>",
				Short: "Get all properties of a block",
				Args: redant.ArgSet{
					{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]any, error) {
					client := NewClient()
					return client.GetBlockProperties(ctx, inv.Args[0])
				}),
			},
			{
				Use:   "set <uuid> <key> <value>",
				Short: "Set a block property",
				Args: redant.ArgSet{
					{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "Property key"},
					{Name: "value", Required: true, Value: redant.StringOf(new(string)), Description: "Property value"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					var value any = inv.Args[2]
					// Try to parse as JSON for structured values
					var parsed any
					if err := json.Unmarshal([]byte(inv.Args[2]), &parsed); err == nil {
						value = parsed
					}
					if err := client.UpsertBlockProperty(ctx, inv.Args[0], inv.Args[1], value); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "set " + inv.Args[1] + " on block " + inv.Args[0]}, nil
				}),
			},
			{
				Use:   "remove <uuid> <key>",
				Short: "Remove a block property",
				Args: redant.ArgSet{
					{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "Property key"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.RemoveBlockProperty(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "removed " + inv.Args[1] + " from block " + inv.Args[0]}, nil
				}),
			},
		},
	}
}

func blockCollapseCmd() *redant.Command {
	var expand bool
	return &redant.Command{
		Use:   "collapse <uuid>",
		Short: "Collapse or expand a block",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "expand",
				Shorthand:   "e",
				Description: "Expand instead of collapse",
				Value:       redant.BoolOf(&expand),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			collapsed := !expand
			if err := client.SetBlockCollapsed(ctx, inv.Args[0], collapsed); err != nil {
				return StatusResult{}, err
			}
			action := "collapsed"
			if expand {
				action = "expanded"
			}
			return StatusResult{OK: true, Message: action + " block: " + inv.Args[0]}, nil
		}),
	}
}
