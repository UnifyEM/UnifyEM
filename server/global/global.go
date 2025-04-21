//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// See LICENSE file for details
//

package global

import "github.com/UnifyEM/UnifyEM/common"

const (
	Version           = common.Version
	Build             = common.Build
	Name              = "UEMServer"
	LogName           = "uem-server"
	Description       = "UnifyEM Server"
	WindowsBinaryName = "uem-server.exe"
	UnixBinaryName    = "uem-server"
	FileDirPattern    = "/files/" // URL pattern for file downloads
	MessageQueueSize  = 500       // Size of the message queue
	TaskTicker        = 10        // seconds between task checks
	ConsoleExitDelay  = 10        // seconds to wait so that user can read the console output when exiting
	TokenLength       = 64        // Length of registration token and JWT authentication key prior to base-64 encoding
	MemoryCacheTTL    = 600       // Time to live for memory cache items in seconds
)

var (
	UnixConfigFiles         = []string{"/etc/uem-server.conf", "/usr/local/etc/uem-server.conf", "/var/root/uem-server.conf"}
	UnixDefaultDataPaths    = []string{"/opt/uem-server", "/var/lib/uem-server", "/usr/local/uem-server"}
	WindowsDefaultDataPaths = []string{"C:\\ProgramData\\uem-server"}
	Debug                   = false
	ListenOverride          = ""
)
