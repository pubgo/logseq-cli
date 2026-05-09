package cmds

import (
	"context"
	"fmt"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

func TagCmd() *redant.Command {
	return &redant.Command{
		Use:   "tag",
		Short: "Tag operations",
		Children: []*redant.Command{
			tagListCmd(),
			tagGetCmd(),
			tagSearchCmd(),
			tagCreateCmd(),
			tagObjectsCmd(),
			tagPropertyCmd(),
			tagExtendsCmd(),
			tagBlockCmd(),
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

func tagGetCmd() *redant.Command {
	return &redant.Command{
		Use:   "get <name-or-id>",
		Short: "Get tag by name or entity id",
		Args: redant.ArgSet{
			{Name: "name-or-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag name or numeric entity id"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Page, error) {
			client := NewClient()
			tag, err := client.GetTag(ctx, inv.Args[0])
			if err != nil {
				return nil, err
			}
			if tag == nil {
				return nil, fmt.Errorf("tag not found: %s", inv.Args[0])
			}
			return tag, nil
		}),
	}
}

func tagSearchCmd() *redant.Command {
	return &redant.Command{
		Use:   "search <name>",
		Short: "Search tags by name",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Tag name to search"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]logseq.Page, error) {
			client := NewClient()
			return client.GetTagsByName(ctx, inv.Args[0])
		}),
	}
}

func tagCreateCmd() *redant.Command {
	var uuid string
	return &redant.Command{
		Use:   "create <name>",
		Short: "Create a tag",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Tag name"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "uuid",
				Description: "Custom UUID for tag page",
				Value:       redant.StringOf(&uuid),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*logseq.Page, error) {
			client := NewClient()
			if uuid != "" {
				return client.CreateTag(ctx, inv.Args[0], uuid)
			}
			return client.CreateTag(ctx, inv.Args[0])
		}),
	}
}

func tagObjectsCmd() *redant.Command {
	return &redant.Command{
		Use:   "objects <name>",
		Short: "Get tag object blocks",
		Args: redant.ArgSet{
			{Name: "name", Required: true, Value: redant.StringOf(new(string)), Description: "Tag name"},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) ([]logseq.Block, error) {
			client := NewClient()
			return client.GetTagObjects(ctx, inv.Args[0])
		}),
	}
}

func tagPropertyCmd() *redant.Command {
	return &redant.Command{
		Use:   "property",
		Short: "Tag property relation operations",
		Children: []*redant.Command{
			{
				Use:   "add <tag-id> <property-id-or-name>",
				Short: "Add property relation to tag",
				Args: redant.ArgSet{
					{Name: "tag-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag id or name"},
					{Name: "property-id-or-name", Required: true, Value: redant.StringOf(new(string)), Description: "Property id or name"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.AddTagProperty(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "tag property relation added"}, nil
				}),
			},
			{
				Use:   "remove <tag-id> <property-id-or-name>",
				Short: "Remove property relation from tag",
				Args: redant.ArgSet{
					{Name: "tag-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag id or name"},
					{Name: "property-id-or-name", Required: true, Value: redant.StringOf(new(string)), Description: "Property id or name"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.RemoveTagProperty(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "tag property relation removed"}, nil
				}),
			},
		},
	}
}

func tagExtendsCmd() *redant.Command {
	return &redant.Command{
		Use:   "extends",
		Short: "Tag extends relation operations",
		Children: []*redant.Command{
			{
				Use:   "add <tag-id> <parent-tag-id-or-name>",
				Short: "Add parent tag relation",
				Args: redant.ArgSet{
					{Name: "tag-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag id or name"},
					{Name: "parent-tag-id-or-name", Required: true, Value: redant.StringOf(new(string)), Description: "Parent tag id or name"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.AddTagExtends(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "tag extends relation added"}, nil
				}),
			},
			{
				Use:   "remove <tag-id> <parent-tag-id-or-name>",
				Short: "Remove parent tag relation",
				Args: redant.ArgSet{
					{Name: "tag-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag id or name"},
					{Name: "parent-tag-id-or-name", Required: true, Value: redant.StringOf(new(string)), Description: "Parent tag id or name"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.RemoveTagExtends(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "tag extends relation removed"}, nil
				}),
			},
		},
	}
}

func tagBlockCmd() *redant.Command {
	return &redant.Command{
		Use:   "block",
		Short: "Block tag relation operations",
		Children: []*redant.Command{
			{
				Use:   "add <block-id> <tag-id>",
				Short: "Add tag to block",
				Args: redant.ArgSet{
					{Name: "block-id", Required: true, Value: redant.StringOf(new(string)), Description: "Block id or uuid"},
					{Name: "tag-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag id or name"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.AddBlockTag(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "block tag relation added"}, nil
				}),
			},
			{
				Use:   "remove <block-id> <tag-id>",
				Short: "Remove tag from block",
				Args: redant.ArgSet{
					{Name: "block-id", Required: true, Value: redant.StringOf(new(string)), Description: "Block id or uuid"},
					{Name: "tag-id", Required: true, Value: redant.StringOf(new(string)), Description: "Tag id or name"},
				},
				ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (StatusResult, error) {
					client := NewClient()
					if err := client.RemoveBlockTag(ctx, inv.Args[0], inv.Args[1]); err != nil {
						return StatusResult{}, err
					}
					return StatusResult{OK: true, Message: "block tag relation removed"}, nil
				}),
			},
		},
	}
}
