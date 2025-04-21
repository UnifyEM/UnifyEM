//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package userAdd

import (
	"errors"
	"fmt"
	"strings"

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

	makeAdmin := false
	admin, ok := request.Parameters["admin"]
	if ok {
		if strings.ToLower(admin) == "true" || strings.ToLower(admin) == "yes" {
			makeAdmin = true
		}
	}

	// Assemble log fields
	f := fields.NewFields(
		fields.NewField("cmd", request.Request),
		fields.NewField("requester", request.Requester),
		fields.NewField("request_id", request.RequestID),
		fields.NewField("username", username),
		fields.NewField("admin", fmt.Sprintf("%t", makeAdmin)),
	)

	a := osActions.New(h.logger)
	err := a.AddUser(username, password, makeAdmin)
	if err != nil {
		h.logger.Error(8201, "failed to add user", f)
		response.Response = fmt.Sprintf("failed to add user: %s", err.Error())
		return response, err
	}

	h.logger.Info(8200, "user added", f)
	response.Success = true
	response.Response = "user added successfully"
	return response, nil
}
