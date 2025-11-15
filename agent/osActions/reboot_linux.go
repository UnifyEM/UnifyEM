//go:build linux

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for Linux
package osActions

import (
	"errors"

	"github.com/UnifyEM/UnifyEM/common/runCmd"
)

// shutdownOrReboot attempts a clean shutdown/reboot
func (a *Actions) shutdownOrReboot(reboot bool) error {
	var err error

	// Try systemctl first (systemd-based distributions)
	if reboot {
		_, err = runCmd.Combined("systemctl", "reboot")
	} else {
		_, err = runCmd.Combined("systemctl", "poweroff")
	}

	if err != nil {
		a.logger.Errorf(8307, "Failed to execute systemctl command: %w", err.Error())
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
				_, err = runCmd.Combined("reboot")
			} else {
				_, err = runCmd.Combined("poweroff")
			}

			if err != nil {
				a.logger.Errorf(8310, "Failed to execute reboot/poweroff command: %s", err.Error())
				return errors.New("all shutdown/reboot methods failed")
			}
		}
	}

	return nil
}
