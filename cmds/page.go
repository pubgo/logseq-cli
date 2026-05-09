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

func PageCmd() *redant.Command {
	return &redant.Command{
		Use:   "page",
		Short: "Page management",
		Children: []*redant.Command{
			pageListCmd(),
			pageCurrentCmd(),
			pageCurrentTreeCmd(),
			pageGetCmd(),
			pageCreateCmd(),
			pageAppendSafeCmd(),
			pageJournalCmd(),
			pageDeleteCmd(),
			pageRenameCmd(),
			pageRefsCmd(),
			pageNamespaceCmd(),
			pagePropertiesCmd(),
		},
	}
}

func pageAppendSafeCmd() *redant.Command {
	var (
		dryRun         bool
		confirm        bool
		idempotencyKey string
	)

	return &redant.Command{
		Use:   "append-safe <name> <content>",
		Short: "Safely append block to page (supports dry-run / idempotency)",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Page name"},
			{Name: "content", Required: true, Value: redant.StringOf(new(string)), Description: "Block content (use '-' for stdin)"},
		},
		Options: redant.OptionSet{
			{Flag: "dry-run", Description: "Preview action without writing", Value: redant.BoolOf(&dryRun)},
			{Flag: "confirm", Description: "Required in some write policies", Value: redant.BoolOf(&confirm)},
			{Flag: "idempotency-key", Description: "Idempotency key to avoid duplicate writes", Value: redant.StringOf(&idempotencyKey)},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*llmEnvelope, error) {
			start := time.Now()
			pageName := strings.TrimSpace(inv.Args[0])
			if pageName == "" {
				return envelopeFailure(start, fmt.Errorf("page name is required"), "BAD_REQUEST", "请提供页面名"), nil
			}

			content, err := readContent(inv, inv.Args[1])
			if err != nil {
				return envelopeFailure(start, err, "BAD_REQUEST", "content 参数错误"), nil
			}
			content = strings.TrimSpace(content)
			if content == "" {
				return envelopeFailure(start, fmt.Errorf("content is required"), "BAD_REQUEST", "请提供非空内容"), nil
			}

			if err := ensureWriteAllowed("page.append-safe", dryRun, confirm, false); err != nil {
				return envelopeFailure(start, err, "SAFETY_BLOCKED", "可先使用 --dry-run 或调整 LOGSEQ_LLM_WRITE_MODE"), nil
			}

			client := NewClient()
			page, err := client.GetPage(ctx, pageName)
			if err != nil {
				return envelopeFailure(start, err, "UPSTREAM_ERROR", "请先确认 API 连接正常", withCapabilityUsed("logseq.Editor.getPage")), nil
			}
			if page == nil {
				return envelopeFailure(start, fmt.Errorf("page '%s' not found", pageName), "RESOURCE_NOT_FOUND", "请确认页面存在"), nil
			}

			trimmedKey := strings.TrimSpace(idempotencyKey)
			if !dryRun && trimmedKey != "" {
				if rec, ok := getAppendSafeRecord(trimmedKey); ok {
					payload := map[string]any{
						"action":          "deduped",
						"page":            pageName,
						"idempotency_key": trimmedKey,
						"block_uuid":      rec.BlockUUID,
						"deduped":         true,
						"write_mode":      currentLLMWriteMode(),
						"message":         "duplicate idempotency key detected, skipped",
						"created_at":      rec.CreatedAt.Format(time.RFC3339),
					}
					return envelopeSuccess(start, payload, withCapabilityUsed("logseq.Editor.getPage"), withHints("幂等键已命中，未重复写入")), nil
				}
			}

			if dryRun {
				payload := map[string]any{
					"action":          "dry-run",
					"page":            pageName,
					"idempotency_key": trimmedKey,
					"content_preview": truncate(content, 240),
					"dry_run":         true,
					"write_mode":      currentLLMWriteMode(),
					"message":         "append preview only, nothing written",
				}
				return envelopeSuccess(start, payload, withCapabilityUsed("logseq.Editor.getPage"), withHints("dry-run 模式未执行写入")), nil
			}

			block, err := client.AppendBlockInPage(ctx, pageName, content)
			if err != nil {
				return envelopeFailure(start, err, "UPSTREAM_ERROR", "写入失败，请检查页面权限和 API 状态", withCapabilityUsed("logseq.Editor.appendBlockInPage")), nil
			}

			if trimmedKey != "" && block != nil {
				setAppendSafeRecord(trimmedKey, appendSafeRecord{
					BlockUUID:      strings.TrimSpace(block.UUID),
					Page:           pageName,
					ContentPreview: truncate(content, 240),
					CreatedAt:      time.Now(),
				})
			}

			payload := map[string]any{
				"action":          "appended",
				"page":            pageName,
				"idempotency_key": trimmedKey,
				"block_uuid":      block.UUID,
				"deduped":         false,
				"write_mode":      currentLLMWriteMode(),
				"message":         "block appended",
			}
			return envelopeSuccess(start, payload, withCapabilityUsed("logseq.Editor.getPage", "logseq.Editor.appendBlockInPage")), nil
		}),
	}
}

func pageCurrentCmd() *redant.Command {
	return &redant.Command{
		Use:   "current",
		Short: "Get currently focused page",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Page, error) {
			client := NewClient()
			page, err := client.GetCurrentPage(ctx)
			if err != nil {
				return nil, err
			}
			if page == nil {
				return nil, fmt.Errorf("no current page")
			}
			return page, nil
		}),
	}
}

func pageCurrentTreeCmd() *redant.Command {
	return &redant.Command{
		Use:   "current-tree",
		Short: "Get block tree of currently focused page",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]logseq.Block, error) {
			client := NewClient()
			return client.GetCurrentPageBlocksTree(ctx)
		}),
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

func pageJournalCmd() *redant.Command {
	return &redant.Command{
		Use:   "journal [date]",
		Short: "Create journal page for date (default: today, format YYYY-MM-DD)",
		Args: redant.ArgSet{
			{Name: "date", Required: false, Value: redant.StringOf(new(string)), Description: "Date string in YYYY-MM-DD"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Page, error) {
			date := time.Now().Format("2006-01-02")
			if len(inv.Args) > 0 && strings.TrimSpace(inv.Args[0]) != "" {
				date = inv.Args[0]
			}

			client := NewClient()
			return client.CreateJournalPage(ctx, date)
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

			enc := json.NewEncoder(inv.Stdout)
			enc.SetIndent("", "  ")

			// Prefer official API when available.
			props, err := client.GetPageProperties(ctx, name)
			if err == nil {
				if props == nil {
					return enc.Encode(map[string]any{})
				}
				return enc.Encode(props)
			}

			// Backward compatibility fallback for versions without getPageProperties.
			if !strings.Contains(strings.ToLower(err.Error()), "methodnotexist") {
				return err
			}

			page, getErr := client.GetPage(ctx, name)
			if getErr != nil {
				return getErr
			}
			if page == nil || page.Properties == nil {
				return enc.Encode(map[string]any{})
			}
			return enc.Encode(page.Properties)
		},
	}
}
