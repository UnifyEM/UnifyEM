//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package ping

import (
	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Ping responds with "pong" to a "ping" request

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
	response.Response = "pong"
	response.Success = true

	// Assemble log fields
	f := fields.NewFields(
		fields.NewField("cmd", request.Request),
		fields.NewField("requester", request.Requester),
		fields.NewField("request_id", request.RequestID),
	)

	// Log the event
	h.logger.Info(8205, "received ping", f)

	// Return the response
	return response, nil
}
