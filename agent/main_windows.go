//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package main

import (
	"context"
	"os/exec"
	"strings"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "manage-bde", "-protectors", "-get", "C:").CombinedOutput()
	if err != nil {
		if logger != nil {
			logger.Warningf(8608, "failed to retrieve BitLocker info: %s", err.Error())
		}
		return ""
	}
	return strings.TrimSpace(string(out))
}
