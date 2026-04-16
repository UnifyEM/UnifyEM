/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"fmt"
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Retrieve request status information
// @Description Obtain information about one or more requests sent to agents This includes the status and agent response, if any.
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} schema.APIRequestStatusResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Failure 500 {object} schema.API500
// @Router /request/{id} [get]
func (a *API) getRequest(req *http.Request) userver.JResponse {
	var requests schema.AgentRequestRecordList
	var err error

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the request ID from the URL
	requestID := userver.GetParam(req, "id")
	if requestID == "" {
		// Get all requests if no request ID is specified
		requests, err = a.data.GetRequestRecords()
		if err != nil {
			a.logger.Error(2841, fmt.Sprintf("error retrieving requests %s", err.Error()), logFields)
			return userver.JResponse{
				HTTPCode: http.StatusInternalServerError,
				JSONData: schema.API500{Details: "error retrieving requests", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
		}
	} else {

		// Add request ID to log fields
		logFields.Append(fields.NewField("requestID", requestID))

		// Get the request from the database (it is provided as a list for consistency)
		requests, err = a.data.GetRequestRecord(requestID)
		if err != nil {
			a.logger.Info(2842, fmt.Sprintf("error retrieving agent request: %s", err.Error()), logFields)
			return userver.JResponse{
				HTTPCode: http.StatusNotFound,
				JSONData: schema.API404{Details: "request not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
		}
	}

	// Mask sensitive parameters before returning via API
	for i := range requests.Requests {
		if _, exists := requests.Requests[i].Parameters["password"]; exists {
			requests.Requests[i].Parameters["password"] = "********"
		}
	}

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIRequestStatusResponse{
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
			Data:   requests}}
}

// @Summary Delete request
// @Description Deletes a request by ID
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /request/{id} [delete]
func (a *API) deleteRequest(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the request ID from the URL
	requestID := userver.GetParam(req, "id")
	if requestID == "" {
		a.logger.Error(2844, "no request specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "request ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add request ID to log fields
	logFields.Append(fields.NewField("requestID", requestID))

	// Delete the request from the database
	err := a.data.DeleteAgentRequest(requestID)
	if err != nil {
		a.logger.Info(2845, fmt.Sprintf("error deleting agent request: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "request not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	a.logger.Info(2846, "agent request deleted", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: "request deleted"}}
}

// @Summary Cancel request
// @Description Cancels a pending request by ID
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /request/{id}/cancel [post]
func (a *API) cancelRequest(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the request ID from the URL
	requestID := userver.GetParam(req, "id")
	if requestID == "" {
		a.logger.Error(2847, "no request specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "request ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add request ID to log fields
	logFields.Append(fields.NewField("requestID", requestID))

	// Cancel the request in the database
	err := a.data.CancelAgentRequest(requestID)
	if err != nil {
		a.logger.Info(2848, fmt.Sprintf("error cancelling agent request: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "request not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	a.logger.Info(2849, "agent request cancelled", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: "request cancelled"}}
}

// @Summary Retrieve all requests for an agent
// @Description Returns all request records for a given agent ID
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} schema.APIRequestStatusResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Failure 500 {object} schema.API500
// @Router /agent/{id}/requests [get]
func (a *API) getAgentRequests(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the agent ID from the URL
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		a.logger.Error(2853, "no agent ID specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add agent ID to log fields
	logFields.Append(fields.NewField("agentID", agentID))

	// Get all request records for the agent
	requests, err := a.data.GetAgentRequestRecords(agentID)
	if err != nil {
		a.logger.Error(2854, fmt.Sprintf("error retrieving agent request records: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving agent requests", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	// Mask sensitive parameters before returning via API
	for i := range requests.Requests {
		if _, exists := requests.Requests[i].Parameters["password"]; exists {
			requests.Requests[i].Parameters["password"] = "********"
		}
	}

	a.logger.Info(2857, "agent requests retrieved", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIRequestStatusResponse{
			Status: schema.APIStatusOK,
			Code:   http.StatusOK,
			Data:   requests}}
}

// @Summary Cancel all requests for an agent
// @Description Cancels all pending requests for the specified agent
// @Tags Agent management
// @Security BearerAuth
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/cancel-requests [post]
func (a *API) cancelAgentRequests(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Extract the agent ID from the URL
	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		a.logger.Error(2867, "no agent ID specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Add agent ID to log fields
	logFields.Append(fields.NewField("agentID", agentID))

	// Cancel all requests for the agent in the database
	err := a.data.CancelAgentRequests(agentID)
	if err != nil {
		a.logger.Info(2868, fmt.Sprintf("error cancelling agent requests: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "agent not found", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	a.logger.Info(2869, "agent requests cancelled", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: "agent requests cancelled"}}
}
