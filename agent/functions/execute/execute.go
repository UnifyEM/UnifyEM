/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package execute

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
)

const maxOutputSize = 10240

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

	returnData := make(map[string]string)

	// Create a response to the server
	response := schema.NewAgentResponse()
	response.Cmd = request.Request
	response.RequestID = request.RequestID
	response.Response = "collected"
	response.Data = &returnData
	response.Success = true

	cmd := request.Parameters["cmd"]
	if cmd == "" {
		response.Response = "cmd parameter is empty or not specified"
		response.Success = false
		return response, errors.New(response.Response)
	}

	// Collect options from params in the correct order
	humanReadable := cmd
	var args []string
	for i := 1; ; i++ {
		key := commands.Arg + strconv.Itoa(i)
		value, exists := request.Parameters[key]
		if !exists {
			break
		}
		args = append(args, value)
		humanReadable += " " + value
	}

	// Assemble log fields
	f := fields.NewFields(
		fields.NewField("cmd", request.Request),
		fields.NewField("requester", request.Requester),
		fields.NewField("request_id", request.RequestID),
	)

	// Log the event
	h.logger.Info(8201, fmt.Sprintf("executing \"%s\"", humanReadable), f)

	// Execute the file with the supplied arguments
	returnData["exit_status"] = "0"
	command := exec.Command(cmd, args...)

	var outputBuffer bytes.Buffer
	command.Stdout = &outputBuffer
	command.Stderr = &outputBuffer

	err := command.Start()
	if err != nil {
		response.Response = fmt.Sprintf("error starting command \"%s\": %s", humanReadable, err.Error())
		response.Success = false
		return response, err
	}

	err = command.Wait()
	if err != nil {
		// Check if the error is an exit status
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			returnData["exit_status"] = fmt.Sprintf("%d", exitError.ExitCode())
		}
		response.Response = fmt.Sprintf("error executing \"%s\": %s", humanReadable, err.Error())
		response.Success = false
	}

	response.Success = true
	response.Response = "executed"

	// Capture output
	output := outputBuffer.Bytes()
	if len(output) > maxOutputSize {
		output = output[:maxOutputSize]
		response.Response += " [output truncated]"
	}

	// Log the output
	returnData["output"] = string(output)
	h.logger.Infof(8202, "executed \"%s\", exit status %s", humanReadable, returnData["exit_status"])

	// Return the response
	return response, nil
}
