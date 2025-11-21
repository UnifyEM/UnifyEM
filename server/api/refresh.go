/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Refresh token
// @Description Refreshes a user authentication token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refreshRequest body schema.RefreshRequest true "Refresh request"
// @Success 200 {object} schema.APITokenRefreshResponse
// @Failure 401 {object} schema.API401
// @Router /refresh [post]
// postLogin handles registration requests from agents
func (a *API) postRefresh(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return failureResponse
	}

	// Deserialize the JSON
	var loginRequest schema.RefreshRequest
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		return failureResponse
	}

	// Information to be logged as fields
	logInfo := fields.NewFields(
		fields.NewField("src_ip", remoteIP))

	tokenData, err := a.data.RefreshToken(loginRequest.RefreshToken, loginRequest.ClientPublicSig, loginRequest.ClientPublicEnc)
	if err != nil {
		logInfo.Append(fields.NewField("refresh-result", "failed"), fields.NewField("error", err.Error()))
		a.logger.Error(2865, "access token refresh failed", logInfo)
		return failureResponse
	}

	logInfo.Append(fields.NewField("refresh-result", "success"))
	a.logger.Info(2866, "successful access token refresh", logInfo)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APITokenRefreshResponse{
			Status:          schema.APIStatusOK,
			Code:            http.StatusOK,
			AccessToken:     tokenData.AccessToken,
			ServerPublicSig: tokenData.ServerPublicSig,
			ServerPublicEnc: tokenData.ServerPublicEnc}}
}
