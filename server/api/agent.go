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

/*
 * TAG MANAGEMENT ENDPOINTS
 */

// @Summary Get agents by tag
// @Description Retrieves all agents that have the specified tag (case-insensitive)
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param tag path string true "Tag"
// @Success 200 {object} schema.AgentsByTagResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/by-tag/{tag} [get]
func (a *API) getAgentsByTag(req *http.Request) userver.JResponse {
	tag := userver.GetParam(req, "tag")
	if tag == "" {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "tag required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	agents, err := a.data.GetAllAgentMeta()
	if err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving agents", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	// Special case: tag "all" returns all agents
	if strings.ToLower(tag) == "all" {
		return userver.JResponse{
			HTTPCode: http.StatusOK,
			JSONData: schema.AgentsByTagResponse{
				Agents: agents.Agents,
				Status: schema.APIStatusOK,
				Code:   http.StatusOK,
			},
		}
	}

	var matched []schema.AgentMeta
	for _, agent := range agents.Agents {
		for _, t := range agent.Tags {
			// Match case-insensitive
			if strings.ToLower(t) == strings.ToLower(tag) {
				matched = append(matched, agent)
				break
			}
		}
	}
	if len(matched) == 0 {
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "no agents found with tag", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.AgentsByTagResponse{
			Agents: matched,
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
		},
	}
}

// @Summary List agent tags
// @Description Retrieves the list of tags for the specified agent
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} schema.AgentTagsResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/tags [get]
func (a *API) getAgentTags(req *http.Request) userver.JResponse {
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.AgentTagsResponse{
			Tags:   agents.Agents[0].Tags,
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
		},
	}
}

// @Summary Add tags to agent
// @Description Adds one or more tags to the specified agent (duplicates ignored)
// @Tags Agent management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param tags body schema.AgentTagsRequest true "Tags to add"
// @Success 200 {object} schema.AgentTagsResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/tags/add [post]
func (a *API) postAgentTagsAdd(req *http.Request) userver.JResponse {
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}
	currentMeta := agents.Agents[0]

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	var tagReq schema.AgentTagsRequest
	if err := json.Unmarshal(body, &tagReq); err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add tags, ensuring uniqueness
	tagSet := make(map[string]struct{})
	for _, t := range currentMeta.Tags {
		tagSet[t] = struct{}{}
	}
	for _, t := range tagReq.Tags {
		if t != "" {
			tagSet[t] = struct{}{}
		}
	}
	newTags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		newTags = append(newTags, tag)
	}
	currentMeta.Tags = newTags

	if err := a.data.SetAgentMeta(currentMeta); err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error updating agent tags", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.AgentTagsResponse{
			Tags:   currentMeta.Tags,
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
		},
	}
}

// @Summary Remove tags from agent
// @Description Removes one or more tags from the specified agent (case-insensitive)
// @Tags Agent management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param tags body schema.AgentTagsRequest true "Tags to remove"
// @Success 200 {object} schema.AgentTagsResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/tags/remove [post]
func (a *API) postAgentTagsRemove(req *http.Request) userver.JResponse {
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}
	currentMeta := agents.Agents[0]

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	var tagReq schema.AgentTagsRequest
	if err := json.Unmarshal(body, &tagReq); err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Create a set of tags to remove, forcing to lower case
	removeSet := make(map[string]struct{})
	for _, t := range tagReq.Tags {
		removeSet[strings.ToLower(t)] = struct{}{}
	}

	// Copy existing tags that are not in the removeSet to a new list
	// in a case-insensitive manner
	newTags := make([]string, 0, len(currentMeta.Tags))
	for _, t := range currentMeta.Tags {
		if _, found := removeSet[strings.ToLower(t)]; !found {
			newTags = append(newTags, t)
		}
	}

	// Update the agent's metadata
	currentMeta.Tags = newTags
	if err := a.data.SetAgentMeta(currentMeta); err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error updating agent tags", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.AgentTagsResponse{
			Tags:   currentMeta.Tags,
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
		},
	}
}

/*
 * USER MANAGEMENT ENDPOINTS
 */

// @Summary Add users to agent
// @Description Adds one or more users to the specified agent (duplicates ignored, users must exist)
// @Tags Agent management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param users body schema.AgentUsersRequest true "Users to add"
// @Success 200 {object} schema.AgentUsersResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/users/add [post]
func (a *API) postAgentUsersAdd(req *http.Request) userver.JResponse {
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}
	currentMeta := agents.Agents[0]

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	var userReq schema.AgentUsersRequest
	if err := json.Unmarshal(body, &userReq); err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Check that each user exists in the user bucket
	validUsers := make([]string, 0, len(userReq.Users))
	for _, u := range userReq.Users {
		exists, err := a.data.UserExists(u)
		if err != nil {
			a.logger.Error(3301, fmt.Sprintf("error checking user existence: %s", err.Error()), nil)
			return userver.JResponse{
				HTTPCode: http.StatusInternalServerError,
				JSONData: schema.API500{Details: "error checking user existence", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
		}
		if !exists {
			a.logger.Error(3302, fmt.Sprintf("user does not exist: %s", u), nil)
			return userver.JResponse{
				HTTPCode: http.StatusNotFound,
				JSONData: schema.API404{Details: fmt.Sprintf("user does not exist: %s", u), Status: schema.APIStatusError, Code: http.StatusNotFound}}
		}
		validUsers = append(validUsers, u)
	}

	// Add users, ensuring uniqueness
	userSet := make(map[string]struct{})
	for _, u := range currentMeta.Users {
		userSet[u] = struct{}{}
	}
	for _, u := range validUsers {
		if u != "" {
			userSet[u] = struct{}{}
		}
	}
	newUsers := make([]string, 0, len(userSet))
	for user := range userSet {
		newUsers = append(newUsers, user)
	}
	currentMeta.Users = newUsers

	if err := a.data.SetAgentMeta(currentMeta); err != nil {
		a.logger.Error(3303, fmt.Sprintf("error updating agent users: %s", err.Error()), nil)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error updating agent users", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	// HOOK: Place for future actions when users are added to an agent

	a.logger.Info(3304, fmt.Sprintf("users added to agent %s: %v", agentID, validUsers), nil)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.AgentUsersResponse{
			Users:  currentMeta.Users,
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
		},
	}
}

// @Summary Remove users from agent
// @Description Removes one or more users from the specified agent
// @Tags Agent management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param users body schema.AgentUsersRequest true "Users to remove"
// @Success 200 {object} schema.AgentUsersResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/users/remove [post]
func (a *API) postAgentUsersRemove(req *http.Request) userver.JResponse {
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	agents, err := a.data.GetAgentMeta(agentID)
	if err != nil || len(agents.Agents) != 1 {
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}
	currentMeta := agents.Agents[0]

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}
	var userReq schema.AgentUsersRequest
	if err := json.Unmarshal(body, &userReq); err != nil {
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Remove users
	removeSet := make(map[string]struct{})
	for _, u := range userReq.Users {
		removeSet[u] = struct{}{}
	}
	newUsers := make([]string, 0, len(currentMeta.Users))
	for _, u := range currentMeta.Users {
		if _, found := removeSet[u]; !found {
			newUsers = append(newUsers, u)
		}
	}
	currentMeta.Users = newUsers

	if err := a.data.SetAgentMeta(currentMeta); err != nil {
		a.logger.Error(3305, fmt.Sprintf("error updating agent users: %s", err.Error()), nil)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error updating agent users", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	// HOOK: Place for future actions when users are removed from an agent

	a.logger.Info(3306, fmt.Sprintf("users removed from agent %s: %v", agentID, userReq.Users), nil)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.AgentUsersResponse{
			Users:  currentMeta.Users,
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
		},
	}
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
