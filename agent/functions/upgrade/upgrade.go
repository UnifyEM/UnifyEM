/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package upgrade

import (
	"errors"
	"fmt"
	"os"
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

	// Check for the hash parameter
	hash, ok := request.Parameters["hash"]
	if !ok {
		if global.DisableHash {
			hash = ""
		} else {
			response.Response = fmt.Sprintf("hash parameter is not specified")
			return response, errors.New(response.Response)
		}
	}

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

	// Download the information file
	url := strings.ToLower(fmt.Sprintf("%s%s/%s", serverURL, schema.EndpointFiles, schema.DeployInfoFile))
	infoFile, err := common.Download(h.logger, h.comms, url, hash)
	if err != nil {
		response.Response = fmt.Sprintf("error downloading %s: %s", url, err.Error())
		return response, err
	}

	// Read the JSON file
	upgradeInfo, err := common.FileToMap(infoFile)
	if err != nil {
		response.Response = fmt.Sprintf("error reading %s: %s", infoFile, err.Error())
		return response, err
	}

	// Delete the file
	_ = os.Remove(infoFile)

	// Get the hash for our desired upgrade
	hash, ok = upgradeInfo[requestFile]
	if !ok {
		if global.DisableHash {
			hash = ""
		} else {
			response.Response = fmt.Sprintf("hash for %s not found in %s", requestFile, infoFile)
			return response, errors.New(response.Response)
		}
	}

	url = strings.ToLower(fmt.Sprintf("%s%s/%s", serverURL, schema.EndpointFiles, requestFile))
	var args = []string{"upgrade"}
	err = common.DownloadExecute(h.logger, h.comms, url, args, hash)
	if err != nil {
		response.Response = fmt.Sprintf("error downloading and executing %s: %s", url, err.Error())
		return response, err
	}

	// Update and return response
	response.Response = "Successfully downloaded and executed " + url
	response.Success = true
	return response, nil
}
