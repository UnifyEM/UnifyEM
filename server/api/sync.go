/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/data"
	"github.com/UnifyEM/UnifyEM/server/queue"
)

// @Summary Agent sync with server
// @Description Agent syncs with the server to send and receive messages
// @Tags Agent communication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param syncRequest body schema.AgentSyncRequest true "Agent sync request"
// @Success 200 {object} schema.APISyncResponse
// @Failure 401 {object} schema.API401
// @Router /sync [post]
// postSync handles sync requests from agents
func (a *API) postSync(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Check if the agent exists - it might have been deleted
	err := a.data.AgentExists(authDetails.ID)
	if err != nil {
		msg := fmt.Sprintf("agent not in database: %s", err.Error())
		if strings.Contains(err.Error(), "key not found") {
			msg = "agent not found, denying access"
		} else {
			msg = fmt.Sprintf("agent check failed, denying access: %s", err.Error())
		}
		a.logger.Error(2800, msg, logFields)

		// Deny access - agent will attempt to re-register if it has a valid token
		return failureResponse
	}

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(2801, fmt.Sprintf("error reading post body: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Deserialize the JSON
	var syncRequest schema.AgentSyncRequest
	err = json.Unmarshal(body, &syncRequest)
	if err != nil {
		a.logger.Error(2802, fmt.Sprintf("deserialization error: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	logFields.Append(
		fields.NewField("version", syncRequest.Version),
		fields.NewField("build", syncRequest.Build),
		fields.NewField("messages", len(syncRequest.Messages)))

	// Queue any received messages immediately in case the connection drops
	// For example, the agent may be acknowledging a wipe command
	for _, message := range syncRequest.Messages {
		// Overwrite the agent ID with the authenticated user and queue it
		message.AgentID = authDetails.ID
		queue.Add(message)
	}

	// Check for missing required fields
	if syncRequest.Version == "" || syncRequest.Build < 1 {
		a.logger.Error(2803, "syn request missing required fields", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "missing required fields", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Get a list of requests for this agent and mark the ones that do not require ack as sent
	requests, err := a.data.GetAgentRequests(authDetails.ID, true)
	if err != nil {
		a.logger.Error(2804, fmt.Sprintf("error retrieving requests: %s", err.Error()), logFields)
	}

	// Record metadata about the sync, process responses from the agent, and retrieve triggers
	triggers := a.data.AgentSync(
		data.SyncData{
			AgentID:       authDetails.ID,
			RemoteIP:      remoteIP,
			Role:          authDetails.Role,
			Version:       syncRequest.Version,
			Build:         syncRequest.Build,
			RequestCount:  len(requests),
			ResponseCount: len(syncRequest.Responses),
			Responses:     syncRequest.Responses,
		})

	// Get service credentials for this agent (encrypted with agent's public key)
	serviceCredentials := a.data.GetServiceCredentials(authDetails.ID)

	// Return the response
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APISyncResponse{
			Status:             schema.APIStatusOK,
			Code:               http.StatusOK,
			Conf:               a.conf.AC.GetMap(),
			Triggers:           triggers,
			Details:            "ok",
			Requests:           requests,
			ServiceCredentials: serviceCredentials}}
}
