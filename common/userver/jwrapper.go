/******************************************************************************
 * Copyright (c) 2024-2025 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package userver

import (
	"encoding/json"
	"net/http"

	"github.com/UnifyEM/UnifyEM/common/fields"
)

// JWrapper wraps a JHandler to a standard http.Handler.
// It marshals the JSON data and logs any errors.
// This allows APIs to avoid providing http.Handler directly.
func (s *HServer) JWrapper(name string, h JHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Get the source IP
		src := s.getIP(req)

		// Call the actual handler to service the agent
		respData := h(req)

		// Set reply headers
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		// Send the response
		w.WriteHeader(respData.HTTPCode)
		if err := json.NewEncoder(w).Encode(respData.JSONData); err != nil {
			s.Logger.Error(s.SEid+11,
				"Error writing response",
				fields.NewFields(
					fields.NewField("error", err.Error()),
					fields.NewField("src_ip", src),
					fields.NewField("method", req.Method),
					fields.NewField("uri", req.RequestURI),
					fields.NewField("handler", name)))
		}
	})
}
