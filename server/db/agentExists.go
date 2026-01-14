/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"fmt"
)

// AgentExists checks if an agent exists in the database
func (d *DB) AgentExists(agentID string) error {
	err := d.GetData(BucketAgentMeta, validateKey(agentID), nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve agent metadata: %w", err)
	}

	return nil
}

// AgentActive checks if an agent is active
// Errors are not returned because this function is primarily used to verify
// that an agent exists and is marked active in the database
func (d *DB) AgentActive(agentID string) bool {
	meta, err := d.GetAgentMeta(agentID)
	if err != nil {
		return false
	}

	return meta.Active
}
