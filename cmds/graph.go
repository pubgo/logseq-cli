package cmds

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

func GraphCmd() *redant.Command {
	return &redant.Command{
		Use:   "graph",
		Short: "Graph operations",
		Children: []*redant.Command{
			{
				Use:   "info",
				Short: "Get current graph info",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.GraphInfo, error) {
					client := NewClient()
					return client.GetCurrentGraph(ctx)
				}),
			},
			{
				Use:   "config",
				Short: "Get user configuration",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]any, error) {
					client := NewClient()
					return client.GetUserConfigs(ctx)
				}),
			},
			{
				Use:   "app-info",
				Short: "Get app info",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]any, error) {
					client := NewClient()
					return client.GetInfo(ctx)
				}),
			},
			{
				Use:   "user-info",
				Short: "Get app user info",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]any, error) {
					client := NewClient()
					return client.GetUserInfo(ctx)
				}),
			},
			{
				Use:   "db-graph",
				Short: "Check if current graph is DB graph",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]bool, error) {
					client := NewClient()
					ok, err := client.CheckCurrentIsDBGraph(ctx)
					if err != nil {
						return nil, err
					}
					return map[string]bool{"dbGraph": ok}, nil
				}),
			},
			{
				Use:   "graph-config",
				Short: "Get current graph configs",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (any, error) {
					client := NewClient()
					return client.GetCurrentGraphConfigs(ctx)
				}),
			},
			{
				Use:   "favorites",
				Short: "Get current graph favorites",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]any, error) {
					client := NewClient()
					return client.GetCurrentGraphFavorites(ctx)
				}),
			},
			{
				Use:   "recent",
				Short: "Get current graph recent items",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]any, error) {
					client := NewClient()
					return client.GetCurrentGraphRecent(ctx)
				}),
			},
			{
				Use:   "templates",
				Short: "Get current graph templates",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (map[string]any, error) {
					client := NewClient()
					return client.GetCurrentGraphTemplates(ctx)
				}),
			},
			{
				Use:   "state <key>",
				Short: "Get app state value from store by key",
				Args: redant.ArgSet{
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "State store key"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (any, error) {
					client := NewClient()
					return client.GetStateFromStore(ctx, inv.Args[0])
				}),
			},
			{
				Use:   "state-set <key> <value>",
				Short: "Set app state value in store by key",
				Args: redant.ArgSet{
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "State store key"},
					{Name: "value", Required: true, Value: redant.StringOf(new(string)), Description: "State value (JSON literal or plain string)"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()

					var value any = inv.Args[1]
					var parsed any
					if err := json.Unmarshal([]byte(inv.Args[1]), &parsed); err == nil {
						value = parsed
					}

					if err := client.SetStateFromStore(ctx, inv.Args[0], value); err != nil {
						return StatusResult{}, fmt.Errorf("set state failed: %w", err)
					}

					return StatusResult{OK: true, Message: "state updated"}, nil
				}),
			},
		},
	}
}

func QueryCmd() *redant.Command {
	return &redant.Command{
		Use:   "query",
		Short: "Query operations",
		Children: []*redant.Command{
			{
				Use:   "datalog <query>",
				Short: "Execute Datalog query",
				Args: redant.ArgSet{
					{Name: "query", Required: true, Value: redant.StringOf(new(string)), Description: "Datalog query string"},
				},
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					client := NewClient()
					result, err := client.DatascriptQuery(ctx, inv.Args[0])
					if err != nil {
						return err
					}
					buf := json.RawMessage(result)
					enc := json.NewEncoder(inv.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(buf)
				},
			},
			{
				Use:   "dsl <query>",
				Short: "Execute Logseq DSL query",
				Args: redant.ArgSet{
					{Name: "query", Required: true, Value: redant.StringOf(new(string)), Description: "DSL query string"},
				},
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					client := NewClient()
					result, err := client.DSLQuery(ctx, inv.Args[0])
					if err != nil {
						return err
					}
					buf := json.RawMessage(result)
					enc := json.NewEncoder(inv.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(buf)
				},
			},
		},
	}
}

func SearchCmd() *redant.Command {
	return &redant.Command{
		Use:   "search <query>",
		Short: "Full-text search",
		Args: redant.ArgSet{
			{Name: "query", Required: true, Value: redant.StringOf(new(string)), Description: "Search query"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.SearchResult, error) {
			client := NewClient()
			return client.Search(ctx, inv.Args[0])
		}),
	}
}
