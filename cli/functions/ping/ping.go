/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package ping

import (
	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func Register() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "ping the server",
		Long:  "ping the server, which also requires login",
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute()
		},
	}
}

func execute() error {

	// Create communications object
	c := communications.New(login.Login())

	// Post the command to the server and display the result
	display.ErrorWrapper(display.GenericResp(c.Get(schema.EndpointPing)))
	return nil
}
