/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

import "github.com/UnifyEM/UnifyEM/common/schema"

func (d *Data) GetEvents(agentID string, startTime, endTime int64, eventType string) ([]schema.AgentEvent, error) {
	return d.database.GetEvents(agentID, startTime, endTime, eventType)
}
