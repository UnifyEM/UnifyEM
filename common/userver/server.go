/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// Package userver implements a production grade HTTP server using the
// standard Go libraries and gorilla/mux. It provides a simple way to create
// a server with a set of routes and handlers. Each handler can be either
// a traditional http.Handler or a customer JHandler that returns an
// object that can be marshalled to JSON.
package userver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/netutil"

	"github.com/gorilla/mux"

	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/common/ulogger"
)

// New returns a HServer struct with default values and options applied
func New(options ...func(*HServer) error) (*HServer, error) {
	s := &HServer{
		Listen:           "127.0.0.1:8080",
		HTTPTimeout:      60,
		HTTPIdleTimeout:  60,
		HandlerTimeout:   60,
		PenaltyBoxMin:    0,
		PenaltyBoxMax:    0,
		MaxConcurrent:    100,
		LogFile:          "", // Default to stdout
		DownFile:         "",
		SEid:             0,
		HealthHandler:    true,
		TestHandler:      false,
		StrictSlash:      false,
		DefaultHeaders:   true,
		TLS:              false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		TLSStrongCiphers: true,
		Debug:            false,
	}

	// Process options (see options.go)
	for _, op := range options {
		err := op(s)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Start starts the API
func (s *HServer) Start() error {
	var err error

	// If there is no logger, create a new one and include stdout
	if s.Logger == nil {
		s.Logger, err = ulogger.New(
			ulogger.WithLogFile(s.LogFile),
			ulogger.WithLogStdout(true),
			ulogger.WithRetention(0),
			ulogger.WithDebug(s.Debug))
		if err != nil {
			return err
		}
	}

	s.Logger.Info(s.SEid+1,
		"Starting server", fields.NewFields(fields.NewField("listen", s.Listen)))

	// Add default headers if requested
	if s.DefaultHeaders {
		s.AddHeader("Cache-Control", "no-cache, no-store, must-revalidate")
		s.AddHeader("Pragma", "no-cache")
		s.AddHeader("Expires", "0")
	}

	// Add the health handler if requested
	if s.HealthHandler {
		s.AddRoute(Route{
			Name:     "health",
			Methods:  []string{"GET"},
			Pattern:  "/health",
			JHandler: s.HandlerHealth,
		})
	}

	// Add the test handler if requested
	if s.TestHandler {
		s.AddRoutes(Routes{
			Route{
				Name:     "test",
				Methods:  []string{"GET"},
				Pattern:  "/test",
				JHandler: s.HandlerTest,
			},
			Route{
				Name:     "test",
				Methods:  []string{"GET"},
				Pattern:  "/test/{id}",
				JHandler: s.HandlerTest,
			},
		})
	}

	// Create a new gorilla/mux router
	router := mux.NewRouter()

	// Iterate through routes
	for _, route := range s.Routes {
		// Use JHandler if set otherwise use Handler
		// Wrap either with Wrapper() for logging
		if route.JHandler != nil {
			handler := s.Wrapper(route.Name, s.JWrapper(route.Name, route.JHandler), route.AuthFunc)
			router.Handle(route.Pattern, handler).Methods(route.Methods...)
		} else if route.Handler != nil {
			handler := s.Wrapper(route.Name, route.Handler, route.AuthFunc)
			router.Handle(route.Pattern, handler).Methods(route.Methods...)
		}
	}

	// Serve files from FileDir if set
	if s.FileSrv.Dir != "" && s.FileSrv.Pattern != "" {
		// Create the file server
		fileServer := http.FileServer(http.Dir(s.FileSrv.Dir))

		// Wrap the file server for logging
		router.PathPrefix(s.FileSrv.Pattern).Handler(s.Wrapper("FileServer", http.StripPrefix(s.FileSrv.Pattern, fileServer), s.FileSrv.AuthFunc))

		// Log creating the file server
		s.Logger.Info(s.SEid+2, fmt.Sprintf("Serving files from %s with pattern %s", s.FileSrv.Dir, s.FileSrv.Pattern), nil)
	}

	// Add catch all and not found handler
	router.NotFoundHandler = s.Wrapper("Handler404", s.JWrapper("Handler404", s.Handler404), s.AuthFunc)
	router.MethodNotAllowedHandler = s.Wrapper("Handler405", s.JWrapper("Handler405", s.Handler405), s.AuthFunc)

	// Create server
	serv := &http.Server{
		Addr:              s.Listen,
		Handler:           router,
		ReadHeaderTimeout: time.Duration(s.HTTPTimeout) * time.Second,
		ReadTimeout:       time.Duration(s.HTTPTimeout) * time.Second,
		WriteTimeout:      time.Duration(s.HTTPTimeout) * time.Second,
		IdleTimeout:       time.Duration(s.HTTPIdleTimeout) * time.Second,
	}

	// Add TLS configuration if option is enabled
	if s.TLS {
		if s.TLSCertFile == "" || s.TLSKeyFile == "" {
			return errors.New("TLS cert or key file not specified")
		}

		// Load the cert and key
		cert, err := tls.LoadX509KeyPair(s.TLSCertFile, s.TLSKeyFile)
		if err != nil {
			return err
		}

		// Create the TLS configuration
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		tlsConfig.MinVersion = tls.VersionTLS12

		if s.TLSStrongCiphers {
			tlsConfig.CipherSuites = []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			}
		}

		// Add to the HTTP server config
		serv.TLSConfig = &tlsConfig
	}

	// Start our customized server
	return s.listen(serv)
}

func (s *HServer) Stop() error {

	// Tell the server it has 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Protect against nil server
	if s.server == nil {
		return errors.New("server is not running")
	}

	// Shutdown the server
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %s", err.Error())
	}

	// Shutdown was successful
	return nil
}

// AddRoutes adds routes to the router
func (s *HServer) AddRoutes(routes Routes) {
	// Iterate over routes and add to the router
	for _, route := range routes {
		s.AddRoute(route)
	}
}

// AddRoute adds a route to the router
func (s *HServer) AddRoute(route Route) {
	s.Routes = append(s.Routes, route)
}

// AddHeader adds a header to the list
func (s *HServer) AddHeader(key, value string) {
	s.Headers = append(s.Headers, Header{key, value})
}

// listen is a replacement for ListenAndServe that implements a concurrent session limit
// using netutil.LimitListener. If maxConcurrent is 0, no limit is imposed.
func (s *HServer) listen(server *http.Server) error {

	// Store the server to allow for a graceful shutdown
	s.server = server

	// Get listen address, default to ":http"
	addr := s.server.Addr
	if addr == "" {
		addr = ":http"
	}

	// Create listener
	rawListener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// If maxConcurrent > 0 wrap the listener with a limited listener
	var listener net.Listener
	if s.MaxConcurrent > 0 {
		listener = netutil.LimitListener(rawListener, s.MaxConcurrent)
	} else {
		listener = rawListener
	}

	// Start TLS or non-TLS listener
	if s.TLS {
		// This will use the previously configured TLS information
		return s.server.ServeTLS(listener, "", "")
	} else {
		return s.server.Serve(listener)
	}
}
