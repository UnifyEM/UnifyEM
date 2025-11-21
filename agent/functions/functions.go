/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package functions

import (
	"errors"
	"fmt"

	"github.com/UnifyEM/UnifyEM/agent/communications"
	"github.com/UnifyEM/UnifyEM/agent/functions/downloadEx"
	"github.com/UnifyEM/UnifyEM/agent/functions/execute"
	"github.com/UnifyEM/UnifyEM/agent/functions/ping"
	"github.com/UnifyEM/UnifyEM/agent/functions/reboot"
	"github.com/UnifyEM/UnifyEM/agent/functions/refreshServiceAccount"
	"github.com/UnifyEM/UnifyEM/agent/functions/shutdown"
	"github.com/UnifyEM/UnifyEM/agent/functions/status"
	"github.com/UnifyEM/UnifyEM/agent/functions/updateServiceAccount"
	"github.com/UnifyEM/UnifyEM/agent/functions/upgrade"
	"github.com/UnifyEM/UnifyEM/agent/functions/userAdd"
	"github.com/UnifyEM/UnifyEM/agent/functions/userAdmin"
	"github.com/UnifyEM/UnifyEM/agent/functions/userDelete"
	"github.com/UnifyEM/UnifyEM/agent/functions/userList"
	"github.com/UnifyEM/UnifyEM/agent/functions/userLock"
	"github.com/UnifyEM/UnifyEM/agent/functions/userPassword"
	"github.com/UnifyEM/UnifyEM/agent/functions/userUnlock"
	"github.com/UnifyEM/UnifyEM/agent/global"
	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/schema/commands"
)

type Command struct {
	logger         interfaces.Logger
	config         *global.AgentConfig
	comms          *communications.Communications
	handlers       map[string]CmdHandler
	userDataSource status.UserDataSource
}

type CmdHandler interface {
	Cmd(schema.AgentRequest) (schema.AgentResponse, error)
}

//goland:noinspection DuplicatedCode
func New(options ...func(*Command) error) (*Command, error) {
	c := &Command{
		handlers: make(map[string]CmdHandler),
	}

	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}

	// Check for mandatory fields
	if c.logger == nil {
		return nil, errors.New("logger is required")
	}

	if c.config == nil {
		return nil, errors.New("config is required")
	}

	// Add command handlers
	c.addHandler(commands.DownloadExecute, downloadEx.New(c.config, c.logger, c.comms))
	c.addHandler(commands.Execute, execute.New(c.config, c.logger, c.comms))
	c.addHandler(commands.Status, status.New(c.config, c.logger, c.comms, c.userDataSource))
	c.addHandler(commands.Ping, ping.New(c.config, c.logger, c.comms))
	c.addHandler(commands.Reboot, reboot.New(c.config, c.logger, c.comms))
	c.addHandler(commands.Shutdown, shutdown.New(c.config, c.logger, c.comms))
	c.addHandler(commands.Upgrade, upgrade.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UpdateServiceAccount, updateServiceAccount.New(c.config, c.logger, c.comms))
	c.addHandler(commands.RefreshServiceAccount, refreshServiceAccount.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserList, userList.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserAdd, userAdd.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserDelete, userDelete.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserAdmin, userAdmin.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserPassword, userPassword.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserLock, userLock.New(c.config, c.logger, c.comms))
	c.addHandler(commands.UserUnlock, userUnlock.New(c.config, c.logger, c.comms))

	return c, nil
}

func WithLogger(logger interfaces.Logger) func(*Command) error {
	return func(c *Command) error {
		if logger == nil {
			return errors.New("logger is nil")
		}
		c.logger = logger
		return nil
	}
}

func WithConfig(config *global.AgentConfig) func(*Command) error {
	return func(c *Command) error {
		if config == nil {
			return errors.New("config is nil")
		}
		c.config = config
		return nil
	}
}

func WithComms(comms *communications.Communications) func(*Command) error {
	return func(c *Command) error {
		if comms == nil {
			return errors.New("comms is nil")
		}
		c.comms = comms
		return nil
	}
}

func WithUserDataSource(userDataSource status.UserDataSource) func(*Command) error {
	return func(c *Command) error {
		c.userDataSource = userDataSource
		return nil
	}
}

// addHandler adds a command handler to the map
func (c *Command) addHandler(name string, handler CmdHandler) {
	c.handlers[name] = handler
}

// ExecuteRequest processes a request and returns a response
// Errors are returned as part of the response, so a separate error return would be redundant
func (c *Command) ExecuteRequest(request schema.AgentRequest) schema.AgentResponse {

	// Create response object, assume failure
	response := schema.NewAgentResponse()
	response.Cmd = request.Request
	response.RequestID = request.RequestID
	response.Success = false

	// Validate the request. This eliminates the need for each function to validate mandatory parameters, etc.
	err := commands.Validate(request.Request, request.Parameters)
	if err != nil {
		response.Response = fmt.Sprintf("command validation for %s failed: %s", request.Request, err.Error())
		return response
	}

	// Check if the command exists in the handlers map
	handler, exists := c.handlers[request.Request]
	if !exists {
		response.Response = fmt.Sprintf("command not found: %s", request.Request)
		return response
	}

	// Dispatch the request and return the response
	response, err = handler.Cmd(request)
	if err != nil {
		//goland:noinspection GoDfaErrorMayBeNotNil
		response.Response = fmt.Sprintf("command execution failed: %s", err.Error())
		response.Success = false
	} else {
		response.Success = true
	}
	return response
}
