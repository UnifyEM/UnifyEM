/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package db

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// GetAllRequestRecords retrieves all agent metadata from the AgentMeta bucket
func (d *DB) GetAllRequestRecords() (schema.AgentRequestRecordList, error) {
	var allRequests schema.AgentRequestRecordList

	// Iterate over all keys in the bucket
	err := d.ForEach(BucketAgentRequests, func(key, value []byte) error {
		var request schema.AgentRequestRecord
		err := d.deserialize(value, &request)
		if err != nil {
			return fmt.Errorf("failed to deserialize request record: %w", err)
		}
		allRequests.Requests = append(allRequests.Requests, request)
		return nil
	})

	if err != nil {
		return schema.AgentRequestRecordList{}, fmt.Errorf("failed to retrieve all agent metadata: %w", err)
	}

	return allRequests, nil
}
