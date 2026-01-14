//go:build windows

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for windows
package main

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
)

// launch handles what would usually be in main() in an OS-specific way
func launch() {
	isSvc, err := svc.IsWindowsService()
	if err != nil {
		fmt.Printf("Failed to determine if we are running in an interactive session: %v\n", err)
		return
	}

	if isSvc {
		// Start the service
		startService()
	} else {
		// Launch in console mode
		exitCode := console()
		exit(exitCode, true) // Set true to force a delay
	}
}
