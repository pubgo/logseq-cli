package cmds

import (
	"context"
	"fmt"

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
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					client := NewClient()
					info, err := client.GetCurrentGraph(ctx)
					if err != nil {
						return err
					}
					return PrintJSON(info)
				},
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
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					if len(inv.Args) == 0 {
						return fmt.Errorf("datalog query required")
					}
					client := NewClient()
					result, err := client.DatascriptQuery(ctx, inv.Args[0])
					if err != nil {
						return err
					}
					fmt.Fprintln(inv.Stdout, string(result))
					return nil
				},
			},
			{
				Use:   "dsl <query>",
				Short: "Execute Logseq DSL query",
				Handler: func(ctx context.Context, inv *redant.Invocation) error {
					if len(inv.Args) == 0 {
						return fmt.Errorf("DSL query required")
					}
					client := NewClient()
					result, err := client.DSLQuery(ctx, inv.Args[0])
					if err != nil {
						return err
					}
					fmt.Fprintln(inv.Stdout, string(result))
					return nil
				},
			},
		},
	}
}

func SearchCmd() *redant.Command {
	return &redant.Command{
		Use:   "search <query>",
		Short: "Full-text search",
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			if len(inv.Args) == 0 {
				return fmt.Errorf("search query required")
			}
			client := NewClient()
			result, err := client.Search(ctx, inv.Args[0])
			if err != nil {
				return err
			}
			return PrintJSON(result)
		},
	}
}
