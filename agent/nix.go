//go:build !windows

/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for operating systems other than windows
package main

import (
	"os"
)

func launch() {

	// Check if arguments were passed
	if len(os.Args) == 1 {
		// If no arguments, start the service
		startService()
	} else {
		// Launch in interactive mode
		exitCode := console()
		exit(exitCode, false) // Set true to force a delay
	}
}
