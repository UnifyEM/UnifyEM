//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

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
		Long:    "query and delete agent requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("A subcommand is required\n")
			}
			return fmt.Errorf("Unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list requests",
		Long:  "request a list of requests",
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

	return cmd
}

func requestList(_ []string, _ *util.NVPairs) error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.RequestList(c.Get(schema.EndpointRequest)))
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
