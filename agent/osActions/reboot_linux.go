// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

// Code for Linux
//go:build linux

package osActions

import (
	"errors"
	"os/exec"
)

// shutdownOrReboot attempts a clean shutdown/reboot
func (a *Actions) shutdownOrReboot(reboot bool) error {
	var cmd *exec.Cmd

	// Try systemctl first (systemd-based distributions)
	if reboot {
		cmd = exec.Command("systemctl", "reboot")
	} else {
		cmd = exec.Command("systemctl", "poweroff")
	}

	err := cmd.Run()
	if err != nil {
		a.logger.Errorf(8307, "Failed to execute systemctl command: %s", err.Error())
		a.logger.Info(8308, "Falling back to shutdown command", nil)

		// Fall back to shutdown command
		if reboot {
			cmd = exec.Command("shutdown", "-r", "now")
		} else {
			cmd = exec.Command("shutdown", "-h", "now")
		}

		err = cmd.Run()
		if err != nil {
			a.logger.Errorf(8309, "Failed to execute shutdown command: %s", err.Error())

			// Last resort, try the older init-based commands
			if reboot {
				cmd = exec.Command("reboot")
			} else {
				cmd = exec.Command("poweroff")
			}

			err = cmd.Run()
			if err != nil {
				a.logger.Errorf(8310, "Failed to execute reboot/poweroff command: %s", err.Error())
				return errors.New("all shutdown/reboot methods failed")
			}
		}
	}

	return nil
}
