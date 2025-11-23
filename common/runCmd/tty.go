/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package runCmd

import (
	"fmt"
)

// Interactive represents a command and series of interactions
type Interactive struct {
	Command []string
	Actions []Action
	AsUser  *UserLogin // Optional: login as different user first
}

// UserLogin contains credentials for logging in as a different user
type UserLogin struct {
	Username string
	Password string
}

// Action represents a single prompt/response interaction in an interactive TTY session
type Action struct {
	WaitFor  string // The string to wait for in the output
	Send     string // The value to send when the prompt is detected
	DebugMsg string // Optional debug message to print when this interaction occurs
}

// TTY runs a command in a pseudo-terminal and handles interactive prompts
// by waiting for specific strings and sending responses.
// If AsUser is specified, it will first login as that user before running the command.
// Returns all output from the command and any error that occurred.
func TTY(def Interactive) (string, error) {
	return osTTY(def)
}

// TTYAsUser runs a non-interactive command as another user and returns output
// This is a convenience function for commands that don't need interaction
func TTYAsUser(asUser *UserLogin, cmdAndArgs ...string) (string, error) {
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("no command specified")
	}

	return TTY(Interactive{
		Command: cmdAndArgs,
		Actions: []Action{}, // No interactions
		AsUser:  asUser,
	})
}
