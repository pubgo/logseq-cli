package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

func BlockCmd() *redant.Command {
	return &redant.Command{
		Use:   "block",
		Short: "Block management",
		Children: []*redant.Command{
			blockCurrentCmd(),
			blockSelectedCmd(),
			blockClearSelectedCmd(),
			blockNewUUIDCmd(),
			blockGetCmd(),
			blockPrevSiblingCmd(),
			blockNextSiblingCmd(),
			blockInsertCmd(),
			blockInsertBatchCmd(),
			blockUpdateCmd(),
			blockDeleteSafeCmd(),
			blockRemoveCmd(),
			blockMoveCmd(),
			blockPrependCmd(),
			blockAppendCmd(),
			blockPropertyCmd(),
			blockCollapseCmd(),
		},
	}
}

func blockDeleteSafeCmd() *redant.Command {
	var (
		dryRun  bool
		confirm bool
	)

	return &redant.Command{
		Use:   "delete-safe <uuid>",
		Short: "Safely delete a block (dry-run + confirm)",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
		},
		Options: redant.OptionSet{
			{Flag: "dry-run", Description: "Preview impact without deleting", Value: redant.BoolOf(&dryRun)},
			{Flag: "confirm", Description: "Required for dangerous delete in most policies", Value: redant.BoolOf(&confirm)},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*llmEnvelope, error) {
			start := time.Now()
			uuid := strings.TrimSpace(inv.Args[0])
			if uuid == "" {
				return envelopeFailure(start, fmt.Errorf("uuid is required"), "BAD_REQUEST", "请提供块 UUID"), nil
			}

			client := NewClient()
			block, err := client.GetBlock(ctx, uuid, true)
			if err != nil {
				return envelopeFailure(start, err, "", "请先确认块可读取", withCapabilityUsed("logseq.Editor.getBlock")), nil
			}
			if block == nil {
				return envelopeFailure(start, fmt.Errorf("block '%s' not found", uuid), "RESOURCE_NOT_FOUND", "请确认 UUID 正确"), nil
			}

			subtreeCount := countBlockSubtree(block)
			preview := truncate(block.Content, 240)

			if err := ensureWriteAllowed("block.delete-safe", dryRun, confirm, true); err != nil {
				return envelopeFailure(start, err, "SAFETY_BLOCKED", "可先使用 --dry-run 或显式 --confirm"), nil
			}

			if dryRun {
				payload := map[string]any{
					"action":          "dry-run",
					"uuid":            uuid,
					"subtree_count":   subtreeCount,
					"content_preview": preview,
					"write_mode":      currentLLMWriteMode(),
					"message":         "delete preview only, nothing deleted",
				}
				return envelopeSuccess(start, payload, withCapabilityUsed("logseq.Editor.getBlock"), withHints("dry-run 模式未执行删除")), nil
			}

			if err := client.RemoveBlock(ctx, uuid); err != nil {
				return envelopeFailure(start, err, "UPSTREAM_ERROR", "删除失败，请检查块状态后重试", withCapabilityUsed("logseq.Editor.removeBlock")), nil
			}

			payload := map[string]any{
				"action":          "deleted",
				"uuid":            uuid,
				"subtree_count":   subtreeCount,
				"content_preview": preview,
				"write_mode":      currentLLMWriteMode(),
				"message":         "block deleted",
			}
			return envelopeSuccess(start, payload, withCapabilityUsed("logseq.Editor.getBlock", "logseq.Editor.removeBlock")), nil
		}),
	}
}

func countBlockSubtree(b *logseq.Block) int {
	if b == nil {
		return 0
	}
	count := 1
	for i := range b.Children {
		child := b.Children[i]
		count += countBlockSubtree(&child)
	}
	return count
}

func blockSelectedCmd() *redant.Command {
	return &redant.Command{
		Use:   "selected",
		Short: "Get currently selected blocks",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]logseq.Block, error) {
			client := NewClient()
			return client.GetSelectedBlocks(ctx)
		}),
	}
}

func blockClearSelectedCmd() *redant.Command {
	return &redant.Command{
		Use:   "clear-selected",
		Short: "Clear currently selected blocks",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			client := NewClient()
			if err := client.ClearSelectedBlocks(ctx); err != nil {
				return StatusResult{}, err
			}
			return StatusResult{OK: true, Message: "selection cleared"}, nil
		}),
	}
}

func blockNewUUIDCmd() *redant.Command {
	return &redant.Command{
		Use:   "new-uuid",
		Short: "Create a new block UUID",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]string, error) {
			client := NewClient()
			uuid, err := client.NewBlockUUID(ctx)
			if err != nil {
				return nil, err
			}
			return map[string]string{"uuid": uuid}, nil
		}),
	}
}

func blockCurrentCmd() *redant.Command {
	return &redant.Command{
		Use:   "current",
		Short: "Get currently focused block",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			block, err := client.GetCurrentBlock(ctx)
			if err != nil {
				return nil, err
			}
			if block == nil {
				return nil, fmt.Errorf("no current block")
			}
			return block, nil
		}),
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

func blockPrevSiblingCmd() *redant.Command {
	return &redant.Command{
		Use:   "prev-sibling <uuid>",
		Short: "Get previous sibling block",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			block, err := client.GetPreviousSiblingBlock(ctx, inv.Args[0])
			if err != nil {
				return nil, err
			}
			if block == nil {
				return nil, fmt.Errorf("previous sibling not found for block '%s'", inv.Args[0])
			}
			return block, nil
		}),
	}
}

func blockNextSiblingCmd() *redant.Command {
	return &redant.Command{
		Use:   "next-sibling <uuid>",
		Short: "Get next sibling block",
		Args: redant.ArgSet{
			{Name: "uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Block UUID"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Block, error) {
			client := NewClient()
			block, err := client.GetNextSiblingBlock(ctx, inv.Args[0])
			if err != nil {
				return nil, err
			}
			if block == nil {
				return nil, fmt.Errorf("next sibling not found for block '%s'", inv.Args[0])
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

func blockInsertBatchCmd() *redant.Command {
	var sibling bool
	return &redant.Command{
		Use:   "insert-batch <target-uuid> <blocks-json>",
		Short: "Insert multiple blocks (JSON array, use '-' to read from stdin)",
		Args: redant.ArgSet{
			{Name: "target-uuid", Required: true, Value: redant.StringOf(new(string)), Description: "Target block UUID"},
			{Name: "blocks-json", Required: true, Value: redant.StringOf(new(string)), Description: "JSON array of batch blocks (use '-' for stdin)"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "sibling",
				Shorthand:   "s",
				Description: "Insert as sibling (default: child)",
				Value:       redant.BoolOf(&sibling),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
			raw, err := readContent(inv, inv.Args[1])
			if err != nil {
				return StatusResult{}, err
			}

			var blocks []logseq.BatchBlock
			if err := json.Unmarshal([]byte(raw), &blocks); err != nil {
				return StatusResult{}, fmt.Errorf("invalid blocks json, expected []BatchBlock: %w", err)
			}
			if len(blocks) == 0 {
				return StatusResult{}, fmt.Errorf("blocks array cannot be empty")
			}

			client := NewClient()
			if err := client.InsertBatchBlock(ctx, inv.Args[0], blocks, sibling); err != nil {
				return StatusResult{}, err
			}

			return StatusResult{OK: true, Message: fmt.Sprintf("inserted %d blocks", len(blocks))}, nil
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
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (any, error) {
			content, err := readContent(inv, inv.Args[1])
			if err != nil {
				return nil, err
			}
			client := NewClient()
			if err := client.UpdateBlock(ctx, inv.Args[0], content); err != nil {
				return nil, err
			}
			// Fetch updated block for consistent response
			block, err := client.GetBlock(ctx, inv.Args[0], false)
			if err == nil && block != nil {
				return block, nil
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
