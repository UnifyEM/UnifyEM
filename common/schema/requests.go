/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package schema

import "time"

const (
	RequestStatusNew      = "new"
	RequestStatusPending  = "pending"
	RequestStatusComplete = "complete"
	RequestStatusFailed   = "failed"
	RequestStatusInvalid  = "invalid"
)

type AgentRequestRecord struct {
	AgentID         string            `json:"agent_id"`
	RequestID       string            `json:"request_id"`
	Request         string            `json:"agent"`
	Requester       string            `json:"requester"`
	AckRequired     bool              `json:"ack_required"`
	Parameters      map[string]string `json:"parameters"`
	Status          string            `json:"status"`
	TimeCreated     time.Time         `json:"time_created"`
	LastUpdated     time.Time         `json:"last_updated"`
	SendCount       int               `json:"send_count"`
	ResponseDetails string            `json:"response_details"`
	ResponseData    any               `json:"response_data,omitempty"`
	Cancelled       bool              `json:"cancelled"`
}

type AgentRequestRecordList struct {
	Requests []AgentRequestRecord `json:"requests"`
}

func NewDBAgentRequest() AgentRequestRecord {
	return AgentRequestRecord{
		Parameters:   make(map[string]string),
		ResponseData: make(map[string]string),
	}
}
