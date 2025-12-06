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

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Agent registration
// @Description Agent requests registration and receives access and refresh tokens
// @Tags Agent communication
// @Accept json
// @Produce json
// @Param registerRequest body schema.AgentRegisterRequest true "Agent registration request"
// @Success 200 {object} schema.APIRegisterResponse "Registration successful"
// @Failure 401 {object} schema.API401 "Invalid registration token"
// @Security RegToken
// @Router /register [post]
// postRegister handles registration requests from agents
// This function does not require a bearer token because it is used by agents to register
func (a *API) postRegister(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(2811, fmt.Sprintf("failed reading body: %s", err.Error()),
			fields.NewFields(fields.NewField("src_ip", remoteIP)))
		return failureResponse
	}

	// Deserialize the JSON
	var regRequest schema.AgentRegisterRequest
	err = json.Unmarshal(body, &regRequest)
	if err != nil {
		a.logger.Error(2812, fmt.Sprintf("deserialization error: %s", err.Error()),
			fields.NewFields(fields.NewField("src_ip", remoteIP)))
		return failureResponse
	}

	// Information to be logged as fields
	logInfo := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("version", regRequest.Version),
		fields.NewField("build", regRequest.Build))

	// Check for missing required fields
	if regRequest.Token == "" || regRequest.Version == "" || regRequest.Build < 1 {
		a.logger.Error(2813, "registration request missing required fields", logInfo)
		return failureResponse
	}

	// Attempt registration, this function will verify the registration token
	// If it is correct, it will return the agent ID, password, and a JWT
	regInfo, err := a.data.Register(regRequest, remoteIP)
	if err != nil {
		a.logger.Error(2814, "registration failed: "+err.Error(), logInfo)
		return failureResponse
	}

	logInfo.Append(fields.NewField("id", regInfo.AgentID))
	a.logger.Info(2815, "registered", logInfo)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIRegisterResponse{
			Status:          schema.APIStatusOK,
			Code:            http.StatusOK,
			Details:         "registered",
			AgentID:         regInfo.AgentID,
			AccessToken:     regInfo.AccessToken,
			RefreshToken:    regInfo.RefreshToken,
			ServerPublicSig: regInfo.ServerPublicSig,
			ServerPublicEnc: regInfo.ServerPublicEnc}}
}
