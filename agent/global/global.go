/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package global provides global variables and functions for UEMAgent that can be imported by any other package
package global

import (
	"github.com/UnifyEM/UnifyEM/common"
)

// These constants are used throughout the agent
const (
	Version                   = common.Version
	Build                     = common.Build
	Name                      = "UEMAgent"
	LogName                   = "uem-agent"
	Description               = "UnifyEM Agent"
	WindowsBinaryName         = "uem-agent.exe"
	UnixBinaryName            = "uem-agent"
	TaskTicker                = 5   // seconds between task checks
	ConsoleExitDelay          = 10  // seconds to wait so that user can read the console output when exiting
	TaskQueueSize             = 100 // maximum number of tasks to queue
	UserHelperFlag            = "--user-helper"
	CollectionIntervalFlag    = "--collection-interval"
	DefaultCollectionInterval = 300 // 5 minutes in seconds
	SocketPath                = "/var/run/uem-agent.sock"
	SocketPerms               = 0666 // Allow user processes to connect
)

// These constants are intended for development purposes only and disable important security features.
// For production use, the following constants should all be false.
//
//goland:noinspection GoUnusedConst
const (
	Unsafe      = false // Allows self-signed certificates and HTTP
	DisableHash = false // Don't require hash verification on file downloads
	// DisableSig is not implemented yet
	DisableSig = true // Don't require signed requests from the server
)

// Global values that either can or should not be constants
var (
	UnixConfigFiles         = []string{"/etc/uem-agent.conf", "/usr/local/etc/uem-agent.conf", "/var/root/uem-agent.conf"}
	UnixDefaultDataPaths    = []string{"/opt/uem-agent", "/var/lib/uem-agent", "/usr/local/uem-agent"}
	WindowsDefaultDataPaths = []string{"C:\\ProgramData\\uem-agent"}
	Debug                   = true
	Lost                    = false
)
