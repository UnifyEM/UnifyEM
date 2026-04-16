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

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// @Summary Upload recovery public key
// @Description Upload the recovery public key to the server for distribution to agents
// @Tags Recovery
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body schema.RecoveryKeyRequest true "Recovery public key"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 500 {object} schema.API500
// @Router /recovery/key [post]
func (a *API) postRecoveryKey(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Error(2910, fmt.Sprintf("failed reading body: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	var keyReq schema.RecoveryKeyRequest
	err = json.Unmarshal(body, &keyReq)
	if err != nil {
		a.logger.Error(2911, fmt.Sprintf("deserialization error: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	if keyReq.PublicKey == "" {
		a.logger.Error(2912, "recovery public key is empty", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "public_key is required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	a.conf.SC.Set(global.ConfigRecoveryPublicKey, keyReq.PublicKey)
	if err = a.conf.Checkpoint(); err != nil {
		a.logger.Error(2913, fmt.Sprintf("failed to save recovery key: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "failed to save recovery key", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	a.logger.Info(2914, "recovery public key updated", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{
			Status:  schema.APIStatusOK,
			Code:    http.StatusOK,
			Details: "recovery public key updated"}}
}

// @Summary Get agent recovery info
// @Description Retrieve the encrypted recovery info blob for an agent
// @Tags Recovery
// @Security BearerAuth
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} schema.APIRecoveryResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 404 {object} schema.API404
// @Router /agent/{id}/recovery [get]
func (a *API) getAgentRecovery(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	agentID := userver.GetParam(req, "id")
	if agentID == "" {
		a.logger.Error(2915, "no agent ID specified", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{Details: "agent ID required", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	logFields.Append(fields.NewField("agentID", agentID))

	info := a.data.GetAgentRecoveryInfo(agentID)
	if info == "" {
		a.logger.Info(2916, "no recovery info available for agent", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusNotFound,
			JSONData: schema.API404{Details: "no recovery info available", Status: schema.APIStatusError, Code: http.StatusNotFound}}
	}

	a.logger.Info(2917, "recovery info retrieved", logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIRecoveryResponse{
			Status:       schema.APIStatusOK,
			Code:         http.StatusOK,
			RecoveryInfo: info}}
}
