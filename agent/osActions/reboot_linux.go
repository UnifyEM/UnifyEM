//go:build linux

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"errors"
)

// shutdownOrReboot attempts a clean shutdown/reboot
func (a *Actions) shutdownOrReboot(reboot bool) error {
	var err error

	// Try systemctl first (systemd-based distributions)
	if reboot {
		_, err = a.runner.Combined("systemctl", "reboot")
	} else {
		_, err = a.runner.Combined("systemctl", "poweroff")
	}

	if err != nil {
		a.logger.Errorf(8307, "Failed to execute systemctl command: %w", err.Error())
		a.logger.Info(8308, "Falling back to shutdown command", nil)

		// Fall back to shutdown command
		if reboot {
			_, err = a.runner.Combined("shutdown", "-r", "now")
		} else {
			_, err = a.runner.Combined("shutdown", "-h", "now")
		}

		if err != nil {
			a.logger.Errorf(8309, "Failed to execute shutdown command: %s", err.Error())

			// Last resort, try the older init-based commands
			if reboot {
				_, err = a.runner.Combined("reboot")
			} else {
				_, err = a.runner.Combined("poweroff")
			}

			if err != nil {
				a.logger.Errorf(8310, "Failed to execute reboot/poweroff command: %s", err.Error())
				return errors.New("all shutdown/reboot methods failed")
			}
		}
	}

	return nil
}
