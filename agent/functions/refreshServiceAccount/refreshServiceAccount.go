/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package refreshServiceAccount

import (
	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/osActions"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

type Handler struct {
	config *global.AgentConfig
	logger interfaces.Logger
	comms  *communications.Communications
}

func New(config *global.AgentConfig, logger interfaces.Logger, comms *communications.Communications) *Handler {
	return &Handler{config: config, logger: logger, comms: comms}
}

func (h *Handler) Cmd(request schema.AgentRequest) (schema.AgentResponse, error) {
	response := schema.NewAgentResponse()
	response.Cmd = request.Request
	response.RequestID = request.RequestID

	if //goland:noinspection GoBoolExpressions
	!global.HaveServiceAccount {
		response.Response = "service account not implemented on this platform"
		response.Success = true
		h.logger.Infof(8126, "service account not implemented on this platform")
		return response, nil
	}

	// Get current credentials from config
	username, oldPassword, err := h.config.GetServiceCredentials()
	if err != nil {
		response.Response = "failed to get current credentials: " + err.Error()
		response.Success = false
		h.logger.Errorf(8120, "failed to get current credentials: %s", err.Error())
		return response, nil
	}

	userInfo := osActions.UserInfo{
		Username: username,
		Password: oldPassword,
	}

	// Create osActions instance
	actions := osActions.New(h.logger)

	// Refresh the service account password
	newPassword, err := actions.RefreshServiceAccount(userInfo)
	if err != nil {
		response.Response = "failed to refresh service account: " + err.Error()
		response.Success = false
		h.logger.Errorf(8122, "failed to refresh service account: %s", err.Error())
		return response, nil
	}

	// Empty newPassword means platform doesn't support refresh (Linux/Windows stubs)
	if newPassword == "" {
		response.Response = "service account refresh not supported on this platform"
		response.Success = true
		h.logger.Info(8123, "service account refresh not supported on this platform", nil)
		return response, nil
	}

	// Store new credentials (automatically encrypts and flags for sending)
	err = h.config.SetServiceCredentials(username, newPassword)
	if err != nil {
		response.Response = "failed to store new credentials: " + err.Error()
		response.Success = false
		h.logger.Errorf(8124, "failed to store new credentials: %s", err.Error())
		return response, nil
	}

	response.Response = "service account password refreshed successfully"
	response.Success = true
	h.logger.Info(8125, "service account password refreshed successfully", nil)
	return response, nil
}
