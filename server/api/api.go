//
// Copyright (c) 2024-2025 Tenebris Technologies Inc.
// Please see the LICENSE file for details
//

package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/UnifyEM/UnifyEM/common/interfaces"
	"github.com/UnifyEM/UnifyEM/common/schema"
	"github.com/UnifyEM/UnifyEM/common/userver"
	"github.com/UnifyEM/UnifyEM/server/data"
	"github.com/UnifyEM/UnifyEM/server/global"
)

type API struct {
	logger interfaces.Logger
	conf   *global.ServerConfig
	data   *data.Data
}

func New(config *global.ServerConfig, logger interfaces.Logger) *API {
	return &API{logger: logger, conf: config}
}

func (a *API) Start() {
	var err error

	// Set up data access
	a.data, err = data.New(a.conf, a.logger)
	if err != nil {
		a.logger.Errorf(2004, "Data error: %s", err.Error())
		return
	}

	// Loop until stopped
	for {
		// Start the API
		a.logger.Infof(2001, "Starting API")
		err := a.startAPI()
		if err != nil {
			a.logger.Errorf(2003, "API error: %s", err.Error())
		} else {
			a.logger.Infof(2002, "API stopped")
			return
		}

		// Sleep before trying again
		time.Sleep(10 * time.Second)
	}
}

func (a *API) startAPI() error {

	// Obtain the listen address and check for command line override
	listen := a.conf.SC.Get(global.ConfigListen).String()
	if global.ListenOverride != "" {
		listen = global.ListenOverride
	}

	// Create a new AServer instance
	s, err := userver.New(
		userver.WithLogger(a.logger),
		userver.WithSEid(2500),
		userver.WithTestHandler(true),
		userver.WithListen(listen),
		userver.WithHTTPTimeout(a.conf.SC.Get(global.ConfigHTTPTimeout).Int()),
		userver.WithHTTPIdleTimeout(a.conf.SC.Get(global.ConfigHTTPIdleTimeout).Int()),
		userver.WithHandlerTimeout(a.conf.SC.Get(global.ConfigHandlerTimeout).Int()),
		userver.WithMaxConcurrent(a.conf.SC.Get(global.ConfigMaxConcurrent).Int()),
		userver.WithPenaltyBox(
			a.conf.SC.Get(global.ConfigPenaltyBoxMin).Int(),
			a.conf.SC.Get(global.ConfigPenaltyBoxMax).Int()),
		userver.WithAuthFunc(a.NewAuthFunc(a.AuthAnyRole())),
		userver.WithFileDir(
			global.FileDirPattern,
			a.conf.SC.Get(global.ConfigFilesPath).String(),
			a.NewAuthFunc(a.AuthAnyRole())))

	if err != nil {
		return err
	}

	if s == nil {
		return errors.New("userver.New() returned nil")
	}

	s.AddRoute(userver.Route{
		Name:     "ping",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointPing,
		JHandler: a.getPing,
		AuthFunc: a.NewAuthFunc(a.AuthAnyRole())})

	s.AddRoute(userver.Route{
		Name:     "login",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointLogin,
		JHandler: a.postLogin,
		AuthFunc: nil})

	s.AddRoute(userver.Route{
		Name:     "sync",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointSync,
		JHandler: a.postSync,
		AuthFunc: a.NewAuthFunc(a.AuthRoles(schema.RoleAgent))})

	s.AddRoute(userver.Route{
		Name:     "register",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointRegister,
		JHandler: a.postRegister,
		AuthFunc: nil})

	s.AddRoute(userver.Route{
		Name:     "refresh",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointRefresh,
		JHandler: a.postRefresh,
		AuthFunc: nil})

	s.AddRoute(userver.Route{
		Name:     "cmd",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointCmd,
		JHandler: a.postCmd,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "agent",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointAgent + "/{id}", // Single agent
		JHandler: a.getAgent,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "agent",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointAgent, // All agents
		JHandler: a.getAgent,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "agent",
		Methods:  []string{"POST", "PUT"}, // Allow either
		Pattern:  schema.EndpointAgent + "/{id}",
		JHandler: a.postAgent,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "agent",
		Methods:  []string{"DELETE"},
		Pattern:  schema.EndpointAgent + "/{id}",
		JHandler: a.deleteAgent,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "reset",
		Methods:  []string{"PUT", "POST"},
		Pattern:  schema.EndpointReset + "/{id}",
		JHandler: a.putAgentResetTriggers,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "report",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointReport,
		JHandler: a.postReport,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "request",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointRequest + "/{id}", // One request
		JHandler: a.getRequest,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "request",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointRequest,
		JHandler: a.getRequest,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())}) // All requests

	s.AddRoute(userver.Route{
		Name:     "request",
		Methods:  []string{"DELETE"},
		Pattern:  schema.EndpointRequest + "/{id}",
		JHandler: a.deleteRequest,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "regToken",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointRegToken,
		JHandler: a.getRegToken,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "regToken-refresh",
		Methods:  []string{"POST"},
		Pattern:  schema.EndpointRegToken,
		JHandler: a.postRegToken,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "events",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointEvents,
		JHandler: a.getEvents,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "agentsConfig",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointConfigAgents,
		JHandler: a.getConfigAgents,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "agentsConfig",
		Methods:  []string{"PUT", "POST"},
		Pattern:  schema.EndpointConfigAgents,
		JHandler: a.putConfigAgents,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "serverConfig",
		Methods:  []string{"GET"},
		Pattern:  schema.EndpointConfigServer,
		JHandler: a.getConfigServer,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "serverConfig",
		Methods:  []string{"PUT", "POST"},
		Pattern:  schema.EndpointConfigServer,
		JHandler: a.putConfigServer,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	s.AddRoute(userver.Route{
		Name:     "createDeployFile",
		Methods:  []string{"PUT", "POST"},
		Pattern:  schema.EndpointCreateDeployFile,
		JHandler: a.createDeployFile,
		AuthFunc: a.NewAuthFunc(a.AuthAdmins())})

	// Start the server
	err = s.Start()
	if err != nil {
		return fmt.Errorf("userver Start(): %w", err)
	}
	return nil
}

// Close closes open files, etc.
func (a *API) Close() {
	a.data.Close()
}

// PruneDB provides a way for the app to trigger database pruning
func (a *API) PruneDB() {
	a.data.PruneDB()
}
