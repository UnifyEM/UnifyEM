/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package communications

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// SendMessage sends a message to the server
func (c *Communications) SendMessage(message string) error {

	// Get the agent ID
	agentID := c.conf.AP.Get(global.ConfigAgentID).String()
	if agentID == "" {
		return errors.New("agentID is empty")
	}

	// Get the server URL
	serverURL := c.conf.AP.Get(global.ConfigServerURL).String()
	if serverURL == "" {
		return fmt.Errorf("unable to obtain ServerURL")
	}

	// Use a sync request object to send a message
	request := schema.AgentSyncRequest{
		Version:   global.Version,
		Build:     global.Build,
		Responses: nil,
		Messages: []schema.AgentMessage{{
			AgentID:     agentID,
			Sent:        time.Now(),
			MessageType: schema.AgentEventMessage,
			Message:     message}}}

	// Send the sync request
	resp, err := c.post(serverURL, schema.EndpointSync, true, request)
	if err != nil {
		return fmt.Errorf("post error: %w", err)
	}

	// Unmarshal the response body into schema.ServerResponse
	// This provides the details in the event of an error
	var serverResponse schema.APISyncResponse
	err = json.Unmarshal(resp, &serverResponse)
	if err != nil {
		return fmt.Errorf("deserialization error: %w", err)
	}

	// Log the result
	if serverResponse.Code != 200 {
		return fmt.Errorf("failed with HTTP code %d: %s", serverResponse.Code, serverResponse.Details)
	}

	return nil
}
