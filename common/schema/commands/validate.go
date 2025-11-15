/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package commands

import (
	"errors"
	"fmt"
)

// Validate checks if the command and parameters are valid
//
//goland:noinspection GoUnusedExportedFunction
func Validate(cmd string, parameters map[string]string) error {
	var err error
	var cmdTemplate Command

	// Check if the command exists
	if c, ok := cmds.Commands[cmd]; !ok {
		err = errors.New("invalid command")
		return err
	} else {
		cmdTemplate = c
	}

	// Check if all required arguments are present
	for _, arg := range cmdTemplate.RequiredArgs {
		if _, ok := parameters[arg]; !ok {
			err = errors.New("missing required argument: " + arg)
			return err
		}
	}

	// Check that all parameters are either a required or optional argument
	for param := range parameters {
		found := false
		for _, arg := range cmdTemplate.RequiredArgs {
			if arg == param {
				found = true
				break
			}
		}
		if !found {
			for _, arg := range cmdTemplate.OptionalArgs {
				if arg == param {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("invalid argument: %s", param)
		}
	}
	return nil
}

// ValidateCmd checks if the command is valid
//
//goland:noinspection GoUnusedExportedFunction
func ValidateCmd(cmd string) error {
	if _, ok := cmds.Commands[cmd]; !ok {
		return errors.New("invalid command")
	}
	return nil
}
