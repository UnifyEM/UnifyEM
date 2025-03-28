//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package db

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// GetAllAgentMeta retrieves all agent metadata from the AgentMeta bucket
func (d *DB) GetAllAgentMeta() (schema.AgentList, error) {
	var allMeta schema.AgentList

	// Iterate over all keys in the AgentMeta bucket
	err := d.ForEach(BucketAgentMeta, func(key, value []byte) error {
		var meta schema.AgentMeta
		err := d.deserialize(value, &meta)
		if err != nil {
			return fmt.Errorf("failed to deserialize agent metadata: %w", err)
		}
		allMeta.Agents = append(allMeta.Agents, meta)
		return nil
	})

	if err != nil {
		return schema.AgentList{}, fmt.Errorf("failed to retrieve all agent metadata: %w", err)
	}

	return allMeta, nil
}
