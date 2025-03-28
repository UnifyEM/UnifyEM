//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package userPassword

import (
	"errors"
	"fmt"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/agent/osActions"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

type Handler struct {
	config *global.AgentConfig
	logger interfaces.Logger
	comms  *communications.Communications
}

func New(config *global.AgentConfig, logger interfaces.Logger, comms *communications.Communications) *Handler {
	return &Handler{
		config: config,
		logger: logger,
		comms:  comms,
	}
}

func (h *Handler) Cmd(request schema.AgentRequest) (schema.AgentResponse, error) {

	// Create a response to the server
	response := schema.NewAgentResponse()
	response.Cmd = request.Request
	response.RequestID = request.RequestID
	response.Success = false

	username, ok := request.Parameters["username"]
	if !ok || username == "" {
		response.Response = "username is missing or invalid"
		return response, errors.New(response.Response)
	}

	password, ok := request.Parameters["password"]
	if !ok || password == "" {
		response.Response = "password is missing or invalid"
		return response, errors.New(response.Response)
	}

	// Assemble log fields
	f := fields.NewFields(
		fields.NewField("cmd", request.Request),
		fields.NewField("requester", request.Requester),
		fields.NewField("request_id", request.RequestID),
		fields.NewField("username", username),
	)

	a := osActions.New(h.logger)
	err := a.SetPassword(username, password)
	if err != nil {
		h.logger.Error(8212, "failed to set password", f)
		response.Response = fmt.Sprintf("failed to set password: %s", err.Error())
		return response, err
	}

	h.logger.Info(8211, "password set", f)
	response.Success = true
	response.Response = "password set successfully"
	return response, nil
}
