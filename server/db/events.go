//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// AddEvent adds an event to the database. Each agent has their own child bucket for events and the
// event time plus an event ID is used as the key
func (d *DB) AddEvent(event schema.AgentEvent) error {
	return d.db.Update(func(tx *bbolt.Tx) error {

		// Get or create the parent bucket
		parentBucket, err := tx.CreateBucketIfNotExists([]byte(BucketAgentEvents))
		if err != nil {
			return fmt.Errorf("failed to create parent bucket: %w", err)
		}

		// Get or create the child bucket for the agent
		childBucket, err := parentBucket.CreateBucketIfNotExists([]byte(event.AgentID))
		if err != nil {
			return fmt.Errorf("failed to create child bucket: %w", err)
		}

		// Generate a UUID for the event ID if it is not set
		if event.EventID == "" {
			event.EventID = "E-" + uuid.New().String()
		}

		// Create the key using event time and event ID
		key := fmt.Sprintf("%d-%s", event.Time.Unix(), event.EventID)

		// Serialize the event
		data, err := d.serialize(event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		// Store the serialized event in the child bucket
		return childBucket.Put([]byte(key), data)
	})
}

// GetEvents returns a list of events for an agent within a specified time range
func (d *DB) GetEvents(agentID string, startTime, endTime int64, eventType string) ([]schema.AgentEvent, error) {
	var events []schema.AgentEvent

	err := d.db.View(func(tx *bbolt.Tx) error {
		// Get the parent bucket
		parentBucket := tx.Bucket([]byte(BucketAgentEvents))
		if parentBucket == nil {
			return fmt.Errorf("parent bucket not found")
		}

		// Get the child bucket for the agent
		childBucket := parentBucket.Bucket([]byte(agentID))
		if childBucket == nil {
			return fmt.Errorf("child bucket not found")
		}

		// Iterate over the events in the child bucket
		return childBucket.ForEach(func(k, v []byte) error {
			// Parse the event time from the key
			var eventTime int64
			_, err := fmt.Sscanf(string(k), "%d-", &eventTime)
			if err != nil {
				return fmt.Errorf("failed to parse event time: %w", err)
			}

			// Check if the event is within the specified time range
			if (startTime == 0 || eventTime >= startTime) && (endTime == 0 || eventTime <= endTime) {
				var event schema.AgentEvent
				err := d.deserialize(v, &event)
				if err != nil {
					return fmt.Errorf("failed to deserialize event: %w", err)
				}
				if eventType == "" || event.EventType == eventType {
					events = append(events, event)
				}
			}
			return nil
		})
	})
	return events, err
}

// ForEachEvent iterates over all events for an agent within a specified time range
func (d *DB) ForEachEvent(agentID string, startTime, endTime int64, eventType string, callback func(schema.AgentEvent) error) error {
	return d.db.View(func(tx *bbolt.Tx) error {
		// Get the parent bucket
		parentBucket := tx.Bucket([]byte(BucketAgentEvents))
		if parentBucket == nil {
			return fmt.Errorf("parent bucket not found")
		}

		// Get the child bucket for the agent
		childBucket := parentBucket.Bucket([]byte(agentID))
		if childBucket == nil {
			return fmt.Errorf("child bucket not found")
		}

		// Iterate over the events in the child bucket
		return childBucket.ForEach(func(k, v []byte) error {
			// Parse the event time from the key
			var eventTime int64
			_, err := fmt.Sscanf(string(k), "%d-", &eventTime)
			if err != nil {
				return fmt.Errorf("failed to parse event time: %w", err)
			}

			// Check if the event is within the specified time range
			if (startTime == 0 || eventTime >= startTime) && (endTime == 0 || eventTime <= endTime) {
				var event schema.AgentEvent
				err := d.deserialize(v, &event)
				if err != nil {
					return fmt.Errorf("failed to deserialize event: %w", err)
				}
				// Check if the event type matches
				if eventType == "" || event.EventType == eventType {
					return callback(event)
				}
			}
			return nil
		})
	})
}

// DeleteAllEvents removes the child bucket for the agent thus removing all events
func (d *DB) DeleteAllEvents(agentID string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {

		// Get the parent bucket
		parentBucket := tx.Bucket([]byte(BucketAgentEvents))
		if parentBucket == nil {
			return fmt.Errorf("parent bucket not found")
		}

		// Delete the child bucket for the agent
		return parentBucket.DeleteBucket([]byte(agentID))
	})
}

// PruneEvents iterates over all child buckets and removes events older than the specified number of days
func (d *DB) PruneEvents(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days).Unix()

	return d.db.Update(func(tx *bbolt.Tx) error {

		// Get the parent bucket
		parentBucket := tx.Bucket([]byte(BucketAgentEvents))
		if parentBucket == nil {
			return fmt.Errorf("parent bucket not found")
		}

		// Iterate over all child buckets
		return parentBucket.ForEach(func(agentID, _ []byte) error {
			childBucket := parentBucket.Bucket(agentID)
			if childBucket == nil {
				return fmt.Errorf("child bucket not found")
			}

			// Collect keys to delete
			var keysToDelete [][]byte
			err := childBucket.ForEach(func(k, v []byte) error {
				var eventTime int64
				_, err := fmt.Sscanf(string(k), "%d-", &eventTime)
				if err != nil {
					return fmt.Errorf("failed to parse event time: %w", err)
				}

				if eventTime < cutoffTime {
					keysToDelete = append(keysToDelete, k)
				}
				return nil
			})
			if err != nil {
				return err
			}

			// Delete the collected keys
			for _, k := range keysToDelete {
				if err := childBucket.Delete(k); err != nil {
					return fmt.Errorf("failed to delete event: %w", err)
				}
			}
			return nil
		})
	})
}
