/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

import "time"

// RecoveryInfo contains recovery information collected by the agent
type RecoveryInfo struct {
	Timestamp       time.Time `json:"timestamp"`
	OS              string    `json:"os"`
	Hostname        string    `json:"hostname"`
	ServiceAccount  string    `json:"service_account,omitempty"`
	ServicePassword string    `json:"service_password,omitempty"`
	BitLockerInfo   string    `json:"bitlocker_info,omitempty"`
}
