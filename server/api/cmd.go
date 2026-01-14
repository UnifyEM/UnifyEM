/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
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
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Send command to agent
// @Description Creates and queues a command request for an agent
// @Tags Agent management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cmdRequest body schema.CmdRequest true "Command request"
// @Success 200 {object} schema.APICmdResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /cmd [post]
// postCmd handles command requests from administrators
func (a *API) postCmd(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(2822, fmt.Sprintf("failed reading body: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Deserialize the JSON
	var cmd schema.CmdRequest
	err = json.Unmarshal(body, &cmd)
	if err != nil {
		a.logger.Error(2823, fmt.Sprintf("deserialization error: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Information to be logged as fields
	logFields.Append(
		fields.NewField("cmd", cmd.Cmd),
		fields.NewField("parameters", cmd.Parameters))

	// Validate the command
	err = commands.Validate(cmd.Cmd, cmd.Parameters)
	if err != nil {
		a.logger.Error(2824, fmt.Sprintf("command validation failed: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "invalid command", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Queue the request
	requestID, err := a.data.AddAgentRequest(schema.AgentRequest{
		Requester:   authDetails.ID,
		Request:     cmd.Cmd,
		AckRequired: commands.IsAckRequired(cmd.Cmd),
		Parameters:  cmd.Parameters,
	})

	if err != nil {
		a.logger.Error(2825, "unable to queue request: "+err.Error(), logFields)
		details := "unable to queue request"
		code := http.StatusInternalServerError

		if strings.Contains(err.Error(), "key not found") {
			details = "agent does not exist"
			code = http.StatusNotFound
		} else {
			if strings.Contains(err.Error(), "agent ID is required") {
				details = "agent ID is required"
				code = http.StatusBadRequest
			}
		}

		return userver.JResponse{
			HTTPCode: code,
			JSONData: schema.API400{
				Details: details,
				Status:  schema.APIStatusError,
				Code:    code}}
	}

	// Log the request
	a.logger.Info(2826, "request queued for agent", logFields)

	// Return response
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APICmdResponse{
			Status:    schema.APIStatusOK,
			Code:      http.StatusOK,
			Details:   "request queued for agent",
			RequestID: requestID,
			AgentID:   cmd.Parameters["agent_id"]}}
}
