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

// failureResponse provides a consistent response to failed authentication attempts
var failureResponse = userver.JResponse{
	HTTPCode: http.StatusUnauthorized,
	JSONData: authFailResponse}

var authFailResponse = schema.API401{
	Status:  schema.APIStatusError,
	Code:    http.StatusUnauthorized,
	Details: "authentication failed"}

// @Summary User authentication
// @Description Authenticate a user and return access and refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body schema.LoginRequest true "User credentials"
// @Success 200 {object} schema.APILoginResponse "Authentication successful"
// @Failure 401 {object} schema.API401 "Authentication failed"
// @Router /login [post]
func (a *API) postLogin(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return failureResponse
	}

	// Deserialize the JSON
	var loginRequest schema.LoginRequest
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		return failureResponse
	}

	// Information to be logged as fields
	logInfo := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", loginRequest.Username))

	// Check for missing required fields
	if loginRequest.Username == "" || loginRequest.Password == "" {
		a.logger.Error(2861, "login missing required fields", logInfo)
		return failureResponse
	}

	// Authenticate user
	accessToken, refreshToken, err := a.data.LoginGetToken(loginRequest.Username, loginRequest.Password)
	if err != nil {
		logInfo.Append(fields.NewField("auth-result", "failed"), fields.NewField("error", err.Error()))
		a.logger.Error(2862, fmt.Sprintf("login failed: %s", err.Error()), logInfo)
		return failureResponse
	}

	logInfo.Append(fields.NewField("auth-result", "success"))
	a.logger.Info(2863, "successful login", logInfo)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APILoginResponse{
			Status:       schema.APIStatusOK,
			Code:         http.StatusOK,
			AccessToken:  accessToken,
			RefreshToken: refreshToken}}
}
