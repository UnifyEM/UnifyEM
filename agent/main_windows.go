//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

//go:build windows

package main

import (
	"github.com/UnifyEM/UnifyEM/agent/userdata"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// checkUserHelperMode is a no-op on Windows
func checkUserHelperMode() bool {
	return false
}

// initUserDataListener is a no-op on Windows
func initUserDataListener(log interfaces.Logger) {
	// Not implemented on Windows
}

// cleanupUserDataListener is a no-op on Windows
func cleanupUserDataListener(log interfaces.Logger) {
	// Not implemented on Windows
}

// getUserDataSource returns nil on Windows
func getUserDataSource() *userdata.UserDataListener {
	return nil
}
