//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

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
)

// @Summary Get agent information
// @Description Retrieves agent information with optional ID
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string false "Agent ID"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Failure 500 {object} schema.API500
// @Router /agent/{id} [get]
func (a *API) getAgent(req *http.Request) userver.JResponse {
	var agents schema.AgentList
	var err error

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the agent ID from the URL
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		// Handle the case where no agent ID is provided
		// For example, return a list of all agents
		agents, err = a.data.GetAllAgentMeta()
		if err != nil {
			a.logger.Error(2882, fmt.Sprintf("error retrieving agents: %s", err.Error()), logFields)
			return userver.JResponse{
				HTTPCode: http.StatusInternalServerError,
				JSONData: schema.API500{Details: "error retrieving agents", Status: "error", Code: http.StatusInternalServerError}}
		}
	} else {
		// Desired agent was specified
		agents, err = a.data.GetAgentMeta(agentID)
		if err != nil {
			var msg string
			code := http.StatusInternalServerError
			if strings.Contains(err.Error(), "key not found") {
				msg = "agent not found"
				code = http.StatusNotFound
			} else {
				msg = fmt.Sprintf("error retrieving agent: %s", err.Error())
			}
			a.logger.Error(2884, msg, logFields)
			return userver.JResponse{
				HTTPCode: code,
				JSONData: schema.API404{Details: msg, Status: schema.APIStatusError, Code: code}}
		}
	}

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIAgentInfoResponse{
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
			Data:   agents}}
}

// @Summary Update agent information
// @Description updates an agent by ID
// @Tags Agent management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param agent body schema.AgentMeta true "Agent data"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Failure 500 {object} schema.API500
// @Router /agent/{id} [post]
// @Router /agent/{id} [put]
func (a *API) postAgent(req *http.Request) userver.JResponse {
	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the agent ID from the URL
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		a.logger.Error(2886, "no agent specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add agent ID to log fields
	logFields.Append(fields.NewField("id", agentID))

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(2887, fmt.Sprintf("error reading body: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Deserialize the JSON into a schema.AgentMeta object
	var AgentMeta schema.AgentMeta
	err = json.Unmarshal(body, &AgentMeta)
	if err != nil {
		a.logger.Error(2888, fmt.Sprintf("deserialization failed: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Get the agent's metadata
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		a.logger.Error(2889, fmt.Sprintf("agent not found: %s", agentID), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	// Get the current metadata (there is only one agent in the list_
	currentMeta := agents.Agents[0]

	// Update the fields that are allowed to be updated
	if AgentMeta.FriendlyName != "" {
		currentMeta.FriendlyName = AgentMeta.FriendlyName
		logFields.Append(fields.NewField("friendlyName", AgentMeta.FriendlyName))
	}

	// Triggers are set, but not reset, otherwise multiple
	// triggers would cancel each other. TriggerReset must be used
	// to clear them.
	if AgentMeta.Triggers.Lost {
		currentMeta.Triggers.Lost = true
		logFields.Append(fields.NewField("lost", "true"))
	}

	if AgentMeta.Triggers.Wipe {
		currentMeta.Triggers.Wipe = true
		logFields.Append(fields.NewField("wipe", "true"))
	}

	if AgentMeta.Triggers.Uninstall {
		currentMeta.Triggers.Uninstall = true
		logFields.Append(fields.NewField("uninstall", "true"))
	}

	// Update the agent's metadata
	err = a.data.SetAgentMeta(currentMeta)
	if err != nil {
		a.logger.Error(2890, fmt.Sprintf("update failed: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error updating agent metadata", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	a.logger.Info(2891, "agent updated", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{Status: schema.APIStatusOK, Code: http.StatusOK}}
}

// @Summary Reset agent triggers
// @Description Resets triggers for the specified agent. With respect to the "wipe" and "uninstall" triggers, this is only useful before the agent's next sync with the server.
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /reset/{id} [put]
// @Router /reset/{id} [post]
func (a *API) putAgentResetTriggers(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the agent ID from the URL
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		a.logger.Error(2892, "no agent specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add agent ID to log fields
	logFields.Append(fields.NewField("id", agentID))

	// Get the agent's metadata
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		a.logger.Error(2893, fmt.Sprintf("agent not found: %s", agentID), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	// Get the current metadata (there is only one agent in the list_
	currentMeta := agents.Agents[0]

	// Reset all triggers
	currentMeta.Triggers = schema.NewAgentTriggers()

	logFields.Append(
		fields.NewField("lost", "false"),
		fields.NewField("lock", "false"),
		fields.NewField("wipe", "false"),
		fields.NewField("uninstall", "false"))

	// Update the agent's metadata
	err = a.data.SetAgentMeta(currentMeta)
	if err != nil {
		a.logger.Error(2894, fmt.Sprintf("update failed: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error updating agent metadata", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	a.logger.Info(2895, "agent updated", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{Status: schema.APIStatusOK, Code: http.StatusOK}}
}

// @Summary Delete agent
// @Description Deletes an agent by ID
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id} [delete]
func (a *API) deleteAgent(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the agent ID from the URL
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		a.logger.Error(2896, "no agent specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add agent ID to log fields
	logFields.Append(fields.NewField("id", agentID))

	// Delete the agent
	err := a.data.AgentDelete(agentID)
	var msg string
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			msg = "agent not found"
		} else {
			msg = fmt.Sprintf("error deleting agent: %s", err.Error())
		}
		a.logger.Warning(2897, msg, logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API400{Details: msg, Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	// Log the deletion
	a.logger.Info(2899, "agent deleted", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: "deleted"}}
}
