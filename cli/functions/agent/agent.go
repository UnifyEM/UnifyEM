//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package agent

import (
	"encoding/json"
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
		Use:     "agent",
		Aliases: []string{"agents"},
		Short:   "agent functions",
		Long:    "agent-related functions",
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

	cmd.AddCommand(&cobra.Command{
		Use:   "user-add <agent_id>|tag=<tag> <user1> [<user2> ...]",
		Short: "add users to an agent",
		Long:  "add one or more users to the specified agent or all agents with a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentAddUsers(args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "user-remove <agent_id>|tag=<tag> <user1> [<user2> ...]",
		Short: "remove users from an agent",
		Long:  "remove one or more users from the specified agent or all agents with a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentRemoveUsers(args)
		},
	})

	return cmd
}

func agentList(_ []string, _ *util.NVPairs) error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.AnyResp(c.Get(schema.EndpointAgent)))
	return nil
}

func agentGet(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Agent ID is required")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.AnyResp(c.Get(schema.EndpointAgent + "/" + args[0])))
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

// Add users to an agent or all agents with a tag
func agentAddUsers(args []string) error {
	if len(args) < 2 {
		return errors.New("Agent ID or tag=<tag> and at least one user are required")
	}
	c := communications.New(login.Login())
	if len(args) > 0 && len(args[0]) > 4 && args[0][:4] == "tag=" {
		tag := args[0][4:]
		// Query agents by tag
		_, body, err := c.Get(schema.EndpointAgent + "/by-tag/" + tag)
		if err != nil {
			return fmt.Errorf("failed to query agents by tag: %v", err)
		}
		var resp schema.AgentsByTagResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse agents by tag response: %v", err)
		}
		if len(resp.Agents) == 0 {
			return fmt.Errorf("no agents found with tag: %s", tag)
		}
		var firstErr error
		for _, agent := range resp.Agents {
			req := schema.AgentUsersRequest{Users: args[1:]}
			status, body, err := c.Post(schema.EndpointAgent+"/"+agent.AgentID+"/users/add", req)
			display.ErrorWrapper(display.GenericResp(status, body, err))
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
	// Single agent
	req := schema.AgentUsersRequest{Users: args[1:]}
	status, body, err := c.Post(schema.EndpointAgent+"/"+args[0]+"/users/add", req)
	display.ErrorWrapper(display.GenericResp(status, body, err))
	return nil
}

// Remove users from an agent or all agents with a tag
func agentRemoveUsers(args []string) error {
	if len(args) < 2 {
		return errors.New("Agent ID or tag=<tag> and at least one user are required")
	}
	c := communications.New(login.Login())
	if len(args) > 0 && len(args[0]) > 4 && args[0][:4] == "tag=" {
		tag := args[0][4:]
		// Query agents by tag
		_, body, err := c.Get(schema.EndpointAgent + "/by-tag/" + tag)
		if err != nil {
			return fmt.Errorf("failed to query agents by tag: %v", err)
		}
		var resp schema.AgentsByTagResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed to parse agents by tag response: %v", err)
		}
		if len(resp.Agents) == 0 {
			return fmt.Errorf("no agents found with tag: %s", tag)
		}
		var firstErr error
		for _, agent := range resp.Agents {
			req := schema.AgentUsersRequest{Users: args[1:]}
			status, body, err := c.Post(schema.EndpointAgent+"/"+agent.AgentID+"/users/remove", req)
			display.ErrorWrapper(display.GenericResp(status, body, err))
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
	// Single agent
	req := schema.AgentUsersRequest{Users: args[1:]}
	status, body, err := c.Post(schema.EndpointAgent+"/"+args[0]+"/users/remove", req)
	display.ErrorWrapper(display.GenericResp(status, body, err))
	return nil
}
