/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// SetAgentMeta stores agent metadata in the AgentMeta bucket
func (d *DB) SetAgentMeta(meta schema.AgentMeta) error {

	// Use the SetData function to serialize and store the agent metadata
	err := d.SetData(BucketAgentMeta, validateKey(meta.AgentID), meta)
	if err != nil {
		return fmt.Errorf("failed to store agent metadata: %w", err)
	}

	return nil
}

// GetAgentMeta retrieves agent metadata from the AgentMeta bucket
func (d *DB) GetAgentMeta(agentID string) (schema.AgentMeta, error) {
	var meta schema.AgentMeta
	err := d.GetData(BucketAgentMeta, validateKey(agentID), &meta)
	if err != nil {
		return meta, fmt.Errorf("failed to retrieve agent metadata: %w", err)
	}

	return meta, nil
}

// DeleteAgentMeta removes agent metadata from the AgentMeta bucket
func (d *DB) DeleteAgentMeta(agentID string) error {
	err := d.DeleteData(BucketAgentMeta, validateKey(agentID))
	if err != nil {
		return fmt.Errorf("failed to delete agent metadata: %w", err)
	}
	return nil
}

// GetOrCreateAgentMeta retrieves agent metadata or creates a new one if it doesn't exist
func (d *DB) GetOrCreateAgentMeta(agentID string) (schema.AgentMeta, error) {
	meta, err := d.GetAgentMeta(agentID)
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			meta = schema.NewAgentMeta(agentID)
		} else {
			return meta, err
		}
	}
	return meta, nil
}

// AgentSync updates the agent metadata and returns triggers
func (d *DB) AgentSync(agentID string, ip string, version string, build int) (schema.AgentTriggers, error) {

	// Ensure all flags are set to false
	triggers := schema.NewAgentTriggers()

	meta, err := d.GetOrCreateAgentMeta(agentID)
	if err != nil {
		return triggers, err
	}

	meta.LastSeen = time.Now()
	meta.LastIP = ip
	meta.Version = version
	meta.Build = build

	// Update any triggers
	triggers = meta.Triggers

	// Save the updated metadata
	err = d.SetAgentMeta(meta)
	if err != nil {
		return triggers, err
	}

	return triggers, nil
}
