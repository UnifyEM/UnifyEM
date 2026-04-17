/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package cert

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/certstore"
)

func Register() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "certificate management commands",
		Long:  "manage pinned TLS certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a subcommand is required\n")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "remove <host:port>",
		Short: "remove a pinned certificate",
		Long:  "remove all pinned certificate entries for the given host:port from ~/.uemcert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeExec(args[0])
		},
	})

	return cmd
}

func removeExec(host string) error {
	removed, err := certstore.Remove(host)
	if err != nil {
		return fmt.Errorf("failed to remove certificate: %w", err)
	}
	if !removed {
		fmt.Printf("No pinned certificate found for %s\n", host)
		return nil
	}
	fmt.Printf("Pinned certificate for %s removed\n", host)
	return nil
}
