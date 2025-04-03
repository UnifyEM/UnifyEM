//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package db

import (
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// PruneAgents removes agents that have been inactive for more than the specified number of days
func (d *DB) PruneAgents(days int) error {

	// Calculate the cutoff time
	cutoff := time.Now().AddDate(0, 0, -days)

	// Iterate over all keys in the AgentMeta bucket
	err := d.ForEach(BucketAgentMeta, func(key, value []byte) error {
		var meta schema.AgentMeta
		err := d.deserialize(value, &meta)
		if err != nil {
			d.logger.Warning(3011, fmt.Sprintf("failed to deserialize agent metadata: %s", err.Error()),
				fields.NewFields(
					fields.NewField("key", string(key)),
					fields.NewField("error", err.Error())))

			// Attempt to delete the bad record
			_ = d.DeleteData(BucketAgentMeta, string(key))
			return nil
		}

		// Check if the agent is inactive and should be pruned
		if meta.LastSeen.Before(cutoff) {

			// Delete any outstanding requests on a best-effort basis
			_ = d.DeleteAgentRequests(meta.AgentID)

			// Delete agent events on a best-effort basis
			_ = d.DeleteAllEvents(meta.AgentID)

			// Delete the agent metadata
			err = d.DeleteData(BucketAgentMeta, string(key))
			if err != nil {
				// Log the error but continue so that one bad record doesn't stop the whole process
				d.logger.Warning(3012, "pruning failed to delete agent metadata",
					fields.NewFields(
						fields.NewField("agent_id", meta.AgentID),
						fields.NewField("last_seen", meta.LastSeen),
						fields.NewField("error", err.Error())))
			} else {
				d.logger.Info(3010, "pruned agent", fields.NewFields(
					fields.NewField("agent_id", meta.AgentID),
					fields.NewField("last_seen", meta.LastSeen)))
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to prune agents: %w", err)
	}
	return nil
}
