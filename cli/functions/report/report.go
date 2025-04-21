//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package report

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/cli/util"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

func Register() *cobra.Command {
	return &cobra.Command{
		Use:   "report <report name>",
		Short: "request report",
		Long:  "request the specified report",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("A report name is required\n")
			}
			execute(args, util.NewNVPairs(args))
			return nil
		},
	}
}

func execute(args []string, pairs *util.NVPairs) {

	// Create communications object
	c := communications.New(login.Login())

	// Initialize a new common command object
	cmd := schema.NewReportRequest()
	cmd.Report = args[0]
	cmd.Parameters = pairs.ToMap()

	// Post the command to the server and display the result
	display.ErrorWrapper(display.ReportResp(c.Post(schema.EndpointReport, cmd)))
}
