/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Code for operating systems other than windows
//go:build !windows

package main

import (
	"os"
)

func launch() {

	// Check if arguments were passed
	if len(os.Args) == 1 {
		// If no arguments, start the service in the background
		startService(true)
	} else {
		// Launch in interactive mode
		console()
		exit(0, false) // Set true to force a delay
	}
}
