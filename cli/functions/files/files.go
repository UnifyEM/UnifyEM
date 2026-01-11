/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package files

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "files",
		Aliases: []string{"file"},
		Short:   "file management functions",
		Long:    "manage files used by agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a subcommand is required\n")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "deploy",
		Short: "create deploy.json",
		Long:  "create deploy.json containing file hashes for agent upgrades",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createDeploy()
		},
	})

	return cmd
}

func createDeploy() error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Post(schema.EndpointCreateDeployFile, nil)))
	return nil
}
