//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package version

import (
	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/global"
)

func Register() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "version, copyright, and legal",
		Long:  "display version, copyright, and legal information",
		RunE: func(cmd *cobra.Command, args []string) error {
			global.Banner()
			return nil
		},
	}
}
