//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

// Code for macOS
//go:build darwin

package osActions

import (
	"errors"
	"os/exec"
)

// shutdownOrReboot attempts a clean shutdown/reboot
func (a *Actions) shutdownOrReboot(reboot bool) error {
	var cmd *exec.Cmd
	if reboot {
		cmd = exec.Command("shutdown", "-r", "now")
	} else {
		cmd = exec.Command("shutdown", "-h", "now")
	}

	err := cmd.Run()
	if err != nil {
		a.logger.Errorf(8301, "Failed to execute shutdown command: %s", err.Error())
		a.logger.Info(8302, "Falling back to osascript", nil)

		if reboot {
			cmd = exec.Command("osascript", "-e", `tell application "System Events" to restart`)
		} else {
			cmd = exec.Command("osascript", "-e", `tell application "System Events" to shut down`)
		}

		err = cmd.Run()
		if err != nil {
			a.logger.Errorf(8303, "Failed to execute osascript command: %s", err.Error())
			return errors.New("shutdown command and osascript both failed")
		}
	}
	return nil
}
