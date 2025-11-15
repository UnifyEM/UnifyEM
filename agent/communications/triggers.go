/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/UnifyEM/UnifyEM/agent/execute"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// triggerStatus provides a local copy to compare and identify new triggers
var triggerStatus schema.AgentTriggers

// initTriggers initializes the local triggers to false
func initTriggers() {
	triggerStatus = schema.NewAgentTriggers()
}

// AnyTriggerChanges checks if the triggers have changed
func (c *Communications) AnyTriggerChanges(newTriggers schema.AgentTriggers) bool {
	changed := false
	if newTriggers != triggerStatus {
		changed = true
	}
	return changed
}

// ProcessTriggers is run as a goroutine to send an immediate reply to the server
// and then activate the triggers. Errors are logged and handled to the extent possible.
func (c *Communications) ProcessTriggers(triggers schema.AgentTriggers) {

	// Marshall the triggers into a JSON string
	triggersJSON, err := json.Marshal(triggers)
	if err != nil {
		c.logger.Errorf(8049, "serialization error: %s", err.Error())
		return
	}

	// Assemble a message
	msg := fmt.Sprintf("triggers ack: %s", string(triggersJSON))

	// Attempt to send an ack message
	c.logMessageError(c.SendMessage(msg))

	// Lost mode can set and reset
	if triggers.Lost != triggerStatus.Lost {
		triggerStatus.Lost = triggers.Lost
		c.triggerSetLost(triggers.Lost)
	}

	// Only process the following triggers if they have changed to avoid conflicts
	if triggers.Uninstall && !triggerStatus.Uninstall {
		triggerStatus.Uninstall = true
		c.triggerUninstall()
	}

	if triggers.Wipe && !triggerStatus.Wipe {
		triggerStatus.Wipe = true
		c.triggerWipe()
	}

	// Update the local copy
	triggerStatus = triggers
}

func (c *Communications) logMessageError(err error) {
	if err != nil {
		c.logger.Errorf(8040, "error sending message: %s", err.Error())
	}
}

func (c *Communications) triggerSetLost(lost bool) {

	// Create a message and log it
	msg := fmt.Sprintf("lost mode changed to %t", lost)
	c.logger.Info(8041, msg, nil)

	// Set or reset the global lost flag
	global.Lost = lost

	// Store it in the config file
	c.conf.AP.Set(global.ConfigLost, lost)

	// Attempt to send a message to the server
	c.logMessageError(c.SendMessage(msg))

	// Best effort checkpoint
	_ = c.conf.Checkpoint()
}

func (c *Communications) triggerUninstall() {
	c.triggerLogAndSend("uninstall")
	if global.PROTECTED {
		return
	}

	prog, err := os.Executable()
	if err != nil {
		c.logger.Errorf(8045, "error getting executable path: %s", err.Error())
		return
	}

	args := []string{"uninstall"}
	err = execute.Execute(c.logger, prog, args)
	if err != nil {
		c.logger.Errorf(8046, "error executing uninstall: %s", err.Error())
		return
	}
}

func (c *Communications) triggerWipe() {
	c.triggerLogAndSend("wipe")
	if global.PROTECTED {
		return
	}

	// Initiate wipe // TODO
}

func (c *Communications) triggerLogAndSend(triggerName string) {
	msg := triggerName
	if global.PROTECTED {
		msg += " trigger ignored in protected mode"
	} else {
		msg += " trigger activated"
	}

	// Log it
	c.logger.Infof(8044, msg)

	// Attempt to send a message to the server
	c.logMessageError(c.SendMessage(msg))
}
