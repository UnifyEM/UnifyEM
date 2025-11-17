/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

import (
	"github.com/UnifyEM/UnifyEM/common/interfaces"
)

// This is shared so that the agent and server load the same defaults

const (
	ConfigAgentSet              = "agent_config"
	ConfigAgentSyncInterval     = "sync_interval"
	ConfigAgentSyncPending      = "sync_pending"
	ConfigAgentSyncRetry        = "sync_retry"
	ConfigAgentSyncLost         = "sync_lost"
	ConfigAgentStatusInterval   = "status_interval"
	ConfigAgentLogRetention     = "log_retention"
	ConfigAgentLogStdout        = "log_stdout"
	ConfigAgentLogWindowsDisk   = "log_windows_disk"
	ConfigAgentLogWindowsEvents = "log_windows_events"
	ConfigAgentLogMacOSDisk     = "log_macos_disk"
	ConfigAgentLogLinuxDisk     = "log_linux_disk"
	ConfigAgentDebug            = "log_debug"
	ConfigAgentPinCA            = "pin_ca"
	ConfigAgentVerification     = "verification"
	configAgentVerificationKey  = "verification_key"
)

func SetAgentDefaults(c interfaces.Config) interfaces.Parameters {
	s := c.NewSet(ConfigAgentSet)
	s.SetConstraint(ConfigAgentSyncInterval, 5, 86400, 300)
	s.SetConstraint(ConfigAgentSyncPending, 5, 86400, 60)
	s.SetConstraint(ConfigAgentSyncRetry, 5, 86400, 10)
	s.SetConstraint(ConfigAgentSyncLost, 5, 86400, 60)
	s.SetConstraint(ConfigAgentStatusInterval, 5, 86400, 21600)
	s.SetConstraint(ConfigAgentLogRetention, 1, 365, 30)
	s.SetConstraint(ConfigAgentLogStdout, 0, 0, true)
	s.SetConstraint(ConfigAgentLogWindowsDisk, 0, 0, true)
	s.SetConstraint(ConfigAgentLogWindowsEvents, 0, 0, true)
	s.SetConstraint(ConfigAgentLogMacOSDisk, 0, 0, true)
	s.SetConstraint(ConfigAgentLogLinuxDisk, 0, 0, true)
	s.SetConstraint(ConfigAgentDebug, 0, 0, false)
	s.SetConstraint(ConfigAgentPinCA, 0, 0, false)
	s.SetConstraint(ConfigAgentVerification, 0, 0, false)
	s.SetConstraint(configAgentVerificationKey, 0, 0, "")
	return s
}
