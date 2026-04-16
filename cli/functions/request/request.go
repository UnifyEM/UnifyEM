/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package request

import (
	"errors"
	"fmt"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/cli/util"
	"github.com/UnifyEM/UnifyEM/common/schema"

	"github.com/spf13/cobra"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "request",
		Aliases: []string{"requests"},
		Short:   "request functions",
		Long:    "query, delete, and cancel agent requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("A subcommand is required\n")
			}
			return fmt.Errorf("Unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list [agent_id]",
		Short: "list requests",
		Long:  "list all requests, or all requests for a specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestList(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <request_id>",
		Short: "get request",
		Long:  "get information about the specified request",
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestGet(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <request_id>",
		Short: "delete request",
		Long:  "delete the specified request",
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestDelete(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "cancel <request_id>",
		Short: "cancel request",
		Long:  "cancel the specified request",
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestCancel(args, util.NewNVPairs(args))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "cancel-agent <agent_id>",
		Short: "cancel agent requests",
		Long:  "cancel all requests for the specified agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestCancelAgent(args, util.NewNVPairs(args))
		},
	})

	return cmd
}

func requestList(args []string, _ *util.NVPairs) error {
	c := communications.New(login.Login())
	if len(args) > 0 {
		display.ErrorWrapper(display.RequestList(c.Get(schema.EndpointAgent + "/" + args[0] + "/requests")))
	} else {
		display.ErrorWrapper(display.RequestList(c.Get(schema.EndpointRequest)))
	}
	return nil
}

func requestGet(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Request ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.RequestList(c.Get(schema.EndpointRequest + "/" + args[0])))
	return nil
}

func requestDelete(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Request ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Delete(schema.EndpointRequest + "/" + args[0])))
	return nil
}

func requestCancel(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Request ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Post(schema.EndpointRequest+"/"+args[0]+"/cancel", nil)))
	return nil
}

func requestCancelAgent(args []string, _ *util.NVPairs) error {
	if len(args) == 0 {
		return errors.New("Agent ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Post(schema.EndpointAgent+"/"+args[0]+"/cancel-requests", nil)))
	return nil
}
