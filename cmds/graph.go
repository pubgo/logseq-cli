package cmds

import (
	"context"
	"encoding/json"

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
					var buf json.RawMessage = result
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
					var buf json.RawMessage = result
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
