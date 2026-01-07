/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package data

import (
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/common/crypto"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
	"github.com/UnifyEM/UnifyEM/server/global"
)

type SyncData struct {
	RemoteIP      string
	AgentID       string
	Role          int
	Version       string
	Build         int
	RequestCount  int
	ResponseCount int
	Responses     []schema.AgentResponse
}

// AgentSync updates metadata about the agent, sends responses for processing, and returns any triggers
func (d *Data) AgentSync(data SyncData) schema.AgentTriggers {

	// Log the sync
	d.logger.Info(2701,
		"agent sync",
		fields.NewFields(
			fields.NewField("id", data.AgentID),
			fields.NewField("role", data.Role),
			fields.NewField("requests", data.RequestCount),
			fields.NewField("responses", data.ResponseCount),
			fields.NewField("messages", len(data.Responses)),
			fields.NewField("src_ip", data.RemoteIP),
			fields.NewField("version", data.Version),
			fields.NewField("build", data.Build)))

	// Update the agent metadata
	triggers, err := d.database.AgentSync(data.AgentID, data.RemoteIP, data.Version, data.Build)
	if err != nil {
		d.logger.Error(2708, "error updating agent metadata",
			fields.NewFields(
				fields.NewField("error", err.Error()),
				fields.NewField("id", data.AgentID)))
	}

	// If there are any responses, process them
	for index, response := range data.Responses {
		d.logger.Info(2702, "processing agent response",
			fields.NewFields(
				fields.NewField("id", data.AgentID),
				fields.NewField("role", data.Role),
				fields.NewField("index", index),
				fields.NewField("cmd", response.Cmd),
				fields.NewField("requestID", response.RequestID),
				fields.NewField("success", response.Success),
				fields.NewField("response", response.Response)))

		err := d.processAgentResponse(data.AgentID, response)
		if err != nil {
			d.logger.Error(2705, "error processing agent response",
				fields.NewFields(
					fields.NewField("error", err.Error()),
					fields.NewField("id", data.AgentID),
					fields.NewField("requestID", response.RequestID)))
		}
	}
	return triggers
}

// processAgentResponse processes a single response from an agent
// Security note: agentID has been authenticated and role indicates if this is a test
func (d *Data) processAgentResponse(agentID string, response schema.AgentResponse) error {

	// Agents can send a status update on their own
	// This is indicated by the request ID being "status"
	if response.RequestID == "status" {
		err := d.agentStatus(agentID, response)
		if err != nil {
			return err
		}
		return d.queueResponse(agentID, response)
	}

	// Agents can send service credentials on their own (unsolicited response)
	// This is indicated by the request ID being "none" and cmd being refresh_service_account
	if response.RequestID == "none" && response.Cmd == commands.RefreshServiceAccount {
		err := d.processServiceCredentials(agentID, response)
		if err != nil {
			return err
		}
		return d.queueResponse(agentID, response)
	}

	// Validate the agent response
	request, err := d.database.GetAgentRequest(response.RequestID)
	if err != nil {
		return fmt.Errorf("failed to get agent request: %w", err)
	}

	// Verify the agent ID matches the request
	if request.AgentID != agentID {
		return fmt.Errorf("agent ID does not match request")
	}

	// Update the request record with the response
	request.ResponseDetails = response.Response
	if response.Success {
		request.Status = schema.RequestStatusComplete
	} else {
		request.Status = schema.RequestStatusFailed
	}

	request.ResponseData = response.Data

	// Redact sensitive parameters from completed or failed requests
	if request.Status == schema.RequestStatusComplete || request.Status == schema.RequestStatusFailed {
		if _, exists := request.Parameters["password"]; exists {
			request.Parameters["password"] = "********"
		}
	}

	// Update the request record
	err = d.database.SetAgentRequest(request)
	if err != nil {
		return fmt.Errorf("failed to update agent request: %w", err)
	}

	// If the request was for status, add it to the status bucket as well
	if response.Cmd == commands.Status {
		err = d.agentStatus(agentID, response)
		if err != nil {
			return err
		}
	}

	return d.queueResponse(agentID, response)
}

// processServiceCredentials handles incoming service credentials from agents
// Credentials arrive double-encrypted: first with agent's public key, then with server's public key
// This function decrypts the outer layer and stores the agent-encrypted version in the database
func (d *Data) processServiceCredentials(agentID string, response schema.AgentResponse) error {
	if response.ServiceCredentials == "" {
		return fmt.Errorf("no service credentials in response")
	}

	// Get server's private encryption key
	serverPrivateEnc := d.conf.SP.Get(global.ConfigServerECPrivateEnc).String()
	if serverPrivateEnc == "" {
		return fmt.Errorf("server private encryption key not available")
	}

	// Decrypt outer layer (server encryption) to get agent-encrypted credentials
	decrypted, err := crypto.Decrypt(response.ServiceCredentials, serverPrivateEnc)
	if err != nil {
		return fmt.Errorf("failed to decrypt service credentials: %w", err)
	}

	// Get current agent metadata
	meta, err := d.database.GetAgentMeta(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent metadata: %w", err)
	}

	// Store the agent-encrypted credentials
	meta.ServiceCredentials = string(decrypted)

	// Update agent metadata
	err = d.database.SetAgentMeta(meta)
	if err != nil {
		return fmt.Errorf("failed to update agent metadata with credentials: %w", err)
	}

	d.logger.Info(2710, "service credentials stored for agent",
		fields.NewFields(fields.NewField("agentID", agentID)))

	return nil
}

// agentStatus handles incoming status messages from agents
func (d *Data) agentStatus(agentID string, response schema.AgentResponse) error {

	// Make sure that response.Data is not nil
	if response.Data == nil {
		return fmt.Errorf("response data is nil")
	}

	// Convert response.Data to AgentStatusData (supports both new and legacy formats)
	statusData, err := schema.ConvertAgentStatusData(response.Data)
	if err != nil {
		return fmt.Errorf("unable to convert response data to AgentStatusData: %w", err)
	}

	// Add to the event store
	err = d.database.AddEvent(schema.AgentEvent{
		AgentID:   agentID,
		Time:      time.Now(),
		EventType: schema.AgentEventStatus,
		Event:     "status",
		Details:   statusData.Details})
	if err != nil {
		return fmt.Errorf("failed to add event to event store: %w", err)
	}

	// Update the agent status
	err = d.database.UpdateAgentStatus(agentID, schema.AgentStatus{
		LastUpdated: time.Now(),
		Details:     statusData.Details,
		Info:        statusData.Info})
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	return nil
}

// queueResponse adds a response to the response queue
func (d *Data) queueResponse(agentID string, response schema.AgentResponse) error {

	// Create a detailed log entry for debugging
	f := fields.NewFields(
		fields.NewField("agentID", agentID),
		fields.NewField("requestID", response.RequestID),
		fields.NewField("cmd", response.Cmd),
		fields.NewField("response", response.Response),
		fields.NewField("success", response.Success))

	// Use a type assertion to check if response.Data is a map[string]string and if so log it
	if responseData, ok := response.Data.(map[string]interface{}); ok {
		mapData, err := schema.ConvertMapString(responseData)
		if err != nil {
			return fmt.Errorf("unable to convert response data to map[string]string: %w", err)
		}
		f.AppendMapString(mapData)
	} else {
		// Otherwise just add a field indicating what type it is
		if response.Data != nil {
			f.AppendKV("resp_data_type", fmt.Sprintf("%T", response.Data))
		}
	}

	// Log the response
	d.logger.Info(2703, "agent response", f)

	// If we want to send a message to notify the admin, etc.,
	// this is where we would do it

	return nil
}
