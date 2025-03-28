//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package global

import (
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"os"
	"runtime"
)

const (
	ConfigPrivate      = "client_private"
	ConfigRegToken     = "reg_token"
	ConfigLost         = "config_lost"
	ConfigAgentLogFile = "log_file"
	ConfigAgentDataDir = "data_dir"
	ConfigAgentID      = "agent_id"
	ConfigServerURL    = "server_url"
	ConfigRefreshToken = "refresh_token"
)

// setDefaults makes sure the sets exist, sets default values, and constraints
func setDefaults(c interfaces.Config) (interfaces.Parameters, interfaces.Parameters) {

	// Server configuration set
	ac := schema.SetAgentDefaults(c)

	ap := c.NewSet(ConfigPrivate)
	ap.SetConstraint(ConfigRegToken, 0, 0, "")
	ap.SetConstraint(ConfigLost, 0, 0, false)
	ap.SetConstraint(ConfigAgentLogFile, 0, 0, "")
	ap.SetConstraint(ConfigAgentDataDir, 0, 0, "")
	ap.SetConstraint(ConfigAgentID, 0, 0, "")
	ap.SetConstraint(ConfigServerURL, 0, 0, "")

	// Return the sets
	return ac, ap
}

// DefaultLog is used to create a log location if the usual approach fails
func DefaultLog() string {
	if runtime.GOOS == "windows" {
		return "C:" + string(os.PathSeparator) + "ProgramData" + string(os.PathSeparator) + LogName + string(os.PathSeparator) + LogName + ".log"
	}
	return "/var/log/" + LogName + ".log"
}
