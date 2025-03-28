//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/cli/util"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		//Use:   "cmd <command> [parameters]",
		Use:   "cmd",
		Short: "send command",
		Long:  "send the specified command to agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a subcommand is required\n")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   commands.DownloadExecute + " agent_id=<agent ID> url=<URL> [arg1=value1] [arg2=value2] ...",
		Short: "download and execute a file",
		Long:  "download a file from the specified URL and execute it on the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.DownloadExecute, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.Ping + " agent_id=<agent ID>",
		Short: "ping an agent",
		Long:  "instruct the server to ping the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.Ping, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.Execute + " agent_id=<agent ID> cmd=<command> [arg1=value1] [arg2=value2] ...",
		Short: "execute a command",
		Long:  "execute the specified command on the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.Execute, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.Reboot + " agent_id=<agent ID>",
		Short: "reboot an agent",
		Long:  "instruct the server to reboot the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.Reboot, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.Shutdown + " agent_id=<agent ID>",
		Short: "shutdown an agent",
		Long:  "instruct the server to shutdown the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.Shutdown, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.Status + " agent_id=<agent ID>",
		Short: "get agent status",
		Long:  "request the status of the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.Status, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.Upgrade + " agent_id=<agent ID>",
		Short: "agent upgrade",
		Long:  "instruct the agent to download and install the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.Upgrade, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.UserAdd + " agent_id=<agent ID> username=<username> password=<password> [admin=true|false]",
		Short: "add a user",
		Long:  "add a user to the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.UserAdd, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.UserAdmin + " agent_id=<agent ID> username=<username> admin=true|false",
		Short: "grant or revoke admin privileges",
		Long:  "set or remove the specified user as an admin on the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.UserAdmin, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.UserPassword + " agent_id=<agent ID> username=<username> password=<password>",
		Short: "set user password",
		Long:  "set the password for the specified user on the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.UserPassword, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.UserList + " agent_id=<agent ID>",
		Short: "list users",
		Long:  "list the users on the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.UserList, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.UserLock + " agent_id=<agent ID> username=<username> [shutdown=true]",
		Short: "lock user account",
		Long:  "lock the specified user on the specified agent and optionally shutdown the device",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.UserLock, args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   commands.UserUnlock + " agent_id=<agent ID> username=<username>",
		Short: "unlock user account",
		Long:  "unlock the specified user account on the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(commands.UserUnlock, args, util.NewNVPairs(args))
		},
	})
	return cmd
}

func execute(subCmd string, _ []string, pairs *util.NVPairs) error {

	// Create communications object
	c := communications.New(login.Login())

	err := commands.Validate(subCmd, pairs.Pairs)
	if err != nil {
		return fmt.Errorf("command validation failed: %s\n", err.Error())
	}

	// Initialize a new command object
	cmd := schema.NewCmdRequest()
	cmd.Cmd = subCmd
	cmd.Parameters = pairs.ToMap()

	// Post the command to the server and display the result
	display.ErrorWrapper(display.CmdResp(c.Post(schema.EndpointCmd, cmd)))
	return nil
}
