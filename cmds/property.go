package cmds

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pubgo/redant"
)

func PropertyCmd() *redant.Command {
	return &redant.Command{
		Use:   "property",
		Short: "Property schema operations",
		Children: []*redant.Command{
			{
				Use:   "list",
				Short: "List all properties",
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (any, error) {
					client := NewClient()
					return client.GetAllProperties(ctx)
				}),
			},
			{
				Use:   "get <key>",
				Short: "Get property by key",
				Args: redant.ArgSet{
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "Property key"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (any, error) {
					client := NewClient()
					return client.GetProperty(ctx, inv.Args[0])
				}),
			},
			{
				Use:   "upsert <key> [schema-json]",
				Short: "Create or update property schema",
				Long:  "schema-json should be a JSON object, e.g. '{\"type\":\"number\",\"cardinality\":\"one\"}'",
				Args: redant.ArgSet{
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "Property key"},
					{Name: "schema-json", Required: false, Value: redant.StringOf(new(string)), Description: "Schema JSON object"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (any, error) {
					client := NewClient()

					var schema map[string]any
					if len(inv.Args) > 1 && inv.Args[1] != "" {
						if err := json.Unmarshal([]byte(inv.Args[1]), &schema); err != nil {
							return nil, fmt.Errorf("invalid schema-json: %w", err)
						}
					}

					return client.UpsertProperty(ctx, inv.Args[0], schema, nil)
				}),
			},
			{
				Use:   "remove <key>",
				Short: "Remove property schema",
				Args: redant.ArgSet{
					{Name: "key", Required: true, Value: redant.StringOf(new(string)), Description: "Property key"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.RemoveProperty(ctx, inv.Args[0]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "property removed"}, nil
				}),
			},
		},
	}
}
