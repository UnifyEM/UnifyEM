//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package agent

import (
	"errors"

	"github.com/UnifyEM/UnifyEM/cli/communications"
	"github.com/UnifyEM/UnifyEM/cli/display"
	"github.com/UnifyEM/UnifyEM/cli/login"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// agentSetTriggers sets one or more triggers for the specified agent
// Note that this function can not be used to reset the triggers
func agentSetTriggers(args []string, triggers schema.AgentTriggers) error {

	// Require one argument
	if len(args) != 1 {
		return errors.New("Agent ID is required\n")
	}

	agentMeta := schema.NewAgentMeta(args[0])
	agentMeta.Triggers = triggers

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Post(schema.EndpointAgent+"/"+args[0], agentMeta)))
	return nil
}

// agentResetTriggers resets all triggers for the specified agent
func agentResetTriggers(args []string) error {

	// Require one argument
	if len(args) != 1 {
		return errors.New("Agent ID is required\n")
	}

	c := communications.New(login.Login())
	display.ErrorWrapper(display.GenericResp(c.Put(schema.EndpointReset+"/"+args[0], nil)))
	return nil
}
