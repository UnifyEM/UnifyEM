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

	return cmd
}

func agentList(_ []string, _ *util.NVPairs) error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Get(schema.EndpointAgent)))
	return nil
}

func agentGet(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Agent ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Get(schema.EndpointAgent + "/" + args[0])))
	return nil
}

func agentDelete(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Agent ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Delete(schema.EndpointAgent + "/" + args[0])))
	return nil
}

func agentSetName(args []string, _ *util.NVPairs) error {
	if len(args) < 2 {
		return errors.New("Agent ID and name are required\n")
	}

	agentMeta := schema.NewAgentMeta(args[0])
	agentMeta.FriendlyName = args[1]

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Put(schema.EndpointAgent+"/"+args[0], agentMeta)))
	return nil
}
