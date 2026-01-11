//go:build linux

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package main

import (
	"github.com/UnifyEM/UnifyEM/agent/userdata"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// checkUserHelperMode is a no-op on Linux
func checkUserHelperMode() bool {
	return false
}

// initUserDataListener is a no-op on Linux
func initUserDataListener(log interfaces.Logger) {
	// Not implemented on Linux
}

// cleanupUserDataListener is a no-op on Linux
func cleanupUserDataListener(log interfaces.Logger) {
	// Not implemented on Linux
}

// getUserDataSource returns nil on Linux
func getUserDataSource() *userdata.UserDataListener {
	return nil
}
