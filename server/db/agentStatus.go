/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// UpdateAgentStatus updates status information in the agent metadata
func (d *DB) UpdateAgentStatus(agentID string, status schema.AgentStatus) error {

	// Always update the LastUpdated field
	status.LastUpdated = time.Now()

	// Retrieve the existing agent meta
	meta, err := d.GetAgentMeta(agentID)
	if err != nil {
		return fmt.Errorf("failed to retrieve agent metadata: %w", err)
	}

	meta.Status = &status

	// Use the SetAgentMeta function to store the updated metadata
	err = d.SetAgentMeta(meta)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	return nil
}
