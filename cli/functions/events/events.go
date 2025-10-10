//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package events

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
	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"event"},
		Short:   "agent event functions",
		Long:    "agent event functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a subcommand is required\n")
			}
			return fmt.Errorf("unknown subcommand: %s\n", args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get agent_id=<agent_id [start=<YYYYMMDD>] [end=<YYYYMMDD>] [start_time=<unix time>] [end_time=<unix time>] [type=<message|alert|status>]",
		Short: "get events",
		Long:  "get events for the specified agent with optional start and end times",
		RunE: func(cmd *cobra.Command, args []string) error {
			return eventsGet(args, util.NewNVPairs(args))
		},
	})

	return cmd
}

func eventsGet(_ []string, pairs *util.NVPairs) error {
	c := communications.New(login.Login())
	display.ErrorWrapper(display.AnyResp(c.GetQuery(schema.EndpointEvents, pairs)))
	return nil
}
