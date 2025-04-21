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
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// @Summary Retrieve global agent configuration
// @Description Retrieves the current global agent configuration
// @Tags Configuration
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schema.APIConfigResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Router /config/agents [get]
func (a *API) getConfigAgents(req *http.Request) userver.JResponse {
	return a.getConfigTarget("agents", req)
}

// @Summary Retrieve server configuration
// @Description Retrieves the current server configuration
// @Tags Configuration
// @Security BearerAuth
// @Produce json
// @Success 200 {object} schema.APIConfigResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Router /config/server [get]
func (a *API) getConfigServer(req *http.Request) userver.JResponse {
	return a.getConfigTarget("server", req)
}

func (a *API) getConfigTarget(target string, req *http.Request) userver.JResponse {
	var msg string

	targetLC := strings.ToLower(target)

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role),
		fields.NewField("config_set", targetLC))

	// Get the global agent config
	msg = fmt.Sprintf("config set '%s' is empty", targetLC)
	code := http.StatusBadRequest
	status := schema.APIStatusError
	configMap := make(map[string]string)

	switch targetLC {
	case "agents":
		configMap = a.conf.AC.GetMap()
	case "server":
		configMap = a.conf.SC.GetMap()
	default:
		msg = fmt.Sprintf("invalid config set '%s'", targetLC)
	}

	if len(configMap) > 0 {
		msg = fmt.Sprintf("config set '%s' retrieved", targetLC)
		code = http.StatusOK
		status = schema.APIStatusOK
	}

	a.logger.Info(2900, msg, logFields)
	return userver.JResponse{
		HTTPCode: code,
		JSONData: schema.APIConfigResponse{
			Status:  status,
			Code:    code,
			Details: msg,
			Data:    configMap}}
}

// @Summary Update global agent configuration
// @Description Updates the global agent configuration
// @Tags Configuration
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param configAgents body schema.ConfigRequest true "Agent configuration"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Router /config/agents [post]
// @Router /config/agents [put]
func (a *API) putConfigAgents(req *http.Request) userver.JResponse {
	return a.putConfigTarget("agents", req)
}

// @Summary Update server configuration
// @Description Updates the server configuration
// @Tags Configuration
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param configServer body schema.ConfigRequest true "Server configuration"
// @Success 200 {object} schema.APIGenericResponse
// @Failure 400 {object} schema.API400
// @Failure 401 {object} schema.API401
// @Router /config/server [post]
// @Router /config/server [put]
func (a *API) putConfigServer(req *http.Request) userver.JResponse {
	return a.putConfigTarget("server", req)
}

func (a *API) putConfigTarget(target string, req *http.Request) userver.JResponse {
	var err error
	var msg string

	targetLC := strings.ToLower(target)

	remoteIP := userver.RemoteIP(req)
	authDetails := GetAuthDetails(req)
	logFields := fields.NewFields(
		fields.NewField("src_ip", remoteIP),
		fields.NewField("id", authDetails.ID),
		fields.NewField("role", authDetails.Role))

	// Get the right target
	msg = ""
	var set interfaces.Parameters
	switch targetLC {
	case "agents":
		set = a.conf.AC
	case "server":
		set = a.conf.SC
	default:
		msg = fmt.Sprintf("invalid config set '%s'", targetLC)
	}

	if //goland:noinspection GoDfaConstantCondition
	set == nil || msg != "" {
		if msg == "" {
			msg = fmt.Sprintf("config set '%s' is nil", targetLC)
		}

		logFields.Append(fields.NewField("error", msg))
		a.logger.Warning(2904, msg, logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{
				Details: msg,
				Status:  schema.APIStatusError,
				Code:    http.StatusBadRequest}}
	}

	// Get the JSON post data
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.logger.Warning(2905, fmt.Sprintf("failed reading body: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.APICmdResponse{Details: "error reading body", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	// Deserialize the JSON
	request := schema.ConfigRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		a.logger.Warning(2906, fmt.Sprintf("deserialization error: %s", err.Error()), logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.APICmdResponse{Details: "error unmarshalling JSON", Status: schema.APIStatusError, Code: http.StatusBadRequest}}
	}

	if len(request.Parameters) == 0 {
		msg = "no parameters provided"
		logFields.Append(fields.NewField("error", msg))
		a.logger.Warning(2907, msg, logFields)
		return userver.JResponse{
			HTTPCode: http.StatusBadRequest,
			JSONData: schema.API400{
				Details: msg,
				Status:  schema.APIStatusError,
				Code:    http.StatusBadRequest}}
	}

	// Iterate over the received parameters and make sure they exist
	for key := range request.Parameters {
		if !set.Exists(key) {
			msg = fmt.Sprintf("parameter '%s' does not exist in target set '%s'", key, targetLC)
			logFields.Append(fields.NewField("error", msg))
			a.logger.Error(2908, msg, logFields)
			return userver.JResponse{
				HTTPCode: http.StatusBadRequest,
				JSONData: schema.API400{
					Details: msg,
					Status:  schema.APIStatusError,
					Code:    http.StatusBadRequest}}
		}
	}

	// Set the new values
	set.SetStringMap(request.Parameters)
	_ = a.conf.Checkpoint()

	// Add the config set to the log fields
	logFields.Append(fields.NewField("config_set", targetLC))

	// Claim success
	msg = fmt.Sprintf("config set '%s' updated", targetLC)
	a.logger.Info(2909, msg, logFields)
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: schema.APIGenericResponse{Status: schema.APIStatusOK, Details: msg, Code: http.StatusOK}}
}
