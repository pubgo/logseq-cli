package cmds

import (
	"context"

	"github.com/pubgo/redant"
)

func TagCmd() *redant.Command {
	return &redant.Command{
		Use:   "tag",
		Short: "Tag operations",
		Children: []*redant.Command{
			tagListCmd(),
		},
	}
}

func tagListCmd() *redant.Command {
	return &redant.Command{
		Use:   "list",
		Short: "List all tags",
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]string, error) {
			client := NewClient()
			return client.GetAllTags(ctx)
		}),
	}
}
