/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

import (
	"fmt"

	"github.com/UnifyEM/UnifyEM/common/schema"
)

// GetAllAgentMeta returns a list of all agent metadata
func (d *Data) GetAllAgentMeta() (schema.AgentList, error) {
	return d.database.GetAllAgentMeta()
}

// GetAgentMeta returns metadata for a single agent but wraps it in
// a schema.AgentList for consistency
func (d *Data) GetAgentMeta(agentID string) (schema.AgentList, error) {
	agent, err := d.database.GetAgentMeta(agentID)
	if err != nil {
		return schema.AgentList{}, err
	}
	return schema.AgentList{Agents: []schema.AgentMeta{agent}}, nil
}

func (d *Data) SetAgentMeta(meta schema.AgentMeta) error {
	return d.database.SetAgentMeta(meta)
}

func (d *Data) AgentExists(agentID string) error {
	return d.database.AgentExists(agentID)
}

// GetServiceCredentials returns service credentials for an agent
// Credentials are stored encrypted with the agent's public key
func (d *Data) GetServiceCredentials(agentID string) string {
	meta, err := d.database.GetAgentMeta(agentID)
	if err != nil {
		return ""
	}
	return meta.ServiceCredentials
}

// SetAgentRecoveryInfo stores the encrypted recovery info blob for an agent
func (d *Data) SetAgentRecoveryInfo(agentID string, info string) error {
	meta, err := d.database.GetAgentMeta(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent metadata: %w", err)
	}
	meta.RecoveryInfo = info
	return d.database.SetAgentMeta(meta)
}

// GetAgentRecoveryInfo returns the encrypted recovery info blob for an agent
func (d *Data) GetAgentRecoveryInfo(agentID string) string {
	meta, err := d.database.GetAgentMeta(agentID)
	if err != nil {
		return ""
	}
	return meta.RecoveryInfo
}

// AgentDelete removes an agent from the database including any requests
func (d *Data) AgentDelete(agentID string) error {
	var err error

	// First delete any requests
	err = d.database.DeleteAgentRequests(agentID)
	if err != nil {
		return err
	}

	// Delete agent events
	err = d.database.DeleteAllEvents(agentID)

	// Delete agent metadata
	return d.database.DeleteAgentMeta(agentID)
}

// NewAgentMessage adds a message event to the database
func (d *Data) NewAgentMessage(message schema.AgentMessage) error {
	return d.database.AddEvent(schema.AgentEvent{
		AgentID:   message.AgentID,
		Event:     message.Message,
		Time:      message.Sent,
		EventType: schema.AgentEventMessage})
}
