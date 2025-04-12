//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package agent

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/cli/util"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "agent functions",
		Long:  "agent-related functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a subcommand is required\n")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list agents",
		Long:  "request a list of agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentList(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <agent_id>",
		Short: "get agent",
		Long:  "get information about the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentGet(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <agent_id>",
		Short: "delete agent",
		Long:  "delete the specified agent from the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentDelete(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "lost <agent_id>",
		Short: "active lost mode",
		Long:  "instruct the agent to enter lost mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			triggers := schema.NewAgentTriggers()
			triggers.Lost = true
			return agentSetTriggers(args, triggers)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "uninstall <agent_id>",
		Short: "uninstall agent",
		Long:  "uninstall the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			triggers := schema.NewAgentTriggers()
			triggers.Uninstall = true
			return agentSetTriggers(args, triggers)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "wipe <agent_id>",
		Short: "wipe agent disk",
		Long:  "instruct the agent to wipe all drives and set lost mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			triggers := schema.NewAgentTriggers()
			triggers.Wipe = true
			triggers.Lost = true
			return agentSetTriggers(args, triggers)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset <agent_id>",
		Short: "reset agent",
		Long:  "clear the lost, lock, wipe, and uninstall flags for the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentResetTriggers(args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "name <agent_id> <name>",
		Short: "set agent name",
		Long:  "set the agent's friendly name",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentSetName(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "tags <agent_id>",
		Short: "list tags for an agent",
		Long:  "list all tags assigned to the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentListTags(args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "tag-add <agent_id> <tag1> [<tag2> ...]",
		Short: "add tags to an agent",
		Long:  "add one or more tags to the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentAddTags(args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "tag-remove <agent_id> <tag1> [<tag2> ...]",
		Short: "remove tags from an agent",
		Long:  "remove one or more tags from the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentRemoveTags(args)
		},
	})

	return cmd
}

func agentList(_ []string, _ *util.NVPairs) error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Get(schema.EndpointAgent)))
	return nil
}

func agentGet(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Agent ID is required")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Get(schema.EndpointAgent + "/" + args[0])))
	return nil
}

func agentDelete(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Agent ID is required")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Delete(schema.EndpointAgent + "/" + args[0])))
	return nil
}

func agentSetName(args []string, _ *util.NVPairs) error {
	if len(args) < 2 {
		return errors.New("Agent ID and name are required")
	}

	agentMeta := schema.NewAgentMeta(args[0])
	agentMeta.FriendlyName = args[1]

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Put(schema.EndpointAgent+"/"+args[0], agentMeta)))
	return nil
}

// List tags for an agent
func agentListTags(args []string) error {
	if len(args) < 1 {
		return errors.New("Agent ID is required")
	}
	c := communications.New(login.Login())
	status, body, err := c.Get(schema.EndpointAgent + "/" + args[0] + "/tags")
	display.ErrorWrapper(display.TagsResp(status, body, err))
	return nil
}

// Add tags to an agent
func agentAddTags(args []string) error {
	if len(args) < 2 {
		return errors.New("Agent ID and at least one tag are required")
	}
	req := schema.AgentTagsRequest{Tags: args[1:]}
	c := communications.New(login.Login())
	status, body, err := c.Post(schema.EndpointAgent+"/"+args[0]+"/tags/add", req)
	display.ErrorWrapper(display.GenericResp(status, body, err))
	return nil
}

// Remove tags from an agent
func agentRemoveTags(args []string) error {
	if len(args) < 2 {
		return errors.New("Agent ID and at least one tag are required")
	}
	req := schema.AgentTagsRequest{Tags: args[1:]}
	c := communications.New(login.Login())
	status, body, err := c.Post(schema.EndpointAgent+"/"+args[0]+"/tags/remove", req)
	display.ErrorWrapper(display.GenericResp(status, body, err))
	return nil
}
