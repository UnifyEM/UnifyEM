//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package regToken

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/common/schema"

	"github.com/spf13/cobra"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regtoken [new]",
		Short: "registration token functions",
		Long:  "view current registration token or generate a new one",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return getRegToken()
			}
			return fmt.Errorf("Unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "get registration token",
		Long:  "get the registration token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getRegToken()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "new",
		Short: "generate new registration token",
		Long:  "generate new registration token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return newRegToken()
		},
	})

	return cmd
}

func getRegToken() error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.AnyResp(c.Get(schema.EndpointRegToken)))
	return nil
}

func newRegToken() error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.AnyResp(c.Post(schema.EndpointRegToken, nil)))
	return nil
}
