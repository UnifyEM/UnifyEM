/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package updateServiceAccount

import (
	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
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

	// Get double-encrypted credentials from config
	encryptedForServer, err := h.config.GetServiceCredentialsForServer()
	if err != nil {
		response.Response = "failed to encrypt credentials for server: " + err.Error()
		response.Success = false
		h.logger.Errorf(8109, "failed to get credentials for server: %s", err.Error())
		return response, nil
	}

	// Set response fields
	response.ServiceCredentials = encryptedForServer
	response.Response = "service credentials encrypted and ready"
	response.Success = true

	h.logger.Info(8110, "service credentials prepared for server", nil)
	return response, nil
}
