/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/common/fields"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// SetAgentRequest stores an agent request in the database
func (d *DB) SetAgentRequest(request schema.AgentRequestRecord) error {

	// Always update the LastUpdated field
	request.LastUpdated = time.Now()

	// Validate critical fields
	if request.AgentID == "" {
		return errors.New("agentID is required")
	}
	if request.RequestID == "" {
		return errors.New("requestID is required")
	}

	// Use the SetData function to serialize and store the agent
	//err := d.SetData(BucketAgentRequests, CombineKeys(request.AgentID, request.RequestID), request)
	err := d.SetData(BucketAgentRequests, request.RequestID, request)
	if err != nil {
		return fmt.Errorf("failed to store agent agent: %w", err)
	}

	return nil
}

// GetAgentRequest retrieves an agent request from the database
func (d *DB) GetAgentRequest(requestKey string) (schema.AgentRequestRecord, error) {
	result := schema.NewDBAgentRequest()
	err := d.GetData(BucketAgentRequests, requestKey, &result)
	return result, err
}

// GetAgentRequests retrieves all agent requests for a given agent
func (d *DB) GetAgentRequests(agentID string) ([]schema.AgentRequestRecord, error) {
	var result []schema.AgentRequestRecord
	err := d.ForEach(BucketAgentRequests, func(key, value []byte) error {
		var request schema.AgentRequestRecord
		err := d.deserialize(value, &request)
		if err != nil {
			// Delete the bad record
			_ = d.DeleteData(BucketAgentRequests, string(key))
			return fmt.Errorf("failed to deserialize agent request, deleting bad record %s: %w", key, err)
		}
		if request.AgentID == agentID {
			result = append(result, request)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve agent requests: %w", err)
	}

	return result, nil
}

// DeleteAgentRequest deletes an agent request from the database
func (d *DB) DeleteAgentRequest(requestKey string) error {
	return d.DeleteData(BucketAgentRequests, requestKey)
}

// DeleteAgentRequests deletes all requests for the specified agent
func (d *DB) DeleteAgentRequests(agentID string) error {
	//prefix := validateKey(agentID) + ":"

	// Iterate over all requests and delete those that match the agentID
	err := d.ForEach(BucketAgentRequests, func(key, value []byte) error {
		//if len(key) > len(prefix) {
		//	if string(key[:len(prefix)]) == prefix {
		var request schema.AgentRequestRecord
		err := d.deserialize(value, &request)
		if err != nil {
			return fmt.Errorf("failed to deserialize agent request: %w", err)
		}
		if request.AgentID == agentID {
			err = d.DeleteData(BucketAgentRequests, string(key))
			if err != nil {
				return fmt.Errorf("failed to delete agent request: %w", err)
			}
		}

		return nil

	})
	return err
}

// PruneAgentRequests deletes all request older than the specified number of days
func (d *DB) PruneAgentRequests(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days).Unix()

	return d.ForEach(BucketAgentRequests, func(key, value []byte) error {
		var request schema.AgentRequestRecord
		err := d.deserialize(value, &request)
		if err != nil {
			d.logger.Warning(3032, fmt.Sprintf("failed to deserialize request record: %s", err.Error()),
				fields.NewFields(
					fields.NewField("key", string(key)),
					fields.NewField("error", err.Error())))

			// Attempt to delete the bad record
			_ = d.DeleteData(BucketAgentRequests, string(key))
			return nil
		}

		if request.LastUpdated.Unix() < cutoffTime {
			err = d.DeleteData(BucketAgentRequests, string(key))
			if err != nil {
				// Log the error but continue so that one bad record doesn't stop the whole process
				d.logger.Warning(3031, "pruning failed to delete request",
					fields.NewFields(
						fields.NewField("key", string(key)),
						fields.NewField("last_updated", request.LastUpdated),
						fields.NewField("error", err.Error())))
			} else {
				d.logger.Info(3030, "pruned request", fields.NewFields(
					fields.NewField("key", string(key)),
					fields.NewField("last_updated", request.LastUpdated)))
			}
		}

		return nil
	})
}

// RequestExists checks if a request exists in the database
func (d *DB) RequestExists(requestID string) (bool, error) {
	return d.KeyExists(BucketAgentRequests, requestID)
}
