//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package main

import (
	"fmt"
	configCmd "github.com/UnifyEM/UnifyEM/cli/functions/config"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/functions/agent"
	"github.com/UnifyEM/UnifyEM/cli/functions/cmd"
	"github.com/UnifyEM/UnifyEM/cli/functions/events"
	"github.com/UnifyEM/UnifyEM/cli/functions/files"
	"github.com/UnifyEM/UnifyEM/cli/functions/ping"
	"github.com/UnifyEM/UnifyEM/cli/functions/regToken"
	"github.com/UnifyEM/UnifyEM/cli/functions/report"
	"github.com/UnifyEM/UnifyEM/cli/functions/request"
	"github.com/UnifyEM/UnifyEM/cli/functions/version"
	"github.com/UnifyEM/UnifyEM/cli/global"
)

func main() {
	var err error

	// Get the name of this binary, eliminating any path information
	progName := os.Args[0]
	progName = progName[strings.LastIndex(progName, "/")+1:]

	// Initialize the root command
	rootCmd := &cobra.Command{
		//Use:   progName + " <command> [flags]",
		Use:   progName,
		Short: global.Description,
		Long:  global.LongDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("A subcommand is required\n")
		},
	}

	// Disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add the functions
	rootCmd.AddCommand(agent.Register())
	rootCmd.AddCommand(cmd.Register())
	rootCmd.AddCommand(configCmd.Register())
	rootCmd.AddCommand(events.Register())
	rootCmd.AddCommand(files.Register())
	rootCmd.AddCommand(ping.Register())
	rootCmd.AddCommand(report.Register())
	rootCmd.AddCommand(request.Register())
	rootCmd.AddCommand(version.Register())
	rootCmd.AddCommand(regToken.Register())

	// Execute the CLI
	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
