//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package api

import (
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Ping the server
// @Description Pinging the server tests authentication and communication
// @Security BearerAuth
// @Produce json
// @Tags Testing
// @Success 200 {object} schema.APIGenericResponse
// @Router /ping [get]
func (a *API) getPing(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)

	authDetails := GetAuthDetails(req)

	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	a.logger.Info(2891, "ping", logFields)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{Details: "pong", Status: schema.APIStatusOK, Code: http.StatusOK}}
}
