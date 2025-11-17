/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

// This is a simple example of how to implement an HTTP server using the userver package

package main

import (
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/ulogger"
	"github.com/UnifyEM/UnifyEM/common/userver"
)

// Response structure for consistency
type Response struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

func main() {

	// Create a new UEMLogger instance
	logger, err := ulogger.New(
		ulogger.WithPrefix("example"),
		ulogger.WithLogFile("example.log"),
		ulogger.WithDebug(true),
		ulogger.WithLogStdout(true))

	// Create a new HServer instance
	server, err := userver.New(
		userver.WithLogFile("example.log"),
		userver.WithListen(":8080"),
		userver.WithTestHandler(true),
		userver.WithDebug(true),
		userver.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	// Add a handler for /help
	server.AddRoute(userver.Route{
		Name:     "help",
		Methods:  []string{"GET"},
		Pattern:  "/help",
		JHandler: getHelp})

	// Start the server
	err = server.Start()
	if err != nil {
		panic(err)
	}
}

// getHelp returns a help message
// Handlers must accept *http.Request and return an userver.JResponse structure
// See userver/handlers.go for more examples
func getHelp(_ *http.Request) userver.JResponse {
	return userver.JResponse{
		HTTPCode: http.StatusOK,
		JSONData: Response{Details: "This is a help message", Status: "ok", Code: http.StatusOK}}
}
