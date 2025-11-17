/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/global"
)

// @Summary Retrieve registration token
// @Description Retrieves the current registration token
// @Tags "Registration token"
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Failure 500 {object} schema.API500
// @Router /regToken [get]
// getRegKey returns the current registration token
func (a *API) getRegToken(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Get the current server URL and registration key from config
	regToken := a.conf.SP.Get(global.ConfigRegToken).String()
	if regToken == "" {
		a.logger.Error(2851, "error retrieving registration token", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving registration token", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	externalURL := a.conf.SC.Get(global.ConfigExternalULR).String()
	if externalURL == "" {
		a.logger.Error(2852, "error retrieving external URL", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving external URL", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	// Make sure externalURL does not end in a slash
	if externalURL[len(externalURL)-1:] == "/" {
		externalURL = externalURL[:len(externalURL)-1]
	}

	// Generate new base64-encoded format: {"s":"server","t":"token"}
	tokenData := fmt.Sprintf(`{"s":"%s","t":"%s"}`, externalURL, regToken)
	rToken := base64.StdEncoding.EncodeToString([]byte(tokenData))
	a.logger.Info(2850, "get regkey", logFields)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{Details: rToken, Status: schema.APIStatusOK, Code: http.StatusOK}}
}

// @Summary Create new registration token
// @Description Creates a new registration token
// @Tags "Registration token"
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schema.APIGenericResponse
// @Failure 401 {object} schema.API401
// @Failure 500 {object} schema.API500
// @Router /regToken [post]
// postRegKey creates and returns a new registration token
func (a *API) postRegToken(req *http.Request) userver.JResponse {

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	externalURL := a.conf.SC.Get(global.ConfigExternalULR).String()
	if externalURL == "" {
		a.logger.Error(2855, "error retrieving external URL", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error retrieving external URL", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	regToken, err := global.GenerateToken()
	if err != nil {
		a.logger.Error(2856, "error generating new registration token", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error generating new registration token", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	a.conf.SP.Set(global.ConfigRegToken, regToken)
	err = a.conf.Checkpoint()
	if err != nil {
		a.logger.Error(2858, "error saving configuration", logFields)
		return userver.JResponse{
			HTTPCode: http.StatusInternalServerError,
			JSONData: schema.API500{Details: "error saving configuration", Status: schema.APIStatusError, Code: http.StatusInternalServerError}}
	}

	// Make sure externalURL does not end in a slash
	if externalURL[len(externalURL)-1:] == "/" {
		externalURL = externalURL[:len(externalURL)-1]
	}

	// Generate new base64-encoded format: {"s":"server","t":"token"}
	tokenData := fmt.Sprintf(`{"s":"%s","t":"%s"}`, externalURL, regToken)
	rToken := base64.StdEncoding.EncodeToString([]byte(tokenData))
	a.logger.Info(2859, "generate new registration key", logFields)

	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{Details: rToken, Status: schema.APIStatusOK, Code: http.StatusOK}}
}
