/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package configCmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/cli/util"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "configuration commands",
		Long:  "get or set configuration information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a subcommand is required\n")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	agents := &cobra.Command{
		Use:   "agents <get> <set arg1=value1 [arg2=value2] ...>",
		Short: "global agent configuration",
		Long:  "get or set global agent configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Assume get
				return get("agents")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	agents.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "get global agent configuration",
		Long:  "get global agent configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return get("agents")
		},
	})

	agents.AddCommand(&cobra.Command{
		Use:   "set arg1=value1 [arg2=value2] ...",
		Short: "set global agent configuration",
		Long:  "set global agent configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return set("agents", util.NewNVPairs(args))
		},
	})

	server := &cobra.Command{
		Use:   "server <get> <set arg1=value1 [arg2=value2] ...>",
		Short: "server configuration",
		Long:  "get or set server configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Assume get
				return get("server")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "get server configuration",
		Long:  "get server configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return get("server")
		},
	})

	server.AddCommand(&cobra.Command{
		Use:   "set arg1=value1 [arg2=value2] ...",
		Short: "set server configuration",
		Long:  "set server configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return set("server", util.NewNVPairs(args))
		},
	})

	// add the agent commands
	cmd.AddCommand(agents)
	cmd.AddCommand(server)
	return cmd
}

func get(cfg string) error {
	var endpoint string

	targetLC := strings.ToLower(cfg)
	switch targetLC {
	case "agents":
		endpoint = schema.EndpointConfigAgents
	case "server":
		endpoint = schema.EndpointConfigServer
	default:
		return fmt.Errorf("config set '%s' does not exist", cfg)
	}

	// Create communications object
	c := communications.New(login.Login())

	// Post the command to the server and display the result
	display.ErrorWrapper(display.AnyResp(c.Get(endpoint)))
	return nil
}

func set(cfg string, pairs *util.NVPairs) error {
	var endpoint string

	targetLC := strings.ToLower(cfg)
	switch targetLC {
	case "agents":
		endpoint = schema.EndpointConfigAgents
	case "server":
		endpoint = schema.EndpointConfigServer
	default:
		return fmt.Errorf("config set '%s' does not exist", cfg)
	}

	// Create communications object
	c := communications.New(login.Login())

	// Initialize a new command object
	req := schema.NewConfigRequest()

	// Iterate through pairs and add them to the request
	for n, v := range pairs.Pairs {
		req.Parameters[n] = v
	}

	// Post the command to the server and display the result
	display.ErrorWrapper(display.GenericResp(c.Post(endpoint, req)))
	return nil
}
