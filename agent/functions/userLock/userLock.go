/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package userLock

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
	var err error

	// Create a response to the server
	response := schema.NewAgentResponse()
	response.Cmd = request.Request
	response.RequestID = request.RequestID
	response.Success = false

	username, ok := request.Parameters["user"]
	if !ok || username == "" {
		response.Response = "username is missing or invalid"
		return response, errors.New(response.Response)
	}

	shutdown := true
	shutdownParam, ok := request.Parameters["shutdown"]
	if ok {
		if strings.ToLower(shutdownParam) == "no" || strings.ToLower(shutdownParam) == "false" {
			shutdown = false
		}
	}

	// Assemble log fields
	f := fields.NewFields(
		fields.NewField("cmd", request.Request),
		fields.NewField("requester", request.Requester),
		fields.NewField("request_id", request.RequestID),
		fields.NewField("user", username),
	)

	userInfo := osActions.UserInfo{
		Username: username,
	}

	if runtime.GOOS == "darwin" {
		userInfo.AdminUser, userInfo.AdminPassword, err = h.config.GetServiceCredentials()
		if err != nil {
			response.Response = fmt.Sprintf("unable to obtain service account credentials: %s", err.Error())
			return response, errors.New(response.Response)
		}
	}

	a := osActions.New(h.logger)
	err = a.LockUser(userInfo, shutdown)
	if err != nil {
		h.logger.Error(8208, "failed to lock user", f)
		response.Response = err.Error()
		return response, err
	}

	h.logger.Info(8207, "user locked", f)
	response.Success = true
	response.Response = fmt.Sprintf("user %s locked successfully", username)
	return response, nil
}
