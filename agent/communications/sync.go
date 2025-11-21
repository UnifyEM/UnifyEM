/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"encoding/json"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Sync communicates with the UEM server
func (c *Communications) Sync() {
	var err error
	var serverURL string

	// Get the server URL
	serverURL = c.conf.AP.Get(global.ConfigServerURL).String()
	if serverURL == "" {

		// This is likely because the agent is not registered get
		// Requesting the token will trigger a refresh or registration
		_, err = c.GetToken()
		if err != nil {
			c.logger.Error(8020, "server URL not set and unable to refresh or register", nil)
			return
		}

		// Try again
		serverURL = c.conf.AP.Get(global.ConfigServerURL).String()
		if serverURL == "" {
			c.logger.Error(8021, "unable to obtain server URL", nil)
			return
		}
	}

	// Get the agent ID
	agentID := c.conf.AP.Get(global.ConfigAgentID).String()
	if agentID == "" {
		c.logger.Error(8022, "agent ID is empty", nil)
		return
	}

	// Get a list of responses waiting to be sent to the server
	responses := c.responses.ReadAll()

	// Create a sync request to send to the server and include any queued responses
	request := schema.AgentSyncRequest{
		Version:   global.Version,
		Build:     global.Build,
		Responses: responses,
	}

	// If lost mode is set, send an alert message
	if global.Lost {
		request.Messages = append(request.Messages,
			schema.AgentMessage{
				AgentID:     agentID,
				Sent:        time.Now(),
				MessageType: schema.AgentEventAlert,
				Message:     "lost mode is active",
			})
	}

	// Send the sync request
	resp, err := c.post(serverURL, schema.EndpointSync, true, request)
	if err != nil {
		c.logger.Errorf(8024, "error sending sync request: %s", err.Error())
		c.responses.ReQueue(responses)
		return
	}

	// Unmarshal the response body into schema.ServerResponse
	var serverResponse schema.APISyncResponse
	err = json.Unmarshal(resp, &serverResponse)
	if err != nil {
		c.logger.Errorf(8025, "error unmarshalling sync response: %s", err.Error())
		c.responses.ReQueue(responses)
		return
	}

	if serverResponse.Code != 200 {
		c.logger.Errorf(8026, "sync failed with code %d: %s", serverResponse.Code, serverResponse.Details)
		c.responses.ReQueue(responses)
		return
	}

	// Check for triggers
	if c.AnyTriggerChanges(serverResponse.Triggers) {
		c.ProcessTriggers(serverResponse.Triggers)
	}

	// Process requests contained in sync response
	for _, req := range serverResponse.Requests {
		c.logger.Info(8027, "queued received request",
			fields.NewFields(
				fields.NewField("request", req.Request),
				fields.NewField("requestID", req.RequestID),
				fields.NewField("requester", req.Requester)))

		c.requests.Add(req)
	}

	// Update the agent config (includes sync intervals)
	c.conf.AC.SetStringMap(serverResponse.Conf)

	// Store service credentials if provided (encrypted with agent's public key)
	if serverResponse.ServiceCredentials != "" {
		c.conf.SetServiceCredentialsEncrypted(serverResponse.ServiceCredentials)
		c.logger.Info(8030, "service credentials received from server", nil)
	}

	// Checkpoint the configuration
	err = c.conf.Checkpoint()
	if err != nil {
		c.logger.Errorf(8028, "error checkpointing configuration: %s", err.Error())
	}

	c.logger.Info(8029, "sync successful", fields.NewFields(
		fields.NewField("SyncInterval", c.conf.AC.Get(schema.ConfigAgentSyncInterval).Int()),
		fields.NewField("StatusInterval", c.conf.AC.Get(schema.ConfigAgentStatusInterval).Int())))
}
