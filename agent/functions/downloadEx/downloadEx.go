//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package downloadEx

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/functions/common"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
)

// This command downloads the specified executable file and executes it with the supplied command line arguments

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

	// Check for the required URL parameter
	url, ok := request.Parameters["url"]
	if !ok {
		response.Response = fmt.Sprintf("url parameter is empty or not specified")
		return response, errors.New(response.Response)
	}

	// Collect options from params in the correct order
	var args []string
	for i := 1; ; i++ {
		key := commands.Arg + strconv.Itoa(i)
		value, exists := request.Parameters[key]
		if !exists {
			break
		}
		args = append(args, value)
	}

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
