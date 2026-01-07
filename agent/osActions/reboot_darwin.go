//go:build darwin

/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package osActions

import (
	"errors"
)

// shutdownOrReboot attempts a clean shutdown/reboot
func (a *Actions) shutdownOrReboot(reboot bool) error {
	var err error

	if reboot {
		_, err = a.runner.Combined("shutdown", "-r", "now")
	} else {
		_, err = a.runner.Combined("shutdown", "-h", "now")
	}

	if err != nil {
		a.logger.Errorf(8301, "Failed to execute shutdown command: %s", err.Error())
		a.logger.Info(8302, "Falling back to osascript", nil)

		if reboot {
			_, err = a.runner.Combined("osascript", "-e", `tell application "System Events" to restart`)
		} else {
			_, err = a.runner.Combined("osascript", "-e", `tell application "System Events" to shut down`)
		}

		if err != nil {
			a.logger.Errorf(8303, "Failed to execute osascript command: %s", err.Error())
			return errors.New("shutdown command and osascript both failed")
		}
	}
	return nil
}
