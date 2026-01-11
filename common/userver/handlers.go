/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package userver

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// HandlerHealth implements a health check for load balancers, etc.
func (s *HServer) HandlerHealth(_ *http.Request) JResponse {
	var r Response

	// Check for presence of the file that indicates the server is down
	if _, err := os.Stat(s.DownFile); err == nil {
		// file exists, send status down and 503
		r.Status = "down"
		r.Code = http.StatusServiceUnavailable
		r.Details = "server is shutting down"
	} else {
		// does not exist - send ok and 200
		r.Status = "ok"
		r.Code = http.StatusOK
		r.Details = "health check ok"
	}
	return JResponse{
		HTTPCode: r.Code,
		JSONData: r}
}

func (s *HServer) Handler401(_ *http.Request) JResponse {
	s.PenaltyBox()

	return JResponse{
		HTTPCode: http.StatusUnauthorized,
		JSONData: Response{Details: "not authorized", Status: "error", Code: http.StatusUnauthorized}}
}

func (s *HServer) Handler404(_ *http.Request) JResponse {
	s.PenaltyBox()
	return JResponse{
		HTTPCode: http.StatusNotFound,
		JSONData: Response{Details: "object does not exist", Status: "error", Code: http.StatusNotFound}}
}

func (s *HServer) Handler405(_ *http.Request) JResponse {
	s.PenaltyBox()
	return JResponse{
		HTTPCode: http.StatusMethodNotAllowed,
		JSONData: Response{Details: "method not allowed", Status: "error", Code: http.StatusMethodNotAllowed}}
}

// HandlerTest accepts an optional 'id' variable and echos it back
// This is an example of a handler that can receive a variable in the URL or not
// Note that two routes are defined in routes.go, one with the variable and one without
func (s *HServer) HandlerTest(req *http.Request) JResponse {
	var r Response

	// Get parameter
	vars := mux.Vars(req)
	id := vars["id"]

	// Create example response
	r.Status = "ok"
	r.Code = http.StatusOK

	if id == "" {
		r.Details = "no ID received"
	} else {
		r.Details = fmt.Sprintf("received ID %s", id)
	}

	// Send Response
	return JResponse{
		HTTPCode: r.Code,
		JSONData: r}
}
