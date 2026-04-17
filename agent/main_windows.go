//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package main

import (
	"os/exec"
	"strings"

	"github.com/UnifyEM/UnifyEM/agent/userdata"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// checkUserHelperMode is a no-op on Windows
func checkUserHelperMode() bool {
	return false
}

// initUserDataListener is a no-op on Windows
func initUserDataListener(_ interfaces.Logger) {
	// Not implemented on Windows
}

// cleanupUserDataListener is a no-op on Windows
func cleanupUserDataListener(_ interfaces.Logger) {
	// Not implemented on Windows
}

// getUserDataSource returns nil on Windows
func getUserDataSource() *userdata.UserDataListener {
	return nil
}

func getBitLockerInfo() string {
	out, err := exec.Command("manage-bde", "-protectors", "-get", "C:").CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
