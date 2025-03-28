//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package upgrade

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/functions/common"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
)

// Upgrade downloads the latest agent for the current OS and architecture from the server and installs it

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

	// Determine our operating system and architecture for the download URL
	requestFile := fmt.Sprintf("uem-agent-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		requestFile += ".exe"
	}

	// Get the URL
	serverURL := h.config.AP.Get(global.ConfigServerURL).String()
	if serverURL == "" {
		response.Response = "unable to obtain server URL"
		return response, errors.New(response.Response)
	}

	url := strings.ToLower(fmt.Sprintf("%s%s/%s", serverURL, schema.EndpointFiles, requestFile))
	var args = []string{"upgrade"}
	err := common.DownloadExecute(h.logger, h.comms, url, args)
	if err != nil {
		response.Response = fmt.Sprintf("error downloading and executing %s: %s", url, err.Error())
		return response, err
	}

	// Update and return response
	response.Response = "Successfully downloaded and executed " + url
	response.Success = true
	return response, nil
}
